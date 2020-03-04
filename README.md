# Glamour

[![Latest Release](https://img.shields.io/github/release/charmbracelet/glamour.svg)](https://github.com/charmbracelet/glamour/releases)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/charmbracelet/glamour?tab=doc)
[![Build Status](https://github.com/charmbracelet/glamour/workflows/build/badge.svg)](https://github.com/charmbracelet/glamour/actions)
[![Coverage Status](https://coveralls.io/repos/github/charmbracelet/glamour/badge.svg?branch=master)](https://coveralls.io/github/charmbracelet/glamour?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/charmbracelet/glamour)](http://goreportcard.com/report/charmbracelet/glamour)

Write handsome command-line tools with *glamour*!

`glamour` lets you use [markdown](https://en.wikipedia.org/wiki/Markdown)
templates to render user-friendly & stylish output on [ANSI](https://en.wikipedia.org/wiki/ANSI_escape_code)
compatible terminals.


## Usage

```go
import "github.com/charmbracelet/glamour"

in := `# Hello World

This is a simple example of glamour!
Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples).

Bye!
`

out, _ := glamour.Render(in, "dark")
fmt.Print(out)
```

![HelloWorld Example](https://github.com/charmbracelet/glamour/raw/master/examples/helloworld/helloworld.png)

### Custom Renderer

```go
import "github.com/charmbracelet/glamour"

r, _ := glamour.NewTermRenderer(
	glamour.WithStandardStyle("dark"),
	glamour.WithWordWrap(40),
)

out, _ := r.Render(in)
fmt.Print(out)
```


## Glamourous Projects

Check out [Glow](https://github.com/charmbracelet/glow), a markdown renderer for
the command-line, which uses `glamour`.


## License

[MIT](https://github.com/charmbracelet/glamour/raw/master/LICENSE)
