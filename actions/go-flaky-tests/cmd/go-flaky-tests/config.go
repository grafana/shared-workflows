package main

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	LokiURL             string
	LokiUsername        string
	LokiPassword        string
	Repository          string
	TimeRange           string
	RepositoryDirectory string
	SkipPostingIssues   bool
	TopK                int
	IgnoredTests        []string
}

func getConfigFromEnv() Config {
	return Config{
		LokiURL:             os.Getenv("LOKI_URL"),
		LokiUsername:        os.Getenv("LOKI_USERNAME"),
		LokiPassword:        os.Getenv("LOKI_PASSWORD"),
		Repository:          os.Getenv("REPOSITORY"),
		TimeRange:           getEnvWithDefault("TIME_RANGE", "24h"),
		RepositoryDirectory: getEnvWithDefault("REPOSITORY_DIRECTORY", "."),
		SkipPostingIssues:   getBoolEnvWithDefault("SKIP_POSTING_ISSUES", true),
		TopK:                getIntEnvWithDefault("TOP_K", 3),
		IgnoredTests:        getStringSliceFromEnv("IGNORED_TESTS"),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func getIntEnvWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getStringSliceFromEnv(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{}
	}
	
	var result []string
	for _, item := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
