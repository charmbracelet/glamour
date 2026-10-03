package autolink_test

import (
	"testing"

	"charm.land/glamour/v2/internal/autolink"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/owner/repo/issue/123", "owner/repo#123"},
		{"https://github.com/owner/repo/issues/123", "owner/repo#123"},
		{"https://github.com/owner/repo/pull/123", "owner/repo#123"},
		{"https://github.com/owner/repo/pulls/123", "owner/repo#123"},
		{"https://github.com/owner/repo/discussions/123", "owner/repo#123"},

		{"https://github.com/owner/repo/issue/123#issuecomment-456", "owner/repo#123 (comment)"},
		{"https://github.com/owner/repo/issues/123#issuecomment-456", "owner/repo#123 (comment)"},
		{"https://github.com/owner/repo/pull/123#issuecomment-456", "owner/repo#123 (comment)"},
		{"https://github.com/owner/repo/pulls/123#issuecomment-456", "owner/repo#123 (comment)"},

		{"https://github.com/owner/repo/pull/123#discussion_r456", "owner/repo#123 (comment)"},
		{"https://github.com/owner/repo/pulls/123#discussion_r456", "owner/repo#123 (comment)"},

		{"https://github.com/owner/repo/pull/123#pullrequestreview-456", "owner/repo#123 (review)"},
		{"https://github.com/owner/repo/pulls/123#pullrequestreview-456", "owner/repo#123 (review)"},

		{"https://github.com/owner/repo/discussions/123#discussioncomment-456", "owner/repo#123 (comment)"},

		{"https://github.com/owner/repo/commit/abcdefghijklmnopqrsxyz", "owner/repo@abcdefg"},

		{"https://github.com/owner/repo/pull/123/commits/abcdefghijklmnopqrsxyz", "owner/repo@abcdefg"},
		{"https://github.com/owner/repo/pulls/123/commits/abcdefghijklmnopqrsxyz", "owner/repo@abcdefg"},

		{"https://github.com/owner/repo/commit/abcdefghijklmnopqrsxyz#diff-123", "owner/repo@abcdefg"},
		{"https://github.com/owner/repo/pull/123/commits/abcdefghijklmnopqrsxyz#diff-123", "owner/repo@abcdefg"},
		{"https://github.com/owner/repo/pulls/123/commits/abcdefghijklmnopqrsxyz#diff-123", "owner/repo@abcdefg"},

		// Repository names may contain dots (e.g. vercel/next.js).
		{"https://github.com/vercel/next.js/issues/123", "vercel/next.js#123"},
		{"https://github.com/owner/repo.github.io/pull/5", "owner/repo.github.io#5"},
	}
	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			result, ok := autolink.Detect(test.url)
			if !ok {
				t.Errorf("expected to detect URL, got nil")
			}
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestDetectRejectsInvalid(t *testing.T) {
	// Characters that are not valid in GitHub owner or repository names must
	// not be treated as part of the slug. These previously matched because the
	// pattern used the `A-z` range, which also covers punctuation such as
	// `[ \ ] ^ ` `.
	urls := []string{
		"https://github.com/ow^ner/repo/issues/1",
		"https://github.com/owner/re]po/issues/1",
		"https://github.com/owner/repo/commit/abcde`fg",
	}
	for _, u := range urls {
		t.Run(u, func(t *testing.T) {
			if result, ok := autolink.Detect(u); ok {
				t.Errorf("expected %q to be rejected, got %q", u, result)
			}
		})
	}
}
