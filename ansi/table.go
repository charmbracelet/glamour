package ansi

import (
	"bytes"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/reflow/indent"
	astext "github.com/yuin/goldmark/extension/ast"
)

// A TableElement is used to render tables.
type TableElement struct {
	lipgloss *table.Table
	table    *astext.Table
	header   []string
	row      []string
}

// A TableRowElement is used to render a single row in a table.
type TableRowElement struct{}

// A TableHeadElement is used to render a table's head element.
type TableHeadElement struct{}

// A TableCellElement is used to render a single cell in a row.
type TableCellElement struct {
	Children []ElementRenderer
	Head     bool
}

// TODO: move this to the styles thingy?
var (
	cellStyle   = lipgloss.NewStyle().Padding(0, 1)
	headerStyle = lipgloss.NewStyle().Padding(0, 1).Bold(true)
)

func (e *TableElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	var indentation uint
	var margin uint
	rules := ctx.options.Styles.Table
	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	if rules.Margin != nil {
		margin = *rules.Margin
	}

	iw := indent.NewWriterPipe(w, indentation+margin, func(wr io.Writer) {
		renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, " ")
	})

	style := bs.With(rules.StylePrimitive)

	renderText(iw, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockPrefix)
	renderText(iw, ctx.options.ColorProfile, style, rules.Prefix)
	ctx.table.lipgloss = table.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			st := cellStyle
			if row == 0 {
				st = headerStyle
			}

			switch e.table.Alignments[col] {
			case astext.AlignLeft:
				st = st.Align(lipgloss.Left)
			case astext.AlignCenter:
				st = st.Align(lipgloss.Center)
			case astext.AlignRight:
				st = st.Align(lipgloss.Right)
			}

			return st
		}).
		Width(int(ctx.blockStack.Width(ctx)))

	return nil
}

func (e *TableElement) setBorders(ctx RenderContext) {
	rules := ctx.options.Styles.Table
	border := lipgloss.NormalBorder()

	if rules.RowSeparator != nil && rules.ColumnSeparator != nil {
		border = lipgloss.Border{
			Top:    *rules.RowSeparator,
			Bottom: *rules.RowSeparator,
			Left:   *rules.ColumnSeparator,
			Right:  *rules.ColumnSeparator,
			Middle: *rules.CenterSeparator,
		}
	}
	ctx.table.lipgloss.Border(border)
	ctx.table.lipgloss.BorderTop(false)
	ctx.table.lipgloss.BorderLeft(false)
	ctx.table.lipgloss.BorderRight(false)
	ctx.table.lipgloss.BorderBottom(false)
}

func (e *TableElement) Finish(w io.Writer, ctx RenderContext) error {
	rules := ctx.options.Styles.Table

	e.setBorders(ctx)

	ow := ctx.blockStack.Current().Block
	if _, err := ow.WriteString(ctx.table.lipgloss.String()); err != nil {
		return err
	}

	renderText(ow, ctx.options.ColorProfile, ctx.blockStack.With(rules.StylePrimitive), rules.Suffix)
	renderText(ow, ctx.options.ColorProfile, ctx.blockStack.Current().Style.StylePrimitive, rules.BlockSuffix)
	ctx.table.lipgloss = nil
	return nil
}

func (e *TableRowElement) Finish(w io.Writer, ctx RenderContext) error {
	if ctx.table.lipgloss == nil {
		return nil
	}

	ctx.table.lipgloss.Row(ctx.table.row...)
	ctx.table.row = []string{}
	return nil
}

func (e *TableHeadElement) Finish(w io.Writer, ctx RenderContext) error {
	if ctx.table.lipgloss == nil {
		return nil
	}

	ctx.table.lipgloss.Headers(ctx.table.header...)
	ctx.table.header = []string{}
	return nil
}

func (e *TableCellElement) Render(w io.Writer, ctx RenderContext) error {
	var b bytes.Buffer
	style := ctx.options.Styles.Table.StylePrimitive
	for _, child := range e.Children {
		if r, ok := child.(StyleOverriderElementRenderer); ok {
			if err := r.StyleOverrideRender(&b, ctx, style); err != nil {
				return err
			}
		} else {
			if err := child.Render(&b, ctx); err != nil {
				return err
			}
			el := &BaseElement{
				Token: b.String(),
				Style: style,
			}
			if err := el.Render(w, ctx); err != nil {
				return err
			}
		}
	}

	if e.Head {
		ctx.table.header = append(ctx.table.header, b.String())
	} else {
		ctx.table.row = append(ctx.table.row, b.String())
	}

	return nil
}
