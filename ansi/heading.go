package ansi

import (
	"bytes"
	"fmt"
	"io"

	"charm.land/lipgloss/v2"
)

// A HeadingElement is used to render headings.
type HeadingElement struct {
	Level int
	First bool
}

const (
	h1 = iota + 1
	h2
	h3
	h4
	h5
	h6
)

// Render renders a HeadingElement.
func (e *HeadingElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := ctx.options.Styles.Heading

	switch e.Level {
	case h1:
		rules = cascadeStyles(rules, ctx.options.Styles.H1)
	case h2:
		rules = cascadeStyles(rules, ctx.options.Styles.H2)
	case h3:
		rules = cascadeStyles(rules, ctx.options.Styles.H3)
	case h4:
		rules = cascadeStyles(rules, ctx.options.Styles.H4)
	case h5:
		rules = cascadeStyles(rules, ctx.options.Styles.H5)
	case h6:
		rules = cascadeStyles(rules, ctx.options.Styles.H6)
	}

	if !e.First {
		_, _ = renderText(w, bs.Current().Style.StylePrimitive, "\n")
	}

	be := BlockElement{
		Block: &bytes.Buffer{},
		Style: cascadeStyle(bs.Current().Style, rules, false),
	}
	bs.Push(be)

	_, _ = renderText(w, bs.Parent().Style.StylePrimitive, rules.BlockPrefix)
	_, _ = renderText(bs.Current().Block, bs.Current().Style.StylePrimitive, rules.Prefix)
	if ctx.options.BidiReordering {
		_, _ = io.WriteString(bs.Current().Block, bidiOpen)
	}
	return nil
}

// Finish finishes rendering a HeadingElement.
func (e *HeadingElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := bs.Current().Style
	mw := NewMarginWriter(ctx, w, rules)
	defer mw.Close() //nolint:errcheck

	blk := bs.Current().Block.String()
	if ctx.options.BidiReordering {
		blk = closeBidiFlow(blk)
	}
	flow := lipgloss.Wrap(blk, int(bs.Width(ctx)), "")
	if ctx.options.BidiReordering {
		flow = reorderBidi(flow)
	}
	_, err := io.WriteString(mw, flow)
	if err != nil {
		return fmt.Errorf("glamour: error writing to writer: %w", err)
	}

	_, _ = renderText(w, bs.Current().Style.StylePrimitive, rules.Suffix)
	_, _ = renderText(w, bs.Parent().Style.StylePrimitive, rules.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
