package ansi

import (
	"strings"
	"testing"
)

func TestFormatHyperlink(t *testing.T) {
	tests := []struct {
		name string
		text string
		url  string
		want string
	}{
		{
			name: "basic hyperlink",
			text: "example",
			url:  "https://example.com",
			want: "\x1b]8;;https://example.com\x1b\\example\x1b]8;;\x1b\\",
		},
		{
			name: "empty text",
			text: "",
			url:  "https://example.com",
			want: "\x1b]8;;https://example.com\x1b\\\x1b]8;;\x1b\\",
		},
		{
			name: "empty url",
			text: "example",
			url:  "",
			want: "example",
		},
		{
			name: "special characters in URL",
			text: "example",
			url:  "https://example.com/path?param=value&other=test",
			want: "\x1b]8;;https://example.com/path?param=value&other=test\x1b\\example\x1b]8;;\x1b\\",
		},
		{
			name: "unicode text",
			text: "例え",
			url:  "https://example.com",
			want: "\x1b]8;;https://example.com\x1b\\例え\x1b]8;;\x1b\\",
		},
		{
			name: "text with spaces",
			text: "click here",
			url:  "https://example.com",
			want: "\x1b]8;;https://example.com\x1b\\click here\x1b]8;;\x1b\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHyperlink(tt.text, tt.url)
			if result != tt.want {
				t.Errorf("formatHyperlink(%q, %q) = %q, want %q", tt.text, tt.url, result, tt.want)
			}
		})
	}
}

func TestSupportsHyperlinks(t *testing.T) {
	tests := []struct {
		name        string
		termProgram string
		term        string
		envVars     map[string]string
		want        bool
	}{
		{
			name:        "iTerm2",
			termProgram: "iTerm.app",
			want:        true,
		},
		{
			name:        "VS Code",
			termProgram: "vscode",
			want:        true,
		},
		{
			name:        "Windows Terminal",
			termProgram: "Windows Terminal",
			want:        true,
		},
		{
			name:        "WezTerm",
			termProgram: "WezTerm",
			want:        true,
		},
		{
			name:        "Hyper",
			termProgram: "Hyper",
			want:        true,
		},
		{
			name: "xterm-256color",
			term: "xterm-256color",
			want: true,
		},
		{
			name: "screen-256color",
			term: "screen-256color",
			want: true,
		},
		{
			name: "tmux-256color",
			term: "tmux-256color",
			want: true,
		},
		{
			name: "alacritty",
			term: "alacritty",
			want: true,
		},
		{
			name: "xterm-kitty",
			term: "xterm-kitty",
			want: true,
		},
		{
			name: "basic xterm",
			term: "xterm",
			want: false,
		},
		{
			name: "unknown terminal",
			want: false,
		},
		{
			name: "kitty terminal",
			envVars: map[string]string{
				"KITTY_WINDOW_ID": "1",
			},
			want: true,
		},
		{
			name: "alacritty by env var",
			envVars: map[string]string{
				"ALACRITTY_LOG": "/tmp/log",
			},
			want: true,
		},
		{
			name: "alacritty by socket env var",
			envVars: map[string]string{
				"ALACRITTY_SOCKET": "/tmp/socket",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.termProgram != "" {
				t.Setenv("TERM_PROGRAM", tt.termProgram)
			} else {
				t.Setenv("TERM_PROGRAM", "")
			}

			if tt.term != "" {
				t.Setenv("TERM", tt.term)
			} else {
				t.Setenv("TERM", "")
			}

			// Set additional environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Test
			result := supportsHyperlinks(RenderContext{})
			if result != tt.want {
				t.Errorf("supportsHyperlinks() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestStripANSISequences(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "no ANSI sequences",
			text: "Hello World",
			want: "Hello World",
		},
		{
			name: "empty string",
			text: "",
			want: "",
		},
		{
			name: "red text",
			text: "\x1b[31mRed Text\x1b[0m",
			want: "Red Text",
		},
		{
			name: "bold text",
			text: "\x1b[1mBold Text\x1b[0m",
			want: "Bold Text",
		},
		{
			name: "hyperlink sequence",
			text: "\x1b]8;;https://example.com\x1b\\Click Here\x1b]8;;\x1b\\",
			want: "Click Here",
		},
		{
			name: "multiple sequences",
			text: "\x1b[31m\x1b[1mRed Bold\x1b[0m\x1b[0m",
			want: "Red Bold",
		},
		{
			name: "cursor movement",
			text: "\x1b[2JClear Screen\x1b[H",
			want: "Clear Screen",
		},
		{
			name: "mixed content",
			text: "Normal \x1b[31mRed\x1b[0m Normal \x1b]8;;url\x1b\\Link\x1b]8;;\x1b\\ Normal",
			want: "Normal Red Normal Link Normal",
		},
		{
			name: "OSC with BEL terminator",
			text: "\x1b]0;Title\x07Content",
			want: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSISequences(tt.text)
			if result != tt.want {
				t.Errorf("stripANSISequences(%q) = %q, want %q", tt.text, result, tt.want)
			}
		})
	}
}

func TestExtractTextFromChildren(t *testing.T) {
	tests := []struct {
		name     string
		children []ElementRenderer
		want     string
		wantErr  bool
	}{
		{
			name:     "no children",
			children: []ElementRenderer{},
			want:     "",
			wantErr:  false,
		},
		{
			name:     "nil children",
			children: nil,
			want:     "",
			wantErr:  false,
		},
		{
			name: "single text element",
			children: []ElementRenderer{
				&BaseElement{Token: "Hello"},
			},
			want:    "Hello",
			wantErr: false,
		},
		{
			name: "multiple text elements",
			children: []ElementRenderer{
				&BaseElement{Token: "Hello"},
				&BaseElement{Token: " "},
				&BaseElement{Token: "World"},
			},
			want:    "Hello World",
			wantErr: false,
		},
		{
			name: "element with ANSI sequences",
			children: []ElementRenderer{
				&BaseElement{
					Token: "Styled",
					Style: StylePrimitive{
						Color: stringPtr("#ff0000"),
					},
				},
			},
			want:    "Styled",
			wantErr: false,
		},
		{
			name: "mixed elements with nil",
			children: []ElementRenderer{
				&BaseElement{Token: "First"},
				nil, // Should be skipped gracefully
				&BaseElement{Token: "Second"},
			},
			want:    "FirstSecond",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := Options{
				Styles: StyleConfig{
					Document: StyleBlock{},
					Link: StylePrimitive{
						Color: stringPtr("#00ff00"),
					},
					LinkText: StylePrimitive{
						Color: stringPtr("#ffffff"),
					},
				},
			}
			ctx := NewRenderContext(options)

			result, err := extractTextFromChildren(tt.children, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTextFromChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("extractTextFromChildren() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestApplyStyleToText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		style StylePrimitive
		want  string
	}{
		{
			name:  "empty text",
			text:  "",
			style: StylePrimitive{},
			want:  "",
		},
		{
			name:  "no style",
			text:  "Hello",
			style: StylePrimitive{},
			want:  "Hello", // Should still contain the text
		},
		{
			name: "colored text",
			text: "Hello",
			style: StylePrimitive{
				Color: stringPtr("#ff0000"),
			},
			want: "Hello", // Should contain the text (ANSI codes will be present too)
		},
		{
			name: "bold text",
			text: "Hello",
			style: StylePrimitive{
				Bold: boolPtr(true),
			},
			want: "Hello", // Should contain the text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := Options{
				Styles: StyleConfig{
					Document: StyleBlock{},
					Link: StylePrimitive{
						Color: stringPtr("#00ff00"),
					},
					LinkText: StylePrimitive{
						Color: stringPtr("#ffffff"),
					},
				},
			}
			ctx := NewRenderContext(options)

			result, err := applyStyleToText(tt.text, tt.style, ctx)
			if err != nil {
				t.Errorf("applyStyleToText() error = %v", err)
				return
			}

			// Strip ANSI to check if text is present
			plainResult := stripANSISequences(result)
			if plainResult != tt.want {
				t.Errorf("applyStyleToText() plain text = %q, want %q", plainResult, tt.want)
			}

			// Check that result contains the original text
			if tt.text != "" && !strings.Contains(result, tt.text) {
				t.Errorf("applyStyleToText() result %q should contain original text %q", result, tt.text)
			}
		})
	}
}

func TestHyperlinkStruct(t *testing.T) {
	t.Run("NewHyperlink", func(t *testing.T) {
		tests := []struct {
			name  string
			url   string
			text  string
			title string
			want  *Hyperlink
		}{
			{
				name:  "basic hyperlink",
				url:   "https://example.com",
				text:  "Example",
				title: "Example Site",
				want: &Hyperlink{
					URL:   "https://example.com",
					Text:  "Example",
					Title: "Example Site",
				},
			},
			{
				name: "with whitespace",
				url:  "  https://example.com  ",
				text: "  Example  ",
				want: &Hyperlink{
					URL:  "https://example.com",
					Text: "Example",
				},
			},
			{
				name: "with ANSI in text",
				url:  "https://example.com",
				text: "\x1b[31mRed Text\x1b[0m",
				want: &Hyperlink{
					URL:  "https://example.com",
					Text: "Red Text",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := NewHyperlink(tt.url, tt.text, tt.title)
				if result.URL != tt.want.URL {
					t.Errorf("NewHyperlink().URL = %q, want %q", result.URL, tt.want.URL)
				}
				if result.Text != tt.want.Text {
					t.Errorf("NewHyperlink().Text = %q, want %q", result.Text, tt.want.Text)
				}
				if result.Title != tt.want.Title {
					t.Errorf("NewHyperlink().Title = %q, want %q", result.Title, tt.want.Title)
				}
			})
		}
	})

	t.Run("RenderOSC8", func(t *testing.T) {
		h := &Hyperlink{
			URL:  "https://example.com",
			Text: "Example",
		}

		result := h.RenderOSC8()
		want := "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\"
		if result != want {
			t.Errorf("RenderOSC8() = %q, want %q", result, want)
		}
	})

	t.Run("RenderPlain", func(t *testing.T) {
		tests := []struct {
			name string
			h    *Hyperlink
			want string
		}{
			{
				name: "with text and URL",
				h: &Hyperlink{
					URL:  "https://example.com",
					Text: "Example",
				},
				want: "Example (https://example.com)",
			},
			{
				name: "URL only",
				h: &Hyperlink{
					URL:  "https://example.com",
					Text: "",
				},
				want: "https://example.com",
			},
			{
				name: "text only",
				h: &Hyperlink{
					URL:  "",
					Text: "Example",
				},
				want: "Example",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.h.RenderPlain()
				if result != tt.want {
					t.Errorf("RenderPlain() = %q, want %q", result, tt.want)
				}
			})
		}
	})

	t.Run("RenderSmart", func(t *testing.T) {
		h := &Hyperlink{
			URL:  "https://example.com",
			Text: "Example",
		}

		// Test with hyperlink support
		t.Setenv("TERM_PROGRAM", "iTerm.app")
		ctx := RenderContext{}
		result := h.RenderSmart(ctx)
		if !strings.Contains(result, "\x1b]8;;") {
			t.Errorf("RenderSmart() with hyperlink support should contain OSC 8 sequences, got %q", result)
		}

		// Test without hyperlink support
		t.Setenv("TERM_PROGRAM", "")
		t.Setenv("TERM", "")
		result = h.RenderSmart(ctx)
		if strings.Contains(result, "\x1b]8;;") {
			t.Errorf("RenderSmart() without hyperlink support should not contain OSC 8 sequences, got %q", result)
		}
		if !strings.Contains(result, "Example") || !strings.Contains(result, "https://example.com") {
			t.Errorf("RenderSmart() should contain both text and URL in plain format, got %q", result)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		tests := []struct {
			name    string
			h       *Hyperlink
			wantErr bool
		}{
			{
				name: "valid with both URL and text",
				h: &Hyperlink{
					URL:  "https://example.com",
					Text: "Example",
				},
				wantErr: false,
			},
			{
				name: "valid with URL only",
				h: &Hyperlink{
					URL:  "https://example.com",
					Text: "",
				},
				wantErr: false,
			},
			{
				name: "valid with text only",
				h: &Hyperlink{
					URL:  "",
					Text: "Example",
				},
				wantErr: false,
			},
			{
				name: "invalid - no URL or text",
				h: &Hyperlink{
					URL:  "",
					Text: "",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.h.Validate()
				if (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

// Helper function for bool pointers
func boolPtr(b bool) *bool {
	return &b
}
