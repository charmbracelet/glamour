package main

import (
	"fmt"

	_ "embed"

	"github.com/charmbracelet/glamour"
)

//go:embed input.md
var input []byte

func main() {
	out, _ := glamour.Render(string(input), "dark")
	fmt.Print(out)
}
