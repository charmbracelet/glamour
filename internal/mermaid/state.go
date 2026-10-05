package mermaid

import "strings"

// parseState converts a stateDiagram (or stateDiagram-v2) definition into
// a flowchart: states become rounded nodes, the start and end markers
// become circles, and transitions become labelled edges. The layout
// engine, including cyclic transition graphs, does the rest.
func parseState(src string) (*diagram, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)
	d := &diagram{direction: DirTD, title: title, nodeByID: make(map[string]*node)}

	seenHeader := false
	lineno := 0
	for _, line := range strings.Split(src, "\n") {
		lineno++
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		word := strings.ToLower(firstWord(trimmed))

		if !seenHeader {
			if word == "statediagram" || word == "statediagram-v2" {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "stateDiagram"`, lineno)
		}

		switch word {
		case "direction":
			dir := strings.ToUpper(strings.TrimSpace(trimmed[len("direction"):]))
			switch dir {
			case "LR":
				d.direction = DirLR
			case "RL", "BT":
				return nil, errf("unsupported direction %q", dir)
			}
		case "state":
			rest := strings.TrimSpace(trimmed[len("state"):])
			if strings.HasSuffix(rest, "{") || strings.Contains(rest, "{") {
				return nil, errf("unsupported syntax (line %d): composite states", lineno)
			}
			if label, id, ok := splitStateAlias(rest); ok {
				d.addNode(&node{id: id, label: label, shape: ShapeRounded})
				continue
			}
			return nil, errf("could not parse (line %d): %q", lineno, trimmed)
		case "note", "fork", "join":
			return nil, errf("unsupported syntax (line %d): %q", lineno, word)
		default:
			if err := parseStateTransition(d, trimmed, lineno); err != nil {
				return nil, err
			}
		}
	}

	if !seenHeader || len(d.nodes) == 0 {
		return nil, errf("empty diagram")
	}
	return d, nil
}

// splitStateAlias parses the `state "Long label" as X` form.
func splitStateAlias(rest string) (label, id string, ok bool) {
	if !strings.HasPrefix(rest, `"`) {
		return "", "", false
	}
	end := strings.Index(rest[1:], `"`)
	if end < 0 {
		return "", "", false
	}
	label = rest[1 : 1+end]
	tail := strings.TrimSpace(rest[2+end:])
	parts := strings.Fields(tail)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "as" {
		return "", "", false
	}
	return label, parts[1], true
}

// parseStateTransition parses `A --> B : label`, mapping the start and end
// markers to circle nodes.
func parseStateTransition(d *diagram, line string, lineno int) error {
	left, right, ok := strings.Cut(line, "-->")
	if !ok {
		return errf("could not parse (line %d): %q", lineno, line)
	}
	fromID := strings.TrimSpace(left)
	rest := strings.TrimSpace(right)
	label := ""
	if toPart, lbl, hasColon := strings.Cut(rest, ":"); hasColon {
		rest = toPart
		label = strings.TrimSpace(lbl)
	}
	toID := firstWord(rest)

	from, err := stateNode(d, fromID, true)
	if err != nil {
		return errf("%v (line %d)", err, lineno)
	}
	to, err := stateNode(d, toID, false)
	if err != nil {
		return errf("%v (line %d)", err, lineno)
	}
	d.edges = append(d.edges, &edge{from: from, to: to, label: label, kind: edgeArrow})
	return nil
}

// stateNode resolves a state reference, creating it on first use. The
// start and end markers render as small circles.
func stateNode(d *diagram, id string, isFrom bool) (*node, error) {
	if id == "[*]" {
		nodeID := stateEndID
		label := "◉"
		if isFrom {
			nodeID, label = stateStartID, "●"
		}
		d.addNode(&node{id: nodeID, label: label, shape: ShapeCircle})
		return d.nodeByID[nodeID], nil
	}
	if id == "" {
		return nil, errf("could not parse: missing state")
	}
	d.addNode(&node{id: id, label: id, shape: ShapeRounded})
	return d.nodeByID[id], nil
}

// Internal ids for the start and end markers.
const (
	stateStartID = "(* start *)"
	stateEndID   = "(* end *)"
)
