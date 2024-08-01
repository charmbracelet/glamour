package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
)

//go:embed artichokes.md
var f embed.FS

func main() {
	// Provide a style by name (optional)
	var style string
	if len(os.Args) < 2 {
		// check env, if unset then use default style
		style = os.Getenv("GLAMOUR_STYLE")
	} else {
		style = os.Args[1]
	}

	// Let's learn a 'lil something about artichokes...
	b, err := f.ReadFile("artichokes.md")
	if err != nil {
		log.Fatal(err)
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithColorProfile(termenv.TrueColor),
	)
	if err != nil {
		log.Fatal(err)
	}
	md, err := r.RenderBytes(b)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", md)
}
