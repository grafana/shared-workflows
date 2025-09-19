package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileSystem interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

type GitClient interface {
	FindTestFile(testName string) (string, error)
	TestCommits(filePath, testName string) ([]CommitInfo, error)
}

type GitHubClient interface {
	GetUsernameForCommit(commitHash string) (string, error)
	CreateOrUpdateIssue(test FlakyTest) error
	SearchForExistingIssue(issueTitle string) (string, error)
	AddCommentToIssue(issueURL string, test FlakyTest) error
	ReopenIssue(issueURL string) error
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
	TestName         string                  `json:"test_name"`
	FilePath         string                  `json:"file_path"`
	TotalFailures    int                     `json:"total_failures"`
	BranchCounts     map[string]int          `json:"branch_counts"`
	ExampleWorkflows []GithubActionsWorkflow `json:"example_workflows"`
	RecentCommits    []CommitInfo            `json:"recent_commits"`
}

type GithubActionsWorkflow struct {
	RunURL  string `json:"run_url"`
	JobName string `json:"job_name"`
	Attempt int    `json:"attempt"`
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
	flakyTests, err := AggregateFlakyTestsFromResponse(lokiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test failures: %w", err)
	}
	if len(flakyTests) > config.TopK {
		flakyTests = flakyTests[:config.TopK]
	}

	log.Printf("🧪 Found %d flaky tests that meet criteria", len(flakyTests))
	log.Printf("📁 Finding test files in repository...")
	err = t.findFilePaths(flakyTests)
	if err != nil {
		return nil, fmt.Errorf("failed to find file paths for flaky tests: %w", err)
	}

	log.Printf("👥 Finding authors of flaky tests...")
	err = t.findTestAuthors(flakyTests)
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
		err := t.previewIssuesForFlakyTests(report.FlakyTests, config)
		if err != nil {
			return fmt.Errorf("failed to preview GitHub issues: %w", err)
		}
	} else {
		log.Printf("📝 Creating GitHub issues for flaky tests...")
		err := t.createIssuesForFlakyTests(report.FlakyTests)
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

func (t *TestFailureAnalyzer) findFilePaths(flakyTests []FlakyTest) error {
	for i, test := range flakyTests {
		filePath, err := t.gitClient.FindTestFile(test.TestName)
		if err != nil {
			return fmt.Errorf("failed to find file path for test %s: %w", test.TestName, err)
		}
		flakyTests[i].FilePath = filePath
	}
	return nil
}

func (t *TestFailureAnalyzer) findTestAuthors(flakyTests []FlakyTest) error {
	for i, test := range flakyTests {
		commits, err := t.gitClient.TestCommits(test.FilePath, test.TestName)
		if err != nil {
			return fmt.Errorf("failed to get authors for test %s in %s: %w", test.TestName, test.FilePath, err)
		}
		flakyTests[i].RecentCommits = commits

		var authors []string
		for commitIdx, commit := range commits {
			authors = append(authors, commit.Author)
			commits[commitIdx].Author, err = t.githubClient.GetUsernameForCommit(commit.Hash)
			if err != nil {
				return fmt.Errorf("failed to get author for test %s in %s: %w", test.TestName, test.FilePath, err)
			}
		}
	}
	return nil
}

func (t *TestFailureAnalyzer) createIssuesForFlakyTests(flakyTests []FlakyTest) error {
	for _, test := range flakyTests {
		err := t.githubClient.CreateOrUpdateIssue(test)
		if err != nil {
			log.Printf("Warning: failed to create issue for test %s: %v", test.TestName, err)
		}
	}
	return nil
}

func (t *TestFailureAnalyzer) previewIssuesForFlakyTests(flakyTests []FlakyTest, config Config) error {
	for _, test := range flakyTests {
		err := previewIssueForTest(test, config)
		if err != nil {
			log.Printf("Warning: failed to preview issue for test %s: %v", test.TestName, err)
		}
	}
	return nil
}

func previewIssueForTest(test FlakyTest, config Config) error {
	issueTitle := fmt.Sprintf("Flaky %s", test.TestName)

	log.Printf("📄 Would create issue for %s:", test.TestName)
	log.Printf("Title: %s", issueTitle)
	log.Printf("Labels: flaky-test")
	log.Printf("")

	// Generate the actual markdown content that would be used
	issueBody, err := generateInitialIssueBody(test)
	if err != nil {
		log.Printf("Warning: failed to generate issue body preview: %v", err)
		return nil
	}

	commentBody, err := generateCommentBody(test, config)
	if err != nil {
		log.Printf("Warning: failed to generate comment body preview: %v", err)
		return nil
	}

	log.Printf("Initial Issue Body Markdown:")
	log.Printf("────────────────────────────────────────────────────────────────────────")
	log.Printf("%s", issueBody)
	log.Printf("────────────────────────────────────────────────────────────────────────")
	log.Printf("")

	log.Printf("Comment Body Markdown:")
	log.Printf("────────────────────────────────────────────────────────────────────────")
	log.Printf("%s", commentBody)
	log.Printf("────────────────────────────────────────────────────────────────────────")
	log.Printf("")

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
