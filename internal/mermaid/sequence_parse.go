package mermaid

import "strings"

// seqKind tags a sequence diagram event.
type seqKind int

// Sequence event kinds.
const (
	seqMessage seqKind = iota
	seqSelf
	seqNote
	seqFragment
	seqDivider
)

// seqHead describes the arrowhead of a message.
type seqHead int

// Sequence arrowhead styles.
const (
	seqHeadFilled seqHead = iota
	seqHeadOpen
	seqHeadCross
)

// seqOps lists the message operators, longest first.
var seqOps = []string{"-->>", "-->", "--x", "--)", "->>", "->", "-x", "-)"}

// seqEvent is one step of a sequence diagram. Fragments carry their children
// in events; dividers separate fragment branches.
type seqEvent struct {
	kind   seqKind
	from   int
	to     int
	label  string
	frag   string
	dotted bool
	head   seqHead
	side   int
	events []seqEvent
}

// seqParticipant is one lifeline.
type seqParticipant struct {
	id    string
	label string
}

type seqDiagram struct {
	title        string
	participants []seqParticipant
	byID         map[string]int
	events       []seqEvent
}

// parseSequence parses a sequenceDiagram definition.
func parseSequence(src string) (*seqDiagram, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)
	d := &seqDiagram{title: title, byID: make(map[string]int)}

	var stack []*[]seqEvent
	current := &d.events
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
			if word == "sequencediagram" {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "sequenceDiagram"`, lineno)
		}

		switch word {
		case "participant", "actor":
			if err := d.parseParticipant(trimmed); err != nil {
				return nil, err
			}
		case "autonumber", "activate", "deactivate":
			// Procedural directives without layout impact; ignored.
		case "note":
			ev, err := d.parseNote(trimmed, lineno)
			if err != nil {
				return nil, err
			}
			*current = append(*current, *ev)
		case "loop", "opt", "alt", "par", "critical", "block":
			*current = append(*current, seqEvent{
				kind:  seqFragment,
				frag:  word,
				label: strings.TrimSpace(trimmed[len(word):]),
			})
			stack = append(stack, current)
			current = &(*current)[len(*current)-1].events
		case "else", "and", "option":
			if len(stack) == 0 {
				return nil, errf("could not parse (line %d): %q outside a fragment", lineno, word)
			}
			*current = append(*current, seqEvent{
				kind:  seqDivider,
				frag:  word,
				label: strings.TrimSpace(trimmed[len(word):]),
			})
		case "end":
			if len(stack) == 0 {
				return nil, errf("could not parse (line %d): unmatched end", lineno)
			}
			current = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		default:
			ev, err := d.parseMessage(trimmed, lineno)
			if err != nil {
				return nil, err
			}
			*current = append(*current, *ev)
		}
	}

	if len(stack) > 0 {
		return nil, errf("could not parse: unterminated fragment")
	}
	if !seenHeader || len(d.participants) == 0 {
		return nil, errf("empty diagram")
	}
	return d, nil
}

// participant returns the index of a participant, declaring it on first use.
func (d *seqDiagram) participant(id, label string) int {
	if i, ok := d.byID[id]; ok {
		return i
	}
	if label == "" {
		label = id
	}
	d.byID[id] = len(d.participants)
	d.participants = append(d.participants, seqParticipant{id: id, label: label})
	return len(d.participants) - 1
}

func (d *seqDiagram) parseParticipant(s string) error {
	rest := strings.TrimSpace(s[len(firstWord(s)):])
	id, label := rest, rest
	if parts := strings.SplitN(rest, " as ", 2); len(parts) == 2 {
		id = strings.TrimSpace(parts[0])
		label = strings.TrimSpace(parts[1])
	}
	if id == "" {
		return errf("could not parse: empty participant id")
	}
	d.participant(id, label)
	return nil
}

func (d *seqDiagram) parseNote(s string, lineno int) (*seqEvent, error) {
	rest := strings.TrimSpace(s[len(firstWord(s)):])
	ev := &seqEvent{kind: seqNote}
	switch strings.ToLower(firstWord(rest)) {
	case "over":
		rest = strings.TrimSpace(rest[len(firstWord(rest)):])
	case "left", "right":
		if strings.ToLower(firstWord(rest)) == "left" {
			ev.side = -1
		} else {
			ev.side = 1
		}
		rest = strings.TrimSpace(rest[len(firstWord(rest)):])
		if strings.ToLower(firstWord(rest)) == "of" {
			rest = strings.TrimSpace(rest[len("of"):])
		}
	default:
		return nil, errf("could not parse (line %d): note needs over, left of or right of", lineno)
	}

	colon := strings.IndexByte(rest, ':')
	if colon < 0 {
		return nil, errf("could not parse (line %d): note needs text after a colon", lineno)
	}
	ids := strings.TrimSpace(rest[:colon])
	ev.label = strings.TrimSpace(rest[colon+1:])
	if ev.side == 0 {
		if parts := strings.Split(ids, ","); len(parts) == 2 {
			ev.from = d.participant(strings.TrimSpace(parts[0]), "")
			ev.to = d.participant(strings.TrimSpace(parts[1]), "")
			if ev.from > ev.to {
				ev.from, ev.to = ev.to, ev.from
			}
			return ev, nil
		}
	}
	ev.from = d.participant(firstWord(ids), "")
	ev.to = ev.from
	return ev, nil
}

func (d *seqDiagram) parseMessage(s string, lineno int) (*seqEvent, error) {
	op, start, length := findSeqOp(s)
	if start < 0 {
		return nil, errf("could not parse (line %d): expected a message", lineno)
	}
	fromID := strings.TrimSpace(s[:start])
	rest := s[start+length:]
	colon := strings.IndexByte(rest, ':')
	if colon < 0 {
		return nil, errf("could parse (line %d): message needs text after a colon", lineno)
	}
	toID := strings.TrimSpace(rest[:colon])
	if fromID == "" || toID == "" {
		return nil, errf("could not parse (line %d): missing participant", lineno)
	}

	ev := &seqEvent{
		kind:   seqMessage,
		from:   d.participant(fromID, ""),
		to:     d.participant(toID, ""),
		label:  strings.TrimSpace(rest[colon+1:]),
		dotted: strings.HasPrefix(op, "--"),
	}
	switch {
	case strings.HasSuffix(op, "x"):
		ev.head = seqHeadCross
	case strings.HasSuffix(op, ")"):
		ev.head = seqHeadOpen
	case strings.HasSuffix(op, ">>"):
		ev.head = seqHeadFilled
	default:
		ev.head = seqHeadOpen
	}
	if ev.from == ev.to {
		ev.kind = seqSelf
	}
	return ev, nil
}

// findSeqOp finds the first message operator in s.
func findSeqOp(s string) (op string, start, length int) {
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			continue
		}
		for _, candidate := range seqOps {
			if strings.HasPrefix(s[i:], candidate) {
				return candidate, i, len(candidate)
			}
		}
	}
	return "", -1, 0
}
