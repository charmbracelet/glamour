package autolink_test

import (
	"testing"

	"github.com/charmbracelet/glamour/internal/autolink"
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
