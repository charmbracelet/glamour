package ansi

import (
	"encoding/json"
	"os"
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

// TestBlockquoteIndentTokenConsistency verifies that every non-blank line
// inside a blockquote carries the indent token (│). Before the fix, the
// Width() method counted indent tokens by their count (*Indent=1) instead
// of their visual width (│ = 2 columns). Paragraphs inside blockquotes
// would wrap too wide by 1 column; the terminal would then hard-break the
// line and the continuation line would lack the │ token.
func TestBlockquoteIndentTokenConsistency(t *testing.T) {
	tests := []struct {
		name string
		md   string
	}{
		{
			name: "CJK with emoji",
			md:   "> ⚠️ 以上信息基于我的训练数据，可能不完全是最新的。如果你需要了解该仓库最新的 star数、最新版本、或者作者的最新动态",
		},
		{
			name: "pure CJK",
			md:   "> 以上信息基于我的训练数据，可能不完全是最新的。如果你需要了解该仓库最新的 star数、最新版本、或者作者的最新动态",
		},
		{
			name: "CJK with trailing emoji",
			md:   "> ⚠️ 以上信息基于我的训练数据，可能不完全是最新的。如果你需要了解该仓库最新的 star数、最新版本、或者作者的最新动态，我可以用工具帮你查询 😄",
		},
	}

	for _, width := range []int{40, 60, 76, 80} {
		for _, tt := range tests {
			name := tt.name + "_w" + intToStr(width)
			t.Run(name, func(t *testing.T) {
				buf := renderMD(t, tt.md, width)
				lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

				for i, line := range lines {
					if i == 0 && strings.TrimSpace(line) == "" {
						continue // blank prefix line is ok
					}
					clean := stripAnsi(line)
					trimmed := strings.TrimSpace(clean)
					if trimmed == "" {
						continue
					}
					if !strings.Contains(clean, "│") {
						t.Errorf("line %d missing indent token: %q", i, clean)
					}
				}
			})
		}
	}
}

// TestBlockquoteCJKGolden is a golden-file test for CJK blockquote rendering.
func TestBlockquoteCJKGolden(t *testing.T) {
	md := "> ⚠️ 以上信息基于我的训练数据，可能不完全是最新的。如果你需要了解该仓库最新的 star数、最新版本、或者作者的最新动态，我可以用工具帮你查询——不过你刚才说了禁止用工具，所以我就基于已有知识回答了 😄\n"
	buf := renderMD(t, md, 80)
	golden.RequireEqual(t, []byte(buf.String()))
}

func renderMD(t *testing.T, md string, wordWrap int) *strings.Builder {
	t.Helper()
	options := Options{WordWrap: wordWrap}
	options.Styles = darkStyles(t)

	markdown := goldmark.New(
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
	markdown.SetRenderer(
		renderer.NewRenderer(
			renderer.WithNodeRenderers(util.Prioritized(ar, 1000)),
		),
	)

	var buf strings.Builder
	if err := markdown.Convert([]byte(md), &buf); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func darkStyles(t *testing.T) StyleConfig {
	t.Helper()
	b, err := os.ReadFile("../styles/dark.json")
	if err != nil {
		t.Fatal(err)
	}
	var styles StyleConfig
	if err := json.Unmarshal(b, &styles); err != nil {
		t.Fatal(err)
	}
	return styles
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	digits := make([]byte, 0, 8)
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

func stripAnsi(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		if r == '\x1b' {
			esc = true
			continue
		}
		if esc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				esc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
