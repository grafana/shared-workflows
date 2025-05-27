package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type FileSystem interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

type TestFailureAnalyzer struct {
	lokiClient   LokiClient
	gitClient    GitClient
	githubClient GitHubClient
	fileSystem   FileSystem
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

func (f *FlakyTest) String() string {
	var authors []string
	for _, commit := range f.RecentCommits {
		if commit.Author != "" && commit.Author != "unknown" {
			authors = append(authors, commit.Author)
		}
	}
	authorsStr := "unknown"
	if len(authors) > 0 {
		authorsStr = strings.Join(authors, ", ")
	}
	return fmt.Sprintf("%s (%d total failures; recently changed by %s)", f.TestName, f.TotalFailures, authorsStr)
}

type RawLogEntry struct {
	TestName       string `json:"test_name"`
	Branch         string `json:"branch"`
	WorkflowRunURL string `json:"workflow_run_url"`
}

type FailuresReport struct {
	TestCount       int         `json:"test_count"`
	AnalysisSummary string      `json:"analysis_summary"`
	ReportPath      string      `json:"report_path"`
	FlakyTests      []FlakyTest `json:"flaky_tests"`
}

type DefaultFileSystem struct{}

func (fs *DefaultFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func NewTestFailureAnalyzer(loki LokiClient, git GitClient, github GitHubClient, fs FileSystem) *TestFailureAnalyzer {
	return &TestFailureAnalyzer{
		lokiClient:   loki,
		gitClient:    git,
		githubClient: github,
		fileSystem:   fs,
	}
}

func NewDefaultTestFailureAnalyzer(config Config) *TestFailureAnalyzer {
	lokiClient := NewDefaultLokiClient(config)
	gitClient := NewDefaultGitClient(config)
	githubClient := NewDefaultGitHubClient(config)
	fileSystem := &DefaultFileSystem{}

	return NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
}

func (t *TestFailureAnalyzer) AnalyzeFailures(config Config) (*FailuresReport, error) {
	log.Printf("🔍 Starting test failure analysis for repository: %s", config.Repository)
	log.Printf("📅 Time range: %s", config.TimeRange)
	log.Printf("🔗 Loki URL: %s", config.LokiURL)
	log.Printf("📊 Top K tests to process: %d", config.TopK)

	log.Printf("📡 Fetching logs from Loki...")
	lokiResp, err := t.lokiClient.FetchLogs()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs from Loki: %w", err)
	}

	log.Printf("📊 Parsing test failures from log data...")
	flakyTests, err := parseTestFailuresFromResponse(lokiResp, config.RepositoryDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test failures: %w", err)
	}
	if len(flakyTests) > config.TopK {
		flakyTests = flakyTests[:config.TopK]
	}

	log.Printf("🧪 Found %d flaky tests that meet criteria", len(flakyTests))
	log.Printf("📁 Finding test files in repository...")
	err = t.findFilePaths(config.RepositoryDirectory, flakyTests)
	if err != nil {
		return nil, fmt.Errorf("failed to find file paths for flaky tests: %w", err)
	}

	log.Printf("👥 Finding authors of flaky tests...")
	err = t.findTestAuthors(config.RepositoryDirectory, config.GitHubToken, flakyTests)
	if err != nil {
		return nil, fmt.Errorf("failed to find test authors: %w", err)
	}

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

	if flakyTests == nil {
		flakyTests = []FlakyTest{}
	}

	result := FailuresReport{
		TestCount:       len(flakyTests),
		AnalysisSummary: generateSummary(flakyTests),
		FlakyTests:      flakyTests,
	}

	log.Printf("📄 Generating analysis report...")
	reportPath, err := t.generateReport(result)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}
	result.ReportPath = reportPath
	log.Printf("💾 Report saved to: %s", reportPath)

	log.Printf("✅ Analysis complete! Summary: %s", result.AnalysisSummary)
	return &result, nil
}

func (t *TestFailureAnalyzer) ActionReport(report *FailuresReport, config Config) error {
	if report == nil || len(report.FlakyTests) == 0 {
		log.Printf("📝 No flaky tests to enact - skipping GitHub issue creation")
		return nil
	}

	if config.SkipPostingIssues {
		log.Printf("🔍 Dry run mode: Generating issue previews...")
		err := t.previewIssuesForFlakyTests(report.FlakyTests)
		if err != nil {
			return fmt.Errorf("failed to preview GitHub issues: %w", err)
		}
	} else {
		log.Printf("📝 Creating GitHub issues for flaky tests...")
		err := t.createIssuesForFlakyTests(config.Repository, config.GitHubToken, report.FlakyTests)
		if err != nil {
			return fmt.Errorf("failed to create GitHub issues: %w", err)
		}
	}

	log.Printf("✅ Report enactment complete!")
	return nil
}

func (t *TestFailureAnalyzer) Run(config Config) error {
	report, err := t.AnalyzeFailures(config)
	if err != nil {
		return fmt.Errorf("analysis phase failed: %w", err)
	}

	err = t.ActionReport(report, config)
	if err != nil {
		return fmt.Errorf("enactment phase failed: %w", err)
	}

	setGitHubOutput("test-count", fmt.Sprintf("%d", report.TestCount))
	setGitHubOutput("analysis-summary", report.AnalysisSummary)
	setGitHubOutput("report-path", report.ReportPath)

	return nil
}

func (t *TestFailureAnalyzer) findFilePaths(repoDir string, flakyTests []FlakyTest) error {
	for i, test := range flakyTests {
		filePath, err := t.gitClient.FindTestFile(test.TestName)
		if err != nil {
			return fmt.Errorf("failed to find file path for test %s: %w", test.TestName, err)
		}
		flakyTests[i].FilePath = filePath
	}
	return nil
}

func (t *TestFailureAnalyzer) findTestAuthors(repoDir, githubToken string, flakyTests []FlakyTest) error {
	for i, test := range flakyTests {
		commits, err := t.gitClient.GetFileAuthors(test.FilePath, test.TestName)
		if err != nil {
			return fmt.Errorf("failed to get authors for test %s in %s: %w", test.TestName, test.FilePath, err)
		}
		flakyTests[i].RecentCommits = commits

		if len(commits) > 0 {
			var authors []string
			for _, commit := range commits {
				authors = append(authors, commit.Author)
			}
			log.Printf("👤 %s: %s", test.TestName, strings.Join(authors, ", "))
		} else {
			log.Printf("👤 %s: no commits found", test.TestName)
		}
	}
	return nil
}

func (t *TestFailureAnalyzer) createIssuesForFlakyTests(repository, githubToken string, flakyTests []FlakyTest) error {
	for _, test := range flakyTests {
		err := t.githubClient.CreateOrUpdateIssue(test)
		if err != nil {
			log.Printf("Warning: failed to create issue for test %s: %v", test.TestName, err)
		}
	}
	return nil
}

func (t *TestFailureAnalyzer) previewIssuesForFlakyTests(flakyTests []FlakyTest) error {
	for _, test := range flakyTests {
		err := previewIssueForTest(test)
		if err != nil {
			log.Printf("Warning: failed to preview issue for test %s: %v", test.TestName, err)
		}
	}
	return nil
}

func (t *TestFailureAnalyzer) generateReport(result FailuresReport) (string, error) {
	reportPath := "test-failure-analysis.json"

	reportData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := t.fileSystem.WriteFile(reportPath, reportData, 0644); err != nil {
		return "", fmt.Errorf("failed to write report file: %w", err)
	}

	return filepath.Abs(reportPath)
}

func parseTestFailures(logsJSON, repoDir string) ([]FlakyTest, error) {
	var lokiResp LokiResponse
	if err := json.Unmarshal([]byte(logsJSON), &lokiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Loki response: %w", err)
	}

	return parseTestFailuresFromResponse(&lokiResp, repoDir)
}

func parseTestFailuresFromResponse(lokiResp *LokiResponse, repoDir string) ([]FlakyTest, error) {
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

	return detectFlakyTestsFromRawEntries(rawEntries, repoDir), nil
}

func detectFlakyTestsFromRawEntries(rawEntries []RawLogEntry, repoDir string) []FlakyTest {
	testMap := make(map[string]map[string]int)
	exampleWorkflows := make(map[string]map[string]bool)

	for _, entry := range rawEntries {
		if entry.TestName == "" || entry.Branch == "" {
			continue
		}

		if testMap[entry.TestName] == nil {
			testMap[entry.TestName] = make(map[string]int)
			exampleWorkflows[entry.TestName] = make(map[string]bool)
		}

		testMap[entry.TestName][entry.Branch]++

		if entry.WorkflowRunURL != "" && len(exampleWorkflows[entry.TestName]) < 3 {
			exampleWorkflows[entry.TestName][entry.WorkflowRunURL] = true
		}
	}

	var flakyTests []FlakyTest

	for testName, branches := range testMap {
		isFlaky := false
		totalFailures := 0

		for branch, count := range branches {
			totalFailures += count

			if branch == "main" {
				isFlaky = true
			}
		}

		if len(branches) > 1 {
			isFlaky = true
		}

		if !isFlaky {
			continue
		}

		var branchSummary []string
		for branch, count := range branches {
			branchSummary = append(branchSummary, fmt.Sprintf("%s:%d", branch, count))
		}

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

func sortFlakyTests(tests []FlakyTest) []FlakyTest {
	slices.SortFunc(tests, func(a, b FlakyTest) int {
		branchesDelta := len(b.BranchCounts) - len(a.BranchCounts)
		if branchesDelta != 0 {
			return branchesDelta
		}
		if a.TestName < b.TestName {
			return -1
		} else if a.TestName > b.TestName {
			return 1
		}
		return 0
	})
	return tests
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

func generateSummary(flakyTests []FlakyTest) string {
	if len(flakyTests) == 0 {
		return "No flaky tests found in the specified time range."
	}

	return fmt.Sprintf("Found %d flaky tests. Most common tests: %s",
		len(flakyTests), formatFlakyTests(flakyTests))
}

func formatFlakyTests(flakyTests []FlakyTest) string {
	if len(flakyTests) == 0 {
		return "none"
	}

	topTests := make([]string, len(flakyTests))
	for i := 0; i < len(flakyTests); i++ {
		topTests[i] = flakyTests[i].String()
	}

	return strings.Join(topTests, ", ")
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
