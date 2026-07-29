package main

import (
	_ "embed"
	"html/template"
)

//go:embed static/htmx.min.js
var htmxJS []byte

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
