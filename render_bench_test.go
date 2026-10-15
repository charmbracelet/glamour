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

func diagramDoc(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "## Heading %d\n\nSome prose around the diagram.\n\n```mermaid\nflowchart TD\nA%d[Build] --> B%d{Tests pass?}\nB%d -->|yes| C%d[Release]\nB%d -->|no| D%d[Fix bugs]\nD%d --> C%d\n```\n\n", i, i, i, i, i, i, i, i, i)
	}
	return b.String()
}

func BenchmarkRenderDiagram(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("diagrams=%d", n), func(b *testing.B) { benchRender(b, diagramDoc(n)) })
	}
}

func BenchmarkRenderDiagramDisabled(b *testing.B) {
	doc := diagramDoc(10)
	r, err := NewTermRenderer(WithStandardStyle("dark"), WithWordWrap(120), WithMermaid(false))
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
