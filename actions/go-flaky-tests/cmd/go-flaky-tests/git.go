package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)


type DefaultGitClient struct {
	config Config
}

func NewDefaultGitClient(config Config) *DefaultGitClient {
	return &DefaultGitClient{config: config}
}

func (g *DefaultGitClient) FindTestFile(testName string) (string, error) {
	return findTestFilePath(g.config.RepositoryDirectory, testName)
}

func (g *DefaultGitClient) GetFileAuthors(filePath, testName string) ([]CommitInfo, error) {
	return getFileAuthors(g.config, filePath, testName)
}

func findTestFilePath(repoDir, testName string) (string, error) {
	if !strings.HasPrefix(testName, "Test") {
		return "", fmt.Errorf("invalid test name format: %s", testName)
	}

	grepCmd := exec.Command("grep", "-rl", "--include=*_test.go", fmt.Sprintf("func %s(", testName), ".")
	grepCmd.Dir = repoDir

	result, err := grepCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to search for test function %s: %w", testName, err)
	}

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		if len(lines) > 1 {
			log.Printf("Warning: test function %s found in multiple files, using first match: %s", testName, lines[0])
		}

		filePath := strings.TrimPrefix(lines[0], "./")
		return filePath, nil
	}

	return "", fmt.Errorf("test function %s not found in repository", testName)
}

func guessTestFilePath(testName string) string {
	if strings.HasPrefix(testName, "Test") {
		name := strings.TrimPrefix(testName, "Test")
		if name != "" {
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

func getFileAuthors(config Config, filePath, testName string) ([]CommitInfo, error) {
	return getFileAuthorsWithClient(config.RepositoryDirectory, filePath, testName)
}

func getFileAuthorsWithClient(repoDir, filePath, testName string) ([]CommitInfo, error) {
	cmd := exec.Command("git", "log", "-3", "-L", fmt.Sprintf(":%s:%s", testName, filePath), "--pretty=format:%H|%ct|%s", "-s")
	cmd.Dir = repoDir

	result, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: failed to get git log for test %s in %s: %v", testName, filePath, err)
		return []CommitInfo{}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		log.Printf("Warning: no git log results for test %s in %s", testName, filePath)
		return []CommitInfo{}, nil
	}

	var commits []CommitInfo
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	for _, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid git log format for test %s in %s: %s", testName, filePath, line)
		}

		hash := parts[0]
		timestampStr := parts[1]
		title := parts[2]

		if hash == "" {
			continue
		}

		var timestamp time.Time
		if timestampUnix, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
			timestamp = time.Unix(timestampUnix, 0)
		}

		if timestamp.Before(sixMonthsAgo) {
			continue
		}

		username := "unknown" // Will be resolved via GitHub in PR3

		if strings.HasSuffix(username, "[bot]") {
			continue
		}

		commitInfo := CommitInfo{
			Hash:      hash,
			Author:    username,
			Timestamp: timestamp,
			Title:     title,
		}
		commits = append(commits, commitInfo)
	}

	return commits, nil
}
