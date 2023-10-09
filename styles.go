package glamour

//go:generate go run ./internal/generate-style-json

import (
	"github.com/charmbracelet/scrapbook"
)

const defaultListIndent = 2
const defaultMargin = 2

var (
	// ASCIIStyleConfig uses only ASCII characters.
	ASCIIStyleConfig = scrapbook.StyleConfig{
		Document: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
			},
			Margin: uintPtr(defaultMargin),
		},
		BlockQuote: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{},
			Indent:         uintPtr(1),
			IndentToken:    stringPtr("| "),
		},
		Paragraph: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{},
		},
		List: scrapbook.StyleList{
			StyleBlock: scrapbook.StyleBlock{
				StylePrimitive: scrapbook.StylePrimitive{},
			},
			LevelIndent: 4,
		},
		Heading: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockSuffix: "\n",
			},
		},
		H1: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "# ",
			},
		},
		H2: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "## ",
			},
		},
		H3: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "### ",
			},
		},
		H4: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "#### ",
			},
		},
		H5: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "##### ",
			},
		},
		H6: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "###### ",
			},
		},
		Strikethrough: scrapbook.StylePrimitive{
			BlockPrefix: "~~",
			BlockSuffix: "~~",
		},
		Emph: scrapbook.StylePrimitive{
			BlockPrefix: "*",
			BlockSuffix: "*",
		},
		Strong: scrapbook.StylePrimitive{
			BlockPrefix: "**",
			BlockSuffix: "**",
		},
		HorizontalRule: scrapbook.StylePrimitive{
			Format: "\n--------\n",
		},
		Item: scrapbook.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: scrapbook.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: scrapbook.StyleTask{
			Ticked:   "[x] ",
			Unticked: "[ ] ",
		},
		ImageText: scrapbook.StylePrimitive{
			Format: "Image: {{.text}} →",
		},
		Code: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockPrefix: "`",
				BlockSuffix: "`",
			},
		},
		CodeBlock: scrapbook.StyleCodeBlock{
			StyleBlock: scrapbook.StyleBlock{
				Margin: uintPtr(defaultMargin),
			},
		},
		Table: scrapbook.StyleTable{
			CenterSeparator: stringPtr("+"),
			ColumnSeparator: stringPtr("|"),
			RowSeparator:    stringPtr("-"),
		},
		DefinitionDescription: scrapbook.StylePrimitive{
			BlockPrefix: "\n* ",
		},
	}

	// DarkStyleConfig is the default dark style.
	DarkStyleConfig = scrapbook.StyleConfig{
		Document: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       stringPtr("252"),
			},
			Margin: uintPtr(defaultMargin),
		},
		BlockQuote: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{},
			Indent:         uintPtr(1),
			IndentToken:    stringPtr("│ "),
		},
		List: scrapbook.StyleList{
			LevelIndent: defaultListIndent,
		},
		Heading: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockSuffix: "\n",
				Color:       stringPtr("39"),
				Bold:        boolPtr(true),
			},
		},
		H1: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           stringPtr("228"),
				BackgroundColor: stringPtr("63"),
				Bold:            boolPtr(true),
			},
		},
		H2: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "## ",
			},
		},
		H3: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "### ",
			},
		},
		H4: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "#### ",
			},
		},
		H5: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "##### ",
			},
		},
		H6: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "###### ",
				Color:  stringPtr("35"),
				Bold:   boolPtr(false),
			},
		},
		Strikethrough: scrapbook.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: scrapbook.StylePrimitive{
			Italic: boolPtr(true),
		},
		Strong: scrapbook.StylePrimitive{
			Bold: boolPtr(true),
		},
		HorizontalRule: scrapbook.StylePrimitive{
			Color:  stringPtr("240"),
			Format: "\n--------\n",
		},
		Item: scrapbook.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: scrapbook.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: scrapbook.StyleTask{
			StylePrimitive: scrapbook.StylePrimitive{},
			Ticked:         "[✓] ",
			Unticked:       "[ ] ",
		},
		Link: scrapbook.StylePrimitive{
			Color:     stringPtr("30"),
			Underline: boolPtr(true),
		},
		LinkText: scrapbook.StylePrimitive{
			Color: stringPtr("35"),
			Bold:  boolPtr(true),
		},
		Image: scrapbook.StylePrimitive{
			Color:     stringPtr("212"),
			Underline: boolPtr(true),
		},
		ImageText: scrapbook.StylePrimitive{
			Color:  stringPtr("243"),
			Format: "Image: {{.text}} →",
		},
		Code: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           stringPtr("203"),
				BackgroundColor: stringPtr("236"),
			},
		},
		CodeBlock: scrapbook.StyleCodeBlock{
			StyleBlock: scrapbook.StyleBlock{
				StylePrimitive: scrapbook.StylePrimitive{
					Color: stringPtr("244"),
				},
				Margin: uintPtr(defaultMargin),
			},
			Chroma: &scrapbook.Chroma{
				Text: scrapbook.StylePrimitive{
					Color: stringPtr("#C4C4C4"),
				},
				Error: scrapbook.StylePrimitive{
					Color:           stringPtr("#F1F1F1"),
					BackgroundColor: stringPtr("#F05B5B"),
				},
				Comment: scrapbook.StylePrimitive{
					Color: stringPtr("#676767"),
				},
				CommentPreproc: scrapbook.StylePrimitive{
					Color: stringPtr("#FF875F"),
				},
				Keyword: scrapbook.StylePrimitive{
					Color: stringPtr("#00AAFF"),
				},
				KeywordReserved: scrapbook.StylePrimitive{
					Color: stringPtr("#FF5FD2"),
				},
				KeywordNamespace: scrapbook.StylePrimitive{
					Color: stringPtr("#FF5F87"),
				},
				KeywordType: scrapbook.StylePrimitive{
					Color: stringPtr("#6E6ED8"),
				},
				Operator: scrapbook.StylePrimitive{
					Color: stringPtr("#EF8080"),
				},
				Punctuation: scrapbook.StylePrimitive{
					Color: stringPtr("#E8E8A8"),
				},
				Name: scrapbook.StylePrimitive{
					Color: stringPtr("#C4C4C4"),
				},
				NameBuiltin: scrapbook.StylePrimitive{
					Color: stringPtr("#FF8EC7"),
				},
				NameTag: scrapbook.StylePrimitive{
					Color: stringPtr("#B083EA"),
				},
				NameAttribute: scrapbook.StylePrimitive{
					Color: stringPtr("#7A7AE6"),
				},
				NameClass: scrapbook.StylePrimitive{
					Color:     stringPtr("#F1F1F1"),
					Underline: boolPtr(true),
					Bold:      boolPtr(true),
				},
				NameDecorator: scrapbook.StylePrimitive{
					Color: stringPtr("#FFFF87"),
				},
				NameFunction: scrapbook.StylePrimitive{
					Color: stringPtr("#00D787"),
				},
				LiteralNumber: scrapbook.StylePrimitive{
					Color: stringPtr("#6EEFC0"),
				},
				LiteralString: scrapbook.StylePrimitive{
					Color: stringPtr("#C69669"),
				},
				LiteralStringEscape: scrapbook.StylePrimitive{
					Color: stringPtr("#AFFFD7"),
				},
				GenericDeleted: scrapbook.StylePrimitive{
					Color: stringPtr("#FD5B5B"),
				},
				GenericEmph: scrapbook.StylePrimitive{
					Italic: boolPtr(true),
				},
				GenericInserted: scrapbook.StylePrimitive{
					Color: stringPtr("#00D787"),
				},
				GenericStrong: scrapbook.StylePrimitive{
					Bold: boolPtr(true),
				},
				GenericSubheading: scrapbook.StylePrimitive{
					Color: stringPtr("#777777"),
				},
				Background: scrapbook.StylePrimitive{
					BackgroundColor: stringPtr("#373737"),
				},
			},
		},
		Table: scrapbook.StyleTable{
			StyleBlock: scrapbook.StyleBlock{
				StylePrimitive: scrapbook.StylePrimitive{},
			},
			CenterSeparator: stringPtr("┼"),
			ColumnSeparator: stringPtr("│"),
			RowSeparator:    stringPtr("─"),
		},
		DefinitionDescription: scrapbook.StylePrimitive{
			BlockPrefix: "\n🠶 ",
		},
	}

	// LightStyleConfig is the default light style.
	LightStyleConfig = scrapbook.StyleConfig{
		Document: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       stringPtr("234"),
			},
			Margin: uintPtr(defaultMargin),
		},
		BlockQuote: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{},
			Indent:         uintPtr(1),
			IndentToken:    stringPtr("│ "),
		},
		List: scrapbook.StyleList{
			LevelIndent: defaultListIndent,
		},
		Heading: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockSuffix: "\n",
				Color:       stringPtr("27"),
				Bold:        boolPtr(true),
			},
		},
		H1: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           stringPtr("228"),
				BackgroundColor: stringPtr("63"),
				Bold:            boolPtr(true),
			},
		},
		H2: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "## ",
			},
		},
		H3: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "### ",
			},
		},
		H4: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "#### ",
			},
		},
		H5: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "##### ",
			},
		},
		H6: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "###### ",
				Bold:   boolPtr(false),
			},
		},
		Strikethrough: scrapbook.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: scrapbook.StylePrimitive{
			Italic: boolPtr(true),
		},
		Strong: scrapbook.StylePrimitive{
			Bold: boolPtr(true),
		},
		HorizontalRule: scrapbook.StylePrimitive{
			Color:  stringPtr("249"),
			Format: "\n--------\n",
		},
		Item: scrapbook.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: scrapbook.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: scrapbook.StyleTask{
			StylePrimitive: scrapbook.StylePrimitive{},
			Ticked:         "[✓] ",
			Unticked:       "[ ] ",
		},
		Link: scrapbook.StylePrimitive{
			Color:     stringPtr("36"),
			Underline: boolPtr(true),
		},
		LinkText: scrapbook.StylePrimitive{
			Color: stringPtr("29"),
			Bold:  boolPtr(true),
		},
		Image: scrapbook.StylePrimitive{
			Color:     stringPtr("205"),
			Underline: boolPtr(true),
		},
		ImageText: scrapbook.StylePrimitive{
			Color:  stringPtr("243"),
			Format: "Image: {{.text}} →",
		},
		Code: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           stringPtr("203"),
				BackgroundColor: stringPtr("254"),
			},
		},
		CodeBlock: scrapbook.StyleCodeBlock{
			StyleBlock: scrapbook.StyleBlock{
				StylePrimitive: scrapbook.StylePrimitive{
					Color: stringPtr("242"),
				},
				Margin: uintPtr(defaultMargin),
			},
			Chroma: &scrapbook.Chroma{
				Text: scrapbook.StylePrimitive{
					Color: stringPtr("#2A2A2A"),
				},
				Error: scrapbook.StylePrimitive{
					Color:           stringPtr("#F1F1F1"),
					BackgroundColor: stringPtr("#FF5555"),
				},
				Comment: scrapbook.StylePrimitive{
					Color: stringPtr("#8D8D8D"),
				},
				CommentPreproc: scrapbook.StylePrimitive{
					Color: stringPtr("#FF875F"),
				},
				Keyword: scrapbook.StylePrimitive{
					Color: stringPtr("#279EFC"),
				},
				KeywordReserved: scrapbook.StylePrimitive{
					Color: stringPtr("#FF5FD2"),
				},
				KeywordNamespace: scrapbook.StylePrimitive{
					Color: stringPtr("#FB406F"),
				},
				KeywordType: scrapbook.StylePrimitive{
					Color: stringPtr("#7049C2"),
				},
				Operator: scrapbook.StylePrimitive{
					Color: stringPtr("#FF2626"),
				},
				Punctuation: scrapbook.StylePrimitive{
					Color: stringPtr("#FA7878"),
				},
				NameBuiltin: scrapbook.StylePrimitive{
					Color: stringPtr("#0A1BB1"),
				},
				NameTag: scrapbook.StylePrimitive{
					Color: stringPtr("#581290"),
				},
				NameAttribute: scrapbook.StylePrimitive{
					Color: stringPtr("#8362CB"),
				},
				NameClass: scrapbook.StylePrimitive{
					Color:     stringPtr("#212121"),
					Underline: boolPtr(true),
					Bold:      boolPtr(true),
				},
				NameConstant: scrapbook.StylePrimitive{
					Color: stringPtr("#581290"),
				},
				NameDecorator: scrapbook.StylePrimitive{
					Color: stringPtr("#A3A322"),
				},
				NameFunction: scrapbook.StylePrimitive{
					Color: stringPtr("#019F57"),
				},
				LiteralNumber: scrapbook.StylePrimitive{
					Color: stringPtr("#22CCAE"),
				},
				LiteralString: scrapbook.StylePrimitive{
					Color: stringPtr("#7E5B38"),
				},
				LiteralStringEscape: scrapbook.StylePrimitive{
					Color: stringPtr("#00AEAE"),
				},
				GenericDeleted: scrapbook.StylePrimitive{
					Color: stringPtr("#FD5B5B"),
				},
				GenericEmph: scrapbook.StylePrimitive{
					Italic: boolPtr(true),
				},
				GenericInserted: scrapbook.StylePrimitive{
					Color: stringPtr("#00D787"),
				},
				GenericStrong: scrapbook.StylePrimitive{
					Bold: boolPtr(true),
				},
				GenericSubheading: scrapbook.StylePrimitive{
					Color: stringPtr("#777777"),
				},
				Background: scrapbook.StylePrimitive{
					BackgroundColor: stringPtr("#373737"),
				},
			},
		},
		Table: scrapbook.StyleTable{
			StyleBlock: scrapbook.StyleBlock{
				StylePrimitive: scrapbook.StylePrimitive{},
			},
			CenterSeparator: stringPtr("┼"),
			ColumnSeparator: stringPtr("│"),
			RowSeparator:    stringPtr("─"),
		},
		DefinitionDescription: scrapbook.StylePrimitive{
			BlockPrefix: "\n🠶 ",
		},
	}

	// PinkStyleConfig is the default pink style.
	PinkStyleConfig = scrapbook.StyleConfig{
		Document: scrapbook.StyleBlock{
			Margin: uintPtr(defaultMargin),
		},
		BlockQuote: scrapbook.StyleBlock{
			Indent:      uintPtr(1),
			IndentToken: stringPtr("│ "),
		},
		List: scrapbook.StyleList{
			LevelIndent: defaultListIndent,
		},
		Heading: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockSuffix: "\n",
				Color:       stringPtr("212"),
				Bold:        boolPtr(true),
			},
		},
		H1: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				BlockSuffix: "\n",
				BlockPrefix: "\n",
				Prefix:      "",
			},
		},
		H2: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "▌ ",
			},
		},
		H3: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "┃ ",
			},
		},
		H4: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "│ ",
			},
		},
		H5: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "┆ ",
			},
		},
		H6: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Prefix: "┊ ",
				Bold:   boolPtr(false),
			},
		},
		Text: scrapbook.StylePrimitive{},
		Strikethrough: scrapbook.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: scrapbook.StylePrimitive{
			Italic: boolPtr(true),
		},
		Strong: scrapbook.StylePrimitive{
			Bold: boolPtr(true),
		},
		HorizontalRule: scrapbook.StylePrimitive{
			Color:  stringPtr("212"),
			Format: "\n──────\n",
		},
		Item: scrapbook.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: scrapbook.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: scrapbook.StyleTask{
			Ticked:   "[✓] ",
			Unticked: "[ ] ",
		},
		Link: scrapbook.StylePrimitive{
			Color:     stringPtr("99"),
			Underline: boolPtr(true),
		},
		LinkText: scrapbook.StylePrimitive{
			Bold: boolPtr(true),
		},
		Image: scrapbook.StylePrimitive{
			Underline: boolPtr(true),
		},
		ImageText: scrapbook.StylePrimitive{
			Format: "Image: {{.text}}",
		},
		Code: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Color:           stringPtr("212"),
				BackgroundColor: stringPtr("236"),
				Prefix:          " ",
				Suffix:          " ",
			},
		},
		Table: scrapbook.StyleTable{
			CenterSeparator: stringPtr("┼"),
			ColumnSeparator: stringPtr("│"),
			RowSeparator:    stringPtr("─"),
		},
		DefinitionList: scrapbook.StyleBlock{},
		DefinitionTerm: scrapbook.StylePrimitive{},
		DefinitionDescription: scrapbook.StylePrimitive{
			BlockPrefix: "\n🠶 ",
		},
		HTMLBlock: scrapbook.StyleBlock{},
		HTMLSpan:  scrapbook.StyleBlock{},
	}

	// NoTTYStyleConfig is the default notty style.
	NoTTYStyleConfig = ASCIIStyleConfig

	// DefaultStyles are the default styles.
	DefaultStyles = map[string]*scrapbook.StyleConfig{
		AsciiStyle:   &ASCIIStyleConfig,
		DarkStyle:    &DarkStyleConfig,
		DraculaStyle: &DraculaStyleConfig,
		LightStyle:   &LightStyleConfig,
		NoTTYStyle:   &NoTTYStyleConfig,
		PinkStyle:    &PinkStyleConfig,
	}
)

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }
