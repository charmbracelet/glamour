package ansi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// MarginWriter is a Writer that applies indentation and padding around
// whatever you write to it.
type MarginWriter struct {
	w  io.Writer
	iw *IndentWriter
}

// NewMarginWriter returns a new MarginWriter.
func NewMarginWriter(ctx RenderContext, w io.Writer, rules StyleBlock) *MarginWriter {
	bs := ctx.blockStack

	var indentation uint
	var margin uint
	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	if rules.Margin != nil {
		margin = *rules.Margin
	}

	// The styled cell is invariant, so build it once here rather than
	// rederiving the style on every padding column of every line.
	padCell := styleText(rules.StylePrimitive, " ")
	pw := NewPaddingWriter(w, int(bs.Width(ctx)), func(_ io.Writer) {
		_, _ = io.WriteString(w, padCell)
	})

	ic := " "
	if rules.IndentToken != nil {
		ic = *rules.IndentToken
	}
	indentCell := styleText(bs.Parent().Style.StylePrimitive, ic)
	iw := NewIndentWriter(pw, int(indentation+margin), func(_ io.Writer) {
		_, _ = io.WriteString(w, indentCell)
	})

	return &MarginWriter{
		w:  lipgloss.NewWrapWriter(w),
		iw: iw,
	}
}

// Write writes to the margin writer and implements [io.Writer].
func (w *MarginWriter) Write(b []byte) (int, error) {
	n, err := w.iw.Write(b)
	if err != nil {
		return 0, fmt.Errorf("glamour: error writing bytes: %w", err)
	}
	return n, nil
}

// Close closes the [MarginWriter].
func (w *MarginWriter) Close() error {
	var werr error
	if c, ok := w.w.(io.WriteCloser); ok {
		werr = c.Close()
	}

	return errors.Join(werr, w.iw.Close())
}

// PaddingFunc is a function that applies padding around whatever you write to it.
type PaddingFunc = func(w io.Writer)

// PaddingWriter is a writer that applies padding around whatever you write to
// it.
type PaddingWriter struct {
	Padding int
	PadFunc PaddingFunc
	w       *lipgloss.WrapWriter
	cache   bytes.Buffer
}

// NewPaddingWriter returns a new PaddingWriter.
func NewPaddingWriter(w io.Writer, padding int, padFunc PaddingFunc) *PaddingWriter {
	return &PaddingWriter{
		Padding: padding,
		PadFunc: padFunc,
		w:       lipgloss.NewWrapWriter(w),
	}
}

// Write writes to the padding writer.
func (w *PaddingWriter) Write(p []byte) (int, error) {
	// Lines are forwarded whole rather than a rune at a time: only a
	// newline needs handling, and a newline byte can never appear inside
	// a multi-byte UTF-8 sequence, so scanning for it is UTF-8 safe
	// without decoding.
	total := len(p)
	for len(p) > 0 {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			w.cache.Write(p)
			if _, err := w.w.Write(p); err != nil {
				return 0, fmt.Errorf("glamour: error writing bytes: %w", err)
			}
			break
		}

		if i > 0 {
			w.cache.Write(p[:i])
			if _, err := w.w.Write(p[:i]); err != nil {
				return 0, fmt.Errorf("glamour: error writing bytes: %w", err)
			}
		}

		linew := ansi.StringWidth(w.cache.String())
		if w.Padding > 0 && linew < w.Padding {
			if w.PadFunc != nil {
				for n := 0; n < w.Padding-linew; n++ {
					w.PadFunc(w.w)
				}
			} else {
				if _, err := io.WriteString(w.w, strings.Repeat(" ", w.Padding-linew)); err != nil {
					return 0, fmt.Errorf("glamour: error writing padding: %w", err)
				}
			}
		}
		w.cache.Reset()

		if _, err := w.w.Write(p[i : i+1]); err != nil {
			return 0, fmt.Errorf("glamour: error writing bytes: %w", err)
		}
		p = p[i+1:]
	}

	return total, nil
}

// Close closes the [PaddingWriter].
func (w *PaddingWriter) Close() error {
	return w.w.Close() //nolint:wrapcheck
}

// IndentFunc is a function that applies indentation around whatever you write to
// it.
type IndentFunc = func(w io.Writer)

// IndentWriter is a writer that applies indentation around whatever you write to
// it.
type IndentWriter struct {
	Indent     int
	IndentFunc PaddingFunc
	w          io.Writer
	pw         *lipgloss.WrapWriter
	skipIndent bool
}

// NewIndentWriter returns a new IndentWriter.
func NewIndentWriter(w io.Writer, indent int, indentFunc IndentFunc) *IndentWriter {
	return &IndentWriter{
		Indent:     indent,
		IndentFunc: indentFunc,
		pw:         lipgloss.NewWrapWriter(w),
		w:          w,
	}
}

func (w *IndentWriter) resetPen() {
	style := w.pw.Style()
	link := w.pw.Link()
	if !style.IsZero() {
		_, _ = io.WriteString(w.w, ansi.ResetStyle)
	}
	if !link.IsZero() {
		_, _ = io.WriteString(w.w, ansi.ResetHyperlink())
	}
}

func (w *IndentWriter) restorePen() {
	style := w.pw.Style()
	link := w.pw.Link()
	if !style.IsZero() {
		_, _ = io.WriteString(w.w, style.String())
	}
	if !link.IsZero() {
		_, _ = io.WriteString(w.w, ansi.SetHyperlink(link.URL, link.Params))
	}
}

// Write writes to the indentation writer.
func (w *IndentWriter) Write(p []byte) (int, error) {
	// Indentation is only ever emitted at the start of a line, so the
	// rest of the line goes downstream in one Write. A newline byte can
	// never appear inside a multi-byte UTF-8 sequence, so scanning for
	// it is UTF-8 safe without decoding.
	total := len(p)
	for len(p) > 0 {
		if !w.skipIndent {
			w.resetPen()
			if w.IndentFunc != nil {
				for j := 0; j < w.Indent; j++ {
					w.IndentFunc(w.pw)
				}
			} else {
				if _, err := io.WriteString(w.pw, strings.Repeat(" ", w.Indent)); err != nil {
					return 0, fmt.Errorf("glamour: error writing indentation: %w", err)
				}
			}

			w.skipIndent = true
			w.restorePen()
		}

		run := p
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			p = nil
		} else {
			run, p = p[:i+1], p[i+1:]
			w.skipIndent = false
		}

		if _, err := w.pw.Write(run); err != nil {
			return 0, fmt.Errorf("glamour: error writing bytes: %w", err)
		}
	}

	return total, nil
}

// Close closes the [IndentWriter].
func (w *IndentWriter) Close() error {
	// Close the wrap writer (w.pw) before the downstream writer (w.w). w.pw
	// wraps w.w, so its Close flushes a trailing style/link reset back through
	// w.w. Closing w.w first would return its parser to the pool and nil it
	// out, turning that flush into a write on a closed writer.
	werr := w.pw.Close()

	if c, ok := w.w.(io.WriteCloser); ok {
		werr = errors.Join(werr, c.Close())
	}

	return werr
}
