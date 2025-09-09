# Custom Link Formatting Examples

This directory contains comprehensive examples demonstrating the custom link formatting capabilities introduced in Glamour. These examples show how to create, configure, and use custom link formatters to control how links appear in terminal output.

## 📁 Example Directories

### [`custom_link_formatting/`](custom_link_formatting/)
**Main comprehensive demo** - Shows all built-in formatters and several custom implementations:
- ✅ Built-in formatters (Default, TextOnly, URLOnly, Hyperlinks, SmartHyperlinks)  
- ✅ Custom formatters (Markdown-style, Domain-based, Length-aware, Context-aware)
- ✅ Advanced patterns (Plugin system, Error handling, Progressive enhancement)

```bash
cd custom_link_formatting && go run main.go
```

### [`terminal_detection/`](terminal_detection/)
**Terminal capability detection** - Learn how to detect and adapt to different terminal capabilities:
- ✅ Environment variable analysis
- ✅ Hyperlink support detection
- ✅ Color and emoji capability checking
- ✅ Adaptive formatting based on terminal features

```bash
cd terminal_detection && go run main.go
```

### [`context_aware/`](context_aware/)
**Context-sensitive formatting** - Advanced formatting that adapts based on link context:
- ✅ Table vs paragraph formatting
- ✅ Autolink vs regular link handling
- ✅ Length-based adaptations
- ✅ Debug formatter for development

```bash
cd context_aware && go run main.go
```

## 🚀 Quick Start

### Basic Usage
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/glamour/ansi"
)

func main() {
    // Create a custom formatter
    customFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
        return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
    })

    // Use it with Glamour
    renderer, _ := glamour.NewTermRenderer(
        glamour.WithLinkFormatter(customFormatter),
    )

    markdown := "[Example](https://example.com)"
    output, _ := renderer.Render(markdown)
    fmt.Print(output)
    // Output: [Example](https://example.com)
}
```

### Built-in Formatters
```go
// Text-only links (clickable in smart terminals)
renderer, _ := glamour.NewTermRenderer(
    glamour.WithTextOnlyLinks(),
)

// URL-only links  
renderer, _ := glamour.NewTermRenderer(
    glamour.WithURLOnlyLinks(),
)

// OSC 8 hyperlinks
renderer, _ := glamour.NewTermRenderer(
    glamour.WithHyperlinks(),
)

// Smart hyperlinks with fallback
renderer, _ := glamour.NewTermRenderer(
    glamour.WithSmartHyperlinks(),
)
```

## 📖 Key Concepts

### LinkFormatter Interface
```go
type LinkFormatter interface {
    FormatLink(data LinkData, ctx RenderContext) (string, error)
}
```

### LinkData Structure
The [`LinkData`](../ansi/link_formatter.go) struct provides comprehensive context:
```go
type LinkData struct {
    URL        string              // Destination URL
    Text       string              // Link text
    Title      string              // Optional title attribute
    BaseURL    string              // Base URL for relative links
    IsAutoLink bool                // Whether this is an autolink
    IsInTable  bool                // Whether link appears in table
    Children   []ElementRenderer   // Original child elements
    LinkStyle  StylePrimitive      // Style for URL portion
    TextStyle  StylePrimitive      // Style for text portion
}
```

### Function-Based Formatters
```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    // Your custom logic here
    return formattedString, nil
})
```

### Struct-Based Formatters
```go
type MyFormatter struct {
    prefix string
}

func (f *MyFormatter) FormatLink(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    return fmt.Sprintf("%s%s (%s)", f.prefix, data.Text, data.URL), nil
}
```

## 🏆 Formatter Examples

### 1. Domain-Based Formatting
Different icons/formatting based on the website domain:
```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    switch {
    case strings.Contains(data.URL, "github.com"):
        return fmt.Sprintf("🐙 %s", data.Text), nil
    case strings.Contains(data.URL, "google.com"):
        return fmt.Sprintf("🔍 %s", data.Text), nil
    default:
        return fmt.Sprintf("🔗 %s", data.Text), nil
    }
})
```

### 2. Context-Aware Formatting
Adapt formatting based on where the link appears:
```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    switch {
    case data.IsInTable:
        return data.Text, nil // Compact for tables
    case data.IsAutoLink:
        return fmt.Sprintf("<%s>", data.URL), nil
    default:
        return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
    }
})
```

### 3. Progressive Enhancement
Adapt to terminal capabilities:
```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    if supportsHyperlinks(ctx) {
        return formatHyperlink(data.Text, data.URL), nil
    }
    return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
})
```

## 🖥️ Terminal Compatibility

### Hyperlink Support (OSC 8)
| Terminal | Support | Notes |
|----------|---------|-------|
| iTerm2 | ✅ | Full support |
| Windows Terminal | ✅ | Full support |
| VS Code Terminal | ✅ | Full support |
| Hyper | ✅ | Full support |
| GNOME Terminal | ✅ | Recent versions |
| macOS Terminal | ❌ | Basic support only |
| SSH Sessions | ❌* | *Depends on client |

### Testing Hyperlinks
Use the [`terminal_detection`](terminal_detection/) example to test your terminal's capabilities.

## ⚡ Performance Tips

### 1. Pre-allocate Buffers
```go
func efficientFormatter(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    var buf strings.Builder
    buf.Grow(len(data.Text) + len(data.URL) + 10) // Pre-allocate
    
    buf.WriteString(data.Text)
    buf.WriteString(" -> ")
    buf.WriteString(data.URL)
    
    return buf.String(), nil
}
```

### 2. Cache Expensive Operations
```go
type CachingFormatter struct {
    cache map[string]string
    mu    sync.RWMutex
}

func (f *CachingFormatter) FormatLink(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    key := data.URL
    f.mu.RLock()
    if cached, ok := f.cache[key]; ok {
        f.mu.RUnlock()
        return cached, nil
    }
    f.mu.RUnlock()
    
    // Expensive operation here
    result := expensiveFormat(data)
    
    f.mu.Lock()
    f.cache[key] = result
    f.mu.Unlock()
    
    return result, nil
}
```

## 🛡️ Error Handling Best Practices

### 1. Graceful Fallbacks
```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    // Validate inputs
    if data.URL == "" {
        return data.Text, nil // Fallback to text
    }
    
    if data.Text == "" {
        data.Text = data.URL // Fallback to URL
    }
    
    // Your formatting logic
    return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
})
```

### 2. Defensive Programming
```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Formatter panic: %v", r)
        }
    }()
    
    // URL parsing with error handling
    u, err := url.Parse(data.URL)
    if err != nil {
        return fmt.Sprintf("%s [invalid URL]", data.Text), nil
    }
    
    return fmt.Sprintf("%s [%s]", data.Text, u.Hostname()), nil
})
```

### 3. Meaningful Errors
```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    if data.URL == "forbidden" {
        return "", fmt.Errorf("link formatting failed: forbidden URL pattern")
    }
    
    // Continue with formatting...
    return result, nil
})
```

## 🧪 Testing Custom Formatters

### Unit Tests
```go
func TestCustomFormatter(t *testing.T) {
    formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
        return fmt.Sprintf("TEST: %s -> %s", data.Text, data.URL), nil
    })
    
    data := ansi.LinkData{
        Text: "example",
        URL:  "https://example.com",
    }
    
    result, err := formatter.FormatLink(data, ansi.RenderContext{})
    assert.NoError(t, err)
    assert.Equal(t, "TEST: example -> https://example.com", result)
}
```

### Integration Tests
```go
func TestFormatterWithRenderer(t *testing.T) {
    formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
        return fmt.Sprintf("CUSTOM: %s", data.Text), nil
    })
    
    renderer, err := glamour.NewTermRenderer(
        glamour.WithLinkFormatter(formatter),
    )
    require.NoError(t, err)
    
    markdown := "[test](https://example.com)"
    result, err := renderer.Render(markdown)
    require.NoError(t, err)
    
    assert.Contains(t, result, "CUSTOM: test")
}
```

## 📚 Additional Resources

- **[CUSTOM_LINK_FORMATTING_ARCHITECTURE.md](../CUSTOM_LINK_FORMATTING_ARCHITECTURE.md)** - Implementation details and architecture
- **[CUSTOM_LINK_FORMATTING_DOCUMENTATION.md](../CUSTOM_LINK_FORMATTING_DOCUMENTATION.md)** - Complete API documentation  
- **[CUSTOM_LINK_FORMATTING_EXAMPLES.md](../CUSTOM_LINK_FORMATTING_EXAMPLES.md)** - Reference implementation examples
- **[CUSTOM_LINK_FORMATTING_TESTS.md](../CUSTOM_LINK_FORMATTING_TESTS.md)** - Testing strategies and examples

## 🤝 Contributing

When adding new formatter examples:

1. **Include comprehensive documentation** with usage examples
2. **Add proper error handling** and graceful fallbacks  
3. **Test with various markdown inputs** and terminal types
4. **Follow naming conventions** and coding standards
5. **Update this index** with new examples

## ⚠️ Known Limitations

- **SSH forwarding**: Hyperlinks may not work through SSH
- **tmux/screen**: May strip hyperlink sequences
- **Terminal detection**: Not 100% accurate for all terminals
- **Performance**: Complex formatters may impact rendering speed

## 📄 License

These examples follow the same license as the main Glamour project.