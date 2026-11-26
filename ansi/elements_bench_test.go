package ansi

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func BenchmarkNewElementCodeBlock(b *testing.B) {
	const line = "var value = 1234567890 // example\n"
	for _, kind := range []ast.NodeKind{ast.KindFencedCodeBlock, ast.KindCodeBlock} {
		for _, size := range []int{8 << 10, 32 << 10, 128 << 10} {
			b.Run(fmt.Sprintf("%s/%dKiB", kind, size>>10), func(b *testing.B) {
				body := strings.Repeat(line, size/len(line))
				markdown := "```go\n" + body + "```\n"
				if kind == ast.KindCodeBlock {
					markdown = "    " + strings.ReplaceAll(body[:len(body)-1], "\n", "\n    ") + "\n"
				}
				source := []byte(markdown)
				// Parse once: measure accumulation, not Markdown parsing or highlighting.
				node := findCodeBlock(b, source, kind)
				renderer := &ANSIRenderer{}
				b.SetBytes(int64(len(body)))
				b.ReportAllocs()
				for b.Loop() {
					renderer.NewElement(node, source)
				}
			})
		}
	}
}
