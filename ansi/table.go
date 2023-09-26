package ansi

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// A TableElement is used to render tables.
type TableElement struct {
	lipgloss    *table.Table
	styleWriter *StyleWriter
	headers     []string
	row        []string
}

// A TableRowElement is used to render a single row in a table.
type TableRowElement struct{}

// A TableHeadElement is used to render a table's head element.
type TableHeadElement struct{}

// A TableCellElement is used to render a single cell in a row.
type TableCellElement struct {
	Text string
	Head bool
}

func (e *TableElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	// TODO add indentation and margin for the table
	// 	var indentation uint
	// 	var margin uint
	rules := ctx.options.Styles.Table
	// 	if rules.Indent != nil {
	// 		indentation = *rules.Indent
	// 	}
	// 	if rules.Margin != nil {
	// 		margin = *rules.Margin
	// 	}

	// iw := indent.NewWriterPipe(w, indentation+margin, func(wr io.Writer) {
	// 	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, " ")
	// })

	style := bs.With(rules.StylePrimitive)
	ctx.table.styleWriter = NewStyleWriter(ctx, w, style)

	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockPrefix)
	renderText(w, ctx.options.ColorProfile, style, rules.Prefix)
	// 	ctx.table.writer = tablewriter.NewWriter(ctx.table.styleWriter)
	ctx.table.lipgloss = table.New()
	return nil
}

func (e *TableElement) Finish(w io.Writer, ctx RenderContext) error {
	rules := ctx.options.Styles.Table

	ctx.table.lipgloss.Border(lipgloss.NormalBorder())

	// TODO remove styleWriter dep; not needed with lipgloss
	ctx.table.styleWriter.Write([]byte(ctx.table.lipgloss.Render()))

	ctx.table.lipgloss = nil

	renderText(ctx.table.styleWriter, ctx.options.ColorProfile, ctx.blockStack.With(rules.StylePrimitive), rules.Suffix)
	renderText(ctx.table.styleWriter, ctx.options.ColorProfile, ctx.blockStack.Current().Style.StylePrimitive, rules.BlockSuffix)
	return ctx.table.styleWriter.Close()
}

func (e *TableRowElement) Finish(w io.Writer, ctx RenderContext) error {
	if ctx.table.lipgloss == nil {
		return nil
	}
	if len(ctx.table.row) == 0 {
		panic(fmt.Sprintf("got an empty row %#v", ctx.table.row))
	}
	ctx.table.lipgloss.Row(StringToAny(ctx.table.row)...)

	// Append the current cell to our current row?
	// Maybe we should just write to TableElement, then render our final table
	// given the data in TableElement ctx.table.writer.Append(ctx.table.cell)

	// reset working row
	ctx.table.row = []string{}
	return nil
}

func (e *TableHeadElement) Finish(w io.Writer, ctx RenderContext) error {
	if ctx.table.lipgloss == nil {
		return nil
	}

	headers := StringToAny(ctx.table.headers)
	ctx.table.lipgloss.Headers(headers...)
	ctx.table.headers = []string{}
	return nil
}

// StringToAny returns the headers as generic types for the lipgloss table.
func StringToAny(s []string) []any {
	out := make([]any, len(s))
	for i, str := range s {
		out[i] = str
	}
	fmt.Print(s)

	fmt.Print(out)
	return out
}

// TODO apply individual cell styling here if desired
func (e *TableCellElement) Render(w io.Writer, ctx RenderContext) error {
	if e.Head {
		ctx.table.headers = append(ctx.table.headers, e.Text)
	} else {
		ctx.table.row = append(ctx.table.row, e.Text)
	}

	return nil
}
