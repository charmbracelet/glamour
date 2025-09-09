# Terminal Detection Example

This example demonstrates how to detect terminal capabilities and adapt link formatting accordingly. It shows various methods to determine what features your terminal supports, particularly for hyperlink rendering.

## Overview

Different terminals have varying levels of support for advanced features like:
- **OSC 8 Hyperlinks**: Clickable text with hidden URLs
- **256 Colors**: Extended color palette
- **True Color**: 24-bit color support
- **Unicode/Emoji**: Full unicode character support

This example shows how to detect these capabilities and adapt your link formatting to provide the best user experience across different terminal environments.

## Running the Example

```bash
# From the terminal_detection directory
go run main.go

# Or build and run
go build -o detect main.go
./detect
```

## What It Demonstrates

### 1. Environment Variable Detection
The example checks various environment variables that indicate terminal capabilities:

```go
termProgram := os.Getenv("TERM_PROGRAM")     // e.g., "iTerm.app", "vscode"
term := os.Getenv("TERM")                    // e.g., "xterm-256color"
terminalEmulator := os.Getenv("TERMINAL_EMULATOR")
colorTerm := os.Getenv("COLORTERM")          // e.g., "truecolor"
```

### 2. Feature Detection Functions
```go
func detectHyperlinkSupport() bool {
    // Check for known terminals with OSC 8 support
}

func detectColorSupport() bool {
    // Check for color terminal capabilities
}

func detectEmojiSupport() bool {
    // Check if terminal likely supports emoji
}
```

### 3. Adaptive Formatting
Shows a formatter that changes behavior based on detected capabilities:

- **Full Support**: Uses OSC 8 hyperlinks
- **Partial Support**: Uses emoji indicators
- **Basic Support**: Simple text format

## Terminal Compatibility Matrix

| Terminal | Hyperlinks | 256 Color | True Color | Emoji | Notes |
|----------|------------|-----------|------------|-------|-------|
| iTerm2 (macOS) | ✅ | ✅ | ✅ | ✅ | Full support |
| Windows Terminal | ✅ | ✅ | ✅ | ✅ | Full support |
| VS Code Terminal | ✅ | ✅ | ✅ | ✅ | Full support |
| Hyper | ✅ | ✅ | ✅ | ✅ | Full support |
| Terminology | ✅ | ✅ | ✅ | ✅ | Full support |
| macOS Terminal | ❌ | ✅ | ❌ | ✅ | Basic support |
| GNOME Terminal | ✅ | ✅ | ✅ | ✅ | Recent versions |
| Konsole | ✅ | ✅ | ✅ | ✅ | Recent versions |
| SSH Sessions | ❌* | ✅ | ❌* | ✅ | *Depends on client |

## Detection Methods

### Environment Variables to Check

#### TERM_PROGRAM
- `iTerm.app` → iTerm2 (full support)
- `vscode` → VS Code (full support)
- `Hyper` → Hyper terminal (full support)
- `Apple_Terminal` → macOS Terminal (basic support)

#### TERM
- `xterm-256color` → Likely modern terminal
- `screen-256color` → tmux/screen with color
- `xterm` → Basic terminal

#### Other Indicators
- `WT_SESSION` → Windows Terminal
- `COLORTERM=truecolor` → True color support
- `LANG=*.UTF-8` → Unicode support

### Hyperlink Detection Logic

```go
func detectHyperlinkSupport() bool {
    termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
    
    switch {
    case strings.Contains(termProgram, "iterm"):
        return true
    case strings.Contains(termProgram, "vscode"):
        return true
    case os.Getenv("WT_SESSION") != "":
        return true
    default:
        return false
    }
}
```

## Usage in Custom Formatters

```go
adaptiveFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
    // Detect terminal capabilities
    supportsHyperlinks := detectHyperlinkSupport()
    supportsEmoji := detectEmojiSupport()
    
    switch {
    case supportsHyperlinks:
        // Use OSC 8 hyperlinks
        return formatOSC8Hyperlink(data.Text, data.URL), nil
    case supportsEmoji:
        // Use emoji indicators
        return fmt.Sprintf("%s 🔗 %s", data.Text, data.URL), nil
    default:
        // Basic fallback
        return fmt.Sprintf("%s [%s]", data.Text, data.URL), nil
    }
})
```

## Testing Different Terminals

To test hyperlink support in your terminal:

1. **Look for clickable text** in the hyperlink examples
2. **Check for escape sequences** in unsupported terminals
3. **Test with different TERM settings**:
   ```bash
   TERM=xterm-256color go run main.go
   TERM=xterm go run main.go
   ```

## Limitations

- **SSH forwarding**: Hyperlinks may not work through SSH
- **tmux/screen**: May strip hyperlink sequences
- **Old terminal versions**: May not support modern features
- **Detection isn't perfect**: Some terminals may support features but not be detected

## Best Practices

1. **Always provide fallbacks** for unsupported terminals
2. **Test in your target environments** before deploying
3. **Use progressive enhancement** (basic → enhanced features)
4. **Consider user preferences** (some users prefer simple output)
5. **Document terminal requirements** for your applications

## Integration with Glamour

Use these detection methods with Glamour's smart formatters:

```go
// Let Glamour handle detection automatically
renderer, err := glamour.NewTermRenderer(
    glamour.WithSmartHyperlinks(), // Auto-detects and falls back
)

// Or use manual detection for custom behavior
if detectHyperlinkSupport() {
    renderer, err = glamour.NewTermRenderer(glamour.WithHyperlinks())
} else {
    renderer, err = glamour.NewTermRenderer(glamour.WithTextOnlyLinks())
}
```

This approach ensures your applications work well across different terminal environments while providing enhanced experiences where possible.