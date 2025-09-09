# Custom Link Formatting Examples

This example demonstrates the comprehensive custom link formatting capabilities of Glamour, showcasing all available built-in formatters and several custom formatter implementations.

## Overview

Glamour provides flexible link formatting through the `LinkFormatter` interface, allowing you to customize how links are rendered in terminal output. This example covers:

- **Built-in formatters**: Ready-to-use formatters for common use cases
- **Custom formatters**: Examples of implementing your own formatting logic
- **Advanced patterns**: Plugin systems, context awareness, and defensive programming

## Running the Example

```bash
# From the custom_link_formatting directory
go run main.go

# Or build and run
go build -o demo main.go
./demo
```

## Built-in Formatters

### 1. Default Formatter
The standard Glamour behavior showing both text and URL with styling.
```
Visit Google https://google.com for searching.
```

### 2. Text-Only Links (`WithTextOnlyLinks()`)
Shows only clickable text in smart terminals that support hyperlinks.
```
Visit Google for searching.  (clickable in compatible terminals)
```

### 3. URL-Only Links (`WithURLOnlyLinks()`)
Shows only URLs, hiding the descriptive text.
```
Visit https://google.com for searching.
```

### 4. Hyperlink Formatter (`WithHyperlinks()`)
Uses OSC 8 hyperlinks to make text clickable while hiding URLs.
```
Visit Google for searching.  (text is clickable, URL hidden)
```

### 5. Smart Hyperlinks (`WithSmartHyperlinks()`)
OSC 8 hyperlinks with intelligent fallback to default format.
```
Modern terminals: Google (clickable)
Older terminals:  Google https://google.com
```

## Custom Formatter Examples

### 6. Markdown-Style Formatter
Outputs links in markdown format:
```go
markdownFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
})
```

### 7. Domain-Based Formatter
Different icons based on the website domain:
```
🐙 GitHub Repo
🔍 Google Search
📚 Stack Overflow
🔗 Other Site (example.com)
```

### 8. Length-Aware Formatter
Truncates long URLs to keep output clean:
```
Short: Google (https://google.com)
Long:  Very Long URL [example.com...]
```

### 9. Context-Aware Formatter
Different formatting based on link context (tables, autolinks, etc.):
- **Tables**: Text only to save space
- **Autolinks**: `<URL>` format
- **Regular**: `text → URL` format

### 10. Error-Safe Formatter
Demonstrates defensive programming with graceful error handling.

### 11. Plugin-Style Formatter
Extensible system allowing multiple formatting plugins with priority ordering.

## Terminal Compatibility

### Hyperlink Support (OSC 8)
- ✅ **iTerm2** (macOS)
- ✅ **Windows Terminal**
- ✅ **VS Code integrated terminal**
- ✅ **Hyper**
- ✅ **Terminology**
- ❌ **macOS Terminal.app** (basic)
- ❌ **Most SSH sessions**

### Testing Hyperlinks
To test if your terminal supports hyperlinks, look for clickable text in the hyperlink formatter examples. In unsupported terminals, you may see escape sequences or plain text.

## Usage Patterns

### Basic Usage
```go
// Create renderer with custom formatter
renderer, err := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithLinkFormatter(customFormatter),
)

// Render markdown
output, err := renderer.Render(markdown)
fmt.Print(output)
```

### Creating Custom Formatters
```go
// Function-based formatter
customFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    // Your custom formatting logic here
    return fmt.Sprintf("CUSTOM: %s -> %s", data.Text, data.URL), nil
})

// Struct-based formatter implementing LinkFormatter interface
type MyFormatter struct {
    prefix string
}

func (f *MyFormatter) FormatLink(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    return fmt.Sprintf("%s%s (%s)", f.prefix, data.Text, data.URL), nil
}
```

### Available LinkData Fields
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

## Best Practices

1. **Graceful Fallbacks**: Always handle edge cases where URL or text might be empty
2. **Terminal Detection**: Use context to detect terminal capabilities
3. **Performance**: Pre-allocate string builders for complex formatting
4. **Error Handling**: Return meaningful errors and provide safe fallbacks
5. **Testing**: Structure formatters to be easily testable with dependency injection

## Contributing

When creating custom formatters:
1. Implement proper error handling
2. Consider terminal compatibility
3. Test with various markdown inputs
4. Document expected behavior
5. Provide usage examples

## License

Same as the main Glamour project.
