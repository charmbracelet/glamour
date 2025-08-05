package ansi

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// WordwrapWithIndent wraps text to the specified width while preserving indentation
// for continuation lines. This is a fork of ansi.Wordwrap that handles indented
// wrapped text properly for zen-mode reading.
func WordwrapWithIndent(text string, width int, breakChars string, indent string) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if line == "" {
			result = append(result, line)
			continue
		}

		wrapped := wrapLineWithIndent(line, width, breakChars, indent)
		result = append(result, wrapped...)
	}

	return strings.Join(result, "\n")
}

// wrapLineWithIndent wraps a single line, adding indent to continuation lines
func wrapLineWithIndent(line string, width int, breakChars string, indent string) []string {
	if ansi.StringWidth(line) <= width {
		return []string{line}
	}

	var result []string
	// Split on spaces but preserve other characters
	words := strings.Fields(line)

	if len(words) == 0 {
		return []string{line}
	}

	var currentLine strings.Builder
	var currentWidth int
	var wordsOnCurrentLine int

	for i, word := range words {
		wordWidth := ansi.StringWidth(word)
		
		// Calculate space needed (word + space if not last word)
		spaceNeeded := wordWidth
		if i < len(words)-1 {
			spaceNeeded += 1 // for space
		}

		// Check if we need to wrap
		if currentWidth+spaceNeeded > width && wordsOnCurrentLine > 0 {
			// Finish current line
			result = append(result, currentLine.String())
			
			// Start new line with indent (all continuation lines get indented)
			currentLine.Reset()
			currentLine.WriteString(indent)
			currentWidth = ansi.StringWidth(indent)
			wordsOnCurrentLine = 0
		}

		// Add word to current line
		if wordsOnCurrentLine > 0 {
			currentLine.WriteString(" ")
			currentWidth += 1
		}
		currentLine.WriteString(word)
		currentWidth += wordWidth
		wordsOnCurrentLine++
	}

	// Add the last line if it has content
	if currentLine.Len() > 0 {
		result = append(result, currentLine.String())
	}

	return result
}