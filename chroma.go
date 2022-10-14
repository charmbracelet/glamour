package glamour

import "github.com/charmbracelet/glamour/ansi"

func init() {
	for _, style := range DefaultStyles {
		ansi.ChromaRegister(style)
	}
}
