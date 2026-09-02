package ansi

import (
	"html"
	"strings"
)

// RenderContext holds the current rendering options and state.
type RenderContext struct {
	options Options

	blockStack *BlockStack
	table      *TableElement
}

// NewRenderContext returns a new RenderContext.
func NewRenderContext(options Options) RenderContext {
	return RenderContext{
		options:    options,
		blockStack: &BlockStack{},
		table:      &TableElement{},
	}
}

// SanitizeHTML sanitizes HTML content.
func (ctx RenderContext) SanitizeHTML(s string, trimSpaces bool) string {
	s = stripTags(s)
	if trimSpaces {
		s = strings.TrimSpace(s)
	}

	return html.UnescapeString(s)
}
