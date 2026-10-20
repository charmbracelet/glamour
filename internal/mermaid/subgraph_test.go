package mermaid

import (
	"strings"
	"testing"
)

func TestRenderSubgraphTD(t *testing.T) {
	src := "flowchart TD\nsubgraph CI\n  A[Build] --> B[Test]\nend\nB --> C[Deploy]"
	got := strings.Join(renderOrFatal(t, src, 0), "\n")
	for _, want := range []string{"┄ CI ┄", "Build", "Test", "Deploy"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
	lines := strings.Split(got, "\n")
	if !strings.HasPrefix(lines[0], "┌┄") && !strings.HasPrefix(lines[0], "┌") {
		t.Errorf("subgraph frame should start the drawing:\n%s", got)
	}
}

func TestRenderSubgraphTitleForms(t *testing.T) {
	for _, head := range []string{`subgraph one [My Group]`, `subgraph "My Group"`, `subgraph My Group`} {
		got := strings.Join(renderOrFatal(t, "flowchart LR\n"+head+"\nA --> B\nend", 0), "\n")
		if !strings.Contains(got, "My Group") {
			t.Errorf("%q: title missing:\n%s", head, got)
		}
	}
}

func TestRenderSubgraphOverlapFallsBack(t *testing.T) {
	// X shares a layer with A, between the members A and B.
	src := "flowchart TD\nS --> A\nS --> X\nA --> B\nX --> B\nsubgraph g\nA\nB\nend"
	_, err := render(src, 0, unicodeGlyphs)
	if err == nil || !strings.Contains(err.Error(), "overlaps") {
		t.Fatalf("render error = %v, want an overlap error", err)
	}
}

func TestRenderSubgraphLabelsStayOffBorder(t *testing.T) {
	src := "flowchart TD\nsubgraph g\nA --> B\nend\nA -->|inside the group| B\nB -->|leaving| C"
	got := renderOrFatal(t, src, 0)
	text := strings.Join(got, "\n")
	for _, want := range []string{"inside the group", "leaving"} {
		if !strings.Contains(text, want) {
			t.Fatalf("label %q missing:\n%s", want, text)
		}
	}
	for _, line := range got {
		if strings.Contains(line, "leaving") && strings.ContainsAny(line, "┄┆") && strings.Contains(line, "┄ leaving") {
			t.Errorf("label sits on the border:\n%s", text)
		}
	}
}
