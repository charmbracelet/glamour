package ansi

import (
	"fmt"
	"html"
	"io"

	"github.com/yuin/goldmark/ast"
)

type strikethroughElement struct {
	node   ast.Node
	source []byte
}

func (e *strikethroughElement) Render(w io.Writer, ctx RenderContext) error {
	style := ctx.options.Styles.Strikethrough
	effective := ctx.blockStack.With(style)
	if hasTextTransform(effective) {
		return e.baseElement(style).Render(w, ctx)
	}
	return e.doRender(w, ctx, ctx.blockStack.Current().Style.StylePrimitive, effective)
}

func (e *strikethroughElement) StyleOverrideRender(w io.Writer, ctx RenderContext, style StylePrimitive) error {
	base := ctx.options.Styles.Strikethrough
	effective := cascadeStylePrimitives(ctx.blockStack.With(base), style)
	if hasTextTransform(effective) {
		return e.baseElement(base).StyleOverrideRender(w, ctx, style)
	}
	return e.doRender(w, ctx,
		cascadeStylePrimitives(ctx.blockStack.Current().Style.StylePrimitive, style),
		effective)
}

func (e *strikethroughElement) baseElement(style StylePrimitive) *BaseElement {
	// Text transforms operate on plain text, as they did before child rendering.
	return &BaseElement{Token: html.UnescapeString(string(e.node.Text(e.source))), Style: style} //nolint:staticcheck
}

func (e *strikethroughElement) doRender(w io.Writer, ctx RenderContext, base, style StylePrimitive) error {
	// Inherit the decoration without repeating the container's delimiters on each child.
	childStyle := style
	childStyle.BlockPrefix, childStyle.BlockSuffix = "", ""
	childStyle.Prefix, childStyle.Suffix, childStyle.Format = "", "", ""
	block := ctx.blockStack.Current()
	block.Style = StyleBlock{StylePrimitive: childStyle}
	stack := append(BlockStack(nil), (*ctx.blockStack)...)
	stack.Push(block)
	ctx.blockStack = &stack

	_, _ = renderText(w, base, style.BlockPrefix)
	_, _ = renderText(w, style, style.Prefix)
	defer func() {
		_, _ = renderText(w, style, style.Suffix)
		_, _ = renderText(w, base, style.BlockSuffix)
	}()

	tr := ANSIRenderer{context: ctx}
	for child := e.node.FirstChild(); child != nil; child = child.NextSibling() {
		r := tr.NewElement(child, e.source).Renderer
		if r != nil {
			if err := r.Render(w, ctx); err != nil {
				return fmt.Errorf("glamour: error rendering: %w", err)
			}
		}
	}
	return nil
}

func hasTextTransform(style StylePrimitive) bool {
	return style.Format != "" ||
		(style.Upper != nil && *style.Upper) ||
		(style.Lower != nil && *style.Lower) ||
		(style.Title != nil && *style.Title)
}
