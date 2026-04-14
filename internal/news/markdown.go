package news

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
)

// RenderMarkdown converts Markdown text to safe HTML.
// On render error falls back to HTML-escaped plain text.
func RenderMarkdown(content string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(content)) //nolint:gosec
	}
	return template.HTML(buf.String()) //nolint:gosec
}
