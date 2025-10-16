package ansi

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
)

// A LinkElement is used to render hyperlinks.
type LinkElement struct {
	BaseURL    string
	URL        string
	Children   []ElementRenderer
	SkipText   bool
	SkipHref   bool
	Title      string        // Optional title attribute from markdown
	Formatter  LinkFormatter // Custom formatter reference (nil = default behavior)
	IsAutoLink bool          // Track if this is an autolink
	IsInTable  bool          // Track table context
}

// Render renders a LinkElement.
func (e *LinkElement) Render(w io.Writer, ctx RenderContext) error {
	// Check if custom formatter is set
	if e.Formatter != nil {
		return e.renderWithFormatter(w, ctx)
	}
	// If no formatter, use default behavior
	return e.renderDefault(w, ctx)
}

// renderWithFormatter creates LinkData struct and calls the custom formatter.
func (e *LinkElement) renderWithFormatter(w io.Writer, ctx RenderContext) error {
	// Extract text from children
	text, err := extractTextFromChildren(e.Children, ctx)
	if err != nil {
		return fmt.Errorf("failed to extract text from children: %w", err)
	}

	// Create LinkData with all context
	data := LinkData{
		URL:        e.URL,
		Text:       text,
		Title:      e.Title,
		BaseURL:    e.BaseURL,
		IsAutoLink: e.IsAutoLink,
		IsInTable:  e.IsInTable,
		Children:   e.Children,
		LinkStyle:  ctx.options.Styles.Link,
		TextStyle:  ctx.options.Styles.LinkText,
	}

	// Call the custom formatter
	result, err := e.Formatter.FormatLink(data, ctx)
	if err != nil {
		return fmt.Errorf("custom formatter error: %w", err)
	}

	// Write the result
	_, err = w.Write([]byte(result))
	return err
}

// renderDefault moves existing rendering logic here for backward compatibility.
func (e *LinkElement) renderDefault(w io.Writer, ctx RenderContext) error {
	if !e.SkipText {
		if err := e.renderTextPart(w, ctx); err != nil {
			return err
		}
	}
	if !e.SkipHref {
		if err := e.renderHrefPart(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (e *LinkElement) renderTextPart(w io.Writer, ctx RenderContext) error {
	for _, child := range e.Children {
		if r, ok := child.(StyleOverriderElementRenderer); ok {
			st := ctx.options.Styles.LinkText
			if err := r.StyleOverrideRender(w, ctx, st); err != nil {
				return fmt.Errorf("glamour: error rendering with style: %w", err)
			}
		} else {
			var b bytes.Buffer
			if err := child.Render(&b, ctx); err != nil {
				return fmt.Errorf("glamour: error rendering: %w", err)
			}
			el := &BaseElement{
				Token: b.String(),
				Style: ctx.options.Styles.LinkText,
			}
			if err := el.Render(w, ctx); err != nil {
				return fmt.Errorf("glamour: error rendering: %w", err)
			}
		}
	}
	return nil
}

func (e *LinkElement) renderHrefPart(w io.Writer, ctx RenderContext) error {
	prefix := ""
	if !e.SkipText {
		prefix = " "
	}

	u, err := url.Parse(e.URL)
	if err == nil && "#"+u.Fragment != e.URL { // if the URL only consists of an anchor, ignore it
		el := &BaseElement{
			Token:  resolveRelativeURL(e.BaseURL, e.URL),
			Prefix: prefix,
			Style:  ctx.options.Styles.Link,
		}
		if err := el.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}
