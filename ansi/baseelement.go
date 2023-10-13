package ansi

import (
	"io"
	"strings"

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

// renderText renders chunks of text provided by a goldmark node. Every higher
// level element uses this function to style its text.
func renderText(w io.Writer, styler lipgloss.Style, s string) {
	if len(s) == 0 {
		return
	}
	// We need to strip the hard line breaks from the markdown file elements.
	// This can happen in paragraphs, link titles, and more so we need to handle
	// it here.
	s = strings.ReplaceAll(s, "\n", " ")
	s = styler.Render(s)
	_, _ = w.Write([]byte(s))
}

func (e *BaseElement) Render(w io.Writer, ctx RenderContext) error {
	block := ctx.blockStack.Current()
	// We need to inherit the styles of the block element containing this base element.
	ps := block.Style.Style()
	// The margins are dictated by block elements, not text.
	ps = ps.Margin(0)
	child := e.Style.Style().Inherit(ps)

	// We don't carry the text styles over to the prefixes. Also, don't make it
	// a block style, we don't want margins and newlines applied to this text.
	renderText(w, block.Style.Style(), e.Prefix)
	defer func() {
		renderText(w, block.Style.Style(), e.Suffix)
	}()

	renderText(w, block.Style.StylePrimitive.Style(), e.Style.BlockPrefix)
	defer func() {
		renderText(w, block.Style.StylePrimitive.Style(), e.Style.BlockSuffix)
	}()

	s := e.Token
	// rendertext using the new style
	renderText(w, child, s)
	return nil
}
