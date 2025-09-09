# Custom Link Formatting - Documentation Updates Plan

## Overview

This document outlines all documentation updates needed to support the custom link formatting feature. Updates maintain consistency with existing documentation style while clearly presenting new capabilities.

## 1. README.md Updates

### New Section: Custom Link Formatting

Add this section after the existing "Custom Renderer" section:

```markdown
### Custom Link Formatting

Glamour supports custom link formatting to control how links are rendered in your terminal output.

#### Built-in Link Formatters

```go
import "github.com/charmbracelet/glamour"

// Text-only links (clickable in smart terminals like iTerm2, VS Code)
r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithTextOnlyLinks(),
)

// URL-only links
r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithURLOnlyLinks(),
)

// Modern terminal hyperlinks (OSC 8 sequences)
r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithHyperlinks(),
)

// Smart hyperlinks with fallback for older terminals
r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithSmartHyperlinks(),
)
```

#### Custom Link Formatters

Create your own link formatting logic:

```go
// Custom formatter that shows links as "text -> url"
customFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    return fmt.Sprintf("%s -> %s", data.Text, data.URL), nil
})

r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithLinkFormatter(customFormatter),
)
```

#### Context-Aware Formatting

Access link context for intelligent formatting:

```go
smartFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    switch {
    case data.IsInTable:
        return data.Text, nil  // Tables: text only
    case data.IsAutoLink:
        return fmt.Sprintf("<%s>", data.URL), nil  // Autolinks: angle brackets
    case glamour.SupportsHyperlinks(ctx):
        return glamour.FormatHyperlink(data.Text, data.URL), nil  // Modern: hyperlinks
    default:
        return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil  // Fallback
    }
})
```

#### Terminal Hyperlink Support

Glamour automatically detects modern terminals that support clickable hyperlinks:

- iTerm2
- VS Code integrated terminal
- Windows Terminal
- WezTerm
- Other terminals with OSC 8 support

Links are automatically clickable in these terminals when using `WithHyperlinks()` or `WithSmartHyperlinks()`.
```

### Update Installation Section

Add note about Go version compatibility:

```markdown
## Installation

```bash
go get github.com/charmbracelet/glamour
```

> **Note**: Custom link formatting requires Go 1.16+ for optimal performance.
```

### Update Examples Section

Add reference to new examples:

```markdown
## Examples

You can find more examples in the [examples](examples/) directory, including:

- [Custom Link Formatting](examples/custom_link_formatting/) - Various link formatting options
- [Terminal Detection](examples/terminal_detection/) - Detecting hyperlink support
- [Context-Aware Rendering](examples/context_aware/) - Formatting based on context
```

## 2. Code Comments and Documentation

### LinkFormatter Interface Comments

```go
// LinkFormatter defines how links should be rendered in terminal output.
// Implementations receive complete link information and rendering context,
// allowing for intelligent formatting decisions based on terminal capabilities,
// link context (table, autolink), and user preferences.
//
// Example custom formatter:
//
//	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
//		if data.IsInTable {
//			return data.Text, nil  // Tables: text only
//		}
//		return fmt.Sprintf("%s [%s]", data.Text, data.URL), nil
//	})
type LinkFormatter interface {
	// FormatLink renders a link according to the formatter's logic.
	// Returns the formatted string and any error encountered.
	//
	// The returned string may contain ANSI escape sequences for styling.
	// Formatters should handle edge cases gracefully (empty text, invalid URLs).
	FormatLink(data LinkData, ctx RenderContext) (string, error)
}

// LinkFormatterFunc is an adapter that allows ordinary functions to implement
// the LinkFormatter interface.
//
// Example usage:
//
//	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
//		return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
//	})
type LinkFormatterFunc func(LinkData, RenderContext) (string, error)

// FormatLink implements the LinkFormatter interface.
func (f LinkFormatterFunc) FormatLink(data LinkData, ctx RenderContext) (string, error) {
	return f(data, ctx)
}
```

### LinkData Struct Comments

```go
// LinkData contains all information about a parsed link, providing formatters
// with complete context for intelligent rendering decisions.
//
// Fields are populated by the renderer based on markdown parsing and context
// analysis. Formatters can use any combination of these fields.
type LinkData struct {
	// URL is the destination URL of the link.
	// For relative links, this may be resolved against BaseURL.
	URL string

	// Text is the display text of the link, extracted from child elements.
	// For autolinks, this typically matches the URL.
	// May be empty if the link has no text content.
	Text string

	// Title is the optional title attribute from markdown syntax.
	// Example: [text](url "title") -> Title = "title"
	Title string

	// BaseURL is the base URL for resolving relative links.
	// Set via WithBaseURL() option.
	BaseURL string

	// IsAutoLink indicates whether this link was automatically detected
	// (e.g., https://example.com) rather than explicit markdown syntax
	// (e.g., [text](https://example.com)).
	IsAutoLink bool

	// IsInTable indicates whether this link appears within a table cell.
	// Useful for space-conscious formatting in tabular contexts.
	IsInTable bool

	// Children contains the original child elements of the link.
	// Advanced formatters can use these for custom text rendering.
	Children []ElementRenderer

	// LinkStyle contains styling information for the URL portion.
	// Use ApplyStyle() to apply these styles to formatted output.
	LinkStyle StylePrimitive

	// TextStyle contains styling information for the text portion.
	// Use ApplyStyle() to apply these styles to formatted output.
	TextStyle StylePrimitive
}
```

### Built-in Formatters Comments

```go
// Built-in link formatters for common use cases.
var (
	// DefaultFormatter replicates the original Glamour link rendering behavior.
	// Outputs: "text url" (text followed by space and URL).
	// Used automatically when no custom formatter is specified.
	DefaultFormatter = LinkFormatterFunc(defaultLinkFormat)

	// TextOnlyFormatter shows only the link text.
	// In terminals supporting hyperlinks, the text becomes clickable.
	// In other terminals, only styled text is shown (not clickable).
	// Useful for clean output where URLs would be distracting.
	TextOnlyFormatter = LinkFormatterFunc(textOnlyLinkFormat)

	// URLOnlyFormatter shows only the URL.
	// Useful for reference lists or when text is redundant.
	URLOnlyFormatter = LinkFormatterFunc(urlOnlyLinkFormat)

	// HyperlinkFormatter creates OSC 8 hyperlinks for compatible terminals.
	// Text appears normal but becomes clickable. URLs are hidden.
	// Does not fall back gracefully - use SmartHyperlinkFormatter for fallback.
	HyperlinkFormatter = LinkFormatterFunc(hyperlinkFormat)

	// SmartHyperlinkFormatter uses OSC 8 hyperlinks when supported,
	// falls back to DefaultFormatter behavior in older terminals.
	// Recommended for most use cases requiring hyperlinks.
	SmartHyperlinkFormatter = LinkFormatterFunc(smartHyperlinkFormat)
)
```

### TermRendererOption Functions Comments

```go
// WithLinkFormatter sets a custom formatter for rendering links.
// The formatter receives complete link context and can make intelligent
// decisions about rendering based on terminal capabilities and link properties.
//
// Example:
//
//	formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
//		return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
//	})
//	r, _ := NewTermRenderer(WithLinkFormatter(formatter))
//
// Pass nil to restore default behavior.
func WithLinkFormatter(formatter LinkFormatter) TermRendererOption

// WithTextOnlyLinks configures the renderer to show only link text.
// In compatible terminals (iTerm2, VS Code, Windows Terminal, WezTerm),
// the text becomes clickable. In other terminals, only styled text is shown.
//
// This is equivalent to WithLinkFormatter(TextOnlyFormatter).
func WithTextOnlyLinks() TermRendererOption

// WithURLOnlyLinks configures the renderer to show only URLs.
// Link text is discarded, which may be useful for reference lists
// or debugging purposes.
//
// This is equivalent to WithLinkFormatter(URLOnlyFormatter).
func WithURLOnlyLinks() TermRendererOption

// WithHyperlinks enables OSC 8 hyperlinks for compatible terminals.
// Text appears normal but becomes clickable, with URLs hidden.
// Does not provide fallback for unsupported terminals.
//
// For automatic fallback, use WithSmartHyperlinks() instead.
// This is equivalent to WithLinkFormatter(HyperlinkFormatter).
func WithHyperlinks() TermRendererOption

// WithSmartHyperlinks enables OSC 8 hyperlinks with automatic fallback.
// Uses modern hyperlinks in compatible terminals, falls back to
// "text url" format in older terminals.
//
// This is the recommended option for most applications.
// This is equivalent to WithLinkFormatter(SmartHyperlinkFormatter).
func WithSmartHyperlinks() TermRendererOption
```

### Hyperlink Utility Functions Comments

```go
// FormatHyperlink creates an OSC 8 hyperlink sequence.
// The text appears normal in the terminal but becomes clickable,
// linking to the specified URL.
//
// OSC 8 format: \e]8;;URL\e\TEXT\e]8;;\e\
//
// Compatible terminals:
//   - iTerm2
//   - VS Code integrated terminal
//   - Windows Terminal
//   - WezTerm
//   - Many modern terminal emulators
//
// In unsupported terminals, the escape sequences may be visible
// or ignored, displaying just the text.
func FormatHyperlink(text, url string) string

// SupportsHyperlinks detects whether the current terminal supports OSC 8
// hyperlinks by examining environment variables and terminal identification.
//
// Detection is based on:
//   - TERM_PROGRAM environment variable
//   - TERM environment variable patterns
//   - Known terminal capabilities
//
// This function is heuristic-based and may not catch all terminal types.
// When in doubt, use SmartHyperlinkFormatter for automatic fallback.
func SupportsHyperlinks(ctx RenderContext) bool

// ApplyStyle applies StylePrimitive formatting to text within the given context.
// Used by formatters to maintain consistent styling with the rest of Glamour.
//
// Example:
//
//	styledText := ApplyStyle(data.Text, data.TextStyle, ctx)
//	styledURL := ApplyStyle(data.URL, data.LinkStyle, ctx)
func ApplyStyle(text string, style StylePrimitive, ctx RenderContext) string
```

## 3. New Examples

### examples/custom_link_formatting/

#### main.go
```go
// Package main demonstrates various custom link formatting options with Glamour.
//
// This example shows:
// - Built-in formatter options
// - Custom formatter implementation
// - Context-aware formatting
// - Terminal capability detection
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
)

const demoMarkdown = `# Link Formatting Demo

Here are various types of links:

1. Regular link: [Glamour](https://github.com/charmbracelet/glamour)
2. Link with title: [Google](https://google.com "Google Search")
3. Autolink: https://github.com
4. Relative link: [README](./README.md)

## Table with Links

| Site | Description |
|------|-------------|
| [GitHub](https://github.com) | Code hosting |
| [Go](https://golang.org) | Programming language |
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "default":
			demonstrateDefault()
		case "text-only":
			demonstrateTextOnly()
		case "url-only":
			demonstrateURLOnly()
		case "hyperlinks":
			demonstrateHyperlinks()
		case "smart":
			demonstrateSmartHyperlinks()
		case "custom":
			demonstrateCustomFormatter()
		case "context":
			demonstrateContextAware()
		default:
			printUsage()
		}
	} else {
		demonstrateAll()
	}
}

func demonstrateAll() {
	fmt.Println("🎨 Glamour Link Formatting Demo\n")
	
	demos := []struct {
		name string
		fn   func()
	}{
		{"Default", demonstrateDefault},
		{"Text-Only", demonstrateTextOnly},
		{"URL-Only", demonstrateURLOnly},
		{"Hyperlinks", demonstrateHyperlinks},
		{"Smart Hyperlinks", demonstrateSmartHyperlinks},
		{"Custom Formatter", demonstrateCustomFormatter},
		{"Context-Aware", demonstrateContextAware},
	}
	
	for _, demo := range demos {
		fmt.Printf("📋 %s:\n", demo.name)
		demo.fn()
		fmt.Println(strings.Repeat("─", 60))
	}
}

// ... (implementation functions as shown in examples document)
```

#### README.md
```markdown
# Custom Link Formatting Example

This example demonstrates the various link formatting options available in Glamour.

## Running the Example

```bash
# Show all formatting options
go run main.go

# Show specific formatter
go run main.go default
go run main.go text-only
go run main.go url-only  
go run main.go hyperlinks
go run main.go smart
go run main.go custom
go run main.go context
```

## Formatters Demonstrated

- **Default**: Current Glamour behavior (`text url`)
- **Text-Only**: Shows only clickable text in smart terminals
- **URL-Only**: Shows only URLs
- **Hyperlinks**: OSC 8 hyperlinks for modern terminals
- **Smart**: Hyperlinks with fallback for older terminals
- **Custom**: User-defined formatting logic
- **Context-Aware**: Different formatting based on link context

## Terminal Compatibility

The hyperlink examples work best in modern terminals:
- iTerm2
- VS Code integrated terminal
- Windows Terminal  
- WezTerm

In older terminals, smart formatters automatically fall back to readable alternatives.
```

### examples/terminal_detection/

#### main.go
```go
// Package main demonstrates terminal capability detection for hyperlinks.
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
)

func main() {
	fmt.Println("🔍 Terminal Hyperlink Support Detection\n")
	
	// Get environment info
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	
	fmt.Printf("TERM: %q\n", term)
	fmt.Printf("TERM_PROGRAM: %q\n", termProgram)
	
	// Test detection
	ctx := glamour.RenderContext{} // Mock context for detection
	supports := glamour.SupportsHyperlinks(ctx)
	
	fmt.Printf("Hyperlinks supported: %t\n\n", supports)
	
	if supports {
		fmt.Println("✅ Your terminal supports hyperlinks!")
		demonstrateHyperlinks()
	} else {
		fmt.Println("❌ Your terminal may not support hyperlinks.")
		demonstrateSmartFallback()
	}
}

// ... (implementation functions)
```

### examples/context_aware/

#### main.go
```go
// Package main demonstrates context-aware link formatting.
package main

import (
	"fmt"
	"github.com/charmbracelet/glamour"
)

const contextMarkdown = `# Context-Aware Demo

Regular paragraph link: [GitHub](https://github.com)

Autolink in paragraph: https://golang.org

## Table with Links

| Platform | Link |
|----------|------|
| [GitHub](https://github.com) | Code hosting |
| [Stack Overflow](https://stackoverflow.com) | Q&A site |

Another paragraph link: [Google](https://google.com)
`

func main() {
	fmt.Println("🧠 Context-Aware Link Formatting\n")
	
	contextFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
		switch {
		case data.IsInTable:
			// Tables: keep it concise
			return fmt.Sprintf("🔗 %s", data.Text), nil
		case data.IsAutoLink:
			// Autolinks: show in brackets
			return fmt.Sprintf("<%s>", data.URL), nil
		case glamour.SupportsHyperlinks(ctx):
			// Modern terminals: hyperlinks
			return glamour.FormatHyperlink(data.Text, data.URL), nil
		default:
			// Fallback: parentheses format
			return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
		}
	})
	
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithLinkFormatter(contextFormatter),
	)
	if err != nil {
		panic(err)
	}
	
	out, err := r.Render(contextMarkdown)
	if err != nil {
		panic(err)
	}
	
	fmt.Print(out)
}
```

## 4. API Documentation Updates

### GoDoc Package Documentation

Update the package-level documentation in `glamour.go`:

```go
// Package glamour lets you render markdown documents & templates on ANSI
// compatible terminals. You can create your own stylesheet or simply use one of
// the stylish defaults.
//
// # Basic Usage
//
//	import "github.com/charmbracelet/glamour"
//
//	in := `# Hello World
//
//	This is a simple example of Markdown rendering with Glamour!
//	Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples).
//
//	Bye!`
//
//	out, err := glamour.Render(in, "dark")
//	fmt.Print(out)
//
// # Custom Link Formatting
//
// Glamour supports custom link formatting to control how links appear in terminal output:
//
//	// Text-only links (clickable in smart terminals)
//	r, _ := glamour.NewTermRenderer(
//		glamour.WithStandardStyle("dark"),
//		glamour.WithTextOnlyLinks(),
//	)
//
//	// Custom formatting logic
//	formatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
//		return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
//	})
//	r, _ := glamour.NewTermRenderer(
//		glamour.WithStandardStyle("dark"),
//		glamour.WithLinkFormatter(formatter),
//	)
//
// # Terminal Hyperlinks
//
// Modern terminals supporting OSC 8 sequences can display clickable hyperlinks:
//
//	r, _ := glamour.NewTermRenderer(
//		glamour.WithStandardStyle("dark"),
//		glamour.WithSmartHyperlinks(), // Automatic fallback for older terminals
//	)
//
// Supported terminals include iTerm2, VS Code, Windows Terminal, and WezTerm.
package glamour
```

## 5. Migration Guide

### MIGRATION.md (new file)

```markdown
# Migration Guide: Custom Link Formatting

This guide helps you adopt Glamour's new custom link formatting features.

## Backward Compatibility

✅ **No breaking changes** - all existing code continues to work unchanged.

## Quick Start

### Enable Text-Only Links
```go
// Before
r, _ := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"))

// After - clickable links in smart terminals
r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithTextOnlyLinks(),
)
```

### Enable Hyperlinks
```go
// Smart hyperlinks with fallback
r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithSmartHyperlinks(),
)
```

### Custom Formatting
```go
// Custom formatter
formatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
})

r, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithLinkFormatter(formatter),
)
```

## Terminal Detection

New terminals are automatically detected for hyperlink support. To test:

```bash
# Check your terminal
go run examples/terminal_detection/main.go
```

## Common Patterns

### Conditional Formatting
```go
formatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    if data.IsInTable {
        return data.Text, nil // Concise for tables
    }
    return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
})
```

### Domain-Specific Icons
```go
formatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    if strings.Contains(data.URL, "github.com") {
        return fmt.Sprintf("🐙 %s", data.Text), nil
    }
    return fmt.Sprintf("🔗 %s", data.Text), nil
})
```

## Performance Notes

- Default behavior has zero overhead
- Custom formatters are called only when configured
- Terminal detection is cached per context

## Troubleshooting

### Links Not Clickable
1. Verify terminal support: `go run examples/terminal_detection/`
2. Use `WithSmartHyperlinks()` for automatic fallback
3. Check TERM environment variables

### Custom Formatter Errors
1. Handle edge cases (empty text, invalid URLs)
2. Return safe fallbacks on errors
3. Test with various link types (autolinks, tables, titles)
```

## 6. CHANGELOG Entry

```markdown
## [v2.0.0] - 2024-XX-XX

### Added
- **Custom Link Formatting**: New `LinkFormatter` interface allows complete control over link rendering
- **Built-in Formatters**: 
  - `WithTextOnlyLinks()` - Show only clickable text
  - `WithURLOnlyLinks()` - Show only URLs  
  - `WithHyperlinks()` - OSC 8 hyperlinks for modern terminals
  - `WithSmartHyperlinks()` - Hyperlinks with automatic fallback
- **Terminal Hyperlink Support**: Automatic detection and support for OSC 8 sequences in modern terminals (iTerm2, VS Code, Windows Terminal, WezTerm)
- **Context-Aware Rendering**: LinkData includes context (table, autolink) for intelligent formatting decisions
- **Link Metadata**: Access to link titles, text, URLs, and styling information

### Enhanced
- Link rendering architecture now supports extensible formatting
- Terminal capability detection for optimal output
- Comprehensive examples and documentation

### Backward Compatibility
- **No breaking changes** - All existing code continues to work unchanged
- Default behavior remains identical to previous versions
- New functionality is opt-in only

### Examples
- Added `examples/custom_link_formatting/` - Comprehensive formatting demonstrations
- Added `examples/terminal_detection/` - Terminal capability detection
- Added `examples/context_aware/` - Context-based formatting examples
```

## 7. Documentation Structure

### File Organization
```
docs/
├── README.md (updated)
├── MIGRATION.md (new)
├── API.md (updated)
└── EXAMPLES.md (new)

examples/
├── custom_link_formatting/ (new)
├── terminal_detection/ (new)
└── context_aware/ (new)

.github/
└── PULL_REQUEST_TEMPLATE.md (updated with link formatting checklist)
```

### Documentation Quality Checklist

- [ ] All new interfaces documented with examples
- [ ] Code comments follow Go conventions
- [ ] README examples work with copy-paste
- [ ] Migration guide covers common use cases
- [ ] Examples demonstrate real-world scenarios
- [ ] API documentation includes parameter descriptions
- [ ] Error handling patterns documented
- [ ] Performance considerations noted
- [ ] Terminal compatibility clearly explained
- [ ] Troubleshooting section included

This comprehensive documentation plan ensures users can easily discover, understand, and adopt the new custom link formatting capabilities while maintaining the high-quality documentation standards of the Glamour project.