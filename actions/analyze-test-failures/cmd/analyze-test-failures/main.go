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
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
			Value  []interface{}     `json:"value,omitempty"`
		} `json:"result"`
	} `json:"data"`
}

type TestFailure struct {
	TestName     string         `json:"test_name"`
	FilePath     string         `json:"file_path"`
	Timestamp    string         `json:"timestamp"`
	Message      string         `json:"message"`
	BranchCounts map[string]int `json:"branch_counts"`
	IsFlaky      bool           `json:"is_flaky"`
}

type TestFailureByBranch struct {
	TestName     string `json:"test_name"`
	Branch       string `json:"branch"`
	FailureCount int    `json:"failure_count"`
	Timestamp    string `json:"timestamp"`
}

type AnalysisResult struct {
	FailureCount    int           `json:"failure_count"`
	AffectedAuthors []string      `json:"affected_authors"`
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
		TimeRange:        getEnvWithDefault("TIME_RANGE", "1h"),
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
	// Fetch logs from Loki
	logs, err := fetchLogsFromLoki(config)
	if err != nil {
		return fmt.Errorf("failed to fetch logs from Loki: %w", err)
	}

	// Parse and analyze test failures
	failures, err := parseTestFailures(logs, config.WorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to parse test failures: %w", err)
	}

	// Find authors of failing tests
	authors, err := findTestAuthors(config.WorkingDirectory, config.GitHubToken, failures)
	if err != nil {
		return fmt.Errorf("failed to find test authors: %w", err)
	}

	// Generate analysis result
	result := AnalysisResult{
		FailureCount:    len(failures),
		AffectedAuthors: authors,
		AnalysisSummary: generateSummary(failures, authors),
		TestFailures:    failures,
	}

	// Generate report
	reportPath, err := generateReport(result)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	result.ReportPath = reportPath

	// Set GitHub Actions outputs
	setGitHubOutput("failure-count", fmt.Sprintf("%d", result.FailureCount))
	setGitHubOutput("affected-authors", mustMarshalJSON(result.AffectedAuthors))
	setGitHubOutput("analysis-summary", result.AnalysisSummary)
	setGitHubOutput("report-path", result.ReportPath)

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
	return fmt.Sprintf(`sum by (parent_test_name, resources_ci_github_workflow_run_head_branch) (
        count_over_time({service_name="%s", service_namespace="cicd-o11y"} |= "--- FAIL: Test" | json | __error__="" | resources_ci_github_workflow_run_conclusion!="cancelled" | line_format "{{.body}}" | regexp "--- FAIL: (?P<test_name>.*) \\(\\d" | line_format "{{.test_name}}" | regexp `+"`(?P<parent_test_name>Test[a-z0-9A-Z_]+)`"+`[7d])
)`, repository)
}

func parseTestFailures(logsJSON, workingDir string) ([]TestFailure, error) {
	var lokiResp LokiResponse
	if err := json.Unmarshal([]byte(logsJSON), &lokiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response: %w", err)
	}

	// Parse raw results grouped by test name and branch
	var branchFailures []TestFailureByBranch
	for _, result := range lokiResp.Data.Result {
		parentTestName := result.Metric["parent_test_name"]
		branch := result.Metric["resources_ci_github_workflow_run_head_branch"]

		if parentTestName == "" || branch == "" {
			continue
		}

		// Get the failure count - use the latest value or single value
		var failureCount int
		var timestamp string

		if len(result.Values) > 0 {
			// Multiple values over time - use the latest
			lastValue := result.Values[len(result.Values)-1]
			if len(lastValue) >= 2 {
				timestamp = fmt.Sprintf("%v", lastValue[0])
				if count, err := strconv.Atoi(fmt.Sprintf("%.0f", lastValue[1])); err == nil {
					failureCount = count
				}
			}
		} else if len(result.Value) >= 2 {
			// Single value
			timestamp = fmt.Sprintf("%v", result.Value[0])
			if count, err := strconv.Atoi(fmt.Sprintf("%.0f", result.Value[1])); err == nil {
				failureCount = count
			}
		}

		if failureCount > 0 {
			branchFailures = append(branchFailures, TestFailureByBranch{
				TestName:     parentTestName,
				Branch:       branch,
				FailureCount: failureCount,
				Timestamp:    timestamp,
			})
		}
	}

	// Aggregate by test name and apply flaky test detection
	return detectFlakyTests(branchFailures, workingDir), nil
}

func detectFlakyTests(branchFailures []TestFailureByBranch, workingDir string) []TestFailure {
	// Group failures by test name
	testMap := make(map[string]map[string]int) // testName -> branch -> failureCount
	timestamps := make(map[string]string)      // testName -> latest timestamp

	for _, failure := range branchFailures {
		if testMap[failure.TestName] == nil {
			testMap[failure.TestName] = make(map[string]int)
		}
		testMap[failure.TestName][failure.Branch] = failure.FailureCount

		// Keep the latest timestamp for each test
		if timestamps[failure.TestName] == "" || failure.Timestamp > timestamps[failure.TestName] {
			timestamps[failure.TestName] = failure.Timestamp
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
		if isFlaky {
			filePath, err := findTestFilePath(workingDir, testName)
			if err != nil {
				log.Printf("Warning: skipping test %s - %v", testName, err)
				continue
			}

			// Create branch summary
			var branchSummary []string
			for branch, count := range branches {
				branchSummary = append(branchSummary, fmt.Sprintf("%s:%d", branch, count))
			}

			flakyTests = append(flakyTests, TestFailure{
				TestName:     testName,
				FilePath:     filePath,
				Timestamp:    timestamps[testName],
				Message:      fmt.Sprintf("Flaky test %s failed %d times across branches: %s", testName, totalFailures, strings.Join(branchSummary, ", ")),
				BranchCounts: branches,
				IsFlaky:      true,
			})
		}
	}

	return flakyTests
}

func findTestFilePath(workingDir, testName string) (string, error) {
	// Search for Go test files containing the test function
	if !strings.HasPrefix(testName, "Test") {
		return "", fmt.Errorf("invalid test name format: %s", testName)
	}

	// Use find command to search for files containing the test function
	// Look for "func TestName(" pattern in *_test.go files
	findCmd := exec.Command("find", ".", "-name", "*_test.go", "-type", "f", "-exec", "grep", "-l", fmt.Sprintf("func %s(", testName), "{}", ";")
	findCmd.Dir = workingDir

	// Execute the search
	result, err := findCmd.Output()
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

	for _, failure := range failures {
		authors, err := getFileAuthors(workingDir, failure.FilePath, failure.TestName, githubToken)
		if err != nil {
			log.Printf("Warning: failed to get authors for test %s in %s: %v", failure.TestName, failure.FilePath, err)
			continue
		}

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
	// Use git log -L to find commits that modified the specific test function
	cmd := exec.Command("git", "log", "-L", fmt.Sprintf(":%s:%s", testName, filePath), "--pretty=format:%H", "-s")
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

	// Extract unique commit hashes
	commitHashes := make(map[string]bool)
	for _, line := range lines {
		hash := strings.TrimSpace(line)
		if hash != "" {
			commitHashes[hash] = true
		}
	}

	if len(commitHashes) == 0 {
		return []string{}, nil
	}

	// Get GitHub usernames for the commits
	var authors []string
	for hash := range commitHashes {
		username, err := getGitHubUsernameForCommit(hash, githubToken)
		if err != nil {
			log.Printf("Warning: failed to get GitHub username for commit %s: %v", hash, err)
			continue
		}
		if username != "" {
			authors = append(authors, username)
		}
	}

	// Remove duplicates
	uniqueAuthors := make(map[string]bool)
	var result_authors []string
	for _, author := range authors {
		if !uniqueAuthors[author] {
			uniqueAuthors[author] = true
			result_authors = append(result_authors, author)
		}
	}

	return result_authors, nil
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
		return "", fmt.Errorf("failed to search commit: %w", err)
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

func generateSummary(failures []TestFailure, authors []string) string {
	if len(failures) == 0 {
		return "No test failures found in the specified time range."
	}

	return fmt.Sprintf("Found %d test failures affecting %d authors. Most common failures: %s",
		len(failures), len(authors), getMostCommonFailures(failures))
}

func getMostCommonFailures(failures []TestFailure) string {
	if len(failures) == 0 {
		return "none"
	}

	// Calculate total failure count for each test across all branches
	testFailureCounts := make(map[string]int)
	for _, failure := range failures {
		totalCount := 0
		for _, count := range failure.BranchCounts {
			totalCount += count
		}
		testFailureCounts[failure.TestName] = totalCount
	}

	// Sort tests by failure count (descending)
	type testCount struct {
		name  string
		count int
	}

	var sortedTests []testCount
	for testName, count := range testFailureCounts {
		sortedTests = append(sortedTests, testCount{name: testName, count: count})
	}

	// Sort by count (highest first)
	for i := 0; i < len(sortedTests); i++ {
		for j := i + 1; j < len(sortedTests); j++ {
			if sortedTests[j].count > sortedTests[i].count {
				sortedTests[i], sortedTests[j] = sortedTests[j], sortedTests[i]
			}
		}
	}

	// Take top 5 and format as "TestName (X failures)"
	limit := min(5, len(sortedTests))
	topFailures := make([]string, limit)
	for i := 0; i < limit; i++ {
		topFailures[i] = fmt.Sprintf("%s (%d failures)", sortedTests[i].name, sortedTests[i].count)
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
