package ansi

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// A TableElement is used to render tables.
type TableElement struct {
	writer      *tablewriter.Table
	header      []string
	cell        []string
}

// A TableRowElement is used to render a single row in a table.
type TableRowElement struct {
}

// A TableHeadElement is used to render a table's head element.
type TableHeadElement struct {
}

// A TableCellElement is used to render a single cell in a row.
type TableCellElement struct {
	Text string
	Head bool
}

func (e *TableElement) Render(w io.Writer, ctx RenderContext) error {
	// TODO handle prefix
	ctx.table.writer = tablewriter.NewWriter(w)
	return nil
}

func (e *TableElement) Finish(w io.Writer, ctx RenderContext) error {
	rules := ctx.options.Styles.Table

	ctx.table.writer.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	if rules.CenterSeparator != nil {
		ctx.table.writer.SetCenterSeparator(*rules.CenterSeparator)
	}
	if rules.ColumnSeparator != nil {
		ctx.table.writer.SetColumnSeparator(*rules.ColumnSeparator)
	}
	if rules.RowSeparator != nil {
		ctx.table.writer.SetRowSeparator(*rules.RowSeparator)
	}

	ctx.table.writer.Render()
	ctx.table.writer = nil

	// TODO handle suffix
	w.Write(ctx.blockStack.Current().Block.Bytes())
	return nil
}

func (e *TableRowElement) Finish(w io.Writer, ctx RenderContext) error {
	if ctx.table.writer == nil {
		return nil
	}

	ctx.table.writer.Append(ctx.table.cell)
	ctx.table.cell = []string{}
	return nil
}

func (e *TableHeadElement) Finish(w io.Writer, ctx RenderContext) error {
	if ctx.table.writer == nil {
		return nil
	}

	ctx.table.writer.SetHeader(ctx.table.header)
	ctx.table.header = []string{}
	return nil
}

func (e *TableCellElement) Render(w io.Writer, ctx RenderContext) error {
	if e.Head {
		ctx.table.header = append(ctx.table.header, e.Text)
	} else {
		ctx.table.cell = append(ctx.table.cell, e.Text)
	}

	return nil
}
