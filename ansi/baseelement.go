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
	Style  scrapbook.Styler
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
	bs := ctx.blockStack
	parentBlock := bs.Current()

	renderText(w, parentBlock.Style, e.Prefix)
	defer func() {
		renderText(w, parentBlock.Style, e.Suffix)
	}()

	// TODO handle prefix/suffix

	s := e.Token
	renderText(w, e.Style, s)
	return nil
}
