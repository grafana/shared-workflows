package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

type DefaultGitHubClient struct {
	config Config
}

func NewDefaultGitHubClient(config Config) *DefaultGitHubClient {
	return &DefaultGitHubClient{config: config}
}

func (gh *DefaultGitHubClient) GetUsernameForCommit(commitHash string) (string, error) {
	cmd := exec.Command("gh", "search", "commits", "--hash", commitHash, "--json", "author")

	result, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to search commit: %w (output: %s)", err, string(result))
	}

	var response GitHubCommitSearchResponse

	if err := json.Unmarshal(result, &response); err != nil {
		return "", fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	if len(response) > 0 && response[0].Author.Login != "" {
		return response[0].Author.Login, nil
	}

	return "", fmt.Errorf("no GitHub username found for commit %s; GitHub response %s", commitHash, string(result))
}

func (gh *DefaultGitHubClient) CreateOrUpdateIssue(test FlakyTest) error {
	issueTitle := fmt.Sprintf("Flaky %s", test.TestName)

	existingIssueURL, err := gh.SearchForExistingIssue(issueTitle)
	if err != nil {
		log.Printf("Warning: failed to search for existing issue: %v", err)
	}

	if existingIssueURL != "" {
		log.Printf("📝 Found existing issue for %s, adding comment: %s", test.TestName, existingIssueURL)
		return gh.AddCommentToIssue(existingIssueURL, test)
	}

	log.Printf("📝 Creating new issue for flaky test: %s", test.TestName)
	issueBody, err := generateInitialIssueBody(test)
	if err != nil {
		return fmt.Errorf("failed to generate issue body: %w", err)
	}

	cmd := exec.Command("gh", "issue", "create",
		"--repo", gh.config.Repository,
		"--title", issueTitle,
		"--body", issueBody,
		"--label", "flaky-test")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create GitHub issue: %w, output: %s", err, string(output))
	}

	log.Printf("📝 Created GitHub issue: %s", strings.TrimSpace(string(output)))

	issueURL := strings.TrimSpace(string(output))
	return gh.AddCommentToIssue(issueURL, test)
}

func (gh *DefaultGitHubClient) SearchForExistingIssue(issueTitle string) (string, error) {
	cmd := exec.Command("gh", "issue", "list",
		"--repo", gh.config.Repository,
		"--search", fmt.Sprintf("\"%s\"", issueTitle),
		"--state", "all",
		"--json", "url,title,state")

	result, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to search issues: %w", err)
	}

	var issues []struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		State string `json:"state"`
	}

	if err := json.Unmarshal(result, &issues); err != nil {
		return "", fmt.Errorf("failed to parse issue search response: %w", err)
	}

	for _, issue := range issues {
		if issue.Title == issueTitle {
			if issue.State == "CLOSED" {
				log.Printf("🔄 Reopening closed issue for %s: %s", issueTitle, issue.URL)
				err := gh.ReopenIssue(issue.URL)
				if err != nil {
					log.Printf("Warning: failed to reopen issue %s: %v", issue.URL, err)
				}
			}
			return issue.URL, nil
		}
	}

	return "", nil
}

func (gh *DefaultGitHubClient) AddCommentToIssue(issueURL string, test FlakyTest) error {
	commentBody, err := generateCommentBody(test, gh.config)
	if err != nil {
		return fmt.Errorf("failed to generate comment body: %w", err)
	}

	cmd := exec.Command("gh", "issue", "comment", issueURL, "--body", commentBody)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add comment to GitHub issue: %w, output: %s", err, string(output))
	}

	log.Printf("📝 Added comment to GitHub issue: %s", issueURL)

	issue, err := gh.getIssueState(issueURL)
	if err != nil {
		log.Printf("Warning: failed to check issue state: %v", err)
		return nil
	}

	if issue.State == "closed" {
		log.Printf("📝 Issue is closed, reopening: %s", issueURL)
		err := gh.ReopenIssue(issueURL)
		if err != nil {
			log.Printf("Warning: failed to reopen issue: %v", err)
		}
	}

	return nil
}

func (gh *DefaultGitHubClient) ReopenIssue(issueURL string) error {
	cmd := exec.Command("gh", "issue", "reopen", issueURL, "--repo", gh.config.Repository)

	result, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to reopen issue: %w (output: %s)", err, string(result))
	}

	log.Printf("✅ Reopened issue: %s", issueURL)
	return nil
}

func (gh *DefaultGitHubClient) getIssueState(issueURL string) (*GitHubIssue, error) {
	cmd := exec.Command("gh", "issue", "view", issueURL,
		"--repo", gh.config.Repository,
		"--json", "state,url,title")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get issue state: %w, output: %s", err, string(output))
	}

	var issue GitHubIssue
	if err := json.Unmarshal(output, &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue state response: %w", err)
	}

	return &issue, nil
}

type GitHubIssue struct {
	State string `json:"state"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

type GitHubCommitSearchResponse []GitHubCommitSearchItem

type GitHubCommitSearchItem struct {
	Author GitHubAuthor `json:"author"`
}

type GitHubAuthor struct {
	Login string `json:"login"`
}

const initialIssueBodyTemplate = `
## ` + "`{{.TestName}}`\n\n`{{.FilePath}}`" + `

### About This Issue
This issue tracks a flaky test that has been detected failing inconsistently. Each week our analysis tool runs and detects this test as flaky, it will add a comment below with recent failure data and who might be able to help.

### 🔍 How to investigate
1. **Run it locally** to see if you can reproduce: ` + "`go test -count=10000 -run {{.TestName}} ./...`" + `
2. **Check the failure logs** in the comments below - they might show a pattern
3. **Look for timing issues** - race conditions, timeouts, or external dependencies
4. **Review recent changes** - check commits that modified this test or the code under test recently

### 🎯 Next steps
- **Can you fix it?** Great! That would help the whole team
- **Need help?** Ask someone who has worked on this area recently
- **Can't fix it right now?** Consider adding logs and metrics so that next time it's easier to debug.
- **Obsolete test?** Maybe it's time to remove it or skip it with ` + "`t.Skip()`" + `

_This test has been identified as flaky by [go-flaky-tests](https://github.com/grafana/shared-workflows/tree/main/actions/go-flaky-tests)._
`

const commentTemplate = `## Hey there! This test is still being flaky

This test failed **{{.TotalFailures}} times** across **{{len .BranchCounts}} different branches** in the last {{.TimeRange}}.

### Who might know about this?
{{- if .RecentCommits}}
{{- range .RecentCommits}}
- @{{.Author}} - made a relevant commit on {{.Timestamp | formatDate}}: {{.Hash}} "{{.Title}}"
{{- end}}

If any of you have a few minutes, could you take a look? You might have context on what could be causing the flakiness.
{{- else}}
_No recent changes to the test definition._
{{- end}}

{{- if .ExampleWorkflows}}

### Recent failures
{{- range .ExampleWorkflows}}
- [Failed run]({{.}}) - check previous attempts' logs for clues
{{- end}}
{{- end}}

**💡 Check the issue description above for investigation tips and next steps!**

Thanks for helping keep our tests reliable!`

func generateInitialIssueBody(test FlakyTest) (string, error) {
	tmpl, err := template.New("initialIssueBody").Parse(initialIssueBodyTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse initial issue template: %w", err)
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, test); err != nil {
		return "", fmt.Errorf("failed to execute initial issue template: %w", err)
	}

	return body.String(), nil
}

func generateCommentBody(test FlakyTest, config Config) (string, error) {
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}).Parse(commentTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse comment template: %w", err)
	}

	type TemplateData struct {
		FlakyTest
		TimeRange string
	}

	templateData := TemplateData{
		FlakyTest: test,
		TimeRange: config.TimeRange,
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, templateData); err != nil {
		return "", fmt.Errorf("failed to execute comment template: %w", err)
	}

	return body.String(), nil
}
