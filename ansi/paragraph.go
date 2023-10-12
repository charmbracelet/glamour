package ansi

import (
	"bytes"
	"io"
)

// A ParagraphElement is used to render individual paragraphs.
type ParagraphElement struct {
	First bool
}

func (e *ParagraphElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	if !e.First {
		_, _ = w.Write([]byte("\n"))
	}
	be := BlockElement{
		Block: &bytes.Buffer{},
		Style: ctx.options.Styles.Paragraph,
	}
	bs.Push(be)

	// TODO handle paragraph prefix
	return nil
}

func (e *ParagraphElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	renderText(w, bs.Current().Style, bs.Current().Block.String())
	// TODO render suffix?

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
