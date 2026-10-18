package glamour

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
)

const tableBorderMarkdown = `
| Header A  | Header B  |
| --------- | --------- |
| Cell 1    | Cell 2    |
| Cell 3    | Cell 4    |
`

// separatorLines counts the horizontal rule lines in a rendered table: lines
// made up solely of border runes (dashes, column/cross separators, spaces)
// that contain at least one dash.
func separatorLines(rendered string) int {
	n := 0
	for _, line := range strings.Split(rendered, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.ContainsRune(trimmed, '-') && strings.Trim(trimmed, "-+|─┼ ") == "" {
			n++
		}
	}
	return n
}

func renderTable(t *testing.T, mutate func(*ansi.StyleConfig)) string {
	t.Helper()
	style := styles.ASCIIStyleConfig
	if mutate != nil {
		mutate(&style)
	}
	renderer, err := NewTermRenderer(WithStyles(style), WithWordWrap(80))
	if err != nil {
		t.Fatal(err)
	}
	out, err := renderer.Render(strings.TrimSpace(tableBorderMarkdown))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestTableDefaultHasNoRowBorder pins the historical default: a table renders
// only the header rule, with no divider between body rows.
func TestTableDefaultHasNoRowBorder(t *testing.T) {
	if got := separatorLines(renderTable(t, nil)); got != 1 {
		t.Errorf("default table want 1 separator line (header rule only), got %d", got)
	}
}

// TestTableRowBorderDrawsDividers verifies RowBorder=true adds a divider
// between every pair of body rows on top of the header rule.
func TestTableRowBorderDrawsDividers(t *testing.T) {
	out := renderTable(t, func(c *ansi.StyleConfig) {
		v := true
		c.Table.RowBorder = &v
	})
	// header rule + one divider between the two body rows.
	if got := separatorLines(out); got != 2 {
		t.Errorf("RowBorder table want 2 separator lines, got %d:\n%s", got, out)
	}
}

func ExampleStyleTable_rowBorder() {
	style := styles.ASCIIStyleConfig
	rowBorder := true
	style.Table.RowBorder = &rowBorder

	renderer, err := NewTermRenderer(WithStyles(style), WithWordWrap(80))
	if err != nil {
		return
	}
	result, err := renderer.Render(strings.TrimSpace(tableBorderMarkdown))
	if err != nil {
		return
	}
	fmt.Println(strings.ReplaceAll(result, " ", "."))

	// Output:
	// ..............................................................................
	// ...Header.A.............................|.Header.B............................
	// ..--------------------------------------|-------------------------------------
	// ...Cell.1...............................|.Cell.2..............................
	// ..--------------------------------------|-------------------------------------
	// ...Cell.3...............................|.Cell.4..............................
}
