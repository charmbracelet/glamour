package ansi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

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
	examplesDir      = "../styles/examples/"
	issuesDir        = "../testdata/issues/"
)

func TestRenderer(t *testing.T) {
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
		err = json.Unmarshal(b, &options.Styles)
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
			err = ioutil.WriteFile(tn, buf.Bytes(), 0644)
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
	files, err := filepath.Glob(issuesDir + "*.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		bn := strings.TrimSuffix(filepath.Base(f), ".md")
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
		err = json.Unmarshal(b, &options.Styles)
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
			err = ioutil.WriteFile(tn, buf.Bytes(), 0644)
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

func TestUrlResolver(t *testing.T) {
	// resolveURL should be compliant with https://tools.ietf.org/html/rfc3986#section-5.2.2

	host := "https://example.com"
	host2 := "https://elpmaxe.moc"

	assertEqual(t, resolveURL("a/b", "/c/d"), "/c/d")
	assertEqual(t, resolveURL("a/b", "./c/d"), "/a/c/d")
	assertEqual(t, resolveURL("a/b", "../c/d"), "/c/d")
	assertEqual(t, resolveURL("a/b/", "./c/d"), "/a/b/c/d")
	assertEqual(t, resolveURL("a/b/", "c/d"), "/a/b/c/d")
	assertEqual(t, resolveURL("a/b/", host2+"/c/d"), host2+"/c/d")

	assertEqual(t, resolveURL("/a/b", "/c/d"), "/c/d")
	assertEqual(t, resolveURL("/a/b", "./c/d"), "/a/c/d")
	assertEqual(t, resolveURL("/a/b", "../c/d"), "/c/d")
	assertEqual(t, resolveURL("/a/b/", "./c/d"), "/a/b/c/d")
	assertEqual(t, resolveURL("/a/b/", "c/d"), "/a/b/c/d")
	assertEqual(t, resolveURL("/a/b/", host2+"/c/d"), host2+"/c/d")

	assertEqual(t, resolveURL(host+"/a/b", "/c/d"), host+"/c/d")
	assertEqual(t, resolveURL(host+"/a/b", "./c/d"), host+"/a/c/d")
	assertEqual(t, resolveURL(host+"/a/b", "../c/d"), host+"/c/d")
	assertEqual(t, resolveURL(host+"/a/b/", "./c/d"), host+"/a/b/c/d")
	assertEqual(t, resolveURL(host+"/a/b/", "c/d"), host+"/a/b/c/d")
	assertEqual(t, resolveURL(host+"/a/b/", host2+"/c/d"), host2+"/c/d")
}

func TestUrlResolverForLocalFiles(t *testing.T) {
	base := "/home/foobar/project/"
	assertEqual(t,
		resolveURL(base, "/assets/logo.png"),
		"/home/foobar/project/assets/logo.png")
}

func TestUrlResolverForGitforgeUsage(t *testing.T) {
	// in git-forges like github & gitea, URLs are often written relative,
	// where the repo-url (`gitea.com/gitea/tea`) treated as reference.
	// The base URL needs to be treated with a trailing slash, acting as a "directory",
	// requiring preprocessing of the URLs by applications (appending trailing slash to base URL).
	base := "https://gitea.com/gitea/tea/"
	assertEqual(t,
		resolveURL(base, "src/branch/master/modules/print/markdown.go"),
		"https://gitea.com/gitea/tea/src/branch/master/modules/print/markdown.go")

	base = "https://raw.githubusercontent.com/foo/bar/master/"
	assertEqual(t,
		resolveURL(base, "/assets/logo.png"),
		"https://raw.githubusercontent.com/foo/bar/master/assets/logo.png")

	base = "https://raw.githubusercontent.com/foo/bar/master/some/dir/"
	assertEqual(t,
		resolveURL(base, "/assets/logo.png"),
		"https://raw.githubusercontent.com/foo/bar/master/assets/logo.png")
}

func assertEqual(t *testing.T, value interface{}, expected interface{}) {
	if value != expected {
		t.Errorf("Expected '%v', but got '%v'", expected, value)
	}
}
