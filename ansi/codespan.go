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

// StyleOverrideRender renders a CodeSpanElement with an overridden style,
// allowing code spans inside emphasis/strong to inherit italic/bold styling
// while keeping their own code-specific styling (color, background).
func (e *CodeSpanElement) StyleOverrideRender(w io.Writer, _ RenderContext, style StylePrimitive) error {
	combined := cascadeStylePrimitive(e.Style, style, false)
	// Preserve the code span's own color and background over the parent's
	if e.Style.Color != nil {
		combined.Color = e.Style.Color
	}
	if e.Style.BackgroundColor != nil {
		combined.BackgroundColor = e.Style.BackgroundColor
	}
	combined.Prefix = e.Style.Prefix
	combined.Suffix = e.Style.Suffix
	_, _ = renderText(w, combined, combined.Prefix+e.Text+combined.Suffix)
	return nil
}
