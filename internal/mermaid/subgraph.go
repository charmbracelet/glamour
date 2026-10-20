package mermaid

import "strings"

// Keywords delimiting a flowchart subgraph.
const (
	kwSubgraph = "subgraph"
	kwEnd      = "end"
)

// subgraph groups flowchart nodes inside a labelled box. Nesting is not
// supported.
type subgraph struct {
	id      string
	title   string
	members []string
}

// openSubgraph handles `subgraph id [title]`, `subgraph "title"` and
// `subgraph title`.
func (d *diagram) openSubgraph(stmt string, lineno int) error {
	if d.open != nil {
		return errf("unsupported syntax %q", "nested subgraph")
	}
	rest := strings.TrimSpace(stmt[len(kwSubgraph):])
	if rest == "" {
		return errf("could not parse (line %d): subgraph needs a name", lineno)
	}
	id, title := rest, rest
	if i := strings.Index(rest, "["); i > 0 && strings.HasSuffix(rest, "]") {
		id = strings.TrimSpace(rest[:i])
		title = rest[i+1 : len(rest)-1]
	}
	id = strings.Trim(id, `"`)
	title = strings.Trim(strings.TrimSpace(title), `"`)
	sg := &subgraph{id: id, title: title}
	d.subgraphs = append(d.subgraphs, sg)
	d.open = sg
	return nil
}

func (d *diagram) closeSubgraph(lineno int) error {
	if d.open == nil {
		return errf("could not parse (line %d): \"end\" without subgraph", lineno)
	}
	d.open = nil
	return nil
}

// assignToOpenSubgraph makes the node a member of the open subgraph unless
// another subgraph already holds it.
func (d *diagram) assignToOpenSubgraph(id string) {
	if d.open == nil {
		return
	}
	for _, sg := range d.subgraphs {
		for _, m := range sg.members {
			if m == id {
				return
			}
		}
	}
	d.open.members = append(d.open.members, id)
}

// Padding between a subgraph box and the nodes it holds.
const sgPad = 1

// sgBox is the canvas rectangle of a subgraph; right and bottom are
// inclusive border coordinates.
type sgBox struct {
	sg                       *subgraph
	left, top, right, bottom int
}

// subgraphBoxes returns the box around each subgraph's members, or an
// error when a box would swallow a non-member node or overlap another box.
func (l *layoutCtx) subgraphBoxes(d *diagram) ([]sgBox, error) {
	var boxes []sgBox
	for _, sg := range d.subgraphs {
		ids := make(map[string]bool, len(sg.members))
		for _, m := range sg.members {
			ids[m] = true
		}
		b := sgBox{sg: sg, left: 1 << 30, top: 1 << 30, right: -1, bottom: -1}
		for _, layer := range l.layers {
			for _, n := range layer {
				if n.virtual || !ids[n.node.id] {
					continue
				}
				b.left = min(b.left, n.x-sgPad-1)
				b.top = min(b.top, n.y-sgPad-1)
				b.right = max(b.right, n.x+n.w+sgPad)
				b.bottom = max(b.bottom, n.y+n.h+sgPad)
			}
		}
		if b.right < 0 {
			continue
		}
		b.right += interiorLabelWidth(d, ids)
		boxes = append(boxes, b)
	}
	for _, b := range boxes {
		ids := make(map[string]bool, len(b.sg.members))
		for _, m := range b.sg.members {
			ids[m] = true
		}
		for _, layer := range l.layers {
			for _, n := range layer {
				if n.virtual || ids[n.node.id] {
					continue
				}
				if n.x < b.right && n.x+n.w > b.left && n.y < b.bottom && n.y+n.h > b.top {
					return nil, errf("subgraph %q overlaps other nodes in this layout", b.sg.id)
				}
			}
		}
	}
	for i, a := range boxes {
		for _, b := range boxes[i+1:] {
			if a.left <= b.right && b.left <= a.right && a.top <= b.bottom && b.top <= a.bottom {
				return nil, errf("subgraphs %q and %q overlap in this layout", a.sg.id, b.sg.id)
			}
		}
	}
	return boxes, nil
}

// interiorLabelWidth returns the room a box needs on its right for labels of
// edges whose both ends are members, since labels are stamped beside the line.
func interiorLabelWidth(d *diagram, members map[string]bool) int {
	w := 0
	for _, e := range d.edges {
		if e.label != "" && members[e.from.id] && members[e.to.id] {
			w = max(w, stringWidth(e.label)+2)
		}
	}
	return w
}

// reserveSubgraphs shifts the layout so subgraph boxes stay on the canvas,
// grows the canvas to hold them and returns the boxes in final coordinates.
func (l *layoutCtx) reserveSubgraphs(d *diagram) ([]sgBox, error) {
	if len(d.subgraphs) == 0 {
		return nil, nil
	}
	boxes, err := l.subgraphBoxes(d)
	if err != nil || len(boxes) == 0 {
		return nil, err
	}
	dx, dy := 0, 0
	for _, b := range boxes {
		dx = max(dx, -b.left)
		dy = max(dy, -b.top)
	}
	if dx > 0 || dy > 0 {
		for _, layer := range l.layers {
			for _, n := range layer {
				n.x += dx
				n.y += dy
			}
		}
		for i := range l.bandPos {
			if l.lr {
				l.bandPos[i] += dx
			} else {
				l.bandPos[i] += dy
			}
		}
		l.size.x += dx
		l.size.y += dy
	}
	for i := range boxes {
		boxes[i].left += dx
		boxes[i].right += dx
		boxes[i].top += dy
		boxes[i].bottom += dy
		l.size.x = max(l.size.x, boxes[i].right+1)
		l.size.y = max(l.size.y, boxes[i].bottom+1)
	}
	return boxes, nil
}

// drawSubgraphs draws each box as a dotted frame with its title set into
// the top border. Frames use line marks so edges crossing them join the
// border instead of leaving a gap.
func drawSubgraphs(c *canvas, boxes []sgBox) {
	for _, b := range boxes {

		c.mark(b.left, b.top, dirDown|dirRight, lineDotted)
		c.mark(b.right, b.top, dirDown|dirLeft, lineDotted)
		c.mark(b.left, b.bottom, dirUp|dirRight, lineDotted)
		c.mark(b.right, b.bottom, dirUp|dirLeft, lineDotted)
		for x := b.left + 1; x < b.right; x++ {
			c.mark(x, b.top, dirLeft|dirRight, lineDotted)
			c.mark(x, b.bottom, dirLeft|dirRight, lineDotted)
		}
		for y := b.top + 1; y < b.bottom; y++ {
			c.mark(b.left, y, dirUp|dirDown, lineDotted)
			c.mark(b.right, y, dirUp|dirDown, lineDotted)
		}
		markFrame(c, b)
		title := " " + b.sg.title + " "
		if room := b.right - b.left - 3; stringWidth(title) <= room {
			c.label(point{x: b.left + 2, y: b.top}, title)
		}
	}
}

// markFrame flags the border cells so labels can avoid them.
func markFrame(c *canvas, b sgBox) {
	for x := b.left; x <= b.right; x++ {
		c.cells[b.top][x].frame = true
		c.cells[b.bottom][x].frame = true
	}
	for y := b.top; y <= b.bottom; y++ {
		c.cells[y][b.left].frame = true
		c.cells[y][b.right].frame = true
	}
}
