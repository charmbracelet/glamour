package ansi

import (
	"fmt"
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// The chroma formatter name used for rendering.
const chromaFormatter = "terminal256"

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

// newChromaStyle builds a chroma style from the given chroma rules.
func newChromaStyle(c *Chroma) *chroma.Style {
	return chroma.MustNewStyle("charm",
		chroma.StyleEntries{
			chroma.Text:                chromaStyle(c.Text),
			chroma.Error:               chromaStyle(c.Error),
			chroma.Comment:             chromaStyle(c.Comment),
			chroma.CommentPreproc:      chromaStyle(c.CommentPreproc),
			chroma.Keyword:             chromaStyle(c.Keyword),
			chroma.KeywordReserved:     chromaStyle(c.KeywordReserved),
			chroma.KeywordNamespace:    chromaStyle(c.KeywordNamespace),
			chroma.KeywordType:         chromaStyle(c.KeywordType),
			chroma.Operator:            chromaStyle(c.Operator),
			chroma.Punctuation:         chromaStyle(c.Punctuation),
			chroma.Name:                chromaStyle(c.Name),
			chroma.NameBuiltin:         chromaStyle(c.NameBuiltin),
			chroma.NameTag:             chromaStyle(c.NameTag),
			chroma.NameAttribute:       chromaStyle(c.NameAttribute),
			chroma.NameClass:           chromaStyle(c.NameClass),
			chroma.NameConstant:        chromaStyle(c.NameConstant),
			chroma.NameDecorator:       chromaStyle(c.NameDecorator),
			chroma.NameException:       chromaStyle(c.NameException),
			chroma.NameFunction:        chromaStyle(c.NameFunction),
			chroma.NameOther:           chromaStyle(c.NameOther),
			chroma.Literal:             chromaStyle(c.Literal),
			chroma.LiteralNumber:       chromaStyle(c.LiteralNumber),
			chroma.LiteralDate:         chromaStyle(c.LiteralDate),
			chroma.LiteralString:       chromaStyle(c.LiteralString),
			chroma.LiteralStringEscape: chromaStyle(c.LiteralStringEscape),
			chroma.GenericDeleted:      chromaStyle(c.GenericDeleted),
			chroma.GenericEmph:         chromaStyle(c.GenericEmph),
			chroma.GenericInserted:     chromaStyle(c.GenericInserted),
			chroma.GenericStrong:       chromaStyle(c.GenericStrong),
			chroma.GenericSubheading:   chromaStyle(c.GenericSubheading),
			chroma.Background:          chromaStyle(c.Background),
		})
}

// highlight is like quick.Highlight, but takes the style directly instead of
// looking it up by name in chroma's global registry.
func highlight(w io.Writer, source, lexer, formatter string, style *chroma.Style) error {
	l := lexers.Get(lexer)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	f := formatters.Get(formatter)
	if f == nil {
		f = formatters.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err //nolint:wrapcheck
	}
	return f.Format(w, style, it) //nolint:wrapcheck
}

// Render renders a CodeBlockElement.
func (e *CodeBlockElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	var indentation uint
	var margin uint
	formatter := chromaFormatter
	rules := ctx.options.Styles.CodeBlock
	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	if rules.Margin != nil {
		margin = *rules.Margin
	}
	if len(ctx.options.ChromaFormatter) > 0 {
		formatter = ctx.options.ChromaFormatter
	}
	// Use the renderer's own style rather than registering it in chroma's
	// global registry, so renderers with different styles don't share one.
	var style *chroma.Style
	if ctx.chromaStyle != nil {
		style = ctx.chromaStyle()
	} else if len(rules.Theme) > 0 {
		style = styles.Get(rules.Theme)
	}

	iw := NewIndentWriter(w, int(indentation+margin), func(_ io.Writer) {
		_, _ = renderText(w, bs.Current().Style.StylePrimitive, " ")
	})
	defer iw.Close() //nolint:errcheck

	if style != nil {
		_, _ = renderText(iw, bs.Current().Style.StylePrimitive, rules.BlockPrefix)

		err := highlight(iw, e.Code, e.Language, formatter, style)
		if err != nil {
			return fmt.Errorf("glamour: error highlighting code: %w", err)
		}
		_, _ = renderText(iw, bs.Current().Style.StylePrimitive, rules.BlockSuffix)
		return nil
	}

	// fallback rendering
	el := &BaseElement{
		Token: e.Code,
		Style: rules.StylePrimitive,
	}

	return el.Render(iw, ctx)
}
