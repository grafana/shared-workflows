package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// Mock implementations for testing

// MockLokiClient implements LokiClient for testing
type MockLokiClient struct {
	response string
	err      error
}

func (m *MockLokiClient) FetchLogs(config Config) (string, error) {
	if m.err != nil {
		return "", m.err
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

func (m *MockGitClient) FindTestFile(workingDir, testName string) (string, error) {
	if m.fileErr != nil {
		return "", m.fileErr
	}
	if path, exists := m.testFiles[testName]; exists {
		return path, nil
	}
	return "", fmt.Errorf("test file not found for %s", testName)
}

func (m *MockGitClient) GetFileAuthors(workingDir, filePath, testName, githubToken string) ([]CommitInfo, error) {
	if m.authorErr != nil {
		return nil, m.authorErr
	}
	if commits, exists := m.authors[testName]; exists {
		return commits, nil
	}
	return []CommitInfo{}, nil
}

// MockGitHubClient implements GitHubClient for testing
type MockGitHubClient struct {
	usernames      map[string]string // commitHash -> username
	existingIssues map[string]string // issueTitle -> issueURL
	createdIssues  []string          // track created issues
	addedComments  []string          // track added comments
	reopenedIssues []string          // track reopened issues
	usernameErr    error
	createIssueErr error
	searchIssueErr error
	commentErr     error
	reopenErr      error
}

func (m *MockGitHubClient) GetUsernameForCommit(commitHash, token string) (string, error) {
	if m.usernameErr != nil {
		return "", m.usernameErr
	}
	if username, exists := m.usernames[commitHash]; exists {
		return username, nil
	}
	return "unknown", nil
}

func (m *MockGitHubClient) CreateOrUpdateIssue(repo, token string, test FlakyTest) error {
	if m.createIssueErr != nil {
		return m.createIssueErr
	}
	issueTitle := fmt.Sprintf("Flaky test: %s", test.TestName)

	// Check if issue exists
	if existingURL, exists := m.existingIssues[issueTitle]; exists {
		m.addedComments = append(m.addedComments, fmt.Sprintf("comment on %s", existingURL))
		return nil
	}

	// Create new issue
	issueURL := fmt.Sprintf("https://github.com/%s/issues/%d", repo, len(m.createdIssues)+1)
	m.createdIssues = append(m.createdIssues, issueURL)
	m.addedComments = append(m.addedComments, fmt.Sprintf("comment on %s", issueURL))
	return nil
}

func (m *MockGitHubClient) SearchForExistingIssue(repository, githubToken, issueTitle string) (string, error) {
	if m.searchIssueErr != nil {
		return "", m.searchIssueErr
	}
	if url, exists := m.existingIssues[issueTitle]; exists {
		return url, nil
	}
	return "", nil
}

func (m *MockGitHubClient) AddCommentToIssue(repository, githubToken, issueURL string, test FlakyTest) error {
	if m.commentErr != nil {
		return m.commentErr
	}
	m.addedComments = append(m.addedComments, fmt.Sprintf("comment on %s", issueURL))
	return nil
}

func (m *MockGitHubClient) ReopenIssue(repository, githubToken, issueURL string) error {
	if m.reopenErr != nil {
		return m.reopenErr
	}
	m.reopenedIssues = append(m.reopenedIssues, issueURL)
	return nil
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

func (m *MockFileSystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Test helper functions

func createTestLokiResponse(entries []RawLogEntry) string {
	response := LokiResponse{
		Status: "success",
		Data: struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Stream map[string]string `json:"stream"`
				Values [][]string        `json:"values"`
			} `json:"result"`
		}{
			ResultType: "streams",
		},
	}

	for _, entry := range entries {
		result := struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		}{
			Stream: map[string]string{
				"parent_test_name": entry.TestName,
				"resources_ci_github_workflow_run_head_branch": entry.Branch,
				"resources_ci_github_workflow_run_html_url":    entry.WorkflowRunURL,
			},
			Values: [][]string{
				{"1640995200000000000", "test log line"},
			},
		}
		response.Data.Result = append(response.Data.Result, result)
	}

	data, _ := json.Marshal(response)
	return string(data)
}

func createTestConfig() Config {
	return Config{
		LokiURL:          "http://localhost:3100",
		LokiUsername:     "user",
		LokiPassword:     "pass",
		Repository:       "test/repo",
		TimeRange:        "24h",
		GitHubToken:      "token",
		WorkingDirectory: "/tmp/test",
		DryRun:           true,
		MaxFailures:      3,
	}
}

// Workflow tests

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
	githubClient := &MockGitHubClient{
		usernames: map[string]string{
			"abc123": "alice",
			"def456": "bob",
		},
	}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
	config := createTestConfig()

	// Run the analysis
	err := analyzer.Run(config)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that report was written
	if len(fileSystem.writtenFiles) != 1 {
		t.Fatalf("Expected 1 file written, got: %d", len(fileSystem.writtenFiles))
	}

	reportData, exists := fileSystem.writtenFiles["test-failure-analysis.json"]
	if !exists {
		t.Fatal("Expected report file to be written")
	}

	var result AnalysisResult
	if err := json.Unmarshal(reportData, &result); err != nil {
		t.Fatalf("Failed to unmarshal report: %v", err)
	}

	// Verify flaky tests were detected
	if result.TestCount != 2 {
		t.Fatalf("Expected 2 flaky tests, got: %d", result.TestCount)
	}

	// Verify test details
	testNames := make(map[string]bool)
	for _, test := range result.FlakyTests {
		testNames[test.TestName] = true
		if test.TotalFailures == 0 {
			t.Errorf("Expected test %s to have failures", test.TestName)
		}
		if test.FilePath == "" {
			t.Errorf("Expected test %s to have file path", test.TestName)
		}
	}

	if !testNames["TestUserLogin"] || !testNames["TestPayment"] {
		t.Error("Expected TestUserLogin and TestPayment to be detected as flaky")
	}
}

func TestAnalyzer_Run_LokiError(t *testing.T) {
	lokiClient := &MockLokiClient{err: fmt.Errorf("loki connection failed")}
	gitClient := &MockGitClient{}
	githubClient := &MockGitHubClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	if err == nil {
		t.Fatal("Expected error from Loki failure")
	}
	if !contains(err.Error(), "failed to fetch logs from Loki") {
		t.Errorf("Expected Loki error message, got: %v", err)
	}
}

func TestAnalyzer_Run_GitError(t *testing.T) {
	logEntries := []RawLogEntry{
		{TestName: "TestExample", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
	}

	lokiClient := &MockLokiClient{response: createTestLokiResponse(logEntries)}
	gitClient := &MockGitClient{fileErr: fmt.Errorf("git command failed")}
	githubClient := &MockGitHubClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	if err == nil {
		t.Fatal("Expected error from Git failure")
	}
	if !contains(err.Error(), "failed to find file paths") {
		t.Errorf("Expected Git error message, got: %v", err)
	}
}

func TestAnalyzer_Run_EmptyLokiResponse(t *testing.T) {
	emptyResponse := `{"status":"success","data":{"resultType":"streams","result":[]}}`

	lokiClient := &MockLokiClient{response: emptyResponse}
	gitClient := &MockGitClient{}
	githubClient := &MockGitHubClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	if err != nil {
		t.Fatalf("Expected no error with empty response, got: %v", err)
	}

	// Check that report was still written with zero tests
	reportData, exists := fileSystem.writtenFiles["test-failure-analysis.json"]
	if !exists {
		t.Fatal("Expected report file to be written")
	}

	var result AnalysisResult
	if err := json.Unmarshal(reportData, &result); err != nil {
		t.Fatalf("Failed to unmarshal report: %v", err)
	}

	if result.TestCount != 0 {
		t.Fatalf("Expected 0 tests with empty response, got: %d", result.TestCount)
	}
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
	githubClient := &MockGitHubClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
	config := createTestConfig()

	err := analyzer.Run(config)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	reportData, _ := fileSystem.writtenFiles["test-failure-analysis.json"]
	var result AnalysisResult
	json.Unmarshal(reportData, &result)

	// Should not detect any flaky tests (only failed on feature branch)
	if result.TestCount != 0 {
		t.Fatalf("Expected 0 flaky tests, got: %d", result.TestCount)
	}
}

// Business logic tests

func TestParseTestFailures_ValidResponse(t *testing.T) {
	logEntries := []RawLogEntry{
		{TestName: "TestUserLogin", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
		{TestName: "TestUserLogin", Branch: "feature", WorkflowRunURL: "https://github.com/test/repo/actions/runs/2"},
		{TestName: "TestPayment", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/3"},
	}

	lokiResponse := createTestLokiResponse(logEntries)

	flakyTests, err := parseTestFailures(lokiResponse, "/tmp/test")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(flakyTests) != 2 {
		t.Fatalf("Expected 2 flaky tests, got: %d", len(flakyTests))
	}

	// Verify TestUserLogin is detected as flaky (fails on main + feature)
	userLoginTest := findTestByName(flakyTests, "TestUserLogin")
	if userLoginTest == nil {
		t.Fatal("Expected TestUserLogin to be detected as flaky")
	}
	if userLoginTest.TotalFailures != 2 {
		t.Errorf("Expected TestUserLogin to have 2 failures, got: %d", userLoginTest.TotalFailures)
	}
	if len(userLoginTest.BranchCounts) != 2 {
		t.Errorf("Expected TestUserLogin to fail on 2 branches, got: %d", len(userLoginTest.BranchCounts))
	}

	// Verify TestPayment is detected as flaky (fails on main)
	paymentTest := findTestByName(flakyTests, "TestPayment")
	if paymentTest == nil {
		t.Fatal("Expected TestPayment to be detected as flaky")
	}
	if paymentTest.TotalFailures != 1 {
		t.Errorf("Expected TestPayment to have 1 failure, got: %d", paymentTest.TotalFailures)
	}
}

func TestParseTestFailures_InvalidJSON(t *testing.T) {
	invalidJSON := `{"invalid": json}`

	_, err := parseTestFailures(invalidJSON, "/tmp/test")

	if err == nil {
		t.Fatal("Expected error with invalid JSON")
	}
	if !contains(err.Error(), "failed to unmarshal Loki response") {
		t.Errorf("Expected JSON unmarshal error, got: %v", err)
	}
}

func TestDetectFlakyTestsFromRawEntries(t *testing.T) {
	tests := []struct {
		name     string
		entries  []RawLogEntry
		expected int
	}{
		{
			name: "flaky test on main branch",
			entries: []RawLogEntry{
				{TestName: "TestExample", Branch: "main", WorkflowRunURL: "https://example.com/1"},
			},
			expected: 1,
		},
		{
			name: "flaky test on multiple branches",
			entries: []RawLogEntry{
				{TestName: "TestExample", Branch: "main", WorkflowRunURL: "https://example.com/1"},
				{TestName: "TestExample", Branch: "feature", WorkflowRunURL: "https://example.com/2"},
			},
			expected: 1,
		},
		{
			name: "non-flaky test (only feature branch)",
			entries: []RawLogEntry{
				{TestName: "TestFeature", Branch: "feature", WorkflowRunURL: "https://example.com/1"},
			},
			expected: 0,
		},
		{
			name: "multiple flaky tests",
			entries: []RawLogEntry{
				{TestName: "TestA", Branch: "main", WorkflowRunURL: "https://example.com/1"},
				{TestName: "TestB", Branch: "feature", WorkflowRunURL: "https://example.com/2"},
				{TestName: "TestB", Branch: "develop", WorkflowRunURL: "https://example.com/3"},
			},
			expected: 2, // TestA (main) and TestB (multiple branches)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectFlakyTestsFromRawEntries(tt.entries, "/tmp/test")
			if len(result) != tt.expected {
				t.Errorf("Expected %d flaky tests, got %d", tt.expected, len(result))
			}
		})
	}
}

// Test edge cases

func TestAnalyzer_Run_MaxFailuresLimit(t *testing.T) {
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
	githubClient := &MockGitHubClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
	config := createTestConfig()
	config.MaxFailures = 3

	err := analyzer.Run(config)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	reportData, _ := fileSystem.writtenFiles["test-failure-analysis.json"]
	var result AnalysisResult
	json.Unmarshal(reportData, &result)

	if result.TestCount != 3 {
		t.Fatalf("Expected test count to be limited to 3, got: %d", result.TestCount)
	}
}

func TestAnalyzer_Run_NoProductionMode(t *testing.T) {
	logEntries := []RawLogEntry{
		{TestName: "TestUserLogin", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/1"},
	}

	lokiClient := &MockLokiClient{response: createTestLokiResponse(logEntries)}
	gitClient := &MockGitClient{
		testFiles: map[string]string{"TestUserLogin": "user_test.go"},
	}
	githubClient := &MockGitHubClient{}
	fileSystem := &MockFileSystem{}

	analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
	config := createTestConfig()
	config.DryRun = false // Production mode

	err := analyzer.Run(config)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify GitHub issue was created
	if len(githubClient.createdIssues) != 1 {
		t.Fatalf("Expected 1 GitHub issue to be created, got: %d", len(githubClient.createdIssues))
	}
	if len(githubClient.addedComments) != 1 {
		t.Fatalf("Expected 1 comment to be added, got: %d", len(githubClient.addedComments))
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
		setupMocks   func() (*MockGitClient, *MockGitHubClient)
		config       func() Config
	}{
		{
			name:         "complex_scenario",
			lokiFile:     "complex_loki_response.json",
			expectedFile: "complex_scenario.json",
			setupMocks: func() (*MockGitClient, *MockGitHubClient) {
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
				githubClient := &MockGitHubClient{
					usernames: map[string]string{
						"abc123def456": "alice",
						"789ghi012jkl": "bob",
						"345mno678pqr": "charlie",
						"901stu234vwx": "dave",
						"567yza890bcd": "eve",
					},
				}
				return gitClient, githubClient
			},
			config: func() Config {
				config := createTestConfig()
				config.MaxFailures = 10 // Don't limit for this test
				return config
			},
		},
		{
			name:         "empty_scenario",
			lokiFile:     "",
			expectedFile: "empty_scenario.json",
			setupMocks: func() (*MockGitClient, *MockGitHubClient) {
				return &MockGitClient{}, &MockGitHubClient{}
			},
			config: createTestConfig,
		},
		{
			name:         "single_test_scenario",
			lokiFile:     "",
			expectedFile: "single_test_scenario.json",
			setupMocks: func() (*MockGitClient, *MockGitHubClient) {
				gitClient := &MockGitClient{
					testFiles: map[string]string{
						"TestLoginFlow": "handlers/login_test.go",
					},
					authors: map[string][]CommitInfo{
						"TestLoginFlow": {}, // No recent commits
					},
				}
				return gitClient, &MockGitHubClient{}
			},
			config: createTestConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load Loki response
			var lokiResponse string
			if tt.lokiFile != "" {
				data, err := os.ReadFile(filepath.Join("testdata", tt.lokiFile))
				if err != nil {
					t.Fatalf("Failed to read Loki file %s: %v", tt.lokiFile, err)
				}
				lokiResponse = string(data)
			} else {
				// Create appropriate empty/single test response
				if tt.name == "single_test_scenario" {
					entries := []RawLogEntry{
						{TestName: "TestLoginFlow", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/401"},
						{TestName: "TestLoginFlow", Branch: "main", WorkflowRunURL: "https://github.com/test/repo/actions/runs/402"},
					}
					lokiResponse = createTestLokiResponse(entries)
				} else {
					lokiResponse = `{"status":"success","data":{"resultType":"streams","result":[]}}`
				}
			}

			// Setup mocks
			lokiClient := &MockLokiClient{response: lokiResponse}
			gitClient, githubClient := tt.setupMocks()
			fileSystem := &MockFileSystem{}

			// Run analysis
			analyzer := NewTestFailureAnalyzer(lokiClient, gitClient, githubClient, fileSystem)
			config := tt.config()

			err := analyzer.Run(config)
			if err != nil {
				t.Fatalf("Analysis failed: %v", err)
			}

			// Load expected result
			expectedData, err := os.ReadFile(filepath.Join("testdata", tt.expectedFile))
			if err != nil {
				t.Fatalf("Failed to read expected file %s: %v", tt.expectedFile, err)
			}

			var expected AnalysisResult
			if err := json.Unmarshal(expectedData, &expected); err != nil {
				t.Fatalf("Failed to unmarshal expected result: %v", err)
			}

			// Get actual result
			actualData, exists := fileSystem.writtenFiles["test-failure-analysis.json"]
			if !exists {
				t.Fatal("Expected report file to be written")
			}

			var actual AnalysisResult
			if err := json.Unmarshal(actualData, &actual); err != nil {
				t.Fatalf("Failed to unmarshal actual result: %v", err)
			}

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

			if string(actualJSON) != string(expectedJSON) {
				t.Errorf("Results don't match.\nExpected:\n%s\n\nActual:\n%s", expectedJSON, actualJSON)

				// Write actual result to file for debugging
				debugFile := filepath.Join("testdata", fmt.Sprintf("%s_actual.json", tt.name))
				os.WriteFile(debugFile, actualJSON, 0644)
				t.Logf("Actual result written to: %s", debugFile)
			}
		})
	}
}

func TestParseTestFailures_GoldenFile(t *testing.T) {
	// Test parsing with the complex Loki response
	data, err := os.ReadFile("testdata/complex_loki_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	flakyTests, err := parseTestFailures(string(data), "/tmp/test")
	if err != nil {
		t.Fatalf("Failed to parse test failures: %v", err)
	}

	// Verify expected results
	if len(flakyTests) != 3 {
		t.Fatalf("Expected 3 flaky tests, got %d", len(flakyTests))
	}

	// Verify specific test details
	dbTest := findTestByName(flakyTests, "TestDatabaseConnection")
	if dbTest == nil {
		t.Fatal("Expected TestDatabaseConnection to be found")
	}
	if dbTest.TotalFailures != 3 {
		t.Errorf("Expected TestDatabaseConnection to have 3 failures, got %d", dbTest.TotalFailures)
	}
	if len(dbTest.BranchCounts) != 2 {
		t.Errorf("Expected TestDatabaseConnection to fail on 2 branches, got %d", len(dbTest.BranchCounts))
	}

	authTest := findTestByName(flakyTests, "TestUserAuthentication")
	if authTest == nil {
		t.Fatal("Expected TestUserAuthentication to be found")
	}
	if authTest.TotalFailures != 2 {
		t.Errorf("Expected TestUserAuthentication to have 2 failures, got %d", authTest.TotalFailures)
	}

	paymentTest := findTestByName(flakyTests, "TestPaymentProcessing")
	if paymentTest == nil {
		t.Fatal("Expected TestPaymentProcessing to be found")
	}
	if paymentTest.TotalFailures != 2 {
		t.Errorf("Expected TestPaymentProcessing to have 2 failures, got %d", paymentTest.TotalFailures)
	}

	// TestNonFlaky should not be included (only failed on feature branch)
	nonFlakyTest := findTestByName(flakyTests, "TestNonFlaky")
	if nonFlakyTest != nil {
		t.Error("Expected TestNonFlaky to not be detected as flaky")
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

func normalizeTimestamps(actual, expected *AnalysisResult) {
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

func normalizeWorkflowOrder(actual, expected *AnalysisResult) {
	// Sort workflow URLs to make comparison order-independent
	for i, actualTest := range actual.FlakyTests {
		for j := range expected.FlakyTests {
			if expected.FlakyTests[j].TestName == actualTest.TestName {
				// Sort both arrays to ensure consistent order
				sort.Strings(actual.FlakyTests[i].ExampleWorkflows)
				sort.Strings(expected.FlakyTests[j].ExampleWorkflows)
				break
			}
		}
	}
}
