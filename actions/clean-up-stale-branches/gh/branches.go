package gh

import (
	"context"
	"time"

	"github.com/google/go-github/v74/github"
)

func (c *Client) GetStaleBranches(ctx context.Context, owner string, repository string) {
	// TODO: what are the opts that are necessary
	// TODO: this needs to be paginated to avoid the rate limiting
	branches, resp, err := c.restClient.Repositories.ListBranches(ctx, owner, repository, nil)

	currentDate := time.Now()
	threeMonthsAgo := currentDate.AddDate(0, -3, 0)

	// TODO: I wonder if I could write a lambda that would then be used to filter within the branches
	staleBranches := []*github.Branch{}

	// calculate three months before today
	for _, branch := range branches {
		opts := &github.PullRequestListOptions{
			Head:  branch.GetName(),
			State: "open",
		}
		// TODO: what happens if the
		pullRequests, _, err := c.restClient.PullRequests.List(ctx, owner, repository, opts)

		// is there an open PR associated with the branch?

		if !isPR {
			// don't include in the stale branches
			continue
		}

		// check the timing of the branch
		if branch.Commit.Commit.Committer.Date.Time < threeMonthsAgo {
			// this is a stale branch and should be removed
			staleBranches.append(branch)
		}
	}

	return staleBranches
}
