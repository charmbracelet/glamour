package ansi

import "io"

// A CodeSpanElement is used to render codespan.
type CodeSpanElement struct {
	Text  string
	Style StylePrimitive
}

// Render renders a CodeSpanElement.
func (e *CodeSpanElement) Render(w io.Writer, _ RenderContext) error {
	_, _ = renderText(w, e.Style, e.Style.Prefix+e.Text+e.Style.Suffix)
	return nil
}

// StyleOverrideRender renders a CodeSpanElement with an overridden style. It is
// called when the code span is nested inside a styled element (emphasis,
// strong, link, table cell, ...) so that the surrounding style (e.g. italic or
// bold) cascades onto the inline code instead of being dropped.
func (e *CodeSpanElement) StyleOverrideRender(w io.Writer, _ RenderContext, style StylePrimitive) error {
	st := cascadeStylePrimitives(e.Style, style)
	_, _ = renderText(w, st, e.Style.Prefix+e.Text+e.Style.Suffix)
	return nil
}
