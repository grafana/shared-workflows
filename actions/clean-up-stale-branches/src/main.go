package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/grafana/shared-workflows/actions/cleanup-stale-branches/config"
	"github.com/grafana/shared-workflows/actions/cleanup-stale-branches/gh"
	"github.com/grafana/shared-workflows/actions/cleanup-stale-branches/gh_action"

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

		ctx = context.Background()
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

			if action == "fetch" && csvFile == "" {
				return fmt.Errorf("If fetch action is specified then output CSV must be specified\n")
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
				githubClient, err = gh.NewGitHubClientWithAppAuth(ctx, owner, githubAppID, githubAppPrivateKey)
				if err != nil {
					return fmt.Errorf("")
				}
			default:
				return fmt.Errorf("The GitHub authentication configuration is missing. Either %s or %s and %s environment variables must be provided.", envGitHubToken, envGitHubAppID, envGitHubAppPrivateKey)
			}

			// Set up the configuration and run the action
			cfg := &config.Config{
				Repository:    repo,
				Owner:         owner,
				DefaultBranch: defaultBranch,
				Fetch:         action == "fetch",
				Delete:        action == "delete",
				CsvFile:       csvFile,
			}
			a := gh_action.Action{
				Cfg:    *cfg,
				Client: githubClient,
			}
			a.Run(ctx)
			return nil
		},
	}

	rootCmd.Flags().StringVar(&action, "action", "", "Whether to delete or fetch")
	rootCmd.Flags().StringVar(&csvFile, "csvFile", "", "Path to the output csv file for all the stale branches to be removed")
	rootCmd.Flags().StringVar(&repository, "repository", "", "Repository to run this action, should be in the format owner/repo-name")

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
