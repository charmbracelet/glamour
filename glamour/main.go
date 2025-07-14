package main

// This app render markdown and downsample colors when
// necessary per the detected color profile of the terminal.

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/glamour"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
		os.Exit(1)
	}
	defer f.Close()

	pager := os.Getenv("PAGER")
	if pager == "" { pager = "less" }
	args := strings.Fields(pager)
	cmd := exec.Command(args[0], args[1:]...)
	p, err := cmd.StdinPipe()
	cmd.Stdout = os.Stdout
	// err = cmd.Run()
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	defer func() {
		p.Close()
		cmd.Wait()
	}()

	// Read the data.
	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(f); readErr != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", readErr)
		os.Exit(1)
	}

	// Create a new colorprofile writer. We'll use it to detect the color profile and downsample colors when necessary.
	os.Setenv("CLICOLOR_FORCE", "1")
	// c := colorprofile.NewWriter(os.Stdout, os.Environ())
	c := colorprofile.NewWriter(p, os.Environ())

	// While we're at it, let's jot down the detected color profile in the
	// markdown output while we're at it.
	fmt.Fprintf(&buf, "\n\nBy the way, this was rendererd as _%s._\n", c.Profile)

	// Okay, now let's render some markdown.
	g, err := glamour.NewTermRenderer(
		glamour.WithEnvironmentConfig(),
		glamour.WithChromaFormatter("terminal16"),
	)
	if err != nil {
		log.Fatal(err)
	}
	md, err := g.RenderBytes(buf.Bytes())
	_ = md
	if err != nil {
		log.Fatal(err)
	}


	// And finally, write it to stdout using the colorprofile writer. This will
	// ensure colors are downsampled if necessary.
	// fmt.Fprintf(os.Stdout, "%s\n", md)
	fmt.Fprintf(c, "%s\n", md)
	// fmt.Fprintf(p, "%s\n", md)
	// fmt.Fprintln(p, "Hello, 世界")

}

