package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

func main() {
	eingang := `# Hello World

This is a simple example of Markdown rendering with Glamour!
Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples) too.

This is [an example](http://example.com/ "Title") inline link.

[This link](http://example.net/) has no title attribute.


This is [an example][id] reference-style link.


See my [About](/about/) page for details.


Bye!
`
	out, _ := glamour.Render(eingang, "dark")
	fmt.Printf("%s", out)
	fmt.Print(out)
}
