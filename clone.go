package glamour

import (
	"encoding/json"
	"fmt"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
)

// CloneStyle returns a deep copy of a built-in style config that can be safely
// modified without affecting the original. This is useful for making small
// adjustments to existing themes.
//
// Example:
//
//	style, _ := glamour.CloneStyle("dark")
//	style.H1.Color = stringPtr("#FF0000")  // red headings
//	r, _ := glamour.NewTermRenderer(glamour.WithStyles(*style))
func CloneStyle(name string) (*ansi.StyleConfig, error) {
	src, ok := styles.DefaultStyles[name]
	if !ok {
		return nil, fmt.Errorf("glamour: style %q not found", name)
	}

	// Deep copy via JSON round-trip to handle all pointer fields
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("glamour: error cloning style: %w", err)
	}

	var dst ansi.StyleConfig
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil, fmt.Errorf("glamour: error cloning style: %w", err)
	}

	return &dst, nil
}
