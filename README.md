# Glamour

Render markdown on the CLI, with _pizzazz_!

## What is it?

A Go library that lets you use JSON-based stylesheets to render Markdown files
in the terminal. Just like CSS, you can define color and style attributes on
Markdown elements. The difference is that you use ANSI color and terminal codes
instead of CSS properties and hex colors.

Available as a library and on the CLI.

## Example Output

![Glamour Dark Style](https://github.com/charmbracelet/glamour/raw/master/styles/gallery/dark.png)

Check out the [Glamour Style Gallery](https://github.com/charmbracelet/glamour/blob/master/styles/gallery/README.md)!

## Colors

Currently `glamour` uses the [Aurora ANSI colors](https://godoc.org/github.com/logrusorgru/aurora#Index).

## Development

Style definitions located in `styles/` can be embedded into the binary with
[statik](https://github.com/rakyll/statik):

```console
statik -f -src styles -include "*.json"
```

You can re-generate screenshots of all available styles running `gallery.sh`.
This requires `termshot`, `convert` and `pngcrush` installed on your system!
