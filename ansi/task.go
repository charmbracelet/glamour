package ansi

import (
	"io"
)

// A TaskElement is used to render tasks inside a todo-list.
type TaskElement struct {
	Checked bool
}

// Render renders a TaskElement.
func (e *TaskElement) Render(w io.Writer, ctx RenderContext) error {
	item := &ItemElement{
		IsTask:      true,
		TaskChecked: e.Checked,
	}
	return item.Render(w, ctx)
}

// Finish finishes rendering a TaskElement.
func (e *TaskElement) Finish(w io.Writer, ctx RenderContext) error {
	item := &ItemElement{
		IsTask:      true,
		TaskChecked: e.Checked,
	}
	return item.Finish(w, ctx)
}
