package mermaid

import "strings"

// parseClass converts a classDiagram definition into a flowchart: classes
// become boxes listing their members, relationships become lines whose
// style follows the operator (dotted operators draw dotted lines, arrowed
// operators draw arrowheads).
func parseClass(src string) (*diagram, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)
	d := &diagram{direction: DirTD, title: title, nodeByID: make(map[string]*node)}

	seenHeader, inClass := false, ""
	lineno := 0
	for _, line := range strings.Split(src, "\n") {
		lineno++
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		word := strings.ToLower(firstWord(trimmed))

		if inClass != "" {
			if trimmed == "}" {
				inClass = ""
				continue
			}
			if err := addClassMember(d, inClass, trimmed); err != nil {
				return nil, errf("%v (line %d)", err, lineno)
			}
			continue
		}

		if !seenHeader {
			if word == "classdiagram" {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "classDiagram"`, lineno)
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
		case "class":
			rest := strings.TrimSpace(trimmed[len("class"):])
			if id, label, ok := strings.Cut(rest, "["); ok {
				if end := strings.Index(label, "]"); end >= 0 {
					d.addNode(&node{id: strings.TrimSpace(id), label: label[:end]})
					continue
				}
			}
			if body, ok := strings.CutSuffix(rest, "{"); ok {
				inClass = strings.TrimSpace(body)
				d.addNode(&node{id: inClass, label: inClass})
				continue
			}
			id := strings.TrimSpace(rest)
			if id == "" {
				return nil, errf("could not parse (line %d): empty class name", lineno)
			}
			d.addNode(&node{id: id, label: id})
		case "note", "namespace", "click":
			return nil, errf("unsupported syntax (line %d): %q", lineno, word)
		default:
			if err := parseClassRelationship(d, trimmed, lineno); err != nil {
				return nil, err
			}
		}
	}

	if inClass != "" {
		return nil, errf("could not parse: unterminated class block")
	}
	if !seenHeader || len(d.nodes) == 0 {
		return nil, errf("empty diagram")
	}
	return d, nil
}

// addClassMember appends one member line to a class's label.
func addClassMember(d *diagram, class, line string) error {
	n, ok := d.nodeByID[class]
	if !ok {
		return errf("could not parse: unknown class %q", class)
	}
	n.label += "\n" + line
	return nil
}

// parseClassRelationship parses `A <|-- B : label`, including optional
// quoted multiplicities around the operator.
func parseClassRelationship(d *diagram, line string, lineno int) error {
	head, label, _ := strings.Cut(line, ":")
	label = strings.TrimSpace(label)

	ids, op, ok := splitClassRelationship(head)
	if !ok {
		return errf("could not parse (line %d): %q", lineno, line)
	}
	left, right := ids[0], ids[1]
	if left == "" || right == "" {
		return errf("could not parse (line %d): missing class", lineno)
	}

	// A leading < points from the right-hand class to the left-hand one.
	from, to := left, right
	if strings.HasPrefix(op, "<") {
		from, to = right, left
	}
	dotted := strings.Contains(op, ".")
	arrow := strings.ContainsAny(op, "<>")
	kind := edgeOpen
	switch {
	case arrow && dotted:
		kind = edgeDottedArrow
	case arrow:
		kind = edgeArrow
	case dotted:
		kind = edgeDottedOpen
	}

	d.addNode(&node{id: left, label: left})
	d.addNode(&node{id: right, label: right})
	d.edges = append(d.edges, &edge{
		from:  d.nodeByID[from],
		to:    d.nodeByID[to],
		label: label,
		kind:  kind,
	})
	return nil
}

// splitClassRelationship splits a relationship head into its two class ids
// and the operator between them, dropping quoted multiplicities.
func splitClassRelationship(head string) ([]string, string, bool) {
	var fields []string
	for _, field := range strings.Fields(head) {
		if len(field) >= 2 && field[0] == '"' && field[len(field)-1] == '"' {
			continue
		}
		fields = append(fields, field)
	}
	if len(fields) != 3 {
		return nil, "", false
	}
	for _, r := range fields[1] {
		switch r {
		case '-', '.', '|', '<', '>', '*', 'o':
		default:
			return nil, "", false
		}
	}
	return []string{fields[0], fields[2]}, fields[1], true
}
