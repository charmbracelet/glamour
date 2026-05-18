package ansi

import (
	"bytes"
	"testing"
)

func TestRenderText_ConcealHidesOutput(t *testing.T) {
	conceal := true
	rules := StylePrimitive{Conceal: &conceal}

	var buf bytes.Buffer
	n, err := renderText(&buf, rules, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 || buf.Len() != 0 {
		t.Errorf("expected concealed output to be empty, got %q (n=%d)", buf.String(), n)
	}
}

func TestRenderText_ConcealFalseStillRenders(t *testing.T) {
	conceal := false
	rules := StylePrimitive{Conceal: &conceal}

	var buf bytes.Buffer
	if _, err := renderText(&buf, rules, "hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("hello")) {
		t.Errorf("expected output to contain %q, got %q", "hello", buf.String())
	}
}
