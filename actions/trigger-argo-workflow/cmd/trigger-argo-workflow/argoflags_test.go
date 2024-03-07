package main

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
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
				"trigger-build-number=1,trigger-commit=abc,trigger-commit-author=actor,trigger-event=event,trigger-repo-name=repo,trigger-repo-owner=owner,trigger-pr=123,trigger-pr-created-at=1702462272,trigger-pr-first-commit-date=1702458672",
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
			gh := createMockGitHubClient(t)
			pr, err := NewPullRequestInfo(context.Background(), slog.Default(), gh)
			require.NoError(t, err)

			output := a.args(md, pr)
			require.Equal(t, tc.expectedOutput, output)
		})
	}
}

func createMockGitHubClient(t *testing.T) *github.Client {
	t.Helper()
	httpClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{
				CreatedAt: &github.Timestamp{
					Time: time.Date(2023, 12, 13, 10, 11, 12, 0, time.UTC),
				},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposPullsCommitsByOwnerByRepoByPullNumber,
			[]*github.RepositoryCommit{
				{
					Commit: &github.Commit{
						Committer: &github.CommitAuthor{
							Date: &github.Timestamp{
								Time: time.Date(2023, 12, 13, 9, 15, 12, 0, time.UTC),
							},
						},
					},
				},
				{
					Commit: &github.Commit{
						Committer: &github.CommitAuthor{
							Date: &github.Timestamp{
								Time: time.Date(2023, 12, 13, 9, 11, 12, 0, time.UTC),
							},
						},
					},
				},
			},
		),
	)
	return github.NewClient(httpClient)
}

func TestPullRequestInfo(t *testing.T) {
	t.Run("no-pr", func(t *testing.T) {
		gh := createMockGitHubClient(t)
		t.Setenv("GITHUB_REF", "something")
		info, err := NewPullRequestInfo(context.Background(), slog.Default(), gh)
		require.NoError(t, err)
		require.Nil(t, info)
	})
	t.Run("no-pr", func(t *testing.T) {
		gh := createMockGitHubClient(t)
		t.Setenv("GITHUB_REF", "refs/pull/123/merge")
		t.Setenv("GITHUB_REPOSITORY", "grafana/shared-workflows")
		info, err := NewPullRequestInfo(context.Background(), slog.Default(), gh)
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 123, info.Number)
	})
}

func TestGetPullRequestNumberFromHead(t *testing.T) {
	gitPath, err := exec.LookPath("git")
	if err != nil || gitPath == "" {
		t.Skipf("Git is not available, skipping")
		return
	}

	gitCommand := func(t *testing.T, workDir string, args ...string) {
		home := t.TempDir()
		cmd := exec.Command(gitPath, args...)
		cmd.Env = []string{
			"GIT_CONFIG_NOSYSTEM=true",
			"GIT_CONFIG_NOGLOBAL=true",
			"HOME=" + home,
		}
		cmd.Dir = workDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Log(string(output))
			t.Fatalf("git failed: %s", err.Error())
		}
	}

	t.Run("without-head-referencing-pr", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitCommand(t, tmpDir, "init")
		gitCommand(t, tmpDir, "config", "user.email", "test@example.org")
		gitCommand(t, tmpDir, "config", "user.name", "test")
		gitCommand(t, tmpDir, "commit", "-m", "Hello world", "--allow-empty")
		num, err := getPullRequestNumberFromHead(context.Background(), tmpDir)
		require.NoError(t, err)
		require.Equal(t, int64(-1), num)
	})

	t.Run("with-head-referencing-pr", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitCommand(t, tmpDir, "init")
		gitCommand(t, tmpDir, "config", "user.email", "test@example.org")
		gitCommand(t, tmpDir, "config", "user.name", "test")
		gitCommand(t, tmpDir, "commit", "-m", "Hello world (#123)", "--allow-empty")
		num, err := getPullRequestNumberFromHead(context.Background(), tmpDir)
		require.NoError(t, err)
		require.Equal(t, int64(123), num)
	})
}
