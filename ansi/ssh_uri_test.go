package ansi

import "testing"

func TestIsSSHURISuffix(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{":gitea/tea", true},
		{":repo.git", true},
		{": something", false},
		{"gitea/tea", false},
		{":", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isSSHURISuffix(tt.text); got != tt.expected {
			t.Errorf("isSSHURISuffix(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}
