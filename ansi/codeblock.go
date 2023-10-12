package ansi

import (
	"io"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/scrapbook"
	"github.com/muesli/termenv"
)

const (
	// The chroma style theme name used for rendering.
	chromaStyleTheme = "charm"
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

func chromaStyle(style scrapbook.StylePrimitive) string {
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
	var margin int
	rules := ctx.options.Styles.CodeBlock
	theme := rules.Theme
	if rules.Margin != nil {
		margin = int(*rules.Margin)
	}

	if rules.Chroma != nil && ctx.options.ColorProfile != termenv.Ascii {
		theme = chromaStyleTheme
		mutex.Lock()
		// Don't register the style if it's already registered.
		_, ok := styles.Registry[theme]
		if !ok {
			styles.Register(chroma.MustNewStyle(theme,
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
				}))
		}
		mutex.Unlock()
	}

	var tmp strings.Builder
	// TODO let's use a stringwriter with lipgloss.Join to set margins then write that to buffer
	if len(theme) > 0 {
		// TODO no longer handling prefix and suffix...
		err := quick.Highlight(&tmp, e.Code, e.Language, "terminal256", theme)
		// TODO can this be cleaned up?
		_, _ = w.Write([]byte(lipgloss.NewStyle().Margin(margin).Render(tmp.String())))
		return err
	}

	// fallback rendering
	// TODO fix this styling
	el := &BaseElement{
		Token: e.Code,
		Style: ctx.options.Styles.Code.StylePrimitive,
	}

	return el.Render(w, ctx)
}
