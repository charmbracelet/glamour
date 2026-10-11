package ansi

import (
	"bytes"
	"io"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestIndentWriterCloseOrder guards against a use-after-close crash in the
// writer chain. IndentWriter.pw wraps IndentWriter.w, so closing w before pw
// used to tear down the writer that pw's closing style/link reset flushes
// through, panicking on a nil ANSI parser. Closing pw first avoids that.
func TestIndentWriterCloseOrder(t *testing.T) {
	var buf bytes.Buffer

	// pw stands in for the downstream PaddingWriter: a passthrough whose own
	// underlying WrapWriter gets closed when pw is closed.
	pw := newWrapCloser(&buf)
	iw := NewIndentWriter(pw, 2, nil)

	// Write content carrying an open SGR style so closing flushes a reset.
	if _, err := io.WriteString(iw, "\x1b[1mhello\n"); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Must not panic.
	if err := iw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

// wrapCloser models glamour's PaddingWriter: it forwards writes through an
// inner lipgloss WrapWriter and closes that writer on Close.
type wrapCloser struct {
	w *lipgloss.WrapWriter
}

func newWrapCloser(w *bytes.Buffer) *wrapCloser {
	return &wrapCloser{w: lipgloss.NewWrapWriter(w)}
}

func (c *wrapCloser) Write(p []byte) (int, error) { return c.w.Write(p) }

func (c *wrapCloser) Close() error { return c.w.Close() }
