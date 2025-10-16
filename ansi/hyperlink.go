package ansi

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// OSC 8 hyperlink escape sequences
const (
	// hyperlinkStart is the OSC 8 sequence to begin a hyperlink
	hyperlinkStart = "\x1b]8;;"
	// hyperlinkMid separates the URL from the display text
	hyperlinkMid = "\x1b\\"
	// hyperlinkEnd terminates the hyperlink sequence
	hyperlinkEnd = "\x1b]8;;\x1b\\"
)

// formatHyperlink formats text as an OSC 8 hyperlink sequence.
// The OSC 8 format is: ESC]8;;URL ESC\ TEXT ESC]8;; ESC\
// This creates a clickable hyperlink in supporting terminals where TEXT is displayed
// but clicking it navigates to URL.
//
// Parameters:
//   - text: The visible text to display (may contain ANSI styling)
//   - url: The target URL for the hyperlink
//
// Returns:
//   - string: The formatted hyperlink with OSC 8 escape sequences
//
// Example:
//
//	formatHyperlink("Click here", "https://example.com")
//	// Returns: "\x1b]8;;https://example.com\x1b\\Click here\x1b]8;;\x1b\\"
func formatHyperlink(text, url string) string {
	if url == "" {
		return text
	}
	return fmt.Sprintf("%s%s%s%s%s", hyperlinkStart, url, hyperlinkMid, text, hyperlinkEnd)
}

// supportsHyperlinks detects if the current terminal supports OSC 8 hyperlinks.
// This function examines environment variables to determine terminal capabilities.
// It checks TERM_PROGRAM first (more specific), then falls back to TERM for broader detection.
//
// Supported terminals:
//   - iTerm2 (TERM_PROGRAM=iTerm.app)
//   - VS Code (TERM_PROGRAM=vscode)
//   - Windows Terminal (TERM_PROGRAM=Windows Terminal)
//   - WezTerm (TERM_PROGRAM=WezTerm)
//   - Terminals with xterm-256color or similar TERM values
//
// Parameters:
//   - ctx: RenderContext containing rendering options and state
//
// Returns:
//   - bool: true if the terminal supports OSC 8 hyperlinks, false otherwise
func supportsHyperlinks(ctx RenderContext) bool {
	// Check TERM_PROGRAM first - this is more specific and reliable
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram != "" {
		supportingPrograms := map[string]bool{
			"iTerm.app":        true, // iTerm2
			"vscode":           true, // VS Code integrated terminal
			"Windows Terminal": true, // Windows Terminal
			"WezTerm":          true, // WezTerm
			"Hyper":            true, // Hyper terminal
		}
		if supported, exists := supportingPrograms[termProgram]; exists {
			return supported
		}
	}

	// Fall back to checking TERM environment variable
	// Many modern terminals support OSC 8 even if TERM_PROGRAM isn't set
	term := os.Getenv("TERM")
	if term != "" {
		// Common terminal types that support hyperlinks
		supportingTerms := []string{
			"xterm-256color",
			"screen-256color",
			"tmux-256color",
			"alacritty",
			"xterm-kitty",
		}

		for _, supportedTerm := range supportingTerms {
			if strings.Contains(term, supportedTerm) {
				return true
			}
		}
	}

	// Check for terminal-specific environment variables
	// Some terminals set their own identification variables
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true // Kitty terminal
	}

	if os.Getenv("ALACRITTY_LOG") != "" || os.Getenv("ALACRITTY_SOCKET") != "" {
		return true // Alacritty terminal
	}

	// Conservative default: assume no hyperlink support
	return false
}

// ansiEscapeRegex matches ANSI escape sequences for removal
// This regex pattern matches:
// - CSI sequences: ESC[ followed by parameters and final byte
// - OSC sequences: ESC] followed by content and terminator
// - Simple escape sequences: ESC followed by a single character
var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[a-zA-Z]`)

// stripANSISequences removes ANSI escape sequences from text to extract plain text.
// This function is essential for extracting readable text content from styled terminal output.
// It handles various ANSI sequence types including:
//   - CSI (Control Sequence Introducer) sequences for colors, cursor movement, etc.
//   - OSC (Operating System Command) sequences for hyperlinks, titles, etc.
//   - Simple escape sequences
//
// Parameters:
//   - text: Input text that may contain ANSI escape sequences
//
// Returns:
//   - string: Clean text with all ANSI sequences removed
//
// Example:
//
//	stripANSISequences("\x1b[31mRed Text\x1b[0m")
//	// Returns: "Red Text"
func stripANSISequences(text string) string {
	if text == "" {
		return text
	}
	return ansiEscapeRegex.ReplaceAllString(text, "")
}

// extractTextFromChildren extracts plain text from a slice of ElementRenderer children.
// This function recursively renders child elements and strips ANSI sequences to produce
// clean text suitable for use in hyperlinks or other contexts where plain text is needed.
//
// The function handles nested elements by:
// 1. Rendering each child element with the provided context
// 2. Extracting the rendered output
// 3. Stripping ANSI escape sequences to get plain text
// 4. Concatenating all text results
//
// Parameters:
//   - children: Slice of ElementRenderer objects to process
//   - ctx: RenderContext for rendering the elements
//
// Returns:
//   - string: Concatenated plain text from all children
//   - error: Any error encountered during rendering
//
// Example usage in link processing:
//
//	text, err := extractTextFromChildren(linkElement.children, renderCtx)
//	if err != nil {
//	    return "", fmt.Errorf("failed to extract link text: %w", err)
//	}
func extractTextFromChildren(children []ElementRenderer, ctx RenderContext) (string, error) {
	if len(children) == 0 {
		return "", nil
	}

	var textBuffer bytes.Buffer

	for i, child := range children {
		if child == nil {
			continue // Skip nil children gracefully
		}

		// Render the child element to capture its output
		var childBuffer bytes.Buffer
		if err := child.Render(&childBuffer, ctx); err != nil {
			return "", fmt.Errorf("failed to render child element %d: %w", i, err)
		}

		// Extract the rendered content and strip ANSI sequences
		renderedText := childBuffer.String()
		plainText := stripANSISequences(renderedText)

		// Add the plain text to our result buffer
		textBuffer.WriteString(plainText)
	}

	return textBuffer.String(), nil
}

// applyStyleToText applies StylePrimitive formatting to text using Glamour's styling system.
// This function integrates with the existing Glamour styling infrastructure to ensure
// consistent text formatting across all components.
//
// The function leverages BaseElement rendering which handles:
//   - Color application (foreground and background)
//   - Text decorations (bold, underline, etc.)
//   - Case transformations (upper, lower, title)
//   - Prefix and suffix text
//
// Parameters:
//   - text: The text content to style
//   - style: StylePrimitive containing formatting rules
//   - ctx: RenderContext for the current rendering session
//
// Returns:
//   - string: The styled text with applied formatting
//   - error: Any error encountered during styling
//
// Example:
//
//	style := StylePrimitive{Color: stringPtr("red"), Bold: boolPtr(true)}
//	styledText, err := applyStyleToText("Hello", style, ctx)
//	// Returns: "\x1b[31;1mHello\x1b[0m" (red bold text)
func applyStyleToText(text string, style StylePrimitive, ctx RenderContext) (string, error) {
	if text == "" {
		return text, nil
	}

	// If the text already contains ANSI escape sequences (like OSC 8 hyperlinks),
	// don't process it through BaseElement as the escapeReplacer would corrupt them.
	// This preserves custom formatter output that includes intentional escape sequences.
	if containsANSISequences(text) {
		return text, nil
	}

	// Create a BaseElement with the text and style
	element := &BaseElement{
		Token: text,
		Style: style,
	}

	// Use BaseElement's render method to apply styling consistently
	var buf bytes.Buffer
	if err := element.Render(&buf, ctx); err != nil {
		return "", fmt.Errorf("failed to apply style to text %q: %w", text, err)
	}

	return buf.String(), nil
}

// Hyperlink represents a complete hyperlink with all necessary data for rendering.
// This struct encapsulates hyperlink information and provides methods for different
// rendering approaches (OSC 8, fallback, etc.).
type Hyperlink struct {
	URL   string // The target URL
	Text  string // Display text (plain, without ANSI sequences)
	Title string // Optional title attribute
}

// NewHyperlink creates a new Hyperlink from the given parameters.
// It automatically strips ANSI sequences from the text to ensure clean display.
//
// Parameters:
//   - url: The target URL for the hyperlink
//   - text: Display text (may contain ANSI sequences)
//   - title: Optional title attribute
//
// Returns:
//   - *Hyperlink: A new Hyperlink instance
func NewHyperlink(url, text, title string) *Hyperlink {
	return &Hyperlink{
		URL:   strings.TrimSpace(url),
		Text:  stripANSISequences(strings.TrimSpace(text)),
		Title: strings.TrimSpace(title),
	}
}

// RenderOSC8 renders the hyperlink using OSC 8 escape sequences.
// This creates a clickable link in supporting terminals.
//
// Returns:
//   - string: The hyperlink formatted with OSC 8 sequences
func (h *Hyperlink) RenderOSC8() string {
	return formatHyperlink(h.Text, h.URL)
}

// RenderPlain renders the hyperlink as plain text with URL.
// This provides a fallback for terminals that don't support OSC 8.
//
// Returns:
//   - string: The hyperlink in "text (url)" format
func (h *Hyperlink) RenderPlain() string {
	if h.Text == "" {
		return h.URL
	}
	if h.URL == "" {
		return h.Text
	}
	return fmt.Sprintf("%s (%s)", h.Text, h.URL)
}

// RenderSmart renders the hyperlink using the best method for the current terminal.
// It uses OSC 8 for supporting terminals and falls back to plain text otherwise.
//
// Parameters:
//   - ctx: RenderContext to determine terminal capabilities
//
// Returns:
//   - string: The appropriately formatted hyperlink
func (h *Hyperlink) RenderSmart(ctx RenderContext) string {
	if supportsHyperlinks(ctx) {
		return h.RenderOSC8()
	}
	return h.RenderPlain()
}

// Validate checks if the hyperlink has valid content.
// A valid hyperlink should have either a URL or display text.
//
// Returns:
//   - error: An error if the hyperlink is invalid, nil otherwise
func (h *Hyperlink) Validate() error {
	if h.URL == "" && h.Text == "" {
		return fmt.Errorf("hyperlink must have either URL or text")
	}
	return nil
}
