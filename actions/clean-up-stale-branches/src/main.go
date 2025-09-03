package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/shared-workflows/actions/cleanup-stale-branches/config"
	"github.com/grafana/shared-workflows/stale-branches/gh"
	"github.com/spf13/cobra"
)

const (
	envGitHubToken         = "GITHUB_TOKEN"
	envGitHubAppID         = "GITHUB_APP_ID"
	envGitHubAppPrivateKey = "GITHUB_APP_PRIVATE_KEY"
)

func main() {
	var (
		repository    string
		defaultBranch string
		action        string
		csvFile       string

		logger = initLogger()
		ctx    = context.Background()
	)

	// TODO: figure out how to have the action be customized to the following specs:
	// list outputs branches to be deleted to a CSV artifact
	// delete if input CSV specified then the CSV should be used as input for branches to delete. Otherwise, the get branches will
	//  be called and used as input for delete (can have dry run option)

	// Plan:
	// For now the only option is to just fetch the stale branches (for testing)
	// run!
	rootCmd := &cobra.Command{
		Use:   "clean-up-stale-branches",
		Short: "Will iterate and either fetch the stale branches or delete them",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the repository and owner
			owner, repo, err := parseRepository(repository)
			if err != nil {
				return fmt.Errorf("Failed to parse respository: %v\n", err)
			}

			// Create the Github client
			var (
				githubToken         = os.Getenv(envGitHubToken)
				githubAppID         = os.Getenv(envGitHubAppID)
				githubAppPrivateKey = os.Getenv(envGitHubAppPrivateKey)
				githubClient        *gh.Client
			)
			switch {
			case githubToken != "":
				githubClient = gh.NewGitHubClientWithTokenAuth(ctx, githubToken)

			case githubAppID != "" && githubAppPrivateKey != "":
				githubClient = gh.NewGitHubClientWithAppAuth(ctx, owner, githubAppID, githubAppPrivateKey)
			default:
				level.Error(logger).Log("msg", fmt.Sprintf("The GitHub authentication configuration is missing. Either %s or %s and %s environment variables must be provided.", envGitHubToken, envGitHubAppID, envGitHubAppPrivateKey))
				os.Exit(1)
			}

			cfg := &config.Config{
				Repository:    repo,
				Owner:         owner,
				DefaultBranch: defaultBranch,
				Fetch:         action == "fetch",
				Delete:        action == "delete",
				csvFile:       csvFile,
			}

			a := &action.Action{
				cfg:    cfg,
				client: githubClient,
				logger: logger,
			}
			a.Run(ctx)
			return nil
		},
	}

	rootCmd.Flags().StringVar(&action, "action", "", "Whether to delete or just fetch")
	rootCmd.Flags().StringVar(&csvFile)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to convert due to error: %v\n", err)
		os.Exit(1)
	}

}

func parseRepository(repository string) (owner, repo string, err error) {
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		err = errors.New("unsupported repository format")
		return
	}

	owner = parts[0]
	repo = parts[1]
	return
}

func initLogger() log.Logger {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	return logger
}
