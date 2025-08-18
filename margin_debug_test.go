package glamour

import (
	"testing"
)

func TestMarginStyleApplication(t *testing.T) {
	// Test if WithMargins actually sets the style correctly
	r, err := NewTermRenderer(
		WithWordWrap(80),
		WithMargins(20, 10),
	)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	// Check if margins were applied to paragraph style
	paragraphMarginLeft := r.ansiOptions.Styles.Paragraph.MarginLeft
	paragraphMarginRight := r.ansiOptions.Styles.Paragraph.MarginRight

	if paragraphMarginLeft == nil {
		t.Error("Paragraph MarginLeft was not set")
	} else if *paragraphMarginLeft != 20 {
		t.Errorf("Expected paragraph MarginLeft to be 20, got %d", *paragraphMarginLeft)
	}

	if paragraphMarginRight == nil {
		t.Error("Paragraph MarginRight was not set")
	} else if *paragraphMarginRight != 10 {
		t.Errorf("Expected paragraph MarginRight to be 10, got %d", *paragraphMarginRight)
	}

	// Check H1 style as well
	h1MarginLeft := r.ansiOptions.Styles.H1.MarginLeft
	if h1MarginLeft == nil {
		t.Error("H1 MarginLeft was not set")
	} else if *h1MarginLeft != 20 {
		t.Errorf("Expected H1 MarginLeft to be 20, got %d", *h1MarginLeft)
	}

	t.Logf("Paragraph MarginLeft: %v", paragraphMarginLeft)
	t.Logf("Paragraph MarginRight: %v", paragraphMarginRight)
	t.Logf("H1 MarginLeft: %v", h1MarginLeft)
}