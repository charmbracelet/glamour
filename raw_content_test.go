package glamour

import (
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
)

func TestRawContentPreservesBackslashes(t *testing.T) {
	const raw = `price: 3\.14, match: foo\+bar, path: C:\\tmp`
	for _, style := range []string{"ascii", "dark"} {
		for _, tc := range []struct {
			name, markdown, want string
		}{
			{"fenced", "```regexp\n" + raw + "\n```\n", raw},
			{"fenced without language", "```\n" + raw + "\n```\n", raw},
			{"indented", "    " + raw + "\n", raw},
			{"html block", "<div>" + raw + "</div>\n", raw},
			{"inline code", "`" + raw + "`\n", raw},
			{"prose", `foo\+bar and 3\.14` + "\n", "foo+bar and 3.14"},
		} {
			t.Run(style+"/"+tc.name, func(t *testing.T) {
				r, err := NewTermRenderer(WithStandardStyle(style), WithWordWrap(120))
				if err != nil {
					t.Fatal(err)
				}
				out, err := r.Render(tc.markdown)
				if err != nil {
					t.Fatal(err)
				}
				if plain := xansi.Strip(out); !strings.Contains(plain, tc.want) {
					t.Fatalf("rendered %q does not contain %q", plain, tc.want)
				}
			})
		}
	}
}

func TestRawHTMLStillSanitized(t *testing.T) {
	r, err := NewTermRenderer(WithStandardStyle("ascii"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Render("<div>foo\\+bar<script>alert(1)</script></div>\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `foo\+bar`) || strings.Contains(out, "alert") || strings.Contains(out, "<script") {
		t.Fatalf("unexpected sanitized HTML: %q", out)
	}
}
