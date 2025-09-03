package action

import (
	"context"
	"log"

	"github.com/grafana/shared-workflows/stale-branches/config"
	"github.com/grafana/shared-workflows/stale-branches/gh"
)

type Action struct {
	cfg    config.Config // should contain whether this is delete or a list and the CSV
	client gh.GithubClient
	logger log.Logger
}

func (a *Action) Run(ctx context.Context) error {
	// if it's a fetch, then call listStaleBranches and write the results to a CSV file
	staleBranches := a.client.FetchStaleBranches()
	// if it's a delete, then either use the previous step or use the CSV file specified
}
