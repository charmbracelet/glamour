package main //nolint:revive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour/ansi"
	styles "github.com/charmbracelet/glamour/styles"
)

func writeStyleJSON(filename string, styleConfig *ansi.StyleConfig) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("glamour: error creating file: %w", err)
	}
	defer f.Close() //nolint: errcheck

	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	if err := e.Encode(styleConfig); err != nil {
		return fmt.Errorf("glamour: error encoding json: %w", err)
	}
	return nil
}

func run() error {
	for style, styleConfig := range styles.DefaultStyles {
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
