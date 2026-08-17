package main

import (
	"fmt"
	"html/template"
	"strings"
)

// firstHeading returns the text of the first "# Heading" line in a markdown
// file, used as the page title in search results.
func FirstHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	return ""
}

// snippetAround builds a short, escaped, <mark>-highlighted excerpt of raw
// markdown source around the first match. Returns "" if there was no
// content match (title-only hit).
func SnipperAround(content string, idx, matchLen int) template.HTML {
	if idx == -1 {
		return ""
	}
	const radius = 60
	start := idx - radius
	if start < 0 {
		start = 0
	}
	end := idx + matchLen + radius
	if end > len(content) {
		end = len(content)
	}

	before := template.HTMLEscapeString(content[start:idx])
	match := template.HTMLEscapeString(content[idx : idx+matchLen])
	after := template.HTMLEscapeString(content[idx+matchLen : end])

	prefix, suffix := "", ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(content) {
		suffix = "..."
	}
	return template.HTML(fmt.Sprintf(`<p class="snippet">%s%s<mark>%s</mark>%s%s</p>`, prefix, before, match, after, suffix))
}
