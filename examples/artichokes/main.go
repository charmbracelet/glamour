package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
)

//go:embed artichokes.md
var f embed.FS

func main() {
	var opt glamour.TermRendererOption

	// Provide a style name via the first arg, or the GLAMOUR_STYLE environment
	// variable. If neither are set, we'll detect the background color and use
	// the standard light or dark style accordingly.
	if len(os.Args) >= 2 { //nolint:mnd
		opt = glamour.WithStandardStyle(os.Args[1])
	} else if v := os.Getenv("GLAMOUR_STYLE"); v != "" {
		opt = glamour.WithStandardStyle(v)
	} else {
		opt = glamour.WithAutoStyle()
	}

	// Let's learn a 'lil something about artichokes...
	const filename = "artichokes.md"
	b, err := f.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Set up a new renderer.
	r, err := glamour.NewTermRenderer(opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create renderer: %v\n", err)
		os.Exit(1)
	}

	// Render markdown.
	md, err := r.RenderBytes(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not render markdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "%s\n", md)
}
