package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

func main() {
	markdown := `
# Test Link Formatters

Here are some test links:

- [Google](https://google.com)
- [Example with title](https://example.com "Example Website")
- <https://autolink.com>
`

	fmt.Println("=== Default Formatter ===")
	renderer1, err := glamour.NewTermRenderer()
	if err != nil {
		log.Fatal(err)
	}
	output1, err := renderer1.Render(markdown)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output1)

	fmt.Println("\n=== Text Only Links ===")
	renderer2, err := glamour.NewTermRenderer(
		glamour.WithTextOnlyLinks(),
	)
	if err != nil {
		log.Fatal(err)
	}
	output2, err := renderer2.Render(markdown)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output2)

	fmt.Println("\n=== URL Only Links ===")
	renderer3, err := glamour.NewTermRenderer(
		glamour.WithURLOnlyLinks(),
	)
	if err != nil {
		log.Fatal(err)
	}
	output3, err := renderer3.Render(markdown)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output3)

	fmt.Println("\n=== Smart Hyperlinks ===")
	renderer4, err := glamour.NewTermRenderer(
		glamour.WithSmartHyperlinks(),
	)
	if err != nil {
		log.Fatal(err)
	}
	output4, err := renderer4.Render(markdown)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output4)

	fmt.Println("\n=== Custom Formatter ===")
	customFormatter := ansi.LinkFormatterFunc(func(data ansi.LinkData, ctx ansi.RenderContext) (string, error) {
		return fmt.Sprintf("[%s](%s)", data.Text, data.URL), nil
	})
	renderer5, err := glamour.NewTermRenderer(
		glamour.WithLinkFormatter(customFormatter),
	)
	if err != nil {
		log.Fatal(err)
	}
	output5, err := renderer5.Render(markdown)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output5)

	fmt.Println("\nAll tests completed successfully!")
}
