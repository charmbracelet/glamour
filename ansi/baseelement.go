package ansi

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// BaseElement renders a styled primitive element.
type BaseElement struct {
	Token  string
	Prefix string
	Suffix string
	Style  StylePrimitive
}

func formatToken(format string, token string) (string, error) {
	var b bytes.Buffer

	v := make(map[string]interface{})
	v["text"] = token

	tmpl, err := template.New(format).Funcs(TemplateFuncMap).Parse(format)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(&b, v)
	return b.String(), err
}

func renderText(w io.Writer, p termenv.Profile, rules StylePrimitive, s string) {
	if len(s) == 0 {
		return
	}

	panic(s)

	ls := lipgloss.NewStyle()

	if rules.Color != nil {
		ls = ls.Foreground(lipgloss.Color(*rules.Color))
	}
	if rules.BackgroundColor != nil {
		ls = ls.Background(lipgloss.Color(*rules.BackgroundColor))
	}
	if rules.Underline != nil && *rules.Underline {
		ls = ls.Underline(true)
	}
	if rules.Bold != nil && *rules.Bold {
		ls = ls.Bold(true)
	}
	if rules.Italic != nil && *rules.Italic {
		ls = ls.Italic(true)
	}
	if rules.CrossedOut != nil && *rules.CrossedOut {
		ls = ls.Strikethrough(true)
	}
	// if s.Overlined != nil && *s.Overlined {
	// 	s.Style = s.Style.Overline(true)
	// }
	// if s.Inverse != nil && *s.Inverse {
	// 	s.Style = s.Style.Reverse()
	// }
	// if s.Blink != nil && *s.Blink {
	// 	s.Style = s.Style.Blink()
	// }

	if rules.Upper != nil && *rules.Upper {
		s = strings.ToUpper(s)
	}
	if rules.Lower != nil && *rules.Lower {
		s = strings.ToLower(s)
	}
	if rules.Title != nil && *rules.Title {
		s = strings.Title(s)
	}

	_, _ = w.Write([]byte(ls.Render(s)))
}

func (e *BaseElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, e.Prefix)
	defer func() {
		renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, e.Suffix)
	}()

	rules := bs.With(e.Style)
	// render unstyled prefix/suffix
	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockPrefix)
	defer func() {
		renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockSuffix)
	}()

	// render styled prefix/suffix
	renderText(w, ctx.options.ColorProfile, rules, rules.Prefix)
	defer func() {
		renderText(w, ctx.options.ColorProfile, rules, rules.Suffix)
	}()

	s := e.Token
	if len(rules.Format) > 0 {
		var err error
		s, err = formatToken(rules.Format, s)
		if err != nil {
			return err
		}
	}
	renderText(w, ctx.options.ColorProfile, rules, s)
	return nil
}
