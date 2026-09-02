package ansi

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// ListItemParagraphElement wraps prose-like list item content independently of
// its prefix.
type ListItemParagraphElement struct {
	First bool
	Child ElementRenderer
}

// Render starts buffering the list item's paragraph.
func (e *ListItemParagraphElement) Render(_ io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	block := &bytes.Buffer{}
	// Pre-render synthetic children before changing the block stack.
	if e.Child != nil {
		if err := e.Child.Render(block, ctx); err != nil {
			return fmt.Errorf("glamour: error rendering list item content: %w", err)
		}
	}
	bs.Push(BlockElement{
		Block: block,
		// Inherit text styling without deducting the list layout twice.
		Style: StyleBlock{StylePrimitive: bs.Current().Style.StylePrimitive},
	})
	return nil
}

// Finish wraps the item text and adds the prefix width after each newline.
func (e *ListItemParagraphElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	prefixWidth := bs.Parent().PrefixWidth
	// Keep wrapping defined when the marker consumes the available width.
	width := max(1, int(bs.Width(ctx))-prefixWidth)
	s := lipgloss.Wrap(bs.Current().Block.String(), width, " ,.;-+|")
	indent := strings.Repeat(" ", prefixWidth)
	// Align every continuation with the first line's item text.
	s = strings.ReplaceAll(s, "\n", "\n"+indent)
	if !e.First {
		// Preserve one blank line without duplicating a preceding block's output.
		trailingNewlines := trailingLineBreaks(bs.Parent().Block.String())
		s = strings.Repeat("\n", max(0, 2-trailingNewlines)) + indent + s
	}

	if _, err := io.WriteString(w, s); err != nil {
		return fmt.Errorf("glamour: error writing to writer: %w", err)
	}

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}

// trailingLineBreaks counts rendered line endings so later paragraphs add only
// the missing separation. ANSI controls and writer padding can occur between
// those endings, so ignore them instead of inspecting the raw suffix.
func trailingLineBreaks(s string) int {
	s = ansi.Strip(s)
	count := 0
	for i := len(s); i > 0; count++ {
		for i > 0 && (s[i-1] == ' ' || s[i-1] == '\t' || s[i-1] == '\r') {
			i--
		}
		if i == 0 || s[i-1] != '\n' {
			break
		}
		i--
	}
	return count
}
