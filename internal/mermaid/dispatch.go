package mermaid

import "strings"

// dKind tags the diagram type declared by a mermaid source.
type dKind int

// Diagram kinds.
const (
	kindOther dKind = iota
	kindFlowchart
	kindSequence
	kindGantt
	kindState
	kindER
	kindClass
	kindPie
	kindJourney
	kindTimeline
	kindMindmap
	kindQuadrant
	kindXYChart
	kindGitGraph
)

// diagramKind sniffs the type from the first significant line.
func diagramKind(src string) dKind {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		switch strings.ToLower(firstWord(trimmed)) {
		case "flowchart", "graph":
			return kindFlowchart
		case "sequencediagram":
			return kindSequence
		case kwGantt:
			return kindGantt
		case "statediagram", "statediagram-v2":
			return kindState
		case "erdiagram":
			return kindER
		case "classdiagram":
			return kindClass
		case "pie":
			return kindPie
		case kwJourney:
			return kindJourney
		case kwTimeline:
			return kindTimeline
		case "mindmap":
			return kindMindmap
		case "quadrantchart":
			return kindQuadrant
		case "xychart-beta", "xychart":
			return kindXYChart
		case "gitgraph":
			return kindGitGraph
		}
		return kindOther
	}
	return kindOther
}
