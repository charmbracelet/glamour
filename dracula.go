package glamour

import "github.com/charmbracelet/scrapbook"

var DraculaStyleConfig = scrapbook.StyleConfig{
	Document: scrapbook.StyleBlock{
		StylePrimitive: scrapbook.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
			Color:       stringPtr("#f8f8f2"),
		},
		Margin: uintPtr(2),
	},
	BlockQuote: scrapbook.StyleBlock{
		StylePrimitive: scrapbook.StylePrimitive{
			Color:  stringPtr("#f1fa8c"),
			Italic: boolPtr(true),
		},
		Indent: uintPtr(2),
	},
	List: scrapbook.StyleList{
		LevelIndent: 2,
		StyleBlock: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Color: stringPtr("#f8f8f2"),
			},
		},
	},
	Heading: scrapbook.StyleBlock{
		StylePrimitive: scrapbook.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#bd93f9"),
			Bold:        boolPtr(true),
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
		CrossedOut: boolPtr(true),
	},
	Emph: scrapbook.StylePrimitive{
		Color:  stringPtr("#f1fa8c"),
		Italic: boolPtr(true),
	},
	Strong: scrapbook.StylePrimitive{
		Bold:  boolPtr(true),
		Color: stringPtr("#ffb86c"),
	},
	HorizontalRule: scrapbook.StylePrimitive{
		Color:  stringPtr("#6272A4"),
		Format: "\n--------\n",
	},
	Item: scrapbook.StylePrimitive{
		BlockPrefix: "• ",
	},
	Enumeration: scrapbook.StylePrimitive{
		BlockPrefix: ". ",
		Color:       stringPtr("#8be9fd"),
	},
	Task: scrapbook.StyleTask{
		StylePrimitive: scrapbook.StylePrimitive{},
		Ticked:         "[✓] ",
		Unticked:       "[ ] ",
	},
	Link: scrapbook.StylePrimitive{
		Color:     stringPtr("#8be9fd"),
		Underline: boolPtr(true),
	},
	LinkText: scrapbook.StylePrimitive{
		Color: stringPtr("#ff79c6"),
	},
	Image: scrapbook.StylePrimitive{
		Color:     stringPtr("#8be9fd"),
		Underline: boolPtr(true),
	},
	ImageText: scrapbook.StylePrimitive{
		Color:  stringPtr("#ff79c6"),
		Format: "Image: {{.text}} →",
	},
	Code: scrapbook.StyleBlock{
		StylePrimitive: scrapbook.StylePrimitive{
			Color: stringPtr("#50fa7b"),
		},
	},
	CodeBlock: scrapbook.StyleCodeBlock{
		StyleBlock: scrapbook.StyleBlock{
			StylePrimitive: scrapbook.StylePrimitive{
				Color: stringPtr("#ffb86c"),
			},
			Margin: uintPtr(2),
		},
		Chroma: &scrapbook.Chroma{
			Text: scrapbook.StylePrimitive{
				Color: stringPtr("#f8f8f2"),
			},
			Error: scrapbook.StylePrimitive{
				Color:           stringPtr("#f8f8f2"),
				BackgroundColor: stringPtr("#ff5555"),
			},
			Comment: scrapbook.StylePrimitive{
				Color: stringPtr("#6272A4"),
			},
			CommentPreproc: scrapbook.StylePrimitive{
				Color: stringPtr("#ff79c6"),
			},
			Keyword: scrapbook.StylePrimitive{
				Color: stringPtr("#ff79c6"),
			},
			KeywordReserved: scrapbook.StylePrimitive{
				Color: stringPtr("#ff79c6"),
			},
			KeywordNamespace: scrapbook.StylePrimitive{
				Color: stringPtr("#ff79c6"),
			},
			KeywordType: scrapbook.StylePrimitive{
				Color: stringPtr("#8be9fd"),
			},
			Operator: scrapbook.StylePrimitive{
				Color: stringPtr("#ff79c6"),
			},
			Punctuation: scrapbook.StylePrimitive{
				Color: stringPtr("#f8f8f2"),
			},
			Name: scrapbook.StylePrimitive{
				Color: stringPtr("#8be9fd"),
			},
			NameBuiltin: scrapbook.StylePrimitive{
				Color: stringPtr("#8be9fd"),
			},
			NameTag: scrapbook.StylePrimitive{
				Color: stringPtr("#ff79c6"),
			},
			NameAttribute: scrapbook.StylePrimitive{
				Color: stringPtr("#50fa7b"),
			},
			NameClass: scrapbook.StylePrimitive{
				Color: stringPtr("#8be9fd"),
			},
			NameConstant: scrapbook.StylePrimitive{
				Color: stringPtr("#bd93f9"),
			},
			NameDecorator: scrapbook.StylePrimitive{
				Color: stringPtr("#50fa7b"),
			},
			NameFunction: scrapbook.StylePrimitive{
				Color: stringPtr("#50fa7b"),
			},
			LiteralNumber: scrapbook.StylePrimitive{
				Color: stringPtr("#6EEFC0"),
			},
			LiteralString: scrapbook.StylePrimitive{
				Color: stringPtr("#f1fa8c"),
			},
			LiteralStringEscape: scrapbook.StylePrimitive{
				Color: stringPtr("#ff79c6"),
			},
			GenericDeleted: scrapbook.StylePrimitive{
				Color: stringPtr("#ff5555"),
			},
			GenericEmph: scrapbook.StylePrimitive{
				Color:  stringPtr("#f1fa8c"),
				Italic: boolPtr(true),
			},
			GenericInserted: scrapbook.StylePrimitive{
				Color: stringPtr("#50fa7b"),
			},
			GenericStrong: scrapbook.StylePrimitive{
				Color: stringPtr("#ffb86c"),
				Bold:  boolPtr(true),
			},
			GenericSubheading: scrapbook.StylePrimitive{
				Color: stringPtr("#bd93f9"),
			},
			Background: scrapbook.StylePrimitive{
				BackgroundColor: stringPtr("#282a36"),
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
