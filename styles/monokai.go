package styles

import "charm.land/glamour/v2/ansi"

// MonokaiStyleConfig is the Monokai style.
var MonokaiStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix:     "\n",
			BlockSuffix:     "\n",
			Color:           stringPtr("#f8f8f2"),
			BackgroundColor: stringPtr("#1c1c1c"),
		},
		Margin: uintPtr(defaultMargin),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#6e7066"),
			Italic: boolPtr(true),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	},
	List: ansi.StyleList{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#f8f8f2"),
			},
		},
		LevelIndent: defaultListIndent,
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#f92672"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          " ",
			Suffix:          " ",
			Color:           stringPtr("#1c1c1c"),
			BackgroundColor: stringPtr("#f92672"),
			Bold:            boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: defaultHeadingPrefixH2,
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: defaultHeadingPrefixH3,
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: defaultHeadingPrefixH4,
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: defaultHeadingPrefixH5,
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: defaultHeadingPrefixH6,
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Color:  stringPtr("#66d9ef"),
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Color: stringPtr("#a6e22e"),
		Bold:  boolPtr(true),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#505050"),
		Format: defaultHorizontalRule,
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "• ",
		Color:       stringPtr("#f92672"),
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
		Color:       stringPtr("#ae81ff"),
	},
	Task: ansi.StyleTask{
		Ticked:   defaultTaskTickedPrefix,
		Unticked: defaultTaskUntickedPrefix,
	},
	Link: ansi.StylePrimitive{
		Color:     stringPtr("#e6db74"),
		Underline: boolPtr(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#a6e22e"),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#ae81ff"),
		Underline: boolPtr(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#66d9ef"),
		Format: defaultImageFormat,
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#e6db74"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#f8f8f2"),
			},
			Margin: uintPtr(defaultMargin),
		},
		Chroma: &ansi.Chroma{
			Text: ansi.StylePrimitive{
				Color: stringPtr("#f8f8f2"),
			},
			Error: ansi.StylePrimitive{
				Color:           stringPtr("#1c1c1c"),
				BackgroundColor: stringPtr("#f4005f"),
			},
			Comment: ansi.StylePrimitive{
				Color:  stringPtr("#6e7066"),
				Italic: boolPtr(true),
			},
			CommentPreproc: ansi.StylePrimitive{
				Color: stringPtr("#ff2d01"),
			},
			Keyword: ansi.StylePrimitive{
				Color: stringPtr("#f92672"),
			},
			KeywordReserved: ansi.StylePrimitive{
				Color: stringPtr("#f92672"),
			},
			KeywordNamespace: ansi.StylePrimitive{
				Color: stringPtr("#f92672"),
			},
			KeywordType: ansi.StylePrimitive{
				Color: stringPtr("#66d9ef"),
			},
			Operator: ansi.StylePrimitive{
				Color: stringPtr("#f92672"),
			},
			Punctuation: ansi.StylePrimitive{
				Color: stringPtr("#6e7066"),
			},
			Name: ansi.StylePrimitive{
				Color: stringPtr("#f8f8f2"),
			},
			NameBuiltin: ansi.StylePrimitive{
				Color: stringPtr("#a6e22e"),
			},
			NameTag: ansi.StylePrimitive{
				Color: stringPtr("#f92672"),
			},
			NameAttribute: ansi.StylePrimitive{
				Color: stringPtr("#66d9ef"),
			},
			NameClass: ansi.StylePrimitive{
				Color: stringPtr("#66d9ef"),
			},
			NameConstant: ansi.StylePrimitive{
				Color:  stringPtr("#ae81ff"),
				Italic: boolPtr(true),
			},
			NameDecorator: ansi.StylePrimitive{
				Color: stringPtr("#a6e22e"),
			},
			NameException: ansi.StylePrimitive{
				Color: stringPtr("#f92672"),
			},
			NameFunction: ansi.StylePrimitive{
				Color: stringPtr("#a6e22e"),
				Bold:  boolPtr(true),
			},
			NameOther: ansi.StylePrimitive{
				Color: stringPtr("#f8f8f2"),
			},
			Literal: ansi.StylePrimitive{
				Color:  stringPtr("#ae81ff"),
				Italic: boolPtr(true),
			},
			LiteralNumber: ansi.StylePrimitive{
				Color:  stringPtr("#ae81ff"),
				Italic: boolPtr(true),
			},
			LiteralDate: ansi.StylePrimitive{
				Color: stringPtr("#e6db74"),
			},
			LiteralString: ansi.StylePrimitive{
				Color: stringPtr("#e6db74"),
			},
			LiteralStringEscape: ansi.StylePrimitive{
				Color: stringPtr("#ae81ff"),
			},
			GenericDeleted: ansi.StylePrimitive{
				Color: stringPtr("#f4005f"),
			},
			GenericEmph: ansi.StylePrimitive{
				Color:  stringPtr("#66d9ef"),
				Italic: boolPtr(true),
			},
			GenericInserted: ansi.StylePrimitive{
				Color: stringPtr("#a6e22e"),
			},
			GenericStrong: ansi.StylePrimitive{
				Color: stringPtr("#a6e22e"),
				Bold:  boolPtr(true),
			},
			GenericSubheading: ansi.StylePrimitive{
				Color: stringPtr("#f92672"),
			},
			Background: ansi.StylePrimitive{
				BackgroundColor: stringPtr("#1c1c1c"),
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
