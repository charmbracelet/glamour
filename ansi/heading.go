package ansi

import (
	"bytes"
	"io"

	"github.com/charmbracelet/lipgloss"
)

// A HeadingElement is used to render headings.
type HeadingElement struct {
	Level int
	First bool
}

const (
	h1 = iota + 1
	h2
	h3
	h4
	h5
	h6
)

func (e *HeadingElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	be := BlockElement{
		Block: &bytes.Buffer{},
		Style: bs.Current().Style,
	}
	bs.Push(be)

	//	renderText(w, bs.Parent().Style.StylePrimitive.Style(), rules.BlockPrefix)
	//	renderText(bs.Current().Block, bs.Current().Style.StylePrimitive.Style(), rules.Prefix)
	return nil
}

func (e *HeadingElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	block := bs.Current().Style

	// heading defines colour, background colour of all heading elements
	// subheadings e.g. h1, h2, etc define the prefix, margin, and text styling of the element.
	// not *always true*
	heading := ctx.options.Styles.Heading
	subheading := ctx.options.Styles.Heading

	switch e.Level {
	case h1:
		subheading = ctx.options.Styles.H1
	case h2:
		subheading = ctx.options.Styles.H2
	case h3:
		subheading = ctx.options.Styles.H3
	case h4:
		subheading = ctx.options.Styles.H4
	case h5:
		subheading = ctx.options.Styles.H5
	case h6:
		subheading = ctx.options.Styles.H6
	}

	// get margin as a style to wrap string
	indent := lipgloss.NewStyle()
	if subheading.Margin != nil {
		indent = indent.MarginLeft(int(*subheading.Margin))
	}
	// apply other styles to the strings
	blockStyle := block.Style().Margin(0)
	headingStyle := heading.Style().Inherit(blockStyle)
	style := subheading.Style().Inherit(headingStyle).Margin(0)

	if !e.First {
		// renderText(w, bs.Current().Style.StylePrimitive.Style(), "\n")
		w.Write([]byte("\n"))
	}

	// TODO apply text styles to prefix and subheading text
	var styledText bytes.Buffer
	renderText(&styledText, subheading.StylePrimitive.Style(),
		subheading.Prefix+bs.Current().Block.String()+subheading.Suffix)

	w.Write([]byte(
		indent.Render(
			heading.BlockPrefix +
				style.Render(styledText.String()) +
				heading.BlockSuffix)))
	// TODO set/handle width

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}
