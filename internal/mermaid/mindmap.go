package mermaid

import "strings"

// renderMindmap renders an indented mindmap as a tree.
func renderMindmap(src string) ([]string, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)

	type node struct {
		text     string
		depth    int
		children []*node
	}
	var root *node
	var rootDepth int

	seenHeader := false
	lineno := 0
	for _, line := range strings.Split(src, "\n") {
		lineno++
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "%%") {
			continue
		}
		if !seenHeader {
			if strings.EqualFold(firstWord(strings.TrimSpace(line)), "mindmap") {
				seenHeader = true
				continue
			}
			return nil, errf(`could not parse (line %d): expected "mindmap"`, lineno)
		}

		indent := 0
		for indent < len(line) && line[indent] == ' ' {
			indent++
		}
		if strings.HasPrefix(line[indent:], "\t") {
			return nil, errf("could not parse (line %d): tabs are not supported", lineno)
		}
		depth := indent / 2
		text := stripMindmapShape(strings.TrimSpace(line))
		if text == "" {
			return nil, errf("could not parse (line %d): empty node", lineno)
		}

		if root == nil {
			rootDepth = depth
			depth = 0
		} else {
			depth -= rootDepth
			if depth < 1 {
				return nil, errf("could not parse (line %d): second root", lineno)
			}
		}

		n := &node{text: text, depth: depth}
		if n.depth == 0 {
			root = n
			continue
		}

		parent := root
		for parent.depth != n.depth-1 {
			if len(parent.children) == 0 {
				return nil, errf("could not parse (line %d): indentation skip", lineno)
			}
			parent = parent.children[len(parent.children)-1]
		}
		parent.children = append(parent.children, n)
	}

	if !seenHeader || root == nil {
		return nil, errf("empty diagram")
	}

	lines := []string{stripMindmapShape(root.text)}
	var walk func(children []*node, prefix string)
	walk = func(children []*node, prefix string) {
		for i, child := range children {
			last := i == len(children)-1
			branch, extension := "├─ ", "│  "
			if last {
				branch, extension = "└─ ", "   "
			}
			lines = append(lines, prefix+branch+child.text)
			walk(child.children, prefix+extension)
		}
	}
	walk(root.children, "")
	if title != "" {
		lines = append([]string{title, ""}, lines...)
	}
	return lines, nil
}

// stripMindmapShape removes a trailing node shape, keeping the inner text
// it displays. Lines without a shape show their id as the text.
func stripMindmapShape(s string) string {
	openers := []struct {
		open, close string
	}{
		{"((", "))"},
		{"[[", "]]"},
		{"(", ")"},
		{"[", "]"},
		{">", "]"},
	}
	for _, pair := range openers {
		open := strings.Index(s, pair.open)
		if open <= 0 {
			continue
		}
		inner := s[open+len(pair.open) : len(s)-len(pair.close)]
		if strings.HasSuffix(s, pair.close) && strings.TrimSpace(inner) != "" && strings.TrimSpace(s[:open]) != "" {
			return strings.TrimSpace(inner)
		}
	}
	return s
}
