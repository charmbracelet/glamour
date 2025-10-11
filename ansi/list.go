package ansi

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// ListElement renders a list with proper hanging indentation.
type ListElement struct {
	IsNested bool
}

// Render renders a ListElement.
func (e *ListElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := ctx.options.Styles.List.StyleBlock

	if rules.Indent == nil {
		var i uint
		rules.Indent = &i
	}

	if e.IsNested {
		i := ctx.options.Styles.List.LevelIndent
		rules.Indent = &i
	}

	be := BlockElement{
		Block:   &bytes.Buffer{},
		Style:   cascadeStyle(bs.Current().Style, rules, false),
		Margin:  true,
		Newline: !e.IsNested,
	}
	bs.Push(be)

	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, rules.BlockPrefix)
	renderText(bs.Current().Block, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.Prefix)
	return nil
}

// Finish finishes rendering a ListElement with hanging indent support.
func (e *ListElement) Finish(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	rules := bs.Current().Style

	wrapWidth := int(bs.Width(ctx))
	content := bs.Current().Block.String()

	wrapped := wrapListContent(content, wrapWidth)

	mw := NewMarginWriter(ctx, w, rules)
	if _, err := io.WriteString(mw, wrapped); err != nil {
		return fmt.Errorf("glamour: error writing to writer: %w", err)
	}

	if !e.IsNested {
		if _, err := io.WriteString(mw, "\n"); err != nil {
			return fmt.Errorf("glamour: error writing to writer: %w", err)
		}
	}

	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.Suffix)
	renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, rules.BlockSuffix)

	bs.Current().Block.Reset()
	bs.Pop()
	return nil
}

// listItemPrefix represents a detected list item prefix.
type listItemPrefix struct {
	isListItem bool
	width      int
	length     int
}

var numberedListRegex = regexp.MustCompile(`^(\d{1,3})\.\s`)

// detectListItemPrefix detects if a line starts with a list item prefix.
func detectListItemPrefix(plainLine string) listItemPrefix {
	trimmed := strings.TrimLeft(plainLine, " ")
	if trimmed == "" {
		return listItemPrefix{isListItem: false}
	}

	bullets := []string{"• ", "◦ ", "▪ ", "▸ ", "‣ ", "⁃ ", "⁌ ", "⁍ "}
	for _, bullet := range bullets {
		if strings.HasPrefix(trimmed, bullet) {
			return listItemPrefix{
				isListItem: true,
				width:      ansi.StringWidth(bullet),
				length:     len(bullet),
			}
		}
	}

	taskPrefixes := []string{"[✓] ", "[ ] ", "[x] ", "[X] "}
	for _, prefix := range taskPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return listItemPrefix{
				isListItem: true,
				width:      ansi.StringWidth(prefix),
				length:     len(prefix),
			}
		}
	}

	simpleTasks := []string{"✓ ", "✗ ", "☑ ", "☐ "}
	for _, prefix := range simpleTasks {
		if strings.HasPrefix(trimmed, prefix) {
			return listItemPrefix{
				isListItem: true,
				width:      ansi.StringWidth(prefix),
				length:     len(prefix),
			}
		}
	}

	if matches := numberedListRegex.FindStringSubmatch(trimmed); matches != nil {
		fullMatch := matches[0]
		return listItemPrefix{
			isListItem: true,
			width:      ansi.StringWidth(fullMatch),
			length:     len(fullMatch),
		}
	}

	return listItemPrefix{isListItem: false}
}

// wrapListContent wraps list content with proper hanging indentation.
func wrapListContent(content string, wrapWidth int) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, line)
			continue
		}

		plainLine := ansi.Strip(line)
		leadingSpaces := len(plainLine) - len(strings.TrimLeft(plainLine, " "))

		prefix := detectListItemPrefix(plainLine)

		if prefix.isListItem {
			effectiveWidth := wrapWidth - leadingSpaces - prefix.width
			if effectiveWidth < 10 {
				// If we don't have enough space, use a reasonable minimum
				// but be aware this may overflow on very narrow terminals
				effectiveWidth = min(10, max(wrapWidth/2, 5))
			}
			wrapped := wrapListItem(line, effectiveWidth, leadingSpaces, prefix)
			result = append(result, wrapped)
		} else {
			wrapped := ansi.Wordwrap(line, wrapWidth, " ,.;-+|")
			result = append(result, wrapped)
		}
	}

	return strings.Join(result, "\n")
}

// wrapListItem wraps a single list item with hanging indent.
func wrapListItem(line string, effectiveWidth, baseIndent int, prefix listItemPrefix) string {
	plainLine := ansi.Strip(line)
	leadingSpaces := len(plainLine) - len(strings.TrimLeft(plainLine, " "))
	contentStartInPlain := leadingSpaces + prefix.length

	prefixPart, content := splitAtPlainTextPosition(line, contentStartInPlain)

	contentWrapped := ansi.Wordwrap(content, effectiveWidth, " ,.;-+|")
	contentLines := strings.Split(contentWrapped, "\n")

	var lines []string
	lines = append(lines, prefixPart+contentLines[0])

	// Add hanging indent to continuation lines
	hangingIndentStr := strings.Repeat(" ", baseIndent+prefix.width)
	for i := 1; i < len(contentLines); i++ {
		if strings.TrimSpace(contentLines[i]) != "" {
			lines = append(lines, hangingIndentStr+contentLines[i])
		}
	}

	return strings.Join(lines, "\n")
}

// splitAtPlainTextPosition splits an ANSI string at a position in its plain text.
func splitAtPlainTextPosition(ansiString string, plainTextPos int) (before, after string) {
	if plainTextPos <= 0 {
		return "", ansiString
	}

	plainText := ansi.Strip(ansiString)
	if plainTextPos >= len(plainText) {
		return ansiString, ""
	}

	// Build the before part by consuming chars until we hit plainTextPos visible chars
	var beforeBuf strings.Builder
	visibleCount := 0
	i := 0

	for i < len(ansiString) && visibleCount < plainTextPos {
		// Check for ANSI escape sequence
		if i < len(ansiString) && ansiString[i] == '\x1b' {
			escapeStart := i
			i++

			// Handle different escape sequence types
			if i < len(ansiString) {
				switch ansiString[i] {
				case '[': // CSI sequence
					i++
					for i < len(ansiString) && !isCSITerminator(ansiString[i]) {
						i++
					}
					if i < len(ansiString) {
						i++
					}
				case ']': // OSC sequence
					i++
					// OSC sequences end with BEL (\x07) or ST (\x1b\\)
					for i < len(ansiString) {
						if ansiString[i] == '\x07' {
							i++
							break
						}
						if i+1 < len(ansiString) && ansiString[i] == '\x1b' && ansiString[i+1] == '\\' {
							i += 2
							break
						}
						i++
					}
				default:
					if i < len(ansiString) {
						i++
					}
				}
			}

			beforeBuf.WriteString(ansiString[escapeStart:i])
		} else {
			r, size := utf8.DecodeRuneInString(ansiString[i:])
			beforeBuf.WriteRune(r)
			i += size
			visibleCount++
		}
	}

	return beforeBuf.String(), ansiString[i:]
}

// isCSITerminator checks if a byte is a CSI sequence terminator.
func isCSITerminator(b byte) bool {
	return b >= 0x40 && b <= 0x7E
}
