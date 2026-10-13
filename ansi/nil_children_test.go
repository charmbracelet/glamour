package ansi

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmphasisElementSkipsNilChildren(t *testing.T) {
	t.Parallel()

	el := &EmphasisElement{
		Children: []ElementRenderer{
			nil,
			&BaseElement{Token: "ok"},
		},
		Level: 1,
	}

	var buf bytes.Buffer
	ctx := NewRenderContext(Options{})

	if err := el.Render(&buf, ctx); err != nil {
		t.Fatalf("render: %v", err)
	}

	if got := buf.String(); !strings.Contains(got, "ok") {
		t.Fatalf("expected rendered output to contain %q, got %q", "ok", got)
	}
}

func TestLinkElementSkipsNilChildren(t *testing.T) {
	t.Parallel()

	el := &LinkElement{
		URL: "https://example.com",
		Children: []ElementRenderer{
			nil,
			&BaseElement{Token: "link"},
		},
	}

	var buf bytes.Buffer
	ctx := NewRenderContext(Options{})

	if err := el.Render(&buf, ctx); err != nil {
		t.Fatalf("render: %v", err)
	}

	if got := buf.String(); !strings.Contains(got, "link") {
		t.Fatalf("expected rendered output to contain %q, got %q", "link", got)
	}
}

func TestTableCellElementSkipsNilChildren(t *testing.T) {
	t.Parallel()

	el := &TableCellElement{
		Children: []ElementRenderer{
			nil,
			&BaseElement{Token: "cell"},
		},
	}

	ctx := NewRenderContext(Options{})

	if err := el.Render(nil, ctx); err != nil {
		t.Fatalf("render: %v", err)
	}

	if len(ctx.table.row) != 1 {
		t.Fatalf("expected 1 table row cell, got %d", len(ctx.table.row))
	}

	if !strings.Contains(ctx.table.row[0], "cell") {
		t.Fatalf("expected table cell to contain %q, got %q", "cell", ctx.table.row[0])
	}
}
