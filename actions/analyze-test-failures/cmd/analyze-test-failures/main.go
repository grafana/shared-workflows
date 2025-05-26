package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Config struct {
	LokiURL          string
	LokiUsername     string
	LokiPassword     string
	Repository       string
	TimeRange        string
	GitHubToken      string
	WorkingDirectory string
	DryRun           bool
	MaxFailures      int
}

type LokiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type CommitInfo struct {
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
	Title     string    `json:"title"`
}

type FlakyTest struct {
	TestName         string         `json:"test_name"`
	FilePath         string         `json:"file_path"`
	TotalFailures    int            `json:"total_failures"`
	BranchCounts     map[string]int `json:"branch_counts"`
	ExampleWorkflows []string       `json:"example_workflows"`
	RecentCommits    []CommitInfo   `json:"recent_commits"`
}

type RawLogEntry struct {
	TestName       string `json:"test_name"`
	Branch         string `json:"branch"`
	WorkflowRunURL string `json:"workflow_run_url"`
}

type AnalysisResult struct {
	TestCount       int         `json:"test_count"`
	AnalysisSummary string      `json:"analysis_summary"`
	ReportPath      string      `json:"report_path"`
	FlakyTests      []FlakyTest `json:"flaky_tests"`
}

func main() {
	config := getConfigFromEnv()

	if err := run(config); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func getConfigFromEnv() Config {
	return Config{
		LokiURL:          os.Getenv("LOKI_URL"),
		LokiUsername:     os.Getenv("LOKI_USERNAME"),
		LokiPassword:     os.Getenv("LOKI_PASSWORD"),
		Repository:       os.Getenv("REPOSITORY"),
		TimeRange:        getEnvWithDefault("TIME_RANGE", "24h"),
		GitHubToken:      os.Getenv("GITHUB_TOKEN"),
		WorkingDirectory: getEnvWithDefault("WORKING_DIRECTORY", "."),
		DryRun:           getBoolEnvWithDefault("DRY_RUN", true),
		MaxFailures:      getIntEnvWithDefault("MAX_FAILURES", 3),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func getIntEnvWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func run(config Config) error {
	log.Printf("🔍 Starting test failure analysis for repository: %s", config.Repository)
	log.Printf("📅 Time range: %s", config.TimeRange)
	log.Printf("🔗 Loki URL: %s", config.LokiURL)
	log.Printf("📊 Max failures to process: %d", config.MaxFailures)

	// Fetch logs from Loki
	log.Printf("📡 Fetching logs from Loki...")
	logs, err := fetchLogsFromLoki(config)
	if err != nil {
		return fmt.Errorf("failed to fetch logs from Loki: %w", err)
	}

	// Parse and analyze test failures
	log.Printf("📊 Parsing test failures from log data...")
	flakyTests, err := parseTestFailures(logs, config.WorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to parse test failures: %w", err)
	}
	// Limit to configured max tests for performance
	if len(flakyTests) > config.MaxFailures {
		flakyTests = flakyTests[:config.MaxFailures]
	}

	log.Printf("🧪 Found %d flaky tests that meet criteria", len(flakyTests))
	log.Printf("📁 Finding test files in repository...")
	err = findFilePaths(config.WorkingDirectory, flakyTests)
	if err != nil {
		return fmt.Errorf("failed to find file paths for flaky tests: %w", err)
	}

	// Find authors of flaky tests
	log.Printf("👥 Finding authors of flaky tests...")
	err = findTestAuthors(config.WorkingDirectory, config.GitHubToken, flakyTests)
	if err != nil {
		return fmt.Errorf("failed to find test authors: %w", err)
	}

	// Log authors for each test
	for _, test := range flakyTests {
		if len(test.RecentCommits) > 0 {
			var authors []string
			for _, commit := range test.RecentCommits {
				if commit.Author != "" && commit.Author != "unknown" {
					authors = append(authors, commit.Author)
				}
			}
			if len(authors) > 0 {
				log.Printf("👤 %s: %s", test.TestName, strings.Join(authors, ", "))
			} else {
				log.Printf("👤 %s: no authors found", test.TestName)
			}
		} else {
			log.Printf("👤 %s: no commits found", test.TestName)
		}
	}

	// Create GitHub issues for flaky tests
	if config.DryRun {
		log.Printf("🔍 Dry run mode: Generating issue previews...")
		err = previewIssuesForFlakyTests(flakyTests)
		if err != nil {
			return fmt.Errorf("failed to preview GitHub issues: %w", err)
		}
	} else {
		log.Printf("📝 Creating GitHub issues for flaky tests...")
		err = createIssuesForFlakyTests(config.Repository, config.GitHubToken, flakyTests)
		if err != nil {
			return fmt.Errorf("failed to create GitHub issues: %w", err)
		}
	}

	// Generate analysis result
	result := AnalysisResult{
		TestCount:       len(flakyTests),
		AnalysisSummary: generateSummary(flakyTests),
		FlakyTests:      flakyTests,
	}

	// Generate report
	log.Printf("📄 Generating analysis report...")
	reportPath, err := generateReport(result)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	result.ReportPath = reportPath
	log.Printf("💾 Report saved to: %s", reportPath)

	// Set GitHub Actions outputs
	setGitHubOutput("test-count", fmt.Sprintf("%d", result.TestCount))
	setGitHubOutput("analysis-summary", result.AnalysisSummary)
	setGitHubOutput("report-path", result.ReportPath)

	log.Printf("✅ Analysis complete! Summary: %s", result.AnalysisSummary)
	return nil
}

func findFilePaths(workingDir string, flakyTests []FlakyTest) error {
	for i, test := range flakyTests {
		filePath, err := findTestFilePath(workingDir, test.TestName)
		if err != nil {
			return fmt.Errorf("failed to find file path for test %s: %w", test.TestName, err)
		}
		flakyTests[i].FilePath = filePath
	}
	return nil
}

func fetchLogsFromLoki(config Config) (string, error) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/loki/api/v1/query_range", config.LokiURL), nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	// Parse time range and calculate start/end times
	endTime := time.Now()
	startTime, err := parseTimeRange(config.TimeRange, endTime)
	if err != nil {
		return "", fmt.Errorf("parsing time range: %w", err)
	}

	// Build the LogQL query with repository parameter
	query := buildLogQLQuery(config.Repository, config.TimeRange)

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("start", fmt.Sprintf("%d", startTime.UnixNano()))
	q.Add("end", fmt.Sprintf("%d", endTime.UnixNano()))
	q.Add("limit", "5000")
	req.URL.RawQuery = q.Encode()

	// Add authentication if username and password are provided
	if config.LokiUsername != "" && config.LokiPassword != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(config.LokiUsername + ":" + config.LokiPassword))
		req.Header.Add("Authorization", "Basic "+auth)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bad response from Loki (status %d): %s",
			resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	// Parse response to count log entries
	var tempResp LokiResponse
	if err := json.Unmarshal(body, &tempResp); err == nil {
		totalEntries := 0
		for _, result := range tempResp.Data.Result {
			totalEntries += len(result.Values)
		}
		log.Printf("📥 Retrieved %d log entries from %d streams", totalEntries, len(tempResp.Data.Result))
	}

	return string(body), nil
}

func parseTimeRange(timeRange string, endTime time.Time) (time.Time, error) {
	// Handle days separately since time.ParseDuration doesn't support "d"
	if strings.HasSuffix(timeRange, "d") {
		daysStr := strings.TrimSuffix(timeRange, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid day format: %s", timeRange)
		}
		duration := time.Duration(days) * 24 * time.Hour
		return endTime.Add(-duration), nil
	}

	// Use time.ParseDuration for standard durations (h, m, s, ms, us, ns)
	duration, err := time.ParseDuration(timeRange)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time range format: %s (supported: 1h, 30m, 7d, etc.)", timeRange)
	}

	return endTime.Add(-duration), nil
}

func buildLogQLQuery(repository, timeRange string) string {
	return fmt.Sprintf(`{service_name="%s", service_namespace="cicd-o11y"} |= "--- FAIL: Test" | json | __error__="" | resources_ci_github_workflow_run_conclusion!="cancelled" | line_format "{{.body}}" | regexp "--- FAIL: (?P<test_name>.*) \\(\\d" | line_format "{{.test_name}}" | regexp `+"`(?P<parent_test_name>Test[a-z0-9A-Z_]+)`", repository)
}

func parseTestFailures(logsJSON, workingDir string) ([]FlakyTest, error) {
	var lokiResp LokiResponse
	if err := json.Unmarshal([]byte(logsJSON), &lokiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response: %w", err)
	}

	// Parse individual log entries
	var rawEntries []RawLogEntry
	for _, result := range lokiResp.Data.Result {
		testName := result.Stream["parent_test_name"]
		branch := result.Stream["resources_ci_github_workflow_run_head_branch"]
		workflowRunURL := result.Stream["resources_ci_github_workflow_run_html_url"]

		if testName == "" || branch == "" {
			continue
		}
		entry := RawLogEntry{
			TestName:       testName,
			Branch:         branch,
			WorkflowRunURL: workflowRunURL,
		}
		rawEntries = append(rawEntries, entry)
	}

	log.Printf("🔄 Processed %d log lines, extracted %d valid test failure entries", len(lokiResp.Data.Result), len(rawEntries))

	// Aggregate raw entries and apply flaky test detection
	return detectFlakyTestsFromRawEntries(rawEntries, workingDir), nil
}

func detectFlakyTestsFromRawEntries(rawEntries []RawLogEntry, workingDir string) []FlakyTest {
	// Group entries by test name
	testMap := make(map[string]map[string]int)           // testName -> branch -> failureCount
	exampleWorkflows := make(map[string]map[string]bool) // testName -> Workflow URL -> seen

	for _, entry := range rawEntries {
		if entry.TestName == "" || entry.Branch == "" {
			continue
		}

		// Initialize maps if needed
		if testMap[entry.TestName] == nil {
			testMap[entry.TestName] = make(map[string]int)
			exampleWorkflows[entry.TestName] = make(map[string]bool)
		}

		// Count failures per branch
		testMap[entry.TestName][entry.Branch]++

		// Collect example Workflow URLs (up to 3 per test)
		if entry.WorkflowRunURL != "" && len(exampleWorkflows[entry.TestName]) < 3 {
			exampleWorkflows[entry.TestName][entry.WorkflowRunURL] = true
		}
	}

	var flakyTests []FlakyTest

	for testName, branches := range testMap {
		isFlaky := false
		totalFailures := 0

		// Count total failures and check flaky criteria
		for branch, count := range branches {
			totalFailures += count

			// Flaky if failed on main branch
			if branch == "main" {
				isFlaky = true
			}
		}

		// Flaky if failed on more than one branch
		if len(branches) > 1 {
			isFlaky = true
		}

		// Only include flaky tests
		if !isFlaky {
			continue
		}

		// Create branch summary
		var branchSummary []string
		for branch, count := range branches {
			branchSummary = append(branchSummary, fmt.Sprintf("%s:%d", branch, count))
		}

		// Convert Workflow map to slice
		var workflowURLs []string
		for workflowURL := range exampleWorkflows[testName] {
			workflowURLs = append(workflowURLs, workflowURL)
		}

		flakyTests = append(flakyTests, FlakyTest{
			TestName:         testName,
			TotalFailures:    totalFailures,
			BranchCounts:     branches,
			ExampleWorkflows: workflowURLs,
		})

		log.Printf("🔍 Detected flaky test: %s (%d total failures) - branches: %s",
			testName, totalFailures, strings.Join(branchSummary, ", "))
	}

	log.Printf("📈 Test analysis stats:")
	log.Printf("   - Total unique tests with failures: %d", len(testMap))
	log.Printf("   - Tests classified as flaky: %d", len(flakyTests))

	return sortFlakyTests(flakyTests)
}

// sortFlakyTests sorts flaky tests by the number of branches they failed on (descending)
func sortFlakyTests(tests []FlakyTest) []FlakyTest {
	slices.SortFunc(tests, func(a, b FlakyTest) int {
		return len(b.BranchCounts) - len(a.BranchCounts)
	})
	return tests
}

func findTestFilePath(workingDir, testName string) (string, error) {
	// Search for Go test files containing the test function
	if !strings.HasPrefix(testName, "Test") {
		return "", fmt.Errorf("invalid test name format: %s", testName)
	}

	// Use grep to recursively search for the test function in *_test.go files
	// -r: recursive, -l: list filenames only, --include: only search *_test.go files
	grepCmd := exec.Command("grep", "-rl", "--include=*_test.go", fmt.Sprintf("func %s(", testName), ".")
	grepCmd.Dir = workingDir

	// Execute the search
	result, err := grepCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to search for test function %s: %w", testName, err)
	}

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		// Found at least one match, return the first one
		if len(lines) > 1 {
			log.Printf("Warning: test function %s found in multiple files, using first match: %s", testName, lines[0])
		}

		// Clean up the path (remove leading ./)
		filePath := strings.TrimPrefix(lines[0], "./")
		return filePath, nil
	}

	return "", fmt.Errorf("test function %s not found in repository", testName)
}

func guessTestFilePath(testName string) string {
	// Fallback implementation for Go tests
	if strings.HasPrefix(testName, "Test") {
		// Convert TestUserLogin -> user_test.go
		name := strings.TrimPrefix(testName, "Test")
		if name != "" {
			// Simple camelCase to snake_case conversion
			var result strings.Builder
			for i, r := range name {
				if i > 0 && r >= 'A' && r <= 'Z' {
					result.WriteRune('_')
				}
				result.WriteRune(r)
			}
			return strings.ToLower(result.String()) + "_test.go"
		}
	}

	return "unknown_test_file"
}

func findTestAuthors(workingDir, githubToken string, flakyTests []FlakyTest) error {
	for i, test := range flakyTests {
		commits, err := getFileAuthors(workingDir, test.FilePath, test.TestName, githubToken)
		if err != nil {
			return fmt.Errorf("failed to get authors for test %s in %s: %w", test.TestName, test.FilePath, err)
		}
		flakyTests[i].RecentCommits = commits
	}
	return nil
}

func getFileAuthors(workingDir, filePath, testName, githubToken string) ([]CommitInfo, error) {
	// Use git log -L to find the last 3 commits that modified the specific test function
	cmd := exec.Command("git", "log", "-3", "-L", fmt.Sprintf(":%s:%s", testName, filePath), "--pretty=format:%H|%ct|%s", "-s")
	cmd.Dir = workingDir

	result, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: failed to get git log for test %s in %s: %v", testName, filePath, err)
		return []CommitInfo{}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		log.Printf("Warning: no git log results for test %s in %s", testName, filePath)
		return []CommitInfo{}, nil
	}

	// Get GitHub usernames for the commits (in order)
	var commits []CommitInfo
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	for _, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid git log format for test %s in %s: %s", testName, filePath, line)
		}

		hash := parts[0]
		timestampStr := parts[1]
		title := parts[2]

		if hash == "" {
			continue
		}

		// Parse timestamp
		var timestamp time.Time
		if timestampUnix, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
			timestamp = time.Unix(timestampUnix, 0)
		}

		// Skip commits older than 6 months
		if timestamp.Before(sixMonthsAgo) {
			continue
		}

		username, err := getGitHubUsernameForCommit(hash, githubToken)
		if err != nil {
			log.Printf("Warning: failed to get GitHub username for commit %s: %v", hash, err)
			username = "unknown"
		}

		if username == "tomwilkie" {
			continue // Tom is unlikely to fix our flaky tests.
		}

		// Create commit info
		commitInfo := CommitInfo{
			Hash:      hash,
			Author:    username,
			Timestamp: timestamp,
			Title:     title,
		}
		commits = append(commits, commitInfo)
	}

	return commits, nil
}

type GitHubCommitSearchResponse []GitHubCommitSearchItem

type GitHubCommitSearchItem struct {
	Author GitHubAuthor `json:"author"`
}

type GitHubAuthor struct {
	Login string `json:"login"`
}

func getGitHubUsernameForCommit(commitHash, githubToken string) (string, error) {
	if githubToken == "" {
		return "", fmt.Errorf("no GitHub token provided")
	}

	// Use gh CLI to search for the commit
	cmd := exec.Command("gh", "search", "commits", "--hash", commitHash, "--json", "author")

	// Set the GitHub token as environment variable for gh CLI
	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", githubToken))

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

	return "", fmt.Errorf("no GitHub username found for commit %s", commitHash)
}

func createIssuesForFlakyTests(repository, githubToken string, flakyTests []FlakyTest) error {
	for _, test := range flakyTests {
		err := createOrUpdateIssueForTest(repository, githubToken, test)
		if err != nil {
			log.Printf("Warning: failed to create issue for test %s: %v", test.TestName, err)
		}
	}
	return nil
}

func previewIssuesForFlakyTests(flakyTests []FlakyTest) error {
	for _, test := range flakyTests {
		err := previewIssueForTest(test)
		if err != nil {
			log.Printf("Warning: failed to preview issue for test %s: %v", test.TestName, err)
		}
	}
	return nil
}

func previewIssueForTest(test FlakyTest) error {
	issueTitle := fmt.Sprintf("Flaky test: %s", test.TestName)

	log.Printf("📄 Issue preview for %s:", test.TestName)
	log.Printf("Title: %s", issueTitle)
	log.Printf("Labels: flaky-test")

	issueBody, err := generateInitialIssueBody(test)
	if err != nil {
		return fmt.Errorf("failed to generate issue body: %w", err)
	}

	log.Printf("Initial Body:\n%s", issueBody)
	log.Printf("────────────────────────────────────────────────────────────────────────")

	commentBody, err := generateCommentBody(test)
	if err != nil {
		return fmt.Errorf("failed to generate comment body: %w", err)
	}

	log.Printf("Comment Body:\n%s", commentBody)
	log.Printf("────────────────────────────────────────────────────────────────────────")

	return nil
}

func createOrUpdateIssueForTest(repository, githubToken string, test FlakyTest) error {
	issueTitle := fmt.Sprintf("Flaky test: %s", test.TestName)

	// Search for existing issue
	existingIssueURL, err := searchForExistingIssue(repository, githubToken, issueTitle)
	if err != nil {
		log.Printf("Warning: failed to search for existing issue: %v", err)
	}

	if existingIssueURL != "" {
		log.Printf("📝 Found existing issue for %s, adding comment: %s", test.TestName, existingIssueURL)
		return addCommentToIssue(repository, githubToken, existingIssueURL, test)
	}

	// Create new issue
	log.Printf("📝 Creating new issue for flaky test: %s", test.TestName)
	issueBody, err := generateInitialIssueBody(test)
	if err != nil {
		return fmt.Errorf("failed to generate issue body: %w", err)
	}

	cmd := exec.Command("gh", "issue", "create",
		"--repo", repository,
		"--title", issueTitle,
		"--body", issueBody,
		"--label", "flaky-test")

	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", githubToken))

	result, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create issue: %w (output: %s)", err, string(result))
	}

	issueURL := strings.TrimSpace(string(result))
	log.Printf("✅ Created issue for %s: %s", test.TestName, issueURL)

	// Add initial comment with current analysis
	return addCommentToIssue(repository, githubToken, issueURL, test)
}

func searchForExistingIssue(repository, githubToken string, issueTitle string) (string, error) {
	cmd := exec.Command("gh", "issue", "list",
		"--repo", repository,
		"--search", fmt.Sprintf("\"%s\"", issueTitle),
		"--state", "all",
		"--json", "url,title,state")

	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", githubToken))

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
			// If the issue is closed, reopen it
			if issue.State == "CLOSED" {
				log.Printf("🔄 Reopening closed issue for %s: %s", issueTitle, issue.URL)
				err := reopenIssue(repository, githubToken, issue.URL)
				if err != nil {
					log.Printf("Warning: failed to reopen issue %s: %v", issue.URL, err)
				}
			}
			return issue.URL, nil
		}
	}

	return "", nil
}

func reopenIssue(repository, githubToken, issueURL string) error {
	cmd := exec.Command("gh", "issue", "reopen", issueURL, "--repo", repository)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", githubToken))

	result, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to reopen issue: %w (output: %s)", err, string(result))
	}

	log.Printf("✅ Reopened issue: %s", issueURL)
	return nil
}

func addCommentToIssue(repository, githubToken, issueURL string, test FlakyTest) error {
	commentBody, err := generateCommentBody(test)
	if err != nil {
		return fmt.Errorf("failed to generate comment body: %w", err)
	}

	cmd := exec.Command("gh", "issue", "comment", issueURL,
		"--repo", repository,
		"--body", commentBody)

	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", githubToken))

	result, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to add comment to issue: %w (output: %s)", err, string(result))
	}

	log.Printf("✅ Added comment to issue for %s", test.TestName)
	return nil
}

const initialIssueBodyTemplate = `
## ` + "`{{.TestName}}`\n\n`{{.FilePath}}`" + `

### About This Issue
This issue tracks a flaky test that has been detected failing inconsistently. Each week our analysis tool runs and detects this test as flaky, it will add a comment below with recent failure data and who might be able to help.

### 🔍 How to investigate
1. **Run it locally** to see if you can reproduce: ` + "`go test -count=10000 -run {{.TestName}} .`" + `
2. **Check the failure logs** in the comments below - they might show a pattern
3. **Look for timing issues** - race conditions, timeouts, or external dependencies
4. **Review recent changes** - check commits that modified this test or the code under test recently

### 🎯 Next steps
- **Can you fix it?** Great! That would help the whole team
- **Need help?** Ask someone who has worked on this area recently
- **Can't fix it right now?** Consider adding logs and metrics so that next time it's easier to debug.
- **Obsolete test?** Maybe it's time to remove it or with it with ` + "`t.Skip()`" + `

_This test has been identified as flaky by [analyze-test-failures](https://github.com/grafana/shared-workflows/tree/main/actions/analyze-test-failures)._
`

const commentTemplate = `## 🚨 Hey there! This test is still being flaky - {{.Timestamp}}

This test failed **{{.TotalFailures}} times** across **{{len .BranchCounts}} different branches** in the last 7 days.

{{- if .RecentCommits}}

### 🕵️ Who might know about this?
{{- range .RecentCommits}}
- @{{.Author}} - made a relevant commit on {{.Timestamp | formatDate}}: {{.Hash}} "{{.Title}}"
{{- end}}

👆 If any of you have a few minutes, could you take a look? You might have context on what could be causing the flakiness.
{{- end}}

{{- if .ExampleWorkflows}}

### 💥 Recent failures
{{- range .ExampleWorkflows}}
- [Failed run]({{.}}) - check the logs for clues
{{- end}}
{{- end}}

**💡 Check the issue description above for investigation tips and next steps!**

Thanks for helping keep our tests reliable! 🙏`

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

type CommentData struct {
	FlakyTest
	Timestamp string
}

func generateCommentBody(test FlakyTest) (string, error) {
	tmpl, err := template.New("comment").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}).Parse(commentTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse comment template: %w", err)
	}

	data := CommentData{
		FlakyTest: test,
		Timestamp: time.Now().Format("2006-01-02 15:04:05 UTC"),
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, data); err != nil {
		return "", fmt.Errorf("failed to execute comment template: %w", err)
	}

	return body.String(), nil
}

func generateSummary(flakyTests []FlakyTest) string {
	if len(flakyTests) == 0 {
		return "No flaky tests found in the specified time range."
	}

	return fmt.Sprintf("Found %d flaky tests. Most common tests: %s",
		len(flakyTests), getMostCommonFailures(flakyTests))
}

func getMostCommonFailures(flakyTests []FlakyTest) string {
	if len(flakyTests) == 0 {
		return "none"
	}

	// Take top 5 and format as "TestName (X failures)"
	limit := min(5, len(flakyTests))
	topTests := make([]string, limit)
	for i := 0; i < limit; i++ {
		var authors []string
		for _, commit := range flakyTests[i].RecentCommits {
			if commit.Author != "" && commit.Author != "unknown" {
				authors = append(authors, commit.Author)
			}
		}
		authorsStr := "unknown"
		if len(authors) > 0 {
			authorsStr = strings.Join(authors, ", ")
		}
		topTests[i] = fmt.Sprintf("%s (%d total failures; recently changed by %s)", flakyTests[i].TestName, flakyTests[i].TotalFailures, authorsStr)
	}

	return strings.Join(topTests, ", ")
}

func generateReport(result AnalysisResult) (string, error) {
	reportPath := "test-failure-analysis.json"

	reportData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(reportPath, reportData, 0644); err != nil {
		return "", fmt.Errorf("failed to write report file: %w", err)
	}

	return filepath.Abs(reportPath)
}

func setGitHubOutput(name, value string) {
	if outputFile := os.Getenv("GITHUB_OUTPUT"); outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Warning: failed to open GITHUB_OUTPUT file: %v", err)
			return
		}
		defer f.Close()

		fmt.Fprintf(f, "%s=%s\n", name, value)
	}

	// Also print to stdout for debugging
	fmt.Printf("::set-output name=%s::%s\n", name, value)
}

func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Warning: failed to marshal JSON: %v", err)
		return "[]"
	}
	return string(data)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
