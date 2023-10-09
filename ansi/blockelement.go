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

	// TODO handle prefix for blocks
	return nil
}

func (e *BlockElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	blockBuffer := bs.Current().Block

	_, err := w.Write([]byte(e.Style.Style().Render(blockBuffer.String())))
	if err != nil {
		return err
	}

	// TODO handle suffix for blocks

	blockBuffer.Reset()
	bs.Pop()
	return nil
}
