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
}

// Render renders a LinkElement.
func (e *LinkElement) Render(w io.Writer, ctx RenderContext) error {
	// Make OSC 8 hyperlink token.
	hyperlink, resetHyperlink, validURL := makeHyperlink(e.URL)
	for _, child := range e.Children {
		if r, ok := child.(StyleOverriderElementRenderer); ok { //nolint:nestif
			var b bytes.Buffer
			st := ctx.options.Styles.LinkText
			if err := r.StyleOverrideRender(&b, ctx, st); err != nil {
				return fmt.Errorf("glamour: error rendering with style: %w", err)
			}

			token := hyperlink + b.String() + resetHyperlink
			if _, err := io.WriteString(w, token); err != nil {
				return fmt.Errorf("glamour: error writing hyperlink: %w", err)
			}
		} else {
			var b bytes.Buffer
			if err := child.Render(&b, ctx); err != nil {
				return fmt.Errorf("glamour: error rendering: %w", err)
			}
			token := hyperlink + b.String() + resetHyperlink
			el := &BaseElement{
				Token: token,
				Style: ctx.options.Styles.LinkText,
			}
			if err := el.Render(w, ctx); err != nil {
				return fmt.Errorf("glamour: error rendering: %w", err)
			}
		}
	}

	if validURL {
		token := hyperlink + resolveRelativeURL(e.BaseURL, e.URL) + resetHyperlink
		el := &BaseElement{
			Token:  token,
			Prefix: " ",
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
	if validURL {                                    // if the URL only consists of an anchor, ignore it
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
