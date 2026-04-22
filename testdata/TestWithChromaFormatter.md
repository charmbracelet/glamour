## Usage

```go
import "github.com/charmbracelet/glamour"

in := `# Hello World

This is a simple example of Markdown rendering with Glamour!
Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples) too.

Bye!
`

out, err := glamour.Render(in, "dark")
fmt.Print(out)
```
