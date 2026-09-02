package styles

import "charm.land/glamour/v2/ansi"

// GruvboxLightStyleConfig is the Gruvbox light style.
var GruvboxLightStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
			Color:       stringPtr("#282828"),
		},
		Margin: uintPtr(defaultMargin),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#b16286"),
			Italic: boolPtr(true),
		},
		Indent:      uintPtr(2),
		IndentToken: stringPtr("│ "),
	},
	List: ansi.StyleList{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#282828"),
			},
		},
		LevelIndent: defaultListIndent,
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#458588"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "# ",
			Color:  stringPtr("#458588"),
			Bold:   boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "## ",
			Color:  stringPtr("#458588"),
			Bold:   boolPtr(true),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "### ",
			Color:  stringPtr("#689d6a"),
			Bold:   boolPtr(true),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "#### ",
			Color:  stringPtr("#d79921"),
			Bold:   boolPtr(true),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "##### ",
			Color:  stringPtr("#d65d0e"),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "###### ",
			Color:  stringPtr("#cc241d"),
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Color:  stringPtr("#b16286"),
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Color: stringPtr("#d65d0e"),
		Bold:  boolPtr(true),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#bdae93"),
		Format: "\n--------\n",
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "• ",
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
	},
	Task: ansi.StyleTask{
		StylePrimitive: ansi.StylePrimitive{},
		Ticked:         "[✓] ",
		Unticked:       "[ ] ",
	},
	Link: ansi.StylePrimitive{
		Color:     stringPtr("#458588"),
		Underline: boolPtr(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#458588"),
		Bold:  boolPtr(true),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#d65d0e"),
		Underline: boolPtr(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#928374"),
		Format: "Image: {{.text}} →",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          " ",
			Suffix:          " ",
			Color:           stringPtr("#cc241d"),
			BackgroundColor: stringPtr("#f2e5bc"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#282828"),
			},
			Margin: uintPtr(defaultMargin),
		},
		Chroma: &ansi.Chroma{
			Text: ansi.StylePrimitive{
				Color: stringPtr("#282828"),
			},
			Error: ansi.StylePrimitive{
				Color:           stringPtr("#fbf1c7"),
				BackgroundColor: stringPtr("#cc241d"),
			},
			Comment: ansi.StylePrimitive{
				Color: stringPtr("#928374"),
			},
			CommentPreproc: ansi.StylePrimitive{
				Color: stringPtr("#d65d0e"),
			},
			Keyword: ansi.StylePrimitive{
				Color: stringPtr("#458588"),
			},
			KeywordReserved: ansi.StylePrimitive{
				Color: stringPtr("#b16286"),
			},
			KeywordNamespace: ansi.StylePrimitive{
				Color: stringPtr("#cc241d"),
			},
			KeywordType: ansi.StylePrimitive{
				Color: stringPtr("#689d6a"),
			},
			Operator: ansi.StylePrimitive{
				Color: stringPtr("#cc241d"),
			},
			Punctuation: ansi.StylePrimitive{
				Color: stringPtr("#cc241d"),
			},
			NameBuiltin: ansi.StylePrimitive{
				Color: stringPtr("#458588"),
			},
			NameTag: ansi.StylePrimitive{
				Color: stringPtr("#689d6a"),
			},
			NameAttribute: ansi.StylePrimitive{
				Color: stringPtr("#689d6a"),
			},
			NameClass: ansi.StylePrimitive{
				Color:     stringPtr("#282828"),
				Underline: boolPtr(true),
				Bold:      boolPtr(true),
			},
			NameConstant: ansi.StylePrimitive{
				Color: stringPtr("#689d6a"),
			},
			NameDecorator: ansi.StylePrimitive{
				Color: stringPtr("#d79921"),
			},
			NameFunction: ansi.StylePrimitive{
				Color: stringPtr("#458588"),
			},
			LiteralNumber: ansi.StylePrimitive{
				Color: stringPtr("#d79921"),
			},
			LiteralString: ansi.StylePrimitive{
				Color: stringPtr("#98971a"),
			},
			LiteralStringEscape: ansi.StylePrimitive{
				Color: stringPtr("#689d6a"),
			},
			GenericDeleted: ansi.StylePrimitive{
				Color: stringPtr("#cc241d"),
			},
			GenericEmph: ansi.StylePrimitive{
				Italic: boolPtr(true),
			},
			GenericInserted: ansi.StylePrimitive{
				Color: stringPtr("#98971a"),
			},
			GenericStrong: ansi.StylePrimitive{
				Bold: boolPtr(true),
			},
			GenericSubheading: ansi.StylePrimitive{
				Color: stringPtr("#928374"),
			},
			Background: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#f2e5bc"),
			},
		},
	},
	Table: ansi.StyleTable{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{},
		},
	},
	DefinitionDescription: ansi.StylePrimitive{
		BlockPrefix: "\n🠶 ",
	},
	HTMLBlock: ansi.StyleBlock{},
	HTMLSpan:  ansi.StyleBlock{},
}
