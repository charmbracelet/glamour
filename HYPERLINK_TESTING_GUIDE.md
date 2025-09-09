# Hyperlink Testing Guide for Glamour

This guide provides comprehensive instructions for testing hyperlink functionality across different terminal environments, validating terminal detection logic, and ensuring proper fallback behavior.

## Table of Contents

- [Overview](#overview)
- [Terminal Compatibility](#terminal-compatibility)
- [Quick Start Testing](#quick-start-testing)
- [Comprehensive Testing](#comprehensive-testing)
- [Development Testing](#development-testing)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Overview

Glamour supports OSC 8 hyperlinks in compatible terminals, providing clickable links that enhance the user experience. When OSC 8 support is not available, Glamour gracefully falls back to displaying links in a `text (url)` format.

### What are OSC 8 Hyperlinks?

OSC 8 (Operating System Command 8) is a terminal escape sequence that creates clickable hyperlinks in supporting terminals. The format is:

```
\x1b]8;;URL\x1b\TEXT\x1b]8;;\x1b\
```

Where:
- `\x1b]8;;URL\x1b\` - Opens the hyperlink with the specified URL
- `TEXT` - The visible text (can include ANSI formatting)
- `\x1b]8;;\x1b\` - Closes the hyperlink

## Terminal Compatibility

### ✅ Fully Supported Terminals

| Terminal | Version | Detection Method | Notes |
|----------|---------|------------------|-------|
| **iTerm2** | 3.1+ | `TERM_PROGRAM=iTerm.app` | Full OSC 8 support, excellent implementation |
| **VS Code** | All | `TERM_PROGRAM=vscode` | Full OSC 8 support in integrated terminal |
| **Windows Terminal** | 1.4+ | `TERM_PROGRAM=Windows Terminal` | Native OSC 8 support |
| **WezTerm** | All | `TERM_PROGRAM=WezTerm` | Excellent OSC 8 implementation |
| **Kitty** | 0.19+ | `TERM=xterm-kitty` or `KITTY_WINDOW_ID` | First terminal to support OSC 8 |
| **Alacritty** | 0.8+ | `TERM=alacritty` or `ALACRITTY_*` vars | Good OSC 8 support |
| **Hyper** | 3.0+ | `TERM_PROGRAM=Hyper` | OSC 8 support via plugins |
| **GNOME Terminal** | 3.36+ | `TERM=xterm-256color` | Recent versions support OSC 8 |

### ⚠️ Partial Support

| Terminal | Detection Method | Notes |
|----------|------------------|-------|
| **tmux/screen** | `TERM=tmux-*` or `screen-*` | Passes through to client terminal |
| **SSH Sessions** | Various | Depends on client terminal capabilities |

### ❌ Unsupported Terminals

| Terminal | Detection Method | Fallback Behavior |
|----------|------------------|-------------------|
| **Terminal.app (macOS)** | `TERM_PROGRAM=Apple_Terminal` | Shows `text (url)` format |
| **Basic xterm** | `TERM=xterm` | Shows `text (url)` format |
| **Console/TTY** | `TERM=linux` or `dumb` | Plain text only |

## Quick Start Testing

### 1. Environment Check

Run this command to see your current terminal environment:

```bash
env | grep -E '^(TERM|COLORTERM|.*_PROGRAM)'
```

### 2. Basic Hyperlink Test

Copy and paste this OSC 8 sequence into your terminal:

```bash
echo -e "\x1b]8;;https://github.com/charmbracelet/glamour\x1b\\Glamour Repository\x1b]8;;\x1b\\"
```

**Expected Results:**
- ✅ **Supporting terminals**: "Glamour Repository" appears as clickable text
- ❌ **Unsupported terminals**: Raw escape sequences are visible

### 3. Glamour Test

Create a test markdown file and render it:

```bash
echo '# Test Document

[GitHub](https://github.com) - Code repository
[Google](https://google.com) - Search engine

Autolink: <https://stackoverflow.com>

Email: <mailto:test@example.com>' > test.md

glamour test.md
```

### 4. Use the Test Tool

Build and run the comprehensive test tool:

```bash
cd cmd/test-hyperlinks
go build -o test-hyperlinks
./test-hyperlinks
```

## Comprehensive Testing

### Testing Script

Use the provided test script for comprehensive validation:

```bash
# Build the test tool
cd cmd/test-hyperlinks
go build -o test-hyperlinks

# Run comprehensive tests
./test-hyperlinks

# Test specific scenarios
TERM_PROGRAM="" TERM="dumb" ./test-hyperlinks  # Test fallback
TERM_PROGRAM="iTerm.app" ./test-hyperlinks     # Test hyperlinks
```

### Manual Test Cases

#### Test Case 1: Basic Links

```markdown
# Basic Links Test

- [Simple Link](https://example.com)
- [Link with Path](https://example.com/path/to/page)
- [Link with Query](https://example.com?param=value&other=test)
- [Link with Fragment](https://example.com/page#section)
```

#### Test Case 2: Special URLs

```markdown
# Special URLs Test

- [Email](mailto:user@example.com)
- [File](file:///etc/hosts)
- [FTP](ftp://ftp.example.com/file.zip)
- [Relative](./relative/path)
```

#### Test Case 3: Unicode and Special Characters

```markdown
# Unicode Test

- [世界のサイト](https://example.com)
- [Emoji Link 🌍](https://example.com)
- [Special & Characters](https://example.com?param=value&other=test)
```

#### Test Case 4: Autolinks

```markdown
# Autolinks Test

- <https://github.com>
- <mailto:hello@example.com>
- <ftp://ftp.example.com>
```

### Expected Behaviors

| Terminal Support | Link Rendering | Click Behavior |
|------------------|----------------|----------------|
| **OSC 8 Supported** | Styled clickable text | Opens URL in browser |
| **Fallback Mode** | `text (url)` format | No click functionality |
| **Styled Fallback** | Colored/styled `text (url)` | No click functionality |

## Development Testing

### Running Unit Tests

```bash
# Terminal detection tests
cd ansi
go test -run TestTerminalDetectionComprehensive -v
go test -run TestSupportsHyperlinks -v

# OSC 8 sequence validation
go test -run TestOSC8SequenceFormat -v
go test -run TestOSC8SequenceTextVariations -v

# Smart fallback behavior
go test -run TestSmartHyperlinkFormatterFallback -v
go test -run TestTerminalDetectionConsistency -v

# All hyperlink-related tests
go test -run Test.*Hyperlink.* -v
go test -run Test.*OSC8.* -v
```

### Benchmarking

```bash
# Performance testing
cd ansi
go test -bench=BenchmarkTerminalDetection -benchmem
go test -bench=BenchmarkOSC8SequenceGeneration -benchmem
go test -bench=BenchmarkSmartFallbackPerformance -benchmem
```

### Testing Different Environments

Simulate different terminal environments:

```bash
# Test iTerm2 environment
TERM_PROGRAM="iTerm.app" TERM="xterm-256color" go test -run TestSupportsHyperlinks

# Test VS Code environment  
TERM_PROGRAM="vscode" TERM="xterm-256color" go test -run TestSupportsHyperlinks

# Test unsupported environment
TERM_PROGRAM="" TERM="dumb" go test -run TestSupportsHyperlinks

# Test edge cases
TERM_PROGRAM="unknown" TERM="xterm" KITTY_WINDOW_ID="1" go test -run TestSupportsHyperlinks
```

## Troubleshooting

### Problem: Links Don't Appear Clickable

**Symptoms:**
- Links are styled but not clickable
- No browser opens when clicking

**Diagnosis:**
1. Check terminal support: Run the quick test above
2. Verify environment: `echo $TERM_PROGRAM $TERM`
3. Check terminal settings for hyperlink support

**Solutions:**
- Update to a supporting terminal version
- Enable hyperlink support in terminal settings
- Use fallback formatting if hyperlinks aren't critical

### Problem: Raw Escape Sequences Visible

**Symptoms:**
- Text like `\x1b]8;;https://example.com\x1b\Link\x1b]8;;\x1b\` appears
- Garbled output in terminal

**Diagnosis:**
1. Terminal doesn't support OSC 8
2. Glamour isn't detecting the terminal correctly
3. Force-hyperlink mode is enabled

**Solutions:**
```go
// Use smart hyperlinks (recommended)
renderer, _ := glamour.NewTermRenderer(
    glamour.WithSmartHyperlinks(),
)

// Or disable hyperlinks entirely
renderer, _ := glamour.NewTermRenderer(
    glamour.WithoutHyperlinks(),
)
```

### Problem: Inconsistent Detection

**Symptoms:**
- Detection varies between runs
- Environment detection seems wrong

**Diagnosis:**
1. Check environment variables: `env | grep -E '^(TERM|.*_PROGRAM)'`
2. Verify environment is stable
3. Check for conflicting settings

**Solutions:**
- Set explicit environment variables
- Use consistent terminal setup
- Report detection bugs with environment details

### Problem: Performance Issues

**Symptoms:**
- Slow rendering with many links
- High memory usage

**Diagnosis:**
1. Run benchmarks: `go test -bench=Benchmark.*Hyperlink.*`
2. Profile memory usage
3. Check for excessive recomputation

**Solutions:**
- Cache terminal detection results
- Use appropriate formatter for use case
- Consider disabling hyperlinks for batch processing

## Environment Variables Reference

### Primary Detection Variables

| Variable | Example Values | Used By |
|----------|----------------|---------|
| `TERM_PROGRAM` | `iTerm.app`, `vscode`, `Windows Terminal` | Primary detection |
| `TERM` | `xterm-256color`, `alacritty`, `xterm-kitty` | Fallback detection |

### Special Detection Variables

| Variable | Purpose | Terminal |
|----------|---------|----------|
| `KITTY_WINDOW_ID` | Kitty terminal detection | Kitty |
| `ALACRITTY_LOG` | Alacritty detection | Alacritty |
| `ALACRITTY_SOCKET` | Alacritty detection | Alacritty |
| `WT_SESSION` | Windows Terminal detection | Windows Terminal |

### Environment Simulation

For testing, you can simulate different terminal environments:

```bash
# Simulate iTerm2
export TERM_PROGRAM="iTerm.app"
export TERM="xterm-256color"

# Simulate basic terminal
unset TERM_PROGRAM
export TERM="xterm"

# Simulate Kitty
unset TERM_PROGRAM  
export TERM="xterm-kitty"
export KITTY_WINDOW_ID="1"
```

## Contributing

### Reporting Issues

When reporting hyperlink-related issues, please include:

1. **Terminal Information:**
   ```bash
   echo "TERM_PROGRAM: $TERM_PROGRAM"
   echo "TERM: $TERM"
   echo "Terminal Version: [manual check]"
   ```

2. **Test Results:**
   ```bash
   # Run the basic test
   echo -e "\x1b]8;;https://example.com\x1b\\Test Link\x1b]8;;\x1b\\"
   
   # Run detection test
   cd cmd/test-hyperlinks && go run main.go | head -20
   ```

3. **Expected vs Actual Behavior**
4. **Operating System and Version**

### Adding Terminal Support

To add support for a new terminal:

1. **Update Detection Logic** in `ansi/hyperlink.go`:
   ```go
   // Add to supportsHyperlinks function
   supportingPrograms := map[string]bool{
       "YourTerminal": true,
   }
   ```

2. **Add Test Cases** in `ansi/terminal_detection_test.go`:
   ```go
   {
       name:        "Your Terminal",
       termProgram: "YourTerminal",
       expected:    true,
       category:    "TERM_PROGRAM",
   },
   ```

3. **Update Documentation** in this file and the compatibility matrix

4. **Test Thoroughly** with real terminal

### Testing Changes

Before submitting changes:

1. Run all tests: `go test ./ansi/...`
2. Run the test tool: `cd cmd/test-hyperlinks && go run main.go`
3. Test in actual terminals
4. Update documentation

---

## Quick Reference

### Test Commands

```bash
# Environment check
env | grep -E '^(TERM|.*_PROGRAM)'

# Basic hyperlink test
echo -e "\x1b]8;;https://github.com\x1b\\GitHub\x1b]8;;\x1b\\"

# Glamour test
echo '[Test](https://example.com)' | glamour

# Comprehensive test
cd cmd/test-hyperlinks && go run main.go

# Unit tests
cd ansi && go test -run Test.*Hyperlink.* -v
```

### Key Functions

- `supportsHyperlinks(ctx)` - Detects terminal hyperlink support
- `formatHyperlink(text, url)` - Generates OSC 8 sequences
- `SmartHyperlinkFormatter` - Adaptive formatter with fallback
- `stripANSISequences(text)` - Removes ANSI sequences for testing

### Environment Detection Priority

1. `TERM_PROGRAM` (most specific)
2. Special environment variables (`KITTY_WINDOW_ID`, etc.)
3. `TERM` patterns (fallback)
4. Conservative default (false)

---

For more information, see the [examples](./examples/) directory and the [test suite](./ansi/).