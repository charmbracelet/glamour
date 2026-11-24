package glamour

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
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
	_, err := NewTermRenderer()
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewTermRenderer(
		WithStandardStyle(styles.DarkStyle),
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
		{name: "style doesn't exist", stylePath: "testdata/notfound.style", err: os.ErrNotExist, expected: styles.DarkStyle},
		{name: "style is empty", stylePath: "", err: nil, expected: styles.DarkStyle},
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

func TestWithHyperlinkModeInline(t *testing.T) {
	r, err := NewTermRenderer(
		WithHyperlinkMode(ansi.HyperlinkModeInline),
	)
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render("[click here](https://charm.land)")
	if err != nil {
		t.Fatal(err)
	}

	// the URL should not be printed as text
	if strings.Contains(b, "https://charm.land\x1b]8;;\a") && strings.Contains(b, "https://charm.land https://charm.land") {
		t.Errorf("expected URL to be hidden, got: %q", b)
	}
	// link text should be wrapped in an OSC 8 hyperlink
	if !strings.Contains(b, "\x1b]8;") || !strings.Contains(b, "click here") {
		t.Errorf("expected OSC 8 hyperlink around link text, got: %q", b)
	}
	// link text should be underlined (SGR 4)
	if !strings.Contains(b, "\x1b[4m") {
		t.Errorf("expected link text to be underlined, got: %q", b)
	}
}

func TestWithHyperlinkModeInlineWithBase(t *testing.T) {
	r, err := NewTermRenderer(
		WithHyperlinkMode(ansi.HyperlinkModeInline),
		WithBaseURL("https://charm.land"),
	)
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render("[click here](/libs)")
	if err != nil {
		t.Fatal(err)
	}

	// the URL should not be printed as text
	if strings.Contains(b, "https://charm.land/libs\x1b]8;;\a") && strings.Contains(b, "https://charm.land/libs https://charm.land/libs") {
		t.Errorf("expected URL to be hidden, got: %q", b)
	}
	// link text should be wrapped in an OSC 8 hyperlink
	if !strings.Contains(b, "\x1b]8;") || !strings.Contains(b, "click here") {
		t.Errorf("expected OSC 8 hyperlink around link text, got: %q", b)
	}
	// link text should be underlined (SGR 4)
	if !strings.Contains(b, "\x1b[4m") {
		t.Errorf("expected link text to be underlined, got: %q", b)
	}
}

func TestWithHyperlinkModeAuto(t *testing.T) {
	r, err := NewTermRenderer(
		WithHyperlinkMode(ansi.HyperlinkModeAuto),
	)
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render("[click here](https://charm.land)")
	if err != nil {
		t.Fatal(err)
	}

	// the URL should be printed as text after the link text
	if !strings.Contains(b, "click here") || !strings.Contains(b, "https://charm.land") {
		t.Errorf("expected link text and URL in output, got: %q", b)
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
	// ...Header.A.............................|.Header.B............................
	// ..--------------------------------------|-------------------------------------
	// ...Cell.1...............................|.Cell.2..............................
	// ...Cell.3...............................|.Cell.4..............................
	// ...Cell.5...............................|.Cell.6..............................
}

const (
	chromaStyleTestMarkdown = "```go\nfunc main() {}\n```\n"
	darkKeywordColor        = "38;2;0;170;255"  // #00AAFF
	lightKeywordColor       = "38;2;39;158;252" // #279EFC
)

func renderChromaStyle(t *testing.T, style string) string {
	t.Helper()

	r, err := NewTermRenderer(
		WithStandardStyle(style),
		WithChromaFormatter("terminal16m"),
	)
	if err != nil {
		t.Error(err)
		return ""
	}

	b, err := r.Render(chromaStyleTestMarkdown)
	if err != nil {
		t.Error(err)
	}
	return b
}

func checkKeywordColor(t *testing.T, style, out string) {
	t.Helper()

	want, unwanted := darkKeywordColor, lightKeywordColor
	if style == styles.LightStyle {
		want, unwanted = lightKeywordColor, darkKeywordColor
	}
	if !strings.Contains(out, want) {
		t.Errorf("%s: expected keyword color %q in output", style, want)
	}
	if strings.Contains(out, unwanted) {
		t.Errorf("%s: unexpected keyword color %q in output", style, unwanted)
	}
}

// TestChromaStylePerRenderer checks that each renderer highlights code with
// its own chroma style, regardless of which renderers were created before it.
func TestChromaStylePerRenderer(t *testing.T) {
	tests := []struct {
		name  string
		order []string
	}{
		{name: "dark then light", order: []string{styles.DarkStyle, styles.LightStyle}},
		{name: "light then dark", order: []string{styles.LightStyle, styles.DarkStyle}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, style := range tc.order {
				checkKeywordColor(t, style, renderChromaStyle(t, style))
			}
		})
	}
}

func TestChromaStyleConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for range 10 {
		for _, style := range []string{styles.DarkStyle, styles.LightStyle} {
			wg.Go(func() {
				checkKeywordColor(t, style, renderChromaStyle(t, style))
			})
		}
	}
	wg.Wait()
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
