package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

func main() {
	in := `# Custom Renderer

Word-wrapping will occur when lines exceed the limit of 40 characters. Just so
you know: Glamour's default word-wrapping limit is set to 100 characters per
line.

Bye!
`

	r, _ := glamour.NewTermRenderer("dark", ansi.Options{
		WordWrap: int(40),
	})

	out, _ := r.Render(in)
	fmt.Printf("%s", out)
}
