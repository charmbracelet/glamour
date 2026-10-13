package ansi

import (
	"bytes"
	"testing"
)

func TestRenderTextEscapeReplacer(t *testing.T) {
	// Regression test for https://github.com/charmbracelet/glamour/issues/503.
	// Any ASCII punctuation backslash-escape should be stripped, per the
	// CommonMark spec. Previously only a subset was handled, so `\~`, `\?`,
	// `\@` etc. rendered verbatim.
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "backslash", in: `\\`, want: `\`},
		{name: "tilde was missing", in: `\~`, want: `~`},
		{name: "question mark was missing", in: `\?`, want: `?`},
		{name: "at sign was missing", in: `\@`, want: `@`},
		{name: "caret was missing", in: `\^`, want: `^`},
		{name: "double quote was missing", in: `\"`, want: `"`},
		{name: "ampersand was missing", in: `\&`, want: `&`},
		{name: "equals was missing", in: `\=`, want: `=`},
		{name: "slash was missing", in: `\/`, want: `/`},
		{name: "colon was missing", in: `\:`, want: `:`},
		{name: "semicolon was missing", in: `\;`, want: `;`},
		{name: "asterisk still works", in: `\*`, want: `*`},
		{name: "underscore still works", in: `\_`, want: `_`},
		{name: "tilde inside a word", in: `foo\~bar`, want: `foo~bar`},
		{name: "double-escaped tilde keeps the backslash", in: `\\\~`, want: `\~`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeReplacer.Replace(tt.in)
			if got != tt.want {
				t.Errorf("escapeReplacer.Replace(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRenderText_EscapesPropagateThroughRender(t *testing.T) {
	// Sanity check that doRender actually feeds the token through
	// escapeReplacer, so the fix is visible in the rendered output and not
	// just in the replacer itself.
	var buf bytes.Buffer
	e := &BaseElement{Token: `hello \~ world`}
	if err := e.doRender(&buf, StylePrimitive{}, StylePrimitive{}); err != nil {
		t.Fatalf("doRender returned error: %v", err)
	}
	got := buf.String()
	want := "hello ~ world"
	if got != want {
		t.Errorf("doRender output: got %q, want %q", got, want)
	}
}
