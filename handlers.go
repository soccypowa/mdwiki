package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

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
