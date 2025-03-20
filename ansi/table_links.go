package ansi

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark/ast"
	astext "github.com/yuin/goldmark/extension/ast"
)

type tableLink struct {
	href    string
	title   string
	content string
}

func (e *TableElement) printTableLinks(ctx RenderContext) error {
	if !ctx.options.AccessibleTableLinks {
		return nil
	}

	links, err := e.collectLinks()
	if err != nil {
		return err
	}
	if len(links) == 0 {
		return nil
	}

	w := ctx.blockStack.Current().Block

	renderLinkText := func(content string) {
		el := &BaseElement{
			Token: content,
			Style: ctx.options.Styles.LinkText,
		}
		_ = el.Render(w, ctx)
	}

	renderLinkHref := func(href string) {
		el := &BaseElement{
			Token: href,
			Style: ctx.options.Styles.Link,
		}
		_ = el.Render(w, ctx)
	}

	renderString := func(str string) {
		renderText(w, ctx.options.ColorProfile, ctx.blockStack.Current().Style.StylePrimitive, str)
	}

	renderString("\n")

	for _, link := range links {
		renderString("\n")
		renderLinkText(link.content)
		renderString(" ")
		renderLinkHref(link.href)
	}

	return nil
}

func (e *TableElement) collectLinks() ([]tableLink, error) {
	links := make([]tableLink, 0)

	err := ast.Walk(e.table, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		linkNode, ok := node.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}

		content, err := e.nodeContent(node)
		if err != nil {
			return ast.WalkStop, err
		}

		link := tableLink{
			href:    string(linkNode.Destination),
			title:   string(linkNode.Title),
			content: string(content),
		}
		links = append(links, link)

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, fmt.Errorf("glamour: error collecting links: %w", err)
	}
	return links, nil
}

func (e *TableElement) nodeContent(node ast.Node) ([]byte, error) {
	var builder bytes.Buffer
	for n := node.FirstChild(); n != nil; n = n.NextSibling() {
		textNode, ok := n.(*ast.Text)
		if !ok {
			continue
		}
		if _, err := builder.Write(textNode.Segment.Value(e.source)); err != nil {
			return nil, fmt.Errorf("glamour: error writing text node: %w", err)
		}
	}
	return builder.Bytes(), nil
}

func isInsideTable(node ast.Node) bool {
	parent := node.Parent()
	for parent != nil {
		switch parent.Kind() {
		case astext.KindTable, astext.KindTableHeader, astext.KindTableRow, astext.KindTableCell:
			return true
		}
		parent = parent.Parent()
	}
	return false
}
