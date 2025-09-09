# Custom Link Formatting - Unit Test Plan

## Overview

This document outlines comprehensive unit tests for the custom link formatting feature. Tests are organized by component and ensure both functionality and backward compatibility.

## Test Structure

### 1. Core Interface Tests

#### LinkFormatter Interface Tests
**File**: `ansi/link_formatter_test.go`

```go
func TestLinkFormatterInterface(t *testing.T) {
    // Test that LinkFormatterFunc implements LinkFormatter
    formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
        return "test", nil
    })
    
    var _ LinkFormatter = formatter
    
    // Test basic functionality
    result, err := formatter.FormatLink(LinkData{}, RenderContext{})
    assert.NoError(t, err)
    assert.Equal(t, "test", result)
}

func TestLinkFormatterError(t *testing.T) {
    formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
        return "", errors.New("formatter error")
    })
    
    _, err := formatter.FormatLink(LinkData{}, RenderContext{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "formatter error")
}
```

#### LinkData Structure Tests
```go
func TestLinkData(t *testing.T) {
    tests := []struct {
        name string
        data LinkData
        want LinkData
    }{
        {
            name: "complete link data",
            data: LinkData{
                URL:        "https://example.com",
                Text:       "Example",
                Title:      "Example Site",
                BaseURL:    "https://base.com",
                IsAutoLink: false,
                IsInTable:  false,
            },
        },
        {
            name: "autolink data",
            data: LinkData{
                URL:        "https://example.com",
                Text:       "https://example.com",
                IsAutoLink: true,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test that all fields are properly set and accessible
            assert.Equal(t, tt.want.URL, tt.data.URL)
            assert.Equal(t, tt.want.Text, tt.data.Text)
            // ... other assertions
        })
    }
}
```

### 2. Built-in Formatter Tests

#### Default Formatter Tests
```go
func TestDefaultFormatter(t *testing.T) {
    tests := []struct {
        name string
        data LinkData
        want string
    }{
        {
            name: "text and url",
            data: LinkData{
                URL:  "https://example.com",
                Text: "Example",
            },
            want: "Example https://example.com", // Current behavior
        },
        {
            name: "autolink",
            data: LinkData{
                URL:        "https://example.com",
                Text:       "https://example.com",
                IsAutoLink: true,
            },
            want: "https://example.com https://example.com",
        },
        {
            name: "empty text",
            data: LinkData{
                URL:  "https://example.com",
                Text: "",
            },
            want: " https://example.com", // Space + URL
        },
    }
    
    ctx := RenderContext{
        options: Options{
            Styles: getDefaultStyles(),
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := DefaultFormatter.FormatLink(tt.data, ctx)
            assert.NoError(t, err)
            assert.Contains(t, result, tt.want) // Contains due to ANSI codes
        })
    }
}
```

#### Text-Only Formatter Tests
```go
func TestTextOnlyFormatter(t *testing.T) {
    tests := []struct {
        name            string
        data            LinkData
        supportsHyperlinks bool
        wantContains    string
    }{
        {
            name: "hyperlink support",
            data: LinkData{
                URL:  "https://example.com",
                Text: "Example",
            },
            supportsHyperlinks: true,
            wantContains: "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\",
        },
        {
            name: "no hyperlink support",
            data: LinkData{
                URL:  "https://example.com", 
                Text: "Example",
            },
            supportsHyperlinks: false,
            wantContains: "Example", // Just text, no URL
        },
        {
            name: "empty text fallback",
            data: LinkData{
                URL:  "https://example.com",
                Text: "",
            },
            supportsHyperlinks: true,
            wantContains: "https://example.com", // Falls back to URL
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Mock terminal support
            originalTermProgram := os.Getenv("TERM_PROGRAM")
            if tt.supportsHyperlinks {
                os.Setenv("TERM_PROGRAM", "iTerm.app")
            } else {
                os.Setenv("TERM_PROGRAM", "")
            }
            defer os.Setenv("TERM_PROGRAM", originalTermProgram)
            
            result, err := TextOnlyFormatter.FormatLink(tt.data, RenderContext{})
            assert.NoError(t, err)
            assert.Contains(t, result, tt.wantContains)
        })
    }
}
```

#### URL-Only Formatter Tests
```go
func TestURLOnlyFormatter(t *testing.T) {
    data := LinkData{
        URL:  "https://example.com",
        Text: "Example Text",
    }
    
    result, err := URLOnlyFormatter.FormatLink(data, RenderContext{})
    assert.NoError(t, err)
    assert.Contains(t, result, "https://example.com")
    assert.NotContains(t, result, "Example Text")
}
```

#### Hyperlink Formatter Tests  
```go
func TestHyperlinkFormatter(t *testing.T) {
    tests := []struct {
        name string
        data LinkData
        want string
    }{
        {
            name: "normal link",
            data: LinkData{
                URL:  "https://example.com",
                Text: "Example",
            },
            want: "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\",
        },
        {
            name: "link with title",
            data: LinkData{
                URL:   "https://example.com",
                Text:  "Example",
                Title: "Example Site",
            },
            want: "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := HyperlinkFormatter.FormatLink(tt.data, RenderContext{})
            assert.NoError(t, err)
            assert.Contains(t, result, tt.want)
        })
    }
}
```

### 3. Integration Tests

#### TermRendererOption Tests
**File**: `glamour_test.go` (addition)

```go
func TestWithLinkFormatter(t *testing.T) {
    customFormatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
        return fmt.Sprintf("CUSTOM[%s](%s)", data.Text, data.URL), nil
    })
    
    r, err := NewTermRenderer(
        WithStandardStyle("dark"),
        WithLinkFormatter(customFormatter),
    )
    assert.NoError(t, err)
    
    markdown := "[example](https://example.com)"
    result, err := r.Render(markdown)
    assert.NoError(t, err)
    assert.Contains(t, result, "CUSTOM[example](https://example.com)")
}

func TestWithTextOnlyLinks(t *testing.T) {
    r, err := NewTermRenderer(
        WithStandardStyle("dark"),
        WithTextOnlyLinks(),
    )
    assert.NoError(t, err)
    
    markdown := "[example](https://example.com)"
    result, err := r.Render(markdown)
    assert.NoError(t, err)
    
    // Should contain text but not the URL as separate text
    assert.Contains(t, result, "example")
    // Should not contain URL as visible text (may be in hyperlink escape)
    assert.NotRegexp(t, `https://example\.com(?!\x1b)`, result)
}

func TestWithHyperlinks(t *testing.T) {
    r, err := NewTermRenderer(
        WithStandardStyle("dark"),
        WithHyperlinks(),
    )
    assert.NoError(t, err)
    
    markdown := "[example](https://example.com)"
    result, err := r.Render(markdown)
    assert.NoError(t, err)
    
    // Should contain OSC 8 sequences
    assert.Contains(t, result, "\x1b]8;;")
}
```

#### LinkElement Integration Tests
**File**: `ansi/link_test.go`

```go
func TestLinkElementWithCustomFormatter(t *testing.T) {
    formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
        return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
    })
    
    element := &LinkElement{
        URL:       "https://example.com",
        Children:  []ElementRenderer{&BaseElement{Token: "example"}},
        Formatter: formatter,
    }
    
    var buf bytes.Buffer
    ctx := RenderContext{
        options: Options{
            Styles: getDefaultStyles(),
        },
    }
    
    err := element.Render(&buf, ctx)
    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "[example](https://example.com)")
}

func TestLinkElementWithoutFormatter(t *testing.T) {
    element := &LinkElement{
        URL:      "https://example.com",
        Children: []ElementRenderer{&BaseElement{Token: "example"}},
        // No Formatter - should use default behavior
    }
    
    var buf bytes.Buffer
    ctx := RenderContext{
        options: Options{
            Styles: getDefaultStyles(),
        },
    }
    
    err := element.Render(&buf, ctx)
    assert.NoError(t, err)
    
    result := buf.String()
    // Should contain both text and URL (default behavior)
    assert.Contains(t, result, "example")
    assert.Contains(t, result, "https://example.com")
}
```

### 4. OSC 8 Hyperlink Tests

#### Hyperlink Generation Tests
**File**: `ansi/hyperlink_test.go`

```go
func TestFormatHyperlink(t *testing.T) {
    tests := []struct {
        name string
        text string
        url  string
        want string
    }{
        {
            name: "basic hyperlink",
            text: "example",
            url:  "https://example.com",
            want: "\x1b]8;;https://example.com\x1b\\example\x1b]8;;\x1b\\",
        },
        {
            name: "empty text",
            text: "",
            url:  "https://example.com", 
            want: "\x1b]8;;https://example.com\x1b\\\x1b]8;;\x1b\\",
        },
        {
            name: "special characters in URL",
            text: "example",
            url:  "https://example.com/path?param=value&other=test",
            want: "\x1b]8;;https://example.com/path?param=value&other=test\x1b\\example\x1b]8;;\x1b\\",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatHyperlink(tt.text, tt.url)
            assert.Equal(t, tt.want, result)
        })
    }
}
```

#### Terminal Detection Tests
```go
func TestSupportsHyperlinks(t *testing.T) {
    tests := []struct {
        name        string
        termProgram string
        term        string
        want        bool
    }{
        {
            name:        "iTerm2",
            termProgram: "iTerm.app",
            want:        true,
        },
        {
            name:        "VS Code",
            termProgram: "vscode",
            want:        true,
        },
        {
            name:        "Windows Terminal",
            termProgram: "Windows Terminal",
            want:        true,
        },
        {
            name:        "WezTerm",
            termProgram: "WezTerm",
            want:        true,
        },
        {
            name: "256 color terminal",
            term: "xterm-256color",
            want: true,
        },
        {
            name: "true color terminal",
            term: "xterm-truecolor",
            want: true,
        },
        {
            name: "basic terminal",
            term: "xterm",
            want: false,
        },
        {
            name: "unknown terminal",
            want: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Save original values
            originalTermProgram := os.Getenv("TERM_PROGRAM")
            originalTerm := os.Getenv("TERM")
            
            // Set test values
            os.Setenv("TERM_PROGRAM", tt.termProgram)
            os.Setenv("TERM", tt.term)
            
            // Test
            result := supportsHyperlinks(RenderContext{})
            assert.Equal(t, tt.want, result)
            
            // Restore
            os.Setenv("TERM_PROGRAM", originalTermProgram)
            os.Setenv("TERM", originalTerm)
        })
    }
}
```

### 5. Backward Compatibility Tests

#### Regression Tests
**File**: `ansi/renderer_test.go` (additions)

```go
func TestBackwardCompatibility(t *testing.T) {
    // Test that all existing golden files still pass
    testCases := []string{
        "link.golden",
        // Add other golden files that contain links
    }
    
    for _, testCase := range testCases {
        t.Run(testCase, func(t *testing.T) {
            // Read the markdown input and expected output
            input := readTestFile(t, testCase+".md")
            expected := readTestFile(t, testCase)
            
            // Render with default settings (no custom formatter)
            r, err := NewTermRenderer(WithStandardStyle("dark"))
            assert.NoError(t, err)
            
            result, err := r.Render(input)
            assert.NoError(t, err)
            
            // Result should match existing golden file exactly
            assert.Equal(t, expected, result)
        })
    }
}

func TestDefaultFormatterMatchesCurrentBehavior(t *testing.T) {
    markdown := "[example](https://example.com)"
    
    // Render without custom formatter (current behavior)
    r1, _ := NewTermRenderer(WithStandardStyle("dark"))
    result1, _ := r1.Render(markdown)
    
    // Render with explicit default formatter
    r2, _ := NewTermRenderer(
        WithStandardStyle("dark"),
        WithLinkFormatter(DefaultFormatter),
    )
    result2, _ := r2.Render(markdown)
    
    // Results should be identical
    assert.Equal(t, result1, result2)
}
```

#### Performance Tests
```go
func BenchmarkLinkRenderingDefault(b *testing.B) {
    r, _ := NewTermRenderer(WithStandardStyle("dark"))
    markdown := "[example](https://example.com)"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        r.Render(markdown)
    }
}

func BenchmarkLinkRenderingCustomFormatter(b *testing.B) {
    formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
        return fmt.Sprintf("%s %s", data.Text, data.URL), nil
    })
    
    r, _ := NewTermRenderer(
        WithStandardStyle("dark"),
        WithLinkFormatter(formatter),
    )
    markdown := "[example](https://example.com)"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        r.Render(markdown)
    }
}
```

### 6. Edge Cases and Error Handling

#### Error Handling Tests
```go
func TestFormatterErrorHandling(t *testing.T) {
    errorFormatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
        return "", errors.New("formatter error")
    })
    
    element := &LinkElement{
        URL:       "https://example.com",
        Children:  []ElementRenderer{&BaseElement{Token: "example"}},
        Formatter: errorFormatter,
    }
    
    var buf bytes.Buffer
    err := element.Render(&buf, RenderContext{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "formatter error")
}

func TestInvalidURLHandling(t *testing.T) {
    tests := []struct {
        name string
        url  string
    }{
        {"empty URL", ""},
        {"malformed URL", "://invalid"},
        {"just fragment", "#fragment"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            data := LinkData{
                URL:  tt.url,
                Text: "example",
            }
            
            // Should not panic
            result, err := DefaultFormatter.FormatLink(data, RenderContext{})
            assert.NoError(t, err)
            assert.NotEmpty(t, result)
        })
    }
}
```

#### Context Variations Tests
```go
func TestLinkInTableContext(t *testing.T) {
    data := LinkData{
        URL:       "https://example.com",
        Text:      "example",
        IsInTable: true,
    }
    
    formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
        if data.IsInTable {
            return data.Text, nil // Table links: text only
        }
        return fmt.Sprintf("%s (%s)", data.Text, data.URL), nil
    })
    
    result, err := formatter.FormatLink(data, RenderContext{})
    assert.NoError(t, err)
    assert.Equal(t, "example", result)
    assert.NotContains(t, result, "https://example.com")
}

func TestAutoLinkContext(t *testing.T) {
    data := LinkData{
        URL:        "https://example.com",
        Text:       "https://example.com",
        IsAutoLink: true,
    }
    
    formatter := LinkFormatterFunc(func(data LinkData, ctx RenderContext) (string, error) {
        if data.IsAutoLink {
            return fmt.Sprintf("<%s>", data.URL), nil
        }
        return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
    })
    
    result, err := formatter.FormatLink(data, RenderContext{})
    assert.NoError(t, err)
    assert.Equal(t, "<https://example.com>", result)
}
```

### 7. Test Utilities

#### Helper Functions
```go
// Test utilities file: ansi/test_utils.go

func getDefaultStyles() StyleConfig {
    // Return default styling configuration for tests
}

func readTestFile(t *testing.T, filename string) string {
    data, err := ioutil.ReadFile(filepath.Join("testdata", filename))
    require.NoError(t, err)
    return string(data)
}

func createTestRenderer(options ...TermRendererOption) (*TermRenderer, error) {
    defaultOptions := []TermRendererOption{WithStandardStyle("dark")}
    return NewTermRenderer(append(defaultOptions, options...)...)
}
```

## Test Coverage Requirements

1. **Interface Coverage**: 100% of LinkFormatter interface methods
2. **Formatter Coverage**: 100% of all built-in formatters
3. **Integration Coverage**: All TermRendererOption functions
4. **Edge Cases**: Error conditions, invalid inputs, empty values
5. **Backward Compatibility**: All existing golden tests pass
6. **Performance**: No regression in rendering speed

## Test Organization

```
ansi/
├── link_formatter_test.go      # Core interface and formatter tests
├── hyperlink_test.go           # OSC 8 and terminal detection tests  
├── link_test.go                # LinkElement integration tests
└── renderer_test.go            # Regression and compatibility tests

glamour_test.go                 # TermRendererOption integration tests

examples/
└── custom_formatter_test.go    # Example usage tests
```

This comprehensive test plan ensures robust functionality while maintaining complete backward compatibility.