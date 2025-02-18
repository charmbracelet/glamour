package ansi

import (
	"bytes"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// A ParagraphElement is used to render individual paragraphs.
type ParagraphElement struct {
	First bool
}

func (e *ParagraphElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := ctx.options.Styles.Paragraph

	if !e.First {
		_, _ = io.WriteString(w, "\n")
	}
	be := BlockElement{
		Block: &bytes.Buffer{},
		Style: cascadeStyle(bs.Current().Style, rules, false),
	}
	bs.Push(be)

	renderText(w, bs.Parent().Style.StylePrimitive, rules.BlockPrefix)
	renderText(bs.Current().Block, bs.Current().Style.StylePrimitive, rules.Prefix)
	return nil
}

func (e *ParagraphElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := bs.Current().Style

	mw := NewMarginWriter(ctx, w, rules)
	block := bs.Current().Block.String()
	if len(strings.TrimSpace(block)) > 0 {
		if !ctx.options.PreserveNewLines {
			block = strings.ReplaceAll(block, "\n", " ")
		}
		_, err := mw.Write([]byte(ansi.Wordwrap(block, int(bs.Width(ctx)), "")))
		if err != nil {
			return err
		}
		_, _ = io.WriteString(mw, "\n")
	}

	renderText(w, bs.Current().Style.StylePrimitive, rules.Suffix)
	renderText(w, bs.Parent().Style.StylePrimitive, rules.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
