package main

// A simple example that renders input through a pipe.
//
// Usage:
//     echo "# Hello, world!" | go run main.go
//
//     cat README.md | go run main.go
//
//     go run main.go < README.md

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/glamour"
)

const defaultWidth = 80

func main() {
	// Read from stdin.
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %s\n", err)
	}

	// Create a new renderer.
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(defaultWidth),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating renderer: %s\n", err)
	}

	// Render markdown.
	md, err := r.RenderBytes(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering markdown: %s\n", err)
	}

	// Write markdown to stdout.
	fmt.Fprintf(os.Stdout, "%s\n", md)
}
