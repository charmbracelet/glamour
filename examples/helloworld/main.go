package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

func main() {
	in := `# Hello World

This is a simple example of Markdown rendering with Glamour!
Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples) too.

` + "```" + `go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello World!")
}
` + "```" + `

Bye!
`

	out, _ := glamour.Render(in, "dark")
	fmt.Print(out)
}
