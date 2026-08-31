package ansi

import (
	"bytes"
	"testing"
)

func TestRenderTextConceal(t *testing.T) {
	// Regression test for https://github.com/charmbracelet/glamour/issues/121
	// and https://github.com/charmbracelet/glow/issues/645. Setting
	// "conceal": true on a style primitive must actually hide the rendered
	// text, not just be parsed into the struct and dropped.
	conceal := true

	tests := []struct {
		name  string
		rules StylePrimitive
		input string
		want  string
	}{
		{
			name:  "conceal true hides text",
			rules: StylePrimitive{Conceal: &conceal},
			input: "https://example.com",
			want:  "",
		},
		{
			name:  "conceal nil renders normally",
			rules: StylePrimitive{},
			input: "https://example.com",
			want:  "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := renderText(&buf, tt.rules, tt.input)
			if err != nil {
				t.Fatalf("renderText returned error: %v", err)
			}
			if buf.String() != tt.want {
				t.Errorf("renderText output: got %q, want %q", buf.String(), tt.want)
			}
			if n != len(tt.want) {
				t.Errorf("renderText byte count: got %d, want %d", n, len(tt.want))
			}
		})
	}
}
