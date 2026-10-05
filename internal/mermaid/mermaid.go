// Package mermaid renders mermaid flowchart definitions as box-drawing
// character diagrams suitable for terminal output. Only flowcharts (graph
// and flowchart) are supported; anything the parser cannot handle returns
// an error describing why, so callers can fall back to showing the source
// instead.
package mermaid

import (
	"fmt"
	"strings"
)

// Guards bounding the worst case before any layout work happens.
const (
	maxSourceBytes = 8 << 10
	maxSourceLines = 400
	maxNodes       = 100
	maxEdges       = 200
)

// point is a canvas coordinate.
type point struct {
	x, y int
}

// renderError carries a user-facing reason a diagram could not be rendered.
type renderError struct{ msg string }

func (e *renderError) Error() string { return e.msg }

func errf(format string, args ...any) error {
	return &renderError{msg: fmt.Sprintf(format, args...)}
}

// Render renders the mermaid diagram in src as lines of terminal art, at
// most limit display columns wide; a limit of zero disables fitting.
// Glyphs are Unicode box-drawing unless the environment locale suggests
// otherwise.
func Render(src string, limit int) ([]string, error) {
	g := unicodeGlyphs
	if detectASCII() {
		g = asciiGlyphs
	}
	return renderCached(src, limit, g)
}

// render renders src without touching the cache, for tests and cache
// misses.
func render(src string, limit int, g glyphSet) ([]string, error) {
	if len(src) > maxSourceBytes {
		return nil, errf("diagram too large (over %d bytes)", maxSourceBytes)
	}
	if strings.Count(src, "\n")+1 > maxSourceLines {
		return nil, errf("diagram too large (over %d lines)", maxSourceLines)
	}

	switch diagramKind(src) {
	case kindSequence:
		d, err := parseSequence(src)
		if err != nil {
			return nil, err
		}
		return drawSequenceFit(d, limit, g)
	case kindGantt:
		d, err := parseGantt(src)
		if err != nil {
			return nil, err
		}
		return drawGanttFit(d, limit, g)
	case kindState:
		d, err := parseState(src)
		if err != nil {
			return nil, err
		}
		return drawFlowchartFit(d, limit, g)
	case kindER:
		d, err := parseER(src)
		if err != nil {
			return nil, err
		}
		return drawFlowchartFit(d, limit, g)
	case kindClass:
		d, err := parseClass(src)
		if err != nil {
			return nil, err
		}
		return drawFlowchartFit(d, limit, g)
	case kindPie:
		return renderPie(src, limit, g)
	case kindJourney:
		return renderJourney(src, limit, g)
	case kindTimeline:
		return renderTimeline(src)
	case kindMindmap:
		return renderMindmap(src)
	case kindQuadrant:
		return renderQuadrant(src, limit, g)
	case kindXYChart:
		return renderXYChart(src, limit, g)
	case kindGitGraph:
		return renderGitGraph(src, limit, g)
	}

	d, err := Parse(src)
	if err != nil {
		return nil, err
	}
	return drawFlowchartFit(d, limit, g)
}

// drawFlowchartFit renders a flowchart-shaped diagram, compacting it when
// it does not fit the limit.
func drawFlowchartFit(d *diagram, limit int, g glyphSet) ([]string, error) {
	if len(d.nodes) > maxNodes {
		return nil, errf("diagram too large (over %d nodes)", maxNodes)
	}
	if len(d.edges) > maxEdges {
		return nil, errf("diagram too large (over %d edges)", maxEdges)
	}

	lines, width, err := draw(d, defaultSettings, g)
	if err != nil {
		return nil, err
	}
	if limit > 0 && width > limit {
		lines, width, err = draw(d, compactSettings, g)
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
