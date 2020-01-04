package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

func writeStyleJSON(filename string, styleConfig *ansi.StyleConfig) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	return e.Encode(styleConfig)
}

func run() error {
	for style, styleConfig := range glamour.DefaultStyles {
		if err := writeStyleJSON(filepath.Join("styles", style+".json"), styleConfig); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
