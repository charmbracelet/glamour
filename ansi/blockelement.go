package ansi

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// BlockElement provides a render buffer for children of a block element.
// After all children have been rendered into it, it applies indentation and
// margins around them and writes everything to the parent rendering buffer.
type BlockElement struct {
	Block   *bytes.Buffer
	Style   StyleBlock
	Margin  bool
	Newline bool
}

// Render renders a BlockElement.
func (e *BlockElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	bs.Push(*e)

	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, e.Style.BlockPrefix)
	renderText(bs.Current().Block, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, e.Style.Prefix)
	return nil
}

// Finish finishes rendering a BlockElement.
func (e *BlockElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	if e.Margin { //nolint: nestif
		s := ansi.Wordwrap(
			bs.Current().Block.String(),
			int(bs.Width(ctx)), //nolint: gosec
			" ,.;-+|",
		)

		mw := NewMarginWriter(ctx, w, bs.Current().Style)

		// Only replace image placeholders at the document level (when writing
		// to the final output). At intermediate levels, placeholders pass through
		// to avoid wordwrap corrupting the image escape sequences.
		if ctx.HasImages() && bs.Len() == 1 {
			// Calculate the margin prefix for images (same as what MarginWriter uses)
			rules := bs.Current().Style
			var indentation, margin uint
			if rules.Indent != nil {
				indentation = *rules.Indent
			}
			if rules.Margin != nil {
				margin = *rules.Margin
			}
			marginPrefix := strings.Repeat(" ", int(indentation+margin))
			if err := ctx.WriteWithImageReplacement(w, mw, s, marginPrefix); err != nil {
				return fmt.Errorf("glamour: error writing to writer: %w", err)
			}
		} else {
			if _, err := io.WriteString(mw, s); err != nil {
				return fmt.Errorf("glamour: error writing to writer: %w", err)
			}
		}

		if e.Newline {
			if _, err := io.WriteString(mw, "\n"); err != nil {
				return fmt.Errorf("glamour: error writing to writer: %w", err)
			}
		}
	} else {
		_, err := bs.Parent().Block.Write(bs.Current().Block.Bytes())
		if err != nil {
			return fmt.Errorf("glamour: error writing to writer: %w", err)
		}
	}

	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, e.Style.Suffix)
	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, e.Style.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
