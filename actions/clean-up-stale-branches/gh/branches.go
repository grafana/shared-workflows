package gh

import (
	"context"
	"time"

	"github.com/google/go-github/v74/github"
)

func (c *Client) GetStaleBranches(ctx context.Context, owner string, repository string, defaultBranch string) ([]*github.Branch, error) {
	// TODO: what are the opts that are necessary
	// TODO: this needs to be paginated to avoid the rate limiting
	branches, _, err := c.restClient.Repositories.ListBranches(ctx, owner, repository, nil)
	if err != nil {
		return nil, err
	}

	currentDate := time.Now()
	threeMonthsAgo := currentDate.AddDate(0, -3, 0)
	staleBranches := []*github.Branch{}

	// calculate three months before today
	for _, branch := range branches {
		if branch.GetName() == defaultBranch {
			continue
		}

		// TODO: also ignore the main branch as well
		opts := &github.PullRequestListOptions{
			Head:  owner + ":" + branch.GetName(),
			State: "open",
		}
		// TODO: what happens if the pull request doesn't exist
		pullRequests, _, err := c.restClient.PullRequests.List(ctx, owner, repository, opts)
		if err != nil {
			return nil, err
		}

		if len(pullRequests) != 0 {
			// there are open pull requests so n
			continue
		}

		// check the timing of the branch
		if branch.Commit.Commit.Committer.Date.Time.Before(threeMonthsAgo) {
			// this is a stale branch and should be removed
			staleBranches = append(staleBranches, branch)
		}
	}

	// TODO: this is where the branches would be deleted
	return staleBranches, nil
}
