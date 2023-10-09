package ansi

import (
	"io"

	"github.com/charmbracelet/scrapbook"
)

// BaseElement renders a styled primitive element.
// TODO do I still need these elements
type BaseElement struct {
	Token  string
	Prefix string
	Suffix string
	Style  scrapbook.Styler
}

// func formatToken(format string, token string) (string, error) {
// 	var b bytes.Buffer
//
// 	v := make(map[string]interface{})
// 	v["text"] = token
//
// 	tmpl, err := template.New(format).Funcs(TemplateFuncMap).Parse(format)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	err = tmpl.Execute(&b, v)
// 	return b.String(), err
// }

func renderText(w io.Writer, styler scrapbook.Styler, s string) {
	if len(s) == 0 {
		return
	}
	_, _ = w.Write([]byte(styler.Style().Render(s)))
}

func (e *BaseElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	parentBlock := bs.Current()

	renderText(w, parentBlock.Style, e.Prefix)
	defer func() {
		renderText(w, parentBlock.Style, e.Suffix)
	}()

	// TODO we're using the parent's style which should contain the rendering specs for the children...
	rules := parentBlock.Style
	// render unstyled prefix/suffix
	// TODO handle prefix/suffix

	s := e.Token
	//if len(rules.Format) > 0 {
	//	var err error
	//	s, err = formatToken(rules.Format, s)
	//	if err != nil {
	//		return err
	//	}
	//}
	renderText(w, rules, s)
	return nil
}
