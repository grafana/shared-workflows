package main

import (
	"fmt"
	"log"
	"slices"
	"strings"
)

type RawLogEntry struct {
	TestName       string `json:"test_name"`
	Branch         string `json:"branch"`
	WorkflowRunURL string `json:"workflow_run_url"`
}

func AggregateTestFailuresFromResponse(lokiResp *LokiResponse) ([]FlakyTest, error) {
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

	return detectFlakyTestsFromRawEntries(rawEntries), nil
}

func detectFlakyTestsFromRawEntries(rawEntries []RawLogEntry) []FlakyTest {
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
