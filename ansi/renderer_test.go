package ansi

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const (
	examplesDir = "../styles/examples/"
	issuesDir   = "../testdata/issues/"
)

func TestRenderer(t *testing.T) {
	files, err := filepath.Glob(examplesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(bn, func(t *testing.T) {
			sn := filepath.Join(examplesDir, bn+".style")

			in, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(sn)
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap: 80,
			}
			err = json.Unmarshal(b, &options.Styles)
			if err != nil {
				t.Fatal(err)
			}

			switch bn {
			case "table_wrap":
				tableWrap := true
				options.TableWrap = &tableWrap
			case "table_truncate":
				tableWrap := false
				options.TableWrap = &tableWrap
			case "table_with_inline_links":
				options.InlineTableLinks = true
			case "table_with_footer_links", "table_with_footer_links_no_color":
				options.InlineTableLinks = false
			}

			md := goldmark.New(
				goldmark.WithExtensions(
					extension.GFM,
					extension.DefinitionList,
					emoji.Emoji,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
			)

			ar := NewRenderer(options)
			md.SetRenderer(
				renderer.NewRenderer(
					renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

			var buf bytes.Buffer
			if err := md.Convert(in, &buf); err != nil {
				t.Error(err)
			}

			golden.RequireEqual(t, buf.Bytes())
		})
	}
}

func TestRendererIssues(t *testing.T) {
	files, err := filepath.Glob(issuesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(bn, func(t *testing.T) {
			in, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile("../styles/dark.json")
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap: 80,
			}
			err = json.Unmarshal(b, &options.Styles)
			if err != nil {
				t.Fatal(err)
			}
			if bn == "493" {
				tableWrap := false
				options.TableWrap = &tableWrap
			}
			if bn == "637" {
				options.BaseURL = "https://github.com/charmbracelet/glamour"
			}

			md := goldmark.New(
				goldmark.WithExtensions(
					extension.GFM,
					extension.DefinitionList,
					emoji.Emoji,
				),
				goldmark.WithParserOptions(
					parser.WithAutoHeadingID(),
				),
			)

			ar := NewRenderer(options)
			md.SetRenderer(
				renderer.NewRenderer(
					renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

			var buf bytes.Buffer
			if err := md.Convert(in, &buf); err != nil {
				t.Error(err)
			}

			golden.RequireEqual(t, buf.Bytes())
		})
	}
}

func TestResolveRelativeURL(t *testing.T) {
	for _, tc := range []struct{ base, rel, want string }{
		{"", "cmd/gold/", "cmd/gold/"},
		{"", "./assets/image.png", "./assets/image.png"},
		{"https://charm.land", "/libs", "https://charm.land/libs"},
		{"https://charm.land", "https://example.com", "https://example.com"},
	} {
		if got := resolveRelativeURL(tc.base, tc.rel); got != tc.want {
			t.Errorf("resolveRelativeURL(%q, %q) = %q, want %q", tc.base, tc.rel, got, tc.want)
		}
	}
}

func TestCodeBlockIndentToken(t *testing.T) {
	in := []byte("```go\nfmt.Println(\"hi\")\nfmt.Println(\"bye\")\n```\n")

	renderWith := func(t *testing.T, cb StyleCodeBlock) string {
		t.Helper()
		options := Options{
			Styles: StyleConfig{
				CodeBlock: cb,
			},
		}
		md := goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.DefinitionList,
				emoji.Emoji,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		)
		ar := NewRenderer(options)
		md.SetRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(util.Prioritized(ar, 1000))))

		var buf bytes.Buffer
		if err := md.Convert(in, &buf); err != nil {
			t.Fatalf("convert: %v", err)
		}
		return buf.String()
	}

	indent := uint(1)
	token := "│ "
	got := renderWith(t, StyleCodeBlock{
		StyleBlock: StyleBlock{
			Indent:      &indent,
			IndentToken: &token,
		},
	})
	want := "\n│ fmt.Println(\"hi\")\n│ fmt.Println(\"bye\")\n"
	if got != want {
		t.Errorf("with IndentToken: got %q, want %q", got, want)
	}

	margin := uint(2)
	gotDefault := renderWith(t, StyleCodeBlock{
		StyleBlock: StyleBlock{
			Margin: &margin,
		},
	})
	wantDefault := "\n  fmt.Println(\"hi\")\n  fmt.Println(\"bye\")\n"
	if gotDefault != wantDefault {
		t.Errorf("without IndentToken: got %q, want %q", gotDefault, wantDefault)
	}
}

