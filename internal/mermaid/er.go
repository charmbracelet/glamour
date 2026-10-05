package mermaid

import "strings"

// erCards maps ER cardinality markers to readable text.
var erCards = map[string]string{
	"||": "1", "o|": "0..1", "|o": "0..1",
	"o{": "0..*", "}o": "0..*",
	"|{": "1..*", "}|": "1..*",
}

// parseER converts an erDiagram definition into a flowchart: entities
// become boxes carrying their attributes, relationships become open lines
// labelled with the cardinality and name.
func parseER(src string) (*diagram, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)
	d := &diagram{direction: DirLR, title: title, nodeByID: make(map[string]*node)}

	seenHeader, inEntity := false, ""
	lineno := 0
	for _, line := range strings.Split(src, "\n") {
		lineno++
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		word := strings.ToLower(firstWord(trimmed))

		if inEntity != "" {
			if trimmed == "}" {
				inEntity = ""
				continue
			}
			if err := addERAttribute(d, inEntity, trimmed); err != nil {
				return nil, errf("%v (line %d)", err, lineno)
			}
			continue
		}

		if !seenHeader {
			if word == "erdiagram" {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "erDiagram"`, lineno)
		}

		if entity, ok := strings.CutSuffix(trimmed, "{"); ok {
			inEntity = strings.TrimSpace(entity)
			d.addNode(&node{id: inEntity, label: inEntity})
			continue
		}
		if err := parseERRelationship(d, trimmed, lineno); err != nil {
			return nil, err
		}
	}

	if inEntity != "" {
		return nil, errf("could not parse: unterminated entity block")
	}
	if !seenHeader || len(d.nodes) == 0 {
		return nil, errf("empty diagram")
	}
	return d, nil
}

// addERAttribute appends one attribute line to an entity's label.
func addERAttribute(d *diagram, entity, line string) error {
	n, ok := d.nodeByID[entity]
	if !ok {
		return errf("could not parse: unknown entity %q", entity)
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return errf("could not parse: %q", line)
	}
	attr := fields[0] + " " + fields[1]
	if len(fields) > 2 {
		switch strings.ToUpper(fields[2]) {
		case "PK", "FK", "UK":
			attr += " " + strings.ToUpper(fields[2])
		}
	}
	n.label += "\n" + attr
	return nil
}

// parseERRelationship parses `A ||--o{ B : label`.
func parseERRelationship(d *diagram, line string, lineno int) error {
	head, label, _ := strings.Cut(line, ":")
	label = strings.TrimSpace(label)
	fields := strings.Fields(head)
	if len(fields) < 3 {
		return errf("could not parse (line %d): %q", lineno, line)
	}

	left, marker, right := fields[0], fields[1], fields[2]
	card, known := erCardinality(marker)
	if !known {
		return errf("could not parse (line %d): unknown cardinality %q", lineno, marker)
	}
	d.addNode(&node{id: left, label: left})
	d.addNode(&node{id: right, label: right})

	if label == "" {
		label = card
	} else {
		label = label + " " + card
	}
	d.edges = append(d.edges, &edge{
		from:  d.nodeByID[left],
		to:    d.nodeByID[right],
		label: label,
		kind:  edgeOpen,
	})
	return nil
}

// erCardinality turns a relationship marker like ||--o{ into readable
// text.
func erCardinality(marker string) (string, bool) {
	left := marker[:2]
	right := marker[len(marker)-2:]
	rest := marker[2 : len(marker)-2]
	if rest != "--" && rest != ".." {
		return "", false
	}
	l, lok := erCards[left]
	r, rok := erCards[right]
	if !lok || !rok {
		return "", false
	}
	return l + ":" + r, true
}
