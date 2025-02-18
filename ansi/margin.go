package ansi

import (
	"bytes"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// MarginWriter is a Writer that applies indentation and padding around
// whatever you write to it.
type MarginWriter struct {
	w     io.Writer
	rules StyleBlock
	ctx   RenderContext
}

// NewMarginWriter returns a new MarginWriter.
func NewMarginWriter(ctx RenderContext, w io.Writer, rules StyleBlock) *MarginWriter {
	return &MarginWriter{
		w:     w,
		ctx:   ctx,
		rules: rules,
	}
}

func (w *MarginWriter) Write(b []byte) (n int, err error) {
	bs := w.ctx.blockStack

	var indentation int
	var margin int
	if w.rules.Indent != nil {
		indentation = int(*w.rules.Indent)
	}
	if w.rules.Margin != nil {
		margin = int(*w.rules.Margin)
	}

	ic := " "
	if w.rules.IndentToken != nil {
		ic = *w.rules.IndentToken
	}

	var buf bytes.Buffer
	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		indentStr := strings.Repeat(ic, indentation+margin)
		renderText(&buf, bs.Parent().Style.StylePrimitive, indentStr)
		buf.WriteString(line)
		if repeat := int(bs.Width(w.ctx)) - ansi.StringWidth(line); repeat > 0 {
			padStr := strings.Repeat(" ", repeat)
			renderText(&buf, w.rules.StylePrimitive, padStr)
		}
		if i < len(lines)-1 {
			buf.WriteByte('\n')
		}
	}

	return w.w.Write(buf.Bytes())
}
