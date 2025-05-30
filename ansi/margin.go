package ansi

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/padding"
)

// MarginWriter is a Writer that applies indentation and padding around
// whatever you write to it.
type MarginWriter struct {
	w               io.Writer
	pw              *padding.Writer
	iw              *indent.Writer
	ctx             RenderContext
	rules           StyleBlock
	availableWidth  uint
	baseIndentation uint
}


// NewMarginWriter returns a new MarginWriter.
func NewMarginWriter(ctx RenderContext, w io.Writer, rules StyleBlock) *MarginWriter {
	bs := ctx.blockStack

	var indentation uint
	var leftMargin uint
	
	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	
	// Handle legacy margin (applies to both sides)
	if rules.Margin != nil {
		leftMargin = *rules.Margin
	}
	
	// Override with specific left margin if provided
	if rules.MarginLeft != nil {
		leftMargin = *rules.MarginLeft
	}
	
	// Handle special alignment cases
	if rules.Align != nil && (*rules.Align == "center" || *rules.Align == "justify") {
		// Note: For center/justify alignment, we'll apply alignment logic in the Write method
		// since we need to measure content width first
		leftMargin = 0 // Will be calculated per-line during writing
	}

	pw := padding.NewWriterPipe(w, bs.Width(ctx), func(_ io.Writer) {
		renderText(w, ctx.options.ColorProfile, rules.StylePrimitive, " ")
	})

	ic := " "
	if rules.IndentToken != nil {
		ic = *rules.IndentToken
	}
	
	// For special alignment (center/justify), we need special handling
	if rules.Align != nil && (*rules.Align == "center" || *rules.Align == "justify") {
		return &MarginWriter{
			w:               w,
			pw:              pw,
			iw:              nil, // We'll handle indentation manually for special alignment
			ctx:             ctx,
			rules:           rules,
			availableWidth:  bs.Width(ctx),
			baseIndentation: indentation,
		}
	}

	// For non-center alignment, use the standard approach
	totalLeftIndent := indentation + leftMargin
	iw := indent.NewWriterPipe(pw, totalLeftIndent, func(_ io.Writer) {
		renderText(w, ctx.options.ColorProfile, bs.Parent().Style.StylePrimitive, ic)
	})

	return &MarginWriter{
		w:               w,
		pw:              pw,
		iw:              iw,
		ctx:             ctx,
		rules:           rules,
		availableWidth:  bs.Width(ctx),
		baseIndentation: indentation,
	}
}

func (w *MarginWriter) Write(b []byte) (int, error) {
	// Handle special alignment cases
	if w.rules.Align != nil {
		switch *w.rules.Align {
		case "center":
			return w.writeCentered(b)
		case "justify":
			return w.writeJustified(b)
		}
	}

	// Standard writing for left alignment
	if w.iw == nil {
		return 0, fmt.Errorf("glamour: indent writer not initialized")
	}
	n, err := w.iw.Write(b)
	if err != nil {
		return 0, fmt.Errorf("glamour: error writing bytes: %w", err)
	}
	return n, nil
}

func (w *MarginWriter) writeCentered(b []byte) (int, error) {
	content := string(b)
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		// Measure the actual display width of the line
		lineWidth := ansi.StringWidth(line)
		
		// Calculate centering margin
		var leftMargin uint
		if lineWidth < int(w.availableWidth) {
			leftMargin = (w.availableWidth - uint(lineWidth)) / 2
		}
		
		// Add base indentation
		totalIndent := w.baseIndentation + leftMargin
		
		// Apply indentation
		indentStr := strings.Repeat(" ", int(totalIndent))
		centeredLine := indentStr + line
		
		// Write the centered line
		if _, err := w.pw.Write([]byte(centeredLine)); err != nil {
			return 0, fmt.Errorf("glamour: error writing centered line: %w", err)
		}
		
		// Add newline for all lines except the last one (if it didn't originally have one)
		if i < len(lines)-1 || (i == len(lines)-1 && strings.HasSuffix(content, "\n")) {
			if _, err := w.pw.Write([]byte("\n")); err != nil {
				return 0, fmt.Errorf("glamour: error writing newline: %w", err)
			}
		}
	}
	
	return len(b), nil
}

func (w *MarginWriter) writeJustified(b []byte) (int, error) {
	content := string(b)
	lines := strings.Split(content, "\n")
	
	// Get margins
	var leftMargin, rightMargin uint
	if w.rules.MarginLeft != nil {
		leftMargin = *w.rules.MarginLeft
	}
	if w.rules.MarginRight != nil {
		rightMargin = *w.rules.MarginRight
	}
	
	// Calculate effective width for justification
	effectiveWidth := w.availableWidth
	if leftMargin+rightMargin < effectiveWidth {
		effectiveWidth = effectiveWidth - leftMargin - rightMargin
	}
	
	for i, line := range lines {
		// Apply left margin and base indentation
		totalLeftIndent := w.baseIndentation + leftMargin
		
		// For justified text, we need to stretch the line to fill the effective width
		justifiedLine := w.justifyLine(line, int(effectiveWidth))
		
		// Apply indentation
		indentStr := strings.Repeat(" ", int(totalLeftIndent))
		finalLine := indentStr + justifiedLine
		
		// Write the justified line
		if _, err := w.pw.Write([]byte(finalLine)); err != nil {
			return 0, fmt.Errorf("glamour: error writing justified line: %w", err)
		}
		
		// Add newline for all lines except the last one (if it didn't originally have one)
		if i < len(lines)-1 || (i == len(lines)-1 && strings.HasSuffix(content, "\n")) {
			if _, err := w.pw.Write([]byte("\n")); err != nil {
				return 0, fmt.Errorf("glamour: error writing newline: %w", err)
			}
		}
	}
	
	return len(b), nil
}

// justifyLine distributes spaces evenly across a line to fill the target width
func (w *MarginWriter) justifyLine(line string, targetWidth int) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return line
	}
	
	// Measure current line width
	currentWidth := ansi.StringWidth(line)
	if currentWidth >= targetWidth {
		return line // Line is already full width or longer
	}
	
	// Don't justify short lines (less than 60% of target width)
	// This prevents awkward justification of paragraph endings
	if currentWidth < int(float64(targetWidth)*0.6) {
		return line
	}
	
	// Split into words
	words := strings.Fields(line)
	if len(words) <= 1 {
		return line // Can't justify single word or empty line
	}
	
	// Calculate how much space to distribute
	spacesToAdd := targetWidth - currentWidth
	gaps := len(words) - 1
	if gaps == 0 {
		return line
	}
	
	// Distribute extra spaces evenly
	baseSpaces := spacesToAdd / gaps
	extraSpaces := spacesToAdd % gaps
	
	var result strings.Builder
	for i, word := range words {
		result.WriteString(word)
		if i < len(words)-1 { // Not the last word
			// Add normal space plus base extra spaces
			result.WriteString(strings.Repeat(" ", 1+baseSpaces))
			// Add one extra space to first 'extraSpaces' gaps
			if i < extraSpaces {
				result.WriteString(" ")
			}
		}
	}
	
	return result.String()
}
