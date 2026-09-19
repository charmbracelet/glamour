//go:build glamour_nochroma

package ansi

import "io"

// codeBlockTheme always returns an empty theme: the glamour_nochroma build
// tag compiles chroma out, so code blocks render with the style's plain
// code block rules.
func codeBlockTheme(StyleCodeBlock) string {
	return ""
}

// highlightCode is never reached without a theme.
func highlightCode(io.Writer, string, string, string, string) error {
	return nil
}
