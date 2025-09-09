package ansi

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestLinkFormatterInterface(t *testing.T) {
	// Test that LinkFormatterFunc implements LinkFormatter
	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		return "test", nil
	})

	var _ LinkFormatter = formatter

	// Test basic functionality
	result, err := formatter.FormatLink(LinkData{}, RenderContext{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "test" {
		t.Errorf("expected 'test', got %q", result)
	}
}

func TestLinkFormatterError(t *testing.T) {
	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		return "", errors.New("formatter error")
	})

	_, err := formatter.FormatLink(LinkData{}, RenderContext{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && err.Error() != "formatter error" {
		t.Errorf("expected 'formatter error', got %q", err.Error())
	}
}

func TestLinkData(t *testing.T) {
	tests := []struct {
		name string
		data LinkData
		want LinkData
	}{
		{
			name: "complete link data",
			data: LinkData{
				URL:        "https://example.com",
				Text:       "Example",
				Title:      "Example Site",
				BaseURL:    "https://base.com",
				IsAutoLink: false,
				IsInTable:  false,
			},
			want: LinkData{
				URL:        "https://example.com",
				Text:       "Example",
				Title:      "Example Site",
				BaseURL:    "https://base.com",
				IsAutoLink: false,
				IsInTable:  false,
			},
		},
		{
			name: "autolink data",
			data: LinkData{
				URL:        "https://example.com",
				Text:       "https://example.com",
				IsAutoLink: true,
			},
			want: LinkData{
				URL:        "https://example.com",
				Text:       "https://example.com",
				IsAutoLink: true,
			},
		},
		{
			name: "table link data",
			data: LinkData{
				URL:       "https://example.com",
				Text:      "Table Link",
				IsInTable: true,
			},
			want: LinkData{
				URL:       "https://example.com",
				Text:      "Table Link",
				IsInTable: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that all fields are properly set and accessible
			if tt.data.URL != tt.want.URL {
				t.Errorf("URL: expected %q, got %q", tt.want.URL, tt.data.URL)
			}
			if tt.data.Text != tt.want.Text {
				t.Errorf("Text: expected %q, got %q", tt.want.Text, tt.data.Text)
			}
			if tt.data.Title != tt.want.Title {
				t.Errorf("Title: expected %q, got %q", tt.want.Title, tt.data.Title)
			}
			if tt.data.BaseURL != tt.want.BaseURL {
				t.Errorf("BaseURL: expected %q, got %q", tt.want.BaseURL, tt.data.BaseURL)
			}
			if tt.data.IsAutoLink != tt.want.IsAutoLink {
				t.Errorf("IsAutoLink: expected %v, got %v", tt.want.IsAutoLink, tt.data.IsAutoLink)
			}
			if tt.data.IsInTable != tt.want.IsInTable {
				t.Errorf("IsInTable: expected %v, got %v", tt.want.IsInTable, tt.data.IsInTable)
			}
		})
	}
}

func TestDefaultFormatter(t *testing.T) {
	tests := []struct {
		name string
		data LinkData
		want string
	}{
		{
			name: "text and url",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			want: "Example https://example.com", // Current behavior
		},
		{
			name: "autolink",
			data: LinkData{
				URL:        "https://example.com",
				Text:       "https://example.com",
				IsAutoLink: true,
			},
			want: "https://example.com https://example.com",
		},
		{
			name: "empty text",
			data: LinkData{
				URL:  "https://example.com",
				Text: "",
			},
			want: "https://example.com", // Just URL when no text
		},
		{
			name: "fragment only URL",
			data: LinkData{
				URL:  "#fragment",
				Text: "Fragment Link",
			},
			want: "Fragment Link", // Fragment URLs are ignored
		},
		{
			name: "relative URL with base",
			data: LinkData{
				URL:     "/path/page",
				Text:    "Page",
				BaseURL: "https://example.com",
			},
			want: "Page https://example.com/path/page",
		},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DefaultFormatter.FormatLink(tt.data, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// Strip ANSI codes for easier comparison
			plainResult := stripANSISequences(result)
			if plainResult != tt.want {
				t.Errorf("expected %q, got %q", tt.want, plainResult)
			}
		})
	}
}

func TestTextOnlyFormatter(t *testing.T) {
	tests := []struct {
		name               string
		data               LinkData
		supportsHyperlinks bool
		wantContains       string
	}{
		{
			name: "hyperlink support",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			supportsHyperlinks: true,
			wantContains:       "\x1b]8;;https://example.com\x1b\\",
		},
		{
			name: "no hyperlink support",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			supportsHyperlinks: false,
			wantContains:       "Example", // Just text, no URL
		},
		{
			name: "empty text fallback",
			data: LinkData{
				URL:  "https://example.com",
				Text: "",
			},
			supportsHyperlinks: true,
			wantContains:       "", // Empty result for empty text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock terminal support by setting environment variable
			if tt.supportsHyperlinks {
				t.Setenv("TERM_PROGRAM", "iTerm.app")
			} else {
				t.Setenv("TERM_PROGRAM", "")
				t.Setenv("TERM", "")
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

			result, err := TextOnlyFormatter.FormatLink(tt.data, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("expected result to contain %q, got %q", tt.wantContains, result)
			}
		})
	}
}

func TestURLOnlyFormatter(t *testing.T) {
	tests := []struct {
		name string
		data LinkData
		want string
	}{
		{
			name: "normal link",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example Text",
			},
			want: "https://example.com",
		},
		{
			name: "fragment only URL",
			data: LinkData{
				URL:  "#fragment",
				Text: "Fragment",
			},
			want: "",
		},
		{
			name: "empty URL",
			data: LinkData{
				URL:  "",
				Text: "Text Only",
			},
			want: "",
		},
		{
			name: "relative URL with base",
			data: LinkData{
				URL:     "/path",
				Text:    "Path",
				BaseURL: "https://example.com",
			},
			want: "https://example.com/path",
		},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := URLOnlyFormatter.FormatLink(tt.data, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			plainResult := stripANSISequences(result)
			if plainResult != tt.want {
				t.Errorf("expected %q, got %q", tt.want, plainResult)
			}
		})
	}
}

func TestHyperlinkFormatter(t *testing.T) {
	tests := []struct {
		name string
		data LinkData
		want string
	}{
		{
			name: "normal link",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			want: "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\",
		},
		{
			name: "link with title",
			data: LinkData{
				URL:   "https://example.com",
				Text:  "Example",
				Title: "Example Site",
			},
			want: "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\",
		},
		{
			name: "empty text",
			data: LinkData{
				URL:  "https://example.com",
				Text: "",
			},
			want: "",
		},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HyperlinkFormatter.FormatLink(tt.data, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.want != "" {
				if !strings.Contains(result, tt.want) {
					t.Errorf("expected result to contain %q, got %q", tt.want, result)
				}
			} else {
				if result != "" {
					t.Errorf("expected empty result, got %q", result)
				}
			}
		})
	}
}

func TestSmartHyperlinkFormatter(t *testing.T) {
	tests := []struct {
		name               string
		data               LinkData
		supportsHyperlinks bool
		wantHyperlink      bool
	}{
		{
			name: "hyperlink support",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			supportsHyperlinks: true,
			wantHyperlink:      true,
		},
		{
			name: "no hyperlink support",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			supportsHyperlinks: false,
			wantHyperlink:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock terminal support
			if tt.supportsHyperlinks {
				t.Setenv("TERM_PROGRAM", "iTerm.app")
			} else {
				t.Setenv("TERM_PROGRAM", "")
				t.Setenv("TERM", "")
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

			result, err := SmartHyperlinkFormatter.FormatLink(tt.data, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantHyperlink {
				if !strings.Contains(result, "\x1b]8;;") {
					t.Errorf("expected hyperlink sequence, got %q", result)
				}
			} else {
				// Should fall back to default format
				plainResult := stripANSISequences(result)
				if !strings.Contains(plainResult, "Example") {
					t.Errorf("expected 'Example' in result, got %q", plainResult)
				}
				if !strings.Contains(plainResult, "https://example.com") {
					t.Errorf("expected 'https://example.com' in result, got %q", plainResult)
				}
			}
		})
	}
}

func TestFormatterErrorHandling(t *testing.T) {
	errorFormatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		return "", errors.New("formatter error")
	})

	_, err := errorFormatter.FormatLink(LinkData{}, RenderContext{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "formatter error") {
		t.Errorf("expected error to contain 'formatter error', got %q", err.Error())
	}
}

func TestInvalidURLHandling(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"malformed URL", "://invalid"},
		{"just fragment", "#fragment"},
		{"just query", "?param=value"},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := LinkData{
				URL:  tt.url,
				Text: "example",
			}

			// Should not panic
			result, err := DefaultFormatter.FormatLink(data, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestLinkInTableContext(t *testing.T) {
	data := LinkData{
		URL:       "https://example.com",
		Text:      "example",
		IsInTable: true,
	}

	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		if data.IsInTable {
			return data.Text, nil // Table links: text only
		}
		return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
	})

	result, err := formatter.FormatLink(data, RenderContext{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "example" {
		t.Errorf("expected 'example', got %q", result)
	}
	if strings.Contains(result, "https://example.com") {
		t.Errorf("result should not contain URL, got %q", result)
	}
}

func TestAutoLinkContext(t *testing.T) {
	data := LinkData{
		URL:        "https://example.com",
		Text:       "https://example.com",
		IsAutoLink: true,
	}

	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
		if data.IsAutoLink {
			return fmt.Sprintf("<%s>", data.URL), nil
		}
		return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
	})

	result, err := formatter.FormatLink(data, RenderContext{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "<https://example.com>" {
		t.Errorf("expected '<https://example.com>', got %q", result)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
