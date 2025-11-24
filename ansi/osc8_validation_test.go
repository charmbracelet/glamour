package ansi

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestOSC8SequenceFormat validates that OSC 8 sequences are generated correctly
func TestOSC8SequenceFormat(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		url      string
		expected string
	}{
		{
			name:     "basic HTTP URL",
			text:     "Example",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\",
		},
		{
			name:     "HTTPS URL",
			text:     "Secure Site",
			url:      "https://secure.example.com",
			expected: "\x1b]8;;https://secure.example.com\x1b\\Secure Site\x1b]8;;\x1b\\",
		},
		{
			name:     "HTTP URL",
			text:     "Plain HTTP",
			url:      "http://plain.example.com",
			expected: "\x1b]8;;http://plain.example.com\x1b\\Plain HTTP\x1b]8;;\x1b\\",
		},
		{
			name:     "URL with path and query",
			text:     "Complex URL",
			url:      "https://example.com/path/to/page?param=value&other=test",
			expected: "\x1b]8;;https://example.com/path/to/page?param=value&other=test\x1b\\Complex URL\x1b]8;;\x1b\\",
		},
		{
			name:     "URL with fragment",
			text:     "Section Link",
			url:      "https://example.com/page#section",
			expected: "\x1b]8;;https://example.com/page#section\x1b\\Section Link\x1b]8;;\x1b\\",
		},
		{
			name:     "relative URL",
			text:     "Relative",
			url:      "/relative/path",
			expected: "\x1b]8;;/relative/path\x1b\\Relative\x1b]8;;\x1b\\",
		},
		{
			name:     "mailto URL",
			text:     "Email Link",
			url:      "mailto:user@example.com",
			expected: "\x1b]8;;mailto:user@example.com\x1b\\Email Link\x1b]8;;\x1b\\",
		},
		{
			name:     "file URL",
			text:     "Local File",
			url:      "file:///path/to/file.txt",
			expected: "\x1b]8;;file:///path/to/file.txt\x1b\\Local File\x1b]8;;\x1b\\",
		},
		{
			name:     "ftp URL",
			text:     "FTP Server",
			url:      "ftp://ftp.example.com/file.zip",
			expected: "\x1b]8;;ftp://ftp.example.com/file.zip\x1b\\FTP Server\x1b]8;;\x1b\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHyperlink(tt.text, tt.url)
			if result != tt.expected {
				t.Errorf("formatHyperlink(%q, %q) = %q, expected %q", tt.text, tt.url, result, tt.expected)
			}
		})
	}
}

// TestOSC8SequenceTextVariations tests various text content in hyperlinks
func TestOSC8SequenceTextVariations(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		url      string
		expected string
	}{
		{
			name:     "empty text",
			text:     "",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\\x1b]8;;\x1b\\",
		},
		{
			name:     "single character",
			text:     "X",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\X\x1b]8;;\x1b\\",
		},
		{
			name:     "text with spaces",
			text:     "Click here for more info",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\Click here for more info\x1b]8;;\x1b\\",
		},
		{
			name:     "text with punctuation",
			text:     "Hello, World!",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\Hello, World!\x1b]8;;\x1b\\",
		},
		{
			name:     "text with numbers",
			text:     "Version 1.2.3",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\Version 1.2.3\x1b]8;;\x1b\\",
		},
		{
			name:     "text with special characters",
			text:     "Cost: $19.99 (50% off!)",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\Cost: $19.99 (50% off!)\x1b]8;;\x1b\\",
		},
		{
			name:     "unicode text",
			text:     "例え",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\例え\x1b]8;;\x1b\\",
		},
		{
			name:     "emoji text",
			text:     "Click here! 👉",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\Click here! 👉\x1b]8;;\x1b\\",
		},
		{
			name:     "mixed unicode and ascii",
			text:     "Hello 世界 World",
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\Hello 世界 World\x1b]8;;\x1b\\",
		},
		{
			name:     "long text",
			text:     strings.Repeat("A", 100),
			url:      "https://example.com",
			expected: "\x1b]8;;https://example.com\x1b\\" + strings.Repeat("A", 100) + "\x1b]8;;\x1b\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHyperlink(tt.text, tt.url)
			if result != tt.expected {
				t.Errorf("formatHyperlink(%q, %q) = %q, expected %q", tt.text, tt.url, result, tt.expected)
			}
		})
	}
}

// TestOSC8SequenceURLVariations tests various URL formats
func TestOSC8SequenceURLVariations(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		url      string
		expected string
	}{
		{
			name:     "empty URL returns text only",
			text:     "Just Text",
			url:      "",
			expected: "Just Text",
		},
		{
			name:     "URL with international domain",
			text:     "International",
			url:      "https://例え.テスト",
			expected: "\x1b]8;;https://例え.テスト\x1b\\International\x1b]8;;\x1b\\",
		},
		{
			name:     "URL with port",
			text:     "Local Server",
			url:      "http://localhost:8080",
			expected: "\x1b]8;;http://localhost:8080\x1b\\Local Server\x1b]8;;\x1b\\",
		},
		{
			name:     "URL with credentials",
			text:     "Auth Required",
			url:      "https://user:pass@example.com",
			expected: "\x1b]8;;https://user:pass@example.com\x1b\\Auth Required\x1b]8;;\x1b\\",
		},
		{
			name:     "URL with encoded characters",
			text:     "Encoded URL",
			url:      "https://example.com/path%20with%20spaces",
			expected: "\x1b]8;;https://example.com/path%20with%20spaces\x1b\\Encoded URL\x1b]8;;\x1b\\",
		},
		{
			name:     "very long URL",
			text:     "Long URL",
			url:      "https://example.com/" + strings.Repeat("segment/", 50),
			expected: "\x1b]8;;https://example.com/" + strings.Repeat("segment/", 50) + "\x1b\\Long URL\x1b]8;;\x1b\\",
		},
		{
			name:     "data URL",
			text:     "Data URL",
			url:      "data:text/plain;base64,SGVsbG8gV29ybGQ=",
			expected: "\x1b]8;;data:text/plain;base64,SGVsbG8gV29ybGQ=\x1b\\Data URL\x1b]8;;\x1b\\",
		},
		{
			name:     "javascript URL",
			text:     "JS URL",
			url:      "javascript:alert('hello')",
			expected: "\x1b]8;;javascript:alert('hello')\x1b\\JS URL\x1b]8;;\x1b\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHyperlink(tt.text, tt.url)
			if result != tt.expected {
				t.Errorf("formatHyperlink(%q, %q) = %q, expected %q", tt.text, tt.url, result, tt.expected)
			}
		})
	}
}

// TestOSC8SequenceConstants validates the OSC 8 constants are correct
func TestOSC8SequenceConstants(t *testing.T) {
	expectedStart := "\x1b]8;;"
	expectedMid := "\x1b\\"
	expectedEnd := "\x1b]8;;\x1b\\"

	if hyperlinkStart != expectedStart {
		t.Errorf("hyperlinkStart = %q, expected %q", hyperlinkStart, expectedStart)
	}

	if hyperlinkMid != expectedMid {
		t.Errorf("hyperlinkMid = %q, expected %q", hyperlinkMid, expectedMid)
	}

	if hyperlinkEnd != expectedEnd {
		t.Errorf("hyperlinkEnd = %q, expected %q", hyperlinkEnd, expectedEnd)
	}
}

// TestOSC8SequenceStructure validates the complete sequence structure
func TestOSC8SequenceStructure(t *testing.T) {
	text := "Example Text"
	url := "https://example.com"

	result := formatHyperlink(text, url)

	// Validate sequence parts
	if !strings.HasPrefix(result, hyperlinkStart+url+hyperlinkMid) {
		t.Errorf("Result should start with hyperlink sequence: %q", result)
	}

	if !strings.HasSuffix(result, hyperlinkEnd) {
		t.Errorf("Result should end with hyperlink end sequence: %q", result)
	}

	// Validate text placement
	expectedTextStart := len(hyperlinkStart + url + hyperlinkMid)
	expectedTextEnd := len(result) - len(hyperlinkEnd)

	if expectedTextEnd <= expectedTextStart {
		t.Error("Invalid sequence structure")
	}

	extractedText := result[expectedTextStart:expectedTextEnd]
	if extractedText != text {
		t.Errorf("Extracted text %q doesn't match original %q", extractedText, text)
	}
}

// TestOSC8SequenceBinaryCompatibility ensures sequences work with binary data
func TestOSC8SequenceBinaryCompatibility(t *testing.T) {
	// Test that sequences don't break with binary data in URLs
	binaryData := string([]byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD})
	text := "Binary Test"

	result := formatHyperlink(text, binaryData)

	// Should still generate a valid sequence structure
	if !strings.Contains(result, text) {
		t.Error("Binary data in URL should not break text rendering")
	}

	if !strings.Contains(result, hyperlinkStart) {
		t.Error("Binary data should not break sequence start")
	}

	if !strings.Contains(result, hyperlinkEnd) {
		t.Error("Binary data should not break sequence end")
	}
}

// TestOSC8SequencePerformance ensures sequence generation is fast
func TestOSC8SequencePerformance(t *testing.T) {
	text := "Performance Test"
	url := "https://performance.example.com"

	// Run many iterations to check for performance issues
	for i := 0; i < 10000; i++ {
		result := formatHyperlink(text, url)
		if result == "" {
			t.Error("Empty result in performance test")
			break
		}
	}
}

// BenchmarkOSC8SequenceGeneration benchmarks sequence generation
func BenchmarkOSC8SequenceGeneration(b *testing.B) {
	scenarios := map[string]struct {
		text string
		url  string
	}{
		"short": {
			text: "Link",
			url:  "https://example.com",
		},
		"medium": {
			text: "This is a medium length link text",
			url:  "https://example.com/path/to/some/resource",
		},
		"long": {
			text: strings.Repeat("Very long link text with lots of content ", 10),
			url:  "https://very-long-domain-name.example.com/very/long/path/with/many/segments/and/parameters?param1=value1&param2=value2&param3=value3",
		},
		"unicode": {
			text: "Unicode: 世界 🌍 Example",
			url:  "https://unicode.example.com/世界",
		},
	}

	for name, scenario := range scenarios {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				formatHyperlink(scenario.text, scenario.url)
			}
		})
	}
}

// TestOSC8SequenceEdgeCases tests edge cases in sequence generation
func TestOSC8SequenceEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		url         string
		shouldPanic bool
		description string
	}{
		{
			name:        "both empty",
			text:        "",
			url:         "",
			shouldPanic: false,
			description: "Both empty should return empty string",
		},
		{
			name:        "text with escape sequences",
			text:        "\x1b[31mRed Text\x1b[0m",
			url:         "https://example.com",
			shouldPanic: false,
			description: "Text with ANSI should be preserved",
		},
		{
			name:        "url with escape sequences",
			text:        "Link",
			url:         "https://example.com\x1b[31m",
			shouldPanic: false,
			description: "URL with ANSI should be preserved",
		},
		{
			name:        "very long inputs",
			text:        strings.Repeat("A", 100000),
			url:         "https://example.com/" + strings.Repeat("segment/", 1000),
			shouldPanic: false,
			description: "Very long inputs should work",
		},
		{
			name:        "null bytes",
			text:        "Text\x00with\x00nulls",
			url:         "https://example.com\x00/path",
			shouldPanic: false,
			description: "Null bytes should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("%s: unexpected panic: %v", tt.description, r)
					}
				} else if tt.shouldPanic {
					t.Errorf("%s: expected panic but didn't get one", tt.description)
				}
			}()

			result := formatHyperlink(tt.text, tt.url)

			// Basic validation that result has expected structure
			if tt.url == "" && result != tt.text {
				t.Errorf("%s: expected %q when URL empty, got %q", tt.description, tt.text, result)
			} else if tt.url != "" && !strings.Contains(result, tt.text) {
				t.Errorf("%s: result should contain text %q", tt.description, tt.text)
			}
		})
	}
}

// TestOSC8SequenceUTF8Validity ensures sequences maintain UTF-8 validity
func TestOSC8SequenceUTF8Validity(t *testing.T) {
	tests := []struct {
		name string
		text string
		url  string
	}{
		{"ascii", "Hello World", "https://example.com"},
		{"utf8", "Hello 世界", "https://例え.com"},
		{"emoji", "Click 👉 here", "https://emoji.example.com"},
		{"mixed", "ASCII中文🌍", "https://mixed.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHyperlink(tt.text, tt.url)

			if !utf8.ValidString(result) {
				t.Errorf("Result is not valid UTF-8: %q", result)
			}

			if !utf8.ValidString(tt.text) {
				t.Errorf("Input text is not valid UTF-8: %q", tt.text)
			}

			if !utf8.ValidString(tt.url) {
				t.Errorf("Input URL is not valid UTF-8: %q", tt.url)
			}
		})
	}
}
