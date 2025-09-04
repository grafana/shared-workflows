package gh_action

import (
	"context"
	"log"

	"github.com/grafana/shared-workflows/actions/cleanup-stale-branches/config"
	"github.com/grafana/shared-workflows/actions/cleanup-stale-branches/gh"
)

type Action struct {
	Cfg    config.Config   // should contain whether this is delete or a list and the CSV
	Client gh.GithubClient // TODO: add the logger later
}

func (a *Action) Run(ctx context.Context) error {
	// if it's a fetch, then call listStaleBranches and write the results to a CSV file
	if a.Cfg.Fetch {
		staleBranches, _ := a.Client.FetchStaleBranches(ctx, a.Cfg.Owner, a.Cfg.Repository, a.Cfg.DefaultBranch)
		log.Printf("stale branch: %v", staleBranches)
	}
	return nil
	// if it's a delete, then either use the previous step or use the CSV file specified
}
