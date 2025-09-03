package gh

import (
	"context"
	"fmt"

	"github.com/google/go-github/v74/github"
)

func (c *Client) FetchStaleBranches(ctx context.Context, owner string, repository string, defaultBranch string) ([]*github.Branch, error) {
	// TODO: what are the opts that are necessary
	// TODO: this needs to be paginated and needs retry to deal with the backoff
	branches, _, err := c.restClient.Repositories.ListBranches(ctx, owner, repository, nil)
	if err != nil {
		return nil, err
	}

	// currentDate := time.Now()
	// threeMonthsAgo := currentDate.AddDate(0, -3, 0)
	staleBranches := []*github.Branch{}

	for _, branch := range branches {
		if branch.GetName() == defaultBranch {
			continue
		}

		opts := &github.PullRequestListOptions{
			Head:  owner + ":" + branch.GetName(),
			State: "open",
		}
		pullRequests, _, err := c.restClient.PullRequests.List(ctx, owner, repository, opts)
		if err != nil {
			return nil, err
		}

		if len(pullRequests) != 0 {
			// there are open pull requests
			continue
		}

		// check the timing of the branch
		fmt.Printf("Branch is here: %v\n", *branch.Name)
		// if branch.Commit.Commit.Committer.Date.Time.Before(threeMonthsAgo) {
		// 	// this is a stale branch and should be removed
		// 	staleBranches = append(staleBranches, branch)
		// }
	}

	// TODO: this is where the branches would be deleted
	return staleBranches, nil
}

func (c *Client) DeleteStaleBranches() error {
	return nil
}
