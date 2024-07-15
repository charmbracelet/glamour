package ansi

import (
	"bytes"
	"io"
	"strings"
	"text/template"

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

	out := termenv.String(s)
	if rules.Upper != nil && *rules.Upper {
		out = termenv.String(strings.ToUpper(s))
	}
	if rules.Lower != nil && *rules.Lower {
		out = termenv.String(strings.ToLower(s))
	}
	if rules.Title != nil && *rules.Title {
		out = termenv.String(strings.Title(s))
	}
	if rules.Color != nil {
		out = out.Foreground(p.Color(*rules.Color))
	}
	if rules.BackgroundColor != nil {
		out = out.Background(p.Color(*rules.BackgroundColor))
	}
	if rules.Underline != nil && *rules.Underline {
		out = out.Underline()
	}
	if rules.Bold != nil && *rules.Bold {
		out = out.Bold()
	}
	if rules.Italic != nil && *rules.Italic {
		out = out.Italic()
	}
	if rules.CrossedOut != nil && *rules.CrossedOut {
		out = out.CrossOut()
	}
	if rules.Overlined != nil && *rules.Overlined {
		out = out.Overline()
	}
	if rules.Inverse != nil && *rules.Inverse {
		out = out.Reverse()
	}
	if rules.Blink != nil && *rules.Blink {
		out = out.Blink()
	}

	_, _ = io.WriteString(w, out.String())
}

func (e *BaseElement) StyleOverrideRender(w io.Writer, ctx RenderContext, style StylePrimitive) error {
	bs := ctx.blockStack
	st1 := cascadeStyles(bs.Current().Style, StyleBlock{
		StylePrimitive: style,
	})
	st2 := cascadeStyles(
		StyleBlock{
			StylePrimitive: bs.With(e.Style),
		},
		StyleBlock{
			StylePrimitive: style,
		},
	)

	return e.doRender(w, ctx.options.ColorProfile, st1, st2)
}

func (e *BaseElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	st1 := bs.Current().Style
	st2 := StyleBlock{
		StylePrimitive: bs.With(e.Style),
	}
	return e.doRender(w, ctx.options.ColorProfile, st1, st2)
}

func (e *BaseElement) doRender(w io.Writer, p termenv.Profile, st1, st2 StyleBlock) error {
	renderText(w, p, st1.StylePrimitive, e.Prefix)
	defer func() {
		renderText(w, p, st1.StylePrimitive, e.Suffix)
	}()

	// render unstyled prefix/suffix
	renderText(w, p, st1.StylePrimitive, st2.BlockPrefix)
	defer func() {
		renderText(w, p, st1.StylePrimitive, st2.BlockSuffix)
	}()

	// render styled prefix/suffix
	renderText(w, p, st2.StylePrimitive, st2.Prefix)
	defer func() {
		renderText(w, p, st2.StylePrimitive, st2.Suffix)
	}()

	s := e.Token
	if len(st2.Format) > 0 {
		var err error
		s, err = formatToken(st2.Format, s)
		if err != nil {
			return err
		}
	}
	renderText(w, p, st2.StylePrimitive, s)
	return nil
}
