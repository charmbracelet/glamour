package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

func main() {
	// Sample markdown with various link types
	markdown := `# Custom Link Formatting Demo

## Basic Links
Here's a [regular link](https://example.com) and another [Google search](https://google.com "Google Search Engine").

## Autolinks
Visit <https://github.com> for repositories.

## Different Contexts
The following table shows some links:

| Site | URL |
|------|-----|
| [GitHub](https://github.com) | Repository hosting |
| [Stack Overflow](https://stackoverflow.com) | Q&A platform |

## Long URLs
Here's a [very long URL](https://example.com/very/long/path/to/some/resource?with=many&query=parameters&and=more&stuff=here) for testing.
`

	fmt.Println("=== BUILT-IN FORMATTERS ===\n")

	// 1. Default behavior (unchanged)
	fmt.Println("1. Default Formatter")
	fmt.Println("   Shows both text and URL with styling")
	renderWithFormatter(markdown, nil, "Default")

	// 2. Text-only links (clickable in smart terminals)
	fmt.Println("2. Text-Only Links")
	fmt.Println("   Shows only clickable text in smart terminals")
	renderer2, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(80),
		glamour.WithTextOnlyLinks(),
	)
	if err != nil {
		log.Fatal(err)
	}
	renderWithRenderer(markdown, renderer2, "TextOnly")

	// 3. URL-only links
	fmt.Println("3. URL-Only Links")
	fmt.Println("   Shows only URLs, hiding link text")
	renderer3, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(80),
		glamour.WithURLOnlyLinks(),
	)
	if err != nil {
		log.Fatal(err)
	}
	renderWithRenderer(markdown, renderer3, "URLOnly")

	// 4. Hyperlinks (OSC 8)
	fmt.Println("4. Hyperlink Formatter")
	fmt.Println("   Uses OSC 8 hyperlinks (clickable text, hidden URLs)")
	renderer4, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(80),
		glamour.WithHyperlinks(),
	)
	if err != nil {
		log.Fatal(err)
	}
	renderWithRenderer(markdown, renderer4, "Hyperlinks")

	// 5. Smart hyperlinks with fallback
	fmt.Println("5. Smart Hyperlinks")
	fmt.Println("   OSC 8 hyperlinks with fallback to default format")
	renderer5, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(80),
		glamour.WithSmartHyperlinks(),
	)
	if err != nil {
		log.Fatal(err)
	}
	renderWithRenderer(markdown, renderer5, "SmartHyperlinks")

	fmt.Println("\n=== CUSTOM FORMATTERS ===\n")

	// 6. Markdown-style formatter
	fmt.Println("6. Markdown-Style Formatter")
	fmt.Println("   Outputs markdown-style links [text](url)")
	markdownFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		if data.Title != "" {
			return fmt.Sprintf("[%s](%s \"%s\")", data.Text, data.URL, data.Title), nil
		}
		return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
	})
	renderWithFormatter(markdown, markdownFormatter, "Markdown")

	// 7. Domain-based formatter with emojis
	fmt.Println("7. Domain-Based Formatter")
	fmt.Println("   Different icons based on domain")
	domainFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		u, err := url.Parse(data.URL)
		if err != nil {
			return fmt.Sprintf("%s [%s]", data.Text, data.URL), nil
		}

		domain := strings.ToLower(u.Hostname())
		switch {
		case strings.Contains(domain, "github.com"):
			return fmt.Sprintf("🐙 %s", data.Text), nil
		case strings.Contains(domain, "google.com"):
			return fmt.Sprintf("🔍 %s", data.Text), nil
		case strings.Contains(domain, "stackoverflow.com"):
			return fmt.Sprintf("📚 %s", data.Text), nil
		default:
			return fmt.Sprintf("🔗 %s (%s)", data.Text, u.Hostname()), nil
		}
	})
	renderWithFormatter(markdown, domainFormatter, "Domain")

	// 8. Length-aware formatter
	fmt.Println("8. Length-Aware Formatter")
	fmt.Println("   Truncates long URLs to keep output clean")
	lengthFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		const maxURLLength = 50

		if len(data.URL) <= maxURLLength {
			// Short URLs: show both text and URL
			return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
		}

		// Long URLs: show only text with domain
		u, err := url.Parse(data.URL)
		if err != nil {
			return data.Text, nil
		}

		return fmt.Sprintf("%s [%s...]", data.Text, u.Hostname()), nil
	})
	renderWithFormatter(markdown, lengthFormatter, "Length")

	// 9. Context-aware formatter
	fmt.Println("9. Context-Aware Formatter")
	fmt.Println("   Different formatting based on link context")
	contextFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		switch {
		case data.IsInTable:
			// Tables: just show text to save space
			return data.Text, nil
		case data.IsAutoLink:
			// Autolinks: show in angle brackets
			return fmt.Sprintf("<%s>", data.URL), nil
		default:
			// Regular links: show both with arrow
			return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
		}
	})
	renderWithFormatter(markdown, contextFormatter, "Context")

	// 10. Error-handling formatter
	fmt.Println("10. Error-Safe Formatter")
	fmt.Println("    Demonstrates defensive programming")
	safeFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		// Always validate inputs
		if data.URL == "" {
			return data.Text, nil // fallback to text
		}

		if data.Text == "" {
			data.Text = data.URL // fallback to URL
		}

		// Handle URL parsing errors gracefully
		u, err := url.Parse(data.URL)
		if err != nil {
			return fmt.Sprintf("%s [invalid URL]", data.Text), nil
		}

		// Safe formatting with validation
		return fmt.Sprintf("%s <%s>", data.Text, u.String()), nil
	})
	renderWithFormatter(markdown, safeFormatter, "Safe")

	fmt.Println("\n=== ADVANCED EXAMPLES ===\n")

	// 11. Plugin-style formatter
	fmt.Println("11. Plugin-Style Formatter")
	fmt.Println("    Extensible formatter system")
	pluginFormatter := createPluginFormatter()
	renderWithFormatter(markdown, pluginFormatter, "Plugin")

	fmt.Println("\n✅ All examples completed successfully!")
	fmt.Println("\nNote: Hyperlink support depends on your terminal.")
	fmt.Println("Try these examples in different terminals to see the differences!")
}

// Helper function to render with a specific formatter
func renderWithFormatter(markdown string, formatter ansi.LinkFormatter, name string) {
	var renderer *glamour.TermRenderer
	var err error

	if formatter == nil {
		// Default formatter
		renderer, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(80),
		)
	} else {
		renderer, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(80),
			glamour.WithLinkFormatter(formatter),
		)
	}

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

// Helper function to render with a pre-configured renderer
func renderWithRenderer(markdown string, renderer *glamour.TermRenderer, name string) {
	output, err := renderer.Render(markdown)
	if err != nil {
		log.Printf("Error rendering with %s renderer: %v", name, err)
		return
	}

	fmt.Print(output)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()
}

// Plugin-style formatter implementation
type FormatterPlugin interface {
	Name() string
	Priority() int
	CanHandle(data ansi.LinkData) bool
	Format(data ansi.LinkData, ctx ansi.RenderContext) (string, error)
}

type GitHubPlugin struct{}

func (p *GitHubPlugin) Name() string  { return "github" }
func (p *GitHubPlugin) Priority() int { return 10 }
func (p *GitHubPlugin) CanHandle(data ansi.LinkData) bool {
	return strings.Contains(strings.ToLower(data.URL), "github.com")
}
func (p *GitHubPlugin) Format(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
	return fmt.Sprintf("🐙 %s [GitHub]", data.Text), nil
}

type GooglePlugin struct{}

func (p *GooglePlugin) Name() string  { return "google" }
func (p *GooglePlugin) Priority() int { return 5 }
func (p *GooglePlugin) CanHandle(data ansi.LinkData) bool {
	return strings.Contains(strings.ToLower(data.URL), "google.com")
}
func (p *GooglePlugin) Format(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
	return fmt.Sprintf("🔍 %s [Google]", data.Text), nil
}

type PluginFormatter struct {
	plugins []FormatterPlugin
}

func (pf *PluginFormatter) FormatLink(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
	// Find the first plugin that can handle this link
	for _, plugin := range pf.plugins {
		if plugin.CanHandle(data) {
			return plugin.Format(data, ctx)
		}
	}

	// Fallback to default format
	return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
}

func createPluginFormatter() *PluginFormatter {
	return &PluginFormatter{
		plugins: []FormatterPlugin{
			&GitHubPlugin{},
			&GooglePlugin{},
		},
	}
}
