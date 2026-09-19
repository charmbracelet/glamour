//go:build glamour_nochroma

package glamour

import (
	"strings"
	"testing"

	"charm.land/glamour/v2/styles"
)

func TestCodeBlockWithoutChroma(t *testing.T) {
	r, err := NewTermRenderer(WithStandardStyle(styles.DarkStyle))
	if err != nil {
		t.Fatal(err)
	}

	const code = "func main() {}"
	b, err := r.Render("```go\n" + code + "\n```\n")
	if err != nil {
		t.Fatal(err)
	}

	// Chroma would split the line into one styled token per lexeme.
	if !strings.Contains(b, code) {
		t.Errorf("expected the code block to render as one plain run, got %q", b)
	}
}
