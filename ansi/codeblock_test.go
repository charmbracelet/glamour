package ansi

import (
	"bytes"
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func uintPtr(u uint) *uint {
	return &u
}

func stringPtr(s string) *string {
	return &s
}

func TestCodeBlockIndentToken(t *testing.T) {
	tests := []struct {
		name        string
		indent      *uint
		margin      *uint
		indentToken *string
		code        string
		language    string
		theme       string
		expected    []string
	}{
		{
			name:        "custom indent token without theme",
			indent:      uintPtr(1),
			indentToken: stringPtr("| "),
			code:        "line1\nline2\nline3",
			expected: []string{
				"| line1",
				"| line2",
				"| line3",
			},
		},
		{
			name:        "custom indent token with indentation count",
			indent:      uintPtr(2),
			indentToken: stringPtr(">"),
			code:        "foo\nbar",
			expected: []string{
				">>foo",
				">>bar",
			},
		},
		{
			name:        "default fallback to space when indent token is nil",
			indent:      uintPtr(2),
			indentToken: nil,
			code:        "hello\nworld",
			expected: []string{
				"  hello",
				"  world",
			},
		},
		{
			name:        "custom indent token with chroma theme",
			indent:      uintPtr(1),
			indentToken: stringPtr("# "),
			code:        "package main\n\nfunc main() {}",
			language:    "go",
			theme:       "monokai",
			expected: []string{
				"# package main",
				"# func main() {}",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewRenderContext(Options{
				Styles: StyleConfig{
					CodeBlock: StyleCodeBlock{
						StyleBlock: StyleBlock{
							Indent:      tc.indent,
							Margin:      tc.margin,
							IndentToken: tc.indentToken,
						},
						Theme: tc.theme,
					},
				},
			})

			el := &CodeBlockElement{
				Code:     tc.code,
				Language: tc.language,
			}

			var buf bytes.Buffer
			if err := el.Render(&buf, ctx); err != nil {
				t.Fatalf("unexpected render error: %v", err)
			}

			output := xansi.Strip(buf.String())
			for _, exp := range tc.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("expected output to contain %q, but got:\n%s", exp, output)
				}
			}
		})
	}
}

func TestCodeBlockMarkdownRenderingWithIndentToken(t *testing.T) {
	indent := uint(1)
	token := "| "
	options := Options{
		Styles: StyleConfig{
			CodeBlock: StyleCodeBlock{
				StyleBlock: StyleBlock{
					Indent:      &indent,
					IndentToken: &token,
				},
			},
		},
	}

	md := goldmark.New()
	ar := NewRenderer(options)
	md.SetRenderer(
		renderer.NewRenderer(
			renderer.WithNodeRenderers(util.Prioritized(ar, 1000)),
		),
	)

	markdown := "```\necho hello\necho world\n```\n"

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("failed to convert markdown: %v", err)
	}

	output := xansi.Strip(buf.String())
	if !strings.Contains(output, "| echo hello") {
		t.Errorf("expected output to contain '| echo hello', got:\n%s", output)
	}
	if !strings.Contains(output, "| echo world") {
		t.Errorf("expected output to contain '| echo world', got:\n%s", output)
	}
}
