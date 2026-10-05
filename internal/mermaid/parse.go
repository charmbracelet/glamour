package mermaid

import "strings"

// Direction is the layout direction of a flowchart.
type Direction int

// Layout directions.
const (
	// DirTD lays the graph out top to bottom.
	DirTD Direction = iota
	// DirLR lays the graph out left to right.
	DirLR
)

// Shape describes the visual form used to draw a node.
type Shape int

// Node shapes.
const (
	ShapeRect Shape = iota
	ShapeRounded
	ShapeStadium
	ShapeCircle
	ShapeDiamond
	ShapeHexagon
)

// edgeKind describes how an edge is drawn.
type edgeKind int

// Edge kinds.
const (
	edgeArrow edgeKind = iota
	edgeOpen
	edgeDottedArrow
	edgeDottedOpen
	edgeThickArrow
	edgeThickOpen
)

// hasArrowhead reports whether an edge kind ends in an arrowhead.
func hasArrowhead(kind edgeKind) bool {
	switch kind {
	case edgeArrow, edgeDottedArrow, edgeThickArrow:
		return true
	case edgeOpen, edgeDottedOpen, edgeThickOpen:
		return false
	}
	return false
}

// lineStyleFor maps an edge kind to its line style.
func lineStyleFor(kind edgeKind) lineStyle {
	switch kind {
	case edgeArrow, edgeOpen:
		return lineSolid
	case edgeDottedArrow, edgeDottedOpen:
		return lineDotted
	case edgeThickArrow, edgeThickOpen:
		return lineThick
	}
	return lineSolid
}

type node struct {
	id    string
	label string
	shape Shape
}

type edge struct {
	from  *node
	to    *node
	label string
	kind  edgeKind
}

type diagram struct {
	direction     Direction
	title         string
	nodes         []*node
	nodeByID      map[string]*node
	edges         []*edge
	haveDirection bool
}

// otherDiagramTypes are mermaid diagram types this package does not render;
// they are reported explicitly so users get a precise reason.
var otherDiagramTypes = []string{
	"sequenceDiagram", "classDiagram", "stateDiagram", "erDiagram",
	"journey", "gantt", "pie", "gitGraph", "mindmap", "timeline",
	"quadrantChart", "sankey", "C4Context", "architecture",
}

// unsupportedKeywords are flowchart statement keywords this package does not
// support; diagrams using them fall back to their source.
var unsupportedKeywords = []string{
	"subgraph", "end", "classdef", "class", "style", "linkstyle",
	"click", "direction", "acctitle", "accdescr", "init",
}

// unsupportedEdgeOps are edge constructs this package does not route.
var unsupportedEdgeOps = []string{"<--", "--o", "--x", "o--", "x--", "&"}

// Parse parses a mermaid flowchart or graph definition.
func Parse(src string) (*diagram, error) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src, title := stripFrontmatter(src)
	d := &diagram{
		direction: DirTD,
		title:     title,
		nodeByID:  make(map[string]*node),
	}

	lineno := 0
	for _, line := range strings.Split(src, "\n") {
		lineno++
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		for _, stmt := range strings.Split(trimmed, ";") {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if err := d.parseStatement(stmt, lineno); err != nil {
				return nil, err
			}
		}
	}
	if !d.haveDirection {
		return nil, errf("empty diagram")
	}
	if len(d.nodes) == 0 {
		return nil, errf("empty diagram")
	}
	return d, nil
}

func (d *diagram) parseStatement(stmt string, lineno int) error {
	if !d.haveDirection {
		return d.parseDirection(stmt, lineno)
	}

	word := strings.ToLower(firstWord(stmt))
	for _, keyword := range unsupportedKeywords {
		if word == keyword {
			return errf("unsupported syntax %q", keyword)
		}
	}
	for _, op := range unsupportedEdgeOps {
		if strings.Contains(stmt, op) {
			return errf("unsupported syntax %q", op)
		}
	}
	return d.parseChain(stmt, lineno)
}

func (d *diagram) parseDirection(stmt string, lineno int) error {
	fields := strings.Fields(stmt)
	keyword := strings.ToLower(fields[0])
	if keyword != "flowchart" && keyword != "graph" {
		lower := strings.ToLower(stmt)
		for _, t := range otherDiagramTypes {
			if strings.HasPrefix(lower, strings.ToLower(t)) {
				return errf("unsupported diagram type %q", t)
			}
		}
		return errf(`could not parse (line %d): expected "flowchart" or "graph"`, lineno)
	}
	if len(fields) > 2 {
		return errf("could not parse (line %d): unexpected %q", lineno, fields[2])
	}
	dir := "TD"
	if len(fields) > 1 {
		dir = strings.ToUpper(fields[1])
	}
	switch dir {
	case "TD", "TB", "":
		d.direction = DirTD
	case "LR":
		d.direction = DirLR
	case "BT", "RL":
		return errf("unsupported direction %q", dir)
	default:
		return errf("could not parse (line %d): unknown direction %q", lineno, dir)
	}
	d.haveDirection = true
	return nil
}

func (d *diagram) addNode(n *node) {
	if existing, ok := d.nodeByID[n.id]; ok {
		if n.label != n.id {
			existing.label = n.label
			existing.shape = n.shape
		}
		return
	}
	d.nodeByID[n.id] = n
	d.nodes = append(d.nodes, n)
}

// chainParser parses a statement of nodes joined by edge operators.
type chainParser struct {
	d      *diagram
	s      string
	i      int
	lineno int
}

func (d *diagram) parseChain(stmt string, lineno int) error {
	p := &chainParser{d: d, s: stmt, lineno: lineno}
	n, err := p.parseNode()
	if err != nil {
		return err
	}
	d.addNode(n)
	cur := n
	for {
		kind, label, ok, err := p.parseEdge()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		n, err := p.parseNode()
		if err != nil {
			return err
		}
		d.addNode(n)
		d.edges = append(d.edges, &edge{from: cur, to: n, label: label, kind: kind})
		cur = n
	}
	if p.i != len(p.s) {
		return errf("could not parse (line %d): unexpected %q", lineno, strings.TrimSpace(p.s[p.i:]))
	}
	return nil
}

func (p *chainParser) skipSpace() {
	for p.i < len(p.s) && (p.s[p.i] == ' ' || p.s[p.i] == '\t') {
		p.i++
	}
}

func (p *chainParser) at(prefix string) bool {
	return strings.HasPrefix(p.s[p.i:], prefix)
}

// parseNode parses a node reference: an identifier with an optional shape
// and label. The label defaults to the identifier.
func (p *chainParser) parseNode() (*node, error) {
	p.skipSpace()
	if p.i >= len(p.s) {
		return nil, errf("could not parse (line %d): expected a node", p.lineno)
	}
	start := p.i
	for p.i < len(p.s) && isIdentByte(p.s[p.i], p.s[p.i:]) {
		p.i++
	}
	id := p.s[start:p.i]
	if id == "" {
		return nil, errf("could not parse (line %d): expected a node id", p.lineno)
	}
	p.skipSpace()

	n := &node{id: id, label: id}
	var (
		label string
		shape Shape
		err   error
	)
	switch {
	case p.at("(["):
		label, err = p.parseLabel("([", "])")
		shape = ShapeStadium
	case p.at("(("):
		label, err = p.parseLabel("((", "))")
		shape = ShapeCircle
	case p.at("("):
		label, err = p.parseLabel("(", ")")
		shape = ShapeRounded
	case p.at("{{"):
		label, err = p.parseLabel("{{", "}}")
		shape = ShapeHexagon
	case p.at("{"):
		label, err = p.parseLabel("{", "}")
		shape = ShapeDiamond
	case p.at("[/"):
		label, err = p.parseLabel("[/", "/]")
		shape = ShapeRect
	case p.at("[\\"):
		label, err = p.parseLabel("[\\", "\\]")
		shape = ShapeRect
	case p.at("["):
		label, err = p.parseLabel("[", "]")
		shape = ShapeRect
	case p.at(">"), p.at("[[["), p.at("((("):
		err = errf("unsupported node shape (line %d)", p.lineno)
	}
	if err != nil {
		return nil, err
	}
	if label != "" {
		n.label = label
		n.shape = shape
	}
	return n, nil
}

// isIdentByte reports whether c continues a node identifier. rest is the
// remainder of the statement starting at c, used to tell identifier hyphens
// apart from edge operators.
func isIdentByte(c byte, rest string) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case c == '_' || c == '.' || c >= 0x80:
		return true
	case c == '-':
		// A hyphen belongs to the id unless it starts an edge operator.
		if len(rest) > 1 && isEdgeByte(rest[1]) {
			return false
		}
		return true
	}
	return false
}

// isEdgeByte reports whether c continues an edge operator.
func isEdgeByte(c byte) bool {
	return c == '-' || c == '=' || c == '.' || c == '>'
}

// parseEdge parses one edge operator, returning ok=false at the end of the
// statement.
func (p *chainParser) parseEdge() (kind edgeKind, label string, ok bool, err error) {
	p.skipSpace()
	if p.i >= len(p.s) {
		return 0, "", false, nil
	}
	switch {
	case p.at("-."):
		return p.parseDottedOp()
	case p.at("=="):
		return p.parseThickOp()
	case p.at("--"):
		return p.parseSolidOp()
	}
	return 0, "", false, errf("could not parse (line %d): unexpected %q", p.lineno, firstWord(p.s[p.i:]))
}

func (p *chainParser) parseSolidOp() (kind edgeKind, label string, ok bool, err error) {
	p.i += 2
	for p.i < len(p.s) && p.s[p.i] == '-' {
		p.i++
	}
	return p.finishOp(edgeArrow, edgeOpen, "-->", "---")
}

func (p *chainParser) parseDottedOp() (kind edgeKind, label string, ok bool, err error) {
	p.i += 2
	for p.i < len(p.s) && p.s[p.i] == '-' {
		p.i++
	}
	return p.finishOp(edgeDottedArrow, edgeDottedOpen, ".->", ".-")
}

func (p *chainParser) parseThickOp() (kind edgeKind, label string, ok bool, err error) {
	p.i += 2
	for p.i < len(p.s) && p.s[p.i] == '=' {
		p.i++
	}
	return p.finishOp(edgeThickArrow, edgeThickOpen, "==>", "===")
}

// finishOp handles both the inline (-->|label|) and spaced (-- label -->)
// label forms after the operator's dashes have been consumed.
func (p *chainParser) finishOp(arrow, open edgeKind, closeArrow, closeOpen string) (kind edgeKind, label string, ok bool, err error) {
	if p.i < len(p.s) && p.s[p.i] == ' ' && p.hasSpacedLabel(closeArrow, closeOpen) {
		return p.parseSpacedLabel(arrow, open, closeArrow, closeOpen)
	}
	kind = open
	if p.i < len(p.s) && p.s[p.i] == '>' {
		kind = arrow
		p.i++
	}
	label, err = p.parsePipeLabel()
	if err != nil {
		return 0, "", false, err
	}
	return kind, label, true, nil
}

// hasSpacedLabel reports whether the remainder of the statement closes a
// spaced label form, telling it apart from a plain operator followed by
// whitespace.
func (p *chainParser) hasSpacedLabel(closeArrow, closeOpen string) bool {
	rest := p.s[p.i:]
	return strings.Contains(rest, closeArrow) || strings.Contains(rest, closeOpen)
}

// parseSpacedLabel parses the "-- text -->" label form, where the label runs
// from here to the closing operator. The arrow closer wins when both
// closers match at the same position.
func (p *chainParser) parseSpacedLabel(arrow, open edgeKind, closeArrow, closeOpen string) (kind edgeKind, label string, ok bool, err error) {
	start := p.i
	kind, end, advance := open, -1, 0
	if idx := strings.Index(p.s[start:], closeArrow); idx >= 0 {
		kind, end, advance = arrow, start+idx, len(closeArrow)
	}
	if idx := strings.Index(p.s[start:], closeOpen); idx >= 0 && (end < 0 || start+idx < end) {
		kind, end, advance = open, start+idx, len(closeOpen)
	}
	if end < 0 {
		return 0, "", false, errf("could not parse (line %d): unterminated edge label", p.lineno)
	}
	p.i = end + advance
	label = strings.TrimSpace(p.s[start:end])
	return kind, unquoteLabel(label), true, nil
}

// parsePipeLabel parses the optional |label| suffix of an edge operator.
func (p *chainParser) parsePipeLabel() (string, error) {
	p.skipSpace()
	if p.i >= len(p.s) || p.s[p.i] != '|' {
		return "", nil
	}
	p.i++
	start := p.i
	inQuote := false
	for p.i < len(p.s) {
		c := p.s[p.i]
		if c == '"' {
			inQuote = !inQuote
			p.i++
			continue
		}
		if !inQuote && c == '|' {
			label := strings.TrimSpace(p.s[start:p.i])
			p.i++
			return unquoteLabel(label), nil
		}
		p.i++
	}
	return "", errf("could not parse (line %d): unterminated edge label", p.lineno)
}

// parseLabel parses a node label delimited by open and close, honouring
// quoted sections that may contain the delimiters.
func (p *chainParser) parseLabel(open, close string) (string, error) {
	p.i += len(open)
	start := p.i
	inQuote := false
	for p.i < len(p.s) {
		c := p.s[p.i]
		switch {
		case inQuote && c == '\\' && p.i+1 < len(p.s):
			p.i += 2
			continue
		case c == '"':
			inQuote = !inQuote
			p.i++
			continue
		}
		if !inQuote && p.at(close) {
			s := p.s[start:p.i]
			p.i += len(close)
			return unquoteLabel(s), nil
		}
		p.i++
	}
	return "", errf("could not parse (line %d): unterminated %q node label", p.lineno, open)
}

// brReplacer turns mermaid line breaks into newlines.
var brReplacer = strings.NewReplacer(
	"<br/>", "\n",
	"<br>", "\n",
	"<br />", "\n",
)

// unquoteLabel strips an optional surrounding pair of quotes or backticks
// and resolves mermaid line breaks.
func unquoteLabel(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = strings.TrimSuffix(strings.TrimPrefix(s, `"`), `"`)
		s = strings.ReplaceAll(s, `\"`, `"`)
		s = strings.ReplaceAll(s, `#quot;`, `"`)
	} else if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		s = s[1 : len(s)-1]
	}
	return brReplacer.Replace(s)
}

// stripFrontmatter removes a YAML frontmatter block, returning the diagram
// body and its title.
func stripFrontmatter(src string) (string, string) {
	s := strings.TrimSpace(src)
	rest, ok := strings.CutPrefix(s, "---")
	if !ok || (rest != "" && !strings.HasPrefix(rest, "\n")) {
		return src, ""
	}
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return src, ""
	}
	var title string
	for _, line := range strings.Split(rest[:end], "\n") {
		if t, found := strings.CutPrefix(strings.TrimSpace(line), "title:"); found {
			title = strings.Trim(strings.TrimSpace(t), `"'`)
		}
	}
	return rest[end+4:], title
}

func firstWord(s string) string {
	if idx := strings.IndexAny(s, " \t"); idx >= 0 {
		return s[:idx]
	}
	return s
}
