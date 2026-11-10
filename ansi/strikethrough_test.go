package ansi

import (
	"bytes"
	"testing"

	astext "github.com/yuin/goldmark/extension/ast"
)

func TestEmptyStrikethrough(t *testing.T) {
	ctx := NewRenderContext(Options{Styles: StyleConfig{
		Strikethrough: StylePrimitive{BlockPrefix: "~", Prefix: "[", Suffix: "]", BlockSuffix: "~"},
	}})
	e := &strikethroughElement{node: astext.NewStrikethrough()}
	var b bytes.Buffer
	if err := e.Render(&b, ctx); err != nil {
		t.Fatal(err)
	}
	if got, want := b.String(), "~[]~"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if ctx.blockStack.Len() != 0 {
		t.Fatal("strikethrough changed the caller's block stack")
	}
}
