package main

import "testing"

func TestTransformGitHubRepoToLokiFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"grafana/mimir", "grafana-mimir"},
		{"grafana/loki", "grafana-loki"},
		{"owner/repo", "owner-repo"},
		{"single-repo", "single-repo"},             // already in correct format
		{"complex/repo/name", "complex-repo-name"}, // multiple slashes
	}

	for _, test := range tests {
		result := transformGitHubRepoToLokiFormat(test.input)
		if result != test.expected {
			t.Errorf("transformGitHubRepoToLokiFormat(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}
