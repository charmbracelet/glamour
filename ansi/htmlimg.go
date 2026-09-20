package ansi

import "regexp"

// reImgTag matches an HTML <img ...> tag (case-insensitive, single- or multi-line).
var (
	reImgTag = regexp.MustCompile(`(?is)<img(\s[^>]*)?>`)
	reSrc    = regexp.MustCompile(`(?i)\bsrc\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s/>]+))`)
	reAlt    = regexp.MustCompile(`(?i)\balt\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s/>]+))`)
)

// parseImgTag tries to extract src and alt attributes from a raw HTML string
// that contains an <img> tag.
//
// Returns (src, alt, true) when an <img> tag is detected.
// Returns ("", "", false) when the string is not an <img> tag.
func parseImgTag(raw string) (src, alt string, ok bool) {
	if !reImgTag.MatchString(raw) {
		return "", "", false
	}
	if m := reSrc.FindStringSubmatch(raw); m != nil {
		// exactly one of the three capture groups will be non-empty
		src = m[1] + m[2] + m[3]
	}
	if m := reAlt.FindStringSubmatch(raw); m != nil {
		alt = m[1] + m[2] + m[3]
	}
	return src, alt, true
}
