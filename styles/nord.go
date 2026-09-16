package styles

import "charm.land/glamour/v2/ansi"

// NordStyleConfig is the nord style.
var NordStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
			Color:       stringPtr("#d8dee9"),
		},
		Margin: uintPtr(defaultMargin),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#8B95A7"),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	},
	List: ansi.StyleList{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#d8dee9"),
			},
		},
		LevelIndent: defaultListIndent,
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#88c0d0"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "# ",
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "## ",
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "### ",
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "#### ",
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "##### ",
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "###### ",
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Color:  stringPtr("#d08770"),
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Color: stringPtr("#ebcb8b"),
		Bold:  boolPtr(true),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#8B95A7"),
		Format: "\n--------\n",
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "• ",
		Color:       stringPtr("#88c0d0"),
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
		Color:       stringPtr("#8fbcbb"),
	},
	Task: ansi.StyleTask{
		StylePrimitive: ansi.StylePrimitive{},
		Ticked:         "[✓] ",
		Unticked:       "[ ] ",
	},
	Link: ansi.StylePrimitive{
		Color:     stringPtr("#81a1c1"),
		Underline: boolPtr(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#8fbcbb"),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#81a1c1"),
		Underline: boolPtr(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#8fbcbb"),
		Format: "Image: {{.text}} →",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#a3be8c"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#d8dee9"),
			},
			Margin: uintPtr(defaultMargin),
		},
		Chroma: &ansi.Chroma{
			Text: ansi.StylePrimitive{
				Color: stringPtr("#d8dee9"),
			},
			Error: ansi.StylePrimitive{
				Color:           stringPtr("#d8dee9"),
				BackgroundColor: stringPtr("#bf616a"),
			},
			Comment: ansi.StylePrimitive{
				Color: stringPtr("#8B95A7"),
			},
			CommentPreproc: ansi.StylePrimitive{
				Color: stringPtr("#b48ead"),
			},
			Keyword: ansi.StylePrimitive{
				Color: stringPtr("#81a1c1"),
			},
			KeywordReserved: ansi.StylePrimitive{
				Color: stringPtr("#81a1c1"),
			},
			KeywordNamespace: ansi.StylePrimitive{
				Color: stringPtr("#81a1c1"),
			},
			KeywordType: ansi.StylePrimitive{
				Color: stringPtr("#8fbcbb"),
			},
			Operator: ansi.StylePrimitive{
				Color: stringPtr("#81a1c1"),
			},
			Punctuation: ansi.StylePrimitive{
				Color: stringPtr("#d8dee9"),
			},
			Name: ansi.StylePrimitive{
				Color: stringPtr("#8fbcbb"),
			},
			NameBuiltin: ansi.StylePrimitive{
				Color: stringPtr("#8fbcbb"),
			},
			NameTag: ansi.StylePrimitive{
				Color: stringPtr("#81a1c1"),
			},
			NameAttribute: ansi.StylePrimitive{
				Color: stringPtr("#a3be8c"),
			},
			NameClass: ansi.StylePrimitive{
				Color: stringPtr("#8fbcbb"),
			},
			NameConstant: ansi.StylePrimitive{
				Color: stringPtr("#b48ead"),
			},
			NameDecorator: ansi.StylePrimitive{
				Color: stringPtr("#d08770"),
			},
			NameFunction: ansi.StylePrimitive{
				Color: stringPtr("#88c0d0"),
			},
			LiteralNumber: ansi.StylePrimitive{
				Color: stringPtr("#b48ead"),
			},
			LiteralString: ansi.StylePrimitive{
				Color: stringPtr("#a3be8c"),
			},
			LiteralStringEscape: ansi.StylePrimitive{
				Color: stringPtr("#ebcb8b"),
			},
			GenericDeleted: ansi.StylePrimitive{
				Color: stringPtr("#bf616a"),
			},
			GenericEmph: ansi.StylePrimitive{
				Color:  stringPtr("#d08770"),
				Italic: boolPtr(true),
			},
			GenericInserted: ansi.StylePrimitive{
				Color: stringPtr("#a3be8c"),
			},
			GenericStrong: ansi.StylePrimitive{
				Color: stringPtr("#ebcb8b"),
				Bold:  boolPtr(true),
			},
			GenericSubheading: ansi.StylePrimitive{
				Color: stringPtr("#88c0d0"),
			},
			Background: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#2e3440"),
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
