package glamour

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

const (
	generate = false
	markdown = "testdata/readme.markdown.in"
	testFile = "testdata/readme.test"
)

func TestTermRendererWriter(t *testing.T) {
	r, err := NewTermRenderer(
		WithStandardStyle(DarkStyle),
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

	// generate
	if generate {
		err := os.WriteFile(testFile, b, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		return
	}

	// verify
	td, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(td, b) {
		t.Errorf("Rendered output doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
			string(td), b)
	}
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

	// verify
	td, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(td, []byte(b)) {
		t.Errorf("Rendered output doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
			string(td), b)
	}
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

	// verify
	td, err := os.ReadFile("testdata/preserved_newline.test")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(td, []byte(b)) {
		t.Errorf("Rendered output doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
			string(td), b)
	}
}

func TestStyles(t *testing.T) {
	_, err := NewTermRenderer(
		WithAutoStyle(),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewTermRenderer(
		WithStandardStyle(AutoStyle),
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
		{name: "style doesn't exist", stylePath: "testdata/notfound.style", err: os.ErrNotExist, expected: AutoStyle},
		{name: "style is empty", stylePath: "", err: nil, expected: AutoStyle},
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

	// verify
	td, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if b != string(td) {
		t.Errorf("Rendered output doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
			string(td), b)
	}
}

func TestCapitalization(t *testing.T) {
	p := true
	style := DarkStyleConfig
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

	// expected outcome
	td, err := os.ReadFile("testdata/capitalization.test")
	if err != nil {
		t.Fatal(err)
	}

	if string(td) != b {
		t.Errorf("Rendered output doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n", td, b)
	}
}
