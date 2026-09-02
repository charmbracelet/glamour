package styles

import "charm.land/glamour/v2/ansi"

type rosePinePalette struct {
	Base    string
	Surface string
	Overlay string
	Muted   string
	Subtle  string
	Text    string
	Love    string
	Gold    string
	Rose    string
	Pine    string
	Foam    string
	Iris    string
}

func buildRosePineStyleConfig(p rosePinePalette) ansi.StyleConfig {
	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       stringPtr(p.Text),
			},
			Margin: uintPtr(defaultMargin),
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  stringPtr(p.Muted),
				Italic: boolPtr(true),
			},
			Indent:      uintPtr(1),
			IndentToken: stringPtr("│ "),
		},
		Paragraph: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{},
		},
		List: ansi.StyleList{
			LevelIndent: defaultListIndent,
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: stringPtr(p.Text),
				},
			},
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       stringPtr(p.Love),
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
		Text: ansi.StylePrimitive{},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: ansi.StylePrimitive{
			Color:  stringPtr(p.Text),
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Color: stringPtr(p.Rose),
			Bold:  boolPtr(true),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  stringPtr(p.Overlay),
			Format: defaultHorizontalRule,
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
			Color:       stringPtr(p.Foam),
		},
		Task: ansi.StyleTask{
			StylePrimitive: ansi.StylePrimitive{},
			Ticked:         defaultTaskTickedPrefix,
			Unticked:       defaultTaskUntickedPrefix,
		},
		Link: ansi.StylePrimitive{
			Color:     stringPtr(p.Pine),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: stringPtr(p.Iris),
		},
		Image: ansi.StylePrimitive{
			Color:     stringPtr(p.Pine),
			Underline: boolPtr(true),
		},
		ImageText: ansi.StylePrimitive{
			Color:  stringPtr(p.Iris),
			Format: defaultImageFormat,
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           stringPtr(p.Gold),
				BackgroundColor: stringPtr(p.Surface),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: stringPtr(p.Gold),
				},
				Margin: uintPtr(defaultMargin),
			},
			Chroma: &ansi.Chroma{
				Text: ansi.StylePrimitive{
					Color: stringPtr(p.Text),
				},
				Error: ansi.StylePrimitive{
					Color:           stringPtr(p.Text),
					BackgroundColor: stringPtr(p.Love),
				},
				Comment: ansi.StylePrimitive{
					Color: stringPtr(p.Muted),
				},
				CommentPreproc: ansi.StylePrimitive{
					Color: stringPtr(p.Iris),
				},
				Keyword: ansi.StylePrimitive{
					Color: stringPtr(p.Pine),
				},
				KeywordReserved: ansi.StylePrimitive{
					Color: stringPtr(p.Pine),
				},
				KeywordNamespace: ansi.StylePrimitive{
					Color: stringPtr(p.Pine),
				},
				KeywordType: ansi.StylePrimitive{
					Color: stringPtr(p.Foam),
				},
				Operator: ansi.StylePrimitive{
					Color: stringPtr(p.Subtle),
				},
				Punctuation: ansi.StylePrimitive{
					Color: stringPtr(p.Text),
				},
				Name: ansi.StylePrimitive{
					Color: stringPtr(p.Text),
				},
				NameBuiltin: ansi.StylePrimitive{
					Color: stringPtr(p.Foam),
				},
				NameTag: ansi.StylePrimitive{
					Color: stringPtr(p.Iris),
				},
				NameAttribute: ansi.StylePrimitive{
					Color: stringPtr(p.Foam),
				},
				NameClass: ansi.StylePrimitive{
					Color: stringPtr(p.Foam),
				},
				NameConstant: ansi.StylePrimitive{
					Color: stringPtr(p.Rose),
				},
				NameDecorator: ansi.StylePrimitive{
					Color: stringPtr(p.Gold),
				},
				NameFunction: ansi.StylePrimitive{
					Color: stringPtr(p.Rose),
				},
				LiteralNumber: ansi.StylePrimitive{
					Color: stringPtr(p.Foam),
				},
				LiteralString: ansi.StylePrimitive{
					Color: stringPtr(p.Gold),
				},
				LiteralStringEscape: ansi.StylePrimitive{
					Color: stringPtr(p.Pine),
				},
				GenericDeleted: ansi.StylePrimitive{
					Color: stringPtr(p.Love),
				},
				GenericEmph: ansi.StylePrimitive{
					Color:  stringPtr(p.Gold),
					Italic: boolPtr(true),
				},
				GenericInserted: ansi.StylePrimitive{
					Color: stringPtr(p.Pine),
				},
				GenericStrong: ansi.StylePrimitive{
					Color: stringPtr(p.Rose),
					Bold:  boolPtr(true),
				},
				GenericSubheading: ansi.StylePrimitive{
					Color: stringPtr(p.Iris),
				},
				Background: ansi.StylePrimitive{
					BackgroundColor: stringPtr(p.Base),
				},
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{},
			},
		},
		DefinitionList: ansi.StyleBlock{},
		DefinitionTerm: ansi.StylePrimitive{},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: defaultArrowBlockPrefix,
		},
		HTMLBlock: ansi.StyleBlock{},
		HTMLSpan:  ansi.StyleBlock{},
	}
}
