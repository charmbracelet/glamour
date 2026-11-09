package ansi

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/yuin/goldmark/ast"
	astext "github.com/yuin/goldmark/extension/ast"
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

	return el.Render(w, ctx)
}

// An ItemParagraphElement separates continuation paragraphs inside a list item
// from the item's leading paragraph.
type ItemParagraphElement struct {
	Indent string
}

// Render renders an ItemParagraphElement.
func (e *ItemParagraphElement) Render(w io.Writer, ctx RenderContext) error {
	_, err := renderText(w, ctx.blockStack.Current().Style.StylePrimitive, "\n"+e.Indent)
	if err != nil {
		return fmt.Errorf("glamour: error writing list item paragraph: %w", err)
	}

	return nil
}

func itemContinuationIndent(node ast.Node, ctx RenderContext) string {
	return strings.Repeat(" ", xansi.StringWidth(itemMarker(node, ctx)))
}

func itemMarker(node ast.Node, ctx RenderContext) string {
	if node.Parent() == nil || node.Parent().Kind() != ast.KindListItem {
		return ""
	}

	item := node.Parent()
	if item.FirstChild() != nil &&
		item.FirstChild().FirstChild() != nil &&
		item.FirstChild().FirstChild().Kind() == astext.KindTaskCheckBox {
		checkbox := item.FirstChild().FirstChild().(*astext.TaskCheckBox)
		marker := ctx.options.Styles.Task.Unticked
		if checkbox.IsChecked {
			marker = ctx.options.Styles.Task.Ticked
		}
		return marker + ctx.options.Styles.Task.BlockPrefix + ctx.options.Styles.Task.Prefix
	}

	list := item.Parent().(*ast.List)
	if list.IsOrdered() {
		enumeration := itemEnumeration(item)
		if list.Start != 1 {
			enumeration += uint(list.Start) - 1
		}
		return strconv.FormatUint(uint64(enumeration), 10) + ctx.options.Styles.Enumeration.BlockPrefix + ctx.options.Styles.Enumeration.Prefix
	}

	return ctx.options.Styles.Item.BlockPrefix + ctx.options.Styles.Item.Prefix
}

func itemEnumeration(node ast.Node) uint {
	var enumeration uint = 1
	for node.PreviousSibling() != nil && node.PreviousSibling().Kind() == ast.KindListItem {
		enumeration++
		node = node.PreviousSibling()
	}

	return enumeration
}
