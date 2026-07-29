package main

import "testing"

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
