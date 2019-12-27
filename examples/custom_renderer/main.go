package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

func main() {
	in := `# Custom Renderer

Word-wrapping will occur when lines exceed the limit of 40 characters. Just so
you know: Glamour's default word-wrapping limit is set to 100 characters per
line.

Bye!
`

	r, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(40),
	)

	out, _ := r.Render(in)
	fmt.Print(out)
}
