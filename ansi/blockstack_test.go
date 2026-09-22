package ansi

import "testing"

func TestBlockStackWidthUsesIndentTokenWidth(t *testing.T) {
	t.Parallel()

	one := uint(1)
	two := uint(2)
	token := "│ "
	stack := BlockStack{
		{
			Style: StyleBlock{
				Indent:      &one,
				IndentToken: &token,
			},
		},
		{
			Style: StyleBlock{
				Indent: &two,
			},
		},
	}

	if got := stack.Width(NewRenderContext(Options{WordWrap: 80})); got != 76 {
		t.Fatalf("width = %d, want 76", got)
	}
}
