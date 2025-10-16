package glamour

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/x/exp/golden"
)

const markdown = "testdata/readme.markdown.in"

func TestTermRendererWriter(t *testing.T) {
	r, err := NewTermRenderer(
		WithStandardStyle(styles.DarkStyle),
	)
	if err != nil {
		t.Fatal(err)
	}

	in, err := os.ReadFile(markdown)
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.Write(in)
	if err != nil {
		t.Fatal(err)
	}
	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	golden.RequireEqual(t, b)
}

func TestTermRenderer(t *testing.T) {
	r, err := NewTermRenderer(
		WithStandardStyle("dark"),
	)
	if err != nil {
		t.Fatal(err)
	}

	in, err := os.ReadFile(markdown)
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render(string(in))
	if err != nil {
		t.Fatal(err)
	}

	golden.RequireEqual(t, []byte(b))
}

func TestWithEmoji(t *testing.T) {
	r, err := NewTermRenderer(
		WithEmoji(),
	)
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render(":+1:")
	if err != nil {
		t.Fatal(err)
	}
	b = strings.TrimSpace(b)

	// Thumbs up unicode character
	td := "\U0001f44d"

	if td != b {
		t.Errorf("Rendered output doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n", td, b)
	}
}

func TestWithPreservedNewLines(t *testing.T) {
	r, err := NewTermRenderer(
		WithPreservedNewLines(),
	)
	if err != nil {
		t.Fatal(err)
	}

	in, err := os.ReadFile("testdata/preserved_newline.in")
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render(string(in))
	if err != nil {
		t.Fatal(err)
	}

	golden.RequireEqual(t, []byte(b))
}

func TestStyles(t *testing.T) {
	_, err := NewTermRenderer(
		WithAutoStyle(),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewTermRenderer(
		WithStandardStyle(styles.AutoStyle),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewTermRenderer(
		WithEnvironmentConfig(),
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestCustomStyle checks the expected errors with custom styling. We need to
// support built-in styles and custom style sheets.
func TestCustomStyle(t *testing.T) {
	md := "testdata/example.md"
	tests := []struct {
		name      string
		stylePath string
		err       error
		expected  string
	}{
		{name: "style exists", stylePath: "testdata/custom.style", err: nil, expected: "testdata/custom.style"},
		{name: "style doesn't exist", stylePath: "testdata/notfound.style", err: os.ErrNotExist, expected: styles.AutoStyle},
		{name: "style is empty", stylePath: "", err: nil, expected: styles.AutoStyle},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GLAMOUR_STYLE", tc.stylePath)
			g, err := NewTermRenderer(
				WithEnvironmentConfig(),
			)
			if !errors.Is(err, tc.err) {
				t.Fatal(err)
			}
			if !errors.Is(tc.err, os.ErrNotExist) {
				w, err := NewTermRenderer(WithStylePath(tc.expected))
				if err != nil {
					t.Fatal(err)
				}
				text, _ := os.ReadFile(md)
				want, err := w.RenderBytes(text)
				got, err := g.RenderBytes(text)
				if !bytes.Equal(want, got) {
					t.Error("Wrong style used")
				}
			}
		})
	}
}

func TestRenderHelpers(t *testing.T) {
	in, err := os.ReadFile(markdown)
	if err != nil {
		t.Fatal(err)
	}

	b, err := Render(string(in), "dark")
	if err != nil {
		t.Error(err)
	}

	golden.RequireEqual(t, []byte(b))
}

func TestCapitalization(t *testing.T) {
	p := true
	style := styles.DarkStyleConfig
	style.H1.Upper = &p
	style.H2.Title = &p
	style.H3.Lower = &p

	r, err := NewTermRenderer(
		WithStyles(style),
	)
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render("# everything is uppercase\n## everything is titled\n### everything is lowercase")
	if err != nil {
		t.Fatal(err)
	}

	golden.RequireEqual(t, []byte(b))
}

func FuzzData(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		func() int {
			_, err := RenderBytes(data, styles.DarkStyle)
			if err != nil {
				return 0
			}
			return 1
		}()
	})
}

func TestTableAscii(t *testing.T) {
	markdown := strings.TrimSpace(`
| Header A  | Header B  |
| --------- | --------- |
| Cell 1    | Cell 2    |
| Cell 3    | Cell 4    |
| Cell 5    | Cell 6    |
`)

	renderer, err := NewTermRenderer(
		WithStyles(styles.ASCIIStyleConfig),
		WithWordWrap(80),
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := renderer.Render(markdown)
	if err != nil {
		t.Fatal(err)
	}

	nonAsciiRegexp := regexp.MustCompile(`[^\x00-\x7f]+`)
	nonAsciiChars := nonAsciiRegexp.FindAllString(result, -1)
	if len(nonAsciiChars) > 0 {
		t.Errorf("Non-ASCII characters found in output: %v", nonAsciiChars)
	}
}

func ExampleASCIIStyleConfig() {
	markdown := strings.TrimSpace(`
| Header A  | Header B  |
| --------- | --------- |
| Cell 1    | Cell 2    |
| Cell 3    | Cell 4    |
| Cell 5    | Cell 6    |
`)

	renderer, err := NewTermRenderer(
		WithStyles(styles.ASCIIStyleConfig),
		WithWordWrap(80),
	)
	if err != nil {
		return
	}

	result, err := renderer.Render(markdown)
	if err != nil {
		return
	}
	result = strings.ReplaceAll(result, " ", ".")
	fmt.Println(result)

	// Output:
	// ..............................................................................
	// ...Header.A............................|.Header.B.............................
	// ..-------------------------------------|------------------------------------..
	// ...Cell.1..............................|.Cell.2...............................
	// ...Cell.3..............................|.Cell.4...............................
	// ...Cell.5..............................|.Cell.6...............................
}

func TestWithChromaFormatterDefault(t *testing.T) {
	r, err := NewTermRenderer(
		WithStandardStyle(styles.DarkStyle),
	)
	if err != nil {
		t.Fatal(err)
	}

	in, err := os.ReadFile("testdata/TestWithChromaFormatter.md")
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render(string(in))
	if err != nil {
		t.Fatal(err)
	}

	golden.RequireEqual(t, []byte(b))
}

func TestWithChromaFormatterCustom(t *testing.T) {
	r, err := NewTermRenderer(
		WithStandardStyle(styles.DarkStyle),
		WithChromaFormatter("terminal16"),
	)
	if err != nil {
		t.Fatal(err)
	}

	in, err := os.ReadFile("testdata/TestWithChromaFormatter.md")
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render(string(in))
	if err != nil {
		t.Fatal(err)
	}

	golden.RequireEqual(t, []byte(b))
}

func TestWithLinkFormatter(t *testing.T) {
	customFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		return fmt.Sprintf("CUSTOM[%s](%s)", data.Text, data.URL), nil
	})

	r, err := NewTermRenderer(
		WithStandardStyle("dark"),
		WithLinkFormatter(customFormatter),
	)
	if err != nil {
		t.Fatal(err)
	}

	markdown := "[example](https://example.com)"
	result, err := r.Render(markdown)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "CUSTOM[example](https://example.com)") {
		t.Errorf("expected custom formatter output, got: %s", result)
	}
}

func TestWithTextOnlyLinks(t *testing.T) {
	tests := []struct {
		name               string
		markdown           string
		supportsHyperlinks bool
		wantContains       []string
		wantNotContains    []string
	}{
		{
			name:               "hyperlink support",
			markdown:           "[example](https://example.com)",
			supportsHyperlinks: true,
			wantContains:       []string{"example", "\x1b]8;;https://example.com\x1b\\"},
			wantNotContains:    []string{},
		},
		{
			name:               "no hyperlink support",
			markdown:           "[example](https://example.com)",
			supportsHyperlinks: false,
			wantContains:       []string{"example"},
			wantNotContains:    []string{"https://example.com"},
		},
		{
			name:               "multiple links",
			markdown:           "[first](https://first.com) and [second](https://second.com)",
			supportsHyperlinks: false,
			wantContains:       []string{"first", "second"},
			wantNotContains:    []string{"https://first.com", "https://second.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up terminal environment
			if tt.supportsHyperlinks {
				t.Setenv("TERM_PROGRAM", "iTerm.app")
			} else {
				t.Setenv("TERM_PROGRAM", "")
				t.Setenv("TERM", "")
			}

			r, err := NewTermRenderer(
				WithStandardStyle("dark"),
				WithTextOnlyLinks(),
			)
			if err != nil {
				t.Fatal(err)
			}

			result, err := r.Render(tt.markdown)
			if err != nil {
				t.Fatal(err)
			}

			for _, expected := range tt.wantContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, got: %s", expected, result)
				}
			}

			for _, notExpected := range tt.wantNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("expected result NOT to contain %q, got: %s", notExpected, result)
				}
			}
		})
	}
}

func TestWithURLOnlyLinks(t *testing.T) {
	tests := []struct {
		name            string
		markdown        string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "normal link",
			markdown:        "[example text](https://example.com)",
			wantContains:    []string{"https://example.com"},
			wantNotContains: []string{"example text"},
		},
		{
			name:            "fragment link ignored",
			markdown:        "[section](#fragment)",
			wantContains:    []string{},
			wantNotContains: []string{"#fragment", "section"},
		},
		{
			name:            "multiple links",
			markdown:        "[Click here](https://example.com) and [Visit site](https://test.org)",
			wantContains:    []string{"https://example.com", "https://test.org"},
			wantNotContains: []string{"Click here", "Visit site"},
		},
		{
			name:            "relative URL with base",
			markdown:        "[path](/relative)",
			wantContains:    []string{"/relative"}, // Base URL would be resolved by formatter
			wantNotContains: []string{"path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewTermRenderer(
				WithStandardStyle("dark"),
				WithURLOnlyLinks(),
			)
			if err != nil {
				t.Fatal(err)
			}

			result, err := r.Render(tt.markdown)
			if err != nil {
				t.Fatal(err)
			}

			// Strip ANSI codes for easier checking
			plainResult := stripANSISequences(result)

			for _, expected := range tt.wantContains {
				if !strings.Contains(plainResult, expected) {
					t.Errorf("expected result to contain %q, got: %s", expected, plainResult)
				}
			}

			for _, notExpected := range tt.wantNotContains {
				if strings.Contains(plainResult, notExpected) {
					t.Errorf("expected result NOT to contain %q, got: %s", notExpected, plainResult)
				}
			}
		})
	}
}

func TestWithHyperlinks(t *testing.T) {
	r, err := NewTermRenderer(
		WithStandardStyle("dark"),
		WithHyperlinks(),
	)
	if err != nil {
		t.Fatal(err)
	}

	markdown := "[example](https://example.com)"
	result, err := r.Render(markdown)
	if err != nil {
		t.Fatal(err)
	}

	// Should contain OSC 8 sequences regardless of terminal support
	if !strings.Contains(result, "\x1b]8;;https://example.com\x1b\\") {
		t.Errorf("expected OSC 8 hyperlink sequences, got: %s", result)
	}
	if !strings.Contains(result, "example") {
		t.Errorf("expected link text, got: %s", result)
	}
	if !strings.Contains(result, "\x1b]8;;\x1b\\") {
		t.Errorf("expected OSC 8 end sequence, got: %s", result)
	}
}

func TestWithSmartHyperlinks(t *testing.T) {
	tests := []struct {
		name               string
		supportsHyperlinks bool
		wantHyperlink      bool
		wantFallback       bool
	}{
		{
			name:               "modern terminal",
			supportsHyperlinks: true,
			wantHyperlink:      true,
			wantFallback:       false,
		},
		{
			name:               "legacy terminal",
			supportsHyperlinks: false,
			wantHyperlink:      false,
			wantFallback:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up terminal environment
			if tt.supportsHyperlinks {
				t.Setenv("TERM_PROGRAM", "iTerm.app")
			} else {
				t.Setenv("TERM_PROGRAM", "")
				t.Setenv("TERM", "")
			}

			r, err := NewTermRenderer(
				WithStandardStyle("dark"),
				WithSmartHyperlinks(),
			)
			if err != nil {
				t.Fatal(err)
			}

			markdown := "[example](https://example.com)"
			result, err := r.Render(markdown)
			if err != nil {
				t.Fatal(err)
			}

			if tt.wantHyperlink {
				if !strings.Contains(result, "\x1b]8;;") {
					t.Errorf("expected hyperlink sequences in modern terminal, got: %s", result)
				}
			}

			if tt.wantFallback {
				plainResult := stripANSISequences(result)
				if !strings.Contains(plainResult, "example") {
					t.Errorf("expected link text in fallback, got: %s", plainResult)
				}
				if !strings.Contains(plainResult, "https://example.com") {
					t.Errorf("expected URL in fallback, got: %s", plainResult)
				}
			}
		})
	}
}

func TestLinkFormatterIntegration(t *testing.T) {
	// Test complex markdown with multiple link types
	complexMarkdown := `# Test Document

Regular link: [GitHub](https://github.com)
Autolink: <https://example.com>
Reference link: [Google][1]

[1]: https://google.com "Google Search"

## In Lists

* [Link 1](https://one.com)
* [Link 2](https://two.com)

## In Tables

| Name | URL |
|------|-----|
| [Site 1](https://site1.com) | Description 1 |
| [Site 2](https://site2.com) | Description 2 |
`

	tests := []struct {
		name   string
		option TermRendererOption
		check  func(t *testing.T, result string)
	}{
		{
			name:   "default behavior",
			option: WithStandardStyle("dark"),
			check: func(t *testing.T, result string) {
				plain := stripANSISequences(result)
				// Should contain both text and URLs
				if !strings.Contains(plain, "GitHub") || !strings.Contains(plain, "https://github.com") {
					t.Error("default should show both text and URL")
				}
			},
		},
		{
			name:   "text only",
			option: WithOptions(WithStandardStyle("dark"), WithTextOnlyLinks()),
			check: func(t *testing.T, result string) {
				plain := stripANSISequences(result)
				// Should contain text but not visible URLs (unless in hyperlinks)
				if !strings.Contains(plain, "GitHub") {
					t.Error("should contain link text")
				}
			},
		},
		{
			name:   "URL only",
			option: WithOptions(WithStandardStyle("dark"), WithURLOnlyLinks()),
			check: func(t *testing.T, result string) {
				plain := stripANSISequences(result)
				// Should contain URLs but not descriptive text
				if !strings.Contains(plain, "https://github.com") {
					t.Error("should contain URLs")
				}
				if strings.Contains(plain, "GitHub") {
					t.Error("should not contain link text")
				}
			},
		},
		{
			name: "custom formatter",
			option: WithOptions(WithStandardStyle("dark"), WithLinkFormatter(ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
				return fmt.Sprintf("{%s->%s}", data.Text, data.URL), nil
			}))),
			check: func(t *testing.T, result string) {
				plain := stripANSISequences(result)
				if !strings.Contains(plain, "{GitHub->https://github.com}") {
					t.Error("should contain custom format")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewTermRenderer(tt.option)
			if err != nil {
				t.Fatal(err)
			}

			result, err := r.Render(complexMarkdown)
			if err != nil {
				t.Fatal(err)
			}

			tt.check(t, result)
		})
	}
}

func TestLinkFormatterErrorHandling(t *testing.T) {
	errorFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		return "", fmt.Errorf("formatter error")
	})

	r, err := NewTermRenderer(
		WithStandardStyle("dark"),
		WithLinkFormatter(errorFormatter),
	)
	if err != nil {
		t.Fatal(err)
	}

	markdown := "[example](https://example.com)"
	_, err = r.Render(markdown)
	if err == nil {
		t.Error("expected formatter error to be propagated")
	}
	if !strings.Contains(err.Error(), "formatter error") {
		t.Errorf("expected formatter error in message, got: %s", err.Error())
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that existing behavior is preserved when no custom formatter is set
	markdown := "[GitHub](https://github.com) and <https://example.com>"

	// Render without custom formatter (current behavior)
	r1, err := NewTermRenderer(WithStandardStyle("dark"))
	if err != nil {
		t.Fatal(err)
	}
	result1, err := r1.Render(markdown)
	if err != nil {
		t.Fatal(err)
	}

	// Render with explicit default formatter
	r2, err := NewTermRenderer(
		WithStandardStyle("dark"),
		WithLinkFormatter(ansi.DefaultFormatter),
	)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := r2.Render(markdown)
	if err != nil {
		t.Fatal(err)
	}

	// Results should be identical
	if result1 != result2 {
		t.Error("default behavior should match explicit DefaultFormatter")
	}

	// Both should contain text and URLs
	plain := stripANSISequences(result1)
	if !strings.Contains(plain, "GitHub") {
		t.Error("should contain link text")
	}
	if !strings.Contains(plain, "https://github.com") {
		t.Error("should contain GitHub URL")
	}
	if !strings.Contains(plain, "https://example.com") {
		t.Error("should contain example URL")
	}
}

// Helper function to strip ANSI sequences for testing
func stripANSISequences(text string) string {
	// Use the same regex as in hyperlink.go
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[a-zA-Z]`)
	return re.ReplaceAllString(text, "")
}

// Performance benchmarks
func BenchmarkLinkRenderingDefault(b *testing.B) {
	r, _ := NewTermRenderer(WithStandardStyle("dark"))
	markdown := "[example](https://example.com)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Render(markdown)
	}
}

func BenchmarkLinkRenderingCustomFormatter(b *testing.B) {
	formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		return fmt.Sprintf("%s %s", data.Text, data.URL), nil
	})

	r, _ := NewTermRenderer(
		WithStandardStyle("dark"),
		WithLinkFormatter(formatter),
	)
	markdown := "[example](https://example.com)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Render(markdown)
	}
}
