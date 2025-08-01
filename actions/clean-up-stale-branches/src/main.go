package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/shared-workflows/actions/clean-up-stale-branches/pkg/gh"
)

// Goal:
// Need to create an API client to fetch all of the branches from Github
//
//	this sounds like a paginated thing
//
// Then see if the last commit was more than three months ago
// collect and mark as stale first
//
// don't need to delete them as a first step? just as a test run to see if I can fetch branches
//
// questions:
// - what are Github API rate limit concerns? how do I avoid getting rate limited?
// - how do I deal with authentication?
func main() {
	var (
		repository    string
		defaultBranch string

		logger = initLogger()
		ctx    = context.Background()
	)

	flag.StringVar(&repository, "repository", "", "The repostiory for which the stale branches should be cleaned up from")
	flag.StringVar(&defaultBranch, "defaultBranch", "", "The branch that should not be marked as stale")

	if repository == "" {
		level.Error(logger).Log("msg", "No repository specified")
		os.Exit(1)
	}

	// Plan:
	// For now the only option is to just fetch the stale branches (for testing)
	// run!

	// Create the Github client
	var (
		gitHubToken         = os.Getenv(envGitHubToken)
		gitHubAppID         = os.Getenv(envGitHubAppID)
		gitHubAppPrivateKey = os.Getenv(envGitHubAppPrivateKey)
		githubClient        *gh.Client
	)
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
