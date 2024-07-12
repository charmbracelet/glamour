package ansi

import (
	"bytes"
	"html"
	"io"
)

// A EmphasisElement is used to render emphasis.
type EmphasisElement struct {
	Children []ElementRenderer
	Level    int
}

func (e *EmphasisElement) Render(w io.Writer, ctx RenderContext) error {
	style := ctx.options.Styles.Emph
	if e.Level > 1 {
		style = ctx.options.Styles.Strong
	}

	var b bytes.Buffer
	for _, child := range e.Children {
		if err := child.Render(&b, ctx); err != nil {
			return err
		}
	}

	el := Element{
		Renderer: &BaseElement{
			Token: html.UnescapeString(b.String()),
			Style: style,
		},
	}
	return el.Renderer.Render(w, ctx)
}
