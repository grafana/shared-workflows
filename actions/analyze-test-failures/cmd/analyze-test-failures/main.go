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

type TestFailure struct {
	TestName         string         `json:"test_name"`
	FilePath         string         `json:"file_path"`
	TotalFailures    int            `json:"total_failures"`
	BranchCounts     map[string]int `json:"branch_counts"`
	ExampleWorkflows []string       `json:"example_workflows"`
	RecentAuthors    []string       `json:"recent_authors"`
}

type RawLogEntry struct {
	TestName       string `json:"test_name"`
	Branch         string `json:"branch"`
	WorkflowRunURL string `json:"workflow_run_url"`
}

type AnalysisResult struct {
	FailureCount    int           `json:"failure_count"`
	AnalysisSummary string        `json:"analysis_summary"`
	ReportPath      string        `json:"report_path"`
	TestFailures    []TestFailure `json:"test_failures"`
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
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func run(config Config) error {
	log.Printf("🔍 Starting test failure analysis for repository: %s", config.Repository)
	log.Printf("📅 Time range: %s", config.TimeRange)
	log.Printf("🔗 Loki URL: %s", config.LokiURL)

	// Fetch logs from Loki
	log.Printf("📡 Fetching logs from Loki...")
	logs, err := fetchLogsFromLoki(config)
	if err != nil {
		return fmt.Errorf("failed to fetch logs from Loki: %w", err)
	}

	// Parse and analyze test failures
	log.Printf("📊 Parsing test failures from log data...")
	failures, err := parseTestFailures(logs, config.WorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to parse test failures: %w", err)
	}
	failures = failures[:4] // Limit to top 4 failures for performance

	log.Printf("🧪 Found %d flaky tests that meet criteria", len(failures))
	log.Printf("📁 Finding test files in repository...")
	err = findFilePaths(config.WorkingDirectory, failures)
	if err != nil {
		return fmt.Errorf("failed to find file paths for test failures: %w", err)
	}

	// Find authors of failing tests
	log.Printf("👥 Finding authors of failing tests...")
	_, err = findTestAuthors(config.WorkingDirectory, config.GitHubToken, failures)
	if err != nil {
		return fmt.Errorf("failed to find test authors: %w", err)
	}

	// Log authors for each test
	for _, failure := range failures {
		if len(failure.RecentAuthors) > 0 {
			log.Printf("👤 %s: %s", failure.TestName, strings.Join(failure.RecentAuthors, ", "))
		} else {
			log.Printf("👤 %s: no authors found", failure.TestName)
		}
	}

	// Generate analysis result
	result := AnalysisResult{
		FailureCount:    len(failures),
		AnalysisSummary: generateSummary(failures),
		TestFailures:    failures,
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
	setGitHubOutput("failure-count", fmt.Sprintf("%d", result.FailureCount))
	setGitHubOutput("analysis-summary", result.AnalysisSummary)
	setGitHubOutput("report-path", result.ReportPath)

	log.Printf("✅ Analysis complete! Summary: %s", result.AnalysisSummary)
	return nil
}

func findFilePaths(workingDir string, failures []TestFailure) error {
	for i, failure := range failures {
		filePath, err := findTestFilePath(workingDir, failure.TestName)
		if err != nil {
			return fmt.Errorf("failed to find file path for test %s: %w", failure.TestName, err)
		}
		failures[i].FilePath = filePath
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

func parseTestFailures(logsJSON, workingDir string) ([]TestFailure, error) {
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

func detectFlakyTestsFromRawEntries(rawEntries []RawLogEntry, workingDir string) []TestFailure {
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

	var flakyTests []TestFailure

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

		flakyTests = append(flakyTests, TestFailure{
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
func sortFlakyTests(tests []TestFailure) []TestFailure {
	slices.SortFunc(tests, func(a, b TestFailure) int {
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

func findTestAuthors(workingDir, githubToken string, failures []TestFailure) ([]string, error) {
	authorsMap := make(map[string]bool)

	for i, failure := range failures {
		authors, err := getFileAuthors(workingDir, failure.FilePath, failure.TestName, githubToken)
		if err != nil {
			log.Printf("Warning: failed to get authors for test %s in %s: %v", failure.TestName, failure.FilePath, err)
			continue
		}

		// Store the recent authors for this test (already limited to 3 by git log -3)
		failures[i].RecentAuthors = authors

		for _, author := range authors {
			authorsMap[author] = true
		}
	}

	var uniqueAuthors []string
	for author := range authorsMap {
		uniqueAuthors = append(uniqueAuthors, author)
	}

	return uniqueAuthors, nil
}

func getFileAuthors(workingDir, filePath, testName, githubToken string) ([]string, error) {
	// Use git log -L to find the last 3 commits that modified the specific test function
	cmd := exec.Command("git", "log", "-3", "-L", fmt.Sprintf(":%s:%s", testName, filePath), "--pretty=format:%H", "-s")
	cmd.Dir = workingDir

	result, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: failed to get git log for test %s in %s: %v", testName, filePath, err)
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		log.Printf("Warning: no git log results for test %s in %s", testName, filePath)
		return []string{}, nil
	}

	// Get GitHub usernames for the commits (in order)
	var authors []string
	seenAuthors := make(map[string]bool)

	for _, line := range lines {
		hash := strings.TrimSpace(line)
		if hash == "" {
			continue
		}

		username, err := getGitHubUsernameForCommit(hash, githubToken)
		if err != nil {
			log.Printf("Warning: failed to get GitHub username for commit %s: %v", hash, err)
			continue
		}

		// Only add unique authors to preserve order
		if username != "" && !strings.HasSuffix(username, "[bot]") && !seenAuthors[username] {
			authors = append(authors, username)
			seenAuthors[username] = true
		}
	}

	return authors, nil
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

func generateSummary(failures []TestFailure) string {
	if len(failures) == 0 {
		return "No test failures found in the specified time range."
	}

	return fmt.Sprintf("Found %d test failures. Most common failures: %s",
		len(failures), getMostCommonFailures(failures))
}

func getMostCommonFailures(failures []TestFailure) string {
	if len(failures) == 0 {
		return "none"
	}

	// Take top 5 and format as "TestName (X failures)"
	limit := min(5, len(failures))
	topFailures := make([]string, limit)
	for i := 0; i < limit; i++ {
		topFailures[i] = fmt.Sprintf("%s (%d total failures; recently changed by %s)", failures[i].TestName, failures[i].TotalFailures, strings.Join(failures[i].RecentAuthors, ", "))
	}

	return strings.Join(topFailures, ", ")
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
