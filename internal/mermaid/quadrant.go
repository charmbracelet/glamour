package mermaid

import (
	"strconv"
	"strings"
)

type quadrantPoint struct {
	name string
	x, y float64
}

// renderQuadrant renders a quadrant chart as a grid with plotted points.
func renderQuadrant(src string, limit int, g glyphSet) ([]string, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)

	var (
		xAxis     [2]string
		yAxis     [2]string
		quadrants [4]string
		points    []quadrantPoint
		seen      bool
	)

	lineno := 0
	for _, line := range strings.Split(src, "\n") {
		lineno++
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		word := strings.ToLower(firstWord(trimmed))
		rest := strings.TrimSpace(trimmed[len(word):])
		if !seen {
			if word == "quadrantchart" {
				seen = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "quadrantChart"`, lineno)
		}
		switch word {
		case kwTitle:
			if title == "" {
				title = rest
			}
		case "x-axis", "xaxis":
			lo, hi, ok := axisPair(rest)
			if !ok {
				return nil, errf("could not parse (line %d): axis needs two labels", lineno)
			}
			xAxis = [2]string{lo, hi}
		case "y-axis", "yaxis":
			lo, hi, ok := axisPair(rest)
			if !ok {
				return nil, errf("could not parse (line %d): axis needs two labels", lineno)
			}
			yAxis = [2]string{lo, hi}
		case "quadrant-1", "quadrant-2", "quadrant-3", "quadrant-4":
			quadrants[word[len(word)-1]-'1'] = rest
		default:
			name, coords, ok := strings.Cut(trimmed, ":")
			if !ok {
				return nil, errf("could not parse (line %d): %q", lineno, trimmed)
			}
			p, err := parseQuadrantPoint(strings.TrimSpace(name), strings.TrimSpace(coords))
			if err != nil {
				return nil, errf("%v (line %d)", err, lineno)
			}
			points = append(points, p)
		}
	}
	if !seen {
		return nil, errf("empty diagram")
	}

	w, h := 40, 13
	if limit > 0 {
		w = min(40, max(24, limit-4))
	}

	grid := make([][]rune, h)
	for y := range grid {
		grid[y] = []rune(strings.Repeat(" ", w))
	}
	for x := 0; x < w; x++ {
		grid[0][x], grid[h-1][x], grid[h/2][x] = g.h, g.h, g.h
	}
	for y := 0; y < h; y++ {
		grid[y][0], grid[y][w-1], grid[y][w/2] = g.v, g.v, g.v
	}
	grid[0][0], grid[0][w-1] = g.tl, g.tr
	grid[h-1][0], grid[h-1][w-1] = g.bl, g.br
	grid[h/2][0], grid[h/2][w-1] = g.teeRight, g.teeLeft
	grid[0][w/2], grid[h-1][w/2] = g.teeDown, g.teeUp
	grid[h/2][w/2] = g.cross

	corner := func(row, col int, s string) {
		for i, r := range truncateLabel(s, w/2-2) {
			if col+i < w-1 {
				grid[row][col+i] = r
			}
		}
	}
	corner(1, 2, quadrants[1])
	corner(1, w/2+2, quadrants[0])
	corner(h-2, 2, quadrants[2])
	corner(h-2, w/2+2, quadrants[3])

	for _, p := range points {
		x := min(max(1+int(clamp01(p.x)*float64(w-3)), 1), w-2)
		y := min(max(h-2-int(clamp01(p.y)*float64(h-3)), 1), h-2)
		marker := []rune("•")
		if p.name != "" {
			marker = []rune(p.name)[:1]
		}
		grid[y][x] = marker[0]
	}

	var lines []string
	if yAxis[1] != "" {
		lines = append(lines, yAxis[1])
	}
	for _, row := range grid {
		lines = append(lines, string(row))
	}
	gap := max(0, w-stringWidth(yAxis[0])-1-stringWidth(xAxis[0])-stringWidth(xAxis[1]))
	lines = append(lines, yAxis[0]+" "+xAxis[0]+strings.Repeat(" ", gap)+xAxis[1])
	if title != "" {
		lines = append([]string{title, ""}, lines...)
	}
	return lines, nil
}

func parseQuadrantPoint(name, coords string) (quadrantPoint, error) {
	open := strings.IndexByte(coords, '[')
	closeIdx := strings.LastIndexByte(coords, ']')
	if open < 0 || closeIdx <= open {
		return quadrantPoint{}, errf("expected [x, y]")
	}
	parts := strings.Split(coords[open+1:closeIdx], ",")
	if len(parts) != 2 {
		return quadrantPoint{}, errf("expected [x, y]")
	}
	x, errX := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	y, errY := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if errX != nil || errY != nil {
		return quadrantPoint{}, errf("bad coordinates")
	}
	return quadrantPoint{name: strings.Trim(name, `"`), x: x, y: y}, nil
}

// axisPair splits `Low --> High` into its two labels.
func axisPair(rest string) (string, string, bool) {
	lo, hi, ok := strings.Cut(rest, "-->")
	if !ok {
		return "", "", false
	}
	return strings.Trim(strings.TrimSpace(lo), `"`), strings.Trim(strings.TrimSpace(hi), `"`), true
}

func clamp01(v float64) float64 {
	return min(max(v, 0), 1)
}
