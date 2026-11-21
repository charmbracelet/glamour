package styles

import "charm.land/glamour/v2/ansi"

// TokyoNightStyleConfig is the tokyo night style.
var TokyoNightStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
			Color:       stringPtr("#c0caf5"),
		},
		Margin: uintPtr(defaultMargin),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#565f89"),
			Italic: boolPtr(true),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	},
	List: ansi.StyleList{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#c0caf5"),
			},
		},
		LevelIndent: defaultListIndent,
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#7aa2f7"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          "# ",
			Color:           stringPtr("#7aa2f7"),
			BackgroundColor: stringPtr("#24293b"),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          defaultHeadingPrefixH2,
			Color:           stringPtr("#e0af68"),
			BackgroundColor: stringPtr("#2e2a2d"),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          defaultHeadingPrefixH3,
			Color:           stringPtr("#9ece6a"),
			BackgroundColor: stringPtr("#272d2d"),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          defaultHeadingPrefixH4,
			Color:           stringPtr("#1abc9c"),
			BackgroundColor: stringPtr("#1a2b32"),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          defaultHeadingPrefixH5,
			Color:           stringPtr("#bb9af7"),
			BackgroundColor: stringPtr("#2a283b"),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          defaultHeadingPrefixH6,
			Color:           stringPtr("#9d7cd8"),
			BackgroundColor: stringPtr("#272538"),
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Bold: boolPtr(true),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#ff9e64"),
		Format: defaultHorizontalRule,
	},
	Item: ansi.StylePrimitive{
		Prefix: "• ",
		Color:  stringPtr("#ff9e64"),
		Bold:   boolPtr(true),
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
	},
	Task: ansi.StyleTask{
		StylePrimitive: ansi.StylePrimitive{},
		Ticked:         defaultTaskTickedPrefix,
		Unticked:       defaultTaskUntickedPrefix,
	},
	Link: ansi.StylePrimitive{
		Color:     stringPtr("#1abc9c"),
		Underline: boolPtr(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#2ac3de"),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#1abc9c"),
		Underline: boolPtr(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#2ac3de"),
		Format: defaultImageFormat,
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#7aa2f7"),
			BackgroundColor: stringPtr("#414868"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{},
			Margin:         uintPtr(defaultMargin),
		},
		Chroma: &ansi.Chroma{
			Text: ansi.StylePrimitive{
				Color: stringPtr("#c0caf5"),
			},
			Error: ansi.StylePrimitive{
				Color: stringPtr("#db4b4b"),
			},
			Comment: ansi.StylePrimitive{
				Color:  stringPtr("#565f89"),
				Italic: boolPtr(true),
			},
			CommentPreproc: ansi.StylePrimitive{
				Color: stringPtr("#7dcfff"),
			},
			Keyword: ansi.StylePrimitive{
				Color:  stringPtr("#9d7cd8"),
				Italic: boolPtr(true),
			},
			KeywordReserved: ansi.StylePrimitive{
				Color: stringPtr("#bb9af7"),
			},
			KeywordNamespace: ansi.StylePrimitive{
				Color: stringPtr("#7dcfff"),
			},
			KeywordType: ansi.StylePrimitive{
				Color: stringPtr("#27a1b9"),
			},
			Operator: ansi.StylePrimitive{
				Color: stringPtr("#89ddff"),
			},
			Punctuation: ansi.StylePrimitive{
				Color: stringPtr("#a9b1d6"),
			},
			Name: ansi.StylePrimitive{
				Color: stringPtr("#c0caf5"),
			},
			NameBuiltin: ansi.StylePrimitive{
				Color: stringPtr("#2ac3de"),
			},
			NameTag: ansi.StylePrimitive{
				Color: stringPtr("#bb9af7"),
			},
			NameAttribute: ansi.StylePrimitive{
				Color: stringPtr("#73daca"),
			},
			NameClass: ansi.StylePrimitive{
				Color: stringPtr("#2ac3de"),
			},
			NameConstant: ansi.StylePrimitive{
				Color: stringPtr("#ff9e64"),
			},
			NameDecorator: ansi.StylePrimitive{
				Color: stringPtr("#7dcfff"),
			},
			NameException: ansi.StylePrimitive{
				Color: stringPtr("#27a1b9"),
			},
			NameFunction: ansi.StylePrimitive{
				Color: stringPtr("#7aa2f7"),
			},
			NameOther: ansi.StylePrimitive{
				Color: stringPtr("#c0caf5"),
			},
			Literal: ansi.StylePrimitive{
				Color: stringPtr("#ff9e64"),
			},
			LiteralNumber: ansi.StylePrimitive{
				Color: stringPtr("#ff9e64"),
			},
			LiteralDate: ansi.StylePrimitive{
				Color: stringPtr("#ff9e64"),
			},
			LiteralString: ansi.StylePrimitive{
				Color: stringPtr("#9ece6a"),
			},
			LiteralStringEscape: ansi.StylePrimitive{
				Color: stringPtr("#bb9af7"),
			},
			GenericDeleted: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#4a272f"),
			},
			GenericEmph: ansi.StylePrimitive{
				Italic: boolPtr(true),
			},
			GenericInserted: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#243e4a"),
			},
			GenericStrong: ansi.StylePrimitive{
				Bold: boolPtr(true),
			},
			GenericSubheading: ansi.StylePrimitive{
				Color: stringPtr("#7aa2f7"),
				Bold:  boolPtr(true),
			},
			Background: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#16161e"),
			},
		},
	},
	Table: ansi.StyleTable{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{},
		},
	},
	DefinitionDescription: ansi.StylePrimitive{
		BlockPrefix: defaultArrowBlockPrefix,
	},
}
