package glamour

import (
	"fmt"
	"strings"
	"testing"
)

func proseDoc(paragraphs int) string {
	var b strings.Builder
	for i := range paragraphs {
		fmt.Fprintf(&b, "Considering angle %d of the question in some detail, the constraint holds for every branch here and the wrapping needs to run across several lines to be representative.\n\n", i)
	}
	return b.String()
}

func mixedDoc(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "## Heading %d\n\nSome **bold** and _italic_ prose with `code` and a [link](https://example.com) that wraps.\n\n- item one\n- item two\n\n```go\nfunc f() { return }\n```\n\n", i)
	}
	return b.String()
}

func benchRender(b *testing.B, doc string) {
	r, err := NewTermRenderer(WithStandardStyle("dark"), WithWordWrap(120))
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(doc)))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := r.Render(doc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderProse(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("paras=%d", n), func(b *testing.B) { benchRender(b, proseDoc(n)) })
	}
}

func BenchmarkRenderMixed(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("blocks=%d", n), func(b *testing.B) { benchRender(b, mixedDoc(n)) })
	}
}
