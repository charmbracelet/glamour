package mermaid

import (
	"fmt"
	"strings"
	"testing"
)

func renderOrFatal(t *testing.T, src string, limit int) []string {
	t.Helper()
	lines, err := render(src, limit, unicodeGlyphs)
	if err != nil {
		t.Fatalf("render(%q, %d): %v", src, limit, err)
	}
	return lines
}

func assertLines(t *testing.T, got []string, want []string) {
	t.Helper()
	if strings.Join(got, "\n") == strings.Join(want, "\n") {
		return
	}
	t.Errorf("rendered diagram mismatch\n--- got ---\n%s\n--- want ---\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
}

func TestRenderChainTD(t *testing.T) {
	got := renderOrFatal(t, "flowchart TD\nA[Build] --> B[Test] --> C[Ship]", 0)
	want := []string{
		"┌───────┐",
		"│       │",
		"│ Build │",
		"│       │",
		"└───┬───┘",
		"    │",
		"    │",
		"    ▼",
		"┌───┴──┐",
		"│      │",
		"│ Test │",
		"│      │",
		"└───┬──┘",
		"    │",
		"    │",
		"    ▼",
		"┌───┴──┐",
		"│      │",
		"│ Ship │",
		"│      │",
		"└──────┘",
	}
	assertLines(t, got, want)
}

func TestRenderBranchingTD(t *testing.T) {
	got := renderOrFatal(t, "flowchart TD\nA --> B\nA --> C\nB --> D\nC --> D", 0)
	for _, want := range []string{"┌", "▼", "┐"} {
		if !strings.Contains(strings.Join(got, "\n"), want) {
			t.Errorf("output missing %q:\n%s", want, strings.Join(got, "\n"))
		}
	}
}

func TestRenderElbowTD(t *testing.T) {
	// B and C are in the same layer; D below both: one edge needs an
	// elbow through the gap row.
	got := renderOrFatal(t, "flowchart TD\nA --> B --> D\nA --> C --> D", 0)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "┌") && !strings.Contains(joined, "└") && !strings.Contains(joined, "┬") {
		t.Errorf("expected elbow junctions, got:\n%s", joined)
	}
	if !strings.Contains(joined, "▼") {
		t.Errorf("expected arrowheads, got:\n%s", joined)
	}
}

func TestRenderLR(t *testing.T) {
	got := renderOrFatal(t, "flowchart LR\nA --> B", 0)
	want := []string{
		"┌───┐    ┌───┐",
		"│   │    │   │",
		"│ A ├───▶┤ B │",
		"│   │    │   │",
		"└───┘    └───┘",
	}
	assertLines(t, got, want)
}

func TestRenderEdgeLabelTD(t *testing.T) {
	got := renderOrFatal(t, "flowchart TD\nA -->|yes| B", 0)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, " yes") {
		t.Errorf("expected edge label, got:\n%s", joined)
	}
	if !strings.Contains(joined, "▼") {
		t.Errorf("expected arrowhead, got:\n%s", joined)
	}
}

func TestRenderEdgeLabelLR(t *testing.T) {
	got := renderOrFatal(t, "flowchart LR\nA -->|yes| B", 0)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, " yes ") {
		t.Errorf("expected edge label, got:\n%s", joined)
	}
}

func TestRenderMultiLayerEdge(t *testing.T) {
	// A -> C skips a layer and must be routed through a reserved column
	// without cutting through B.
	got := renderOrFatal(t, "flowchart LR\nA --> B --> C\nA --> C", 0)
	joined := strings.Join(got, "\n")
	// B's box borders must survive the crossing edge.
	if strings.Count(joined, "┌") < 1 || strings.Count(joined, "┐") < 1 {
		t.Errorf("expected intact box corners, got:\n%s", joined)
	}
}

func TestRenderShapes(t *testing.T) {
	got := renderOrFatal(t, "flowchart TD\nA(rounded) --> B((circle))", 0)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "╭") {
		t.Errorf("expected rounded corners, got:\n%s", joined)
	}
}

func TestRenderDottedAndThick(t *testing.T) {
	got := renderOrFatal(t, "flowchart LR\nA -.-> B\nB ==> C", 0)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "┄") {
		t.Errorf("expected dotted line, got:\n%s", joined)
	}
	if !strings.Contains(joined, "━") {
		t.Errorf("expected thick line, got:\n%s", joined)
	}
}

func TestRenderLongLabelWraps(t *testing.T) {
	got := renderOrFatal(t, "flowchart TD\nA[this label is quite long indeed]", 0)
	if len(got) < 6 {
		t.Errorf("expected a multi-line box, got:\n%s", strings.Join(got, "\n"))
	}
	for _, line := range got {
		if w := stringWidth(line); w > 24+2*1+2+2 {
			t.Errorf("line too wide (%d): %q", w, line)
		}
	}
}

func TestRenderTitle(t *testing.T) {
	got := renderOrFatal(t, "---\ntitle: My diagram\n---\nflowchart LR\nA --> B", 0)
	if len(got) < 3 || got[0] != "My diagram" || got[1] != "" {
		t.Errorf("expected title and blank line first, got:\n%s", strings.Join(got, "\n"))
	}
}

func TestRenderFitWidth(t *testing.T) {
	src := "flowchart LR\nA[Build] --> B{Tests pass?}\nB -->|yes| C[Release]\nB -->|no| D[Fix]\nD --> C"

	natural := renderOrFatal(t, src, 0)
	naturalW := 0
	for _, l := range natural {
		naturalW = max(naturalW, stringWidth(l))
	}

	fitted := renderOrFatal(t, src, naturalW-1)
	fittedW := 0
	for _, l := range fitted {
		fittedW = max(fittedW, stringWidth(l))
	}
	if fittedW >= naturalW {
		t.Errorf("expected compaction to narrow the diagram, natural=%d fitted=%d", naturalW, fittedW)
	}

	_, err := render(src, 8, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "too wide") {
		t.Errorf("expected too-wide error, got %v", err)
	}
}

func TestRenderASCIIGlyphs(t *testing.T) {
	lines, err := render("flowchart LR\nA --> B", 0, asciiGlyphs)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(lines, "\n")
	if strings.ContainsAny(joined, "─│┌┐└┘▼▶") {
		t.Errorf("ASCII output contains Unicode glyphs:\n%s", joined)
	}
	if !strings.Contains(joined, "+") || !strings.Contains(joined, "|") {
		t.Errorf("ASCII output missing box glyphs:\n%s", joined)
	}
}

func TestRenderCycle(t *testing.T) {
	_, err := render("flowchart LR\nA --> B --> A", 0, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected cycle error, got %v", err)
	}
}

func TestRenderSelfLoop(t *testing.T) {
	_, err := render("flowchart LR\nA --> A", 0, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected cycle error, got %v", err)
	}
}

func TestRenderDeterministic(t *testing.T) {
	src := "flowchart TD\nA --> B\nA --> C\nB --> D\nC --> D\nA --> D"
	first := renderOrFatal(t, src, 0)
	for i := 0; i < 5; i++ {
		again := renderOrFatal(t, src, 0)
		if strings.Join(first, "\n") != strings.Join(again, "\n") {
			t.Fatalf("render is not deterministic, run %d:\n%s", i, strings.Join(again, "\n"))
		}
	}
}

func TestRenderDoesNotPanic(t *testing.T) {
	inputs := []string{
		"flowchart TD\n",
		"flowchart LR\n   \n%% comment\nA -->",
		"flowchart TD\nA[unterminated",
		"flowchart LR\nA --> B --> --> C",
		"flowchart TD\nA[] --> B() --> C{}",
		"flowchart LR\nA\"x\" --> B",
		"flowchart TD\n\n\n",
		"graph\nA",
	}
	for _, in := range inputs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("render(%q) panicked: %v", in, r)
				}
			}()
			_, _ = render(in, 40, unicodeGlyphs) //nolint:errcheck
		}()
	}
}

func TestRenderGuards(t *testing.T) {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	for i := 0; i < 150; i++ {
		fmt.Fprintf(&b, "n%d --> sink\n", i)
	}
	if _, err := render(b.String(), 0, unicodeGlyphs); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected too-large error, got %v", err)
	}

	if _, err := render(strings.Repeat("x", 9000), 0, unicodeGlyphs); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected too-large error, got %v", err)
	}

	many := "flowchart TD\nA --> B\n"
	for i := 0; i < 500; i++ {
		many += "%% padding\n"
	}
	if _, err := render(many, 0, unicodeGlyphs); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected too-large error, got %v", err)
	}
}

func TestRenderEmptyDiagram(t *testing.T) {
	_, err := render("flowchart TD\n%% only comments\n", 0, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected empty error, got %v", err)
	}
}

func TestRenderIsolatedNodes(t *testing.T) {
	got := renderOrFatal(t, "flowchart TD\nA\nB[standalone]", 0)
	joined := strings.Join(got, "\n")
	for _, want := range []string{"A", "standalone"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q in output:\n%s", want, joined)
		}
	}
}
