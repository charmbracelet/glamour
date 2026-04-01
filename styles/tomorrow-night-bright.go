package styles

import "charm.land/glamour/v2/ansi"

// TomorrowNightBrightStyleConfig is a theme based on the Tomorrow Night Bright palette.
var TomorrowNightBrightStyleConfig = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "\n",
			BlockSuffix: "\n",
			Color:       stringPtr("#eaeaea"),
		},
		Margin: uintPtr(defaultMargin),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:  stringPtr("#969896"),
			Italic: boolPtr(true),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	},
	List: ansi.StyleList{
		LevelIndent: defaultListIndent,
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#eaeaea"),
			},
		},
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       stringPtr("#7aa6da"),
			Bold:        boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          " ",
			Suffix:          " ",
			Color:           stringPtr("#000000"),
			BackgroundColor: stringPtr("#7aa6da"),
			Bold:            boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "## ",
			Color:  stringPtr("#e7c547"),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "### ",
			Color:  stringPtr("#b9ca4a"),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "#### ",
			Color:  stringPtr("#c397d8"),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "##### ",
			Color:  stringPtr("#70c0b1"),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "###### ",
			Color:  stringPtr("#969896"),
			Bold:   boolPtr(false),
		},
	},
	Strikethrough: ansi.StylePrimitive{
		CrossedOut: boolPtr(true),
	},
	Emph: ansi.StylePrimitive{
		Color:  stringPtr("#e7c547"),
		Italic: boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Bold:  boolPtr(true),
		Color: stringPtr("#d54e53"),
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:  stringPtr("#424242"),
		Format: "\n--------\n",
	},
	Item: ansi.StylePrimitive{
		BlockPrefix: "• ",
	},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
		Color:       stringPtr("#70c0b1"),
	},
	Task: ansi.StyleTask{
		StylePrimitive: ansi.StylePrimitive{},
		Ticked:         "[✓] ",
		Unticked:       "[ ] ",
	},
	Link: ansi.StylePrimitive{
		Color:     stringPtr("#7aa6da"),
		Underline: boolPtr(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: stringPtr("#c397d8"),
		Bold:  boolPtr(true),
	},
	Image: ansi.StylePrimitive{
		Color:     stringPtr("#7aa6da"),
		Underline: boolPtr(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  stringPtr("#969896"),
		Format: "Image: {{.text}} →",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#b9ca4a"),
			BackgroundColor: stringPtr("#2a2a2a"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#d0d0d0"),
			},
			Margin: uintPtr(defaultMargin),
		},
		Chroma: &ansi.Chroma{
			Text:                ansi.StylePrimitive{Color: stringPtr("#eaeaea")},
			Error:               ansi.StylePrimitive{Color: stringPtr("#eaeaea"), BackgroundColor: stringPtr("#d54e53")},
			Comment:             ansi.StylePrimitive{Color: stringPtr("#969896")},
			CommentPreproc:      ansi.StylePrimitive{Color: stringPtr("#70c0b1")},
			Keyword:             ansi.StylePrimitive{Color: stringPtr("#c397d8")},
			KeywordReserved:     ansi.StylePrimitive{Color: stringPtr("#c397d8")},
			KeywordNamespace:    ansi.StylePrimitive{Color: stringPtr("#d54e53")},
			KeywordType:         ansi.StylePrimitive{Color: stringPtr("#e7c547")},
			Operator:            ansi.StylePrimitive{Color: stringPtr("#d54e53")},
			Punctuation:         ansi.StylePrimitive{Color: stringPtr("#eaeaea")},
			Name:                ansi.StylePrimitive{Color: stringPtr("#eaeaea")},
			NameBuiltin:         ansi.StylePrimitive{Color: stringPtr("#e7c547")},
			NameTag:             ansi.StylePrimitive{Color: stringPtr("#d54e53")},
			NameAttribute:       ansi.StylePrimitive{Color: stringPtr("#b9ca4a")},
			NameClass:           ansi.StylePrimitive{Color: stringPtr("#e7c547"), Underline: boolPtr(true), Bold: boolPtr(true)},
			NameConstant:        ansi.StylePrimitive{Color: stringPtr("#c397d8")},
			NameDecorator:       ansi.StylePrimitive{Color: stringPtr("#b9ca4a")},
			NameFunction:        ansi.StylePrimitive{Color: stringPtr("#7aa6da")},
			LiteralNumber:       ansi.StylePrimitive{Color: stringPtr("#70c0b1")},
			LiteralString:       ansi.StylePrimitive{Color: stringPtr("#b9ca4a")},
			LiteralStringEscape: ansi.StylePrimitive{Color: stringPtr("#d54e53")},
			GenericDeleted:      ansi.StylePrimitive{Color: stringPtr("#d54e53")},
			GenericEmph:         ansi.StylePrimitive{Color: stringPtr("#e7c547"), Italic: boolPtr(true)},
			GenericInserted:     ansi.StylePrimitive{Color: stringPtr("#b9ca4a")},
			GenericStrong:       ansi.StylePrimitive{Color: stringPtr("#d54e53"), Bold: boolPtr(true)},
			GenericSubheading:   ansi.StylePrimitive{Color: stringPtr("#969896")},
			Background:          ansi.StylePrimitive{BackgroundColor: stringPtr("#000000")},
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
