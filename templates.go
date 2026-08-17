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
  .search-box { position: relative; margin-left: auto; }
  .search-box input[type=search] {
    font: inherit;
    padding: 0.4em 0.7em;
    border: 1px solid #8886;
    border-radius: 6px;
    background: Canvas;
    color: CanvasText;
    width: 14rem;
  }
  .search-box input[type=search]:focus { outline: 2px solid #4488ff88; }
  .search-panel {
    display: none;
    position: absolute;
    right: 0;
    top: calc(100% + 0.4rem);
    width: 22rem;
    max-height: 60vh;
    overflow-y: auto;
    background: Canvas;
    border: 1px solid #8886;
    border-radius: 8px;
    box-shadow: 0 4px 16px #0003;
    padding: 0.5rem;
    z-index: 10;
  }
  .search-box:focus-within .search-panel { display: block; }
  .search-results { list-style: none; margin: 0; padding: 0; }
  .search-results li { padding: 0.5rem 0.4rem; border-bottom: 1px solid #8883; }
  .search-results li:last-child { border-bottom: none; }
  .search-results a { font-weight: 600; text-decoration: none; }
  .search-results a:hover { text-decoration: underline; }
  .search-results .snippet { margin: 0.25rem 0 0; font-size: 0.85rem; opacity: 0.8; line-height: 1.4; }
  .search-results mark { background: #ffdd5766; color: inherit; border-radius: 3px; }
  .search-hint, .search-empty { margin: 0.4rem; font-size: 0.85rem; opacity: 0.7; }
</style>
</head>
<body hx-boost="true">
<header>
  <a href="/">Home</a>
  <nav class="breadcrumb">{{.Breadcrumb}}</nav>
  <div class="search-box">
    <input type="search" name="q" placeholder="Search…" autocomplete="off"
           hx-get="/search" hx-trigger="keyup changed delay:300ms, search"
           hx-target="#search-results" hx-swap="innerHTML">
    <div class="search-panel">
      <div id="search-results"></div>
    </div>
  </div>
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
