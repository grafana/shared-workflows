package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildCommand(t *testing.T) {
	testCases := []struct {
		name           string
		command        string
		addCILabels    bool
		logLevel       string
		expectedError  bool
		expectedOutput []string
		envVars        map[string]string
	}{
		{
			name:           "Add labels to submit command",
			command:        "submit",
			addCILabels:    true,
			logLevel:       "info",
			expectedOutput: []string{"--labels", "trigger-build-number=1,trigger-commit=abc,trigger-commit-author=actor,trigger-repo-name=repo,trigger-repo-owner=owner,trigger-event=event", "--loglevel", "info", "submit"},
			envVars: map[string]string{
				"GITHUB_RUN_NUMBER": "1",
				"GITHUB_SHA":        "abc",
				"GITHUB_ACTOR":      "actor",
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_EVENT_NAME": "event",
			},
		},
		{
			name:           "No labels when addCILabels is false",
			command:        "stop",
			addCILabels:    false,
			logLevel:       "info",
			expectedOutput: []string{"--loglevel", "info", "stop"},
			envVars: map[string]string{
				"GITHUB_RUN_NUMBER": "2",
				"GITHUB_SHA":        "abc",
				"GITHUB_ACTOR":      "actor",
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_EVENT_NAME": "event",
			},
		},
		{
			name:          "Invalid repository format",
			command:       "submit",
			addCILabels:   true,
			logLevel:      "info",
			expectedError: true,
			envVars: map[string]string{
				"GITHUB_RUN_NUMBER": "3",
				"GITHUB_SHA":        "abc",
				"GITHUB_ACTOR":      "actor",
				"GITHUB_REPOSITORY": "invalid",
				"GITHUB_EVENT_NAME": "event",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			level, err := parseLogLevel(tc.logLevel)
			require.NoError(t, err)

			var lv slog.LevelVar
			lv.Set(level)

			a := App{
				levelVar: &lv,

				addCILabels: tc.addCILabels,

				command: tc.command,
			}

			md, err := NewGitHubActionsMetadata()
			if tc.expectedError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			output := a.args(md)
			require.Equal(t, tc.expectedOutput, output)
		})
	}
}
