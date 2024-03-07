package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/kelseyhightower/envconfig"
)

// sanitisedString is a string which conforms with
// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
var labelValueRegexp = regexp.MustCompile(`[^-A-Za-z0-9_.]`)

type sanitisedString string

func (s *sanitisedString) Decode(str string) error {
	*s = sanitisedString(labelValueRegexp.ReplaceAllString(str, ""))
	return nil
}

type RepoInfo struct {
	Name  string
	Owner string
}

func (r *RepoInfo) Decode(text string) error {
	split := strings.SplitN(text, "/", 2)
	if len(split) != 2 {
		return fmt.Errorf("invalid repository format")
	}
	r.Owner, r.Name = labelValueRegexp.ReplaceAllString(split[0], ""), labelValueRegexp.ReplaceAllString(split[1], "")
	return nil
}

func (r RepoInfo) ToLabels() []string {
	return []string{
		fmt.Sprintf("trigger-repo-name=%s", r.Name),
		fmt.Sprintf("trigger-repo-owner=%s", r.Owner),
	}
}

// PullRequestInfo represents collected information about the pull request this
// action was executed in.
type PullRequestInfo struct {
	Number      int
	CreatedAt   *time.Time
	FirstCommit *github.RepositoryCommit
}

func (pri *PullRequestInfo) ToLabels() []string {
	if pri == nil {
		return []string{}
	}
	result := []string{
		fmt.Sprintf("trigger-pr=%d", pri.Number),
	}
	if pri.CreatedAt != nil {
		result = append(result, fmt.Sprintf("trigger-pr-created-at=%d", pri.CreatedAt.UTC().Unix()))
	}
	if pri.FirstCommit != nil && pri.FirstCommit.Commit != nil && pri.FirstCommit.Commit.Committer != nil {
		result = append(result, fmt.Sprintf("trigger-pr-first-commit-date=%d", pri.FirstCommit.Commit.Committer.Date.UTC().Unix()))
	}
	return result
}

var errPRLookupNotSupported error = errors.New("PR lookup not supported")

func getPullRequestNumberFromHead(ctx context.Context, workdir string) (int64, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return -1, errPRLookupNotSupported
	}
	cmd := exec.CommandContext(ctx, gitPath, "log", "--pretty=%s", "--max-count=1", "HEAD")
	cmd.Dir = workdir
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return -1, err
	}
	output := strings.TrimSpace(string(raw))
	re := regexp.MustCompile(`^.*\\(#([0-9]+)\\)$`)
	match := re.FindStringSubmatch(output)
	if len(match) < 2 {
		return -1, nil
	}
	return strconv.ParseInt(match[1], 10, 64)
}

// NewPullRequestInfo tries to generate a new PullRequestInfo object based on
// information available inside the GitHub API and environment variables. If
// now PR information is available, nil is returned without an error!
func NewPullRequestInfo(ctx context.Context, gh *github.Client) (*PullRequestInfo, error) {
	var err error
	var number int64
	ref := os.Getenv("GITHUB_REF")
	re := regexp.MustCompile(`^refs/pull/([0-9]+)/merge$`)
	match := re.FindStringSubmatch(ref)
	if len(match) == 0 {
		// This is happening outside of a pull request. This means that we
		// cannot simply get the pull request number from the ref and need to
		// look somewhere else. For our purposes we can also take a look at the
		// HEAD commit message and continue from there.
		number, err = getPullRequestNumberFromHead(ctx, ".")
		if err != nil {
			if err == errPRLookupNotSupported {
				return nil, nil
			}
			return nil, err
		}
		if number == -1 {
			return nil, nil
		}
	}
	number, err = strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PR number: %w", err)
	}
	info := PullRequestInfo{
		Number: int(number),
	}

	repo := &RepoInfo{}
	if err := repo.Decode(os.Getenv("GITHUB_REPOSITORY")); err != nil {
		return nil, err
	}
	pr, _, err := gh.PullRequests.Get(ctx, repo.Owner, repo.Name, info.Number)
	if err != nil {
		return nil, err
	}
	if pr.CreatedAt != nil {
		info.CreatedAt = &pr.CreatedAt.Time
	}

	// Now let's also try to retrieve the first commit in this PR:
	opts := github.ListOptions{
		Page: 1,
	}
	var firstCommit *github.RepositoryCommit
	for {
		commits, resp, err := gh.PullRequests.ListCommits(ctx, repo.Owner, repo.Name, info.Number, &opts)
		if err != nil {
			return nil, err
		}
		if resp.NextPage <= opts.Page {
			if len(commits) > 0 {
				firstCommit = commits[len(commits)-1]
			}
			break
		}
		opts.Page = resp.NextPage
	}
	if firstCommit != nil {
		info.FirstCommit = firstCommit
	}
	return &info, nil
}

// GitHubActionsMetadata contains the metadata provided by GitHub Actions in
// environment variables, which is used to populate labels on the Argo Workflow.
type GitHubActionsMetadata struct {
	BuildNumber  sanitisedString `envconfig:"GITHUB_RUN_NUMBER"`
	Commit       sanitisedString `envconfig:"GITHUB_SHA"`
	CommitAuthor sanitisedString `envconfig:"GITHUB_ACTOR"`
	Repo         RepoInfo        `envconfig:"GITHUB_REPOSITORY"`
	BuildEvent   sanitisedString `envconfig:"GITHUB_EVENT_NAME"`
}

func NewGitHubActionsMetadata() (GitHubActionsMetadata, error) {
	var m GitHubActionsMetadata
	err := envconfig.Process("", &m)
	if err != nil {
		return m, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return m, nil
}

func (m GitHubActionsMetadata) ToLabels() []string {
	var z GitHubActionsMetadata
	if m == z {
		return []string{}
	}

	repoLabels := m.Repo.ToLabels()

	result := []string{
		fmt.Sprintf("trigger-build-number=%s", m.BuildNumber),
		fmt.Sprintf("trigger-commit=%s", m.Commit),
		fmt.Sprintf("trigger-commit-author=%s", m.CommitAuthor),
		fmt.Sprintf("trigger-event=%s", m.BuildEvent),
	}
	return append(result, repoLabels...)
}

type LabelsProvider interface {
	ToLabels() []string
}

func (a App) args(providers ...LabelsProvider) []string {
	// Force the labels when the command is `submit`
	addCILabels := a.addCILabels || a.command == "submit"

	var args []string
	if addCILabels {
		var labels []string
		for _, prov := range providers {
			if prov == nil {
				continue
			}
			labels = append(labels, prov.ToLabels()...)
		}
		args = append(args, "--labels", strings.Join(labels, ","))
	}

	if a.workflowTemplate != "" {
		args = append(
			args,
			"--from",
			fmt.Sprintf("workflowtemplate/%s", a.workflowTemplate),
		)
	}

	args = append(
		args,
		"--loglevel",
		strings.ToLower(a.levelVar.Level().String()),
		a.command,
	)

	args = append(args, a.extraArgs...)

	return args
}
