# Custom Link Formatting - Code Examples

## Overview

This document provides comprehensive code examples demonstrating the custom link formatting capabilities of Glamour. Examples range from basic usage to advanced custom implementations.

## Basic Usage Examples

### 1. Default Behavior (Unchanged)
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    markdown := `# Links Demo

Here's a [regular link](https://example.com) and an autolink: https://github.com

Visit [Google](https://google.com "Google Search") for searching.`

    // Default behavior - no changes needed
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithWordWrap(80),
    )
    
    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output:
    // Here's a regular link https://example.com and an autolink: https://github.com https://github.com
    // Visit Google https://google.com for searching.
}
```

### 2. Text-Only Links
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    markdown := `# Text-Only Links

Click [here](https://example.com) to visit the site.
Check out https://github.com for code repositories.`

    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithTextOnlyLinks(), // Show only clickable text
    )
    
    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output in smart terminals (iTerm2, VS Code, etc.):
    // Click here to visit the site.  (clickable)
    // Check out https://github.com for code repositories. (clickable)
    //
    // Output in basic terminals:
    // Click here to visit the site.
    // Check out https://github.com for code repositories.
}
```

### 3. URL-Only Links
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    markdown := `# URL-Only Links

Visit [Google](https://google.com) and [GitHub](https://github.com).`

    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithURLOnlyLinks(), // Show only URLs
    )
    
    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output:
    // Visit https://google.com and https://github.com.
}
```

### 4. Modern Terminal Hyperlinks
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    markdown := `# Hyperlink Demo

Visit [Google](https://google.com) and [GitHub](https://github.com).`

    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithHyperlinks(), // Enable OSC 8 hyperlinks
    )
    
    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output in compatible terminals:
    // Visit Google and GitHub. (both clickable with invisible URLs)
}
```

### 5. Smart Hyperlinks with Fallback
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    markdown := `# Smart Links

Visit [Google](https://google.com) for searching.`

    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithSmartHyperlinks(), // OSC 8 with fallback
    )
    
    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output in modern terminals: Google (clickable)
    // Output in older terminals: Google https://google.com
}
```

## Custom Formatter Examples

### 1. Markdown-Style Formatter
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    // Custom formatter that outputs markdown-style links
    markdownFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
        if data.Title != "" {
            return fmt.Sprintf("[%s](%s \"%s\")", data.Text, data.URL, data.Title), nil
        }
        return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
    })
    
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithLinkFormatter(markdownFormatter),
    )
    
    markdown := `Visit [Google](https://google.com "Google Search") for help.`
    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output: Visit [Google](https://google.com "Google Search") for help.
}
```

### 2. Context-Aware Formatter
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    // Smart formatter that adapts based on context
    contextFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
        // Different formatting based on context
        switch {
        case data.IsInTable:
            // Tables: just show text to save space
            return data.Text, nil
            
        case data.IsAutoLink:
            // Autolinks: show in angle brackets
            return fmt.Sprintf("<%s>", data.URL), nil
            
        case glamour.SupportsHyperlinks(ctx):
            // Modern terminals: use hyperlinks
            return glamour.FormatHyperlink(data.Text, data.URL), nil
            
        default:
            // Fallback: show both text and URL
            return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
        }
    })
    
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithLinkFormatter(contextFormatter),
    )
    
    markdown := `# Context Demo

Regular link: [Google](https://google.com)
Autolink: https://github.com

| Site | URL |
|------|-----|  
| [Google](https://google.com) | Search engine |`

    out, _ := r.Render(markdown)
    fmt.Print(out)
}
```

### 3. Styled Custom Formatter
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    // Custom formatter using existing styles
    styledFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
        // Apply styles manually for custom format
        styledText := glamour.ApplyStyle(data.Text, data.TextStyle, ctx)
        styledURL := glamour.ApplyStyle(data.URL, data.LinkStyle, ctx)
        
        // Custom format: text -> url
        return fmt.Sprintf("%s -> %s", styledText, styledURL), nil
    })
    
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithLinkFormatter(styledFormatter),
    )
    
    markdown := `Visit [Google](https://google.com) now!`
    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output: Google -> https://google.com (with styling)
}
```

### 4. Domain-Based Formatter  
```go
package main

import (
    "fmt"
    "net/url"
    "strings"
    "github.com/charmbracelet/glamour"
)

func main() {
    // Formatter that customizes based on domain
    domainFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
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
    
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithLinkFormatter(domainFormatter),
    )
    
    markdown := `# Resource Links

- [GitHub Repo](https://github.com/charmbracelet/glamour)
- [Google Search](https://google.com)
- [Stack Overflow](https://stackoverflow.com/questions/tagged/go)
- [Other Site](https://example.com)`

    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output:
    // • 🐙 GitHub Repo
    // • 🔍 Google Search  
    // • 📚 Stack Overflow
    // • 🔗 Other Site (example.com)
}
```

### 5. Length-Aware Formatter
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    // Formatter that adapts based on URL length
    lengthFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
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
    
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithLinkFormatter(lengthFormatter),
    )
    
    markdown := `# Length Demo

Short: [Google](https://google.com)
Long: [Very Long URL](https://example.com/very/long/path/to/some/resource?with=many&query=parameters&and=more&stuff=here)`

    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output:
    // Short: Google (https://google.com)
    // Long: Very Long URL [example.com...]
}
```

## Unit Test Examples

### 1. Basic Custom Formatter Test
```go
func TestCustomFormatter(t *testing.T) {
    formatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
        return fmt.Sprintf("LINK: %s -> %s", data.Text, data.URL), nil
    })
    
    r, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithLinkFormatter(formatter),
    )
    require.NoError(t, err)
    
    markdown := "[test](https://example.com)"
    result, err := r.Render(markdown)
    require.NoError(t, err)
    
    assert.Contains(t, result, "LINK: test -> https://example.com")
}
```

### 2. Error Handling Test
```go
func TestFormatterErrorHandling(t *testing.T) {
    errorFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
        if data.URL == "error" {
            return "", errors.New("test error")
        }
        return data.Text, nil
    })
    
    r, err := glamour.NewTermRenderer(
        glamour.WithLinkFormatter(errorFormatter),
    )
    require.NoError(t, err)
    
    // This should trigger the error
    markdown := "[test](error)"
    _, err = r.Render(markdown)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "test error")
}
```

### 3. Context Testing
```go
func TestFormatterWithContext(t *testing.T) {
    contextFormatter := glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
        if data.IsAutoLink {
            return fmt.Sprintf("AUTO: %s", data.URL), nil
        }
        return fmt.Sprintf("NORMAL: %s", data.Text), nil
    })
    
    r, err := glamour.NewTermRenderer(
        glamour.WithLinkFormatter(contextFormatter),
    )
    require.NoError(t, err)
    
    tests := []struct {
        markdown string
        contains string
    }{
        {"[text](https://example.com)", "NORMAL: text"},
        {"https://example.com", "AUTO: https://example.com"},
    }
    
    for _, tt := range tests {
        result, err := r.Render(tt.markdown)
        require.NoError(t, err)
        assert.Contains(t, result, tt.contains)
    }
}
```

## Advanced Usage Examples

### 1. Multiple Formatters for Different Content
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

// DocumentRenderer wraps multiple Glamour renderers with different formatters
type DocumentRenderer struct {
    defaultRenderer  *glamour.TermRenderer
    tableRenderer    *glamour.TermRenderer 
    hyperlinkRenderer *glamour.TermRenderer
}

func NewDocumentRenderer() (*DocumentRenderer, error) {
    // Default renderer
    defaultR, err := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"))
    if err != nil {
        return nil, err
    }
    
    // Table-optimized renderer (text-only links)
    tableR, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithTextOnlyLinks(),
    )
    if err != nil {
        return nil, err
    }
    
    // Hyperlink renderer for modern output
    hyperlinkR, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithSmartHyperlinks(),
    )
    if err != nil {
        return nil, err
    }
    
    return &DocumentRenderer{
        defaultRenderer:   defaultR,
        tableRenderer:     tableR,
        hyperlinkRenderer: hyperlinkR,
    }, nil
}

func (dr *DocumentRenderer) RenderForContext(markdown string, context string) (string, error) {
    switch context {
    case "table":
        return dr.tableRenderer.Render(markdown)
    case "modern":
        return dr.hyperlinkRenderer.Render(markdown)
    default:
        return dr.defaultRenderer.Render(markdown)
    }
}
```

### 2. Configuration-Driven Formatter
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

type LinkConfig struct {
    ShowText     bool   `json:"show_text"`
    ShowURL      bool   `json:"show_url"`
    Separator    string `json:"separator"`
    UseHyperlinks bool  `json:"use_hyperlinks"`
    MaxURLLength int    `json:"max_url_length"`
}

func CreateConfigurableFormatter(config LinkConfig) glamour.LinkFormatter {
    return glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
        var parts []string
        
        if config.ShowText && data.Text != "" {
            if config.UseHyperlinks && glamour.SupportsHyperlinks(ctx) {
                parts = append(parts, glamour.FormatHyperlink(data.Text, data.URL))
            } else {
                parts = append(parts, data.Text)
            }
        }
        
        if config.ShowURL {
            url := data.URL
            if config.MaxURLLength > 0 && len(url) > config.MaxURLLength {
                url = url[:config.MaxURLLength-3] + "..."
            }
            parts = append(parts, url)
        }
        
        if len(parts) == 0 {
            return data.Text, nil // fallback
        }
        
        return strings.Join(parts, config.Separator), nil
    })
}

func main() {
    config := LinkConfig{
        ShowText:     true,
        ShowURL:      true,
        Separator:    " | ",
        UseHyperlinks: false,
        MaxURLLength: 30,
    }
    
    formatter := CreateConfigurableFormatter(config)
    
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithLinkFormatter(formatter),
    )
    
    markdown := "[Long Example](https://example.com/very/long/url/path)"
    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output: Long Example | https://example.com/very/lon...
}
```

### 3. Plugin-Style Formatter System
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

// FormatterPlugin interface for extensible formatters
type FormatterPlugin interface {
    Name() string
    Priority() int
    CanHandle(data glamour.LinkData) bool
    Format(data glamour.LinkData, ctx glamour.RenderContext) (string, error)
}

// GitHubPlugin handles GitHub links specially
type GitHubPlugin struct{}

func (p *GitHubPlugin) Name() string { return "github" }
func (p *GitHubPlugin) Priority() int { return 10 }
func (p *GitHubPlugin) CanHandle(data glamour.LinkData) bool {
    return strings.Contains(data.URL, "github.com")
}
func (p *GitHubPlugin) Format(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    return fmt.Sprintf("🐙 %s", data.Text), nil
}

// PluginFormatter manages multiple plugins
type PluginFormatter struct {
    plugins []FormatterPlugin
    fallback glamour.LinkFormatter
}

func (pf *PluginFormatter) FormatLink(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    // Sort plugins by priority and find first match
    for _, plugin := range pf.plugins {
        if plugin.CanHandle(data) {
            return plugin.Format(data, ctx)
        }
    }
    
    // Use fallback formatter
    if pf.fallback != nil {
        return pf.fallback.FormatLink(data, ctx)
    }
    
    return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
}

func main() {
    pluginFormatter := &PluginFormatter{
        plugins: []FormatterPlugin{
            &GitHubPlugin{},
            // Add more plugins here
        },
        fallback: glamour.DefaultFormatter,
    }
    
    r, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle("dark"),
        glamour.WithLinkFormatter(pluginFormatter),
    )
    
    markdown := `# Plugin Demo
    
- [Glamour](https://github.com/charmbracelet/glamour) 
- [Google](https://google.com)`

    out, _ := r.Render(markdown)
    fmt.Print(out)
    
    // Output:
    // • 🐙 Glamour
    // • Google (https://google.com)
}
```

## Best Practices Examples

### 1. Defensive Programming
```go
func SafeFormatter(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    // Always validate inputs
    if data.URL == "" {
        return data.Text, nil // fallback to text
    }
    
    if data.Text == "" {
        data.Text = data.URL // fallback to URL
    }
    
    // Handle edge cases gracefully
    defer func() {
        if r := recover(); r != nil {
            // Log the panic and return safe fallback
            log.Printf("Formatter panic: %v", r)
        }
    }()
    
    // Your custom formatting logic here
    return fmt.Sprintf("%s [%s]", data.Text, data.URL), nil
}
```

### 2. Performance-Conscious Formatter
```go
func PerformantFormatter(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
    // Pre-allocate buffer for known maximum size
    var buf strings.Builder
    buf.Grow(len(data.Text) + len(data.URL) + 10) // estimate
    
    // Use efficient string building
    buf.WriteString(data.Text)
    buf.WriteString(" (")
    buf.WriteString(data.URL)
    buf.WriteByte(')')
    
    return buf.String(), nil
}
```

### 3. Testing-Friendly Formatter
```go
// FormatterConfig makes testing easier by allowing dependency injection
type FormatterConfig struct {
    HyperlinkDetector func(ctx glamour.RenderContext) bool
    URLShortener      func(string) string
}

func ConfigurableFormatter(config FormatterConfig) glamour.LinkFormatter {
    return glamour.LinkFormatterFunc(func(data glamour.LinkData, ctx glamour.RenderContext) (string, error) {
        if config.HyperlinkDetector != nil && config.HyperlinkDetector(ctx) {
            return glamour.FormatHyperlink(data.Text, data.URL), nil
        }
        
        url := data.URL
        if config.URLShortener != nil {
            url = config.URLShortener(url)
        }
        
        return fmt.Sprintf("%s (%s)", data.Text, url), nil
    })
}

// Test becomes simple:
func TestConfigurableFormatter(t *testing.T) {
    config := FormatterConfig{
        HyperlinkDetector: func(ctx glamour.RenderContext) bool { return false },
        URLShortener:     func(url string) string { return "short.ly/xyz" },
    }
    
    formatter := ConfigurableFormatter(config)
    result, _ := formatter.FormatLink(glamour.LinkData{
        Text: "example",
        URL:  "https://very-long-url.com/path",
    }, glamour.RenderContext{})
    
    assert.Equal(t, "example (short.ly/xyz)", result)
}
```

These examples provide a comprehensive guide for implementing and testing custom link formatters while following Go best practices and maintaining compatibility with the Glamour library.