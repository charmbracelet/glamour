package ansi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// A chroma color chroma can't parse, such as an ANSI index, which the rest of
// a style file accepts, used to panic inside chroma.MustNewStyle. It should
// come back as an error instead.
func TestCodeBlockInvalidChromaStyle(t *testing.T) {
	// The chroma style is registered once under a fixed name; clear it so this
	// test builds its own, and leave it clear for tests that follow. No lock:
	// these tests don't run in parallel, and before the fix the panic left the
	// mutex held, so locking here would hang instead of failing.
	unregister := func() {
		delete(styles.Registry, chromaStyleTheme)
	}
	unregister()
	t.Cleanup(unregister)

	color := "5"
	options := Options{
		Styles: StyleConfig{
			CodeBlock: StyleCodeBlock{
				Chroma: &Chroma{
					Keyword: StylePrimitive{Color: &color},
				},
			},
		},
	}

	md := goldmark.New()
	md.SetRenderer(renderer.NewRenderer(
		renderer.WithNodeRenderers(util.Prioritized(NewRenderer(options), 1000))))

	var buf bytes.Buffer
	err := md.Convert([]byte("```go\nfunc main() {}\n```\n"), &buf)
	if err == nil {
		t.Fatal("expected an error for an invalid chroma color, got none")
	}
	if !strings.Contains(err.Error(), "invalid code block style") {
		t.Fatalf("unexpected error: %v", err)
	}
}
