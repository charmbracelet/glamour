package ansi

import (
	"bytes"
	"fmt"
	"net/url"
	"slices"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/slice"
	"github.com/yuin/goldmark/ast"
	astext "github.com/yuin/goldmark/extension/ast"
)

type tableLink struct {
	href    string
	title   string
	content string
}

type groupedTableLinks map[string][]tableLink

type linkType int

const (
	_ linkType = iota
	linkTypeAuto
	linkTypeImage
	linkTypeRegular
)

func (e *TableElement) printTableLinks(ctx RenderContext) {
	if !e.shouldPrintTableLinks(ctx) {
		return
	}

	w := ctx.blockStack.Current().Block
	termWidth := int(ctx.blockStack.Width(ctx)) //nolint: gosec

	renderLinkText := func(link tableLink, linkType linkType) string {
		style := ctx.options.Styles.LinkText

		var token string
		switch linkType {
		case linkTypeAuto:
			token = linkWithSuffix(link, ctx.table.groupedAutoLinks)
		case linkTypeImage:
			token = linkWithSuffix(link, ctx.table.groupedImages)
			style = ctx.options.Styles.ImageText
		case linkTypeRegular:
			token = linkWithSuffix(link, ctx.table.groupedLinks)
		}

		el := &BaseElement{Token: token, Style: style}
		_ = el.Render(w, ctx)

		return token
	}

	renderLinkHref := func(link tableLink, linkType linkType, linkText string) {
		style := ctx.options.Styles.Link
		if linkType == linkTypeImage {
			style = ctx.options.Styles.Image
		}

		// XXX(@andreynering): Once #411 is merged, use the hyperlink
		// protocol to make the link work for the full URL even if we
		// show it truncated.
		linkMaxWidth := max(termWidth-len(linkText)-1, 0)
		token := xansi.Truncate(link.href, linkMaxWidth, "…")

		el := &BaseElement{Token: token, Style: style}
		_ = el.Render(w, ctx)
	}

	renderString := func(str string) {
		renderText(w, ctx.options.ColorProfile, ctx.blockStack.Current().Style.StylePrimitive, str)
	}

	if len(ctx.table.tableAutoLinks) > 0 || len(ctx.table.tableLinks) > 0 {
		renderString("\n")
	}
	for _, link := range ctx.table.tableAutoLinks {
		renderString("\n")
		linkText := renderLinkText(link, linkTypeAuto)
		renderString(" ")
		renderLinkHref(link, linkTypeAuto, linkText)
	}
	for _, link := range ctx.table.tableLinks {
		renderString("\n")
		linkText := renderLinkText(link, linkTypeRegular)
		renderString(" ")
		renderLinkHref(link, linkTypeRegular, linkText)
	}

	if len(ctx.table.tableImages) > 0 {
		renderString("\n")
	}
	for _, image := range ctx.table.tableImages {
		renderString("\n")
		linkText := renderLinkText(image, linkTypeImage)
		renderString(" ")
		renderLinkHref(image, linkTypeImage, linkText)
	}
}

func (e *TableElement) shouldPrintTableLinks(ctx RenderContext) bool {
	if ctx.options.InlineTableLinks {
		return false
	}
	if len(ctx.table.tableAutoLinks) == 0 && len(ctx.table.tableLinks) == 0 && len(ctx.table.tableImages) == 0 {
		return false
	}
	return true
}

func (e *TableElement) collectLinksAndImages(ctx RenderContext) error {
	autoLinks := make([]tableLink, 0)
	images := make([]tableLink, 0)
	links := make([]tableLink, 0)

	err := ast.Walk(e.table, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.AutoLink:
			uri := string(n.URL(e.source))
			autoLink := tableLink{
				href:    uri,
				content: linkDomain(uri),
			}
			autoLinks = append(autoLinks, autoLink)
		case *ast.Image:
			content, err := nodeContent(node, e.source)
			if err != nil {
				return ast.WalkStop, err
			}
			image := tableLink{
				href:    string(n.Destination),
				title:   string(n.Title),
				content: string(content),
			}
			if image.content == "" {
				image.content = linkDomain(image.href)
			}
			images = append(images, image)
		case *ast.Link:
			content, err := nodeContent(node, e.source)
			if err != nil {
				return ast.WalkStop, err
			}
			link := tableLink{
				href:    string(n.Destination),
				title:   string(n.Title),
				content: string(content),
			}
			links = append(links, link)
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return fmt.Errorf("glamour: error collecting links: %w", err)
	}

	ctx.table.tableAutoLinks = autoLinks
	ctx.table.tableImages = images
	ctx.table.tableLinks = links
	return nil
}

func (e *TableElement) uniqAndGroupLinks(ctx RenderContext) {
	groupByContentFunc := func(l tableLink) string { return l.content }

	// auto links
	ctx.table.tableAutoLinks = slice.Uniq(ctx.table.tableAutoLinks)
	ctx.table.groupedAutoLinks = slice.GroupBy(ctx.table.tableAutoLinks, groupByContentFunc)

	// images
	ctx.table.tableImages = slice.Uniq(ctx.table.tableImages)
	ctx.table.groupedImages = slice.GroupBy(ctx.table.tableImages, groupByContentFunc)

	// links
	ctx.table.tableLinks = slice.Uniq(ctx.table.tableLinks)
	ctx.table.groupedLinks = slice.GroupBy(ctx.table.tableLinks, groupByContentFunc)
}

func isInsideTable(node ast.Node) bool {
	parent := node.Parent()
	for parent != nil {
		switch parent.Kind() {
		case astext.KindTable, astext.KindTableHeader, astext.KindTableRow, astext.KindTableCell:
			return true
		default:
			parent = parent.Parent()
		}
	}
	return false
}

func nodeContent(node ast.Node, source []byte) ([]byte, error) {
	var builder bytes.Buffer

	var traverse func(node ast.Node) error
	traverse = func(node ast.Node) error {
		for n := node.FirstChild(); n != nil; n = n.NextSibling() {
			switch nn := n.(type) {
			case *ast.Text:
				if _, err := builder.Write(nn.Segment.Value(source)); err != nil {
					return fmt.Errorf("glamour: error writing text node: %w", err)
				}
			default:
				if err := traverse(nn); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := traverse(node); err != nil {
		return nil, err
	}

	return builder.Bytes(), nil
}

func linkDomain(href string) string {
	if uri, err := url.Parse(href); err == nil {
		return uri.Hostname()
	}
	return "link"
}

func linkWithSuffix(tl tableLink, grouped groupedTableLinks) string {
	token := tl.content
	if len(grouped[token]) < 2 {
		return token
	}
	index := slices.Index(grouped[token], tl)
	return fmt.Sprintf("%s[%d]", token, index+1)
}
