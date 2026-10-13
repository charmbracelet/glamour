package ansi

import (
	"bytes"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

func TestChromaStyleUpdatesOnThemeSwitch(t *testing.T) {
	originalLastChromaColors := lastChromaColors
	originalStyle, originalExists := styles.Registry[chromaStyleTheme]
	t.Cleanup(func() {
		mutex.Lock()
		lastChromaColors = originalLastChromaColors
		if originalExists {
			styles.Register(originalStyle)
		} else {
			delete(styles.Registry, chromaStyleTheme)
		}
		mutex.Unlock()
	})

	red := "#ff0000"
	blue := "#0000ff"

	chroma1 := &Chroma{
		Text:       StylePrimitive{Color: &red},
		Keyword:    StylePrimitive{Color: &red},
		Background: StylePrimitive{BackgroundColor: &red},
	}
	chroma2 := &Chroma{
		Text:       StylePrimitive{Color: &blue},
		Keyword:    StylePrimitive{Color: &blue},
		Background: StylePrimitive{BackgroundColor: &blue},
	}

	render := func(chroma *Chroma) {
		e := &CodeBlockElement{Code: "var x = 1", Language: "go"}
		ctx := NewRenderContext(Options{
			Styles: StyleConfig{
				CodeBlock: StyleCodeBlock{
					Chroma: chroma,
				},
			},
		})
		var buf bytes.Buffer
		if err := e.Render(&buf, ctx); err != nil {
			t.Fatal(err)
		}
	}

	render(chroma1)
	if lastChromaColors != chroma1 {
		t.Error("lastChromaColors should point to chroma1 after first render")
	}
	registered := styles.Registry[chromaStyleTheme]
	if registered == nil {
		t.Fatal("chroma style should be registered after first render")
	}
	if got := registered.Get(chroma.Text).Colour.String(); got != red {
		t.Errorf("text color after first render: got %q, want %q", got, red)
	}

	render(chroma1)
	if lastChromaColors != chroma1 {
		t.Error("lastChromaColors should still point to chroma1 after re-render with same pointer")
	}

	render(chroma2)
	if lastChromaColors != chroma2 {
		t.Error("lastChromaColors should point to chroma2 after render with new pointer")
	}
	registered = styles.Registry[chromaStyleTheme]
	if got := registered.Get(chroma.Text).Colour.String(); got != blue {
		t.Errorf("text color after theme switch: got %q, want %q", got, blue)
	}
}
