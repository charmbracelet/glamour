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

	// SkipWordwrap, when true, causes Finish to skip the lipgloss.Wrap call
	// and only apply the MarginWriter (indent + padding). This is needed for
	// elements like blockquotes whose children (paragraphs) already perform
	// word-wrapping. Running Wrap a second time on already-wrapped content
	// causes incorrect re-wrapping when the indent token's visual width
	// differs from the indent count, leading to lost indent tokens on
	// continuation lines.
	SkipWordwrap bool
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
		var s string
		if e.SkipWordwrap {
			// Children (e.g. paragraphs) already word-wrapped the content.
			// Only apply indent + padding, do NOT re-wrap.
			s = bs.Current().Block.String()
		} else {
			s = lipgloss.Wrap(
				bs.Current().Block.String(),
				int(bs.Width(ctx)), //nolint: gosec
				" ,.;-+|",
			)
		}

		mw := NewMarginWriter(ctx, w, bs.Current().Style)
		defer mw.Close() //nolint:errcheck
		if _, err := io.WriteString(mw, s); err != nil {
			return fmt.Errorf("glamour: error writing to writer: %w", err)
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

	_, _ = renderText(w, bs.Current().Style.StylePrimitive, e.Style.Suffix)
	_, _ = renderText(w, bs.Parent().Style.StylePrimitive, e.Style.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
