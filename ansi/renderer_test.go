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

func TestRendererDoesNotDecodeControlEntitiesAcrossNodes(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	md.SetRenderer(renderer.NewRenderer(
		renderer.WithNodeRenderers(util.Prioritized(NewRenderer(Options{}), 1000))))

	var buf bytes.Buffer
	if err := md.Convert([]byte("~~&**#27;**[31mFAKE~~"), &buf); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "&#27;[31mFAKE\n" {
		t.Fatalf("unexpected render: %q", got)
	}
}

func TestUnescapeHTMLNumericReferencePolicy(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"decimal escape", "&#27;", "&#27;"},
		{"hex escape", "&#x1b;", "&#x1b;"},
		{"uppercase hex escape", "&#X1B;", "&#X1B;"},
		{"semicolonless decimal escape", "&#27", "&#27"},
		{"semicolonless hex escape", "&#x1b", "&#x1b"},
		{"carriage return", "&#13;", "&#13;"},
		{"delete", "&#127;", "&#127;"},
		{"C1 control", "&#129;", "&#129;"},
		{"tab", "&#9;", "\t"},
		{"newline", "&#10;", "\n"},
		{"printable", "&#65;", "A"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := unescapeHTML(test.in); got != test.want {
				t.Fatalf("unescapeHTML(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

func TestRendererDoesNotDecodeControlEntitiesInHTMLBlock(t *testing.T) {
	md := goldmark.New()
	md.SetRenderer(renderer.NewRenderer(
		renderer.WithNodeRenderers(util.Prioritized(NewRenderer(Options{}), 1000))))

	var buf bytes.Buffer
	if err := md.Convert([]byte("<div>\n&#27;[31mFAKE\n</div>\n"), &buf); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "&#27;[31mFAKE" {
		t.Fatalf("unexpected render: %q", got)
	}
}
