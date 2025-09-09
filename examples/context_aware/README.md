# Context-Aware Link Formatting Example

This example demonstrates sophisticated context-aware link formatting that adapts based on where links appear in the document and the capabilities of the terminal.

## Overview

Links can appear in various contexts within markdown documents:
- **Regular paragraphs** - Standard text flow
- **Tables** - Space-constrained tabular data
- **Autolinks** - Direct URL references like `<https://example.com>`
- **Block quotes** - Quoted content sections
- **Lists** - Bulleted or numbered items

Each context may benefit from different formatting approaches to optimize readability and space usage.

## Running the Example

```bash
# From the context_aware directory
go run main.go

# Or build and run
go build -o demo main.go
./demo
```

## What It Demonstrates

### 1. Standard Context-Aware Formatter
Basic context switching with different formatting for different contexts:

```go
switch {
case data.IsInTable:
    return data.Text, nil  // Compact for tables
case data.IsAutoLink:
    return fmt.Sprintf("<%s>", data.URL), nil  // Angle brackets
default:
    return fmt.Sprintf("%s → %s", data.Text, data.URL), nil  // Full format
}
```

### 2. Advanced Context-Aware Formatter
More sophisticated logic considering multiple factors:
- Text length (short vs long)
- URL length (compact vs detailed)
- Title presence
- Context combinations

### 3. Table-Optimized Formatter
Specifically designed for tabular data where space is critical:
- Ultra-compact formatting in tables
- Full formatting outside tables
- Handles both regular links and autolinks appropriately

### 4. Progressive Enhancement Formatter
Combines context awareness with terminal capability detection:
- Uses hyperlinks when supported
- Falls back to emoji indicators
- Provides basic text formatting as last resort

### 5. Debug Context Formatter
Development tool that shows context information:
- Displays context flags (TABLE, AUTO, TITLED, etc.)
- Helps understand how context detection works
- Useful for testing custom formatters

## Context Detection

### Available Context Information

The [`LinkData`](../../ansi/link_formatter.go:11) struct provides context through several fields:

```go
type LinkData struct {
    URL        string              // The destination URL
    Text       string              // The link text  
    Title      string              // Optional title attribute
    BaseURL    string              // Base URL for relative links
    IsAutoLink bool                // Whether this is an autolink
    IsInTable  bool                // Whether link appears in a table
    Children   []ElementRenderer   // Original child elements
    LinkStyle  StylePrimitive      // Style for URL portion
    TextStyle  StylePrimitive      // Style for text portion
}
```

### Context-Specific Formatting Strategies

| Context | Strategy | Example Output |
|---------|----------|----------------|
| **Regular** | Full formatting | `GitHub → https://github.com` |
| **Table** | Text only | `GitHub` |
| **Autolink** | Angle brackets | `<https://github.com>` |
| **Long URL** | Domain hint | `Long Article [example.com]` |
| **With Title** | Include title | `GitHub → https://github.com (Code)` |

## Implementation Patterns

### Basic Context Switching

```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    if data.IsInTable {
        // Table-specific formatting
        return data.Text, nil
    }
    
    if data.IsAutoLink {
        // Autolink-specific formatting  
        return fmt.Sprintf("<%s>", data.URL), nil
    }
    
    // Default formatting
    return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
})
```

### Multi-Factor Context Analysis

```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    isShortText := len(data.Text) <= 10
    isLongURL := len(data.URL) > 50
    hasTitle := data.Title != ""
    
    // Combine multiple context factors
    switch {
    case data.IsInTable && isShortText:
        return data.Text, nil
    case data.IsInTable && !isShortText:
        return data.Text[:12] + "...", nil
    case data.IsAutoLink && isLongURL:
        return formatDomainOnly(data.URL), nil
    case hasTitle:
        return fmt.Sprintf("%s → %s (%s)", data.Text, data.URL, data.Title), nil
    default:
        return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
    }
})
```

### Progressive Enhancement

```go
formatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    // Detect terminal capabilities
    supportsHyperlinks := detectHyperlinkSupport()
    supportsEmoji := detectEmojiSupport()
    
    // Combine context and capabilities
    switch {
    case data.IsInTable && supportsHyperlinks:
        return formatHyperlink(data.Text, data.URL), nil
    case data.IsInTable:
        return data.Text, nil
    case data.IsAutoLink && supportsEmoji:
        return fmt.Sprintf("🔗 %s", data.URL), nil
    case supportsHyperlinks:
        return formatHyperlink(data.Text, data.URL), nil
    default:
        return fmt.Sprintf("%s → %s", data.Text, data.URL), nil
    }
})
```

## Use Cases

### Documentation Sites
- **Headers**: Minimal formatting to maintain clean hierarchy
- **Tables**: Ultra-compact to preserve table structure  
- **Paragraphs**: Full formatting for maximum information

### CLI Tools
- **Lists**: Consistent bullet formatting with links
- **Error messages**: Clear, unambiguous link formatting
- **Help text**: Concise formatting that doesn't overwhelm

### Terminal Dashboards
- **Status displays**: Space-efficient formatting
- **Interactive elements**: Hyperlink support when available
- **Log outputs**: Distinguishable link formatting

## Best Practices

### 1. Context Priority
Establish clear priority for context factors:
1. Space constraints (tables, lists)
2. Content type (autolinks, titled links)
3. Terminal capabilities
4. User preferences

### 2. Graceful Degradation
Always provide fallbacks:
```go
switch {
case canUseAdvancedFeature():
    return advancedFormat(data)
case canUseBasicFeature():
    return basicFormat(data)
default:
    return fallbackFormat(data)
}
```

### 3. Consistency Within Context
Maintain consistent formatting within the same context:
- All table links should use the same format
- All autolinks should be handled similarly
- Regular paragraph links should be uniform

### 4. Testing Across Contexts
Test formatters with varied markdown:
```markdown
Regular [link](url) and <autolink>

| Table | [Link](url) |
|-------|-------------|

> Quote with [link](url)
```

## Integration Tips

### With Glamour Renderers
```go
// Create context-aware renderer
renderer, err := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithLinkFormatter(contextAwareFormatter),
)

// Handle different document types
switch docType {
case "table-heavy":
    renderer.SetLinkFormatter(tableOptimizedFormatter)
case "interactive":
    renderer.SetLinkFormatter(hyperlinkFormatter)
}
```

### Custom Context Detection
Extend context detection for specific needs:
```go
type ExtendedLinkData struct {
    ansi.LinkData
    IsInCodeBlock bool
    IsInQuote     bool
    NestingLevel  int
}
```

This example provides a foundation for building sophisticated, context-aware link formatting that adapts to both content structure and user environment.