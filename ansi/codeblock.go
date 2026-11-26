package ansi

import (
	"fmt"
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

const (
	// The name given to the chroma style built from a style's own chroma
	// section. It is not registered, so it is not looked up by name.
	chromaStyleTheme = "charm"

	// The chroma formatter name used for rendering.
	chromaFormatter = "terminal256"
)

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

	var style *chroma.Style
	if rules.Chroma != nil {
		// Build the style for this render. Registering it under a fixed
		// name would keep the first style built in the process.
		var err error
		style, err = chroma.NewStyle(chromaStyleTheme,
			chroma.StyleEntries{
				chroma.Text:                chromaStyle(rules.Chroma.Text),
				chroma.Error:               chromaStyle(rules.Chroma.Error),
				chroma.Comment:             chromaStyle(rules.Chroma.Comment),
				chroma.CommentPreproc:      chromaStyle(rules.Chroma.CommentPreproc),
				chroma.Keyword:             chromaStyle(rules.Chroma.Keyword),
				chroma.KeywordReserved:     chromaStyle(rules.Chroma.KeywordReserved),
				chroma.KeywordNamespace:    chromaStyle(rules.Chroma.KeywordNamespace),
				chroma.KeywordType:         chromaStyle(rules.Chroma.KeywordType),
				chroma.Operator:            chromaStyle(rules.Chroma.Operator),
				chroma.Punctuation:         chromaStyle(rules.Chroma.Punctuation),
				chroma.Name:                chromaStyle(rules.Chroma.Name),
				chroma.NameBuiltin:         chromaStyle(rules.Chroma.NameBuiltin),
				chroma.NameTag:             chromaStyle(rules.Chroma.NameTag),
				chroma.NameAttribute:       chromaStyle(rules.Chroma.NameAttribute),
				chroma.NameClass:           chromaStyle(rules.Chroma.NameClass),
				chroma.NameConstant:        chromaStyle(rules.Chroma.NameConstant),
				chroma.NameDecorator:       chromaStyle(rules.Chroma.NameDecorator),
				chroma.NameException:       chromaStyle(rules.Chroma.NameException),
				chroma.NameFunction:        chromaStyle(rules.Chroma.NameFunction),
				chroma.NameOther:           chromaStyle(rules.Chroma.NameOther),
				chroma.Literal:             chromaStyle(rules.Chroma.Literal),
				chroma.LiteralNumber:       chromaStyle(rules.Chroma.LiteralNumber),
				chroma.LiteralDate:         chromaStyle(rules.Chroma.LiteralDate),
				chroma.LiteralString:       chromaStyle(rules.Chroma.LiteralString),
				chroma.LiteralStringEscape: chromaStyle(rules.Chroma.LiteralStringEscape),
				chroma.GenericDeleted:      chromaStyle(rules.Chroma.GenericDeleted),
				chroma.GenericEmph:         chromaStyle(rules.Chroma.GenericEmph),
				chroma.GenericInserted:     chromaStyle(rules.Chroma.GenericInserted),
				chroma.GenericStrong:       chromaStyle(rules.Chroma.GenericStrong),
				chroma.GenericSubheading:   chromaStyle(rules.Chroma.GenericSubheading),
				chroma.Background:          chromaStyle(rules.Chroma.Background),
			})
		if err != nil {
			return fmt.Errorf("glamour: error building chroma style: %w", err)
		}
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

// highlight writes code to w, colored with style. It chooses the lexer and
// the formatter as quick.Highlight does, but takes the style itself rather
// than the name of a registered one.
func highlight(w io.Writer, code, language, formatter string, style *chroma.Style) error {
	l := lexers.Get(language)
	if l == nil {
		l = lexers.Analyse(code)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	f := formatters.Get(formatter)
	if f == nil {
		f = formatters.Fallback
	}

	it, err := l.Tokenise(nil, code)
	if err != nil {
		return fmt.Errorf("glamour: error tokenizing code: %w", err)
	}
	if err := f.Format(w, style, it); err != nil {
		return fmt.Errorf("glamour: error formatting code: %w", err)
	}
	return nil
}
