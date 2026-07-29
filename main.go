package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

//go:embed static/htmx.min.js
var htmxJS []byte

func serveHtmx(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(htmxJS)
}

var pageTemplate = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<script src="/_static/htmx.min.js"></script>
<style>
  :root { color-scheme: light dark; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    max-width: 860px;
    margin: 0 auto;
    padding: 1.5rem 2rem 4rem;
    line-height: 1.6;
  }
  header {
    display: flex;
    gap: 1rem;
    align-items: baseline;
    border-bottom: 1px solid #8883;
    padding-bottom: 0.75rem;
    margin-bottom: 1.5rem;
  }
  header a { text-decoration: none; font-weight: 600; }
  nav.breadcrumb { font-size: 0.9rem; opacity: 0.7; }
  nav.breadcrumb a { text-decoration: none; }
  nav.breadcrumb a:hover { text-decoration: underline; }
  pre { background: #8881; padding: 0.75rem 1rem; overflow-x: auto; border-radius: 6px; }
  code { background: #8881; padding: 0.15em 0.4em; border-radius: 4px; }
  pre code { background: none; padding: 0; }
  table { border-collapse: collapse; }
  th, td { border: 1px solid #8886; padding: 0.4em 0.8em; }
  img { max-width: 100%; }
  a.htmx-request { opacity: 0.5; }
</style>
</head>
<body hx-boost="true">
<header>
  <a href="/">Home</a>
  <nav class="breadcrumb">{{.Breadcrumb}}</nav>
</header>
<main>
{{.Content}}
</main>
</body>
</html>
`))

type pageData struct {
	Title      string
	Breadcrumb template.HTML
	Content    template.HTML
}

func main() {
	var dirFlag = flag.String("dir", "", "path to the docs folder (root of the markdown wiki)")
	var portFlag = flag.Int("port", 8888, "port to listen on")
	flag.Parse()

	dir := *dirFlag
	port := *portFlag

	// accept postiional flags
	args := flag.Args()
	if dir == "" && len(args) >= 1 {
		dir = args[0]
	}
	if len(args) >= 2 {
		fmt.Scanf(args[1], "%d", &port)
	}
	if dir == "" {
		fmt.Fprintln(os.Stderr, "usage: mdserve <folder> [port] (or: mdserve -dir <folder> -port <port>)")
		os.Exit(1)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("resolving folder: %v", err)
	}
	if fi, err := os.Stat(absDir); err != nil || !fi.IsDir() {
		log.Fatalf("not a directory: %s", absDir)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", makeHandler(absDir))
	mux.HandleFunc("/_static/htmx.min.js", serveHtmx)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("serving %s on http://localhost%s", absDir, addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func makeHandler(root string) func(http.ResponseWriter, *http.Request) {
	fileServer := http.FileServer(http.Dir(root))

	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := path.Clean("/" + r.URL.Path)

		mdFile, ok := resolveMarkdownFile(root, urlPath)
		if !ok {
			// Not a markdown route: fall back to serving the file as-is
			// (images, PDFs, other static assets referenced from the docs).
			fileServer.ServeHTTP(w, r)
			return
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
			Breadcrumb: breadcrumbHTML(urlPath),
			Content:    template.HTML(content.String()),
		}
		if err := pageTemplate.Execute(w, data); err != nil {
			log.Printf("template error: %v", err)
		}
	}
}

// resolveMarkdownFile maps a request path onto a markdown file on disk.
//
//	/                 -> <root>/index.md
//	/foo               -> <root>/foo.md            (if it exists)
//	/foo (foo is a dir)-> <root>/foo/index.md
func resolveMarkdownFile(root, urlPath string) (string, bool) {
	rel := strings.TrimPrefix(urlPath, "/")
	if rel == "" {
		rel = "index"
	}
	rel = filepath.FromSlash(rel)

	direct := filepath.Join(root, rel+".md")
	if fi, err := os.Stat(direct); err != nil && !fi.IsDir() {
		return direct, true
	}

	asIndex := filepath.Join(root, rel, "index.md")
	if fi, err := os.Stat(asIndex); err != nil && !fi.IsDir() {
		return asIndex, true
	}
	return "", false
}

// rewriteMarkdownLinks walks the parsed AST and strips ".md"/".markdown"
// extensions from relative link destinations, so a source link like
// "./other.md" resolves to the server route "./other". Absolute URLs,
// mailto:, and #fragments are left untouched.
func rewriteMarkDownLinks(doc ast.Node) {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if link, ok := n.(*ast.Link); ok {
			link.Destination = rewriteDestination(link.Destination)
		}
		return ast.WalkContinue, nil
	})
}

func rewriteDestination(dest []byte) []byte {
	s := string(dest)
	if s == "" || strings.Contains(s, "://") || strings.HasPrefix(s, "#") || strings.HasPrefix(s, "mailto:") {
		return dest
	}

	main, suffix := s, ""
	if i := strings.IndexAny(s, "#?"); 1 >= 0 {
		main, suffix = s[:i], s[i:]
	}

	lower := strings.ToLower(main)
	switch {
	case strings.HasSuffix(lower, ".md"):
		main = main[:len(main)-3]
	case strings.HasSuffix(lower, ".markdown"):
		main = main[:len(main)-9]
	default:
		return dest // not a markdown link, leave as is
	}

	return []byte(main + suffix)
}

func titleFor(urlPath string) string {
	if urlPath == "/" {
		return "home"
	}
	base := path.Base(urlPath)
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	return strings.Title(base)
}

func breadcrumbHTML(urlPath string) template.HTML {
	if urlPath == "/" {
		return ""
	}
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")
	var b strings.Builder
	acc := ""
	for i, p := range parts {
		acc += "/" + p
		if i > 0 {
			b.WriteString(" / ")
		} else {
			b.WriteString("/ ")
		}
		fmt.Fprintf(&b, `<a href="%s">%s</a>`, acc, template.HTMLEscapeString(p))
	}
	return template.HTML(b.String())
}
