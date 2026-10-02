package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

type searchResult struct {
	URL     string
	Title   string
	Snippet template.HTML
	Score   int
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func safeFileServer(root string) http.Handler {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		rootPath = root
	}
	resolvedRoot, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		resolvedRoot = rootPath
	}
	fileServer := http.FileServer(http.Dir(rootPath))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := path.Clean("/" + r.URL.Path)
		candidate := filepath.Join(rootPath, filepath.FromSlash(strings.TrimPrefix(urlPath, "/")))
		if fi, err := os.Lstat(candidate); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			http.NotFound(w, r)
			return
		}

		resolvedCandidate, err := filepath.EvalSymlinks(candidate)
		if err == nil {
			rel, err := filepath.Rel(resolvedRoot, resolvedCandidate)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				http.NotFound(w, r)
				return
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		if !os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func makeHandler(root string) func(http.ResponseWriter, *http.Request) {
	fileServer := safeFileServer(root)

	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := path.Clean("/" + r.URL.Path)

		mdFile, ok := resolveMarkdownFile(root, urlPath)
		if !ok {
			fileServer.ServeHTTP(w, r)
			return
		}
		if isIndexFile(root, mdFile) {
			rel, err := filepath.Rel(root, mdFile)
			if err == nil {
				canonicalPath := "/"
				if rel != "index.md" {
					canonicalPath = "/" + filepath.ToSlash(filepath.Join(filepath.Dir(rel), "index"))
				}
				if urlPath != canonicalPath {
					target := *r.URL
					target.Path = canonicalPath
					http.Redirect(w, r, target.String(), http.StatusPermanentRedirect)
					return
				}
			}
		}

		src, err := os.ReadFile(mdFile)
		if err != nil {
			http.Error(w, "could not read "+urlPath, http.StatusInternalServerError)
			return
		}

		reader := text.NewReader(src)
		doc := md.Parser().Parse(reader)
		rewriteMarkDownLinks(doc)

		var content bytes.Buffer
		if err := md.Renderer().Render(&content, src, doc); err != nil {
			http.Error(w, "render error:", http.StatusInternalServerError)
			return
		}

		data := pageData{
			Title:      titleFor(urlPath),
			Breadcrumb: breadcrumbHTML(root, urlPath),
			Content:    template.HTML(content.String()),
		}
		if err := pageTemplate.Execute(w, data); err != nil {
			log.Printf("template error: %v", err)
		}
	}
}

func searchHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if q == "" {
			fmt.Fprint(w, `<p class="search-hint">Type to search...</p>`)
			return
		}
		needle := strings.ToLower(q)

		var results []searchResult
		filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
				return nil
			}
			src, err := os.ReadFile(p)
			if err != nil {
				return nil // skip unreadable instead of failing
			}
			content := string(src)
			lowerContent := strings.ToLower(content)

			title := FirstHeading(content)
			if title == "" {
				title = strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
			}
			titleHint := strings.Contains(strings.ToLower(title), needle)
			contentIdx := strings.Index(lowerContent, needle)

			if !titleHint && contentIdx == -1 {
				return nil // no match in this file
			}

			score := strings.Count(lowerContent, needle)
			if titleHint {
				score += 100 // title match ranks well above body match
			}

			rel, err := filepath.Rel(root, p)
			if err != nil {
				return nil
			}
			url := "/" + strings.TrimSuffix(filepath.ToSlash(rel), ".md")
			if url == "/index" {
				url = "/"
			}

			results = append(results, searchResult{
				URL:     url,
				Title:   title,
				Snippet: SnipperAround(content, contentIdx, len(q)),
				Score:   score,
			})
			return nil
		})

		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
		const maxResults = 20
		if len(results) > maxResults {
			results = results[:maxResults]
		}

		if len(results) == 0 {
			fmt.Fprintf(w, `<p class="search-empty">No results for &quot;%s&quot;</p>`, template.HTMLEscapeString(q))
			return
		}

		var b strings.Builder
		b.WriteString(`<ul class="search-results">`)
		for _, res := range results {
			fmt.Fprintf(&b, `<li><a href="%s" hx-boost="false">%s</a>%s</li>`,
				template.HTMLEscapeString(res.URL),
				template.HTMLEscapeString(res.Title),
				res.Snippet,
			)
		}
		b.WriteString(`</ul>`)
		w.Write([]byte(b.String()))
	}
}

func serveHtmx(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	_, _ = w.Write(htmxJS)
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC handling %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
