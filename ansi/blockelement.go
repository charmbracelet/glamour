package ansi

import (
	"bytes"
	"io"

	"github.com/charmbracelet/x/cellbuf"
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

func (e *BlockElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	bs.Push(*e)

	renderText(w, bs.Parent().Style.StylePrimitive, e.Style.BlockPrefix)
	renderText(bs.Current().Block, bs.Current().Style.StylePrimitive, e.Style.Prefix)
	return nil
}

func (e *BlockElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	if e.Margin {
		s := cellbuf.Wrap(
			bs.Current().Block.String(),
			int(bs.Width(ctx)), //nolint:gosec
			" ,.;-+|",
		)

		mw := NewMarginWriter(ctx, w, bs.Current().Style)
		if _, err := io.WriteString(mw, s); err != nil {
			return err
		}

		if e.Newline {
			if _, err := io.WriteString(mw, "\n"); err != nil {
				return err
			}
		}
	} else {
		_, err := bs.Parent().Block.Write(bs.Current().Block.Bytes())
		if err != nil {
			return err
		}
	}

	renderText(w, bs.Current().Style.StylePrimitive, e.Style.Suffix)
	renderText(w, bs.Parent().Style.StylePrimitive, e.Style.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
