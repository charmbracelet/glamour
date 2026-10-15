package ansi

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// An ItemElement is used to render items inside a list.
type ItemElement struct {
	IsOrdered   bool
	Enumeration uint
	IsTask      bool
	TaskChecked bool
}

// Render renders an ItemElement.
func (e *ItemElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	st := bs.Current().Style
	st.Indent = nil
	st.Margin = nil
	be := BlockElement{
		Block: &bytes.Buffer{},
		Style: st,
	}
	bs.Push(be)
	return nil
}

// Finish finishes rendering an ItemElement.
func (e *ItemElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	content := bs.Current().Block.String()
	bs.Current().Block.Reset()
	bs.Pop()

	var el *BaseElement
	if e.IsTask {
		pre := ctx.options.Styles.Task.Unticked
		if e.TaskChecked {
			pre = ctx.options.Styles.Task.Ticked
		}
		el = &BaseElement{
			Prefix: pre,
			Style:  ctx.options.Styles.Task.StylePrimitive,
		}
	} else if e.IsOrdered {
		el = &BaseElement{
			Style:  ctx.options.Styles.Enumeration,
			Prefix: strconv.FormatInt(int64(e.Enumeration), 10), //nolint:gosec
		}
	} else {
		el = &BaseElement{
			Style: ctx.options.Styles.Item,
		}
	}

	var buf bytes.Buffer
	if err := el.Render(&buf, ctx); err != nil {
		return err
	}
	prefix := buf.String()
	prefixWidth := ansi.StringWidth(prefix)

	if len(content) == 0 {
		_, err := io.WriteString(w, prefix)
		return err
	}

	width := int(bs.Width(ctx))
	hangingIndent := strings.Repeat(" ", prefixWidth)

	wrapWidth := width - prefixWidth
	if wrapWidth <= 0 {
		wrapWidth = 1
	}

	// Split content into lines.
	// Lines belonging to the list item's text should be wrapped with hanging indent.
	// Nested blocks (e.g. nested lists) already have indentation and formatting applied.
	lines := strings.Split(content, "\n")
	var out strings.Builder
	firstLine := true

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// An empty trailing element from a trailing newline
		if i == len(lines)-1 && len(line) == 0 {
			break
		}

		// If this is a nested block (already indented by child margin writer)
		if !firstLine && len(line) > 0 && strings.HasPrefix(ansi.Strip(line), " ") {
			out.WriteString("\n")
			out.WriteString(line)
			continue
		}

		// Wrap the text line
		var flow string
		if width > 0 {
			flow = lipgloss.Wrap(line, wrapWidth, " ,.;-+|")
		} else {
			flow = line
		}

		wrappedLines := strings.Split(flow, "\n")
		for j, wl := range wrappedLines {
			if firstLine && j == 0 {
				out.WriteString(prefix)
				out.WriteString(wl)
				firstLine = false
			} else {
				out.WriteString("\n")
				if len(wl) > 0 {
					out.WriteString(hangingIndent)
					out.WriteString(wl)
				}
			}
		}
	}

	if strings.HasSuffix(content, "\n") && !strings.HasSuffix(out.String(), "\n") {
		out.WriteString("\n")
	}

	_, err := io.WriteString(w, out.String())
	return err
}
