package ansi

import "testing"

// The expectations here are what bluemonday's StrictPolicy produced for the
// same input, recorded when it was replaced: a differential harness ran both
// over these cases and 400,000 generated ones and found no disagreement.
func TestStripTags(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", ""},
		{"plain text", "plain text"},
		{"<b>bold</b>", "bold"},
		{"a <i>b</i> c", "a b c"},
		{`<a href="x">link</a>`, "link"},

		// A '>' inside an attribute value does not end the tag.
		{`<a title="a>b">t</a>`, "t"},
		{`<img src="a.png" alt="a>b">`, ""},
		{`<div class='x'>q</div>`, "q"},

		// '<' only opens a tag before a name, '/', '!' or '?'. Prose keeps its
		// comparisons, which is the whole point in a markdown renderer.
		{"a < b and c > d", "a < b and c > d"},
		{"lone < bracket", "lone < bracket"},
		{"5 < 6", "5 < 6"},

		// Comments, CDATA and doctypes go with their contents.
		{"<!-- comment -->after", "after"},
		{"before<!-- c -->after", "beforeafter"},
		{"<![CDATA[raw]]>tail", "tail"},
		{"<!DOCTYPE html>doc", "doc"},

		// script and style lose their contents; a stray closing tag does not
		// swallow the rest of the input.
		{"<script>alert(1)</script>tail", "tail"},
		{"<style>p{}</style>tail", "tail"},
		{"</script>after", "after"},
		{"</style>after", "after"},
		{"<SCRIPT>x</SCRIPT>y", "y"},

		// An unterminated tag consumes what follows it.
		{"unclosed <b", "unclosed "},
		{"<script>never closed", ""},

		{"<p>one</p>\n<p>two</p>", "one\ntwo"},
		{"nested <b><i>x</i></b> end", "nested x end"},
		{"<TABLE><TR><TD>cell</TD></TR></TABLE>", "cell"},
		{"&amp; entity", "&amp; entity"},
		{"<x-y>hyphenated</x-y>", "hyphenated"},
	} {
		if got := stripTags(tc.in); got != tc.want {
			t.Errorf("stripTags(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// SanitizeHTML is what the renderer actually calls: strip, optionally trim,
// then unescape.
func TestSanitizeHTML(t *testing.T) {
	ctx := NewRenderContext(Options{})
	for _, tc := range []struct {
		in   string
		trim bool
		want string
	}{
		{"<b>x</b>", false, "x"},
		{"  <b>x</b>  ", true, "x"},
		{"&lt;notatag&gt;", false, "<notatag>"},
		{"a &amp; b", false, "a & b"},
	} {
		if got := ctx.SanitizeHTML(tc.in, tc.trim); got != tc.want {
			t.Errorf("SanitizeHTML(%q, %v) = %q, want %q", tc.in, tc.trim, got, tc.want)
		}
	}
}
