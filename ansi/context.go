package ansi

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// imageStore holds rendered image data for deferred insertion.
type imageStore struct {
	data map[string]string
	seq  int
}

// RenderContext holds the current rendering options and state.
type RenderContext struct {
	options Options

	blockStack *BlockStack
	table      *TableElement

	stripper *bluemonday.Policy

	// images stores rendered image escape sequences keyed by placeholder ID.
	// Uses a pointer so changes persist when context is passed by value.
	images *imageStore
}

// NewRenderContext returns a new RenderContext.
func NewRenderContext(options Options) RenderContext {
	return RenderContext{
		options:    options,
		blockStack: &BlockStack{},
		table:      &TableElement{},
		stripper:   bluemonday.StrictPolicy(),
		images:     &imageStore{data: make(map[string]string)},
	}
}

// imagePlaceholderPrefix is used to mark where images should be inserted.
// Uses Unicode private use area characters that won't appear in normal text
// and won't be split by word-wrapping.
const imagePlaceholderPrefix = "\uFFF0IMG"
const imagePlaceholderSuffix = "\uFFF1"

// StoreImage stores rendered image data and returns a placeholder string.
func (ctx RenderContext) StoreImage(data string) string {
	ctx.images.seq++
	// Use a simple numeric ID padded to fixed width
	id := fmt.Sprintf("%04d", ctx.images.seq)
	placeholder := imagePlaceholderPrefix + id + imagePlaceholderSuffix
	ctx.images.data[placeholder] = data
	return placeholder
}

// ReplaceImagePlaceholders replaces all image placeholders with actual image data.
func (ctx RenderContext) ReplaceImagePlaceholders(content string) string {
	for placeholder, data := range ctx.images.data {
		if strings.Contains(content, placeholder) {
			content = strings.Replace(content, placeholder, data, 1)
		}
	}
	return content
}

// WriteWithImageReplacement writes content, sending text through the margin writer
// but writing image data directly to rawOut to bypass text processing.
// marginPrefix is prepended before each image to maintain alignment.
func (ctx RenderContext) WriteWithImageReplacement(rawOut, marginWriter io.Writer, content string, marginPrefix string) error {
	// Process placeholders in order of appearance (not random map order)
	remaining := content
	for {
		// Find the next placeholder prefix
		idx := strings.Index(remaining, imagePlaceholderPrefix)
		if idx == -1 {
			break
		}

		// Write text before the placeholder through margin writer
		if idx > 0 {
			if _, err := marginWriter.Write([]byte(remaining[:idx])); err != nil {
				return err
			}
		}

		// Find the end of the placeholder
		suffixIdx := strings.Index(remaining[idx:], imagePlaceholderSuffix)
		if suffixIdx == -1 {
			// Malformed placeholder, skip the prefix and continue
			remaining = remaining[idx+len(imagePlaceholderPrefix):]
			continue
		}

		// Extract the full placeholder
		placeholder := remaining[idx : idx+suffixIdx+len(imagePlaceholderSuffix)]

		// Look up the image data
		imageData, ok := ctx.images.data[placeholder]
		if !ok {
			// Unknown placeholder, write it as-is
			if _, err := marginWriter.Write([]byte(placeholder)); err != nil {
				return err
			}
		} else {
			// Write margin prefix then image data directly to raw output
			// (bypass margin processing to avoid corrupting escape sequences)
			if marginPrefix != "" {
				if _, err := rawOut.Write([]byte(marginPrefix)); err != nil {
					return err
				}
			}
			if _, err := rawOut.Write([]byte(imageData)); err != nil {
				return err
			}
		}

		// Continue with remaining content
		remaining = remaining[idx+len(placeholder):]
	}

	// Write any remaining text through margin writer
	if len(remaining) > 0 {
		if _, err := marginWriter.Write([]byte(remaining)); err != nil {
			return err
		}
	}

	return nil
}

// HasImages returns true if there are any stored images.
func (ctx RenderContext) HasImages() bool {
	return len(ctx.images.data) > 0
}

// SanitizeHTML sanitizes HTML content.
func (ctx RenderContext) SanitizeHTML(s string, trimSpaces bool) string {
	s = ctx.stripper.Sanitize(s)
	if trimSpaces {
		s = strings.TrimSpace(s)
	}

	return html.UnescapeString(s)
}
