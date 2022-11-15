package ansi

import (
	"fmt"
	"github.com/BourgeoisBear/rasterm"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
)

// An ImageElement is used to render images elements.
type ImageElement struct {
	Text    string
	BaseURL string
	URL     string
	Child   ElementRenderer // FIXME
}

// ImageSkipChildrenChecker should tell whether the ast walker should skip
// all children based on ctx.options.ImageDisplay
type ImageSkipChildrenChecker struct{}

func (e *ImageSkipChildrenChecker) CheckShouldSkip(ctx RenderContext) (bool, error) {
	return ctx.options.ImageDisplay, nil
}

func (e *ImageElement) Render(w io.Writer, ctx RenderContext) error {
	handleImageDisplay := func(imageAbsUrl string, w io.Writer) error {
		file, err := os.Open(imageAbsUrl)
		if err != nil {
			return err
		}

		img, _, err := image.Decode(file)
		if err != nil {
			return err
		}

		err = rasterm.Settings{}.ItermWriteImage(w, img)
		if err != nil {
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}
		return nil
	}

	if ctx.options.ImageDisplay && len(e.URL) > 0 && rasterm.IsTermItermWez() {
		url := resolveRelativeURL(e.BaseURL, e.URL)
		err := handleImageDisplay(url, w)
		if err != nil {
			fmt.Printf("Warning: failed to display image %v: %v\n", url, err)
			// fallback to text display
		} else {
			// all done
			return nil
		}
	}

	if len(e.Text) > 0 {
		el := &BaseElement{
			Token: e.Text,
			Style: ctx.options.Styles.ImageText,
		}
		err := el.Render(w, ctx)
		if err != nil {
			return err
		}
	}
	if len(e.URL) > 0 {
		el := &BaseElement{
			Token:  resolveRelativeURL(e.BaseURL, e.URL),
			Prefix: " ",
			Style:  ctx.options.Styles.Image,
		}
		err := el.Render(w, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
