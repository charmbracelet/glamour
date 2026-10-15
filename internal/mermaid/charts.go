package mermaid

import (
	"strconv"
	"strings"
)

// renderPie renders a pie chart as proportional bars with percentages.
func renderPie(src string, limit int, g glyphSet) ([]string, error) {
	title, entries, err := parsePie(src)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errf("empty diagram")
	}

	labelCol := 4
	for _, e := range entries {
		labelCol = max(labelCol, min(stringWidth(e.label), 20))
	}
	barW := 30
	if limit > 0 {
		barW = limit - labelCol - 8
	}
	if barW < 4 {
		barW = 4
	}

	total := 0.0
	for _, e := range entries {
		total += e.value
	}
	if total <= 0 {
		return nil, errf("could not parse: slice values sum to zero")
	}

	var lines []string
	for _, e := range entries {
		pct := e.value / total * 100
		n := int(pct/100*float64(barW) + 0.5)
		n = min(max(n, 1), barW)
		label := padRight(e.label, labelCol)
		lines = append(lines, label+" "+strings.Repeat(string(g.bar), n)+" "+strconv.FormatFloat(pct, 'f', 0, 64)+"%")
	}
	if title != "" {
		lines = append([]string{title, ""}, lines...)
	}
	return lines, nil
}

type pieEntry struct {
	label string
	value float64
}

// pieInlineTitle extracts the title from an inline "pie Show buckets"
// header, keeping an explicit title directive or frontmatter title.
func pieInlineTitle(trimmed, title string) string {
	rest := strings.TrimSpace(trimmed[len("pie"):])
	if rest == "" || title != "" {
		return title
	}
	if strings.HasPrefix(strings.ToLower(rest), kwTitle) {
		rest = strings.TrimSpace(rest[len(kwTitle):])
	}
	return strings.Trim(rest, `"`)
}

func parsePie(src string) (string, []pieEntry, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)
	var entries []pieEntry

	seenHeader := false
	lineno := 0
	for _, line := range strings.Split(src, "\n") {
		lineno++
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		if !seenHeader {
			if !strings.EqualFold(firstWord(trimmed), "pie") {
				return "", nil, errf(`could not parse (line %d): expected "pie"`, lineno)
			}
			seenHeader = true
			title = pieInlineTitle(trimmed, title)
			continue
		}
		word := strings.ToLower(firstWord(trimmed))
		rest := strings.TrimSpace(trimmed[len(word):])
		switch word {
		case kwTitle:
			if title == "" {
				title = strings.Trim(rest, `"`)
			}
		case "showdata":
		default:
			label, value, ok := strings.Cut(trimmed, ":")
			if !ok {
				return "", nil, errf("could not parse (line %d): %q", lineno, trimmed)
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil {
				return "", nil, errf("could not parse (line %d): %q", lineno, trimmed)
			}
			entries = append(entries, pieEntry{label: strings.Trim(strings.TrimSpace(label), `"`), value: v})
		}
	}
	return title, entries, nil
}

// renderJourney renders a user journey with score bars.
func renderJourney(src string, limit int, g glyphSet) ([]string, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)

	type task struct {
		name   string
		score  float64
		actors string
	}
	type section struct {
		name  string
		tasks []task
	}
	var sections []section
	current := -1

	seenHeader := false
	lineno := 0
	for _, line := range strings.Split(src, "\n") {
		lineno++
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		word := strings.ToLower(firstWord(trimmed))
		rest := strings.TrimSpace(trimmed[len(word):])
		if !seenHeader {
			if word == kwJourney {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "journey"`, lineno)
		}
		switch word {
		case kwTitle:
			if title == "" {
				title = rest
			}
		case "section":
			sections = append(sections, section{name: rest})
			current = len(sections) - 1
		default:
			if current < 0 {
				return nil, errf("could not parse (line %d): task outside a section", lineno)
			}
			parts := strings.Split(trimmed, ":")
			if len(parts) < 3 {
				return nil, errf("could not parse (line %d): task needs name, score and actors", lineno)
			}
			score, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return nil, errf("could not parse (line %d): bad score", lineno)
			}
			sections[current].tasks = append(sections[current].tasks, task{
				name:   strings.TrimSpace(parts[0]),
				score:  score,
				actors: strings.TrimSpace(parts[len(parts)-1]),
			})
		}
	}

	if !seenHeader || len(sections) == 0 {
		return nil, errf("empty diagram")
	}

	barW := 10
	if limit > 0 {
		barW = min(10, max(4, limit/8))
	}
	var lines []string
	for _, sec := range sections {
		lines = append(lines, sec.name)
		for _, task := range sec.tasks {
			n := int(task.score / 10 * float64(barW))
			n = min(max(n, 0), barW)
			bar := strings.Repeat(string(g.bar), n) + strings.Repeat(" ", barW-n)
			score := strconv.FormatFloat(task.score, 'f', -1, 64)
			pad := 20
			if limit > 0 {
				pad = min(20, max(8, limit/3))
			}
			lines = append(lines, "  "+padRight(truncateLabel(task.name, pad), pad)+" "+bar+" "+score+"  "+task.actors)
		}
		lines = append(lines, "")
	}
	lines = lines[:len(lines)-1]
	if title != "" {
		lines = append([]string{title, ""}, lines...)
	}
	return lines, nil
}

// renderTimeline renders a timeline as periods with bulleted events.
func renderTimeline(src string) ([]string, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)

	type period struct {
		name   string
		events []string
	}
	var periods []period

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
			if word == kwTimeline {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "timeline"`, lineno)
		}
		if word == kwTitle {
			if title == "" {
				title = strings.TrimSpace(trimmed[len(kwTitle):])
			}
			continue
		}
		head, tail, ok := strings.Cut(trimmed, ":")
		if !ok {
			return nil, errf("could not parse (line %d): %q", lineno, trimmed)
		}
		periods = append(periods, period{name: strings.TrimSpace(head)})
		idx := len(periods) - 1
		for _, event := range strings.Split(tail, ":") {
			if event = strings.TrimSpace(event); event != "" {
				periods[idx].events = append(periods[idx].events, event)
			}
		}
	}

	if !seenHeader || len(periods) == 0 {
		return nil, errf("empty diagram")
	}

	var lines []string
	for _, p := range periods {
		lines = append(lines, p.name)
		for _, event := range p.events {
			lines = append(lines, "  • "+event)
		}
	}
	if title != "" {
		lines = append([]string{title, ""}, lines...)
	}
	return lines, nil
}

// renderXYChart renders an xychart as horizontal bars, one row per x
// value. Line series are drawn as bars too.
func renderXYChart(src string, limit int, g glyphSet) ([]string, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)

	var labels []string
	var series [][]float64
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
			if word == "xychart-beta" || word == "xychart" {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "xychart-beta"`, lineno)
		}
		switch word {
		case kwTitle:
			if title == "" {
				title = strings.TrimSpace(trimmed[len(kwTitle):])
			}
		case "x-axis", "xaxis":
			labels = parseAxisLabels(trimmed)
		case "y-axis", "yaxis":
			// Scale is taken from the data itself.
		case "bar", "line":
			rest := strings.TrimSpace(trimmed[len(word):])
			values, err := parseFloatList(rest)
			if err != nil {
				return nil, errf("could not parse (line %d): %v", lineno, err)
			}
			series = append(series, values)
		default:
			return nil, errf("could not parse (line %d): %q", lineno, trimmed)
		}
	}

	if !seenHeader || len(series) == 0 {
		return nil, errf("empty diagram")
	}
	rows := len(series[0])
	for _, s := range series {
		if len(s) != rows {
			return nil, errf("could not parse: series lengths differ")
		}
	}

	maxV := 0.0
	for _, s := range series {
		for _, v := range s {
			maxV = max(maxV, v)
		}
	}

	labelCol := 4
	for i := range rows {
		label := strconv.Itoa(i + 1)
		if i < len(labels) {
			label = labels[i]
		}
		labelCol = max(labelCol, min(stringWidth(label), 16))
	}
	barW := 30
	if limit > 0 {
		barW = limit - labelCol - 10
	}
	barW = max(barW, 4)

	var lines []string
	for _, s := range series {
		for i, v := range s {
			label := strconv.Itoa(i + 1)
			if i < len(labels) {
				label = labels[i]
			}
			n := 0
			if maxV > 0 {
				n = int(v / maxV * float64(barW))
			}
			n = min(max(n, 0), barW)
			bar := strings.Repeat(string(g.bar), n) + strings.Repeat(" ", barW-n)
			lines = append(lines, padRight(truncateLabel(label, labelCol), labelCol)+" "+bar+" "+strconv.FormatFloat(v, 'f', -1, 64))
		}
		lines = append(lines, "")
	}
	lines = lines[:len(lines)-1]
	if title != "" {
		lines = append([]string{title, ""}, lines...)
	}
	return lines, nil
}

// parseAxisLabels extracts the bracketed labels of an axis declaration.
func parseAxisLabels(line string) []string {
	open := strings.IndexByte(line, '[')
	if open < 0 {
		return nil
	}
	closeIdx := strings.LastIndexByte(line, ']')
	if closeIdx <= open {
		return nil
	}
	var labels []string
	for _, label := range strings.Split(line[open+1:closeIdx], ",") {
		labels = append(labels, strings.Trim(strings.TrimSpace(label), `"`))
	}
	return labels
}

// parseFloatList parses a bracketed, comma-separated number list.
func parseFloatList(s string) ([]float64, error) {
	open := strings.IndexByte(s, '[')
	if open < 0 {
		return nil, errf("expected [values]")
	}
	closeIdx := strings.LastIndexByte(s, ']')
	if closeIdx <= open {
		return nil, errf("expected [values]")
	}
	var values []float64
	for _, field := range strings.Split(s[open+1:closeIdx], ",") {
		v, err := strconv.ParseFloat(strings.TrimSpace(field), 64)
		if err != nil {
			return nil, errf("bad value %q", field)
		}
		values = append(values, v)
	}
	return values, nil
}

func padRight(s string, w int) string {
	if pad := w - stringWidth(s); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}
	return s
}

// truncateLabel shortens a label to at most w columns.
func truncateLabel(s string, w int) string {
	if stringWidth(s) <= w {
		return s
	}
	out := make([]rune, 0, w)
	width := 0
	for _, r := range s {
		rw := runeWidth(r)
		if width+rw > w-1 {
			break
		}
		out = append(out, r)
		width += rw
	}
	return string(out) + "…"
}
