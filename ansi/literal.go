package ansi

import "io"

// literalElement keeps the shared styling path without interpreting Markdown escapes.
type literalElement BaseElement

// Render renders literal content.
func (e *literalElement) Render(w io.Writer, ctx RenderContext) error {
	return (*BaseElement)(e).render(w, ctx, true)
}

// StyleOverrideRender renders literal content with an overridden style.
func (e *literalElement) StyleOverrideRender(w io.Writer, ctx RenderContext, style StylePrimitive) error {
	return (*BaseElement)(e).styleOverrideRender(w, ctx, style, true)
}
