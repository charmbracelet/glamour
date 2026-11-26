package ansi

import (
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TemplateFuncMap contains a few useful template helpers.
var (
	TemplateFuncMap = template.FuncMap{
		"Left": func(values ...interface{}) string {
			s := values[0].(string)
			return s[:clamp(values[1].(int), len(s))]
		},
		"Matches": func(values ...interface{}) bool {
			ok, _ := regexp.MatchString(values[1].(string), values[0].(string))
			return ok
		},
		"Mid": func(values ...interface{}) string {
			s := values[0].(string)
			l := clamp(values[1].(int), len(s))

			if len(values) > 2 { //nolint:mnd
				r := clamp(values[2].(int), len(s))
				if r < l {
					return ""
				}
				return s[l:r]
			}
			return s[l:]
		},
		"Right": func(values ...interface{}) string {
			s := values[0].(string)
			return s[len(s)-clamp(values[1].(int), len(s)):]
		},
		"Last": func(values ...interface{}) string {
			list := values[0].([]string)
			if len(list) == 0 {
				return ""
			}
			return list[len(list)-1]
		},
		// strings functions
		"Compare":      strings.Compare, // 1.5+ only
		"Contains":     strings.Contains,
		"ContainsAny":  strings.ContainsAny,
		"Count":        strings.Count,
		"EqualFold":    strings.EqualFold,
		"HasPrefix":    strings.HasPrefix,
		"HasSuffix":    strings.HasSuffix,
		"Index":        strings.Index,
		"IndexAny":     strings.IndexAny,
		"Join":         strings.Join,
		"LastIndex":    strings.LastIndex,
		"LastIndexAny": strings.LastIndexAny,
		"Repeat":       strings.Repeat,
		"Replace":      strings.Replace,
		"Split":        strings.Split,
		"SplitAfter":   strings.SplitAfter,
		"SplitAfterN":  strings.SplitAfterN,
		"SplitN":       strings.SplitN,
		"Title":        cases.Title(language.English).String,
		"ToLower":      cases.Lower(language.English).String,
		"ToTitle":      cases.Upper(language.English).String,
		"ToUpper":      strings.ToUpper,
		"Trim":         strings.Trim,
		"TrimLeft":     strings.TrimLeft,
		"TrimPrefix":   strings.TrimPrefix,
		"TrimRight":    strings.TrimRight,
		"TrimSpace":    strings.TrimSpace,
		"TrimSuffix":   strings.TrimSuffix,
	}
)

// clamp keeps an index i within 0 and n, so slicing a string of length n
// with it can never panic.
func clamp(i, n int) int {
	return max(0, min(i, n))
}
