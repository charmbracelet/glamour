package mermaid

import (
	"strings"
	"testing"
)

func renderOK(t *testing.T, src string) []string {
	t.Helper()
	lines, err := render(src, 60, unicodeGlyphs)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return lines
}

func mustContain(t *testing.T, lines []string, wants ...string) {
	t.Helper()
	joined := strings.Join(lines, "\n")
	for _, want := range wants {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q in:\n%s", want, joined)
		}
	}
}

func TestStateDiagram(t *testing.T) {
	src := "stateDiagram-v2\n[*] --> Active\nActive --> Inactive : pause\nInactive --> Active : resume\nActive --> [*]"
	lines := renderOK(t, src)
	mustContain(t, lines, "●", "◉", "Active", "Inactive", "resume", "▼", "└─")
}

func TestStateDiagramAliases(t *testing.T) {
	src := "stateDiagram\nstate \"Long waiting\" as W\n[*] --> W\ndirection LR"
	lines := renderOK(t, src)
	mustContain(t, lines, "Long waiting", "●")
}

func TestStateDiagramCyclic(t *testing.T) {
	src := "stateDiagram\n[*] --> A\nA --> B\nB --> A\nB --> [*]"
	lines := renderOK(t, src)
	// The A->B->A cycle renders as a feedback lane instead of an error.
	mustContain(t, lines, "A", "B", "┘")
}

func TestStateDiagramUnsupported(t *testing.T) {
	cases := []struct {
		src   string
		match string
	}{
		{"stateDiagram\nnote right of A: x\nA --> B", "unsupported syntax"},
		{"stateDiagram\nstate C {\n  X\n}", "composite states"},
		{"stateDiagram\nfork\n  X\njoin", "unsupported syntax"},
		{"stateDiagram-v2", "empty"},
	}
	for _, c := range cases {
		_, err := render(c.src, 60, unicodeGlyphs)
		if err == nil || !strings.Contains(err.Error(), c.match) {
			t.Errorf("render(%q) error = %v, want containing %q", c.src, err, c.match)
		}
	}
}

func TestERDiagram(t *testing.T) {
	src := "erDiagram\nCUSTOMER ||--o{ ORDER : places\nORDER ||--|{ ITEM : contains"
	lines := renderOK(t, src)
	mustContain(t, lines, "CUSTOMER", "ORDER", "ITEM", "places", "1:0..*", "1:1..*", "┌")
}

func TestERDiagramAttributes(t *testing.T) {
	src := "erDiagram\nCUSTOMER {\n  string name PK\n  int age\n}\nCUSTOMER ||--o{ ORDER : places"
	lines := renderOK(t, src)
	mustContain(t, lines, "string name PK", "int age")
}

func TestERDiagramBadCardinality(t *testing.T) {
	_, err := render("erDiagram\nA xx--o{ B : x", 60, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "unknown cardinality") {
		t.Errorf("expected cardinality error, got %v", err)
	}
}

func TestClassDiagram(t *testing.T) {
	src := "classDiagram\nclass Animal {\n  +String name\n  +makeSound()\n}\nclass Duck\nAnimal <|-- Duck\nDuck ..|> Flyable"
	lines := renderOK(t, src)
	mustContain(t, lines, "Animal", "+makeSound()", "Duck", "Flyable", "▼", "┄")
}

func TestClassDiagramDirectionsAndLabels(t *testing.T) {
	src := "classDiagram\ndirection LR\nA --> B : uses\nA \"1\" *-- \"n\" C : holds"
	lines := renderOK(t, src)
	mustContain(t, lines, "uses", "holds")
}

func TestClassDiagramOpenLinks(t *testing.T) {
	src := "classDiagram\nA -- B\nA o-- C"
	lines := renderOK(t, src)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "▶") {
		t.Errorf("open links must not draw arrowheads, got:\n%s", joined)
	}
}

func TestClassDiagramUnsupported(t *testing.T) {
	_, err := render("classDiagram\nnote for A \"x\"", 60, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "unsupported syntax") {
		t.Errorf("expected unsupported error, got %v", err)
	}
}

func TestPie(t *testing.T) {
	src := "pie title Pets\n\"dogs\" : 386\n\"cats\" : 85\n\"rats\" : 15"
	lines := renderOK(t, src)
	mustContain(t, lines, "Pets", "dogs", "79%", "17%", "3%", "█")
}

func TestPieErrors(t *testing.T) {
	for _, src := range []string{"pie\n\"a\" : x", "pie\n\"a\":0", "pie"} {
		_, err := render(src, 60, unicodeGlyphs)
		if err == nil {
			t.Errorf("render(%q) unexpectedly succeeded", src)
		}
	}
}

func TestJourney(t *testing.T) {
	src := "journey\ntitle Day\nsection Work\n  Make tea: 5: Me\n  Go home: 3: Me, Cat"
	lines := renderOK(t, src)
	mustContain(t, lines, "Day", "Work", "Make tea", "5", "Me, Cat")
}

func TestJourneyTaskOutsideSection(t *testing.T) {
	_, err := render("journey\ntask: 1: Me", 60, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "outside a section") {
		t.Errorf("expected section error, got %v", err)
	}
}

func TestTimeline(t *testing.T) {
	src := "timeline\ntitle History\n1920 : radio : tv\n1990 : web"
	lines := renderOK(t, src)
	mustContain(t, lines, "History", "1920", "• radio", "• tv", "• web")
}

func TestMindmap(t *testing.T) {
	src := "mindmap\n  root((smart))\n    A\n      A1\n      A2\n    B(round)"
	lines := renderOK(t, src)
	mustContain(t, lines, "smart", "├─ A", "│  ├─ A1", "│  └─ A2", "└─ round")
}

func TestMindmapErrors(t *testing.T) {
	cases := []struct {
		src   string
		match string
	}{
		{"mindmap\n  A\n    B\n  C", "second root"},
		{"mindmap\n  A\n      B", "indentation skip"},
		{"mindmap", "empty"},
	}
	for _, c := range cases {
		_, err := render(c.src, 60, unicodeGlyphs)
		if err == nil || !strings.Contains(err.Error(), c.match) {
			t.Errorf("render(%q) error = %v, want containing %q", c.src, err, c.match)
		}
	}
}

func TestQuadrantChart(t *testing.T) {
	src := "quadrantChart\nx-axis Low --> High\ny-axis Cheap --> Pricey\nquadrant-1 Expand\nquadrant-3 Re-eval\nA: [0.3, 0.6]\nB: [0.8, 0.2]"
	lines := renderOK(t, src)
	mustContain(t, lines, "Expand", "Re-eval", "Pricey", "Cheap", "Low", "High", "┼", "A", "B")
}

func TestQuadrantChartClamps(t *testing.T) {
	src := "quadrantChart\nX: [-1, 2]"
	lines := renderOK(t, src)
	if len(lines) == 0 {
		t.Fatal("expected rendered output")
	}
}

func TestXYChart(t *testing.T) {
	src := "xychart-beta\ntitle Sales\nx-axis [jan, feb, mar]\nbar [10, 24, 30]"
	lines := renderOK(t, src)
	mustContain(t, lines, "Sales", "jan", "feb", "mar", "10", "24", "30", "█")
}

func TestXYChartSeriesMismatch(t *testing.T) {
	_, err := render("xychart-beta\nx-axis [a, b]\nbar [1, 2]\nline [1, 2, 3]", 60, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "lengths differ") {
		t.Errorf("expected mismatch error, got %v", err)
	}
}

func TestGitGraph(t *testing.T) {
	src := "gitGraph\ncommit\ncommit\nbranch feature\ncommit\ncheckout main\ncommit\nmerge feature"
	lines := renderOK(t, src)
	mustContain(t, lines, "main", "feature", "●", "└", "┘", "┄")
}

func TestGitGraphUnknownBranch(t *testing.T) {
	_, err := render("gitGraph\ncheckout nope", 60, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "unknown branch") {
		t.Errorf("expected unknown branch error, got %v", err)
	}
}

func TestAllNewDeterministic(t *testing.T) {
	srcs := []string{
		"stateDiagram\n[*] --> A\nA --> B : x",
		"erDiagram\nA ||--o{ B : x",
		"classDiagram\nA <|-- B",
		"pie\n\"a\" : 1\n\"b\" : 2",
		"mindmap\n  root\n    A",
		"timeline\n2000 : x",
		"journey\ntitle T\nsection S\n  task: 5: Me",
		"quadrantChart\nA: [0.1, 0.2]",
		"xychart-beta\nbar [1, 2]",
		"gitGraph\ncommit\nbranch x\ncommit\nmerge x",
	}
	for _, src := range srcs {
		first, err := render(src, 60, unicodeGlyphs)
		if err != nil {
			t.Fatalf("render(%q): %v", src, err)
		}
		for i := 0; i < 3; i++ {
			again, err := render(src, 60, unicodeGlyphs)
			if err != nil {
				t.Fatalf("render(%q): %v", src, err)
			}
			if strings.Join(first, "\n") != strings.Join(again, "\n") {
				t.Fatalf("render(%q) is not deterministic, run %d", src, i)
			}
		}
	}
}
