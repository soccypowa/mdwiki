package main

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/ast"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// statSafe wraps os.Stat with a local recover, so that even an unexpected
// panic deep in the stdlib/OS layer (seen intermittently on some platforms)
// is treated as "not found" rather than crashing the request.
func statSafe(pathName string) (fi os.FileInfo, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("stat panic on %q: %v", pathName, rec)
			fi = nil
		}
	}()
	return os.Stat(pathName)
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
	if fi, err := statSafe(direct); err == nil && !fi.IsDir() {
		return direct, true
	}

	asIndex := filepath.Join(root, rel, "index.md")
	if fi, err := statSafe(asIndex); err == nil && !fi.IsDir() {
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
	if i := strings.IndexAny(s, "#?"); i >= 0 {
		main, suffix = s[:i], s[i:]
	}

	lower := strings.ToLower(main)
	switch {
	case strings.HasSuffix(lower, ".md"):
		main = main[:len(main)-3]
	case strings.HasSuffix(lower, ".markdown"):
		main = main[:len(main)-9]
	default:
		return dest
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

	caser := cases.Title(language.English)
	return caser.String(base)
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
