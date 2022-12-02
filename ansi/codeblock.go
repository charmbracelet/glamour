package ansi

import (
	"io"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/termenv"
)

// A CodeBlockElement is used to render code blocks.
type CodeBlockElement struct {
	Code     string
	Language string
}

func (e *CodeBlockElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	var indentation uint
	var margin uint
	rules := ctx.options.Styles.CodeBlock
	theme := rules.Theme

	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	if rules.Margin != nil {
		margin = *rules.Margin
	}

	iw := indent.NewWriterPipe(w, indentation+margin, func(wr io.Writer) {
		renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, " ")
	})

	if len(theme) > 0 {
		var formatter string
		switch ctx.options.ColorProfile {
		case termenv.TrueColor, termenv.ANSI256:
			formatter = "terminal256"
		case termenv.ANSI:
			formatter = "terminal16"
		default:
			formatter = "terminal"
		}
		renderText(iw, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockPrefix)
		err := quick.Highlight(iw, e.Code, e.Language, formatter, theme)
		if err != nil {
			return err
		}
		renderText(iw, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockSuffix)
		return nil
	}

	// fallback rendering
	el := &BaseElement{
		Token: e.Code,
		Style: rules.StylePrimitive,
	}

	return el.Render(iw, ctx)
}
