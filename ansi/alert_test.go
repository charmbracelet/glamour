package ansi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const alertTestPriority = 300

func parseWithAlerts(in string) ast.Node {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithParagraphTransformers(
				util.Prioritized(NewAlertTransformer(), alertTestPriority),
			),
		),
	)
	return md.Parser().Parse(text.NewReader([]byte(in)))
}

func blockquotes(root ast.Node) []*ast.Blockquote {
	var found []*ast.Blockquote
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if bq, ok := n.(*ast.Blockquote); ok {
				found = append(found, bq)
			}
		}
		return ast.WalkContinue, nil
	})
	return found
}

func nodeText(node ast.Node, source []byte) string {
	var b strings.Builder
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok {
			b.Write(t.Segment.Value(source))
			if t.SoftLineBreak() {
				b.WriteString("\n")
			}
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

func TestAlertTransformer(t *testing.T) {
	tests := []struct {
		name string
		in   string
		kind AlertKind
		body string
	}{
		{
			name: "marker with body on the next line",
			in:   "> [!NOTE]\n> Useful information.\n",
			kind: AlertNote,
			body: "Useful information.",
		},
		{
			name: "marker with a multi-line body",
			in:   "> [!TIP]\n> first line\n> second line\n",
			kind: AlertTip,
			body: "first line\nsecond line",
		},
		{
			name: "marker in its own paragraph",
			in:   "> [!WARNING]\n>\n> Careful.\n",
			kind: AlertWarning,
			body: "Careful.",
		},
		{
			name: "marker alone",
			in:   "> [!CAUTION]\n",
			kind: AlertCaution,
		},
		{
			name: "marker followed by a list",
			in:   "> [!IMPORTANT]\n> - one\n",
			kind: AlertImportant,
			body: "one",
		},
		{
			name: "indented marker",
			in:   ">   [!NOTE]\n> body\n",
			kind: AlertNote,
			body: "body",
		},
		{
			name: "carriage returns",
			in:   "> [!NOTE]\r\n> body\r\n",
			kind: AlertNote,
			body: "body",
		},
		{
			name: "trailing text is not an alert",
			in:   "> [!NOTE] trailing text\n> body\n",
			body: "[!NOTE] trailing text\nbody",
		},
		{
			name: "lowercase is not an alert",
			in:   "> [!note]\n> body\n",
			body: "[!note]\nbody",
		},
		{
			name: "unknown type is not an alert",
			in:   "> [!DANGER]\n> body\n",
			body: "[!DANGER]\nbody",
		},
		{
			name: "escaped marker is not an alert",
			in:   "> \\[!NOTE\\]\n> body\n",
			body: "\\[!NOTE\\]\nbody",
		},
		{
			name: "marker after other content is not an alert",
			in:   "> intro\n> [!NOTE]\n",
			body: "intro\n[!NOTE]",
		},
		{
			name: "marker in a later paragraph is not an alert",
			in:   "> intro\n>\n> [!NOTE]\n",
			body: "intro[!NOTE]",
		},
		{
			name: "nested blockquote",
			in:   "> > [!WARNING]\n> > body\n",
			kind: AlertWarning,
			body: "body",
		},
		{
			name: "blockquote in a list item",
			in:   "- > [!TIP]\n  > body\n",
			kind: AlertTip,
			body: "body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.in)
			doc := parseWithAlerts(tt.in)
			found := blockquotes(doc)
			if len(found) == 0 {
				t.Fatal("expected a blockquote")
			}

			// The innermost blockquote holds the alert. Outer blockquotes,
			// if any, are untouched.
			for _, bq := range found[:len(found)-1] {
				if kind, ok := AlertKindFromNode(bq); ok {
					t.Errorf("expected no alert on the outer blockquote, got %q", kind)
				}
			}
			bq := found[len(found)-1]

			kind, ok := AlertKindFromNode(bq)
			if tt.kind == "" {
				if ok {
					t.Fatalf("expected no alert, got %q", kind)
				}
			} else {
				if !ok {
					t.Fatalf("expected a %q alert", tt.kind)
				}
				if kind != tt.kind {
					t.Errorf("expected a %q alert, got %q", tt.kind, kind)
				}
			}

			if got := nodeText(bq, source); got != tt.body {
				t.Errorf("expected body %q, got %q", tt.body, got)
			}
		})
	}
}

func TestAlertTransformerOutsideBlockquote(t *testing.T) {
	source := []byte("[!NOTE]\n")
	doc := parseWithAlerts(string(source))

	if found := blockquotes(doc); len(found) != 0 {
		t.Fatalf("expected no blockquote, got %d", len(found))
	}
	if got := nodeText(doc, source); got != "[!NOTE]" {
		t.Errorf("expected the marker to be left alone, got %q", got)
	}
}

func TestAlertTransformerDropsMarkerParagraph(t *testing.T) {
	found := blockquotes(parseWithAlerts("> [!NOTE]\n"))
	if len(found) != 1 {
		t.Fatalf("expected a single blockquote, got %d", len(found))
	}
	if n := found[0].ChildCount(); n != 0 {
		t.Errorf("expected the marker paragraph to be dropped, got %d children", n)
	}
}

func renderWithAlerts(t *testing.T, options Options, in string) string {
	t.Helper()

	source := []byte(in)
	doc := parseWithAlerts(in)

	ar := NewRenderer(options)
	r := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(ar, 1000)))

	var buf bytes.Buffer
	if err := r.Render(&buf, source, doc); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// TestAlertFallbackStyle checks that alerts still render a title line when the
// style doesn't know about them, e.g. a custom stylesheet.
func TestAlertFallbackStyle(t *testing.T) {
	out := renderWithAlerts(t, Options{WordWrap: 40}, "> [!NOTE]\n> Useful information.\n")

	if !strings.Contains(out, "Note") {
		t.Errorf("expected an alert title, got %q", out)
	}
	if !strings.Contains(out, "Useful information.") {
		t.Errorf("expected the alert body, got %q", out)
	}
	if strings.Contains(out, "[!NOTE]") {
		t.Errorf("expected the marker to be consumed, got %q", out)
	}
}

// TestAlertIndentTokenStyle checks that the alert's quote line can be colored
// without coloring the body.
func TestAlertIndentTokenStyle(t *testing.T) {
	options := Options{WordWrap: 40}
	options.Styles.Alerts.Note = StyleAlert{
		StyleBlock: StyleBlock{
			Indent:           uintPtr(1),
			IndentToken:      stringPtr("│ "),
			IndentTokenStyle: &StylePrimitive{Color: stringPtr("39")},
		},
	}

	out := renderWithAlerts(t, options, "> [!NOTE]\n> body\n")

	if !strings.Contains(out, "\x1b[38;5;39m│ ") {
		t.Errorf("expected a colored quote line, got %q", out)
	}
	if strings.Contains(out, "\x1b[38;5;39mbody") {
		t.Errorf("expected the body to keep the default color, got %q", out)
	}
}

func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }

func TestAlertKindTitle(t *testing.T) {
	tests := map[AlertKind]string{
		AlertNote:      "Note",
		AlertTip:       "Tip",
		AlertImportant: "Important",
		AlertWarning:   "Warning",
		AlertCaution:   "Caution",
		"":             "",
	}
	for kind, want := range tests {
		if got := kind.Title(); got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	}
}

func TestStyleAlertsFor(t *testing.T) {
	alerts := StyleAlerts{
		Note:    StyleAlert{Title: StylePrimitive{Prefix: "note"}},
		Tip:     StyleAlert{Title: StylePrimitive{Prefix: "tip"}},
		Warning: StyleAlert{Title: StylePrimitive{Prefix: "warning"}},
	}

	if got := alerts.For(AlertNote).Title.Prefix; got != "note" {
		t.Errorf("expected the note style, got %q", got)
	}
	if got := alerts.For(AlertTip).Title.Prefix; got != "tip" {
		t.Errorf("expected the tip style, got %q", got)
	}
	if got := alerts.For(AlertWarning).Title.Prefix; got != "warning" {
		t.Errorf("expected the warning style, got %q", got)
	}
	if got := alerts.For(AlertCaution); got.Title.Prefix != "" || got.Indent != nil {
		t.Errorf("expected a zero style for unset kinds, got %+v", got)
	}
	if got := alerts.For(AlertKind("bogus")); got.Title.Prefix != "" || got.Indent != nil {
		t.Errorf("expected a zero style for unknown kinds, got %+v", got)
	}
}
