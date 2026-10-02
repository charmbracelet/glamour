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

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	xansi "github.com/charmbracelet/x/ansi"
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

func TestWrappedListIndentation(t *testing.T) {
	t.Run("unordered list wrap", func(t *testing.T) {
		in := `- Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla mattis dignissim leo et tempus. Cras sit amet nisi id leo eleifend iaculis nec in lectus. Nam dictum laoreet ex eu laoreet. Morbi quis malesuada lacus, et blandit erat.
    - Cras quis ornare mi, in condimentum tortor. Vivamus id convallis ligula. Morbi ac commodo lacus, in blandit augue. Cras sed nulla risus`

		r, err := NewTermRenderer(
			WithStandardStyle("dark"),
			WithWordWrap(80),
		)
		if err != nil {
			t.Fatal(err)
		}

		out, err := r.Render(in)
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(out, "\n")
		var textLines []string
		for _, l := range lines {
			stripped := xansi.Strip(strings.TrimRight(l, " "))
			if len(strings.TrimSpace(stripped)) > 0 {
				textLines = append(textLines, stripped)
			}
		}

		if len(textLines) < 6 {
			t.Fatalf("expected at least 6 non-empty lines, got %d:\n%v", len(textLines), textLines)
		}

		// First line of outer item: 2 spaces doc margin + bullet (col 2, text col 4)
		if !strings.HasPrefix(textLines[0], "  • ") {
			t.Errorf("expected outer item first line to start with '  • ', got: %q", textLines[0])
		}
		// Continuation lines of outer item: 4 spaces (2 doc margin + 2 hanging indent)
		for i := 1; i <= 3; i++ {
			if !strings.HasPrefix(textLines[i], "    ") || strings.HasPrefix(textLines[i], "    •") {
				t.Errorf("expected outer item continuation line %d to start with 4 spaces, got: %q", i, textLines[i])
			}
		}

		// First line of nested item: 4 spaces (2 doc margin + 2 list indent) + bullet
		if !strings.HasPrefix(textLines[4], "    • ") {
			t.Errorf("expected nested item first line to start with '    • ', got: %q", textLines[4])
		}
		// Continuation line of nested item: 6 spaces (2 doc margin + 2 list indent + 2 hanging indent)
		if !strings.HasPrefix(textLines[5], "      ") {
			t.Errorf("expected nested item continuation line to start with 6 spaces, got: %q", textLines[5])
		}
	})

	t.Run("ordered list wrap", func(t *testing.T) {
		in := `1. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla mattis dignissim leo et tempus. Cras sit amet nisi id leo eleifend iaculis nec in lectus. Nam dictum laoreet ex eu laoreet.`

		r, err := NewTermRenderer(
			WithStandardStyle("dark"),
			WithWordWrap(80),
		)
		if err != nil {
			t.Fatal(err)
		}

		out, err := r.Render(in)
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(out, "\n")
		var textLines []string
		for _, l := range lines {
			stripped := xansi.Strip(strings.TrimRight(l, " "))
			if len(strings.TrimSpace(stripped)) > 0 {
				textLines = append(textLines, stripped)
			}
		}

		if len(textLines) < 2 {
			t.Fatalf("expected at least 2 non-empty lines, got %d:\n%v", len(textLines), textLines)
		}

		// First line: 2 spaces doc margin + "1. " (col 2, text col 5)
		if !strings.HasPrefix(textLines[0], "  1. ") {
			t.Errorf("expected ordered item first line to start with '  1. ', got: %q", textLines[0])
		}
		// Continuation lines: 5 spaces (2 doc margin + 3 hanging indent for '1. ')
		for i := 1; i < len(textLines); i++ {
			if !strings.HasPrefix(textLines[i], "     ") {
				t.Errorf("expected ordered continuation line %d to start with 5 spaces, got: %q", i, textLines[i])
			}
		}
	})

	t.Run("task list wrap", func(t *testing.T) {
		in := `- [ ] Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla mattis dignissim leo et tempus. Cras sit amet nisi id leo eleifend iaculis nec in lectus.`

		r, err := NewTermRenderer(
			WithStandardStyle("dark"),
			WithWordWrap(80),
		)
		if err != nil {
			t.Fatal(err)
		}

		out, err := r.Render(in)
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(out, "\n")
		var textLines []string
		for _, l := range lines {
			stripped := xansi.Strip(strings.TrimRight(l, " "))
			if len(strings.TrimSpace(stripped)) > 0 {
				textLines = append(textLines, stripped)
			}
		}

		if len(textLines) < 2 {
			t.Fatalf("expected at least 2 non-empty lines, got %d:\n%v", len(textLines), textLines)
		}

		// First line: 2 spaces doc margin + "[ ] " (col 2, text col 6)
		if !strings.HasPrefix(textLines[0], "  [ ] ") {
			t.Errorf("expected task item first line to start with '  [ ] ', got: %q", textLines[0])
		}
		// Continuation lines: 6 spaces (2 doc margin + 4 hanging indent for '[ ] ')
		for i := 1; i < len(textLines); i++ {
			if !strings.HasPrefix(textLines[i], "      ") {
				t.Errorf("expected task continuation line %d to start with 6 spaces, got: %q", i, textLines[i])
			}
		}
	})
}
