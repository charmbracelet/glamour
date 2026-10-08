package mermaid

import (
	"os"
	"strings"
)

// Direction bits describing which neighbours a canvas cell connects to.
const (
	dirUp uint8 = 1 << iota
	dirDown
	dirLeft
	dirRight
)

// lineStyle describes how an edge is drawn.
type lineStyle uint8

// Line styles.
const (
	lineSolid lineStyle = iota
	lineDotted
	lineThick
)

// glyphSet holds the drawing characters for one output flavour.
type glyphSet struct {
	h, v               rune
	dotH, dotV         rune
	thickH, thickV     rune
	tl, tr, bl, br     rune
	rtl, rtr, rbl, rbr rune
	teeDown, teeUp     rune
	teeLeft, teeRight  rune
	cross              rune
	arrowUp            rune
	arrowDown          rune
	arrowLeft          rune
	arrowRight         rune
	openRight          rune
	openLeft           rune
	crossGlyph         rune
	bar                rune
	circle             rune
	barDone            rune
	barActive          rune
	barCrit            rune
	milestone          rune
}

var unicodeGlyphs = glyphSet{
	h: '─', v: '│',
	dotH: '┄', dotV: '┆',
	thickH: '━', thickV: '┃',
	tl: '┌', tr: '┐', bl: '└', br: '┘',
	rtl: '╭', rtr: '╮', rbl: '╰', rbr: '╯',
	teeDown: '┬', teeUp: '┴', teeLeft: '┤', teeRight: '├',
	cross:   '┼',
	arrowUp: '▲', arrowDown: '▼', arrowLeft: '◀', arrowRight: '▶',
	openRight: '▷', openLeft: '◁', crossGlyph: '✗',
	bar: '█', barDone: '░', barActive: '▓', barCrit: '▒', milestone: '◆', circle: '●',
}

var asciiGlyphs = glyphSet{
	h: '-', v: '|',
	dotH: '.', dotV: ':',
	thickH: '=', thickV: '|',
	tl: '+', tr: '+', bl: '+', br: '+',
	rtl: '+', rtr: '+', rbl: '+', rbr: '+',
	teeDown: '+', teeUp: '+', teeLeft: '+', teeRight: '+',
	cross:   '+',
	arrowUp: '^', arrowDown: 'v', arrowLeft: '<', arrowRight: '>',
	openRight: '>', openLeft: '<', crossGlyph: 'X',
	bar: '#', barDone: '.', barActive: '@', barCrit: '*', milestone: 'D', circle: 'o',
}

// line returns the glyph for a cell whose connections are bits, drawn with
// the given style.
func (g glyphSet) line(bits uint8, style lineStyle) rune {
	switch bits {
	case dirUp | dirDown:
		return g.axisV(style)
	case dirLeft | dirRight:
		return g.axisH(style)
	case dirDown | dirRight:
		return g.tl
	case dirDown | dirLeft:
		return g.tr
	case dirUp | dirRight:
		return g.bl
	case dirUp | dirLeft:
		return g.br
	case dirUp | dirDown | dirRight:
		return g.teeRight
	case dirUp | dirDown | dirLeft:
		return g.teeLeft
	case dirDown | dirLeft | dirRight:
		return g.teeDown
	case dirUp | dirLeft | dirRight:
		return g.teeUp
	case dirUp | dirDown | dirLeft | dirRight:
		return g.cross
	case dirUp, dirDown:
		return g.axisV(style)
	case dirLeft, dirRight:
		return g.axisH(style)
	}
	return ' '
}

func (g glyphSet) axisV(style lineStyle) rune {
	switch style {
	case lineSolid:
		return g.v
	case lineDotted:
		return g.dotV
	case lineThick:
		return g.thickV
	}
	return g.v
}

func (g glyphSet) axisH(style lineStyle) rune {
	switch style {
	case lineSolid:
		return g.h
	case lineDotted:
		return g.dotH
	case lineThick:
		return g.thickH
	}
	return g.h
}

// arrow returns the arrowhead glyph pointing in the given direction.
func (g glyphSet) arrow(dir uint8) rune {
	switch dir {
	case dirUp:
		return g.arrowUp
	case dirDown:
		return g.arrowDown
	case dirLeft:
		return g.arrowLeft
	case dirRight:
		return g.arrowRight
	}
	return ' '
}

// detectASCII reports whether the environment locale suggests a terminal
// that cannot display Unicode box-drawing characters. Terminals with no
// locale set at all are assumed to be UTF-8 capable.
func detectASCII() bool {
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := os.Getenv(key); v != "" {
			return !strings.Contains(strings.ToLower(v), "utf")
		}
	}
	return false
}
