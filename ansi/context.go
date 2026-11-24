package ansi

import (
	"html"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/microcosm-cc/bluemonday"
)

// RenderContext holds the current rendering options and state.
type RenderContext struct {
	options Options

	blockStack *BlockStack
	table      *TableElement

	stripper *bluemonday.Policy

	// chromaStyle returns the code block style built from the stylesheet's
	// chroma section. It is nil when the stylesheet has no chroma section.
	chromaStyle func() *chroma.Style
}

// NewRenderContext returns a new RenderContext.
func NewRenderContext(options Options) RenderContext {
	ctx := RenderContext{
		options:    options,
		blockStack: &BlockStack{},
		table:      &TableElement{},
		stripper:   bluemonday.StrictPolicy(),
	}
	if c := options.Styles.CodeBlock.Chroma; c != nil {
		ctx.chromaStyle = sync.OnceValue(func() *chroma.Style {
			return newChromaStyle(c)
		})
	}

	return ctx
}

// SanitizeHTML sanitizes HTML content.
func (ctx RenderContext) SanitizeHTML(s string, trimSpaces bool) string {
	s = ctx.stripper.Sanitize(s)
	if trimSpaces {
		s = strings.TrimSpace(s)
	}

	return html.UnescapeString(s)
}
