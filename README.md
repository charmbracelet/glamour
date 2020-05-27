# Glamour

[![Latest Release](https://img.shields.io/github/release/charmbracelet/glamour.svg)](https://github.com/charmbracelet/glamour/releases)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/charmbracelet/glamour?tab=doc)
[![Build Status](https://github.com/charmbracelet/glamour/workflows/build/badge.svg)](https://github.com/charmbracelet/glamour/actions)
[![Coverage Status](https://coveralls.io/repos/github/charmbracelet/glamour/badge.svg?branch=master)](https://coveralls.io/github/charmbracelet/glamour?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/charmbracelet/glamour)](http://goreportcard.com/report/charmbracelet/glamour)

Write handsome command-line tools with *glamour*!

`glamour` lets you render [markdown](https://en.wikipedia.org/wiki/Markdown)
documents & templates on [ANSI](https://en.wikipedia.org/wiki/ANSI_escape_code)
compatible terminals. You can create your own stylesheet or use one of our
glamourous default themes.


## Usage

```go
import "github.com/charmbracelet/glamour"

in := `# Hello World

This is a simple example of glamour!
Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples).

Bye!
`

out, err := glamour.Render(in, "dark")
fmt.Print(out)
```

![HelloWorld Example](https://github.com/charmbracelet/glamour/raw/master/examples/helloworld/helloworld.png)

### Custom Renderer

```go
import "github.com/charmbracelet/glamour"

r, _ := glamour.NewTermRenderer(
    // detect background color and pick either the default dark or light theme
    glamour.WithAutoStyle(),
    // wrap output at specific width
    glamour.WithWordWrap(40),
)

out, err := r.Render(in)
fmt.Print(out)
```


## Styles

You can find all available default styles in our [gallery](https://github.com/charmbracelet/glamour/tree/master/styles/gallery).
Want to create your own style? [Learn how!](https://github.com/charmbracelet/glamour/tree/master/styles)

There are a few options for using a custom style:
1. Call `glamour.Render(inputText, "desiredStyle")`
1. Set the `GLAMOUR_STYLE` environment variable to your desired default style or a file location for a style and call `glamour.RenderWithEnvironmentConfig(inputText)`
1. Set the `GLAMOUR_STYLE` environment variable and pass `glamour.WithEnvironmentConfig()` to your custom renderer


## Glamourous Projects

Check out [Glow](https://github.com/charmbracelet/glow), a markdown renderer for
the command-line, which uses `glamour`.


## License

[MIT](https://github.com/charmbracelet/glamour/raw/master/LICENSE)
