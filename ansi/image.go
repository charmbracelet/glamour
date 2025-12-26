package ansi

import (
	"io"
	"strings"

	"github.com/charmbracelet/glamour/internal/images"
)

// An ImageElement is used to render images elements.
type ImageElement struct {
	Text     string
	BaseURL  string
	URL      string
	Child    ElementRenderer
	TextOnly bool
}

// Render renders an ImageElement.
func (e *ImageElement) Render(w io.Writer, ctx RenderContext) error {
	// Try to render the actual image if image rendering is enabled
	if !e.TextOnly && len(e.URL) > 0 && ctx.options.ImageProtocol != "" && ctx.options.ImageProtocol != ImageProtocolNone {
		rendered, err := e.renderImage(ctx)
		if err == nil && rendered != "" {
			// Store image data and write a placeholder that will be replaced
			// after text processing is complete. This avoids corruption of
			// terminal graphics escape sequences by word-wrapping.
			placeholder := ctx.StoreImage(rendered + "\n")
			_, _ = io.WriteString(w, placeholder)

			// Optionally render alt text below the image (through normal text processing)
			if len(e.Text) > 0 {
				style := ctx.options.Styles.ImageText
				el := &BaseElement{
					Token: e.Text,
					Style: style,
				}
				return el.Render(w, ctx)
			}
			return nil
		}
		// Fall through to text-only rendering on error
	}

	// Text-only rendering (original behavior)
	return e.renderTextOnly(w, ctx)
}

// renderImage attempts to render the image using terminal graphics protocols.
func (e *ImageElement) renderImage(ctx RenderContext) (string, error) {
	opts := images.RenderOptions{
		Protocol:    protocolToImages(ctx.options.ImageProtocol),
		BaseURL:     e.BaseURL,
		FetchRemote: ctx.options.ImageFetchRemote,
	}

	resolvedURL := resolveRelativeURL(e.BaseURL, e.URL)
	return images.LoadAndRender(resolvedURL, opts)
}

// renderTextOnly renders the image as text (alt text + URL).
func (e *ImageElement) renderTextOnly(w io.Writer, ctx RenderContext) error {
	style := ctx.options.Styles.ImageText
	if e.TextOnly {
		style.Format = strings.TrimSuffix(style.Format, " →")
	}

	// Determine what text to show
	altText := e.Text
	if len(altText) == 0 {
		altText = "[image]"
	}

	el := &BaseElement{
		Token: altText,
		Style: style,
	}
	err := el.Render(w, ctx)
	if err != nil {
		return err
	}

	if e.TextOnly {
		return nil
	}

	// Show URL or [embedded] indicator for data URIs
	if len(e.URL) > 0 {
		urlText := resolveRelativeURL(e.BaseURL, e.URL)
		if strings.HasPrefix(e.URL, "data:") {
			urlText = "[img:embedded]"
		}
		el := &BaseElement{
			Token:  urlText,
			Prefix: " ",
			Style:  ctx.options.Styles.Image,
		}
		err := el.Render(w, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// protocolToImages converts ansi.ImageProtocol to images.Protocol.
func protocolToImages(p ImageProtocol) images.Protocol {
	switch p {
	case ImageProtocolAuto:
		return images.ProtocolAuto
	case ImageProtocolKitty:
		return images.ProtocolKitty
	case ImageProtocolSixel:
		return images.ProtocolSixel
	case ImageProtocolITerm:
		return images.ProtocolITerm
	default:
		return images.ProtocolNone
	}
}
