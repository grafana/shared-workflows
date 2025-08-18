package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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

func (r RepoInfo) ToArgs() string {
	return fmt.Sprintf(
		"trigger-repo-name=%s,trigger-repo-owner=%s",
		r.Name,
		r.Owner,
	)
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

	if author, ok := os.LookupEnv("OVERRIDE_ACTOR"); ok {
		m.CommitAuthor = sanitisedString(author)
	}

	return m, nil
}

func (m GitHubActionsMetadata) ToArgs() []string {
	var z GitHubActionsMetadata
	if m == z {
		return []string{}
	}

	return []string{
		"--labels",
		strings.Join([]string{
			fmt.Sprintf("trigger-build-number=%s", m.BuildNumber),
			fmt.Sprintf("trigger-commit=%s", m.Commit),
			fmt.Sprintf("trigger-commit-author=%s", m.CommitAuthor),
			m.Repo.ToArgs(),
			fmt.Sprintf("trigger-event=%s", m.BuildEvent),
		}, ","),
	}
}

func (a App) args(m GitHubActionsMetadata) []string {
	// Force the labels when the command is `submit`
	addCILabels := a.addCILabels || a.command == "submit"

	var args []string
	if addCILabels {
		args = m.ToArgs()
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
