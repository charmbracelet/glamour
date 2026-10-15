package mermaid

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func renderSeq(t *testing.T, src string) []string {
	t.Helper()
	lines, err := render(src, 0, unicodeGlyphs)
	if err != nil {
		t.Fatalf("render sequence: %v", err)
	}
	return lines
}

func TestSequenceBasic(t *testing.T) {
	got := renderSeq(t, "sequenceDiagram\nA->>B: hi\nB-->>A: ok")
	want := []string{
		"┌───┐      ┌───┐",
		"│   │      │   │",
		"│ A │      │ B │",
		"│   │      │   │",
		"└───┘      └───┘",
		"",
		"  ┆          ┆",
		"  ┆   hi     ┆",
		"  ┆─────────▶┆",
		"  ┆          ┆",
		"  ┆   ok     ┆",
		"  ┆◀┄┄┄┄┄┄┄┄┄┆",
		"  ┆          ┆",
	}
	assertLines(t, got, want)
}

func TestSequenceParticipantsAndSelf(t *testing.T) {
	got := renderSeq(t, "sequenceDiagram\nparticipant U as user\nU->>U: refresh")
	joined := strings.Join(got, "\n")
	for _, want := range []string{"┌──────┐", "│ user │", "refresh", "├──────────┐", "◀──────────┘"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q, got:\n%s", want, joined)
		}
	}
}

func TestSequenceArrowStyles(t *testing.T) {
	got := renderSeq(t, "sequenceDiagram\nA->>B: filled\nA-->B: open\nA-x B: cross\nA-)B: async")
	joined := strings.Join(got, "\n")
	for _, want := range []string{"▶", "▷", "✗"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected head %q, got:\n%s", want, joined)
		}
	}
	if !strings.Contains(joined, "┄") {
		t.Errorf("expected a dashed arrow, got:\n%s", joined)
	}
}

func TestSequenceNotes(t *testing.T) {
	src := "sequenceDiagram\nA->>B: hi\nNote over A: single\nNote over A,B: span\nNote left of A: l\nNote right of B: r"
	got := renderSeq(t, src)
	joined := strings.Join(got, "\n")
	for _, want := range []string{"single", "span", "│l│", "│r│"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected note %q, got:\n%s", want, joined)
		}
	}
}

func TestSequenceFragments(t *testing.T) {
	src := "sequenceDiagram\nA->>B: start\nloop until ok\nA->>B: ping\nB-->>A: pong\nend"
	got := renderSeq(t, src)
	joined := strings.Join(got, "\n")
	for _, want := range []string{"┌─loop: until ok─┐", "└────────────────┘"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q, got:\n%s", want, joined)
		}
	}
}

func TestSequenceAltElse(t *testing.T) {
	src := "sequenceDiagram\nalt yes\nA->>B: 1\nelse no\nA->>B: 2\nend"
	got := renderSeq(t, src)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "┌─alt: yes") {
		t.Errorf("expected alt frame, got:\n%s", joined)
	}
	if !strings.Contains(joined, "│─else: no") {
		t.Errorf("expected else divider, got:\n%s", joined)
	}
}

func TestSequenceParseErrors(t *testing.T) {
	cases := []struct {
		src   string
		match string
	}{
		{"sequenceDiagram\nA B: hi", "expected a message"},
		{"sequenceDiagram\nA->>B hi", "message needs text"},
		{"sequenceDiagram\nNote over A hi", "note needs text"},
		{"sequenceDiagram\nloop x\nA->>B: hi", "unterminated fragment"},
		{"sequenceDiagram\nend", "unmatched end"},
		{"sequenceDiagram\nA->>B: hi\nelse later", "outside a fragment"},
	}
	for _, c := range cases {
		_, err := render(c.src, 0, unicodeGlyphs)
		if err == nil || !strings.Contains(err.Error(), c.match) {
			t.Errorf("render(%q) error = %v, want containing %q", c.src, err, c.match)
		}
	}
}

func TestSequenceTooWide(t *testing.T) {
	var b strings.Builder
	b.WriteString("sequenceDiagram\n")
	for i := 0; i < 12; i++ {
		fmt.Fprintf(&b, "P%d->>P%d: msg\n", i, (i+1)%12)
	}
	_, err := render(b.String(), 30, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "too wide") {
		t.Errorf("expected too-wide error, got %v", err)
	}
}

func TestSequenceDeterministic(t *testing.T) {
	src := "sequenceDiagram\nA->>B: hi\nloop x\nB-->>A: ok\nend"
	first := renderSeq(t, src)
	for i := 0; i < 3; i++ {
		again := renderSeq(t, src)
		if strings.Join(first, "\n") != strings.Join(again, "\n") {
			t.Fatalf("sequence render is not deterministic, run %d", i)
		}
	}
}

func TestGanttBasic(t *testing.T) {
	src := "gantt\ntitle Plan\nsection Prep\nDesign :done, 2024-01-01, 5d\nBuild :crit, 2024-01-06, 10d\nLaunch :milestone, 2024-01-20, 0d"
	lines, err := render(src, 60, unicodeGlyphs)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Plan", "Prep", "Design", "Build", "Launch", "░", "▒", "◆", "Jan"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q in gantt output, got:\n%s", want, joined)
		}
	}
	for _, line := range lines {
		if w := stringWidth(line); w > 60 {
			t.Errorf("line too wide (%d): %q", w, line)
		}
	}
}

func TestGanttAfterDependencies(t *testing.T) {
	src := "gantt\nA :a, 2024-01-01, 1w\nB :after a, 3d\nC :after b, 2w"
	lines, err := render(src, 60, unicodeGlyphs)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"A", "B", "C", "█"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q, got:\n%s", want, joined)
		}
	}
}

func TestGanttDurationParsing(t *testing.T) {
	cases := []struct {
		s    string
		want time.Duration
	}{
		{"5d", 5 * dayDur},
		{"2w", 14 * dayDur},
		{"36h", 36 * time.Hour},
		{"0d", 0},
	}
	for _, c := range cases {
		got, ok := parseGanttDuration(c.s)
		if !ok || got != c.want {
			t.Errorf("parseGanttDuration(%q) = %v, want %v", c.s, got, c.want)
		}
	}
	for _, bad := range []string{"d", "5x", "week", "5"} {
		if _, ok := parseGanttDuration(bad); ok {
			t.Errorf("parseGanttDuration(%q) unexpectedly parsed", bad)
		}
	}
}

func TestGanttParseErrors(t *testing.T) {
	cases := []struct {
		src   string
		match string
	}{
		{"gantt\nDesign 2024-01-01, 5d", "task needs a colon"},
		{"gantt\nA :after missing, 1d", "unknown task"},
		{"gantt", "empty diagram"},
		{"gantt\nA :x, y, z", "unknown field"},
		{"gantt\nB :after a, 1d\nA :nope, z, w", "unknown field"},
	}
	for _, c := range cases {
		_, err := render(c.src, 0, unicodeGlyphs)
		if err == nil || !strings.Contains(err.Error(), c.match) {
			t.Errorf("render(%q) error = %v, want containing %q", c.src, err, c.match)
		}
	}
}

func TestGanttTooNarrow(t *testing.T) {
	src := "gantt\nLong task name :done, 2024-01-01, 5d"
	_, err := render(src, 10, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "too wide") {
		t.Errorf("expected too-wide error, got %v", err)
	}
}

func TestGanttDeterministic(t *testing.T) {
	src := "gantt\ntitle P\nA :done, 2024-01-01, 5d\nB :crit, after a, 3d"
	first, err := render(src, 60, unicodeGlyphs)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		again, err := render(src, 60, unicodeGlyphs)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Join(first, "\n") != strings.Join(again, "\n") {
			t.Fatalf("gantt render is not deterministic, run %d", i)
		}
	}
}
