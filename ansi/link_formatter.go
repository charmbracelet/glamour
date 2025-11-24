package ansi

import (
	"fmt"
	"net/url"
	"strings"
)

// LinkData contains all parsed link information available to formatters.
// It provides formatters with comprehensive context about the link being rendered,
// including styling, positioning context, and original markdown elements.
type LinkData struct {
	// Basic link properties
	URL     string // The destination URL
	Text    string // The link text (extracted from children)
	Title   string // Optional title attribute from markdown
	BaseURL string // Base URL for relative link resolution

	// Formatting context
	IsAutoLink bool              // Whether this is an autolink (e.g., <https://example.com>)
	IsInTable  bool              // Whether link appears in a table context
	Children   []ElementRenderer // Original child elements for advanced rendering

	// Style context
	LinkStyle StylePrimitive // Style configuration for the URL portion
	TextStyle StylePrimitive // Style configuration for the text portion
}

// LinkFormatter defines how links should be rendered.
// Custom formatters implement this interface to provide alternative link rendering.
// The FormatLink method receives complete link context and should return the
// formatted string representation of the link.
type LinkFormatter interface {
	// FormatLink renders a link using the provided data and context.
	// It returns the formatted link string and any formatting error.
	FormatLink(data LinkData, ctx RenderContext) (string, error)
}

// LinkFormatterFunc is an adapter type that allows functions to implement LinkFormatter.
// This enables convenient creation of link formatters using function literals.
type LinkFormatterFunc func(LinkData, RenderContext) (string, error)

// FormatLink implements the LinkFormatter interface for LinkFormatterFunc.
func (f LinkFormatterFunc) FormatLink(data LinkData, ctx RenderContext) (string, error) {
	return f(data, ctx)
}

// Built-in Link Formatters
//
// These formatters provide common link rendering patterns and serve as examples
// for custom formatter implementations.

// DefaultFormatter replicates the current Glamour link rendering behavior.
// It renders links in the format "text url" with appropriate styling applied.
// This formatter maintains backward compatibility with existing Glamour output.
var DefaultFormatter = LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
	var result strings.Builder

	// Render text part if present
	if data.Text != "" {
		styledText, err := applyStyleToText(data.Text, data.TextStyle, ctx)
		if err != nil {
			return "", fmt.Errorf("failed to apply text style: %w", err)
		}
		result.WriteString(styledText)
	}

	// Render URL part with space prefix if text exists
	if data.URL != "" && !isFragmentOnlyURL(data.URL) {
		if data.Text != "" {
			result.WriteString(" ")
		}

		resolvedURL := resolveRelativeURL(data.BaseURL, data.URL)
		styledURL, err := applyStyleToText(resolvedURL, data.LinkStyle, ctx)
		if err != nil {
			return "", fmt.Errorf("failed to apply link style: %w", err)
		}
		result.WriteString(styledURL)
	}

	return result.String(), nil
})

// TextOnlyFormatter shows only the link text, making it clickable in smart terminals.
// In terminals that support OSC 8 hyperlinks, the text becomes a clickable hyperlink.
// In other terminals, only the styled text is shown without the URL.
var TextOnlyFormatter = LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
	if data.Text == "" {
		return "", nil
	}

	styledText, err := applyStyleToText(data.Text, data.TextStyle, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to apply text style: %w", err)
	}

	// Make text clickable in supporting terminals
	if supportsHyperlinks(ctx) {
		return formatHyperlink(styledText, data.URL), nil
	}

	return styledText, nil
})

// URLOnlyFormatter shows only URLs, hiding the link text.
// This formatter is useful for cases where space is limited or when
// the URL itself is more important than descriptive text.
var URLOnlyFormatter = LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
	if data.URL == "" || isFragmentOnlyURL(data.URL) {
		return "", nil
	}

	resolvedURL := resolveRelativeURL(data.BaseURL, data.URL)
	styledURL, err := applyStyleToText(resolvedURL, data.LinkStyle, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to apply link style: %w", err)
	}

	return styledURL, nil
})

// HyperlinkFormatter renders links as OSC 8 hyperlinks in supporting terminals.
// The link text becomes clickable, while the URL remains hidden.
// In terminals without OSC 8 support, this formatter will not provide fallback
// and may result in escape sequences being displayed.
var HyperlinkFormatter = LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
	if data.Text == "" {
		return "", nil
	}

	styledText, err := applyStyleToText(data.Text, data.TextStyle, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to apply text style: %w", err)
	}

	return formatHyperlink(styledText, data.URL), nil
})

// SmartHyperlinkFormatter renders OSC 8 hyperlinks with intelligent fallback.
// In terminals that support hyperlinks, it shows clickable text.
// In other terminals, it falls back to the default "text url" format.
// This provides the best user experience across different terminal environments.
var SmartHyperlinkFormatter = LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
	if supportsHyperlinks(ctx) {
		return HyperlinkFormatter.FormatLink(data, ctx)
	}
	return DefaultFormatter.FormatLink(data, ctx)
})

// Helper Functions
//
// These functions will be implemented in the hyperlink support file but are
// referenced here to define the formatter interfaces clearly.

// Helper functions are implemented in ansi/hyperlink.go

// isFragmentOnlyURL checks if a URL consists only of a fragment (anchor).
// This replicates the logic from the original LinkElement.renderHrefPart method:
// if err == nil && "#"+u.Fragment != e.URL { // if the URL only consists of an anchor, ignore it
func isFragmentOnlyURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false // If can't parse, treat as normal URL
	}
	// Original logic: render if "#"+u.Fragment != e.URL
	// So fragment-only if "#"+u.Fragment == e.URL
	return "#"+u.Fragment == urlStr
}
