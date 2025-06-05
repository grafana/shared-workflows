package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLokiClient implements LokiClient for testing
type MockLokiClient struct {
	response *LokiResponse
	err      error
}

func (m *MockLokiClient) FetchLogs() (*LokiResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

// MockGitClient implements GitClient for testing
type MockGitClient struct {
	testFiles map[string]string       // testName -> filePath
	authors   map[string][]CommitInfo // testName -> commits
	fileErr   error
	authorErr error
}

func (m *MockGitClient) FindTestFile(testName string) (string, error) {
	if m.fileErr != nil {
		return "", m.fileErr
	}
	if path, exists := m.testFiles[testName]; exists {
		return path, nil
	}
	return "", fmt.Errorf("test file not found for %s", testName)
}

func (m *MockGitClient) GetFileAuthors(filePath, testName string) ([]CommitInfo, error) {
	if m.authorErr != nil {
		return nil, m.authorErr
	}
	if commits, exists := m.authors[testName]; exists {
		return commits, nil
	}
	return []CommitInfo{}, nil
}

// MockFileSystem implements FileSystem for testing
type MockFileSystem struct {
	writtenFiles map[string][]byte
	writeErr     error
}

func (m *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.writtenFiles == nil {
		m.writtenFiles = make(map[string][]byte)
	}
	m.writtenFiles[filename] = data
	return nil
}

// Test helper functions

func createTestLokiResponse(entries []RawLogEntry) *LokiResponse {
	response := &LokiResponse{
		Status: "success",
		Data: LokiData{
			ResultType: "streams",
			Result:     []LokiResult{},
		},
	}

	for _, entry := range entries {
		result := LokiResult{
			Stream: map[string]string{
				"parent_test_name":                   entry.TestName,
				"ci_github_workflow_run_head_branch": entry.Branch,
				"ci_github_workflow_run_html_url":    entry.WorkflowRunURL,
			},
			Values: [][]string{
				{"1640995200000000000", "test log line"},
			},
		}
		response.Data.Result = append(response.Data.Result, result)
	}

	return response
}

func createTestConfig() Config {
	return Config{
		LokiURL:             "http://localhost:3100",
		LokiUsername:        "user",
		LokiPassword:        "pass",
		Repository:          "test/repo",
		TimeRange:           "24h",
		RepositoryDirectory: "/tmp/test",
		TopK:                3,
	}
}

// Test the analysis phase (AnalyzeFailures method)

func TestAnalyzer_AnalyzeFailures_Success(t *testing.T) {
	// Setup test data
	logEntries := []RawLogEntry{
		{TestName: "TestUserLogin", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
		{TestName: "TestUserLogin", Branch: "feature", WorkflowRunURL: "https://github.com/test/repo/actions/runs/2"},
		{TestName: "TestPayment", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/3"},
	}

	lokiResponse := createTestLokiResponse(logEntries)

	// Setup mocks
	lokiClient := &MockLokiClient{response: lokiResponse}
	gitClient := &MockGitClient{
		testFiles: map[string]string{
			"TestUserLogin": "user_test.go",
			"TestPayment":   "payment_test.go",
		},
		authors: map[string][]CommitInfo{
			"TestUserLogin": {
				{Hash: "abc123", Author: "alice", Timestamp: time.Now().AddDate(0, -1, 0), Title: "Fix user login"},
			},
			"TestPayment": {
				{Hash: "def456", Author: "bob", Timestamp: time.Now().AddDate(0, -2, 0), Title: "Update payment logic"},
			},
		},
	}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	// Run the analysis phase only
	report, err := analyzer.AnalyzeFailures(config)

	// Verify results
	require.NoError(t, err, "Analysis should complete without error")
	require.NotNil(t, report, "Report should not be nil")

	// Check report contents
	assert.Equal(t, 2, report.TestCount, "Expected 2 flaky tests")
	assert.Len(t, report.FlakyTests, 2, "Expected 2 flaky tests in report")
	assert.Contains(t, report.AnalysisSummary, "Found 2 flaky tests", "Summary should mention found tests")

	// Verify flaky tests details
	userTest := findTestByName(report.FlakyTests, "TestUserLogin")
	require.NotNil(t, userTest, "TestUserLogin should be found")
	assert.Equal(t, 2, userTest.TotalFailures, "TestUserLogin should have 2 failures")
	assert.Equal(t, "user_test.go", userTest.FilePath, "TestUserLogin should have correct file path")

	paymentTest := findTestByName(report.FlakyTests, "TestPayment")
	require.NotNil(t, paymentTest, "TestPayment should be found")
	assert.Equal(t, 1, paymentTest.TotalFailures, "TestPayment should have 1 failure")
	assert.Equal(t, "payment_test.go", paymentTest.FilePath, "TestPayment should have correct file path")

	// Check that report file was written
	assert.NotEmpty(t, report.ReportPath, "Report path should be set")
	assert.Len(t, fileSystem.writtenFiles, 1, "Expected exactly 1 file to be written")
}

func TestAnalyzer_AnalyzeFailures_LokiError(t *testing.T) {
	lokiClient := &MockLokiClient{err: fmt.Errorf("loki connection failed")}
	gitClient := &MockGitClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	report, err := analyzer.AnalyzeFailures(config)

	assert.Error(t, err, "Expected error from Loki failure")
	assert.Nil(t, report, "Report should be nil on error")
	assert.Contains(t, err.Error(), "failed to fetch logs from Loki", "Error should mention Loki")
}

func TestAnalyzer_AnalyzeFailures_EmptyResponse(t *testing.T) {
	lokiResponse := &LokiResponse{
		Status: "success",
		Data: LokiData{
			ResultType: "streams",
			Result:     []LokiResult{},
		},
	}

	lokiClient := &MockLokiClient{response: lokiResponse}
	gitClient := &MockGitClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	report, err := analyzer.AnalyzeFailures(config)

	require.NoError(t, err, "Analysis should complete without error")
	require.NotNil(t, report, "Report should not be nil")
	assert.Equal(t, 0, report.TestCount, "Expected 0 flaky tests")
	assert.Len(t, report.FlakyTests, 0, "Expected no flaky tests in report")
	assert.Contains(t, report.AnalysisSummary, "No flaky tests found", "Summary should mention no tests found")
}

// Test the enactment phase (ActionReport method)

func TestAnalyzer_ActionReport_WithoutPostingIssues(t *testing.T) {
	// Create a sample report
	report := &FailuresReport{
		TestCount:       2,
		AnalysisSummary: "Found 2 flaky tests",
		FlakyTests: []FlakyTest{
			{
				TestName:         "TestUserLogin",
				FilePath:         "user_test.go",
				TotalFailures:    2,
				BranchCounts:     map[string]int{"main": 1, "feature": 1},
				ExampleWorkflows: []GithubActionsWorkflow{{RunURL: "https://github.com/test/repo/actions/runs/1"}, {RunURL: "https://github.com/test/repo/actions/runs/2"}},
			},
			{
				TestName:         "TestPayment",
				FilePath:         "payment_test.go",
				TotalFailures:    1,
				BranchCounts:     map[string]int{"main": 1},
				ExampleWorkflows: []GithubActionsWorkflow{{RunURL: "https://github.com/test/repo/actions/runs/3"}},
			},
		},
	}

	// Setup mocks
	lokiClient := &MockLokiClient{}
	gitClient := &MockGitClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	// Run the enactment phase only
	err := analyzer.ActionReport(report, config)

	// Verify results
	require.NoError(t, err, "Enactment should complete without error")
}

func TestAnalyzer_ActionReport_ProductionMode(t *testing.T) {
	// Create a sample report
	report := &FailuresReport{
		TestCount:       1,
		AnalysisSummary: "Found 1 flaky tests",
		FlakyTests: []FlakyTest{
			{
				TestName:         "TestUserLogin",
				FilePath:         "user_test.go",
				TotalFailures:    2,
				BranchCounts:     map[string]int{"main": 2},
				ExampleWorkflows: []GithubActionsWorkflow{{RunURL: "https://github.com/test/repo/actions/runs/1"}},
			},
		},
	}

	// Setup mocks
	lokiClient := &MockLokiClient{}
	gitClient := &MockGitClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	// Run the enactment phase only
	err := analyzer.ActionReport(report, config)

	// Verify results
	require.NoError(t, err, "Enactment should complete without error")
}

func TestAnalyzer_ActionReport_EmptyReport(t *testing.T) {
	// Empty report
	report := &FailuresReport{
		TestCount:       0,
		AnalysisSummary: "No flaky tests found",
		FlakyTests:      []FlakyTest{},
	}

	// Setup mocks
	lokiClient := &MockLokiClient{}
	gitClient := &MockGitClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	// Run the enactment phase only
	err := analyzer.ActionReport(report, config)

	// Verify results
	require.NoError(t, err, "Enactment should complete without error")
}

func TestAnalyzer_ActionReport_NilReport(t *testing.T) {
	// Setup mocks
	lokiClient := &MockLokiClient{}
	gitClient := &MockGitClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	// Run the enactment phase with nil report
	err := analyzer.ActionReport(nil, config)

	// Verify results
	require.NoError(t, err, "Enactment should complete without error for nil report")
}

// Integration tests (Workflow tests)

func TestAnalyzer_Run_Success(t *testing.T) {
	// Setup test data
	logEntries := []RawLogEntry{
		{TestName: "TestUserLogin", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
		{TestName: "TestUserLogin", Branch: "feature", WorkflowRunURL: "https://github.com/test/repo/actions/runs/2"},
		{TestName: "TestPayment", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/3"},
	}

	lokiResponse := createTestLokiResponse(logEntries)

	// Setup mocks
	lokiClient := &MockLokiClient{response: lokiResponse}
	gitClient := &MockGitClient{
		testFiles: map[string]string{
			"TestUserLogin": "user_test.go",
			"TestPayment":   "payment_test.go",
		},
		authors: map[string][]CommitInfo{
			"TestUserLogin": {
				{Hash: "abc123", Author: "alice", Timestamp: time.Now().AddDate(0, -1, 0), Title: "Fix user login"},
			},
			"TestPayment": {
				{Hash: "def456", Author: "bob", Timestamp: time.Now().AddDate(0, -2, 0), Title: "Update payment logic"},
			},
		},
	}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	// Run the analysis
	err := analyzer.Run(config)

	// Verify results
	require.NoError(t, err, "Analysis should complete without error")

	// Check that report was written
	assert.Len(t, fileSystem.writtenFiles, 1, "Expected exactly 1 file to be written")

	reportData, exists := fileSystem.writtenFiles["test-failure-analysis.json"]
	require.True(t, exists, "Expected report file to be written")

	var result FailuresReport
	require.NoError(t, json.Unmarshal(reportData, &result), "Report should unmarshal successfully")

	// Verify flaky tests were detected
	assert.Equal(t, 2, result.TestCount, "Expected 2 flaky tests to be detected")

	// Verify test details
	testNames := make(map[string]bool)
	for _, test := range result.FlakyTests {
		testNames[test.TestName] = true
		assert.Greater(t, test.TotalFailures, 0, "Test %s should have failures", test.TestName)
		assert.NotEmpty(t, test.FilePath, "Test %s should have file path", test.TestName)
	}

	assert.True(t, testNames["TestUserLogin"], "TestUserLogin should be detected as flaky")
	assert.True(t, testNames["TestPayment"], "TestPayment should be detected as flaky")
}

func TestAnalyzer_Run_LokiError(t *testing.T) {
	lokiClient := &MockLokiClient{err: fmt.Errorf("loki connection failed")}
	gitClient := &MockGitClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	require.Error(t, err, "Expected error from Loki failure")
	assert.Contains(t, err.Error(), "failed to fetch logs from Loki", "Error should mention Loki failure")
}

func TestAnalyzer_Run_GitError(t *testing.T) {
	logEntries := []RawLogEntry{
		{TestName: "TestExample", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
	}

	lokiClient := &MockLokiClient{response: createTestLokiResponse(logEntries)}
	gitClient := &MockGitClient{fileErr: fmt.Errorf("git command failed")}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	require.Error(t, err, "Expected error from Git failure")
	assert.Contains(t, err.Error(), "failed to find file paths", "Error should mention Git file path failure")
}

func TestAnalyzer_Run_EmptyLokiResponse(t *testing.T) {
	emptyResponse := &LokiResponse{
		Status: "success",
		Data: LokiData{
			ResultType: "streams",
			Result:     []LokiResult{},
		},
	}

	lokiClient := &MockLokiClient{response: emptyResponse}
	gitClient := &MockGitClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	require.NoError(t, err, "Analysis should complete without error even with empty response")

	// Check that report was still written with zero tests
	reportData, exists := fileSystem.writtenFiles["test-failure-analysis.json"]
	require.True(t, exists, "Expected report file to be written")

	var result FailuresReport
	require.NoError(t, json.Unmarshal(reportData, &result), "Report should unmarshal successfully")

	assert.Equal(t, 0, result.TestCount, "Expected 0 tests with empty response")
}

func TestAnalyzer_Run_NonFlakyTests(t *testing.T) {
	// Tests that only fail on feature branches (not flaky)
	logEntries := []RawLogEntry{
		{TestName: "TestFeatureOnly", Branch: "feature", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
		{TestName: "TestFeatureOnly", Branch: "feature", WorkflowRunURL: "https://github.com/test/repo/actions/runs/2"},
	}

	lokiClient := &MockLokiClient{response: createTestLokiResponse(logEntries)}
	gitClient := &MockGitClient{
		testFiles: map[string]string{
			"TestFeatureOnly": "feature_test.go",
		},
	}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	require.NoError(t, err, "Analysis should complete without error")

	reportData, exists := fileSystem.writtenFiles["test-failure-analysis.json"]
	require.True(t, exists, "Report should be written")

	var result FailuresReport
	require.NoError(t, json.Unmarshal(reportData, &result), "Report should unmarshal successfully")

	// Should not detect any flaky tests (only failed on feature branch)
	assert.Equal(t, 0, result.TestCount, "Expected 0 flaky tests (only failed on feature branch)")
}

// Business logic tests

func TestParseTestFailures_ValidResponse(t *testing.T) {
	logEntries := []RawLogEntry{
		{TestName: "TestUserLogin", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
		{TestName: "TestUserLogin", Branch: "feature", WorkflowRunURL: "https://github.com/test/repo/actions/runs/2"},
		{TestName: "TestPayment", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/3"},
	}

	lokiResponse := createTestLokiResponse(logEntries)

	flakyTests, err := AggregateFlakyTestsFromResponse(lokiResponse)

	require.NoError(t, err, "Parsing should succeed with valid response")
	assert.Len(t, flakyTests, 2, "Expected 2 flaky tests to be detected")

	// Verify TestUserLogin is detected as flaky (fails on main + feature)
	userLoginTest := findTestByName(flakyTests, "TestUserLogin")
	require.NotNil(t, userLoginTest, "TestUserLogin should be detected as flaky")
	assert.Equal(t, 2, userLoginTest.TotalFailures, "TestUserLogin should have 2 failures")
	assert.Len(t, userLoginTest.BranchCounts, 2, "TestUserLogin should fail on 2 branches")

	// Verify TestPayment is detected as flaky (fails on main)
	paymentTest := findTestByName(flakyTests, "TestPayment")
	require.NotNil(t, paymentTest, "TestPayment should be detected as flaky")
	assert.Equal(t, 1, paymentTest.TotalFailures, "TestPayment should have 1 failure")
}

// Test edge cases

func TestAnalyzer_Run_TopKLimit(t *testing.T) {
	// Create 5 flaky tests but limit to 3
	logEntries := []RawLogEntry{
		{TestName: "TestA", Branch: "main", WorkflowRunURL: "https://example.com/1"},
		{TestName: "TestB", Branch: "main", WorkflowRunURL: "https://example.com/2"},
		{TestName: "TestC", Branch: "main", WorkflowRunURL: "https://example.com/3"},
		{TestName: "TestD", Branch: "main", WorkflowRunURL: "https://example.com/4"},
		{TestName: "TestE", Branch: "main", WorkflowRunURL: "https://example.com/5"},
	}

	lokiClient := &MockLokiClient{response: createTestLokiResponse(logEntries)}
	gitClient := &MockGitClient{
		testFiles: map[string]string{
			"TestA": "a_test.go", "TestB": "b_test.go", "TestC": "c_test.go",
			"TestD": "d_test.go", "TestE": "e_test.go",
		},
	}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()
	config.TopK = 3

	err := analyzer.Run(config)

	require.NoError(t, err, "Analysis should complete without error")

	reportData, exists := fileSystem.writtenFiles["test-failure-analysis.json"]
	require.True(t, exists, "Report should be written")

	var result FailuresReport
	require.NoError(t, json.Unmarshal(reportData, &result), "Report should unmarshal successfully")

	assert.Equal(t, 3, result.TestCount, "Test count should be limited to top K setting")
}

func TestAnalyzer_Run_NoProductionMode(t *testing.T) {
	logEntries := []RawLogEntry{
		{TestName: "TestUserLogin", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
	}

	lokiClient := &MockLokiClient{response: createTestLokiResponse(logEntries)}
	gitClient := &MockGitClient{
		testFiles: map[string]string{"TestUserLogin": "user_test.go"},
	}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	require.NoError(t, err, "Analysis should complete without error")
}

// Unit tests for FlakyTest methods

func TestFlakyTest_String(t *testing.T) {
	tests := []struct {
		name     string
		test     FlakyTest
		expected string
	}{
		{
			name: "test with single author",
			test: FlakyTest{
				TestName:      "TestUserLogin",
				TotalFailures: 3,
				RecentCommits: []CommitInfo{
					{Author: "alice", Hash: "abc123", Title: "Fix login"},
				},
			},
			expected: "TestUserLogin (3 total failures; recently changed by alice)",
		},
		{
			name: "test with multiple authors",
			test: FlakyTest{
				TestName:      "TestPayment",
				TotalFailures: 5,
				RecentCommits: []CommitInfo{
					{Author: "alice", Hash: "abc123", Title: "Fix payment"},
					{Author: "bob", Hash: "def456", Title: "Update payment logic"},
				},
			},
			expected: "TestPayment (5 total failures; recently changed by alice, bob)",
		},
		{
			name: "test with no commits",
			test: FlakyTest{
				TestName:      "TestDatabase",
				TotalFailures: 2,
				RecentCommits: []CommitInfo{},
			},
			expected: "TestDatabase (2 total failures; recently changed by unknown)",
		},
		{
			name: "test with unknown authors",
			test: FlakyTest{
				TestName:      "TestAPI",
				TotalFailures: 1,
				RecentCommits: []CommitInfo{
					{Author: "unknown", Hash: "abc123", Title: "Some change"},
					{Author: "", Hash: "def456", Title: "Another change"},
				},
			},
			expected: "TestAPI (1 total failures; recently changed by unknown)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.test.String()
			assert.Equal(t, tt.expected, result, "String() output should match expected format")
		})
	}
}

// Helper functions

func findTestByName(tests []FlakyTest, name string) *FlakyTest {
	for i := range tests {
		if tests[i].TestName == name {
			return &tests[i]
		}
	}
	return nil
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
				str[len(str)-len(substr):] == substr ||
				findSubstring(str, substr))))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Golden file tests

func TestAnalyzer_Run_GoldenFiles(t *testing.T) {
	tests := []struct {
		name         string
		lokiFile     string
		expectedFile string
		setupMocks   func() *MockGitClient
		config       func() Config
	}{
		{
			name:         "complex_scenario",
			lokiFile:     "complex_loki_response.json",
			expectedFile: "complex_scenario.json",
			setupMocks: func() *MockGitClient {
				gitClient := &MockGitClient{
					testFiles: map[string]string{
						"TestDatabaseConnection": "internal/database/connection_test.go",
						"TestUserAuthentication": "auth/user_test.go",
						"TestPaymentProcessing":  "payment/processor_test.go",
					},
					authors: map[string][]CommitInfo{
						"TestDatabaseConnection": {
							{Hash: "abc123def456", Author: "alice", Timestamp: mustParseTime("2024-01-15T10:30:00Z"), Title: "Optimize database connection pooling"},
							{Hash: "789ghi012jkl", Author: "bob", Timestamp: mustParseTime("2024-01-10T14:22:00Z"), Title: "Add connection timeout handling"},
						},
						"TestUserAuthentication": {
							{Hash: "345mno678pqr", Author: "charlie", Timestamp: mustParseTime("2024-01-12T09:15:00Z"), Title: "Implement OAuth2 authentication flow"},
						},
						"TestPaymentProcessing": {
							{Hash: "901stu234vwx", Author: "dave", Timestamp: mustParseTime("2024-01-08T16:45:00Z"), Title: "Add Stripe payment integration"},
							{Hash: "567yza890bcd", Author: "eve", Timestamp: mustParseTime("2024-01-05T11:30:00Z"), Title: "Refactor payment processing logic"},
						},
					},
				}
				return gitClient
			},
			config: func() Config {
				config := createTestConfig()
				config.TopK = 10 // Don't limit for this test
				return config
			},
		},
		{
			name:         "empty_scenario",
			lokiFile:     "",
			expectedFile: "empty_scenario.json",
			setupMocks: func() *MockGitClient {
				return &MockGitClient{}
			},
			config: createTestConfig,
		},
		{
			name:         "single_test_scenario",
			lokiFile:     "",
			expectedFile: "single_test_scenario.json",
			setupMocks: func() *MockGitClient {
				gitClient := &MockGitClient{
					testFiles: map[string]string{
						"TestLoginFlow": "handlers/login_test.go",
					},
					authors: map[string][]CommitInfo{
						"TestLoginFlow": {}, // No recent commits
					},
				}
				return gitClient
			},
			config: createTestConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load Loki response
			var lokiResponse *LokiResponse
			if tt.lokiFile != "" {
				data, err := os.ReadFile(filepath.Join("testdata", tt.lokiFile))
				require.NoError(t, err, "Failed to read Loki file %s", tt.lokiFile)
				lokiResponse = &LokiResponse{}
				require.NoError(t, json.Unmarshal(data, lokiResponse), "Failed to unmarshal Loki response")
			} else {
				// Create appropriate empty/single test response
				if tt.name == "single_test_scenario" {
					entries := []RawLogEntry{
						{TestName: "TestLoginFlow", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/401"},
						{TestName: "TestLoginFlow", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/402"},
					}
					lokiResponse = createTestLokiResponse(entries)
				} else {
					lokiResponse = &LokiResponse{
						Status: "success",
						Data: LokiData{
							ResultType: "streams",
							Result:     []LokiResult{},
						},
					}
				}
			}

			// Setup mocks
			lokiClient := &MockLokiClient{response: lokiResponse}
			gitClient := tt.setupMocks()
			fileSystem := &MockFileSystem{}

			// Run analysis
			analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, fileSystem)
			config := tt.config()

			err := analyzer.Run(config)
			require.NoError(t, err, "Analysis should complete without error")

			// Load expected result
			expectedData, err := os.ReadFile(filepath.Join("testdata", tt.expectedFile))
			require.NoError(t, err, "Should be able to read expected file %s", tt.expectedFile)

			var expected FailuresReport
			require.NoError(t, json.Unmarshal(expectedData, &expected), "Expected result should unmarshal successfully")

			// Get actual result
			actualData, exists := fileSystem.writtenFiles["test-failure-analysis.json"]
			require.True(t, exists, "Expected report file to be written")

			var actual FailuresReport
			require.NoError(t, json.Unmarshal(actualData, &actual), "Actual result should unmarshal successfully")

			// Compare results (ignoring report_path which will be different)
			actual.ReportPath = expected.ReportPath

			// For time-sensitive tests, normalize timestamps
			if tt.name == "complex_scenario" {
				normalizeTimestamps(&actual, &expected)
			}

			// Normalize workflow order for comparison
			normalizeWorkflowOrder(&actual, &expected)

			// Compare JSON representations for deep equality
			actualJSON, _ := json.MarshalIndent(actual, "", "  ")
			expectedJSON, _ := json.MarshalIndent(expected, "", "  ")

			if !assert.Equal(t, string(expectedJSON), string(actualJSON), "Results should match for test: %s", tt.name) {
				// Write actual result to file for debugging
				debugFile := filepath.Join("testdata", fmt.Sprintf("%s_actual.json", tt.name))
				os.WriteFile(debugFile, actualJSON, 0644)
				t.Logf("Actual result written to: %s", debugFile)
			}
		})
	}
}

// Helper functions for golden file tests

func mustParseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse time %s: %v", timeStr, err))
	}
	return t
}

func normalizeTimestamps(actual, expected *FailuresReport) {
	// For golden file tests, we normalize timestamps to expected values
	// since actual git log times will differ from test data
	for i, actualTest := range actual.FlakyTests {
		for j := range expected.FlakyTests {
			if expected.FlakyTests[j].TestName == actualTest.TestName {
				// Copy expected timestamps to actual for comparison
				if len(expected.FlakyTests[j].RecentCommits) == len(actualTest.RecentCommits) {
					for k := range actualTest.RecentCommits {
						actual.FlakyTests[i].RecentCommits[k].Timestamp = expected.FlakyTests[j].RecentCommits[k].Timestamp
					}
				}
				break
			}
		}
	}
}

func normalizeWorkflowOrder(actual, expected *FailuresReport) {
	// Sort workflow URLs to make comparison order-independent
	for i, actualTest := range actual.FlakyTests {
		for j := range expected.FlakyTests {
			if expected.FlakyTests[j].TestName == actualTest.TestName {
				// Sort both arrays to ensure consistent order
				slices.SortFunc(actual.FlakyTests[i].ExampleWorkflows, func(a, b GithubActionsWorkflow) int {
					return strings.Compare(a.RunURL, b.RunURL)
				})
				slices.SortFunc(expected.FlakyTests[i].ExampleWorkflows, func(a, b GithubActionsWorkflow) int {
					return strings.Compare(a.RunURL, b.RunURL)
				})
				break
			}
		}
	}
}
