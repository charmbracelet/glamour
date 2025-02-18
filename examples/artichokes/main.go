package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/glamour"
)

//go:embed artichokes.md
var f embed.FS

func main() {
	var (
		output   = os.Stdout
		styleOpt glamour.TermRendererOption
	)

	// Provide a style name via the first arg, or the GLAMOUR_STYLE environment
	// variable. If neither are set, we'll detect the background color and use
	// the standard light or dark style accordingly.
	if len(os.Args) >= 2 { //nolint:mnd
		styleOpt = glamour.WithStandardStyle(os.Args[1])
	} else if v := os.Getenv("GLAMOUR_STYLE"); v != "" {
		styleOpt = glamour.WithStandardStyle(v)
	} else {
		styleOpt = glamour.WithAutoStyle()
	}

	// Let's learn a 'lil something about artichokes...
	const filename = "artichokes.md"
	b, err := f.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Detect the color profile, and note it.
	p := colorprofile.Detect(output, os.Environ())
	var buf bytes.Buffer
	buf.Write(b)
	fmt.Fprintf(&buf, "***\n\n## By the Way\n\nWe detected the following color profile: %s.", p)

	// Set up a new renderer. Note that we're auto-detecting the color profile.
	r, err := glamour.NewTermRenderer(styleOpt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create renderer: %v\n", err)
		os.Exit(1)
	}

	// Render markdown.
	md, err := r.RenderBytes(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not render markdown: %v\n", err)
		os.Exit(1)
	}

	// Write output to the terminal, downsampling colors as necessary.
	w := &colorprofile.Writer{Forward: output, Profile: p}
	_, _ = w.Write(md)
}
