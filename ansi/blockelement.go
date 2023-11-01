package ansi

import (
	"bytes"
	"io"

	"github.com/charmbracelet/scrapbook"
)

// BlockElement provides a render buffer for children of a block element.
// After all children have been rendered into it, it applies indentation and
// margins around them and writes everything to the parent rendering buffer.
type BlockElement struct {
	Block   *bytes.Buffer
	Style   scrapbook.StyleBlock
	Margin  bool
	Newline bool
}

func (e *BlockElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	bs.Push(*e)
	return nil
}

// setMargins sets the margins given the prefix and suffix values being non-nil.

func (e *BlockElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	blockBuffer := bs.Current().Block
	ls := e.Style.Style().Inherit(e.Style.StylePrimitive.Style()).Width(ctx.options.WordWrap)

	_, err := w.Write([]byte(ls.Render(blockBuffer.String())))
	if err != nil {
		return err
	}

	if e.Newline {
		_, err := w.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	blockBuffer.Reset()
	bs.Pop()
	return nil
}
