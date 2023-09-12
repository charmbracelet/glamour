package ansi

import (
	"bytes"
	"io"

	"github.com/charmbracelet/lipgloss"
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

	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, e.Style.BlockPrefix)
	renderText(bs.Current().Block, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, e.Style.Prefix)
	return nil
}

func (e *BlockElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	// TODO clean this
	/*
	if e.Margin {
		mw := NewMarginWriter(ctx, w, bs.Current().Style)
		_, err := mw.Write(
			wordwrap.Bytes(bs.Current().Block.Bytes(), int(bs.Width(ctx))))
		if err != nil {
			return err
		}

		if e.Newline {
			// _, err = mw.Write([]byte("\n"))
			if err != nil {
				return err
			}
		}
	} else {
		_, err := bs.Parent().Block.Write(bs.Current().Block.Bytes())
		if err != nil {
			return err
		}
	}
*/

	// TODO add margin
	flow := lipgloss.NewStyle().Width(int(bs.Width(ctx)))
	_, err := w.Write([]byte(flow.Render(bs.Current().Block.String())))
	if err != nil {
		return err
	}
	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, e.Style.Suffix)
	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, e.Style.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
