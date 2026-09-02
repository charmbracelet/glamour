package ansi

import (
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

var htmlNumericReference = regexp.MustCompile(`&#(?:[xX][0-9A-Fa-f]+|[0-9]+);?`)

func protectHTMLControlReferences(s string) string {
	return htmlNumericReference.ReplaceAllStringFunc(s, func(reference string) string {
		if strings.ContainsFunc(html.UnescapeString(reference), func(r rune) bool {
			return r != '\n' && r != '\t' && unicode.IsControl(r)
		}) {
			return "&amp;" + reference[1:]
		}
		return reference
	})
}

func unescapeHTML(s string) string {
	return html.UnescapeString(protectHTMLControlReferences(s))
}

// RenderContext holds the current rendering options and state.
type RenderContext struct {
	options Options

	blockStack *BlockStack
	table      *TableElement

	stripper *bluemonday.Policy
}

// NewRenderContext returns a new RenderContext.
func NewRenderContext(options Options) RenderContext {
	return RenderContext{
		options:    options,
		blockStack: &BlockStack{},
		table:      &TableElement{},
		stripper:   bluemonday.StrictPolicy(),
	}
}

// SanitizeHTML sanitizes HTML content.
func (ctx RenderContext) SanitizeHTML(s string, trimSpaces bool) string {
	s = protectHTMLControlReferences(s)
	s = ctx.stripper.Sanitize(s)
	if trimSpaces {
		s = strings.TrimSpace(s)
	}

	return unescapeHTML(s)
}
