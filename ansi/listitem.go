package ansi

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/charmbracelet/x/ansi"
)

// An ItemElement is used to render items inside a list.
type ItemElement struct {
	IsOrdered   bool
	Enumeration uint
}

// Render renders an ItemElement.
func (e *ItemElement) Render(w io.Writer, ctx RenderContext) error {
	var el *BaseElement
	if e.IsOrdered {
		el = &BaseElement{
			Style:  ctx.options.Styles.Enumeration,
			Prefix: strconv.FormatInt(int64(e.Enumeration), 10), //nolint:gosec
		}
	} else {
		el = &BaseElement{
			Style: ctx.options.Styles.Item,
		}
	}

	return renderListItemPrefix(w, ctx, el)
}

// renderListItemPrefix renders and records the display width of a list marker.
func renderListItemPrefix(w io.Writer, ctx RenderContext, el *BaseElement) error {
	var prefix bytes.Buffer
	// Measure the fully styled prefix rather than inferring its visible width.
	if err := el.Render(&prefix, ctx); err != nil {
		return err
	}
	ctx.blockStack.SetPrefixWidth(ansi.StringWidth(prefix.String()))
	if _, err := prefix.WriteTo(w); err != nil {
		return fmt.Errorf("glamour: error writing list item prefix: %w", err)
	}
	return nil
}
