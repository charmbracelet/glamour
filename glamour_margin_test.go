package glamour

import (
	"strings"
	"testing"

	"github.com/charmbracelet/glamour/ansi"
)

func TestWithMargins(t *testing.T) {
	tests := []struct {
		name         string
		leftMargin   uint
		rightMargin  uint
		markdown     string
		expectedLeft string // Expected left margin pattern
		width        int
	}{
		{
			name:         "Basic left margin",
			leftMargin:   10,
			rightMargin:  0,
			markdown:     "# Test\n\nSimple paragraph.",
			expectedLeft: "          ", // 10 spaces
			width:        80,
		},
		{
			name:         "Large left margin",
			leftMargin:   20,
			rightMargin:  20,
			markdown:     "# Test\n\nSimple paragraph text that should wrap with proper indentation.",
			expectedLeft: "                    ", // 20 spaces
			width:        80,
		},
		{
			name:         "Zero margins",
			leftMargin:   0,
			rightMargin:  0,
			markdown:     "# Test\n\nNo margins applied.",
			expectedLeft: "", // No margin
			width:        80,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create renderer with margins
			r, err := NewTermRenderer(
				WithWordWrap(test.width),
				WithMargins(test.leftMargin, test.rightMargin),
			)
			if err != nil {
				t.Fatalf("Failed to create renderer: %v", err)
			}

			// Render the markdown
			output, err := r.Render(test.markdown)
			if err != nil {
				t.Fatalf("Failed to render markdown: %v", err)
			}

			lines := strings.Split(output, "\n")
			
			// Check that non-empty lines start with expected left margin
			foundMarginedLine := false
			for _, line := range lines {
				if strings.TrimSpace(line) != "" { // Skip empty lines
					if test.leftMargin > 0 {
						if !strings.HasPrefix(line, test.expectedLeft) {
							t.Errorf("Line does not start with expected margin.\nExpected prefix: %q\nActual line: %q", test.expectedLeft, line)
						} else {
							foundMarginedLine = true
						}
					}
				}
			}

			if test.leftMargin > 0 && !foundMarginedLine {
				t.Errorf("No lines found with expected left margin.\nOutput:\n%s", output)
			}
		})
	}
}

func TestWithJustifiedAlignment(t *testing.T) {
	tests := []struct {
		name        string
		leftMargin  uint
		rightMargin uint
		markdown    string
		width       int
	}{
		{
			name:        "Justified with margins",
			leftMargin:  10,
			rightMargin: 10,
			markdown:    "This is a long paragraph that should be justified with proper spacing between words to fill the available width.",
			width:       80,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := NewTermRenderer(
				WithWordWrap(test.width),
				WithJustifiedAlignment(test.leftMargin, test.rightMargin),
			)
			if err != nil {
				t.Fatalf("Failed to create renderer: %v", err)
			}

			output, err := r.Render(test.markdown)
			if err != nil {
				t.Fatalf("Failed to render markdown: %v", err)
			}

			// Basic check - output should not be empty
			if strings.TrimSpace(output) == "" {
				t.Error("Justified alignment produced empty output")
			}

			t.Logf("Justified output:\n%s", output)
		})
	}
}

func TestWithCenterAlignment(t *testing.T) {
	tests := []struct {
		name        string
		leftMargin  uint
		rightMargin uint
		markdown    string
		width       int
	}{
		{
			name:        "Center alignment with margins",
			leftMargin:  5,
			rightMargin: 5,
			markdown:    "# Centered Title\n\nThis paragraph should be centered.",
			width:       80,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := NewTermRenderer(
				WithWordWrap(test.width),
				WithCenterAlignment(test.leftMargin, test.rightMargin),
			)
			if err != nil {
				t.Fatalf("Failed to create renderer: %v", err)
			}

			output, err := r.Render(test.markdown)
			if err != nil {
				t.Fatalf("Failed to render markdown: %v", err)
			}

			// Basic check - output should not be empty
			if strings.TrimSpace(output) == "" {
				t.Error("Center alignment produced empty output")
			}

			t.Logf("Centered output:\n%s", output)
		})
	}
}

// Test the WordwrapWithIndent function directly
func TestWordwrapWithIndent(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		width      int
		breakChars string
		indent     string
		expected   []string
	}{
		{
			name:       "Basic wrapping with indent",
			text:       "This is a long line that should wrap with proper indentation",
			width:      20,
			breakChars: " ",
			indent:     "  ",
			expected: []string{
				"This is a long line",
				"  that should wrap",
				"  with proper",
				"  indentation",
			},
		},
		{
			name:       "Short text no wrap",
			text:       "Short",
			width:      20,
			breakChars: " ",
			indent:     "  ",
			expected:   []string{"Short"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ansi.WordwrapWithIndent(test.text, test.width, test.breakChars, test.indent)
			lines := strings.Split(result, "\n")

			if len(lines) != len(test.expected) {
				t.Errorf("Expected %d lines, got %d.\nExpected: %v\nActual: %v", 
					len(test.expected), len(lines), test.expected, lines)
				return
			}

			for i, line := range lines {
				if line != test.expected[i] {
					t.Errorf("Line %d mismatch.\nExpected: %q\nActual: %q", i, test.expected[i], line)
				}
			}
		})
	}
}