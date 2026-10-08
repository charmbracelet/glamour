package mermaid

import (
	"strings"
	"testing"
)

func TestParseDirection(t *testing.T) {
	cases := []struct {
		src  string
		want Direction
	}{
		{"flowchart TD\nA --> B", DirTD},
		{"flowchart TB\nA --> B", DirTD},
		{"flowchart\nA --> B", DirTD},
		{"graph\nA --> B", DirTD},
		{"flowchart LR\nA --> B", DirLR},
		{"graph LR\nA --> B", DirLR},
	}
	for _, c := range cases {
		d, err := parse(c.src)
		if err != nil {
			t.Errorf("parse(%q): %v", c.src, err)
			continue
		}
		if d.direction != c.want {
			t.Errorf("parse(%q) direction = %v, want %v", c.src, d.direction, c.want)
		}
	}
}

func TestParseDirectionErrors(t *testing.T) {
	cases := []struct {
		src   string
		match string
	}{
		{"flowchart BT\nA --> B", "unsupported direction"},
		{"flowchart RL\nA --> B", "unsupported direction"},
		{"sequenceDiagram\nA->>B: hi", "unsupported diagram type"},
		{"stateDiagram-v2\nA --> B", "unsupported diagram type"},
		{"classDiagram\nclass A", "unsupported diagram type"},
		{"gantt\napple :a", "unsupported diagram type"},
		{"something weird", `expected "flowchart"`},
		{"flowchart TD\nsubgraph box\nend", "unsupported syntax"},
		{"flowchart TD\nclassDef bold fill:#f00", "unsupported syntax"},
		{"flowchart TD\nstyle A fill:#f00", "unsupported syntax"},
		{"flowchart TD\nA o-- B", "unsupported syntax"},
		{"flowchart TD\nA --> B & C", "unsupported syntax"},
		{"flowchart TD\nA <-- B", "unsupported syntax"},
		{"flowchart TD\n%% comment only", "empty"},
	}
	for _, c := range cases {
		_, err := parse(c.src)
		if err == nil || !strings.Contains(err.Error(), c.match) {
			t.Errorf("parse(%q) error = %v, want containing %q", c.src, err, c.match)
		}
	}
}

func TestParseNodes(t *testing.T) {
	d, err := parse("flowchart TD\nA[Build] --> B(Test) --> C((Ship)) --> D{Gate} --> E{{Hex}} --> F([Stadium])")
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		id    string
		label string
		shape Shape
	}{
		{"A", "Build", ShapeRect},
		{"B", "Test", ShapeRounded},
		{"C", "Ship", ShapeCircle},
		{"D", "Gate", ShapeDiamond},
		{"E", "Hex", ShapeHexagon},
		{"F", "Stadium", ShapeStadium},
	}
	for _, w := range want {
		n, ok := d.nodeByID[w.id]
		if !ok {
			t.Fatalf("missing node %q", w.id)
		}
		if n.label != w.label || n.shape != w.shape {
			t.Errorf("node %q = (%q, %v), want (%q, %v)", w.id, n.label, n.shape, w.label, w.shape)
		}
	}
}

func TestParsePlainNodesDefaultToID(t *testing.T) {
	d, err := parse("flowchart TD\nA --> B[Label]")
	if err != nil {
		t.Fatal(err)
	}
	if n := d.nodeByID["A"]; n.label != "A" || n.shape != ShapeRect {
		t.Errorf("plain node = (%q, %v)", n.label, n.shape)
	}
}

func TestParseMergesLaterShape(t *testing.T) {
	d, err := parse("flowchart TD\nA --> B\nB[Label]")
	if err != nil {
		t.Fatal(err)
	}
	if n := d.nodeByID["B"]; n.label != "Label" {
		t.Errorf("merged node label = %q, want Label", n.label)
	}
}

func TestParseQuotedLabels(t *testing.T) {
	d, err := parse(`flowchart TD
A["label with ] bracket"] --> B --> C["quoted \" esc"]`)
	if err != nil {
		t.Fatal(err)
	}
	if n := d.nodeByID["A"]; n.label != `label with ] bracket` {
		t.Errorf("label = %q", n.label)
	}
	if n := d.nodeByID["C"]; n.label != `quoted " esc` {
		t.Errorf("label = %q", n.label)
	}
}

func TestParseLineBreaks(t *testing.T) {
	d, err := parse("flowchart TD\nA[two<br/>lines]")
	if err != nil {
		t.Fatal(err)
	}
	if n := d.nodeByID["A"]; n.label != "two\nlines" {
		t.Errorf("label = %q, want two\\nlines", n.label)
	}
}

func TestParseEdgeKinds(t *testing.T) {
	d, err := parse("flowchart TD\nA --> B\nB --- C\nC -.-> D\nD -.- E\nE ==> F\nF === G\nG ---> H")
	if err != nil {
		t.Fatal(err)
	}
	want := []edgeKind{edgeArrow, edgeOpen, edgeDottedArrow, edgeDottedOpen, edgeThickArrow, edgeThickOpen, edgeArrow}
	if len(d.edges) != len(want) {
		t.Fatalf("got %d edges, want %d", len(d.edges), len(want))
	}
	for i, k := range want {
		if d.edges[i].kind != k {
			t.Errorf("edge %d kind = %v, want %v", i, d.edges[i].kind, k)
		}
	}
}

func TestParseEdgeLabels(t *testing.T) {
	cases := []struct {
		src   string
		label string
	}{
		{"A -->|yes| B", "yes"},
		{"A -- yes --> B", "yes"},
		{"A -. maybe .-> B", "maybe"},
		{"A == fast ==> B", "fast"},
		{"A ---|plain| B", "plain"},
		{`A -->|"quoted|"| B`, "quoted|"},
	}
	for _, c := range cases {
		d, err := parse("flowchart LR\n" + c.src)
		if err != nil {
			t.Errorf("parse(%q): %v", c.src, err)
			continue
		}
		if len(d.edges) != 1 || d.edges[0].label != c.label {
			t.Errorf("parse(%q) label = %q, want %q", c.src, d.edges[0].label, c.label)
		}
	}
}

func TestParseChains(t *testing.T) {
	d, err := parse("flowchart TD\nA --> B --> C")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.edges) != 2 {
		t.Fatalf("got %d edges, want 2", len(d.edges))
	}
	if d.edges[0].to.id != "B" || d.edges[1].to.id != "C" {
		t.Error("chain edges are not in order")
	}
}

func TestParseSemicolonsAndComments(t *testing.T) {
	d, err := parse("flowchart TD\n%% a comment\nA --> B; B --> C")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.edges) != 2 {
		t.Fatalf("got %d edges, want 2", len(d.edges))
	}
}

func TestParseFrontmatter(t *testing.T) {
	d, err := parse("---\ntitle: Build flow\n---\nflowchart LR\nA --> B")
	if err != nil {
		t.Fatal(err)
	}
	if d.title != "Build flow" {
		t.Errorf("title = %q", d.title)
	}
}

func TestParseHyphenatedIDs(t *testing.T) {
	d, err := parse("flowchart TD\nbuild-1 --> test-2.something")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := d.nodeByID["build-1"]; !ok {
		t.Error("missing hyphenated id build-1")
	}
	if _, ok := d.nodeByID["test-2.something"]; !ok {
		t.Error("missing dotted id test-2.something")
	}
}

func TestParseTrailingGarbage(t *testing.T) {
	if _, err := parse("flowchart TD\nA --> B trailing"); err == nil {
		t.Error("expected an error for trailing garbage")
	}
}

func TestParseIsDeterministic(t *testing.T) {
	src := "flowchart TD\nA --> B\nA --> C\nB --> D\nC --> D"
	first, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		again, err := parse(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(again.nodes) != len(first.nodes) || len(again.edges) != len(first.edges) {
			t.Fatal("node or edge count differs between runs")
		}
		for j := range again.nodes {
			if again.nodes[j].id != first.nodes[j].id {
				t.Fatalf("node order differs at %d", j)
			}
		}
	}
}
