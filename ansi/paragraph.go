package ansi

import (
	"bytes"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// A ParagraphElement is used to render individual paragraphs.
type ParagraphElement struct {
	First bool
}

func (e *ParagraphElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := ctx.options.Styles.Paragraph

	if !e.First {
		_, _ = w.Write([]byte("\n"))
	}
	be := BlockElement{
		Block: &bytes.Buffer{},
		Style: cascadeStyle(bs.Current().Style, rules, false),
	}
	bs.Push(be)

	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, rules.BlockPrefix)
	renderText(bs.Current().Block, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.Prefix)
	return nil
}

func (e *ParagraphElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := bs.Current().Style

	// TODO clean this up
	if len(strings.TrimSpace(bs.Current().Block.String())) > 0 {
		// flow := wordwrap.NewWriter(int(bs.Width(ctx)))
		// flow.KeepNewlines = ctx.options.PreserveNewLines
		flow := lipgloss.NewStyle().Width(int(bs.Width(ctx)))
		// panic(ctx.options.WordWrap)
		//		_, _ = flow.Write(bs.Current().Block.Bytes())
		//		flow.Close()

		//panic(flow.Render(bs.Current().Block.String()))

		_, err := w.Write([]byte(flow.Render(bs.Current().Block.String())))
		if err != nil {
			return err
		}
		_, _ = w.Write([]byte("\n"))

		//		_, err := mw.Write([]byte(flow.Render(bs.Current().Block.String()))))
		//		if err != nil {
		//			return err
		//		}
		//		_, _ = mw.Write([]byte("\n"))

	}

	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.Suffix)
	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, rules.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
