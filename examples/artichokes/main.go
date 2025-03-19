package main

// This example illustrates how to render markdown and downsample colors when
// necessary per the detected color profile of the terminal.

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/glamour"
)

//go:embed artichokes.md
var f embed.FS

func main() {
	// Open the file to learn a thing or two about artichokes.
	f, err := f.Open("artichokes.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Read the data.
	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(f); readErr != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", readErr)
		os.Exit(1)
	}

	// Create a new colorprofile writer. We'll use it to detect the color
	// profile and downsample colors when necessary.
	w := colorprofile.NewWriter(os.Stdout, os.Environ())

	// While we're at it, let's jot down the detected color profile in the
	// markdown output while we're at it.
	fmt.Fprintf(&buf, "\n\nBy the way, this was rendererd as _%s._\n", w.Profile)

	// Okay, now let's render some markdown.
	r, err := glamour.NewTermRenderer(glamour.WithEnvironmentConfig())
	if err != nil {
		log.Fatal(err)
	}
	md, err := r.RenderBytes(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	// And finally, write it to stdout using the colorprofile writer. This will
	// ensure colors are downsampled if necessary.
	fmt.Fprintf(w, "%s\n", md)
}
