package main

import (
	"os"
	"testing"
)

// Mock implementations for PR1 - basic Loki and analysis functionality

type MockLokiClient struct {
	response *LokiResponse
	err      error
}

func (m *MockLokiClient) FetchLogs() (*LokiResponse, error) {
	return m.response, m.err
}

type MockFileSystem struct {
	writeFileFunc func(filename string, data []byte, perm os.FileMode) error
}

func (m *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if m.writeFileFunc != nil {
		return m.writeFileFunc(filename, data, perm)
	}
	return nil
}

// Test data for PR1
func createTestLokiResponse() *LokiResponse {
	return &LokiResponse{
		Status: "success",
		Data: LokiData{
			ResultType: "streams",
			Result: []LokiResult{
				{
					Stream: map[string]string{
						"parent_test_name": "TestFlakyExample",
						"resources_ci_github_workflow_run_head_branch": "main",
						"resources_ci_github_workflow_run_html_url":    "https://github.com/test/repo/actions/runs/123",
					},
					Values: [][]string{
						{"1640995200000000000", "TestFlakyExample"},
					},
				},
				{
					Stream: map[string]string{
						"parent_test_name": "TestFlakyExample",
						"resources_ci_github_workflow_run_head_branch": "feature-1",
						"resources_ci_github_workflow_run_html_url":    "https://github.com/test/repo/actions/runs/124",
					},
					Values: [][]string{
						{"1640995300000000000", "TestFlakyExample"},
					},
				},
				{
					Stream: map[string]string{
						"parent_test_name": "TestAnotherFlaky",
						"resources_ci_github_workflow_run_head_branch": "main",
						"resources_ci_github_workflow_run_html_url":    "https://github.com/test/repo/actions/runs/125",
					},
					Values: [][]string{
						{"1640995400000000000", "TestAnotherFlaky"},
					},
				},
			},
		},
	}
}

func TestTestFailureAnalyzer_AnalyzeFailures(t *testing.T) {
	// Test successful analysis
	mockLoki := &MockLokiClient{
		response: createTestLokiResponse(),
		err:      nil,
	}
	mockFS := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(mockLoki, mockFS)

	config := Config{
		Repository: "test/repo",
		TimeRange:  "1h",
		TopK:       3,
	}

	report, err := analyzer.AnalyzeFailures(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report == nil {
		t.Fatal("Expected report to be non-nil")
	}

	if report.TestCount == 0 {
		t.Error("Expected test count to be greater than 0")
	}

	if len(report.FlakyTests) == 0 {
		t.Error("Expected flaky tests to be found")
	}
}

func TestTestFailureAnalyzer_AnalyzeFailures_LokiError(t *testing.T) {
	// Test Loki error handling
	mockLoki := &MockLokiClient{
		response: nil,
		err:      &MockError{message: "loki connection failed"},
	}
	mockFS := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(mockLoki, mockFS)

	config := Config{
		Repository: "test/repo",
		TimeRange:  "1h",
		TopK:       3,
	}

	_, err := analyzer.AnalyzeFailures(config)
	if err == nil {
		t.Fatal("Expected error when Loki fails")
	}

	if !contains(err.Error(), "failed to fetch logs from Loki") {
		t.Errorf("Expected Loki error message, got: %v", err)
	}
}

func TestTestFailureAnalyzer_ActionReport(t *testing.T) {
	mockLoki := &MockLokiClient{}
	mockFS := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(mockLoki, mockFS)

	// Test with empty report
	emptyReport := &FailuresReport{
		TestCount:  0,
		FlakyTests: []FlakyTest{},
	}

	config := Config{}
	err := analyzer.ActionReport(emptyReport, config)
	if err != nil {
		t.Fatalf("Expected no error for empty report, got: %v", err)
	}

	// Test with flaky tests
	reportWithTests := &FailuresReport{
		TestCount: 1,
		FlakyTests: []FlakyTest{
			{
				TestName:      "TestExample",
				TotalFailures: 5,
				BranchCounts:  map[string]int{"main": 3, "feature": 2},
			},
		},
	}

	err = analyzer.ActionReport(reportWithTests, config)
	if err != nil {
		t.Fatalf("Expected no error for report with tests, got: %v", err)
	}
}

func TestGenerateSummary(t *testing.T) {
	// Test empty flaky tests
	summary := generateSummary([]FlakyTest{})
	expected := "No flaky tests found in the specified time range."
	if summary != expected {
		t.Errorf("Expected: %s, got: %s", expected, summary)
	}

	// Test with flaky tests
	flakyTests := []FlakyTest{
		{
			TestName:      "TestExample",
			TotalFailures: 5,
		},
		{
			TestName:      "TestAnother",
			TotalFailures: 3,
		},
	}

	summary = generateSummary(flakyTests)
	if !contains(summary, "Found 2 flaky tests") {
		t.Errorf("Expected summary to contain count, got: %s", summary)
	}
	if !contains(summary, "TestExample") {
		t.Errorf("Expected summary to contain test name, got: %s", summary)
	}
}

func TestFlakyTest_String(t *testing.T) {
	test := FlakyTest{
		TestName:      "TestExample",
		TotalFailures: 5,
	}

	result := test.String()
	expected := "TestExample (5 total failures)"
	if result != expected {
		t.Errorf("Expected: %s, got: %s", expected, result)
	}
}

// Helper types and functions
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			str[:len(substr)] == substr ||
			str[len(str)-len(substr):] == substr ||
			findSubstring(str, substr))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
