package ansi

import (
	"io"
	"sync"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/termenv"
)

var (
	// mutex for synchronizing access to the chroma style registry.
	// Related https://github.com/alecthomas/chroma/pull/650
	mutex = sync.Mutex{}
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
	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	if rules.Margin != nil {
		margin = *rules.Margin
	}

	// Don't register the style if it's already registered.
	if rules.Chroma != nil && ctx.options.ColorProfile != termenv.Ascii {
		mutex.Lock()
		ChromaRegister(&ctx.options.Styles)
		mutex.Unlock()
	}
	theme := rules.Theme

	iw := indent.NewWriterPipe(w, indentation+margin, func(wr io.Writer) {
		renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, " ")
	})

	if len(theme) > 0 {
		renderText(iw, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockPrefix)
		err := quick.Highlight(iw, e.Code, e.Language, "terminal256", theme)
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
