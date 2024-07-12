package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

func main() {
	in := `# Hello World

This is a simple example of Markdown rendering with Glamour!
Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples) too.

**just bold text**

_just italic text_

**[bold with url within](https://www.example.com)**

[normal](https://www.example.com)


[url with **bold** within](https://www.example.com)

[url with _italic_ within](https://www.example.com)

[**entire url text is bold**](https://www.example.com)

[_entire url text is italic_](https://www.example.com)

_URL_

Bye!
`

	out, _ := glamour.Render(in, "dark")
	fmt.Print(out)
}
