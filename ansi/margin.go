package ansi

import (
	"io"

	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/padding"
	"github.com/muesli/termenv"
)

// MarginWriter is a Writer that applies indentation and padding around
// whatever you write to it.
type MarginWriter struct {
	indentation, margin  uint
	indentPos, marginPos uint
	indentToken          string

	profile            termenv.Profile
	rules, parentRules StylePrimitive

	w  io.Writer
	pw *padding.Writer
	iw *indent.Writer
}

// NewMarginWriter returns a new MarginWriter.
func NewMarginWriter(ctx RenderContext, w io.Writer, rules StyleBlock, padded bool) *MarginWriter {
	bs := ctx.blockStack
	mw := &MarginWriter{
		w:           w,
		profile:     ctx.options.ColorProfile,
		rules:       rules.StylePrimitive,
		parentRules: bs.Parent().Style.StylePrimitive,
	}

	if rules.Indent != nil {
		mw.indentation = *rules.Indent
		mw.indentToken = " "
		if rules.IndentToken != nil {
			mw.indentToken = *rules.IndentToken
		}
	}
	if rules.Margin != nil {
		mw.margin = *rules.Margin
	}

	fwd := mw.w
	if padded {
		mw.pw = padding.NewWriterPipe(mw.w, bs.Width(ctx), func(wr io.Writer) {
			renderText(mw.w, mw.profile, mw.rules, "")
		})
		fwd = mw.pw
	}

	mw.iw = indent.NewWriterPipe(fwd, mw.indentation+(mw.margin*2), mw.indentFunc)
	return mw
}

func (w *MarginWriter) Write(b []byte) (int, error) {
	return w.iw.Write(b)
}

// indentFunc is called when writing each the margin and indentation tokens.
// The margin is written first, using an empty space character as the token.
// The indentation is written next, using the token specified in the rules.
func (w *MarginWriter) indentFunc(iw io.Writer) {
	ic := " "
	switch {
	case w.margin == 0 && w.indentation == 0:
		return
	case w.margin >= 1 && w.indentation == 0:
		break
	case w.margin >= 1 && w.marginPos < w.margin:
		w.marginPos++
	case w.indentation >= 1 && w.indentPos < w.indentation:
		w.indentPos++
		ic = w.indentToken
		if w.indentPos == w.indentation {
			w.marginPos = 0
			w.indentPos = 0
		}
	}
	renderText(w.w, w.profile, w.parentRules, ic)
}
