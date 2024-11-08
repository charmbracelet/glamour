package ansi

import (
	"bytes"
	"io"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/muesli/termenv"
)

const (
	// The chroma style theme name used for rendering.
	chromaStyleTheme = "charm"
)

// mutex for synchronizing access to the chroma style registry.
// Related https://github.com/alecthomas/chroma/pull/650
var mutex = sync.Mutex{}

// A CodeBlockElement is used to render code blocks.
type CodeBlockElement struct {
	Code     string
	Language string
}

func chromaStyle(style StylePrimitive) string {
	var s string

	if style.Color != nil {
		s = *style.Color
	}
	if style.BackgroundColor != nil {
		if s != "" {
			s += " "
		}
		s += "bg:" + *style.BackgroundColor
	}
	if style.Italic != nil && *style.Italic {
		if s != "" {
			s += " "
		}
		s += "italic"
	}
	if style.Bold != nil && *style.Bold {
		if s != "" {
			s += " "
		}
		s += "bold"
	}
	if style.Underline != nil && *style.Underline {
		if s != "" {
			s += " "
		}
		s += "underline"
	}

	return s
}

func (e *CodeBlockElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := ctx.options.Styles.CodeBlock

	be := BlockElement{
		Block: &bytes.Buffer{},
		Style: rules.StyleBlock,
	}
	bs.Push(be)
	return nil
}

func (e *CodeBlockElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := bs.Current().Style

	cb := ctx.options.Styles.CodeBlock
	theme := cb.Theme
	chromaRules := cb.Chroma

	if chromaRules != nil && ctx.options.ColorProfile != termenv.Ascii {
		theme = chromaStyleTheme
		mutex.Lock()
		// Don't register the style if it's already registered.
		_, ok := styles.Registry[theme]
		if !ok {
			styles.Register(chroma.MustNewStyle(theme,
				chroma.StyleEntries{
					chroma.Text:                chromaStyle(chromaRules.Text),
					chroma.Error:               chromaStyle(chromaRules.Error),
					chroma.Comment:             chromaStyle(chromaRules.Comment),
					chroma.CommentPreproc:      chromaStyle(chromaRules.CommentPreproc),
					chroma.Keyword:             chromaStyle(chromaRules.Keyword),
					chroma.KeywordReserved:     chromaStyle(chromaRules.KeywordReserved),
					chroma.KeywordNamespace:    chromaStyle(chromaRules.KeywordNamespace),
					chroma.KeywordType:         chromaStyle(chromaRules.KeywordType),
					chroma.Operator:            chromaStyle(chromaRules.Operator),
					chroma.Punctuation:         chromaStyle(chromaRules.Punctuation),
					chroma.Name:                chromaStyle(chromaRules.Name),
					chroma.NameBuiltin:         chromaStyle(chromaRules.NameBuiltin),
					chroma.NameTag:             chromaStyle(chromaRules.NameTag),
					chroma.NameAttribute:       chromaStyle(chromaRules.NameAttribute),
					chroma.NameClass:           chromaStyle(chromaRules.NameClass),
					chroma.NameConstant:        chromaStyle(chromaRules.NameConstant),
					chroma.NameDecorator:       chromaStyle(chromaRules.NameDecorator),
					chroma.NameException:       chromaStyle(chromaRules.NameException),
					chroma.NameFunction:        chromaStyle(chromaRules.NameFunction),
					chroma.NameOther:           chromaStyle(chromaRules.NameOther),
					chroma.Literal:             chromaStyle(chromaRules.Literal),
					chroma.LiteralNumber:       chromaStyle(chromaRules.LiteralNumber),
					chroma.LiteralDate:         chromaStyle(chromaRules.LiteralDate),
					chroma.LiteralString:       chromaStyle(chromaRules.LiteralString),
					chroma.LiteralStringEscape: chromaStyle(chromaRules.LiteralStringEscape),
					chroma.GenericDeleted:      chromaStyle(chromaRules.GenericDeleted),
					chroma.GenericEmph:         chromaStyle(chromaRules.GenericEmph),
					chroma.GenericInserted:     chromaStyle(chromaRules.GenericInserted),
					chroma.GenericStrong:       chromaStyle(chromaRules.GenericStrong),
					chroma.GenericSubheading:   chromaStyle(chromaRules.GenericSubheading),
					chroma.Background:          chromaStyle(chromaRules.Background),
				}))
		}
		mutex.Unlock()
	}

	mw := NewMarginWriter(ctx, w, bs.Current().Style, false)
	renderText(mw, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockPrefix)
	if len(theme) > 0 {
		err := quick.Highlight(mw, e.Code, e.Language, "terminal256", theme)
		if err != nil {
			return err
		}
	} else {
		// fallback rendering
		el := &BaseElement{
			Token: e.Code,
			Style: rules.StylePrimitive,
		}

		err := el.Render(mw, ctx)
		if err != nil {
			return err
		}
	}
	renderText(mw, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
