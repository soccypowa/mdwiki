package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want cliOptions
	}{
		{
			name: "version flag",
			args: []string{"-version"},
			want: cliOptions{port: 8888, showVersion: true},
		},
		{
			name: "named flags",
			args: []string{"-dir", "docs", "-port", "9000"},
			want: cliOptions{dir: "docs", port: 9000},
		},
		{
			name: "positional arguments",
			args: []string{"docs", "9000"},
			want: cliOptions{dir: "docs", port: 9000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("parseFlags(%v) returned error: %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("parseFlags(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestParseFlagsReturnsError(t *testing.T) {
	if _, err := parseFlags([]string{"-unknown"}); err == nil {
		t.Fatal("parseFlags(-unknown) returned no error")
	}
}

func TestStaticFileSymlinkEscapeBlocked(t *testing.T) {
	root := t.TempDir()
	outsideDir := t.TempDir()
	secretPath := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte("SECRET"), 0644); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(root, "link.txt")
	if err := os.Symlink(secretPath, linkPath); err != nil {
		t.Skipf("symlinks not supported in this environment: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/link.txt", nil)
	res := httptest.NewRecorder()
	safeFileServer(root).ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
}
