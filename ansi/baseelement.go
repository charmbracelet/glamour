package ansi

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss/v2"
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

func renderText(w io.Writer, rules StylePrimitive, s string) {
	if len(s) == 0 {
		return
	}

	out := lipgloss.NewStyle().SetString(s)
	if rules.Upper != nil && *rules.Upper {
		out = out.SetString(strings.ToUpper(s))
	}
	if rules.Lower != nil && *rules.Lower {
		out = out.SetString(strings.ToLower(s))
	}
	if rules.Title != nil && *rules.Title {
		out = out.SetString(strings.Title(s))
	}
	if rules.Color != nil {
		out = out.Foreground(lipgloss.Color(*rules.Color))
	}
	if rules.BackgroundColor != nil {
		out = out.Background(lipgloss.Color(*rules.BackgroundColor))
	}
	if rules.Underline != nil {
		out = out.Underline(*rules.Underline)
	}
	if rules.Bold != nil {
		out = out.Bold(*rules.Bold)
	}
	if rules.Italic != nil {
		out = out.Italic(*rules.Italic)
	}
	if rules.CrossedOut != nil {
		out = out.Strikethrough(*rules.CrossedOut)
	}
	if rules.Inverse != nil {
		out = out.Reverse(*rules.Inverse)
	}
	if rules.Blink != nil {
		out = out.Blink(*rules.Blink)
	}

	_, _ = io.WriteString(w, out.String())
}

func (e *BaseElement) StyleOverrideRender(w io.Writer, ctx RenderContext, style StylePrimitive) error {
	bs := ctx.blockStack
	st1 := cascadeStylePrimitives(bs.Current().Style.StylePrimitive, style)
	st2 := cascadeStylePrimitives(bs.With(e.Style), style)

	return e.doRender(w, st1, st2)
}

func (e *BaseElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack
	st1 := bs.Current().Style.StylePrimitive
	st2 := bs.With(e.Style)
	return e.doRender(w, st1, st2)
}

func (e *BaseElement) doRender(w io.Writer, st1, st2 StylePrimitive) error {
	renderText(w, st1, e.Prefix)
	defer func() {
		renderText(w, st1, e.Suffix)
	}()

	// render unstyled prefix/suffix
	renderText(w, st1, st2.BlockPrefix)
	defer func() {
		renderText(w, st1, st2.BlockSuffix)
	}()

	// render styled prefix/suffix
	renderText(w, st2, st2.Prefix)
	defer func() {
		renderText(w, st2, st2.Suffix)
	}()

	s := e.Token
	if len(st2.Format) > 0 {
		var err error
		s, err = formatToken(st2.Format, s)
		if err != nil {
			return err
		}
	}
	renderText(w, st2, escapeReplacer.Replace(s))
	return nil
}

// https://www.markdownguide.org/basic-syntax/#characters-you-can-escape
var escapeReplacer = strings.NewReplacer(
	"\\\\", "\\",
	"\\`", "`",
	"\\*", "*",
	"\\_", "_",
	"\\{", "{",
	"\\}", "}",
	"\\[", "[",
	"\\]", "]",
	"\\<", "<",
	"\\>", ">",
	"\\(", "(",
	"\\)", ")",
	"\\#", "#",
	"\\+", "+",
	"\\-", "-",
	"\\.", ".",
	"\\!", "!",
	"\\|", "|",
)
