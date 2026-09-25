package ansi

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"net/url"

	"github.com/charmbracelet/x/ansi"
)

// A LinkElement is used to render hyperlinks.
type LinkElement struct {
	BaseURL  string
	URL      string
	Children []ElementRenderer
	SkipText bool
	SkipHref bool

	hyperlink, resetHyperlink string
	validURL                  bool
}

// Render renders a LinkElement.
func (e *LinkElement) Render(w io.Writer, ctx RenderContext) error {
	// Make OSC 8 hyperlink token.
	e.hyperlink, e.resetHyperlink, e.validURL = makeHyperlink(e.URL)

	// When inline hyperlinks are enabled and the URL is valid, render only
	// the link text (underlined, hyperlinked) and hide the URL.
	inline := ctx.options.HyperlinkMode == HyperlinkModeInline && e.validURL && !e.SkipText

	if err := e.renderTextPart(w, ctx, inline); err != nil {
		return err
	}
	if !inline && !e.SkipHref {
		if err := e.renderHrefPart(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (e *LinkElement) renderTextPart(w io.Writer, ctx RenderContext, inline bool) error {
	style := ctx.options.Styles.LinkText
	if inline {
		underline := true
		style.Underline = &underline
	}

	if e.SkipText {
		return nil
	}

	for _, child := range e.Children {
		if child == nil {
			continue
		}
		if r, ok := child.(StyleOverriderElementRenderer); ok { //nolint:nestif
			var b bytes.Buffer
			if err := r.StyleOverrideRender(&b, ctx, style); err != nil {
				return fmt.Errorf("glamour: error rendering with style: %w", err)
			}

			token := e.hyperlink + b.String() + e.resetHyperlink
			if _, err := io.WriteString(w, token); err != nil {
				return fmt.Errorf("glamour: error writing hyperlink: %w", err)
			}
		} else {
			var b bytes.Buffer
			if err := child.Render(&b, ctx); err != nil {
				return fmt.Errorf("glamour: error rendering: %w", err)
			}
			token := e.hyperlink + b.String() + e.resetHyperlink
			el := &BaseElement{
				Token: token,
				Style: style,
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

	if e.validURL {
		token := e.hyperlink + resolveRelativeURL(e.BaseURL, e.URL) + e.resetHyperlink
		el := &BaseElement{
			Token:  token,
			Prefix: prefix,
			Style:  ctx.options.Styles.Link,
		}
		if err := el.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

// makeHyperlink takes a URL and returns an OSC 8 hyperlink token.
func makeHyperlink(link string) (string, string, bool) {
	// Make OSC 8 hyperlink token.
	var hyperlink, resetHyperlink string

	u, err := url.Parse(link)
	validURL := err == nil && "#"+u.Fragment != link // if the URL only consists of an anchor, ignore it
	if validURL {
		h := fnv.New32a()
		if _, err := io.WriteString(h, link); err != nil {
			return "", "", false
		}
		urlID := fmt.Sprintf("id=%d", h.Sum32())
		hyperlink = ansi.SetHyperlink(link, urlID)
		resetHyperlink = ansi.ResetHyperlink()
	}

	return hyperlink, resetHyperlink, validURL
}
