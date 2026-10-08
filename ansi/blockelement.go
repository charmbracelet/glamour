package ansi

import (
	"bytes"
	"fmt"
	"io"

	"charm.land/lipgloss/v2"
)

// BlockElement provides a render buffer for children of a block element.
// After all children have been rendered into it, it applies indentation and
// margins around them and writes everything to the parent rendering buffer.
type BlockElement struct {
	Block   *bytes.Buffer
	Style   StyleBlock
	Margin  bool
	Newline bool
	// Prewrapped prevents structurally wrapped content from being reflowed.
	Prewrapped  bool
	PrefixWidth int
	// OuterIndent shifts a block without repeating its configured indent token.
	OuterIndent uint
}

// Render renders a BlockElement.
func (e *BlockElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	bs.Push(*e)

	_, _ = renderText(w, bs.Parent().Style.StylePrimitive, e.Style.BlockPrefix)
	_, _ = renderText(bs.Current().Block, bs.Current().Style.StylePrimitive, e.Style.Prefix)
	return nil
}

// Finish finishes rendering a BlockElement.
func (e *BlockElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	if e.Margin { //nolint: nestif
		s := bs.Current().Block.String()
		// Lists wrap their children while the current item prefix is still known.
		if !bs.Current().Prewrapped {
			s = lipgloss.Wrap(s, int(bs.Width(ctx)), " ,.;-+|")
		}

		ow := w
		// Apply structural offsets outside the block's configured indent token.
		if bs.Current().OuterIndent > 0 {
			iw := NewIndentWriter(w, int(bs.Current().OuterIndent), nil)
			defer iw.Close() //nolint:errcheck
			ow = iw
		}

		mw := NewMarginWriter(ctx, ow, bs.Current().Style)
		defer mw.Close() //nolint:errcheck
		if _, err := io.WriteString(mw, s); err != nil {
			return fmt.Errorf("glamour: error writing to writer: %w", err)
		}

		if bs.Current().Newline {
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

	_, _ = renderText(w, bs.Current().Style.StylePrimitive, e.Style.Suffix)
	_, _ = renderText(w, bs.Parent().Style.StylePrimitive, e.Style.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
