package ansi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// chromaTestStyle returns a style whose code blocks differ only in the color
// given to every token, by way of the chroma text and name entries.
func chromaTestStyle(color string) StyleConfig {
	var s StyleConfig
	s.CodeBlock.Chroma = &Chroma{}
	s.CodeBlock.Chroma.Text.Color = &color
	s.CodeBlock.Chroma.Name.Color = &color
	return s
}

func renderCodeBlock(s StyleConfig) (string, error) {
	md := goldmark.New()
	ar := NewRenderer(Options{WordWrap: 80, Styles: s})
	md.SetRenderer(renderer.NewRenderer(
		renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

	var buf bytes.Buffer
	err := md.Convert([]byte("```go\nfmt.Println(x)\n```\n"), &buf)
	return buf.String(), err
}

// TestCodeBlockChromaStyleIsPerRender renders the same code block with
// different chroma styles in one process. Each render must carry its own
// colors, not those of whichever style was rendered first.
func TestCodeBlockChromaStyleIsPerRender(t *testing.T) {
	// terminal256 writes #ff0000 as color 196 and #00ff00 as color 46.
	red := chromaTestStyle("#ff0000")
	green := chromaTestStyle("#00ff00")

	for i, tc := range []struct {
		name string
		s    StyleConfig
		has  string
		not  string
	}{
		{"red first", red, "38;5;196", "38;5;46"},
		{"green second", green, "38;5;46", "38;5;196"},
		{"red again", red, "38;5;196", "38;5;46"},
	} {
		got, err := renderCodeBlock(tc.s)
		if err != nil {
			t.Fatalf("%d %s: %v", i, tc.name, err)
		}
		if !strings.Contains(got, tc.has) {
			t.Errorf("%d %s: no %q in %q", i, tc.name, tc.has, got)
		}
		if strings.Contains(got, tc.not) {
			t.Errorf("%d %s: found %q from another style in %q", i, tc.name, tc.not, got)
		}
	}
}

// TestCodeBlockChromaStyleBadColor checks that a chroma color chroma cannot
// parse is returned as an error and does not panic. An ANSI number or a color
// name is valid elsewhere in a style, but not in a chroma section.
func TestCodeBlockChromaStyleBadColor(t *testing.T) {
	// A good style first, so that the bad one is not the first of the process.
	if _, err := renderCodeBlock(chromaTestStyle("#ff0000")); err != nil {
		t.Fatal(err)
	}

	for _, color := range []string{"203", "red", "#ggg"} {
		t.Run(color, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic: %v", r)
				}
			}()
			if _, err := renderCodeBlock(chromaTestStyle(color)); err == nil {
				t.Error("no error for a color chroma cannot parse")
			}
		})
	}
}
