package ansi

import (
	"html"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// RenderContext holds the current rendering options and state.
type RenderContext struct {
	options Options

	blockStack *BlockStack
	table      *TableElement

	stripper *bluemonday.Policy

	// htmlStyleStack tracks active inline HTML style tags (<b>, <i>, <u>, etc.)
	htmlStyleStack *[]StylePrimitive
}

// NewRenderContext returns a new RenderContext.
func NewRenderContext(options Options) RenderContext {
	return RenderContext{
		options:        options,
		blockStack:     &BlockStack{},
		table:          &TableElement{},
		stripper:       bluemonday.StrictPolicy(),
		htmlStyleStack: &[]StylePrimitive{},
	}
}

// SanitizeHTML sanitizes HTML content.
func (ctx RenderContext) SanitizeHTML(s string, trimSpaces bool) string {
	s = ctx.stripper.Sanitize(s)
	if trimSpaces {
		s = strings.TrimSpace(s)
	}

	return html.UnescapeString(s)
}

// htmlInlineTagStyle returns a StylePrimitive for recognized inline HTML tags.
// Returns the style and true if the tag is recognized, false otherwise.
func htmlInlineTagStyle(tag string) (StylePrimitive, bool) {
	t := true
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "<b>", "<strong>":
		return StylePrimitive{Bold: &t}, true
	case "<i>", "<em>":
		return StylePrimitive{Italic: &t}, true
	case "<u>", "<ins>":
		return StylePrimitive{Underline: &t}, true
	case "<s>", "<del>", "<strike>":
		return StylePrimitive{CrossedOut: &t}, true
	}
	return StylePrimitive{}, false
}

// isHTMLClosingTag checks if a tag is a closing variant of an inline style tag.
func isHTMLClosingTag(tag string) bool {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "</b>", "</strong>", "</i>", "</em>", "</u>", "</ins>", "</s>", "</del>", "</strike>":
		return true
	}
	return false
}

// PushHTMLStyle adds an inline HTML style to the stack.
func (ctx RenderContext) PushHTMLStyle(s StylePrimitive) {
	*ctx.htmlStyleStack = append(*ctx.htmlStyleStack, s)
}

// PopHTMLStyle removes the last inline HTML style from the stack.
func (ctx RenderContext) PopHTMLStyle() {
	stack := *ctx.htmlStyleStack
	if len(stack) > 0 {
		*ctx.htmlStyleStack = stack[:len(stack)-1]
	}
}

// CurrentHTMLStyle returns the combined style of all active HTML inline tags.
func (ctx RenderContext) CurrentHTMLStyle() StylePrimitive {
	var combined StylePrimitive
	for _, s := range *ctx.htmlStyleStack {
		combined = cascadeStylePrimitive(combined, s, false)
	}
	return combined
}
