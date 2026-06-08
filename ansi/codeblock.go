package ansi

import (
	"fmt"
	"io"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
)

const (
	// The chroma style theme name used for rendering.
	chromaStyleTheme = "charm"

	// The chroma formatter name used for rendering.
	chromaFormatter = "terminal256"
)

// mutex for synchronizing access to the chroma style registry.
// Related https://github.com/alecthomas/chroma/pull/650
var (
	mutex            = sync.Mutex{}
	lastChromaColors *Chroma
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

// registerChromaStyle registers the chroma style with the given colors.
// This is thread-safe and should be called before rendering when switching themes.
func registerChromaStyle(colors *Chroma) {
	if colors == nil {
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	styles.Register(chroma.MustNewStyle(chromaStyleTheme,
		chroma.StyleEntries{
			chroma.Text:                chromaStyle(colors.Text),
			chroma.Error:               chromaStyle(colors.Error),
			chroma.Comment:             chromaStyle(colors.Comment),
			chroma.CommentPreproc:      chromaStyle(colors.CommentPreproc),
			chroma.Keyword:             chromaStyle(colors.Keyword),
			chroma.KeywordReserved:     chromaStyle(colors.KeywordReserved),
			chroma.KeywordNamespace:    chromaStyle(colors.KeywordNamespace),
			chroma.KeywordType:         chromaStyle(colors.KeywordType),
			chroma.Operator:            chromaStyle(colors.Operator),
			chroma.Punctuation:         chromaStyle(colors.Punctuation),
			chroma.Name:                chromaStyle(colors.Name),
			chroma.NameBuiltin:         chromaStyle(colors.NameBuiltin),
			chroma.NameTag:             chromaStyle(colors.NameTag),
			chroma.NameAttribute:       chromaStyle(colors.NameAttribute),
			chroma.NameClass:           chromaStyle(colors.NameClass),
			chroma.NameConstant:        chromaStyle(colors.NameConstant),
			chroma.NameDecorator:       chromaStyle(colors.NameDecorator),
			chroma.NameException:       chromaStyle(colors.NameException),
			chroma.NameFunction:        chromaStyle(colors.NameFunction),
			chroma.NameOther:           chromaStyle(colors.NameOther),
			chroma.Literal:             chromaStyle(colors.Literal),
			chroma.LiteralNumber:       chromaStyle(colors.LiteralNumber),
			chroma.LiteralDate:         chromaStyle(colors.LiteralDate),
			chroma.LiteralString:       chromaStyle(colors.LiteralString),
			chroma.LiteralStringEscape: chromaStyle(colors.LiteralStringEscape),
			chroma.GenericDeleted:      chromaStyle(colors.GenericDeleted),
			chroma.GenericEmph:         chromaStyle(colors.GenericEmph),
			chroma.GenericInserted:     chromaStyle(colors.GenericInserted),
			chroma.GenericStrong:       chromaStyle(colors.GenericStrong),
			chroma.GenericSubheading:   chromaStyle(colors.GenericSubheading),
			chroma.Background:          chromaStyle(colors.Background),
		}))
	lastChromaColors = colors
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
	theme := rules.Theme

	if rules.Chroma != nil {
		theme = chromaStyleTheme
		if rules.Chroma != lastChromaColors {
			registerChromaStyle(rules.Chroma)
		}
	}

	iw := NewIndentWriter(w, int(indentation+margin), func(_ io.Writer) { //nolint:gosec
		_, _ = renderText(w, bs.Current().Style.StylePrimitive, " ")
	})
	defer iw.Close() //nolint:errcheck

	if len(theme) > 0 {
		_, _ = renderText(iw, bs.Current().Style.StylePrimitive, rules.BlockPrefix)

		err := quick.Highlight(iw, e.Code, e.Language, formatter, theme)
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
