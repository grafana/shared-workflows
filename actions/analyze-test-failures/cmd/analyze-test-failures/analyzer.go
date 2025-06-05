package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FileSystem interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

type TestFailureAnalyzer struct {
	lokiClient LokiClient
	fileSystem FileSystem
}

type FlakyTest struct {
	TestName         string         `json:"test_name"`
	TotalFailures    int            `json:"total_failures"`
	BranchCounts     map[string]int `json:"branch_counts"`
	ExampleWorkflows []string       `json:"example_workflows"`
}

func (f *FlakyTest) String() string {
	return fmt.Sprintf("%s (%d total failures)", f.TestName, f.TotalFailures)
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

func NewTestFailureAnalyzer(loki LokiClient, fs FileSystem) *TestFailureAnalyzer {
	return &TestFailureAnalyzer{
		lokiClient: loki,
		fileSystem: fs,
	}
}

func NewDefaultTestFailureAnalyzer(config Config) *TestFailureAnalyzer {
	lokiClient := NewDefaultLokiClient(config)
	fileSystem := &DefaultFileSystem{}

	return NewTestFailureAnalyzer(lokiClient, fileSystem)
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
	log.Printf("📝 Report generated successfully - no additional actions in this version")
	log.Printf("✅ Analysis complete!")
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