// Package tests_test contains integration tests to validate the rendering
// behaviour of HTML images (<img src alt>) in glamour.
//
// These tests cover:
//   - Block HTML image (<img> on its own paragraph) → KindHTMLBlock
//   - Inline HTML image (<img> inside a paragraph with text) → KindRawHTML
//   - Consistency with the equivalent Markdown syntax (![alt](src))
//   - Preservation of code blocks containing <img> (must not be treated as images)
//   - Attribute variations: single quotes, uppercase tag, missing alt, missing src
package tests_test

import (
	"regexp"
	"strings"
	"testing"

	glamour "charm.land/glamour/v2"
)

// stripANSI strips ANSI escape sequences and OSC 8 hyperlinks from a string,
// returning only the visible text.
func stripANSI(s string) string {
	// Remove OSC 8 hyperlinks (\x1b]8;...;\x07 and variants)
	oscRe := regexp.MustCompile(`\x1b\][^\x07]*\x07`)
	s = oscRe.ReplaceAllString(s, "")
	// Remove remaining ESC sequences
	ansiRe := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	s = ansiRe.ReplaceAllString(s, "")
	return s
}

func newDarkRenderer(t *testing.T) *glamour.TermRenderer {
	t.Helper()
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}
	return r
}

// ---------------------------------------------------------------------------
// Expected behaviour tests (after the fix)
// ---------------------------------------------------------------------------

// TestHTMLImgBlockRendersURL checks that a block <img> includes the URL in the output.
func TestHTMLImgBlockRendersURL(t *testing.T) {
	r := newDarkRenderer(t)

	out, err := r.Render("<img src=\"https://charm.sh/logo.png\" alt=\"Charm Logo\">\n")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	visible := stripANSI(out)
	if !strings.Contains(visible, "https://charm.sh/logo.png") {
		t.Errorf("URL not found in output.\nVisible output: %q", visible)
	}
}

// TestHTMLImgBlockRendersAlt checks that the alt text of a block <img> appears in the output.
func TestHTMLImgBlockRendersAlt(t *testing.T) {
	r := newDarkRenderer(t)

	out, err := r.Render("<img src=\"https://charm.sh/logo.png\" alt=\"Charm Logo\">\n")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	visible := stripANSI(out)
	if !strings.Contains(visible, "Charm Logo") {
		t.Errorf("alt text not found in output.\nVisible output: %q", visible)
	}
}

// TestHTMLImgInlineRendersURL checks that an inline <img> inside a paragraph renders the URL.
func TestHTMLImgInlineRendersURL(t *testing.T) {
	r := newDarkRenderer(t)

	out, err := r.Render("Text before <img src=\"https://charm.sh/logo.png\" alt=\"Logo\"> text after.\n")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	visible := stripANSI(out)
	if !strings.Contains(visible, "https://charm.sh/logo.png") {
		t.Errorf("URL not found in inline output.\nVisible output: %q", visible)
	}
}

// TestHTMLImgConsistentWithMarkdownImg verifies that <img src alt> produces the same
// visible content as the equivalent Markdown syntax ![alt](src).
//
// Note: padding spaces added by KindParagraph (which wraps Markdown images) are
// not present in KindHTMLBlock, so we compare only the visible content
// (ANSI-stripped, whitespace-collapsed).
func TestHTMLImgConsistentWithMarkdownImg(t *testing.T) {
	cases := []struct {
		name string
		md   string
		html string
	}{
		{
			name: "block image",
			md:   "![Charm Logo](https://charm.sh/logo.png)\n",
			html: "<img src=\"https://charm.sh/logo.png\" alt=\"Charm Logo\">\n",
		},
		{
			name: "single quotes",
			md:   "![Logo](https://example.com/img.png)\n",
			html: "<img src='https://example.com/img.png' alt='Logo'>\n",
		},
		{
			name: "uppercase tag",
			md:   "![Logo](https://example.com/img.png)\n",
			html: "<IMG SRC=\"https://example.com/img.png\" ALT=\"Logo\">\n",
		},
	}

	r := newDarkRenderer(t)

	normalizeVisible := func(s string) string {
		// strip ANSI, collapse multiple spaces, trim
		s = stripANSI(s)
		s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
		return strings.TrimSpace(s)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mdOut, err := r.Render(tc.md)
			if err != nil {
				t.Fatalf("failed to render markdown: %v", err)
			}
			htmlOut, err := r.Render(tc.html)
			if err != nil {
				t.Fatalf("failed to render html: %v", err)
			}
			mdVisible := normalizeVisible(mdOut)
			htmlVisible := normalizeVisible(htmlOut)
			if mdVisible != htmlVisible {
				t.Errorf(
					"<img> HTML should produce the same visible content as ![alt](src) Markdown.\n"+
						"Markdown visible: %q\n"+
						"HTML visible:     %q",
					mdVisible, htmlVisible,
				)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Preservation tests — <img> inside code must NOT be treated as an image
// ---------------------------------------------------------------------------

// TestImgInsideFencedCodeBlockIsPreserved ensures that <img> inside a fenced
// code block is rendered literally and not treated as an image.
func TestImgInsideFencedCodeBlockIsPreserved(t *testing.T) {
	r := newDarkRenderer(t)

	md := "```html\n<img src=\"https://example.com/img.png\" alt=\"Logo\">\n```\n"
	out, err := r.Render(md)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	visible := stripANSI(out)
	// The literal tag must appear in the output
	if !strings.Contains(visible, "<img") {
		t.Errorf("<img> tag should be literal inside a code block.\nVisible output: %q", visible)
	}
	// The src attribute must also appear literally (not rendered as a hyperlink)
	if !strings.Contains(visible, "src=") {
		t.Errorf("src attribute should appear literally inside a code block.\nVisible output: %q", visible)
	}
}

// TestImgInsideIndentedCodeBlockIsPreserved ensures that <img> inside a
// 4-space-indented code block is rendered literally and not treated as an image.
func TestImgInsideIndentedCodeBlockIsPreserved(t *testing.T) {
	r := newDarkRenderer(t)

	md := "    <img src=\"https://example.com/img.png\" alt=\"Logo\">\n"
	out, err := r.Render(md)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	visible := stripANSI(out)
	if !strings.Contains(visible, "<img") {
		t.Errorf("<img> tag should be literal inside an indented code block.\nVisible output: %q", visible)
	}
}

// TestImgInsideInlineCodeIsPreserved ensures that <img> inside inline code
// (`...`) is rendered literally and not treated as an image.
func TestImgInsideInlineCodeIsPreserved(t *testing.T) {
	r := newDarkRenderer(t)

	md := "Use `<img src=\"url\" alt=\"text\">` to insert images.\n"
	out, err := r.Render(md)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	visible := stripANSI(out)
	if !strings.Contains(visible, "<img") {
		t.Errorf("<img> tag should be literal inside inline code.\nVisible output: %q", visible)
	}
}

// ---------------------------------------------------------------------------
// Edge case tests
// ---------------------------------------------------------------------------

// TestHTMLImgNoAlt checks that <img> without an alt attribute still renders the URL.
func TestHTMLImgNoAlt(t *testing.T) {
	r := newDarkRenderer(t)

	out, err := r.Render("<img src=\"https://charm.sh/logo.png\">\n")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	visible := stripANSI(out)
	if !strings.Contains(visible, "https://charm.sh/logo.png") {
		t.Errorf("URL should appear in output even without alt.\nVisible output: %q", visible)
	}
}

// TestHTMLImgNoSrc checks that <img> without a src attribute does not panic
// and produces some output (even if empty or alt-only).
func TestHTMLImgNoSrc(t *testing.T) {
	r := newDarkRenderer(t)

	// Must not panic
	_, err := r.Render("<img alt=\"Logo without src\">\n")
	if err != nil {
		t.Errorf("rendering <img> without src should not return an error: %v", err)
	}
}

// TestNonImgHTMLIsUnchanged checks that other HTML tags (e.g. <br>) continue
// to be handled normally (sanitized, not confused with images).
func TestNonImgHTMLIsUnchanged(t *testing.T) {
	r := newDarkRenderer(t)

	out, err := r.Render("<br>\n")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	// <br> is sanitized to an empty string — must not produce any URL
	visible := stripANSI(out)
	if strings.Contains(visible, "http") {
		t.Errorf("<br> tag should not produce URLs in the output.\nVisible output: %q", visible)
	}
}
