package ansi

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// RenderContext holds the current rendering options and state.
type RenderContext struct {
	options Options

	blockStack *BlockStack
	table      *TableElement

	stripper *bluemonday.Policy

	// pendingImages holds graphics protocol sequences to be written after the
	// current paragraph block has been fully rendered.
	pendingImages *[]string

	// graphicsCommands holds out-of-band graphics protocol sequences, i.e.
	// commands that must be written to the terminal before the rendered
	// document is displayed. These are not part of the rendered document and
	// must be retrieved via [ANSIRenderer.GraphicsCommands] after rendering.
	graphicsCommands *[]string
}

// NewRenderContext returns a new RenderContext.
func NewRenderContext(options Options) RenderContext {
	return RenderContext{
		options:          options,
		blockStack:       &BlockStack{},
		table:            &TableElement{},
		stripper:         bluemonday.StrictPolicy(),
		pendingImages:    &[]string{},
		graphicsCommands: &[]string{},
	}
}

// SanitizeHTML sanitizes HTML content.
func (ctx RenderContext) SanitizeHTML(s string, trimSpaces bool) string {
	s = ctx.stripper.Sanitize(s)
	if trimSpaces {
		s = strings.TrimSpace(s)
	}

	return html.UnescapeString(s)
}

// flushPendingImages writes any pending graphics protocol sequences to w.
func (ctx *RenderContext) flushPendingImages(w io.Writer) error {
	for _, seq := range *ctx.pendingImages {
		if _, err := io.WriteString(w, seq); err != nil {
			return fmt.Errorf("glamour: error writing graphics sequence: %w", err)
		}
	}
	*ctx.pendingImages = (*ctx.pendingImages)[:0]
	return nil
}
