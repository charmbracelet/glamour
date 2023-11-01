package ansi

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/scrapbook"

	"github.com/muesli/termenv"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const (
	generateExamples = false
	generateIssues   = false
	generateHeadings = false
	examplesDir      = "../styles/examples/"
	issuesDir        = "../testdata/issues/"
)

func TestRenderer(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	files, err := filepath.Glob(examplesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		sn := filepath.Join(examplesDir, bn+".style")
		tn := filepath.Join("../testdata", bn+".test")

		in, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		b, err := ioutil.ReadFile(sn)
		if err != nil {
			t.Fatal(err)
		}

		options := Options{
			WordWrap:     80,
			ColorProfile: termenv.TrueColor,
		}
		options.Styles, err = scrapbook.ImportJSONBytes(b)
		if err != nil {
			t.Fatal(err)
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
		err = md.Convert(in, &buf)
		if err != nil {
			t.Error(err)
		}

		// generate
		if generateExamples {
			err = ioutil.WriteFile(tn, buf.Bytes(), 0o644)
			if err != nil {
				t.Fatal(err)
			}
			continue
		}

		// verify
		td, err := ioutil.ReadFile(tn)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(td, buf.Bytes()) {
			t.Errorf("Rendered output for %s doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
				bn, string(td), buf.String())
		}
	}
}

func TestRendererIssues(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	files, err := filepath.Glob(issuesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(bn, func(t *testing.T) {
			tn := filepath.Join(issuesDir, bn+".test")

			in, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			b, err := ioutil.ReadFile("../styles/dark.json")
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap:     80,
				ColorProfile: termenv.TrueColor,
			}
			options.Styles, err = scrapbook.ImportJSONBytes(b)
			if err != nil {
				t.Fatal(err)
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
			err = md.Convert(in, &buf)
			if err != nil {
				t.Error(err)
			}

			// generate
			if generateIssues {
				err = ioutil.WriteFile(tn, buf.Bytes(), 0o644)
				if err != nil {
					t.Fatal(err)
				}
				return
			}

			// verify
			td, err := ioutil.ReadFile(tn)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(td, buf.Bytes()) {
				t.Errorf("Rendered output for %s doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
					bn, string(td), buf.String())
			}
		})
	}
}

// TODO generate the test files with correct ansi sequences.
func TestHeadings(t *testing.T) {
	headingDir := "../testdata/headings/"
	files, err := filepath.Glob(headingDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(bn, func(t *testing.T) {
			tn := filepath.Join(headingDir, bn+".test")

			// use custom style if there is one. Name must match the markdown
			// file. If not, use dark theme.
			styleSheet := filepath.Join(headingDir, bn+".style")
			if _, err := os.Stat(styleSheet); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					styleSheet = "../styles/dark.json"
				} else {
					t.Fatal(err)
				}
			}

			b, err := ioutil.ReadFile(styleSheet)
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap:     80,
				ColorProfile: termenv.TrueColor,
			}
			options.Styles, err = scrapbook.ImportJSONBytes(b)
			if err != nil {
				t.Fatal(err)
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
			in, err := os.ReadFile(f)
			if err != nil {
				t.Error(err)
			}
			err = md.Convert(in, &buf)
			if err != nil {
				t.Error(err)
			}

			// if we don't have a test file, make one.
			_, err = os.Stat(tn)
			if generateHeadings || err != nil {
				err = os.WriteFile(tn, buf.Bytes(), 0o644)
				if err != nil {
					t.Fatal(err)
				}
			}

			want, err := os.ReadFile(tn)
			if err != nil {
				t.Fatal(err)
			}

			// Uncomment when the desired result exists.
			//if !bytes.Equal(want, buf.Bytes()) {
			//	t.Errorf("Rendered output for headings don't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
			//		string(want), buf.String())
			//}

			fmt.Printf("Rendered output for headings don't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
				string(want), buf.String())
		})
	}
}

func TestCodeBlocks(t *testing.T) {
	dir := "../testdata/code/"
	files, err := filepath.Glob(dir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		if f == dir+"code.md" {
			fmt.Println("yo")
		}
		t.Run(bn, func(t *testing.T) {
			tn := filepath.Join(dir, bn+".test")

			// use custom style if there is one. Name must match the markdown
			// file. If not, use dark theme.
			styleSheet := filepath.Join(dir, bn+".style")
			if _, err := os.Stat(styleSheet); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					styleSheet = "../styles/dark.json"
				} else {
					t.Fatal(err)
				}
			}

			b, err := ioutil.ReadFile(styleSheet)
			if err != nil {
				t.Fatal(err)
			}

			options := Options{
				WordWrap:     80,
				ColorProfile: termenv.TrueColor,
			}
			options.Styles, err = scrapbook.ImportJSONBytes(b)
			if err != nil {
				t.Fatal(err)
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
			in, err := os.ReadFile(f)
			if err != nil {
				t.Error(err)
			}
			err = md.Convert(in, &buf)
			if err != nil {
				t.Error(err)
			}

			// if we don't have a test file, make one.
			_, err = os.Stat(tn)
			if generateHeadings || err != nil {
				err = os.WriteFile(tn, buf.Bytes(), 0o644)
				if err != nil {
					t.Fatal(err)
				}
			}

			want, err := os.ReadFile(tn)
			if err != nil {
				t.Fatal(err)
			}

			// Uncomment when the desired result exists.
			//if !bytes.Equal(want, buf.Bytes()) {
			//	t.Errorf("Rendered output for headings don't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
			//		string(want), buf.String())
			//}

			fmt.Printf("Rendered output for headings don't match!\nExpected: `\n%s`\nGot: `\n%s`\n",
				string(want), buf.String())
		})
	}
}
