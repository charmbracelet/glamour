package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

const in1 = `# Hello World

This is a simple example of Markdown rendering with Glamour!
Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples) too.

## links

**just bold text**

_just italic text_

**[bold with url within](https://www.example.com)**

_[italic with url within](https://www.example.com)_

[normal](https://www.example.com)

[url with **bold** within](https://www.example.com)

[url with _italic_ within](https://www.example.com)

[**entire url text is bold**](https://www.example.com)

[aaa _entire url_ aaaa _text is italic_](https://www.example.com)

**test@example.com**

https://google.com

_https://google.com_

This is a [link](https://charm.sh).

## tables

| h1 | h2 |
|---|---|
| a | b |

### table with markup and escapes inside

| a | b | c |
|---|---|---|
|test1|test2|test3|
|pipe \| pipe | 2 | 3 **bold** |
` + "|test1|var a = 1|test3|\n\n## escapes\n" +
	"- \\`hi\\`\n" +
	`- \\hi
- \*hi
- \_hi
- \{hi\}
- \[hi\]
- \<hi\>
- \(hi\)
- \# hi
- \+ hi
- \- hi
- \. hi
- \! hi
- \| hi
`

func main() {
	out, _ := glamour.Render(in1, "dark")
	fmt.Print(out)
}
