package ansi

import (
	"fmt"
	"io"

	"charm.land/glamour/v2/internal/mermaid"
)

// A DiagramElement is used to render mermaid code blocks as box-drawing
// diagrams. Diagrams are laid out to the available width; anything the
// renderer cannot handle falls back to the source shown as a regular code
// block, preceded by a note explaining why.
type DiagramElement struct {
	Source   string
	Language string
}

// Render renders a DiagramElement.
func (e *DiagramElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	var indentation uint
	var margin uint
	rules := ctx.options.Styles.CodeBlock
	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	if rules.Margin != nil {
		margin = *rules.Margin
	}

	lines, err := safeRender(e.Source, availableWidth(ctx, int(indentation+margin)))
	if err != nil {
		return e.renderFallback(w, ctx, err)
	}

	iw := NewIndentWriter(w, int(indentation+margin), func(_ io.Writer) {
		_, _ = renderText(w, bs.Current().Style.StylePrimitive, " ")
	})
	defer iw.Close() //nolint:errcheck

	_, _ = renderText(iw, bs.Current().Style.StylePrimitive, rules.BlockPrefix)
	style := cascadeStyle(bs.Current().Style, rules.StyleBlock, false).StylePrimitive
	for _, line := range lines {
		if _, err := renderText(iw, style, line); err != nil {
			return fmt.Errorf("glamour: error rendering diagram: %w", err)
		}
		if _, err := io.WriteString(iw, "\n"); err != nil {
			return fmt.Errorf("glamour: error rendering diagram: %w", err)
		}
	}
	_, _ = renderText(iw, bs.Current().Style.StylePrimitive, rules.BlockSuffix)
	return nil
}

// renderFallback writes a note carrying the reason the diagram could not be
// rendered, followed by the source as a regular code block.
func (e *DiagramElement) renderFallback(w io.Writer, ctx RenderContext, cause error) error {
	bs := ctx.blockStack
	rules := ctx.options.Styles.CodeBlock

	var indentation uint
	var margin uint
	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	if rules.Margin != nil {
		margin = *rules.Margin
	}

	iw := NewIndentWriter(w, int(indentation+margin), func(_ io.Writer) {
		_, _ = renderText(w, bs.Current().Style.StylePrimitive, " ")
	})
	note := "mermaid: " + cause.Error() + " (showing source)"
	if _, err := renderText(iw, rules.StylePrimitive, note); err != nil {
		return fmt.Errorf("glamour: error rendering diagram note: %w", err)
	}
	if _, err := io.WriteString(iw, "\n"); err != nil {
		return fmt.Errorf("glamour: error rendering diagram note: %w", err)
	}
	if err := iw.Close(); err != nil {
		return fmt.Errorf("glamour: error rendering diagram note: %w", err)
	}

	return (&CodeBlockElement{Code: e.Source, Language: e.Language}).Render(w, ctx)
}

// isMermaidEnabled reports whether mermaid rendering is switched on. It is
// on unless explicitly disabled.
func isMermaidEnabled(o Options) bool {
	return o.Mermaid == nil || *o.Mermaid
}

// availableWidth returns the number of columns a diagram may occupy:
// the wrapping width minus the block stack's indentation and margins, and
// the code block style's own indent and margin. Zero means no wrapping.
func availableWidth(ctx RenderContext, indent int) int {
	w := int(ctx.blockStack.Width(ctx))
	if w <= 0 {
		return 0
	}
	return max(0, w-indent)
}

// safeRender guards against panics in the diagram renderer, which handles
// untrusted input.
func safeRender(src string, limit int) (lines []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			lines, err = nil, fmt.Errorf("internal error: %v", r)
		}
	}()
	return mermaid.Render(src, limit)
}
