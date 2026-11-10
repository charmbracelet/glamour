package glamour

import (
	"strings"
	"testing"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	"github.com/charmbracelet/x/exp/golden"
)

func TestStrikethroughInlineContent(t *testing.T) {
	for name, style := range map[string]ansi.StyleConfig{
		"notty": styles.NoTTYStyleConfig,
		"dark":  styles.DarkStyleConfig,
	} {
		t.Run(name, func(t *testing.T) {
			style.Strikethrough = ansi.StylePrimitive{}
			r, err := NewTermRenderer(WithStyles(style), WithWordWrap(120), WithEmoji(), WithBaseURL("https://example.com/base/"))
			if err != nil {
				t.Fatal(err)
			}
			for _, in := range []string{
				"plain",
				"[docs](https://example.com)",
				"[docs](/guide)",
				"[section](#section)",
				"[invalid](http://example.com/%zz)",
				"<https://example.com>",
				"<hello@example.com>",
				"![image](https://example.com/image.png)",
				"`&gt; \\*literal\\*`",
				"*emphasis* and **strong**",
				"before [docs](https://example.com) after `code` :smile:",
			} {
				t.Run(in, func(t *testing.T) {
					want, err := r.Render(in)
					if err != nil {
						t.Fatal(err)
					}
					got, err := r.Render("~~" + in + "~~")
					if err != nil {
						t.Fatal(err)
					}
					if got != want {
						t.Errorf("strikethrough changed inline rendering:\ngot  %q\nwant %q", got, want)
					}
				})
			}
		})
	}
}

func TestStrikethroughStyles(t *testing.T) {
	const in = "~~plain~~\n\n~~before [docs](https://example.com) after `code` *emphasis* **strong**~~\n\n**~~[docs](https://example.com)~~**\n\n[~~docs~~](https://example.com)\n\n| resource |\n| --- |\n| ~~[docs](https://example.com)~~ |\n"
	for _, name := range []string{"ascii", "notty", "dark", "light", "pink", "dracula", "tokyo-night"} {
		t.Run(name, func(t *testing.T) {
			r, err := NewTermRenderer(WithStandardStyle(name), WithWordWrap(120))
			if err != nil {
				t.Fatal(err)
			}
			out, err := r.Render(in)
			if err != nil {
				t.Fatal(err)
			}
			golden.RequireEqual(t, []byte(out))
		})
	}
}

func TestStrikethroughDelimiters(t *testing.T) {
	r, err := NewTermRenderer(WithStyles(ansi.StyleConfig{
		Strikethrough: ansi.StylePrimitive{BlockPrefix: "<", Prefix: "[", Suffix: "]", BlockSuffix: ">"},
		Emph:          ansi.StylePrimitive{Prefix: "*", Suffix: "*"},
		Code:          ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "`", Suffix: "`"}},
	}), WithWordWrap(120))
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Render("~~before *emphasis* after `code`~~")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(out), "<[before *emphasis* after `code`]>"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStrikethroughErrors(t *testing.T) {
	for name, style := range map[string]ansi.StyleConfig{
		"container": {Strikethrough: ansi.StylePrimitive{Format: "{{"}},
		"child":     {LinkText: ansi.StylePrimitive{Format: "{{"}},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := NewTermRenderer(WithStyles(style))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := r.Render("~~[docs](https://example.com)~~"); err == nil {
				t.Fatal("expected invalid style template to return an error")
			}
		})
	}
}

func TestStrikethroughLinkModes(t *testing.T) {
	for _, inline := range []bool{false, true} {
		for _, mode := range []ansi.HyperlinkMode{ansi.HyperlinkModeAuto, ansi.HyperlinkModeInline} {
			r, err := NewTermRenderer(WithInlineTableLinks(inline), WithHyperlinkMode(mode))
			if err != nil {
				t.Fatal(err)
			}
			for _, in := range []string{
				"~~[docs](https://example.com/Guide)~~",
				"| resource |\n| --- |\n| ~~[docs](https://example.com/Guide)~~ |",
			} {
				out, err := r.Render(in)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(out, "https://example.com/Guide") || !strings.Contains(out, "\x1b]8;") {
					t.Errorf("inline table=%v, hyperlink mode=%v: lost destination or OSC 8 hyperlink: %q", inline, mode, out)
				}
			}
		}
	}
}

func TestStrikethroughFormat(t *testing.T) {
	for _, tc := range []struct{ format, want string }{
		{"[{{.text}}]", "[Docs]"},
		{"{{ToUpper .text}}", "DOCS"},
		{"{{ToLower .text}}", "docs"},
		{"{{Left .text 2}}", "Do"},
	} {
		t.Run(tc.format, func(t *testing.T) {
			style := ansi.StyleConfig{
				Strikethrough: ansi.StylePrimitive{Format: tc.format, BlockPrefix: "~", BlockSuffix: "~"},
				Strong:        ansi.StylePrimitive{Format: tc.format, BlockPrefix: "*", BlockSuffix: "*"},
			}
			r, err := NewTermRenderer(WithStyles(style))
			if err != nil {
				t.Fatal(err)
			}
			for in, want := range map[string]string{
				"~~[Docs](https://example.com/Guide)~~":     "~" + tc.want + "~",
				"**~~[Docs](https://example.com/Guide)~~**": "*" + tc.want + "*",
			} {
				out, err := r.Render(in)
				if err != nil {
					t.Fatal(err)
				}
				if got := strings.TrimSpace(out); got != want {
					t.Errorf("got %q, want %q", got, want)
				}
			}
		})
	}
}

func TestStrikethroughCasing(t *testing.T) {
	yes := true
	for _, tc := range []struct {
		name  string
		style ansi.StylePrimitive
		want  string
	}{
		{"upper", ansi.StylePrimitive{Upper: &yes}, "DOCS"},
		{"lower", ansi.StylePrimitive{Lower: &yes}, "docs"},
		{"title", ansi.StylePrimitive{Title: &yes}, "Docs"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewTermRenderer(WithStyles(ansi.StyleConfig{Strikethrough: tc.style}))
			if err != nil {
				t.Fatal(err)
			}
			out, err := r.Render("~~[Docs](https://example.com/Guide)~~")
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(out); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestStrikethroughInheritedCasing(t *testing.T) {
	yes := true
	for _, tc := range []struct {
		name  string
		style ansi.StylePrimitive
		want  string
	}{
		{"upper", ansi.StylePrimitive{Upper: &yes}, "~DOCS~"},
		{"lower", ansi.StylePrimitive{Lower: &yes}, "~docs~"},
		{"title", ansi.StylePrimitive{Title: &yes}, "~Docs~"},
	} {
		for _, block := range []string{"document", "heading"} {
			t.Run(block+"/"+tc.name, func(t *testing.T) {
				style := ansi.StyleConfig{
					Strikethrough: ansi.StylePrimitive{BlockPrefix: "~", BlockSuffix: "~"},
				}
				if block == "document" {
					style.Document.StylePrimitive = tc.style
				} else {
					style.Heading.StylePrimitive = tc.style
				}
				r, err := NewTermRenderer(WithStyles(style))
				if err != nil {
					t.Fatal(err)
				}
				for _, in := range []string{
					"~~Docs~~",
					"~~[Docs](https://example.com/Guide)~~",
					"**~~[Docs](https://example.com/Guide)~~**",
				} {
					if block == "heading" {
						in = "# " + in
					}
					out, err := r.Render(in)
					if err != nil {
						t.Fatal(err)
					}
					if got := strings.TrimSpace(out); got != tc.want {
						t.Errorf("%q: got %q, want %q", in, got, tc.want)
					}
				}
			})
		}
	}
}

func TestStrikethroughBlockFormat(t *testing.T) {
	for _, block := range []string{"document", "heading"} {
		t.Run(block, func(t *testing.T) {
			style := ansi.StyleConfig{
				Strikethrough: ansi.StylePrimitive{BlockPrefix: "~", BlockSuffix: "~"},
			}
			format := ansi.StylePrimitive{Format: "[{{ToUpper .text}}]"}
			if block == "document" {
				style.Document.StylePrimitive = format
			} else {
				style.Heading.StylePrimitive = format
			}
			r, err := NewTermRenderer(WithStyles(style))
			if err != nil {
				t.Fatal(err)
			}
			for _, in := range []string{"~~Docs~~", "~~[Docs](https://example.com/Guide)~~"} {
				if block == "heading" {
					in = "# " + in
				}
				out, err := r.Render(in)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(out, "DOCS") {
					t.Errorf("block Format was applied to inline content: %q", out)
				}
				if !strings.Contains(in, "https://") && strings.TrimSpace(out) != "~Docs~" {
					t.Errorf("block Format changed plain strike: %q", out)
				}
				if strings.Contains(in, "https://") && !strings.Contains(out, "https://example.com/Guide") {
					t.Errorf("unused block Format flattened the link: %q", out)
				}
			}
		})
	}
}

func TestStrikethroughClearedFormat(t *testing.T) {
	style := ansi.StyleConfig{Strikethrough: ansi.StylePrimitive{Format: "[{{ToUpper .text}}]"}}
	r, err := NewTermRenderer(WithStyles(style))
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Render("**~~[Docs](https://example.com/Guide)~~**")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "DOCS") || !strings.Contains(out, "https://example.com/Guide") {
		t.Errorf("empty outer override did not clear strike Format: %q", out)
	}
}
