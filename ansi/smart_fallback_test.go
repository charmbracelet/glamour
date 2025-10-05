package ansi

import (
	"strings"
	"testing"
)

// TestSmartHyperlinkFormatterFallback tests the smart formatter's fallback behavior
func TestSmartHyperlinkFormatterFallback(t *testing.T) {
	tests := []struct {
		name            string
		data            LinkData
		termProgram     string
		term            string
		expectHyperlink bool
		expectPlainText bool
		description     string
	}{
		{
			name: "iTerm2 - should use hyperlinks",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			termProgram:     "iTerm.app",
			expectHyperlink: true,
			expectPlainText: false,
			description:     "iTerm2 supports OSC 8, should generate hyperlink sequence",
		},
		{
			name: "VS Code - should use hyperlinks",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			termProgram:     "vscode",
			expectHyperlink: true,
			expectPlainText: false,
			description:     "VS Code supports OSC 8, should generate hyperlink sequence",
		},
		{
			name: "Unknown terminal - should fallback to plain text",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			termProgram:     "unknown",
			term:            "dumb",
			expectHyperlink: false,
			expectPlainText: true,
			description:     "Unknown terminal should fallback to text + URL format",
		},
		{
			name: "Basic xterm - should fallback to plain text",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			term:            "xterm",
			expectHyperlink: false,
			expectPlainText: true,
			description:     "Basic xterm doesn't support hyperlinks, should fallback",
		},
		{
			name: "Empty environment - should fallback to plain text",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			expectHyperlink: false,
			expectPlainText: true,
			description:     "Empty environment should fallback to plain text",
		},
		{
			name: "Text only with hyperlink support",
			data: LinkData{
				URL:  "https://example.com",
				Text: "",
			},
			termProgram:     "iTerm.app",
			expectHyperlink: false,
			expectPlainText: false,
			description:     "Empty text should return empty result even with hyperlink support",
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

			// Clear other terminal detection variables
			t.Setenv("KITTY_WINDOW_ID", "")
			t.Setenv("ALACRITTY_LOG", "")
			t.Setenv("ALACRITTY_SOCKET", "")

			ctx := NewRenderContext(Options{})
			result, err := SmartHyperlinkFormatter.FormatLink(tt.data, ctx)

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}

			if tt.expectHyperlink {
				if !strings.Contains(result, "\x1b]8;;") {
					t.Errorf("%s: expected OSC 8 hyperlink sequence, got: %q", tt.description, result)
				}
				// Should not contain URL separately when using hyperlinks (unless fallback)
				plainResult := stripANSISequences(result)
				if strings.Contains(plainResult, tt.data.URL) && tt.data.URL != "" {
					t.Errorf("%s: hyperlink mode should not show URL separately, got: %q", tt.description, plainResult)
				}
			}

			if tt.expectPlainText {
				if strings.Contains(result, "\x1b]8;;") {
					t.Errorf("%s: should not contain hyperlink sequences in fallback mode, got: %q", tt.description, result)
				}

				// Should contain both text and URL for fallback (if both exist)
				plainResult := stripANSISequences(result)
				if tt.data.Text != "" && !strings.Contains(plainResult, tt.data.Text) {
					t.Errorf("%s: fallback should contain text %q, got: %q", tt.description, tt.data.Text, plainResult)
				}
				if tt.data.URL != "" && !strings.Contains(plainResult, tt.data.URL) {
					t.Errorf("%s: fallback should contain URL %q, got: %q", tt.description, tt.data.URL, plainResult)
				}
			}

			// Handle case where empty result is expected
			if !tt.expectHyperlink && !tt.expectPlainText {
				plainResult := stripANSISequences(result)
				if plainResult != "" {
					t.Errorf("%s: expected empty result, got: %q", tt.description, plainResult)
				}
			}
		})
	}
}

// TestTextOnlyFormatterFallback tests TextOnlyFormatter's fallback behavior
func TestTextOnlyFormatterFallback(t *testing.T) {
	tests := []struct {
		name            string
		data            LinkData
		termProgram     string
		expectHyperlink bool
		description     string
	}{
		{
			name: "With hyperlink support - should hide URL",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			termProgram:     "iTerm.app",
			expectHyperlink: true,
			description:     "TextOnly formatter should use hyperlinks when supported",
		},
		{
			name: "Without hyperlink support - should show only text",
			data: LinkData{
				URL:  "https://example.com",
				Text: "Example",
			},
			termProgram:     "unknown",
			expectHyperlink: false,
			description:     "TextOnly formatter should show only text without support",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.termProgram != "" {
				t.Setenv("TERM_PROGRAM", tt.termProgram)
			} else {
				t.Setenv("TERM_PROGRAM", "")
			}
			t.Setenv("TERM", "")

			ctx := NewRenderContext(Options{})
			result, err := TextOnlyFormatter.FormatLink(tt.data, ctx)

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}

			plainResult := stripANSISequences(result)

			if tt.expectHyperlink {
				if !strings.Contains(result, "\x1b]8;;") {
					t.Errorf("%s: expected hyperlink sequence, got: %q", tt.description, result)
				}
				// Should not show URL separately
				if strings.Contains(plainResult, tt.data.URL) {
					t.Errorf("%s: should not show URL in hyperlink mode, got: %q", tt.description, plainResult)
				}
			} else {
				// Should not show URL at all in text-only mode
				if strings.Contains(plainResult, tt.data.URL) {
					t.Errorf("%s: should not show URL in text-only fallback, got: %q", tt.description, plainResult)
				}
			}

			// Should always show text (if present)
			if tt.data.Text != "" && !strings.Contains(plainResult, tt.data.Text) {
				t.Errorf("%s: should always show text, got: %q", tt.description, plainResult)
			}
		})
	}
}

// TestHyperlinkStructSmartRendering tests Hyperlink struct's smart rendering
func TestHyperlinkStructSmartRendering(t *testing.T) {
	tests := []struct {
		name            string
		hyperlink       *Hyperlink
		termProgram     string
		expectHyperlink bool
		expectPlain     bool
		description     string
	}{
		{
			name: "Smart rendering with hyperlink support",
			hyperlink: &Hyperlink{
				URL:  "https://example.com",
				Text: "Example",
			},
			termProgram:     "iTerm.app",
			expectHyperlink: true,
			expectPlain:     false,
			description:     "Should use OSC 8 when supported",
		},
		{
			name: "Smart rendering without hyperlink support",
			hyperlink: &Hyperlink{
				URL:  "https://example.com",
				Text: "Example",
			},
			termProgram:     "unknown",
			expectHyperlink: false,
			expectPlain:     true,
			description:     "Should fallback to plain text format",
		},
		{
			name: "Smart rendering with empty text",
			hyperlink: &Hyperlink{
				URL:  "https://example.com",
				Text: "",
			},
			termProgram:     "iTerm.app",
			expectHyperlink: true, // Still generates hyperlink with empty text
			expectPlain:     false,
			description:     "Should generate hyperlink even with empty text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.termProgram != "" {
				t.Setenv("TERM_PROGRAM", tt.termProgram)
			} else {
				t.Setenv("TERM_PROGRAM", "")
			}
			t.Setenv("TERM", "")

			ctx := NewRenderContext(Options{})
			result := tt.hyperlink.RenderSmart(ctx)

			if tt.expectHyperlink {
				if !strings.Contains(result, "\x1b]8;;") {
					t.Errorf("%s: expected OSC 8 sequence, got: %q", tt.description, result)
				}
			}

			if tt.expectPlain {
				if strings.Contains(result, "\x1b]8;;") {
					t.Errorf("%s: should not contain OSC 8 in plain mode, got: %q", tt.description, result)
				}

				// Plain mode should show both text and URL (if both exist)
				if tt.hyperlink.Text != "" && tt.hyperlink.URL != "" {
					expected := tt.hyperlink.Text + " (" + tt.hyperlink.URL + ")"
					if result != expected {
						t.Errorf("%s: expected %q, got %q", tt.description, expected, result)
					}
				}
			}
		})
	}
}

// TestFallbackWithStyling tests that fallback still applies styling correctly
func TestFallbackWithStyling(t *testing.T) {
	data := LinkData{
		URL:  "https://example.com",
		Text: "Example",
		LinkStyle: StylePrimitive{
			Color: stringPtr("#ff0000"),
		},
		TextStyle: StylePrimitive{
			Color: stringPtr("#00ff00"),
		},
	}

	// Test with unsupported terminal
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "dumb")

	ctx := NewRenderContext(Options{})
	result, err := SmartHyperlinkFormatter.FormatLink(data, ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should fallback to default formatter but still apply styling
	if strings.Contains(result, "\x1b]8;;") {
		t.Error("Should not contain hyperlink sequences in fallback")
	}

	// Result should contain styled content (ANSI color codes)
	if !strings.Contains(result, "\x1b[") {
		t.Error("Fallback should still apply styling with ANSI codes")
	}

	// Should contain both text and URL
	plainResult := stripANSISequences(result)
	if !strings.Contains(plainResult, "Example") {
		t.Errorf("Should contain text, got: %q", plainResult)
	}
	if !strings.Contains(plainResult, "https://example.com") {
		t.Errorf("Should contain URL, got: %q", plainResult)
	}
}

// TestFallbackEdgeCases tests edge cases in fallback behavior
func TestFallbackEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		data        LinkData
		description string
		expectEmpty bool
	}{
		{
			name: "Empty text and URL",
			data: LinkData{
				URL:  "",
				Text: "",
			},
			description: "Both empty should produce empty result",
			expectEmpty: true,
		},
		{
			name: "Only URL",
			data: LinkData{
				URL:  "https://example.com",
				Text: "",
			},
			description: "URL only should show just URL",
			expectEmpty: false,
		},
		{
			name: "Only text",
			data: LinkData{
				URL:  "",
				Text: "Just text",
			},
			description: "Text only should show just text",
			expectEmpty: false,
		},
		{
			name: "Fragment URL",
			data: LinkData{
				URL:  "#fragment",
				Text: "Fragment",
			},
			description: "Fragment URLs should be handled specially",
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use unsupported terminal to force fallback
			t.Setenv("TERM_PROGRAM", "")
			t.Setenv("TERM", "dumb")

			ctx := NewRenderContext(Options{})
			result, err := SmartHyperlinkFormatter.FormatLink(tt.data, ctx)

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}

			plainResult := stripANSISequences(result)

			if tt.expectEmpty && plainResult != "" {
				t.Errorf("%s: expected empty result, got %q", tt.description, plainResult)
			}

			if !tt.expectEmpty && plainResult == "" {
				t.Errorf("%s: expected non-empty result, got empty", tt.description)
			}
		})
	}
}

// TestTerminalDetectionConsistency ensures consistent behavior across formatters
func TestTerminalDetectionConsistency(t *testing.T) {
	data := LinkData{
		URL:  "https://example.com",
		Text: "Example",
	}

	terminals := []struct {
		name          string
		termProgram   string
		term          string
		shouldSupport bool
	}{
		{"iTerm2", "iTerm.app", "", true},
		{"VS Code", "vscode", "", true},
		{"Windows Terminal", "Windows Terminal", "", true},
		{"xterm-256color", "", "xterm-256color", true},
		{"basic xterm", "", "xterm", false},
		{"dumb terminal", "", "dumb", false},
		{"empty env", "", "", false},
	}

	for _, term := range terminals {
		t.Run(term.name, func(t *testing.T) {
			t.Setenv("TERM_PROGRAM", term.termProgram)
			t.Setenv("TERM", term.term)
			t.Setenv("KITTY_WINDOW_ID", "")
			t.Setenv("ALACRITTY_LOG", "")

			ctx := NewRenderContext(Options{})

			// Test direct detection
			detected := supportsHyperlinks(ctx)
			if detected != term.shouldSupport {
				t.Errorf("Detection mismatch for %s: got %v, expected %v", term.name, detected, term.shouldSupport)
			}

			// Test SmartHyperlinkFormatter consistency
			smartResult, _ := SmartHyperlinkFormatter.FormatLink(data, ctx)
			containsHyperlink := strings.Contains(smartResult, "\x1b]8;;")

			if containsHyperlink != term.shouldSupport {
				t.Errorf("SmartHyperlinkFormatter inconsistent for %s: hyperlink=%v, shouldSupport=%v",
					term.name, containsHyperlink, term.shouldSupport)
			}

			// Test Hyperlink struct consistency
			h := NewHyperlink(data.URL, data.Text, "")
			structResult := h.RenderSmart(ctx)
			structContainsHyperlink := strings.Contains(structResult, "\x1b]8;;")

			if structContainsHyperlink != term.shouldSupport {
				t.Errorf("Hyperlink.RenderSmart inconsistent for %s: hyperlink=%v, shouldSupport=%v",
					term.name, structContainsHyperlink, term.shouldSupport)
			}
		})
	}
}

// BenchmarkSmartFallbackPerformance benchmarks the performance of smart fallback
func BenchmarkSmartFallbackPerformance(b *testing.B) {
	data := LinkData{
		URL:  "https://example.com",
		Text: "Example Link",
	}

	scenarios := map[string]string{
		"with_hyperlinks":    "iTerm.app",
		"without_hyperlinks": "unknown",
	}

	for name, termProgram := range scenarios {
		b.Run(name, func(b *testing.B) {
			b.Setenv("TERM_PROGRAM", termProgram)
			ctx := NewRenderContext(Options{})

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				SmartHyperlinkFormatter.FormatLink(data, ctx)
			}
		})
	}
}
