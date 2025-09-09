package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

func main() {
	fmt.Println("=== TERMINAL DETECTION DEMO ===\n")

	// Sample markdown for testing
	markdown := `# Terminal Capability Detection

Test your terminal's link support with these examples:

- [GitHub](https://github.com) - Popular repository hosting
- [Google](https://google.com) - Search engine
- <https://stackoverflow.com> - Developer Q&A

Try this in different terminals to see the differences!`

	// Get terminal information
	fmt.Println("🔍 Terminal Environment Detection:")
	detectTerminalCapabilities()
	fmt.Println()

	// Test hyperlink support with different methods
	fmt.Println("🧪 Hyperlink Support Tests:")
	testHyperlinkSupport(markdown)
	fmt.Println()

	// Demonstrate adaptive formatter
	fmt.Println("🎯 Adaptive Formatter Demo:")
	demonstrateAdaptiveFormatter(markdown)
	fmt.Println()

	fmt.Println("✅ Terminal detection demo completed!")
	fmt.Println("\n💡 Try running this in different terminals:")
	fmt.Println("   • iTerm2 (macOS) - Full hyperlink support")
	fmt.Println("   • Windows Terminal - Full hyperlink support")
	fmt.Println("   • VS Code Terminal - Full hyperlink support")
	fmt.Println("   • macOS Terminal - Basic support only")
	fmt.Println("   • SSH session - Usually basic support only")
}

// detectTerminalCapabilities shows how to detect terminal capabilities
func detectTerminalCapabilities() {
	// Check environment variables that indicate terminal capabilities
	termProgram := os.Getenv("TERM_PROGRAM")
	term := os.Getenv("TERM")
	terminalEmulator := os.Getenv("TERMINAL_EMULATOR")

	fmt.Printf("   TERM_PROGRAM: %q\n", termProgram)
	fmt.Printf("   TERM: %q\n", term)
	fmt.Printf("   TERMINAL_EMULATOR: %q\n", terminalEmulator)

	// Check for hyperlink support indicators
	fmt.Printf("\n   Likely hyperlink support: %v\n", detectHyperlinkSupport())
	fmt.Printf("   Color support: %v\n", detectColorSupport())
	fmt.Printf("   Emoji support: %v\n", detectEmojiSupport())
}

// detectHyperlinkSupport attempts to detect if terminal supports OSC 8 hyperlinks
func detectHyperlinkSupport() bool {
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	term := strings.ToLower(os.Getenv("TERM"))

	// Known terminals with hyperlink support
	switch {
	case strings.Contains(termProgram, "iterm"):
		return true
	case strings.Contains(termProgram, "vscode"):
		return true
	case strings.Contains(termProgram, "hyper"):
		return true
	case strings.Contains(term, "xterm-256"):
		return true // Many modern terminals identify as this
	case os.Getenv("WT_SESSION") != "": // Windows Terminal
		return true
	default:
		return false
	}
}

// detectColorSupport checks if terminal supports colors
func detectColorSupport() bool {
	term := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")

	return strings.Contains(term, "color") ||
		strings.Contains(term, "256") ||
		colorTerm != ""
}

// detectEmojiSupport checks if terminal likely supports emoji
func detectEmojiSupport() bool {
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))

	// Most modern terminals support emoji
	return termProgram == "iterm.app" ||
		termProgram == "vscode" ||
		os.Getenv("WT_SESSION") != ""
}

// testHyperlinkSupport tests different hyperlink rendering approaches
func testHyperlinkSupport(markdown string) {
	fmt.Println("   Testing with Smart Hyperlinks (auto-fallback):")
	renderer1, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(60),
		glamour.WithSmartHyperlinks(),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	output1, err := renderer1.Render(markdown)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Print(output1)
	fmt.Println("   " + strings.Repeat("-", 50))

	fmt.Println("\n   Testing with Force Hyperlinks (OSC 8 only):")
	renderer2, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(60),
		glamour.WithHyperlinks(),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	output2, err := renderer2.Render(markdown)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Print(output2)
	fmt.Println("   " + strings.Repeat("-", 50))
}

// demonstrateAdaptiveFormatter shows a formatter that adapts based on detected capabilities
func demonstrateAdaptiveFormatter(markdown string) {
	// Create an adaptive formatter that changes behavior based on terminal capabilities
	adaptiveFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		// Detect capabilities
		supportsHyperlinks := detectHyperlinkSupport()
		supportsEmoji := detectEmojiSupport()

		// Adapt formatting based on capabilities
		switch {
		case supportsHyperlinks:
			// Use OSC 8 hyperlinks for supported terminals
			return formatOSC8Hyperlink(data.Text, data.URL), nil

		case supportsEmoji:
			// Use emoji indicators for terminals with emoji support
			return fmt.Sprintf("%s 🔗 %s", data.Text, data.URL), nil

		default:
			// Fallback to simple format for basic terminals
			return fmt.Sprintf("%s [%s]", data.Text, data.URL), nil
		}
	})

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(60),
		glamour.WithLinkFormatter(adaptiveFormatter),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	output, err := renderer.Render(markdown)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("   Adaptive formatting (detected: hyperlinks=%v, emoji=%v):\n",
		detectHyperlinkSupport(), detectEmojiSupport())
	fmt.Print(output)
	fmt.Println("   " + strings.Repeat("-", 50))
}

// formatOSC8Hyperlink creates an OSC 8 hyperlink sequence
func formatOSC8Hyperlink(text, url string) string {
	// OSC 8 format: \033]8;;URL\033\\TEXT\033]8;;\033\\
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// Additional utility functions for comprehensive terminal detection

// DetectTerminalFeatures returns a comprehensive feature set
type TerminalFeatures struct {
	SupportsHyperlinks bool
	SupportsColors     bool
	Supports256Colors  bool
	SupportsTrueColor  bool
	SupportsEmoji      bool
	SupportsUnicode    bool
	Name               string
	Version            string
}

// GetTerminalFeatures returns detected terminal features
func GetTerminalFeatures() TerminalFeatures {
	features := TerminalFeatures{
		SupportsHyperlinks: detectHyperlinkSupport(),
		SupportsColors:     detectColorSupport(),
		Supports256Colors:  detect256ColorSupport(),
		SupportsTrueColor:  detectTrueColorSupport(),
		SupportsEmoji:      detectEmojiSupport(),
		SupportsUnicode:    detectUnicodeSupport(),
		Name:               detectTerminalName(),
		Version:            detectTerminalVersion(),
	}

	return features
}

func detect256ColorSupport() bool {
	term := os.Getenv("TERM")
	return strings.Contains(term, "256color")
}

func detectTrueColorSupport() bool {
	colorTerm := strings.ToLower(os.Getenv("COLORTERM"))
	return colorTerm == "truecolor" || colorTerm == "24bit"
}

func detectUnicodeSupport() bool {
	// Most modern terminals support unicode
	lang := os.Getenv("LANG")
	return strings.Contains(strings.ToUpper(lang), "UTF")
}

func detectTerminalName() string {
	if name := os.Getenv("TERM_PROGRAM"); name != "" {
		return name
	}
	if name := os.Getenv("TERMINAL_EMULATOR"); name != "" {
		return name
	}
	return os.Getenv("TERM")
}

func detectTerminalVersion() string {
	return os.Getenv("TERM_PROGRAM_VERSION")
}
