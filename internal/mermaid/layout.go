package mermaid

import (
	"sort"
	"strings"
)

// settings control the spacing of a layout; the compact variant is the
// second attempt when the diagram must fit a narrower width.
type settings struct {
	hGap, vGap, boxPad int
	maxLabelW          int
}

var defaultSettings = settings{hGap: 4, vGap: 3, boxPad: 1, maxLabelW: 24}
var compactSettings = settings{hGap: 1, vGap: 1, boxPad: 0, maxLabelW: 12}

// maxLRLabelGap caps how wide left-to-right diagrams grow to make room for
// edge labels.
const maxLRLabelGap = 20

// lnode is a node during layout: either a real node or a virtual waypoint
// that reserves a column (top-down) or row (left-to-right) so a long edge
// can pass through an intermediate layer without hitting any box.
type lnode struct {
	node    *node
	virtual bool
	layer   int
	x, y    int
	w, h    int
	lines   []string
}

// ledge is an edge during layout. chain holds the endpoints and, for edges
// skipping layers, the virtual waypoints in between. Edges closing a cycle
// are marked as feedback edges and routed around the drawing instead of
// taking part in the layering.
type ledge struct {
	from, to *lnode
	label    string
	kind     edgeKind
	chain    []*lnode
	feedback bool
}

// layoutCtx carries the laid out diagram through positioning and routing.
type layoutCtx struct {
	st       settings
	lr       bool
	layers   [][]*lnode
	bandPos  []int
	size     point
	overhang int
}

// draw lays out the diagram and renders it onto a canvas.
func draw(d *diagram, st settings, g glyphSet) ([]string, int, error) {
	nodes := make([]*lnode, len(d.nodes))
	byID := make(map[string]*lnode, len(d.nodes))
	for i, n := range d.nodes {
		lines := wrapLabel(n.label, st.maxLabelW)
		w := 0
		for _, l := range lines {
			w = max(w, stringWidth(l))
		}
		ln := &lnode{
			node:  n,
			lines: lines,
			w:     w + 2*st.boxPad + 2,
			h:     len(lines) + 2*st.boxPad + 2,
		}
		nodes[i] = ln
		byID[n.id] = ln
	}

	edges := make([]*ledge, len(d.edges))
	for i, e := range d.edges {
		edges[i] = &ledge{
			from:  byID[e.from.id],
			to:    byID[e.to.id],
			label: e.label,
			kind:  e.kind,
		}
	}

	if err := computeLayers(nodes, edges); err != nil {
		return nil, 0, err
	}
	insertVirtuals(edges)
	layers := buildLayers(nodes, edges)
	preds, succs := chainAdjacency(edges)
	orderLayers(layers, preds, succs)

	l := &layoutCtx{st: st, lr: d.direction == DirLR, layers: layers}
	l.widenForLabels(edges)
	l.position()
	l.overhang = labelOverhang(edges)

	feedback := feedbackEdges(edges)
	l.shiftForFeedback(feedback)

	c := newCanvas(l.size.x+l.overhang+l.lanePad(feedback), l.size.y+l.lanePadV(feedback))
	for _, layer := range layers {
		for _, n := range layer {
			drawNode(c, n, st, g)
		}
	}
	l.route(c, edges, g)
	l.routeFeedback(c, feedback)
	attachConnectors(c, layers, g, l.lr)

	lines := c.render(g)
	width := 0
	for _, line := range lines {
		width = max(width, stringWidth(line))
	}
	return lines, width, nil
}

// computeLayers assigns each node a layer via a longest-path ranking.
// Edges that close a cycle, including self-loops, are marked as feedback
// edges, excluded from the ranking, and routed around the drawing instead.
func computeLayers(nodes []*lnode, edges []*ledge) error {
	for {
		for _, n := range nodes {
			n.layer = 0
		}
		stuck := kahnRank(nodes, edges)
		if len(stuck) == 0 {
			return nil
		}

		pruneToCycle(stuck, edges)
		var candidate *ledge
		for _, e := range edges {
			if !e.feedback && stuck[e.from] && stuck[e.to] {
				candidate = e
			}
		}
		if candidate == nil {
			return errf("graph contains a cycle")
		}
		candidate.feedback = true
	}
}

// kahnRank ranks the nodes along non-feedback edges and returns the set of
// nodes that could not be ranked because they sit in, or downstream of, a
// cycle.
func kahnRank(nodes []*lnode, edges []*ledge) map[*lnode]bool {
	indeg := make(map[*lnode]int, len(nodes))
	adj := make(map[*lnode][]*lnode, len(nodes))
	for _, e := range edges {
		if e.feedback {
			continue
		}
		adj[e.from] = append(adj[e.from], e.to)
		indeg[e.to]++
	}

	queue := make([]*lnode, 0, len(nodes))
	for _, n := range nodes {
		if indeg[n] == 0 {
			queue = append(queue, n)
		}
	}

	processed := 0
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		processed++
		for _, s := range adj[n] {
			if n.layer+1 > s.layer {
				s.layer = n.layer + 1
			}
			if indeg[s]--; indeg[s] == 0 {
				queue = append(queue, s)
			}
		}
	}

	stuck := make(map[*lnode]bool)
	for _, n := range nodes {
		if indeg[n] > 0 {
			stuck[n] = true
		}
	}
	return stuck
}

// pruneToCycle narrows a stuck node set to the nodes actually part of a
// cycle, dropping nodes that merely feed into or out of it.
func pruneToCycle(stuck map[*lnode]bool, edges []*ledge) {
	for changed := true; changed; {
		changed = false
		for n := range stuck {
			hasIn, hasOut := false, false
			for _, e := range edges {
				if e.feedback {
					continue
				}
				if stuck[e.from] && e.to == n {
					hasIn = true
				}
				if stuck[e.to] && e.from == n {
					hasOut = true
				}
			}
			if !hasIn || !hasOut {
				delete(stuck, n)
				changed = true
			}
		}
	}
}

// insertVirtuals gives every non-feedback edge a waypoint chain, inserting
// virtual nodes on edges that skip layers.
func insertVirtuals(edges []*ledge) {
	for _, e := range edges {
		if e.feedback {
			continue
		}
		e.chain = []*lnode{e.from}
		for l := e.from.layer + 1; l < e.to.layer; l++ {
			e.chain = append(e.chain, &lnode{virtual: true, layer: l, w: 1, h: 1})
		}
		e.chain = append(e.chain, e.to)
	}
}

// feedbackEdges returns the edges routed around the drawing.
func feedbackEdges(edges []*ledge) []*ledge {
	var feedback []*ledge
	for _, e := range edges {
		if e.feedback {
			feedback = append(feedback, e)
		}
	}
	return feedback
}

// lanePad returns the canvas columns reserved for feedback lanes of
// top-down diagrams, which wrap around the right side.
func (l *layoutCtx) lanePad(feedback []*ledge) int {
	if l.lr || len(feedback) == 0 {
		return 0
	}
	return 2 + 2*len(feedback)
}

// lanePadV returns the canvas rows reserved for feedback lanes of
// left-to-right diagrams, which wrap around the bottom.
func (l *layoutCtx) lanePadV(feedback []*ledge) int {
	if !l.lr || len(feedback) == 0 {
		return 0
	}
	return 2 + 2*len(feedback)
}

// shiftForFeedback moves the drawing out of the way of feedback edges that
// would otherwise leave the canvas: a feedback edge entering the first band
// from above needs a top margin, and one leaving the last band needs an
// escape row or column below it.
func (l *layoutCtx) shiftForFeedback(feedback []*ledge) {
	if len(feedback) == 0 {
		return
	}
	lastLayer := len(l.layers) - 1
	shift, escape := false, false
	for _, e := range feedback {
		if e.to.layer == 0 {
			shift = true
		}
		if e.from.layer == lastLayer {
			escape = true
		}
	}
	if shift {
		for _, layer := range l.layers {
			for _, n := range layer {
				if l.lr {
					n.x++
				} else {
					n.y++
				}
			}
		}
		for i := range l.bandPos {
			l.bandPos[i]++
		}
	}
	// Escape room for edges leaving the last band: a row below the drawing
	// for top-down diagrams, a column to the right for left-to-right ones.
	if shift || escape {
		if l.lr {
			l.size.x += 2
		} else {
			l.size.y += 2
		}
	}
}

// routeFeedback draws feedback edges around the margin of the drawing.
func (l *layoutCtx) routeFeedback(c *canvas, feedback []*ledge) {
	for k, e := range feedback {
		if l.lr {
			l.routeBackEdgeLR(c, e, k)
		} else {
			l.routeBackEdgeTD(c, e, k)
		}
	}
}

// routeBackEdgeTD wraps a top-down feedback edge around the right side of
// the drawing: exit the source downwards, cross to a lane beyond the
// drawing, climb to the row above the target band, and enter the target
// from above.
func (l *layoutCtx) routeBackEdgeTD(c *canvas, e *ledge, lane int) {
	from, to := e.from, e.to
	sx := from.x + from.w/2
	sy := from.y + from.h
	tx := to.x + to.w/2
	ty := to.y - 1

	laneX := l.size.x + l.overhang + 1 + 2*lane
	pts := []point{
		{x: sx, y: sy},
		{x: sx, y: l.escapeBelow(from.layer)},
		{x: laneX, y: l.escapeBelow(from.layer)},
		{x: laneX, y: l.bandPos[to.layer] - 1},
		{x: tx, y: l.bandPos[to.layer] - 1},
		{x: tx, y: ty},
	}

	arrowDir := uint8(0)
	if hasArrowhead(e.kind) {
		arrowDir = dirDown
	}
	c.drawPath(pts, lineStyleFor(e.kind), arrowDir)
	l.labelTD(c, e, pts)
}

// routeBackEdgeLR wraps a left-to-right feedback edge around the bottom of
// the drawing: exit the source sideways, drop to a lane below the drawing,
// run to the column left of the target band, and enter the target from the
// left.
func (l *layoutCtx) routeBackEdgeLR(c *canvas, e *ledge, lane int) {
	from, to := e.from, e.to
	sx := from.x + from.w
	sy := from.y + from.h/2
	tx := to.x - 1
	ty := to.y + to.h/2

	laneY := l.size.y + 1 + 2*lane
	pts := []point{
		{x: sx, y: sy},
		{x: l.escapeRight(from.layer), y: sy},
		{x: l.escapeRight(from.layer), y: laneY},
		{x: l.bandPos[to.layer] - 1, y: laneY},
		{x: l.bandPos[to.layer] - 1, y: ty},
		{x: tx, y: ty},
	}

	arrowDir := uint8(0)
	if hasArrowhead(e.kind) {
		arrowDir = dirRight
	}
	c.drawPath(pts, lineStyleFor(e.kind), arrowDir)
	l.labelLR(c, e, pts)
}

// escapeBelow returns the row a feedback edge crosses on its way out of
// the given layer: the middle of the gap below the band, or a dedicated
// row underneath the drawing for the last band.
func (l *layoutCtx) escapeBelow(layer int) int {
	if layer+1 < len(l.layers) {
		return l.gapCoord(layer)
	}
	return l.size.y - 1
}

// escapeRight returns the column a feedback edge crosses on its way out of
// the given layer: the middle of the gap right of the band, or a dedicated
// column past the drawing for the last band.
func (l *layoutCtx) escapeRight(layer int) int {
	if layer+1 < len(l.layers) {
		return l.gapCoord(layer)
	}
	return l.size.x - 1
}

// buildLayers buckets nodes and virtual waypoints into their layers. Order
// within a layer is the definition order, later refined by ordering.
func buildLayers(nodes []*lnode, edges []*ledge) [][]*lnode {
	maxLayer := 0
	for _, n := range nodes {
		maxLayer = max(maxLayer, n.layer)
	}
	layers := make([][]*lnode, maxLayer+1)
	for _, n := range nodes {
		layers[n.layer] = append(layers[n.layer], n)
	}
	for _, e := range edges {
		if e.feedback {
			continue
		}
		for _, v := range e.chain[1 : len(e.chain)-1] {
			layers[v.layer] = append(layers[v.layer], v)
		}
	}
	return layers
}

// chainAdjacency builds predecessor and successor maps following the
// waypoint chains, so ordering sees the actual edge paths.
func chainAdjacency(edges []*ledge) (preds, succs map[*lnode][]*lnode) {
	preds = make(map[*lnode][]*lnode)
	succs = make(map[*lnode][]*lnode)
	for _, e := range edges {
		if e.feedback {
			continue
		}
		for i := 0; i+1 < len(e.chain); i++ {
			a, b := e.chain[i], e.chain[i+1]
			succs[a] = append(succs[a], b)
			preds[b] = append(preds[b], a)
		}
	}
	return preds, succs
}

// orderLayers reduces edge crossings with barycenter sweeps, alternating
// downward and upward.
func orderLayers(layers [][]*lnode, preds, succs map[*lnode][]*lnode) {
	pos := make(map[*lnode]float64)
	for _, layer := range layers {
		for i, n := range layer {
			pos[n] = float64(i)
		}
	}

	for sweep := 0; sweep < 4; sweep++ {
		neighbours := preds
		from, to, step := 1, len(layers), 1
		if sweep%2 != 0 {
			neighbours = succs
			from, to, step = len(layers)-2, -1, -1
		}
		for i := from; i != to; i += step {
			if i < 0 || i >= len(layers) {
				break
			}
			layer := layers[i]
			keys := make([]float64, len(layer))
			for j, n := range layer {
				keys[j] = barycenter(n, neighbours, pos)
			}
			idx := make([]int, len(layer))
			for j := range idx {
				idx[j] = j
			}
			sort.SliceStable(idx, func(a, b int) bool {
				return keys[idx[a]] < keys[idx[b]]
			})
			sorted := make([]*lnode, len(layer))
			for j, k := range idx {
				sorted[j] = layer[k]
			}
			copy(layers[i], sorted)
			for j, n := range sorted {
				pos[n] = float64(j)
			}
		}
	}
}

func barycenter(n *lnode, neighbours map[*lnode][]*lnode, pos map[*lnode]float64) float64 {
	nbs := neighbours[n]
	if len(nbs) == 0 {
		return pos[n]
	}
	sum := 0.0
	for _, nb := range nbs {
		sum += pos[nb]
	}
	return sum / float64(len(nbs))
}

// widenForLabels grows the gap between layers of left-to-right diagrams so
// edge labels fit next to the line without touching any box.
func (l *layoutCtx) widenForLabels(edges []*ledge) {
	if !l.lr {
		return
	}
	for _, e := range edges {
		if e.label != "" {
			l.st.hGap = max(l.st.hGap, min(stringWidth(e.label)+2, maxLRLabelGap))
		}
	}
}

// position places every node and records the coordinate of each layer band
// along the layer axis.
func (l *layoutCtx) position() {
	if l.lr {
		l.positionLR()
	} else {
		l.positionTD()
	}
}

func (l *layoutCtx) positionTD() {
	st := l.st
	maxW := 0
	for _, layer := range l.layers {
		maxW = max(maxW, layerWidth(layer, st.hGap, func(n *lnode) int { return n.w }))
	}

	l.bandPos = make([]int, len(l.layers))
	y := 0
	for i, layer := range l.layers {
		l.bandPos[i] = y
		layerH := layerMax(layer, func(n *lnode) int { return n.h })
		w := layerWidth(layer, st.hGap, func(n *lnode) int { return n.w })
		x := (maxW - w) / 2
		for _, n := range layer {
			n.x = x
			n.y = y + (layerH-n.h)/2
			x += n.w + st.hGap
		}
		y += layerH + st.vGap
	}
	l.size = point{maxW, y - st.vGap}
}

func (l *layoutCtx) positionLR() {
	st := l.st
	maxH := 0
	for _, layer := range l.layers {
		maxH = max(maxH, layerWidth(layer, st.vGap, func(n *lnode) int { return n.h }))
	}

	l.bandPos = make([]int, len(l.layers))
	x := 0
	for i, layer := range l.layers {
		l.bandPos[i] = x
		layerW := layerMax(layer, func(n *lnode) int { return n.w })
		h := layerWidth(layer, st.vGap, func(n *lnode) int { return n.h })
		y := (maxH - h) / 2
		for _, n := range layer {
			n.x = x + (layerW-n.w)/2
			n.y = y
			y += n.h + st.vGap
		}
		x += layerW + st.hGap
	}
	l.size = point{x - st.hGap, maxH}
}

// layerWidth packs the nodes of one layer along the stacking axis and
// returns the extent, including size gaps.
func layerWidth(layer []*lnode, gap int, size func(*lnode) int) int {
	if len(layer) == 0 {
		return 0
	}
	total := (len(layer) - 1) * gap
	for _, n := range layer {
		total += size(n)
	}
	return max(1, total)
}

// layerMax returns the extent of one layer along the layer axis, which is
// the largest single node.
func layerMax(layer []*lnode, size func(*lnode) int) int {
	m := 0
	for _, n := range layer {
		m = max(m, size(n))
	}
	return max(1, m)
}

// gapCoord returns the row (top-down) or column (left-to-right) in the
// middle of the gap between bands i and i+1, where elbow segments run.
func (l *layoutCtx) gapCoord(i int) int {
	gap := l.st.vGap
	if l.lr {
		gap = l.st.hGap
	}
	return l.bandPos[i+1] - 1 - (gap-1)/2
}

// route draws all non-feedback edges onto the canvas.
func (l *layoutCtx) route(c *canvas, edges []*ledge, g glyphSet) {
	for _, e := range edges {
		if e.feedback {
			continue
		}
		if l.lr {
			l.routeEdgeLR(c, e)
		} else {
			l.routeEdgeTD(c, e)
		}
	}
}

// routeEdgeTD routes a top-down edge: descend from the source, cross
// through the gaps below each intermediate layer, and enter the target from
// above.
func (l *layoutCtx) routeEdgeTD(c *canvas, e *ledge) {
	from, to := e.from, e.to
	sx := from.x + from.w/2
	sy := from.y + from.h
	tx := to.x + to.w/2
	ty := to.y - 1

	pts := []point{{x: sx, y: sy}}
	x, li := sx, from.layer
	for _, v := range e.chain[1 : len(e.chain)-1] {
		gy := l.gapCoord(li)
		pts = append(pts, point{x: x, y: gy}, point{x: v.x, y: gy})
		x = v.x
		li++
	}
	gy := l.gapCoord(li)
	pts = append(pts, point{x: x, y: gy}, point{x: tx, y: gy}, point{x: tx, y: ty})

	arrowDir := uint8(0)
	if hasArrowhead(e.kind) {
		arrowDir = dirDown
	}
	c.drawPath(pts, lineStyleFor(e.kind), arrowDir)
	l.labelTD(c, e, pts)
}

// routeEdgeLR routes a left-to-right edge: the mirror of the top-down
// routing, with horizontal and vertical roles swapped.
func (l *layoutCtx) routeEdgeLR(c *canvas, e *ledge) {
	from, to := e.from, e.to
	sx := from.x + from.w
	sy := from.y + from.h/2
	tx := to.x - 1
	ty := to.y + to.h/2

	pts := []point{{x: sx, y: sy}}
	y, li := sy, from.layer
	for _, v := range e.chain[1 : len(e.chain)-1] {
		gx := l.gapCoord(li)
		pts = append(pts, point{x: gx, y: y}, point{x: gx, y: v.y})
		y = v.y
		li++
	}
	gx := l.gapCoord(li)
	pts = append(pts, point{x: gx, y: y}, point{x: gx, y: ty}, point{x: tx, y: ty})

	arrowDir := uint8(0)
	if hasArrowhead(e.kind) {
		arrowDir = dirRight
	}
	c.drawPath(pts, lineStyleFor(e.kind), arrowDir)
	l.labelLR(c, e, pts)
}

// labelTD stamps the edge label onto the first horizontal segment, or next
// to the line for straight vertical edges.
func (l *layoutCtx) labelTD(c *canvas, e *ledge, pts []point) {
	if e.label == "" {
		return
	}
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		if a.y == b.y && a.x != b.x {
			w := stringWidth(e.label)
			mid := (a.x + b.x) / 2
			c.labelSoft(point{x: mid - w/2, y: a.y}, " "+e.label+" ")
			return
		}
	}
	midY := (pts[0].y + pts[len(pts)-1].y) / 2
	c.labelSoft(point{x: pts[0].x + 1, y: midY}, " "+e.label)
}

// labelLR stamps the edge label next to the first vertical segment, or
// above the line for straight horizontal edges.
func (l *layoutCtx) labelLR(c *canvas, e *ledge, pts []point) {
	if e.label == "" {
		return
	}
	if e.feedback {
		for i := 0; i+1 < len(pts); i++ {
			a, b := pts[i], pts[i+1]
			if a.y == b.y && abs(b.x-a.x) > stringWidth(e.label)+2 {
				mid := (a.x + b.x) / 2
				w := stringWidth(e.label)
				c.labelSoft(point{x: mid - w/2, y: a.y}, " "+e.label+" ")
				return
			}
		}
	}
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		if a.x == b.x && a.y != b.y {
			w := stringWidth(e.label)
			mid := (a.y + b.y) / 2
			c.labelSoft(point{x: a.x - w/2, y: mid}, " "+e.label+" ")
			return
		}
	}
	w := stringWidth(e.label)
	midX := (pts[0].x + pts[len(pts)-1].x) / 2
	c.labelSoft(point{x: midX - w/2, y: pts[0].y - 1}, e.label)
}

// drawNode draws a node as a bordered box with its label centred. Virtual
// waypoints draw nothing.
func drawNode(c *canvas, n *lnode, st settings, g glyphSet) {
	if n.virtual {
		return
	}

	rounded := n.node.shape == ShapeRounded ||
		n.node.shape == ShapeStadium ||
		n.node.shape == ShapeCircle
	tl, tr, bl, br := g.tl, g.tr, g.bl, g.br
	if rounded {
		tl, tr, bl, br = g.rtl, g.rtr, g.rbl, g.rbr
	}

	c.text(n.x, n.y, tl)
	c.text(n.x+n.w-1, n.y, tr)
	c.text(n.x, n.y+n.h-1, bl)
	c.text(n.x+n.w-1, n.y+n.h-1, br)
	for x := n.x + 1; x < n.x+n.w-1; x++ {
		c.text(x, n.y, g.h)
		c.text(x, n.y+n.h-1, g.h)
	}

	inner := n.w - 2 - 2*st.boxPad
	labelAt := func(row int, line string) {
		c.text(n.x, row, g.v)
		c.text(n.x+n.w-1, row, g.v)
		off := (inner - stringWidth(line)) / 2
		if off < 0 {
			off = 0
		}
		c.label(point{x: n.x + 1 + st.boxPad + off, y: row}, line)
	}
	for i, line := range n.lines {
		labelAt(n.y+1+st.boxPad+i, line)
	}
	for p := 0; p < st.boxPad; p++ {
		labelAt(n.y+1+p, "")
		labelAt(n.y+n.h-2-p, "")
	}
}

// labelOverhang returns extra canvas columns so edge labels stamped next
// to vertical lines are not clipped; trailing blanks are trimmed on render,
// so this costs nothing when unused.
func labelOverhang(edges []*ledge) int {
	w := 0
	for _, e := range edges {
		if e.label != "" {
			w = max(w, stringWidth(e.label))
		}
	}
	if w > 0 {
		return min(w+2, maxLRLabelGap+4)
	}
	return 0
}

// attachConnectors opens the box borders where edges attach, so lines
// visually join the nodes.
func attachConnectors(c *canvas, layers [][]*lnode, g glyphSet, lr bool) {
	for _, layer := range layers {
		for _, n := range layer {
			if n.virtual {
				continue
			}
			if lr {
				cy := n.y + n.h/2
				if connects(c, n.x+n.w, cy) {
					c.text(n.x+n.w-1, cy, g.teeRight)
				}
				if connects(c, n.x-1, cy) {
					c.text(n.x, cy, g.teeLeft)
				}
			} else {
				cx := n.x + n.w/2
				if connects(c, cx, n.y+n.h) {
					c.text(cx, n.y+n.h-1, g.teeDown)
				}
				if connects(c, cx, n.y-1) {
					c.text(cx, n.y, g.teeUp)
				}
			}
		}
	}
}

// connects reports whether the cell at x, y carries any line or arrowhead.
func connects(c *canvas, x, y int) bool {
	if !c.inside(x, y) {
		return false
	}
	cl := c.cells[y][x]
	return cl.bits != 0 || cl.arrow != 0
}

// wrapLabel wraps a label to maxW display columns, hard-splitting words
// that are too long, and honouring explicit line breaks.
func wrapLabel(label string, maxW int) []string {
	if maxW < 1 {
		maxW = 1
	}
	var lines []string
	for _, hard := range strings.Split(label, "\n") {
		words := strings.Fields(hard)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		var cur string
		for _, word := range words {
			for stringWidth(word) > maxW {
				prefix := clampWord(word, maxW)
				if prefix == "" {
					prefix = string([]rune(word)[0])
				}
				if cur != "" {
					lines = append(lines, cur)
					cur = ""
				}
				lines = append(lines, prefix)
				word = word[len(prefix):]
			}

			if cur == "" {
				cur = word
			} else if stringWidth(cur)+1+stringWidth(word) <= maxW {
				cur += " " + word
			} else {
				lines = append(lines, cur)
				cur = word
			}
		}
		if cur != "" {
			lines = append(lines, cur)
		}
	}
	return lines
}

// clampWord returns the longest prefix of word whose display width is at
// most maxW.
func clampWord(word string, maxW int) string {
	w := 0
	for i, r := range word {
		rw := runeWidth(r)
		if w+rw > maxW {
			return word[:i]
		}
		w += rw
	}
	return word
}
