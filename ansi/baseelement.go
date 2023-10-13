package ansi

import (
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/scrapbook"
)

// BaseElement renders a styled primitive element.
// TODO do I still need these elements
type BaseElement struct {
	Token  string
	Prefix string
	Suffix string
	Style  scrapbook.StylePrimitive
}

func renderText(w io.Writer, styler lipgloss.Style, s string) {
	if len(s) == 0 {
		return
	}
	s = styler.Render(s)
	_, _ = w.Write([]byte(s))
}

func (e *BaseElement) Render(w io.Writer, ctx RenderContext) error {
	block := ctx.blockStack.Current()
	// get parent styles
	// TODO is this Current or parent? We should be inheriting those styles at some point...
	ps := ctx.blockStack.Current().Style.Style()
	ps = ps.Margin(0)
	// Unset the values we don't want applied to the child (spacing).

	// inherit to child
	child := e.Style.Style().Inherit(ps)

	// Don't add text styling to filler text.
	renderText(w, block.Style.Style(), e.Prefix)
	defer func() {
		renderText(w, block.Style.Style(), e.Suffix)
	}()

	// We don't carry the text styles over to the prefixes. Also, don't make it
	// a block style, we don't want margins and newlines applied to this text.
	renderText(w, block.Style.StylePrimitive.Style(), e.Style.BlockPrefix)
	defer func() {
		renderText(w, block.Style.StylePrimitive.Style(), e.Style.BlockSuffix)
	}()

	s := e.Token
	// rendertext using the new style
	renderText(w, child, s)
	return nil
}
