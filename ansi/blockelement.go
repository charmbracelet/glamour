package ansi

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// BlockElement provides a render buffer for children of a block element.
// After all children have been rendered into it, it applies indentation and
// margins around them and writes everything to the parent rendering buffer.
type BlockElement struct {
	Block   *bytes.Buffer
	Style   StyleBlock
	Margin  bool
	Newline bool
}

// Render renders a BlockElement.
func (e *BlockElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	bs.Push(*e)

	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, e.Style.BlockPrefix)
	renderText(bs.Current().Block, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, e.Style.Prefix)
	return nil
}

// Finish finishes rendering a BlockElement.
func (e *BlockElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	if e.Margin { //nolint: nestif
		// Calculate target width based on alignment and margins
		targetWidth := int(bs.Width(ctx))
		
		// Always account for margins in wrapping width
		var leftMargin, rightMargin uint
		if bs.Current().Style.MarginLeft != nil {
			leftMargin = *bs.Current().Style.MarginLeft
		}
		if bs.Current().Style.MarginRight != nil {
			rightMargin = *bs.Current().Style.MarginRight
		}
		
		// Subtract margins from available width for all cases
		if int(leftMargin+rightMargin) < targetWidth {
			targetWidth = targetWidth - int(leftMargin+rightMargin)
		}
		
		// Additional width adjustments for specific alignment types
		if bs.Current().Style.Align != nil && *bs.Current().Style.Align == "center" {
			// Use about 70% of remaining width for better text flow when centering
			targetWidth = int(float64(targetWidth) * 0.7)
		}
		
		// Calculate the indent string for wrapped lines using the actual style margins
		var baseIndentation uint
		var styleLeftMargin uint
		
		if rules := bs.Current().Style; rules.Indent != nil {
			baseIndentation = *rules.Indent
		}
		if rules := bs.Current().Style; rules.MarginLeft != nil {
			styleLeftMargin = *rules.MarginLeft
		}
		
		totalIndent := baseIndentation + styleLeftMargin
		indentStr := strings.Repeat(" ", int(totalIndent))
		
		s := WordwrapWithIndent(
			bs.Current().Block.String(),
			targetWidth, //nolint: gosec
			" ,.;-+|",
			indentStr,
		)

		mw := NewMarginWriter(ctx, w, bs.Current().Style)
		if _, err := io.WriteString(mw, s); err != nil {
			return fmt.Errorf("glamour: error writing to writer: %w", err)
		}

		if e.Newline {
			if _, err := io.WriteString(mw, "\n"); err != nil {
				return fmt.Errorf("glamour: error writing to writer: %w", err)
			}
		}
	} else {
		_, err := bs.Parent().Block.Write(bs.Current().Block.Bytes())
		if err != nil {
			return fmt.Errorf("glamour: error writing to writer: %w", err)
		}
	}

	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, e.Style.Suffix)
	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, e.Style.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
