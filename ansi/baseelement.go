package ansi

import (
	"io"
	"strings"

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

func renderText(w io.Writer, styler scrapbook.Styler, s string) {
	if len(s) == 0 {
		return
	}
	s = strings.ReplaceAll(s, "\n", " ")
	if styler != nil {
		// styler is nil if we get a type of BaseElement with no styles.
		s = styler.Style().Render(s)
	}
	_, _ = w.Write([]byte(s))
}

func (e *BaseElement) Render(w io.Writer, ctx RenderContext) error {
	block := ctx.blockStack.Current()

	renderText(w, block.Style, e.Prefix)
	defer func() {
		renderText(w, block.Style, e.Suffix)
	}()

	// We don't carry the text styles over to the prefixes. Also, don't make it
	// a block style, we don't want margins and newlines applied to this text.
	renderText(w, block.Style.StylePrimitive, e.Style.BlockPrefix)
	defer func() {
		renderText(w, block.Style.StylePrimitive, e.Style.BlockSuffix)
	}()

	s := e.Token
	renderText(w, e.Style, s)
	return nil
}
