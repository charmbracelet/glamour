package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

func main() {
	fmt.Println("=== CONTEXT-AWARE FORMATTING DEMO ===\n")

	// Complex markdown with various contexts
	markdown := `# Context-Aware Link Formatting

## Regular Paragraphs
Here are some links in regular paragraphs:
- Visit [GitHub](https://github.com) for code repositories
- Search on [Google](https://google.com "Google Search Engine") 
- Check out [Stack Overflow](https://stackoverflow.com) for Q&A

## Autolinks
Direct URLs are treated as autolinks:
- <https://example.com>
- <https://docs.github.com>
- Visit https://golang.org for Go documentation

## Table Context
Links behave differently in tables to save space:

| Service | Link | Description |
|---------|------|-------------|
| GitHub | [Repository](https://github.com/charmbracelet/glamour) | Code hosting |
| Google | [Search Engine](https://google.com) | Web search |
| Stack Overflow | [Q&A Platform](https://stackoverflow.com) | Developer help |

## Mixed Content
Paragraph with both [regular links](https://example.com) and autolinks like https://github.com.

> **Quote Context**: Links in quotes like [this one](https://example.com) can be formatted differently.

## Code Context
In code blocks, links are usually plain text:
` + "```" + `
Visit https://example.com for more info
[Not a link](https://example.com)
` + "```" + `

But inline code can contain links: ` + "`" + `see https://example.com` + "`" + `
`

	fmt.Println("1. Standard Context-Aware Formatter")
	fmt.Println("   Different formatting based on where links appear")
	standardContextFormatter := createStandardContextFormatter()
	renderWithFormatter(markdown, standardContextFormatter, "StandardContext")

	fmt.Println("2. Advanced Context-Aware Formatter")
	fmt.Println("   More sophisticated context detection and formatting")
	advancedContextFormatter := createAdvancedContextFormatter()
	renderWithFormatter(markdown, advancedContextFormatter, "AdvancedContext")

	fmt.Println("3. Table-Optimized Formatter")
	fmt.Println("   Specifically optimized for table content")
	tableFormatter := createTableOptimizedFormatter()
	renderWithFormatter(markdown, tableFormatter, "TableOptimized")

	fmt.Println("4. Progressive Enhancement Formatter")
	fmt.Println("   Adapts based on both context and terminal capabilities")
	progressiveFormatter := createProgressiveFormatter()
	renderWithFormatter(markdown, progressiveFormatter, "Progressive")

	fmt.Println("5. Debug Context Formatter")
	fmt.Println("   Shows context information for development/debugging")
	debugFormatter := createDebugContextFormatter()
	renderWithFormatter(markdown, debugFormatter, "Debug")

	fmt.Println("✅ Context-aware formatting demo completed!")
	fmt.Println("\n💡 Key Context Types:")
	fmt.Println("   • Regular paragraphs - Full formatting")
	fmt.Println("   • Tables - Compact formatting")
	fmt.Println("   • Autolinks - Special handling")
	fmt.Println("   • Quotes/blocks - Alternative styling")
}

// Standard context-aware formatter with basic context switching
func createStandardContextFormatter() ansi.LinkFormatter {
	return ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		switch {
		case data.IsInTable:
			// Tables: compact format to save space
			return data.Text, nil

		case data.IsAutoLink:
			// Autolinks: show with angle brackets
			return fmt.Sprintf("<%s>", data.URL), nil

		default:
			// Regular links: show both text and URL
			return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
		}
	})
}

// Advanced context formatter with more sophisticated logic
func createAdvancedContextFormatter() ansi.LinkFormatter {
	return ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		// Determine context from various signals
		isShortText := len(data.Text) <= 10
		isLongURL := len(data.URL) > 50
		hasTitle := data.Title != ""

		switch {
		case data.IsInTable:
			if isShortText {
				// Short text in table: show as clickable if possible
				return data.Text, nil
			} else {
				// Long text in table: truncate
				if len(data.Text) > 15 {
					return data.Text[:12] + "...", nil
				}
				return data.Text, nil
			}

		case data.IsAutoLink:
			if isLongURL {
				// Long autolinks: show domain only
				return formatDomainOnly(data.URL), nil
			}
			return fmt.Sprintf("<%s>", data.URL), nil

		case hasTitle:
			// Links with titles: include title in output
			return fmt.Sprintf("%s → %s (%s)", data.Text, data.URL, data.Title), nil

		case isLongURL:
			// Long URLs: show text with domain hint
			return fmt.Sprintf("%s [%s]", data.Text, formatDomainOnly(data.URL)), nil

		default:
			// Regular formatting
			return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
		}
	})
}

// Table-optimized formatter that prioritizes space efficiency
func createTableOptimizedFormatter() ansi.LinkFormatter {
	return ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		if data.IsInTable {
			// In tables: very compact
			return data.Text, nil
		}

		// Outside tables: normal formatting
		if data.IsAutoLink {
			return data.URL, nil
		}

		return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
	})
}

// Progressive formatter that combines context awareness with terminal detection
func createProgressiveFormatter() ansi.LinkFormatter {
	return ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		// Simulate terminal capability detection
		supportsHyperlinks := detectHyperlinkCapability()
		supportsEmoji := detectEmojiCapability()

		switch {
		case data.IsInTable:
			if supportsHyperlinks {
				// Table + hyperlinks: clickable text only
				return formatHyperlink(data.Text, data.URL), nil
			}
			// Table + no hyperlinks: text only
			return data.Text, nil

		case data.IsAutoLink:
			if supportsEmoji {
				// Autolink with emoji indicator
				return fmt.Sprintf("🔗 %s", data.URL), nil
			}
			return fmt.Sprintf("<%s>", data.URL), nil

		default:
			if supportsHyperlinks {
				// Regular link with hyperlinks
				return formatHyperlink(data.Text, data.URL), nil
			} else if supportsEmoji {
				// Regular link with emoji
				return fmt.Sprintf("%s 🔗 %s", data.Text, data.URL), nil
			}
			// Fallback
			return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
		}
	})
}

// Debug formatter that shows context information (useful for development)
func createDebugContextFormatter() ansi.LinkFormatter {
	return ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		// Build context info
		var contextFlags []string
		if data.IsInTable {
			contextFlags = append(contextFlags, "TABLE")
		}
		if data.IsAutoLink {
			contextFlags = append(contextFlags, "AUTO")
		}
		if data.Title != "" {
			contextFlags = append(contextFlags, "TITLED")
		}
		if len(contextFlags) == 0 {
			contextFlags = append(contextFlags, "REGULAR")
		}

		contextInfo := strings.Join(contextFlags, "|")

		// Format with debug info
		return fmt.Sprintf("%s → %s [%s]", data.Text, data.URL, contextInfo), nil
	})
}

// Helper function to render with a specific formatter
func renderWithFormatter(markdown string, formatter ansi.LinkFormatter, name string) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(70),
		glamour.WithLinkFormatter(formatter),
	)
	if err != nil {
		log.Printf("Error creating renderer for %s: %v", name, err)
		return
	}

	output, err := renderer.Render(markdown)
	if err != nil {
		log.Printf("Error rendering with %s formatter: %v", name, err)
		return
	}

	fmt.Print(output)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()
}

// Utility functions

// formatDomainOnly extracts and formats just the domain from a URL
func formatDomainOnly(url string) string {
	// Simple domain extraction (for demo purposes)
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}

	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	return url
}

// formatHyperlink creates an OSC 8 hyperlink (simplified)
func formatHyperlink(text, url string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// detectHyperlinkCapability simulates terminal hyperlink detection
func detectHyperlinkCapability() bool {
	// Simplified detection for demo
	return strings.Contains(strings.ToLower(getEnvOr("TERM_PROGRAM", "")), "iterm") ||
		strings.Contains(strings.ToLower(getEnvOr("TERM_PROGRAM", "")), "vscode") ||
		getEnvOr("WT_SESSION", "") != ""
}

// detectEmojiCapability simulates emoji support detection
func detectEmojiCapability() bool {
	// Most modern terminals support emoji
	return true
}

// getEnvOr returns environment variable value or default
func getEnvOr(key, defaultValue string) string {
	if value := strings.TrimSpace(key); value != "" { // Simplified for demo
		return defaultValue
	}
	return defaultValue
}

// Example of context-specific styling
func createContextAwareStyledFormatter() ansi.LinkFormatter {
	return ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		// Apply different styles based on context
		switch {
		case data.IsInTable:
			// Minimal styling for tables
			return applyTableLinkStyle(data.Text), nil

		case data.IsAutoLink:
			// Special styling for autolinks
			return applyAutoLinkStyle(data.URL), nil

		default:
			// Full styling for regular links
			styledText := applyRegularLinkTextStyle(data.Text)
			styledURL := applyRegularLinkURLStyle(data.URL)
			return fmt.Sprintf("%s %s", styledText, styledURL), nil
		}
	})
}

// Placeholder styling functions (would use actual styling in real implementation)
func applyTableLinkStyle(text string) string {
	return fmt.Sprintf("[TABLE:%s]", text)
}

func applyAutoLinkStyle(url string) string {
	return fmt.Sprintf("[AUTO:%s]", url)
}

func applyRegularLinkTextStyle(text string) string {
	return fmt.Sprintf("[TEXT:%s]", text)
}

func applyRegularLinkURLStyle(url string) string {
	return fmt.Sprintf("[URL:%s]", url)
}
