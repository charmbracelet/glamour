package glamour

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

	in, err := ioutil.ReadFile(markdown)
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

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	// generate
	if generate {
		err := ioutil.WriteFile(testFile, b, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		return
	}

	// verify
	td, err := ioutil.ReadFile(testFile)
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

	in, err := ioutil.ReadFile(markdown)
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render(string(in))
	if err != nil {
		t.Fatal(err)
	}

	// verify
	td, err := ioutil.ReadFile(testFile)
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

	in, err := ioutil.ReadFile("testdata/preserved_newline.in")
	if err != nil {
		t.Fatal(err)
	}

	b, err := r.Render(string(in))
	if err != nil {
		t.Fatal(err)
	}

	// verify
	td, err := ioutil.ReadFile("testdata/preserved_newline.test")
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

func TestRenderHelpers(t *testing.T) {
	in, err := ioutil.ReadFile(markdown)
	if err != nil {
		t.Fatal(err)
	}

	b, err := Render(string(in), "dark")
	if err != nil {
		t.Error(err)
	}

	// verify
	td, err := ioutil.ReadFile(testFile)
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
	td, err := ioutil.ReadFile("testdata/capitalization.test")
	if err != nil {
		t.Fatal(err)
	}

	if string(td) != b {
		t.Errorf("Rendered output doesn't match!\nExpected: `\n%s`\nGot: `\n%s`\n", td, b)
	}
}

func TestWrapping(t *testing.T) {
	langDir := "testdata/lang-support/"
	files, err := filepath.Glob(langDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("No files found, please resolve paths before trying again.")
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
		goldpath := filepath.Join(langDir, bn+".test")

		t.Run(bn, func(t *testing.T) {
			r, err := NewTermRenderer(
				WithStyles(DarkStyleConfig),
				WithWordWrap(80),
			)
			if err != nil {
				t.Fatal(err)
			}
			// get markdown contents
			in, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			got, err := r.RenderBytes(in)
			if err != nil {
				t.Fatal(err)
			}
			// get desired contents
			want, err := ioutil.ReadFile(goldpath)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(string(got), string(want)); diff != "" {
				t.Fatalf("got != want\n-want +got:\ndiff:\n%s", diff)
			}
		})
	}
}
