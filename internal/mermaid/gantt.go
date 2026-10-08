package mermaid

import (
	"strconv"
	"strings"
	"time"
)

// Gantt guards and parsing helpers.
const (
	maxGanttTasks = 100
	dayDur        = 24 * time.Hour
)

type ganttTask struct {
	name      string
	alias     string
	section   string
	done      bool
	active    bool
	crit      bool
	milestone bool
	after     string
	duration  time.Duration
	start     time.Time
	end       time.Time
	hasStart  bool
	resolved  bool
}

type ganttDiagram struct {
	title string
	tasks []*ganttTask
}

// ganttDateLayouts are the date formats tried when reading task dates.
var ganttDateLayouts = []string{
	"2006-01-02",
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"02 Jan 2006",
	"Jan 2, 2006",
	"02/01/2006",
}

func parseGanttDate(s string) (time.Time, bool) {
	for _, layout := range ganttDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseGanttDuration(s string) (time.Duration, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(s[:len(s)-1]))
	if err != nil || n < 0 {
		return 0, false
	}
	switch s[len(s)-1] {
	case 'd':
		return time.Duration(n) * dayDur, true
	case 'w':
		return time.Duration(n) * 7 * dayDur, true
	case 'h':
		return time.Duration(n) * time.Hour, true
	}
	return 0, false
}

// parseGantt parses a gantt chart definition.
func parseGantt(src string) (*ganttDiagram, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)
	d := &ganttDiagram{title: title}

	seenHeader, section := false, ""
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
			if word == kwGantt {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "gantt"`, lineno)
		}

		switch word {
		case kwTitle:
			if d.title == "" {
				d.title = rest
			}
		case "dateformat", "axisformat", "excludes", "includes", "todaymarker", kwClick:
			// Rendering directives; dates are parsed independently.
		case "section":
			section = rest
		default:
			task, err := parseGanttTask(trimmed, section, lineno)
			if err != nil {
				return nil, err
			}
			d.tasks = append(d.tasks, task)
		}
	}

	if !seenHeader || len(d.tasks) == 0 {
		return nil, errf("empty diagram")
	}
	if err := resolveGanttTasks(d); err != nil {
		return nil, err
	}
	return d, nil
}

// parseGanttTask parses one task line: name : alias, flags, start, end.
func parseGanttTask(s, section string, lineno int) (*ganttTask, error) {
	name, rest, ok := strings.Cut(s, ":")
	if !ok {
		return nil, errf("could not parse (line %d): task needs a colon", lineno)
	}
	task := &ganttTask{name: strings.TrimSpace(name), section: section}

	var dates []time.Time
	for _, field := range strings.Split(rest, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		switch strings.ToLower(field) {
		case "done":
			task.done = true
			continue
		case "active":
			task.active = true
			continue
		case "crit":
			task.crit = true
			continue
		case "milestone":
			task.milestone = true
			continue
		}
		if strings.HasPrefix(strings.ToLower(field), "after ") {
			task.after = strings.TrimSpace(field[len("after "):])
			continue
		}
		if t, isDate := parseGanttDate(field); isDate {
			dates = append(dates, t)
			continue
		}
		if dur, isDur := parseGanttDuration(field); isDur {
			task.duration = dur
			continue
		}
		if task.alias == "" {
			task.alias = field
			continue
		}
		return nil, errf("could not parse (line %d): unknown field %q", lineno, field)
	}

	if len(dates) > 0 {
		task.start = dates[0]
		task.hasStart = true
	}
	if len(dates) > 1 {
		task.end = dates[1]
	}
	return task, nil
}

// resolveGanttTasks fills in starts from after dependencies and ends from
// durations.
func resolveGanttTasks(d *ganttDiagram) error {
	byAlias := make(map[string]*ganttTask)
	for _, t := range d.tasks {
		key := t.alias
		if key == "" {
			key = t.name
		}
		if _, dup := byAlias[key]; dup {
			return errf("could not parse: duplicate task %q", key)
		}
		if _, dup := byAlias[strings.ToLower(key)]; dup && strings.ToLower(key) != key {
			return errf("could not parse: duplicate task %q", key)
		}
		byAlias[key] = t
		byAlias[strings.ToLower(key)] = t
	}

	for range d.tasks {
		for _, t := range d.tasks {
			if t.after != "" && !t.hasStart {
				dep, ok := byAlias[t.after]
				if !ok {
					return errf("could not parse: unknown task %q", t.after)
				}
				if dep.resolved {
					t.start = dep.end
					t.hasStart = true
				}
			}
			if t.hasStart && !t.resolved {
				if t.end.IsZero() {
					if t.milestone || t.duration == 0 {
						t.end = t.start
					} else {
						t.end = t.start.Add(t.duration)
					}
				}
				t.resolved = true
			}
		}
	}

	for _, t := range d.tasks {
		if !t.resolved {
			return errf("could not parse: task %q has no start date", t.name)
		}
	}
	return nil
}

// drawGanttFit renders a gantt chart with the time axis scaled to the
// limit.
func drawGanttFit(d *ganttDiagram, limit int, g glyphSet) ([]string, error) {
	if len(d.tasks) > maxGanttTasks {
		return nil, errf("diagram too large (over %d tasks)", maxGanttTasks)
	}

	labelCol := ganttLabelCol(d, 20)
	barArea := 60
	if limit > 0 {
		barArea = limit - labelCol - 3
	}
	lines, width := drawGantt(d, g, labelCol, barArea)
	if limit > 0 && width > limit {
		labelCol = ganttLabelCol(d, 12)
		barArea = limit - labelCol - 3
		if barArea < 8 {
			return nil, errf("too wide to render (needs %d columns, %d available)", labelCol+11, limit)
		}
		lines, _ = drawGantt(d, g, labelCol, barArea)
	}
	if d.title != "" {
		lines = append([]string{d.title, ""}, lines...)
	}
	return lines, nil
}

// ganttLabelCol returns the column budget for task names.
func ganttLabelCol(d *ganttDiagram, maxW int) int {
	w := 4
	for _, t := range d.tasks {
		for _, line := range wrapLabel(t.name, maxW) {
			w = max(w, stringWidth(line))
		}
		if t.section != "" {
			w = max(w, stringWidth(t.section))
		}
	}
	return min(w, maxW)
}

// drawGantt renders the chart as text rows with a time axis.
func drawGantt(d *ganttDiagram, g glyphSet, labelCol, barArea int) ([]string, int) {
	if barArea < 8 {
		barArea = 8
	}

	lo, hi := d.tasks[0].start, d.tasks[0].end
	for _, t := range d.tasks {
		lo = timeMin(lo, t.start)
		hi = timeMax(hi, t.end)
	}
	span := hi.Sub(lo)

	col := func(t time.Time) int {
		if span <= 0 {
			return 0
		}
		frac := float64(t.Sub(lo)) / float64(span)
		return min(barArea-1, int(frac*float64(barArea-1)+0.5))
	}

	// Time axis: date labels with a ruler underneath.
	dateFmt := "Jan 2"
	if lo.Year() != hi.Year() {
		dateFmt = "Jan 2 06"
	}
	ticks := max(2, barArea/10)
	labels := make([]rune, barArea)
	for x := range labels {
		labels[x] = ' '
	}
	ruler := make([]rune, barArea)
	for x := range ruler {
		ruler[x] = g.h
	}
	for i := 0; i < ticks; i++ {
		c := 0
		if ticks > 1 {
			c = i * (barArea - 1) / (ticks - 1)
		}
		ruler[c] = g.v
		label := []rune(lo.Add(time.Duration(float64(span) * float64(i) / float64(max(ticks-1, 1)))).Format(dateFmt))
		x := c - len(label)/2
		if x < 0 {
			x = 0
		}
		if x+len(label) > barArea {
			continue
		}
		for j, r := range label {
			if x+j < barArea {
				labels[x+j] = r
			}
		}
	}

	var lines []string
	blank := strings.Repeat(" ", labelCol+1)
	lines = append(lines,
		blank+string(labels),
		blank+string(ruler),
	)

	pad := func(s string) string {
		if w := stringWidth(s); w < labelCol {
			return s + strings.Repeat(" ", labelCol-w)
		}
		return s
	}

	section := ""
	for _, t := range d.tasks {
		if t.section != section {
			section = t.section
			if section != "" {
				lines = append(lines, "", section)
			}
		}

		nameLines := wrapLabel(t.name, labelCol)
		start, end := col(t.start), col(t.end)
		if end < start {
			start, end = end, start
		}
		if t.milestone {
			end = start
		}

		bar := make([]rune, barArea)
		for x := range bar {
			bar[x] = ' '
		}
		fill := g.bar
		switch {
		case t.done:
			fill = g.barDone
		case t.active:
			fill = g.barActive
		case t.crit:
			fill = g.barCrit
		}
		if t.milestone {
			bar[start] = g.milestone
		} else {
			for x := start; x <= end; x++ {
				bar[x] = fill
			}
		}

		for i, line := range nameLines {
			if i == 0 {
				lines = append(lines, pad(line)+" "+string(bar))
			} else {
				lines = append(lines, pad(line))
			}
		}
	}

	width := 0
	for _, line := range lines {
		width = max(width, stringWidth(line))
	}
	return lines, width
}

func timeMin(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func timeMax(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
