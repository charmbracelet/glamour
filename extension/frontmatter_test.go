package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"
	"testing"
)

func TestFrontMatterParsing(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			DefaultFrontMatterParser,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/frontmatter.txt", t, testutil.ParseCliCaseArg()...)
}
