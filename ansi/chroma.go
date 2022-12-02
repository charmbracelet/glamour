package ansi

import (
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

var (
	// mutex for synchronizing access to the chroma style registry.
	// Related https://github.com/alecthomas/chroma/pull/650
	mutex = sync.Mutex{}
)

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

// ChromaRegister registers a chroma style from a glamour style.
func ChromaRegister(cfg *StyleConfig) *chroma.Style {
	var style *chroma.Style
	rules := cfg.CodeBlock
	theme := rules.Theme
	if theme == "" && rules.Chroma == nil {
		return nil
	}

	mutex.Lock()
	// Register the theme if it doesn't already exist.
	_, ok := styles.Registry[theme]
	if !ok && rules.Chroma != nil {
		style = styles.Register(
			chroma.MustNewStyle(theme,
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
				},
			),
		)
	}
	mutex.Unlock()

	return style
}
