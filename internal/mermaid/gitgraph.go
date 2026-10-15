package mermaid

import "strings"

// gitGraph guards.
const maxGitEvents = 100

type gitCommit struct {
	lane int
	col  int
}

type gitBranch struct {
	name string
	lane int
	from int
	col  int // column of the commit it forked from
}

type gitMerge struct {
	name    string
	from    int
	to      int
	fromCol int
	toCol   int
}

type gitGraph struct {
	title    string
	lanes    []string
	commits  []gitCommit
	branches []gitBranch
	merges   []gitMerge
}

// renderGitGraph renders a git graph as branch lanes with commit dots.
func renderGitGraph(src string, limit int, g glyphSet) ([]string, error) {
	d, err := parseGitGraph(src)
	if err != nil {
		return nil, err
	}

	laneCol := 4
	for _, lane := range d.lanes {
		laneCol = max(laneCol, stringWidth(lane))
	}
	cols := 0
	for _, c := range d.commits {
		cols = max(cols, c.col+1)
	}
	for _, m := range d.merges {
		cols = max(cols, m.toCol+1)
	}
	if cols == 0 {
		return nil, errf("empty diagram")
	}
	if limit > 0 && laneCol+1+cols > limit {
		return nil, errf("too wide to render (needs %d columns, %d available)", laneCol+1+cols, limit)
	}

	// Cells per lane: the lane line, commit dots, and the branch and
	// merge connectors.
	cells := make([][]rune, len(d.lanes))
	commitAt := make([]map[int]bool, len(d.lanes))
	for i := range d.lanes {
		cells[i] = []rune(strings.Repeat(" ", cols))
		commitAt[i] = make(map[int]bool)
	}
	for _, c := range d.commits {
		commitAt[c.lane][c.col] = true
		cells[c.lane][c.col] = g.circle
	}

	// Lane extents: from the fork column (or 0) to the last event.
	start := make([]int, len(d.lanes))
	end := make([]int, len(d.lanes))
	for i := range d.lanes {
		end[i] = -1
	}
	for _, b := range d.branches {
		start[b.lane] = b.col + 1
	}
	for _, c := range d.commits {
		end[c.lane] = max(end[c.lane], c.col)
	}
	for _, m := range d.merges {
		end[m.to] = max(end[m.to], m.toCol)
		if end[m.from] < m.toCol {
			end[m.from] = m.toCol
		}
	}

	for i := range d.lanes {
		for x := start[i]; x <= end[i] && x < cols; x++ {
			if !commitAt[i][x] && cells[i][x] == ' ' {
				cells[i][x] = g.h
			}
		}
	}
	// Fork connectors: a corner on the new lane under the forked commit.
	for _, b := range d.branches {
		if b.col < cols {
			cells[b.lane][b.col] = g.bl
		}
		for y := min(b.from, b.lane) + 1; y < max(b.from, b.lane); y++ {
			cells[y][b.col] = g.dotV
		}
		if from := b.from; from < len(d.lanes) && commitAt[from][b.col] {
			cells[from][b.col] = g.circle
		}
	}
	// Merge connectors: dotted runs and a corner turning up into the
	// target lane.
	for _, m := range d.merges {
		for x := m.fromCol + 1; x < m.toCol && x < cols; x++ {
			cells[m.from][x] = g.dotH
		}
		if m.toCol < cols {
			cells[m.from][m.toCol] = g.br
		}
		for y := min(m.from, m.to) + 1; y < max(m.from, m.to); y++ {
			cells[y][m.toCol] = g.dotV
		}
		if commitAt[m.to][m.toCol] {
			cells[m.to][m.toCol] = g.circle
		}
	}

	var lines []string
	for i, lane := range d.lanes {
		lines = append(lines, padRight(lane, laneCol)+" "+strings.TrimRight(string(cells[i]), " "))
	}
	if d.title != "" {
		lines = append([]string{d.title, ""}, lines...)
	}
	return lines, nil
}

// parseGitGraph parses gitGraph commands: commit, branch, checkout and
// merge.
func parseGitGraph(src string) (*gitGraph, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)
	d := &gitGraph{title: title, lanes: []string{"main"}}

	laneOf := map[string]int{"main": 0}
	current, col := 0, -1

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
			if word == "gitgraph" {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "gitGraph"`, lineno)
		}

		switch word {
		case "commit":
			col++
			d.commits = append(d.commits, gitCommit{lane: current, col: col})
		case "branch":
			name := gitArg(trimmed, "branch")
			if name == "" {
				return nil, errf("could not parse (line %d): branch needs a name", lineno)
			}
			if _, ok := laneOf[name]; ok {
				return nil, errf("could not parse (line %d): duplicate branch %q", lineno, name)
			}
			lane := len(d.lanes)
			d.lanes = append(d.lanes, name)
			laneOf[name] = lane
			d.branches = append(d.branches, gitBranch{name: name, lane: lane, from: current, col: col})
			current = lane
		case "checkout":
			name := gitArg(trimmed, "checkout")
			lane, ok := laneOf[name]
			if !ok {
				return nil, errf("could not parse (line %d): unknown branch %q", lineno, name)
			}
			current = lane
		case "merge":
			name := gitArg(trimmed, "merge")
			from, ok := laneOf[name]
			if !ok {
				return nil, errf("could not parse (line %d): unknown branch %q", lineno, name)
			}
			fromCol := -1
			for i := len(d.commits) - 1; i >= 0; i-- {
				if d.commits[i].lane == from {
					fromCol = d.commits[i].col
					break
				}
			}
			if fromCol < 0 {
				return nil, errf("could not parse (line %d): %q has no commits", lineno, name)
			}
			col++
			d.merges = append(d.merges, gitMerge{name: name, from: from, to: current, fromCol: fromCol, toCol: col})
			d.commits = append(d.commits, gitCommit{lane: current, col: col})
		default:
			return nil, errf("unsupported syntax (line %d): %q", lineno, word)
		}
	}

	if !seenHeader || len(d.commits) > maxGitEvents {
		if len(d.commits) > maxGitEvents {
			return nil, errf("diagram too large (over %d commits)", maxGitEvents)
		}
		return nil, errf("empty diagram")
	}
	return d, nil
}

// gitArg extracts the branch name argument of a command, ignoring id, tag
// and other qualifiers.
func gitArg(line, word string) string {
	rest := strings.TrimSpace(line[len(word):])
	rest = strings.TrimSpace(rest)
	if quoted := strings.IndexByte(rest, '"'); quoted >= 0 {
		rest = rest[:quoted]
	}
	return firstWord(rest)
}
