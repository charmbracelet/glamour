// Package autolink provides a function to detect and format GitHub links into
// a more readable manner.
package autolink

import (
	"fmt"
	"regexp"
)

type pattern struct {
	pattern *regexp.Regexp
	yield   func(m []string) string
}

const (
	// owner matches a GitHub account name (letters, digits and hyphens).
	owner = `[\w-]+`
	// repo matches a GitHub repository name, which may additionally contain
	// dots (e.g. next.js).
	repo = `[\w.-]+`
	// sha matches a Git commit hash.
	sha = `[A-Za-z0-9]{7,}`
)

var patterns = []pattern{
	{
		regexp.MustCompile(`^https?://github\.com/(` + owner + `)/(` + repo + `)/(issues?|pulls?|discussions?)/([0-9]+)$`),
		func(m []string) string { return fmt.Sprintf("%s/%s#%s", m[1], m[2], m[4]) },
	},
	{
		regexp.MustCompile(`^https?://github\.com/(` + owner + `)/(` + repo + `)/(issues?|pulls?|discussions?)/([0-9]+)#issuecomment-[0-9]+$`),
		func(m []string) string { return fmt.Sprintf("%s/%s#%s (comment)", m[1], m[2], m[4]) },
	},
	{
		regexp.MustCompile(`^https?://github\.com/(` + owner + `)/(` + repo + `)/pulls?/([0-9]+)#discussion_r[0-9]+$`),
		func(m []string) string { return fmt.Sprintf("%s/%s#%s (comment)", m[1], m[2], m[3]) },
	},
	{
		regexp.MustCompile(`^https?://github\.com/(` + owner + `)/(` + repo + `)/pulls?/([0-9]+)#pullrequestreview-[0-9]+$`),
		func(m []string) string { return fmt.Sprintf("%s/%s#%s (review)", m[1], m[2], m[3]) },
	},
	{
		regexp.MustCompile(`^https?://github\.com/(` + owner + `)/(` + repo + `)/discussions/([0-9]+)#discussioncomment-[0-9]+$`),
		func(m []string) string { return fmt.Sprintf("%s/%s#%s (comment)", m[1], m[2], m[3]) },
	},
	{
		regexp.MustCompile(`^https?://github\.com/(` + owner + `)/(` + repo + `)/commit/(` + sha + `)(#.*)?$`),
		func(m []string) string { return fmt.Sprintf("%s/%s@%s", m[1], m[2], m[3][:7]) },
	},
	{
		regexp.MustCompile(`^https?://github\.com/(` + owner + `)/(` + repo + `)/pulls?/[0-9]+/commits/(` + sha + `)(#.*)?$`),
		func(m []string) string { return fmt.Sprintf("%s/%s@%s", m[1], m[2], m[3][:7]) },
	},
}

// Detect checks if the given URL matches any of the known patterns and
// returns a human-readable formatted string if a match is found.
func Detect(u string) (string, bool) {
	for _, p := range patterns {
		if m := p.pattern.FindStringSubmatch(u); len(m) > 0 {
			return p.yield(m), true
		}
	}
	return "", false
}
