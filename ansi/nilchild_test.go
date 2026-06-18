package ansi

import (
	"bytes"
	"testing"
)

// TestEmphasisRenderSkipsNilChild guards against a nil pointer dereference when
// an emphasis element holds a child whose node kind has no renderer. Such
// children enter the Children slice as a nil ElementRenderer (NewElement
// returns an empty Element for unhandled kinds), and doRender used to call
// child.Render on them directly, panicking. The renderer must tolerate a nil
// child instead of crashing.
func TestEmphasisRenderSkipsNilChild(t *testing.T) {
	var buf bytes.Buffer
	e := &EmphasisElement{Children: []ElementRenderer{nil}}
	if err := e.Render(&buf, RenderContext{}); err != nil {
		t.Fatalf("render: %v", err)
	}
}
