package main

import (
	"context"
	"log/slog"
	"os"
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
			name:        "Add labels to submit command",
			command:     "submit",
			addCILabels: true,
			logLevel:    "info",
			expectedOutput: []string{
				"--labels",
				"trigger-build-number=1,trigger-commit=abc,trigger-commit-author=actor,trigger-event=event,trigger-repo-name=repo,trigger-repo-owner=owner,trigger-pr=123",
				"--loglevel",
				"info",
				"submit",
			},
			envVars: map[string]string{
				"GITHUB_RUN_NUMBER": "1",
				"GITHUB_SHA":        "abc",
				"GITHUB_ACTOR":      "actor",
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_EVENT_NAME": "event",
				"GITHUB_REF":        "refs/pull/123/merge",
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
				os.Setenv(k, v)
				defer os.Unsetenv(k)
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
			pr, err := NewPullRequestInfo(context.Background())
			require.NoError(t, err)

			output := a.args(md, pr)
			require.Equal(t, tc.expectedOutput, output)
		})
	}
}

func TestPullRequestInfo(t *testing.T) {
	t.Run("no-pr", func(t *testing.T) {
		t.Setenv("GITHUB_REF", "something")
		info, err := NewPullRequestInfo(context.Background())
		require.NoError(t, err)
		require.Nil(t, info)
	})
	t.Run("no-pr", func(t *testing.T) {
		t.Setenv("GITHUB_REF", "refs/pull/123/merge")
		info, err := NewPullRequestInfo(context.Background())
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, int64(123), info.Number)
	})
}
