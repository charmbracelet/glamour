package ansi

import (
	"bytes"
	"testing"
)

func TestLiteralElement(t *testing.T) {
	for _, override := range []bool{false, true} {
		ctx := NewRenderContext(Options{})
		ctx.blockStack.Push(BlockElement{})
		e := &literalElement{
			Token:  `foo\+bar`,
			Prefix: "[",
			Suffix: "]",
			Style: StylePrimitive{
				Prefix: "(", Suffix: ")", Format: "{{.text}}",
			},
		}
		var out bytes.Buffer
		var err error
		want := `[(foo\+bar)]`
		if override {
			upper := true
			err = e.StyleOverrideRender(&out, ctx, StylePrimitive{Upper: &upper})
			want = `[(FOO\+BAR)]`
		} else {
			err = e.Render(&out, ctx)
		}
		if err != nil {
			t.Fatal(err)
		}
		if out.String() != want {
			t.Fatalf("override=%v: got %q, want %q", override, out.String(), want)
		}
	}
}
