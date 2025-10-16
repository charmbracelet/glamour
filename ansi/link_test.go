package ansi

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func TestLinkElementWithCustomFormatter(t *testing.T) {
	tests := []struct {
		name      string
		element   *LinkElement
		formatter LinkFormatter
		want      string
	}{
		{
			name: "custom formatter with text and URL",
			element: &LinkElement{
				URL:      "https://example.com",
				Children: []ElementRenderer{&BaseElement{Token: "example"}},
			},
			formatter: LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
				return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
			}),
			want: "[example](https://example.com)",
		},
		{
			name: "custom formatter text only",
			element: &LinkElement{
				URL:      "https://example.com",
				Children: []ElementRenderer{&BaseElement{Token: "click here"}},
			},
			formatter: LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
				return data.Text, nil
			}),
			want: "click here",
		},
		{
			name: "custom formatter with title",
			element: &LinkElement{
				URL:      "https://example.com",
				Title:    "Example Site",
				Children: []ElementRenderer{&BaseElement{Token: "example"}},
			},
			formatter: LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
				if data.Title != "" {
					return fmt.Sprintf("%s (%s - %s)", data.Text, data.URL, data.Title), nil
				}
				return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
			}),
			want: "example (https://example.com - Example Site)",
		},
		{
			name: "autolink context",
			element: &LinkElement{
				URL:        "https://example.com",
				Children:   []ElementRenderer{&BaseElement{Token: "https://example.com"}},
				IsAutoLink: true,
			},
			formatter: LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
				if data.IsAutoLink {
					return fmt.Sprintf("<%s>", data.URL), nil
				}
				return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
			}),
			want: "<https://example.com>",
		},
		{
			name: "table context",
			element: &LinkElement{
				URL:       "https://example.com",
				Children:  []ElementRenderer{&BaseElement{Token: "link"}},
				IsInTable: true,
			},
			formatter: LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
				if data.IsInTable {
					return data.Text, nil // Tables: text only
				}
				return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
			}),
			want: "link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.element.Formatter = tt.formatter

			var buf bytes.Buffer
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

			err := tt.element.Render(&buf, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			result := buf.String()
			// Strip ANSI codes for easier comparison
			plainResult := stripANSISequences(result)
			if plainResult != tt.want {
				t.Errorf("expected %q, got %q", tt.want, plainResult)
			}
		})
	}
}

func TestLinkElementWithoutFormatter(t *testing.T) {
	tests := []struct {
		name    string
		element *LinkElement
		want    []string // Multiple strings that should be present in output
	}{
		{
			name: "default behavior with text and URL",
			element: &LinkElement{
				URL:      "https://example.com",
				Children: []ElementRenderer{&BaseElement{Token: "example"}},
			},
			want: []string{"example", "https://example.com"},
		},
		{
			name: "with base URL",
			element: &LinkElement{
				URL:      "/path",
				BaseURL:  "https://example.com",
				Children: []ElementRenderer{&BaseElement{Token: "path"}},
			},
			want: []string{"path", "https://example.com/path"},
		},
		{
			name: "fragment only URL (should be ignored)",
			element: &LinkElement{
				URL:      "#fragment",
				Children: []ElementRenderer{&BaseElement{Token: "fragment"}},
			},
			want: []string{"fragment"}, // URL should be ignored
		},
		{
			name: "skip text",
			element: &LinkElement{
				URL:      "https://example.com",
				Children: []ElementRenderer{&BaseElement{Token: "example"}},
				SkipText: true,
			},
			want: []string{"https://example.com"}, // Only URL
		},
		{
			name: "skip href",
			element: &LinkElement{
				URL:      "https://example.com",
				Children: []ElementRenderer{&BaseElement{Token: "example"}},
				SkipHref: true,
			},
			want: []string{"example"}, // Only text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
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

			err := tt.element.Render(&buf, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			result := buf.String()
			plainResult := stripANSISequences(result)

			for _, expected := range tt.want {
				if !strings.Contains(plainResult, expected) {
					t.Errorf("expected result to contain %q, got %q", expected, plainResult)
				}
			}
		})
	}
}

func TestLinkElementFormatterError(t *testing.T) {
	errorFormatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		return "", errors.New("formatter error")
	})

	element := &LinkElement{
		URL:       "https://example.com",
		Children:  []ElementRenderer{&BaseElement{Token: "example"}},
		Formatter: errorFormatter,
	}

	var buf bytes.Buffer
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
	err := element.Render(&buf, ctx)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "formatter error") {
		t.Errorf("expected error to contain 'formatter error', got %q", err.Error())
	}
}

func TestLinkElementDataExtraction(t *testing.T) {
	tests := []struct {
		name    string
		element *LinkElement
		check   func(t *testing.T, data LinkData)
	}{
		{
			name: "basic data extraction",
			element: &LinkElement{
				URL:        "https://example.com",
				Title:      "Example Site",
				BaseURL:    "https://base.com",
				IsAutoLink: true,
				IsInTable:  true,
				Children:   []ElementRenderer{&BaseElement{Token: "example"}},
			},
			check: func(t *testing.T, data LinkData) {
				if data.URL != "https://example.com" {
					t.Errorf("URL: expected %q, got %q", "https://example.com", data.URL)
				}
				if data.Text != "example" {
					t.Errorf("Text: expected %q, got %q", "example", data.Text)
				}
				if data.Title != "Example Site" {
					t.Errorf("Title: expected %q, got %q", "Example Site", data.Title)
				}
				if data.BaseURL != "https://base.com" {
					t.Errorf("BaseURL: expected %q, got %q", "https://base.com", data.BaseURL)
				}
				if !data.IsAutoLink {
					t.Error("IsAutoLink: expected true")
				}
				if !data.IsInTable {
					t.Error("IsInTable: expected true")
				}
			},
		},
		{
			name: "multiple children",
			element: &LinkElement{
				URL: "https://example.com",
				Children: []ElementRenderer{
					&BaseElement{Token: "Hello"},
					&BaseElement{Token: " "},
					&BaseElement{Token: "World"},
				},
			},
			check: func(t *testing.T, data LinkData) {
				if data.Text != "Hello World" {
					t.Errorf("Text: expected %q, got %q", "Hello World", data.Text)
				}
			},
		},
		{
			name: "styled children",
			element: &LinkElement{
				URL: "https://example.com",
				Children: []ElementRenderer{
					&BaseElement{
						Token: "Styled",
						Style: StylePrimitive{
							Color: stringPtr("#ff0000"),
						},
					},
				},
			},
			check: func(t *testing.T, data LinkData) {
				if data.Text != "Styled" {
					t.Errorf("Text: expected %q, got %q", "Styled", data.Text)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedData LinkData
			formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
				capturedData = data
				return "test", nil
			})

			tt.element.Formatter = formatter

			var buf bytes.Buffer
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

			err := tt.element.Render(&buf, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			tt.check(t, capturedData)
		})
	}
}

func TestLinkElementStyleContext(t *testing.T) {
	customStyles := StyleConfig{
		Link: StylePrimitive{
			Color: stringPtr("#00ff00"),
			Bold:  boolPtr(true),
		},
		LinkText: StylePrimitive{
			Color:     stringPtr("#0000ff"),
			Underline: boolPtr(true),
		},
	}

	var capturedData LinkData
	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		capturedData = data
		return "test", nil
	})

	element := &LinkElement{
		URL:       "https://example.com",
		Children:  []ElementRenderer{&BaseElement{Token: "example"}},
		Formatter: formatter,
	}

	var buf bytes.Buffer
	options := Options{
		Styles: customStyles,
	}
	ctx := NewRenderContext(options)

	err := element.Render(&buf, ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that styles were passed correctly
	if capturedData.LinkStyle.Color == nil || *capturedData.LinkStyle.Color != "#00ff00" {
		t.Errorf("LinkStyle.Color: expected %q, got %v", "#00ff00", capturedData.LinkStyle.Color)
	}
	if capturedData.TextStyle.Color == nil || *capturedData.TextStyle.Color != "#0000ff" {
		t.Errorf("TextStyle.Color: expected %q, got %v", "#0000ff", capturedData.TextStyle.Color)
	}
}

func TestLinkElementComplexChildren(t *testing.T) {
	// Test with nested elements that might have complex rendering
	element := &LinkElement{
		URL: "https://example.com",
		Children: []ElementRenderer{
			&BaseElement{
				Token: "Bold ",
				Style: StylePrimitive{Bold: boolPtr(true)},
			},
			&BaseElement{
				Token: "and ",
			},
			&BaseElement{
				Token: "Italic",
				Style: StylePrimitive{Italic: boolPtr(true)},
			},
		},
		Formatter: LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
			return fmt.Sprintf("TEXT:%s URL:%s", data.Text, data.URL), nil
		}),
	}

	var buf bytes.Buffer
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

	err := element.Render(&buf, ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	result := stripANSISequences(buf.String())
	expected := "TEXT:Bold and Italic URL:https://example.com"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestLinkElementNilChildren(t *testing.T) {
	element := &LinkElement{
		URL: "https://example.com",
		Children: []ElementRenderer{
			&BaseElement{Token: "First"},
			nil, // Should be handled gracefully
			&BaseElement{Token: "Second"},
		},
		Formatter: LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
			return fmt.Sprintf("TEXT:%s", data.Text), nil
		}),
	}

	var buf bytes.Buffer
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

	err := element.Render(&buf, ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	result := stripANSISequences(buf.String())
	expected := "TEXT:FirstSecond"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestLinkElementRenderContext(t *testing.T) {
	var capturedContext RenderContext
	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		capturedContext = ctx
		return "test", nil
	})

	element := &LinkElement{
		URL:       "https://example.com",
		Children:  []ElementRenderer{&BaseElement{Token: "example"}},
		Formatter: formatter,
	}

	var buf bytes.Buffer
	options := Options{
		ColorProfile: 256, // Example context data
		WordWrap:     80,
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

	err := element.Render(&buf, ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify context was passed through
	if capturedContext.options.ColorProfile != 256 {
		t.Errorf("ColorProfile: expected 256, got %v", capturedContext.options.ColorProfile)
	}
	if capturedContext.options.WordWrap != 80 {
		t.Errorf("WordWrap: expected 80, got %v", capturedContext.options.WordWrap)
	}
}

// Benchmark tests for performance
func BenchmarkLinkElementDefault(b *testing.B) {
	element := &LinkElement{
		URL:      "https://example.com",
		Children: []ElementRenderer{&BaseElement{Token: "example"}},
	}

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		element.Render(&buf, ctx)
	}
}

func BenchmarkLinkElementCustomFormatter(b *testing.B) {
	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
	})

	element := &LinkElement{
		URL:       "https://example.com",
		Children:  []ElementRenderer{&BaseElement{Token: "example"}},
		Formatter: formatter,
	}

	bs := &BlockStack{}
	bs.Push(BlockElement{Style: StyleBlock{}})

	ctx := RenderContext{
		blockStack: bs,
		options: Options{
			ColorProfile: termenv.TrueColor,
			WordWrap:     80,
			Styles: StyleConfig{
				Document: StyleBlock{},
				Link: StylePrimitive{
					Color: stringPtr("#00ff00"),
				},
				LinkText: StylePrimitive{
					Color: stringPtr("#ffffff"),
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		element.Render(&buf, ctx)
	}
}
