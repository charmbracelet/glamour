package mermaid

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// cell is one canvas position: either a text rune, an arrowhead, or the
// connections of a line passing through.
type cell struct {
	text  rune
	wide  bool
	bits  uint8
	style lineStyle
	arrow uint8
}

// canvas is the grid diagrams are drawn onto.
type canvas struct {
	w, h  int
	cells [][]cell
}

func newCanvas(w, h int) *canvas {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	cells := make([][]cell, h)
	rows := make([]cell, w*h)
	for i := range cells {
		cells[i], rows = rows[:w], rows[w:]
	}
	return &canvas{w: w, h: h, cells: cells}
}

func (c *canvas) inside(x, y int) bool {
	return x >= 0 && y >= 0 && x < c.w && y < c.h
}

// text stamps a rune, which takes precedence over any line passing through
// the cell. Wide runes reserve the following cell as well.
func (c *canvas) text(x, y int, r rune) {
	if !c.inside(x, y) {
		return
	}
	cl := &c.cells[y][x]
	cl.text = r
	cl.bits = 0
	cl.arrow = 0
	cl.style = lineSolid
	if runeWidth(r) > 1 && c.inside(x+1, y) {
		next := &c.cells[y][x+1]
		next.wide = true
		next.text = 0
		next.bits = 0
		next.arrow = 0
	}
}

// mark records a line connection on a cell.
func (c *canvas) mark(x, y int, bits uint8, style lineStyle) {
	if !c.inside(x, y) || bits == 0 {
		return
	}
	cl := &c.cells[y][x]
	if cl.text != 0 || cl.wide {
		return
	}
	if cl.style != style && cl.bits != 0 {
		// Lines of different styles sharing a cell cross; draw the
		// crossing as a solid junction.
		cl.style = lineSolid
	} else {
		cl.style = style
	}
	cl.bits |= bits
}

// arrow stamps an arrowhead pointing in dir.
func (c *canvas) arrow(x, y int, dir uint8) {
	if !c.inside(x, y) {
		return
	}
	cl := &c.cells[y][x]
	if cl.text != 0 {
		return
	}
	cl.arrow = dir
}

// drawPath draws a polyline through the given points, marking each cell
// with connections toward its path neighbours, and stamps an arrowhead on
// the final point when arrowDir is non-zero.
func (c *canvas) drawPath(pts []point, style lineStyle, arrowDir uint8) {
	pts = dedupPoints(pts)
	if len(pts) < 2 {
		return
	}

	var cells []point
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		dx, dy := sign(b.x-a.x), sign(b.y-a.y)
		if dx != 0 && dy != 0 {
			return
		}
		cells = append(cells, a)
		x, y := a.x+dx, a.y+dy
		for x != b.x || y != b.y {
			cells = append(cells, point{x, y})
			x += dx
			y += dy
		}
	}
	cells = append(cells, pts[len(pts)-1])

	for i, p := range cells {
		var bits uint8
		if i > 0 {
			bits |= direction(cells[i-1], p)
		}
		if i+1 < len(cells) {
			bits |= direction(cells[i+1], p)
		}
		c.mark(p.x, p.y, bits, style)
	}

	if arrowDir != 0 {
		last := pts[len(pts)-1]
		c.arrow(last.x, last.y, arrowDir)
	}
}

// direction returns the direction bit for the side of p that faces
// neighbour.
func direction(neighbour, p point) uint8 {
	switch {
	case neighbour.y > p.y:
		return dirDown
	case neighbour.y < p.y:
		return dirUp
	case neighbour.x > p.x:
		return dirRight
	default:
		return dirLeft
	}
}

// label stamps a string of runes starting at p.
func (c *canvas) label(p point, s string) {
	x := p.x
	for _, r := range s {
		c.text(x, p.y, r)
		x += runeWidth(r)
	}
}

// labelSoft stamps a string of runes, skipping cells already holding text
// so labels never overwrite node boxes.
func (c *canvas) labelSoft(p point, s string) {
	x := p.x
	for _, r := range s {
		if c.inside(x, p.y) {
			cl := &c.cells[p.y][x]
			if cl.text == 0 && !cl.wide {
				c.text(x, p.y, r)
			}
		}
		x += runeWidth(r)
	}
}

// render turns the canvas into lines with trailing blanks trimmed.
func (c *canvas) render(g glyphSet) []string {
	lines := make([]string, 0, c.h)
	for y := 0; y < c.h; y++ {
		var b strings.Builder
		b.Grow(c.w)
		for x := 0; x < c.w; x++ {
			cl := c.cells[y][x]
			switch {
			case cl.wide:
				continue
			case cl.text != 0:
				b.WriteRune(cl.text)
			case cl.arrow != 0:
				b.WriteRune(g.arrow(cl.arrow))
			case cl.bits != 0:
				b.WriteRune(g.line(cl.bits, cl.style))
			default:
				b.WriteByte(' ')
			}
		}
		lines = append(lines, strings.TrimRight(b.String(), " "))
	}

	start, end := 0, len(lines)
	for start < end && lines[start] == "" {
		start++
	}
	for end > start && lines[end-1] == "" {
		end--
	}
	return lines[start:end]
}

// stringWidth returns the display width of s in terminal cells.
func stringWidth(s string) int {
	return ansi.StringWidth(s)
}

// runeWidth returns the display width of r in terminal cells.
func runeWidth(r rune) int {
	if r >= 0x20 && r < 0x7f {
		return 1
	}
	return max(1, ansi.StringWidth(string(r)))
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}

func dedupPoints(pts []point) []point {
	out := make([]point, 0, len(pts))
	for _, p := range pts {
		if len(out) == 0 || out[len(out)-1] != p {
			out = append(out, p)
		}
	}
	return out
}
