package styles

import "charm.land/glamour/v2/ansi"

// GruvboxDarkStyleConfig is the Gruvbox dark style.
var GruvboxDarkStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
			Color:       stringPtr("#ebdbb2"),
		},
		Margin: uintPtr(defaultMargin),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#d3869b"),
			Italic: boolPtr(true),
		},
		Indent:      uintPtr(2),
		IndentToken: stringPtr("│ "),
	},
	List: ansi.StyleList{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#ebdbb2"),
			},
		},
		LevelIndent: defaultListIndent,
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#83a598"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "# ",
			Color:  stringPtr("#83a598"),
			Bold:   boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "## ",
			Color:  stringPtr("#83a598"),
			Bold:   boolPtr(true),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "### ",
			Color:  stringPtr("#8ec07c"),
			Bold:   boolPtr(true),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "#### ",
			Color:  stringPtr("#fabd2f"),
			Bold:   boolPtr(true),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "##### ",
			Color:  stringPtr("#fe8019"),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "###### ",
			Color:  stringPtr("#fb4934"),
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Color:  stringPtr("#d3869b"),
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Color: stringPtr("#fe8019"),
		Bold:  boolPtr(true),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#a89984"),
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
		Color:     stringPtr("#83a598"),
		Underline: boolPtr(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#83a598"),
		Bold:  boolPtr(true),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#fe8019"),
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
			Color:           stringPtr("#fb4934"),
			BackgroundColor: stringPtr("#3c3836"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#ebdbb2"),
			},
			Margin: uintPtr(defaultMargin),
		},
		Chroma: &ansi.Chroma{
			Text: ansi.StylePrimitive{
				Color: stringPtr("#ebdbb2"),
			},
			Error: ansi.StylePrimitive{
				Color:           stringPtr("#282828"),
				BackgroundColor: stringPtr("#fb4934"),
			},
			Comment: ansi.StylePrimitive{
				Color: stringPtr("#928374"),
			},
			CommentPreproc: ansi.StylePrimitive{
				Color: stringPtr("#fe8019"),
			},
			Keyword: ansi.StylePrimitive{
				Color: stringPtr("#83a598"),
			},
			KeywordReserved: ansi.StylePrimitive{
				Color: stringPtr("#d3869b"),
			},
			KeywordNamespace: ansi.StylePrimitive{
				Color: stringPtr("#fb4934"),
			},
			KeywordType: ansi.StylePrimitive{
				Color: stringPtr("#8ec07c"),
			},
			Operator: ansi.StylePrimitive{
				Color: stringPtr("#fb4934"),
			},
			Punctuation: ansi.StylePrimitive{
				Color: stringPtr("#fb4934"),
			},
			NameBuiltin: ansi.StylePrimitive{
				Color: stringPtr("#83a598"),
			},
			NameTag: ansi.StylePrimitive{
				Color: stringPtr("#8ec07c"),
			},
			NameAttribute: ansi.StylePrimitive{
				Color: stringPtr("#8ec07c"),
			},
			NameClass: ansi.StylePrimitive{
				Color:     stringPtr("#ebdbb2"),
				Underline: boolPtr(true),
				Bold:      boolPtr(true),
			},
			NameConstant: ansi.StylePrimitive{
				Color: stringPtr("#8ec07c"),
			},
			NameDecorator: ansi.StylePrimitive{
				Color: stringPtr("#fabd2f"),
			},
			NameFunction: ansi.StylePrimitive{
				Color: stringPtr("#83a598"),
			},
			LiteralNumber: ansi.StylePrimitive{
				Color: stringPtr("#fabd2f"),
			},
			LiteralString: ansi.StylePrimitive{
				Color: stringPtr("#b8bb26"),
			},
			LiteralStringEscape: ansi.StylePrimitive{
				Color: stringPtr("#8ec07c"),
			},
			GenericDeleted: ansi.StylePrimitive{
				Color: stringPtr("#fb4934"),
			},
			GenericEmph: ansi.StylePrimitive{
				Italic: boolPtr(true),
			},
			GenericInserted: ansi.StylePrimitive{
				Color: stringPtr("#b8bb26"),
			},
			GenericStrong: ansi.StylePrimitive{
				Bold: boolPtr(true),
			},
			GenericSubheading: ansi.StylePrimitive{
				Color: stringPtr("#928374"),
			},
			Background: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#32302f"),
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
