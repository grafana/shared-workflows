package main

import (
	"testing"
)

func TestAggregateFlakyTestsFromResponse(t *testing.T) {
	tests := []struct {
		name        string
		rawEntries  []RawLogEntry
		expected    []FlakyTest
		description string
	}{
		{
			name: "test fails on main branch",
			rawEntries: []RawLogEntry{
				{TestName: "TestExample", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/123"},
				{TestName: "TestExample", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/124"},
			},
			expected: []FlakyTest{
				{
					TestName:      "TestExample",
					TotalFailures: 2,
					BranchCounts:  map[string]int{"main": 2},
					ExampleWorkflows: []GithubActionsWorkflow{
						{RunURL: "https://github.com/owner/repo/actions/runs/123"},
						{RunURL: "https://github.com/owner/repo/actions/runs/124"},
					},
				},
			},
			description: "Test that fails on main branch should be classified as flaky",
		},
		{
			name: "test fails on multiple branches",
			rawEntries: []RawLogEntry{
				{TestName: "TestMultiBranch", Branch: "feature-1", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/200"},
				{TestName: "TestMultiBranch", Branch: "feature-2", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/201"},
				{TestName: "TestMultiBranch", Branch: "feature-1", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/202"},
			},
			expected: []FlakyTest{
				{
					TestName:      "TestMultiBranch",
					TotalFailures: 3,
					BranchCounts:  map[string]int{"feature-1": 2, "feature-2": 1},
					ExampleWorkflows: []GithubActionsWorkflow{
						{RunURL: "https://github.com/owner/repo/actions/runs/200"},
						{RunURL: "https://github.com/owner/repo/actions/runs/201"},
						{RunURL: "https://github.com/owner/repo/actions/runs/202"},
					},
				},
			},
			description: "Test that fails on multiple branches should be classified as flaky",
		},
		{
			name: "test fails only on single non-main branch",
			rawEntries: []RawLogEntry{
				{TestName: "TestSingleBranch", Branch: "feature-only", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/300"},
				{TestName: "TestSingleBranch", Branch: "feature-only", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/301"},
			},
			expected:    []FlakyTest{},
			description: "Test that fails only on a single non-main branch should NOT be classified as flaky",
		},
		{
			name: "mixed scenario with flaky and non-flaky tests",
			rawEntries: []RawLogEntry{
				// Flaky test - fails on main
				{TestName: "TestFlaky1", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/400"},
				// Flaky test - fails on multiple branches
				{TestName: "TestFlaky2", Branch: "feature-1", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/401"},
				{TestName: "TestFlaky2", Branch: "feature-2", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/402"},
				// Non-flaky test - single branch only
				{TestName: "TestNotFlaky", Branch: "feature-only", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/403"},
			},
			expected: []FlakyTest{
				{
					TestName:      "TestFlaky2",
					TotalFailures: 2,
					BranchCounts:  map[string]int{"feature-1": 1, "feature-2": 1},
					ExampleWorkflows: []GithubActionsWorkflow{
						{RunURL: "https://github.com/owner/repo/actions/runs/401"},
						{RunURL: "https://github.com/owner/repo/actions/runs/402"},
					},
				},
				{
					TestName:      "TestFlaky1",
					TotalFailures: 1,
					BranchCounts:  map[string]int{"main": 1},
					ExampleWorkflows: []GithubActionsWorkflow{
						{RunURL: "https://github.com/owner/repo/actions/runs/400"},
					},
				},
			},
			description: "Should correctly identify flaky tests and ignore non-flaky ones",
		},
		{
			name: "limits example workflows to 3",
			rawEntries: []RawLogEntry{
				{TestName: "TestManyWorkflows", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/500"},
				{TestName: "TestManyWorkflows", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/501"},
				{TestName: "TestManyWorkflows", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/502"},
				{TestName: "TestManyWorkflows", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/503"},
				{TestName: "TestManyWorkflows", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/504"},
			},
			expected: []FlakyTest{
				{
					TestName:      "TestManyWorkflows",
					TotalFailures: 5,
					BranchCounts:  map[string]int{"main": 5},
					ExampleWorkflows: []GithubActionsWorkflow{
						{RunURL: "https://github.com/owner/repo/actions/runs/500"},
						{RunURL: "https://github.com/owner/repo/actions/runs/501"},
						{RunURL: "https://github.com/owner/repo/actions/runs/502"},
					},
				},
			},
			description: "Should limit example workflows to maximum of 3",
		},
		{
			name:        "empty input",
			rawEntries:  []RawLogEntry{},
			expected:    []FlakyTest{},
			description: "Should handle empty input gracefully",
		},
		{
			name: "entries with missing data",
			rawEntries: []RawLogEntry{
				{TestName: "", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/600"},
				{TestName: "TestValid", Branch: "", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/601"},
				{TestName: "TestValidEntry", Branch: "main", WorkflowRunURL: "https://github.com/owner/repo/actions/runs/602"},
			},
			expected: []FlakyTest{
				{
					TestName:      "TestValidEntry",
					TotalFailures: 1,
					BranchCounts:  map[string]int{"main": 1},
					ExampleWorkflows: []GithubActionsWorkflow{
						{RunURL: "https://github.com/owner/repo/actions/runs/602"},
					},
				},
			},
			description: "Should skip entries with missing test name or branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectFlakyTestsFromRawEntries(tt.rawEntries)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d flaky tests, got %d", len(tt.expected), len(result))
				return
			}

			// Convert to map for easier comparison since order might vary
			resultMap := make(map[string]FlakyTest)
			for _, test := range result {
				resultMap[test.TestName] = test
			}

			for _, expected := range tt.expected {
				actual, found := resultMap[expected.TestName]
				if !found {
					t.Errorf("Expected test %s not found in results", expected.TestName)
					continue
				}

				if actual.TotalFailures != expected.TotalFailures {
					t.Errorf("Test %s: expected %d total failures, got %d",
						expected.TestName, expected.TotalFailures, actual.TotalFailures)
				}

				if len(actual.BranchCounts) != len(expected.BranchCounts) {
					t.Errorf("Test %s: expected %d branches, got %d",
						expected.TestName, len(expected.BranchCounts), len(actual.BranchCounts))
				}

				for branch, expectedCount := range expected.BranchCounts {
					if actualCount, exists := actual.BranchCounts[branch]; !exists || actualCount != expectedCount {
						t.Errorf("Test %s: expected branch %s to have %d failures, got %d",
							expected.TestName, branch, expectedCount, actualCount)
					}
				}

				if len(actual.ExampleWorkflows) != len(expected.ExampleWorkflows) {
					t.Errorf("Test %s: expected %d example workflows, got %d",
						expected.TestName, len(expected.ExampleWorkflows), len(actual.ExampleWorkflows))
				}
			}
		})
	}
}

func TestSortFlakyTests(t *testing.T) {
	tests := []FlakyTest{
		{
			TestName:     "TestSingleBranch",
			BranchCounts: map[string]int{"main": 5},
		},
		{
			TestName:     "TestZAlphabeticallyLast",
			BranchCounts: map[string]int{"main": 3, "feature": 2},
		},
		{
			TestName:     "TestAAlphabeticallyFirst",
			BranchCounts: map[string]int{"main": 1, "feature": 1, "dev": 1},
		},
		{
			TestName:     "TestBSecondAlphabetically",
			BranchCounts: map[string]int{"main": 2, "feature": 1},
		},
	}

	result := sortFlakyTests(tests)

	// Should be sorted by number of branches (descending), then alphabetically by name
	expected := []string{
		"TestAAlphabeticallyFirst",  // 3 branches
		"TestBSecondAlphabetically", // 2 branches
		"TestZAlphabeticallyLast",   // 2 branches
		"TestSingleBranch",          // 1 branch
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d tests, got %d", len(expected), len(result))
		return
	}

	for i, expectedName := range expected {
		if result[i].TestName != expectedName {
			t.Errorf("Position %d: expected %s, got %s", i, expectedName, result[i].TestName)
		}
	}
}

func TestParseTestFailuresFromResponse(t *testing.T) {
	lokiResp := &LokiResponse{
		Status: "success",
		Data: LokiData{
			ResultType: "streams",
			Result: []LokiResult{
				{
					Stream: map[string]string{
						"parent_test_name":                   "TestExample",
						"ci_github_workflow_run_head_branch": "main",
						"ci_github_workflow_run_html_url":    "https://github.com/owner/repo/actions/runs/123",
					},
					Values: [][]string{{"1640995200000000000", "log line"}},
				},
				{
					Stream: map[string]string{
						"parent_test_name":                   "TestAnother",
						"ci_github_workflow_run_head_branch": "feature",
						"ci_github_workflow_run_html_url":    "https://github.com/owner/repo/actions/runs/124",
					},
					Values: [][]string{{"1640995300000000000", "another log line"}},
				},
				{
					Stream: map[string]string{
						"parent_test_name":                   "", // Missing test name
						"ci_github_workflow_run_head_branch": "main",
					},
					Values: [][]string{{"1640995400000000000", "invalid log line"}},
				},
			},
		},
	}

	result, err := AggregateFlakyTestsFromResponse(lokiResp)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Should only return the valid test that fails on main branch
	if len(result) != 1 {
		t.Errorf("Expected 1 flaky test, got %d", len(result))
		return
	}

	if result[0].TestName != "TestExample" {
		t.Errorf("Expected TestExample, got %s", result[0].TestName)
	}

	if result[0].TotalFailures != 1 {
		t.Errorf("Expected 1 total failure, got %d", result[0].TotalFailures)
	}

	if result[0].BranchCounts["main"] != 1 {
		t.Errorf("Expected 1 failure on main branch, got %d", result[0].BranchCounts["main"])
	}
}
