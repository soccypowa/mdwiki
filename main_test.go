package main

import "testing"

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
