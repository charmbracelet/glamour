//go:build !glamour_nochroma

package ansi

import (
	"fmt"
	"io"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
)

// The chroma style theme name used for rendering.
const chromaStyleTheme = "charm"

// mutex for synchronizing access to the chroma style registry.
// Related https://github.com/alecthomas/chroma/pull/650
var mutex = sync.Mutex{}

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

// codeBlockTheme returns the chroma theme used to highlight code blocks,
// registering the style's custom chroma theme on first use. An empty theme
// renders code blocks without highlighting.
func codeBlockTheme(rules StyleCodeBlock) string {
	if rules.Chroma == nil {
		return rules.Theme
	}

	theme := chromaStyleTheme
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

	return theme
}

// highlightCode writes syntax-highlighted code to w.
func highlightCode(w io.Writer, code, language, formatter, theme string) error {
	if err := quick.Highlight(w, code, language, formatter, theme); err != nil {
		return fmt.Errorf("glamour: error highlighting code: %w", err)
	}
	return nil
}
