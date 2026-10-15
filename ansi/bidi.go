package ansi

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"golang.org/x/text/unicode/bidi"
)

// Most terminals print each line from left to right in the order its
// characters are stored, which leaves right-to-left scripts such as Arabic
// and Hebrew reversed. When [Options.BidiReordering] is set, each text flow
// (the text of a paragraph, heading or list item) is opened with FSI and
// closed with PDI as it's rendered. Once the flow has been word wrapped,
// reorderBidi puts each line's share of it into visual order following the
// Unicode Bidirectional Algorithm (UAX #9), and removes the marks.
const (
	bidiOpen  = "⁨" // FIRST STRONG ISOLATE
	bidiClose = "⁩" // POP DIRECTIONAL ISOLATE
)

// textFlow returns the Element for a text flow with no renderer of its own,
// such as the text of a list item.
func textFlow(ctx RenderContext) Element {
	if !ctx.options.BidiReordering {
		return Element{}
	}
	return Element{Entering: bidiOpen, Exiting: bidiClose}
}

// closeBidiFlow closes the text flow opened in s, or drops its opening mark
// if only whitespace was written after it.
func closeBidiFlow(s string) string {
	i := strings.LastIndex(s, bidiOpen)
	if i < 0 {
		return s
	}
	if strings.TrimSpace(s[i+len(bidiOpen):]) == "" {
		return s[:i] + s[i+len(bidiOpen):]
	}
	return s + bidiClose
}

// reorderBidi puts the text flows in s into visual order, line by line, and
// removes their marks. Anything outside of a flow, such as a list bullet, is
// left where it is.
func reorderBidi(s string) string {
	if !strings.Contains(s, bidiOpen) {
		return s
	}

	levels := flowLevels(s)
	var (
		b      strings.Builder
		st     sgrState // styling in effect
		inFlow bool
		level  int      // paragraph embedding level of the current flow
		line   []string // tokens of the current flow on this line
		pstate byte
	)
	b.Grow(len(s))
	flush := func() {
		st = reorderLine(&b, line, st, level)
		line = line[:0]
	}

	for len(s) > 0 {
		tok, _, n, newState := ansi.DecodeSequence(s, pstate, nil)
		pstate = newState
		s = s[n:]

		switch {
		case tok == bidiOpen:
			if inFlow {
				flush()
			}
			inFlow = true
			level, levels = levels[0], levels[1:]
		case tok == bidiClose:
			if inFlow {
				flush()
				inFlow = false
			}
		case inFlow && tok == "\n":
			flush()
			b.WriteString(tok)
		case inFlow:
			line = append(line, tok)
		default:
			if isEscape(tok) {
				st.apply(tok)
			}
			b.WriteString(tok)
		}
	}
	if inFlow {
		flush()
	}
	return b.String()
}

// flowLevels returns the paragraph embedding level of each text flow in s:
// 1 if its first strong character is right-to-left, 0 otherwise (rules P2
// and P3).
func flowLevels(s string) []int {
	var levels []int
	for {
		i := strings.Index(s, bidiOpen)
		if i < 0 {
			return levels
		}
		s = s[i+len(bidiOpen):]
		j := strings.IndexAny(s, bidiOpen+bidiClose)
		if j < 0 {
			j = len(s)
		}
		level := 0
	loop:
		for _, r := range ansi.Strip(s[:j]) {
			switch bidiClass(r) {
			case bidi.L:
				break loop
			case bidi.R, bidi.AL:
				level = 1
				break loop
			}
		}
		levels = append(levels, level)
		s = s[j:]
	}
}

// reorderLine writes one line of a text flow, given as a sequence of
// graphemes and escape sequences, to b in visual order. st is the styling in
// effect before the line; the styling in effect after it is returned.
func reorderLine(b *strings.Builder, tokens []string, st sgrState, level int) sgrState {
	type cell struct {
		text string
		st   sgrState
	}
	var cells []cell
	end := st
	supported := true
	rtl := level%2 == 1
	for _, tok := range tokens {
		if isEscape(tok) {
			supported = end.apply(tok) && supported
			continue
		}
		cells = append(cells, cell{tok, end})
		if !rtl {
			switch bidiClass(firstRune(tok)) {
			case bidi.R, bidi.AL, bidi.AN:
				rtl = true
			}
		}
	}
	if !supported || !rtl {
		// Nothing to reorder, or escape sequences we can't carry over.
		for _, tok := range tokens {
			b.WriteString(tok)
		}
		return end
	}

	// Leading and trailing whitespace stays put, so the line keeps its
	// alignment.
	isSpace := func(c cell) bool { return bidiClass(firstRune(c.text)) == bidi.WS }
	start := 0
	for start < len(cells) && isSpace(cells[start]) {
		start++
	}
	stop := len(cells)
	for stop > start && isSpace(cells[stop-1]) {
		stop--
	}

	chars := make([]rune, stop-start)
	for i, c := range cells[start:stop] {
		chars[i] = firstRune(c.text)
	}
	levels := bidiLevels(chars, level)

	cur := st
	write := func(c cell) {
		cur.transition(b, c.st)
		cur = c.st
		b.WriteString(c.text)
	}
	for _, c := range cells[:start] {
		write(c)
	}
	for _, i := range visualOrder(levels) {
		c := cells[start+i]
		if levels[i]%2 == 1 {
			// Rule L4: mirror glyphs such as brackets on right-to-left levels.
			r, size := utf8.DecodeRuneInString(c.text)
			if m := mirror(r); m != r {
				c.text = string(m) + c.text[size:]
			}
		}
		write(c)
	}
	for _, c := range cells[stop:] {
		write(c)
	}
	cur.transition(b, end)
	return end
}

// sgrState is the styling that applies to text: the SGR sequences set since
// the last reset, and the open OSC 8 hyperlink, if any.
type sgrState struct {
	sgr  string
	link string
}

// apply updates s with the escape sequence seq, and reports whether it's one
// that can be carried over, i.e. an SGR sequence or an OSC 8 hyperlink.
func (s *sgrState) apply(seq string) bool {
	switch {
	case strings.HasPrefix(seq, "\x1b[") && strings.HasSuffix(seq, "m"):
		params := seq[2 : len(seq)-1]
		switch {
		case strings.Trim(params, "0123456789;:") != "":
			return false
		case params == "" || params == "0":
			s.sgr = ""
		case strings.HasPrefix(params, "0;"):
			s.sgr = seq
		default:
			s.sgr += seq
		}
		return true
	case strings.HasPrefix(seq, "\x1b]8;"):
		body := strings.TrimSuffix(strings.TrimSuffix(seq[len("\x1b]8;"):], "\a"), "\x1b\\")
		if _, uri, _ := strings.Cut(body, ";"); uri != "" {
			s.link = seq
		} else {
			s.link = ""
		}
		return true
	}
	return false
}

// transition writes the escape sequences that change styling s to t.
func (s sgrState) transition(b *strings.Builder, t sgrState) {
	if s.link != t.link && s.link != "" {
		b.WriteString(ansi.ResetHyperlink())
	}
	if strings.HasPrefix(t.sgr, s.sgr) {
		b.WriteString(t.sgr[len(s.sgr):])
	} else {
		b.WriteString(ansi.ResetStyle)
		b.WriteString(t.sgr)
	}
	if s.link != t.link && t.link != "" {
		b.WriteString(t.link)
	}
}

func isEscape(tok string) bool {
	return tok != "" && (tok[0] < ' ' || tok[0] == 0x7f)
}

func firstRune(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

func bidiClass(r rune) bidi.Class {
	p, _ := bidi.LookupRune(r)
	return p.Class()
}

// bidiLevels resolves the embedding level of each character in a paragraph
// at the given level (0 for left-to-right, 1 for right-to-left), following
// rules W1–W7, N0–N2, I1–I2 and L1 of UAX #9. Explicit embeddings,
// overrides and isolates aren't supported, and are treated as boundary
// neutrals.
func bidiLevels(chars []rune, level int) []int {
	n := len(chars)
	orig := make([]bidi.Class, n)
	for i, r := range chars {
		c := bidiClass(r)
		if c >= bidi.Control {
			c = bidi.BN
		}
		orig[i] = c
	}
	t := slices.Clone(orig)
	e := bidi.L // embedding direction, also used for sos and eos
	if level%2 == 1 {
		e = bidi.R
	}

	// W1: nonspacing marks take the type of the character before them.
	prev := e
	for i, c := range t {
		if c == bidi.NSM || c == bidi.BN {
			t[i] = prev
		}
		prev = t[i]
	}

	// W2, W3: European numbers after Arabic letters become Arabic numbers,
	// and Arabic letters become R.
	last := e
	for i, c := range t {
		switch c {
		case bidi.L, bidi.R, bidi.AL:
			last = c
		case bidi.EN:
			if last == bidi.AL {
				t[i] = bidi.AN
			}
		}
	}
	for i, c := range t {
		if c == bidi.AL {
			t[i] = bidi.R
		}
	}

	// W4: a single separator between two numbers of the same type joins them.
	for i := 1; i+1 < n; i++ {
		switch {
		case t[i] == bidi.ES && t[i-1] == bidi.EN && t[i+1] == bidi.EN:
			t[i] = bidi.EN
		case t[i] == bidi.CS && t[i-1] == t[i+1] && (t[i-1] == bidi.EN || t[i-1] == bidi.AN):
			t[i] = t[i-1]
		}
	}

	// W5: terminators next to European numbers join them.
	for i := 0; i < n; {
		if t[i] != bidi.ET {
			i++
			continue
		}
		j := i
		for j < n && t[j] == bidi.ET {
			j++
		}
		if (i > 0 && t[i-1] == bidi.EN) || (j < n && t[j] == bidi.EN) {
			for k := i; k < j; k++ {
				t[k] = bidi.EN
			}
		}
		i = j
	}

	// W6: remaining separators and terminators are neutral.
	for i, c := range t {
		if c == bidi.ES || c == bidi.ET || c == bidi.CS {
			t[i] = bidi.ON
		}
	}

	// W7: European numbers in left-to-right context become L.
	last = e
	for i, c := range t {
		switch c {
		case bidi.L, bidi.R:
			last = c
		case bidi.EN:
			if last == bidi.L {
				t[i] = bidi.L
			}
		}
	}

	resolveBrackets(chars, orig, t, e)

	// N1, N2: neutrals between characters of the same direction take that
	// direction, the rest take the embedding direction.
	for i := 0; i < n; {
		if !isNeutral(t[i]) {
			i++
			continue
		}
		j := i
		for j < n && isNeutral(t[j]) {
			j++
		}
		before, after := e, e
		if i > 0 {
			before, _ = strongDirection(t[i-1])
		}
		if j < n {
			after, _ = strongDirection(t[j])
		}
		d := e
		if before == after {
			d = before
		}
		for k := i; k < j; k++ {
			t[k] = d
		}
		i = j
	}

	// I1, I2
	levels := make([]int, n)
	for i, c := range t {
		levels[i] = level
		switch {
		case level%2 == 0 && c == bidi.R:
			levels[i]++
		case level%2 == 0 && (c == bidi.AN || c == bidi.EN):
			levels[i] += 2
		case level%2 == 1 && c != bidi.R:
			levels[i]++
		}
	}

	// L1: separators, and whitespace before them or at the end of the line,
	// are reset to the paragraph level.
	trailing := true
	for i := n - 1; i >= 0; i-- {
		switch orig[i] {
		case bidi.S, bidi.B:
			levels[i] = level
			trailing = true
		case bidi.WS, bidi.BN:
			if trailing {
				levels[i] = level
			}
		default:
			trailing = false
		}
	}
	return levels
}

// resolveBrackets applies rule N0: paired brackets take the direction of the
// text they enclose, or of the text before them.
func resolveBrackets(chars []rune, orig, t []bidi.Class, e bidi.Class) {
	type pair struct{ open, close int }
	type opener struct {
		pos    int
		closer rune
	}

	// BD16: find the bracket pairs.
	var pairs []pair
	var stack []opener
	for i, r := range chars {
		if t[i] != bidi.ON {
			continue
		}
		p, _ := bidi.LookupRune(r)
		if !p.IsBracket() {
			continue
		}
		if p.IsOpeningBracket() {
			if len(stack) == 63 {
				break
			}
			stack = append(stack, opener{i, mirror(r)})
			continue
		}
		for k := len(stack) - 1; k >= 0; k-- {
			if stack[k].closer == r {
				pairs = append(pairs, pair{stack[k].pos, i})
				stack = stack[:k]
				break
			}
		}
	}
	slices.SortFunc(pairs, func(a, b pair) int { return a.open - b.open })

	opposite := bidi.R
	if e == bidi.R {
		opposite = bidi.L
	}
	set := func(i int, d bidi.Class) {
		t[i] = d
		for j := i + 1; j < len(t) && orig[j] == bidi.NSM; j++ {
			t[j] = d
		}
	}
	for _, p := range pairs {
		found := false
		d := opposite
		for k := p.open + 1; k < p.close; k++ {
			if s, ok := strongDirection(t[k]); ok {
				found = true
				if s == e {
					d = e
					break
				}
			}
		}
		if !found {
			continue
		}
		if d == opposite {
			// Keep the opposite direction only if the text before the
			// brackets has it too.
			context := e
			for k := p.open - 1; k >= 0; k-- {
				if s, ok := strongDirection(t[k]); ok {
					context = s
					break
				}
			}
			if context != opposite {
				d = e
			}
		}
		set(p.open, d)
		set(p.close, d)
	}
}

func isNeutral(c bidi.Class) bool {
	switch c {
	case bidi.B, bidi.S, bidi.WS, bidi.ON:
		return true
	}
	return false
}

// strongDirection reports the direction a resolved character type counts as
// when resolving neutrals; numbers count as right-to-left.
func strongDirection(c bidi.Class) (bidi.Class, bool) {
	switch c {
	case bidi.L:
		return bidi.L, true
	case bidi.R, bidi.EN, bidi.AN:
		return bidi.R, true
	}
	return bidi.ON, false
}

// visualOrder returns the indices of characters with the given embedding
// levels in the order they're displayed (rule L2).
func visualOrder(levels []int) []int {
	order := make([]int, len(levels))
	highest, lowestOdd := 0, -1
	for i, l := range levels {
		order[i] = i
		highest = max(highest, l)
		if l%2 == 1 && (lowestOdd < 0 || l < lowestOdd) {
			lowestOdd = l
		}
	}
	if lowestOdd < 0 {
		return order
	}
	for l := highest; l >= lowestOdd; l-- {
		for i := 0; i < len(order); {
			if levels[order[i]] < l {
				i++
				continue
			}
			j := i
			for j < len(order) && levels[order[j]] >= l {
				j++
			}
			slices.Reverse(order[i:j])
			i = j
		}
	}
	return order
}

// mirroredGlyphs are mirrored characters that aren't paired brackets.
var mirroredGlyphs = map[rune]rune{
	'<': '>', '>': '<',
	'«': '»', '»': '«',
	'‹': '›', '›': '‹',
	'≤': '≥', '≥': '≤',
}

// mirror returns the mirror image of r, or r if it has none.
func mirror(r rune) rune {
	if p, _ := bidi.LookupRune(r); p.IsBracket() {
		return firstRune(bidi.ReverseString(string(r)))
	}
	if m, ok := mirroredGlyphs[r]; ok {
		return m
	}
	return r
}
