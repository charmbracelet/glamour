package styles

import "charm.land/glamour/v2/ansi"

// GruvboxStyleConfig is the gruvbox dark style.
var GruvboxStyleConfig = ansi.StyleConfig{
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
			Color:  stringPtr("#a89984"),
			Italic: boolPtr(true),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	},
	List: ansi.StyleList{
		LevelIndent: defaultListIndent,
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#ebdbb2"),
			},
		},
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#fabd2f"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          " ",
			Suffix:          " ",
			Color:           stringPtr("#282828"),
			BackgroundColor: stringPtr("#fabd2f"),
			Bold:            boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "## ",
			Color:  stringPtr("#b8bb26"),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "### ",
			Color:  stringPtr("#83a598"),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "#### ",
			Color:  stringPtr("#d3869b"),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "##### ",
			Color:  stringPtr("#8ec07c"),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "###### ",
			Color:  stringPtr("#a89984"),
			Bold:   boolPtr(false),
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Color:  stringPtr("#fabd2f"),
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Bold:  boolPtr(true),
		Color: stringPtr("#fe8019"),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#504945"),
		Format: "\n--------\n",
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "• ",
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
		Color:       stringPtr("#83a598"),
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
		Color: stringPtr("#d3869b"),
		Bold:  boolPtr(true),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#83a598"),
		Underline: boolPtr(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#a89984"),
		Format: "Image: {{.text}} →",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#b8bb26"),
			BackgroundColor: stringPtr("#3c3836"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#d5c4a1"),
			},
			Margin: uintPtr(defaultMargin),
		},
		Chroma: &ansi.Chroma{
			Text: ansi.StylePrimitive{
				Color: stringPtr("#ebdbb2"),
			},
			Error: ansi.StylePrimitive{
				Color:           stringPtr("#ebdbb2"),
				BackgroundColor: stringPtr("#fb4934"),
			},
			Comment: ansi.StylePrimitive{
				Color: stringPtr("#928374"),
			},
			CommentPreproc: ansi.StylePrimitive{
				Color: stringPtr("#8ec07c"),
			},
			Keyword: ansi.StylePrimitive{
				Color: stringPtr("#fb4934"),
			},
			KeywordReserved: ansi.StylePrimitive{
				Color: stringPtr("#fb4934"),
			},
			KeywordNamespace: ansi.StylePrimitive{
				Color: stringPtr("#fe8019"),
			},
			KeywordType: ansi.StylePrimitive{
				Color: stringPtr("#fabd2f"),
			},
			Operator: ansi.StylePrimitive{
				Color: stringPtr("#fe8019"),
			},
			Punctuation: ansi.StylePrimitive{
				Color: stringPtr("#ebdbb2"),
			},
			Name: ansi.StylePrimitive{
				Color: stringPtr("#ebdbb2"),
			},
			NameBuiltin: ansi.StylePrimitive{
				Color: stringPtr("#fabd2f"),
			},
			NameTag: ansi.StylePrimitive{
				Color: stringPtr("#fb4934"),
			},
			NameAttribute: ansi.StylePrimitive{
				Color: stringPtr("#b8bb26"),
			},
			NameClass: ansi.StylePrimitive{
				Color:     stringPtr("#fabd2f"),
				Underline: boolPtr(true),
				Bold:      boolPtr(true),
			},
			NameConstant: ansi.StylePrimitive{
				Color: stringPtr("#d3869b"),
			},
			NameDecorator: ansi.StylePrimitive{
				Color: stringPtr("#b8bb26"),
			},
			NameFunction: ansi.StylePrimitive{
				Color: stringPtr("#b8bb26"),
			},
			LiteralNumber: ansi.StylePrimitive{
				Color: stringPtr("#d3869b"),
			},
			LiteralString: ansi.StylePrimitive{
				Color: stringPtr("#b8bb26"),
			},
			LiteralStringEscape: ansi.StylePrimitive{
				Color: stringPtr("#fe8019"),
			},
			GenericDeleted: ansi.StylePrimitive{
				Color: stringPtr("#fb4934"),
			},
			GenericEmph: ansi.StylePrimitive{
				Color:  stringPtr("#fabd2f"),
				Italic: boolPtr(true),
			},
			GenericInserted: ansi.StylePrimitive{
				Color: stringPtr("#b8bb26"),
			},
			GenericStrong: ansi.StylePrimitive{
				Color: stringPtr("#fe8019"),
				Bold:  boolPtr(true),
			},
			GenericSubheading: ansi.StylePrimitive{
				Color: stringPtr("#928374"),
			},
			Background: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#282828"),
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
}
