package mermaid

// Sequence guards.
const (
	maxSeqParticipants = 20
	maxSeqEvents       = 200
)

// seqLayout carries the sequence diagram through layout and drawing.
type seqLayout struct {
	d        *seqDiagram
	st       settings
	g        glyphSet
	c        *canvas
	x, w, cx []int
	lines    [][]string
	headerH  int
	y        int
	maxY     int
	width    int
}

// drawSequenceFit renders a sequence diagram, compacting it when it does
// not fit the limit.
func drawSequenceFit(d *seqDiagram, limit int, g glyphSet) ([]string, error) {
	if len(d.participants) > maxSeqParticipants {
		return nil, errf("diagram too large (over %d participants)", maxSeqParticipants)
	}
	if countSeqEvents(d.events) > maxSeqEvents {
		return nil, errf("diagram too large (over %d events)", maxSeqEvents)
	}

	lines, width, err := drawSequence(d, defaultSettings, g)
	if err != nil {
		return nil, err
	}
	if limit > 0 && width > limit {
		lines, width, err = drawSequence(d, compactSettings, g)
		if err != nil {
			return nil, err
		}
		if width > limit {
			return nil, errf("too wide to render (needs %d columns, %d available)", width, limit)
		}
	}
	if d.title != "" {
		lines = append([]string{d.title, ""}, lines...)
	}
	return lines, nil
}

func countSeqEvents(events []seqEvent) int {
	count := 0
	for _, ev := range events {
		count += 1 + countSeqEvents(ev.events)
	}
	return count
}

// drawSequence lays out and draws a sequence diagram.
func drawSequence(d *seqDiagram, st settings, g glyphSet) ([]string, int, error) {
	l := &seqLayout{d: d, st: st, g: g}
	n := len(d.participants)
	l.x = make([]int, n)
	l.w = make([]int, n)
	l.cx = make([]int, n)
	l.lines = make([][]string, n)

	x := 0
	for i, p := range d.participants {
		lines := wrapLabel(p.label, st.maxLabelW)
		w := 0
		for _, line := range lines {
			w = max(w, stringWidth(line))
		}
		w += 2*st.boxPad + 2
		l.lines[i] = lines
		l.w[i] = w
		l.x[i] = x
		l.cx[i] = x + w/2
		l.headerH = max(l.headerH, len(lines)+2*st.boxPad+2)
		x += w + st.hGap + 2
	}

	shift := l.leftShift()
	for i := range d.participants {
		l.x[i] += shift
		l.cx[i] += shift
	}
	l.width = x - st.hGap - 2 + shift + l.rightExtra()

	l.y = l.headerH + 2
	height := l.headerH + 3
	for _, ev := range d.events {
		height += measureSeqEvent(ev, st) + 1
	}
	l.c = newCanvas(l.width+4, height)

	for i, p := range d.participants {
		drawNode(l.c, &lnode{node: &node{label: p.label}, lines: l.lines[i],
			x: l.x[i], y: 0, w: l.w[i], h: len(l.lines[i]) + 2*st.boxPad + 2}, st, g)
	}

	for _, ev := range d.events {
		l.event(ev, 0)
		l.y++
	}
	l.maxY = max(l.maxY, l.y)

	// Lifelines last: they pass under everything already drawn.
	for i := range d.participants {
		for y := l.headerH + 1; y < l.maxY; y++ {
			l.c.mark(l.cx[i], y, dirUp|dirDown, lineDotted)
		}
	}

	lines := l.c.render(g)
	width := 0
	for _, line := range lines {
		width = max(width, stringWidth(line))
	}
	return lines, width, nil
}

// rightExtra returns the columns needed right of the last participant for
// self-messages, right notes and overflowing labels. Rendered lines are
// right-trimmed, so a generous value costs nothing.
func (l *seqLayout) rightExtra() int {
	extra := 4
	for _, ev := range allEvents(l.d.events) {
		if w := stringWidth(ev.label); w > 0 {
			extra = max(extra, w+4)
		}
	}
	return extra
}

// leftShift returns the columns needed left of the first participant so
// notes never start off the canvas.
func (l *seqLayout) leftShift() int {
	shift := 0
	for _, ev := range allEvents(l.d.events) {
		if ev.kind != seqNote {
			continue
		}
		w := stringWidth(ev.label) + 2
		if ev.side < 0 {
			shift = max(shift, w+2-l.cx[ev.from])
		} else if ev.side == 0 {
			shift = max(shift, w/2+1-l.cx[ev.from])
		}
	}
	return shift
}

// allEvents flattens nested fragment events.
func allEvents(events []seqEvent) []seqEvent {
	var flat []seqEvent
	for _, ev := range events {
		flat = append(flat, ev)
		flat = append(flat, allEvents(ev.events)...)
	}
	return flat
}

// measureSeqEvent returns the rows an event occupies, mirroring the layout.
func measureSeqEvent(ev seqEvent, st settings) int {
	switch ev.kind {
	case seqMessage:
		return 2
	case seqSelf:
		return 3
	case seqNote:
		return len(wrapLabel(ev.label, st.maxLabelW)) + 2
	case seqFragment:
		rows := 2
		for _, child := range ev.events {
			rows += measureSeqEvent(child, st) + 1
		}
		return rows
	}
	return 1
}

func (l *seqLayout) event(ev seqEvent, depth int) {
	switch ev.kind {
	case seqMessage:
		l.message(ev)
	case seqSelf:
		l.selfMessage(ev)
	case seqNote:
		l.note(ev)
	case seqFragment:
		l.fragment(ev, depth)
	}
	l.maxY = max(l.maxY, l.y)
}

// message draws an arrow between two lifelines with its label above.
func (l *seqLayout) message(ev seqEvent) {
	from, to := l.cx[ev.from], l.cx[ev.to]
	r := l.y + 1

	w := stringWidth(ev.label)
	mid := (from + to) / 2
	l.c.labelSoft(point{x: mid - w/2, y: l.y}, ev.label)

	style := lineSolid
	if ev.dotted {
		style = lineDotted
	}
	right := to > from
	if right {
		l.c.drawPath([]point{{from + 1, r}, {to - 1, r}}, style, 0)
		l.c.text(to-1, r, l.headGlyph(ev.head, right))
	} else {
		l.c.drawPath([]point{{from - 1, r}, {to + 1, r}}, style, 0)
		l.c.text(to+1, r, l.headGlyph(ev.head, right))
	}
	l.y = r + 1
}

// selfMessage draws a loop to the right of the lifeline.
func (l *seqLayout) selfMessage(ev seqEvent) {
	cx := l.cx[ev.from]
	loopW := max(stringWidth(ev.label)+4, 6)
	style := lineSolid
	if ev.dotted {
		style = lineDotted
	}

	w := stringWidth(ev.label)
	l.c.labelSoft(point{x: cx + 2 + max(0, (loopW-w)/2), y: l.y}, ev.label)
	l.c.drawPath([]point{
		{cx, l.y + 1},
		{cx + loopW, l.y + 1},
		{cx + loopW, l.y + 2},
		{cx, l.y + 2},
	}, style, 0)
	l.c.text(cx, l.y+2, l.headGlyph(ev.head, false))
	l.y += 3
}

// note draws a boxed note near its participants.
func (l *seqLayout) note(ev seqEvent) {
	lines := wrapLabel(ev.label, l.st.maxLabelW)
	w := 0
	for _, line := range lines {
		w = max(w, stringWidth(line))
	}
	w += 2
	h := len(lines) + 2

	x0 := 0
	switch {
	case ev.side < 0:
		x0 = l.cx[ev.from] - w - 2
	case ev.side > 0:
		x0 = l.cx[ev.from] + 2
	default:
		x0 = (l.cx[ev.from]+l.cx[ev.to])/2 - w/2
	}

	// Fill the interior so lifelines do not show through the note.
	for y := l.y; y < l.y+h; y++ {
		for x := x0; x < x0+w; x++ {
			l.c.text(x, y, ' ')
		}
	}
	drawNode(l.c, &lnode{node: &node{}, lines: lines, x: x0, y: l.y, w: w, h: h}, settings{}, l.g)
	l.y += h
}

// fragment draws a labelled frame around its children, with dividers
// between branches.
func (l *seqLayout) fragment(ev seqEvent, depth int) {
	n := len(l.x)
	x0 := max(0, l.x[0]-1+2*depth)
	x1 := l.x[n-1] + l.w[n-1] + 1 - 2*depth
	if x1-x0 < 10 {
		x0 = max(0, x1-10)
	}

	top := l.y
	l.y++
	for _, child := range ev.events {
		if child.kind == seqDivider {
			l.divider(child, x0, x1)
			continue
		}
		l.event(child, depth+1)
		l.y++
	}
	bottom := l.y

	l.hBorder(x0, x1, top, true)
	l.hBorder(x0, x1, bottom, false)
	label := ev.frag
	if ev.label != "" {
		label += ": " + ev.label
	}
	l.c.label(point{x: x0 + 2, y: top}, label)
	for y := top + 1; y < bottom; y++ {
		l.c.text(x0, y, l.g.v)
		l.c.text(x1, y, l.g.v)
	}
	l.y++
}

// divider draws an else/and/option branch separator.
func (l *seqLayout) divider(ev seqEvent, x0, x1 int) {
	for x := x0; x <= x1; x++ {
		l.c.text(x, l.y, l.g.h)
	}
	l.c.text(x0, l.y, l.g.teeRight)
	l.c.text(x1, l.y, l.g.teeLeft)
	label := ev.frag
	if ev.label != "" {
		label += ": " + ev.label
	}
	l.c.label(point{x: x0 + 2, y: l.y}, label)
	l.y++
}

// hBorder draws a fragment top or bottom border.
func (l *seqLayout) hBorder(x0, x1, y int, top bool) {
	for x := x0; x <= x1; x++ {
		l.c.text(x, y, l.g.h)
	}
	if top {
		l.c.text(x0, y, l.g.tl)
		l.c.text(x1, y, l.g.tr)
	} else {
		l.c.text(x0, y, l.g.bl)
		l.c.text(x1, y, l.g.br)
	}
}

// headGlyph returns the arrowhead glyph for a message.
func (l *seqLayout) headGlyph(head seqHead, right bool) rune {
	switch head {
	case seqHeadFilled:
		if right {
			return l.g.arrowRight
		}
		return l.g.arrowLeft
	case seqHeadOpen:
		if right {
			return l.g.openRight
		}
		return l.g.openLeft
	default:
		return l.g.crossGlyph
	}
}
