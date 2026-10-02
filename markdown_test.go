package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteDestination(t *testing.T) {
	tests := []struct {
		name string
		dest []byte
		want []byte
	}{
		{name: "simple markdown file", dest: []byte("./guide.md"), want: []byte("./guide")},
		{name: "markdown file with query", dest: []byte("./guide.md?raw=1"), want: []byte("./guide?raw=1")},
		{name: "markdown file with fragment", dest: []byte("./guide.md#intro"), want: []byte("./guide#intro")},
		{name: "absolute url", dest: []byte("https://example.com/docs/page.md"), want: []byte("https://example.com/docs/page.md")},
		{name: "plain file", dest: []byte("./image.png"), want: []byte("./image.png")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewriteDestination(tt.dest); string(got) != string(tt.want) {
				t.Fatalf("rewriteDestination(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

func TestDirectoryIndexRedirect(t *testing.T) {
	root := t.TempDir()
	indexPath := filepath.Join(root, "guide", "topic", "index.md")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, []byte("# Topic"), 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/guide/topic?from=breadcrumb", nil)
	res := httptest.NewRecorder()
	makeHandler(root)(res, req)

	if res.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusPermanentRedirect)
	}
	if got, want := res.Header().Get("Location"), "/guide/topic/index?from=breadcrumb"; got != want {
		t.Fatalf("Location = %q, want %q", got, want)
	}
}

func TestBreadcrumbHTMLUsesDirectoryIndexRoutes(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"guide/index.md", "guide/topic/index.md"} {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("# Page"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got := string(breadcrumbHTML(root, "/guide/topic/index"))
	for _, want := range []string{`href="/guide/index">guide`, `href="/guide/topic/index">topic`} {
		if !strings.Contains(got, want) {
			t.Errorf("breadcrumb %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, ">index</a>") {
		t.Errorf("breadcrumb should hide the index segment: %q", got)
	}
}

func TestTitleForDirectoryIndex(t *testing.T) {
	if got, want := titleFor("/guide/topic/index"), "Topic"; got != want {
		t.Fatalf("titleFor() = %q, want %q", got, want)
	}
}

func TestSearchHandlerResponseMarkup(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "guide.md"), []byte("# Go Guide\nSearchable text."), 0644); err != nil {
		t.Fatal(err)
	}
	handler := searchHandler(root)

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "empty query hint", want: `<p class="search-hint">`},
		{name: "matching result", query: "searchable", want: `<ul class="search-results"><li><a href="/guide" hx-boost="false">Go Guide</a>`},
		{name: "empty result", query: "missing", want: `<p class="search-empty">No results for &quot;missing&quot;</p>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/search?q="+tt.query, nil)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if !strings.Contains(res.Body.String(), tt.want) {
				t.Fatalf("response body %q does not contain %q", res.Body.String(), tt.want)
			}
		})
	}
}

func TestSearchResultForDirectoryIndexUsesCanonicalRoute(t *testing.T) {
	root := t.TempDir()
	indexPath := filepath.Join(root, "guide", "index.md")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, []byte("# Guide\nCanonicalIndexNeedle"), 0644); err != nil {
		t.Fatal(err)
	}

	searchReq := httptest.NewRequest(http.MethodGet, "/search?q=CanonicalIndexNeedle", nil)
	searchRes := httptest.NewRecorder()
	searchHandler(root).ServeHTTP(searchRes, searchReq)
	if !strings.Contains(searchRes.Body.String(), `href="/guide/index" hx-boost="false"`) {
		t.Fatalf("search response does not link to canonical route: %q", searchRes.Body.String())
	}

	pageReq := httptest.NewRequest(http.MethodGet, "/guide/index", nil)
	pageRes := httptest.NewRecorder()
	makeHandler(root)(pageRes, pageReq)
	if pageRes.Code != http.StatusOK {
		t.Fatalf("canonical route status = %d, want %d", pageRes.Code, http.StatusOK)
	}
}
