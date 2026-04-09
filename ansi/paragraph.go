package ansi

import (
	"bytes"
	"io"
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

// A ParagraphElement is used to render individual paragraphs.
type ParagraphElement struct {
	First bool
}

// Render renders a ParagraphElement.
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

	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, rules.BlockPrefix)
	renderText(bs.Current().Block, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.Prefix)
	return nil
}

// Finish finishes rendering a ParagraphElement.
func (e *ParagraphElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := bs.Current().Style

	mw := NewMarginWriter(ctx, w, rules)
	blk := bs.Current().Block.String()
	if len(strings.TrimSpace(blk)) > 0 {
		if !ctx.options.PreserveNewLines {
			blk = strings.ReplaceAll(strings.TrimSpace(blk), "\n", " ")
		}
		width := int(bs.Width(ctx)) //nolint: gosec
		if width > 0 {
			blk = xansi.Wrap(blk, width, "-")
		}

		_, err := io.WriteString(mw, blk)
		if err != nil {
			return err
		}
		_, _ = io.WriteString(mw, "\n")
	}

	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.Suffix)
	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, rules.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
