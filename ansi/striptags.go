package ansi

import "strings"

// stripTags removes every HTML tag from s and keeps the text between them.
//
// This is what a bluemonday StrictPolicy did here — an allow-nothing policy,
// which is a tag stripper wearing a sanitiser's clothes. That policy was this
// package's only use of bluemonday, and it brought golang.org/x/net/html and
// five golang.org/x/text packages with it, whose generated tables are a large
// share of a TinyGo compile and of the resulting binary.
//
// It follows the tokenizer's rule that '<' only opens a tag when a name, '/',
// '!' or '?' follows it, so prose like "a < b and c > d" survives intact —
// which matters more here than anywhere, because this is markdown being
// rendered for a terminal and comparisons are ordinary text.
//
// The contents of script and style elements are dropped with their tags, as a
// strict policy does: they are code, not prose, and printing them to a terminal
// helps nobody. Entities are left as found, because SanitizeHTML unescapes
// afterwards and a strict policy's escape-then-unescape round trip is a no-op.
func stripTags(s string) string {
	if !strings.ContainsRune(s, '<') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] != '<' || !opensTag(s, i) {
			b.WriteByte(s[i])
			i++
			continue
		}
		// A comment, CDATA section or doctype runs to its own terminator, and
		// everything inside it goes with it.
		if rest, ok := skipDelimited(s, i, "<!--", "-->"); ok {
			i = rest
			continue
		}
		if rest, ok := skipDelimited(s, i, "<![CDATA[", "]]>"); ok {
			i = rest
			continue
		}
		name, end := tagName(s, i)
		if end < 0 {
			return b.String() // unterminated tag: the rest is not text
		}
		// script and style: drop the element's contents too, but only for the
		// opening tag. A stray closing tag is just a tag, and treating it as one
		// end of a pair would swallow everything after it.
		if (name == "script" || name == "style") && s[i+1] != '/' {
			if c := strings.Index(strings.ToLower(s[end:]), "</"+name); c >= 0 {
				if _, e := tagName(s, end+c); e >= 0 {
					i = e
					continue
				}
			}
			return b.String()
		}
		i = end
	}
	return b.String()
}

// opensTag reports whether the '<' at i begins a tag rather than being text.
func opensTag(s string, i int) bool {
	if i+1 >= len(s) {
		return false
	}
	c := s[i+1]
	return c == '/' || c == '!' || c == '?' ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// skipDelimited consumes s[i:] when it starts with open, up to and including
// close, and reports the index after it.
func skipDelimited(s string, i int, open, close string) (int, bool) {
	if !strings.HasPrefix(s[i:], open) {
		return i, false
	}
	if e := strings.Index(s[i+len(open):], close); e >= 0 {
		return i + len(open) + e + len(close), true
	}
	return len(s), true // unterminated: the rest belongs to it
}

// tagName returns the lower-cased element name of the tag starting at i and the
// index just past its '>', or -1 if the tag never closes. Quotes are honoured,
// so <a title="a>b"> is one tag.
func tagName(s string, i int) (string, int) {
	j := i + 1
	if j < len(s) && s[j] == '/' {
		j++
	}
	start := j
	for j < len(s) && (isNameByte(s[j])) {
		j++
	}
	name := strings.ToLower(s[start:j])
	quote := byte(0)
	for ; j < len(s); j++ {
		switch c := s[j]; {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return name, j + 1
		}
	}
	return name, -1
}

func isNameByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == ':'
}
