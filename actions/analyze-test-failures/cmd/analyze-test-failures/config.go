package main

import (
	"os"
	"strconv"
)

type Config struct {
	LokiURL             string
	LokiUsername        string
	LokiPassword        string
	Repository          string
	TimeRange           string
	GitHubToken         string
	RepositoryDirectory string
	DryRun              bool
	MaxFailures         int
}

func getConfigFromEnv() Config {
	return Config{
		LokiURL:             os.Getenv("LOKI_URL"),
		LokiUsername:        os.Getenv("LOKI_USERNAME"),
		LokiPassword:        os.Getenv("LOKI_PASSWORD"),
		Repository:          os.Getenv("REPOSITORY"),
		TimeRange:           getEnvWithDefault("TIME_RANGE", "24h"),
		GitHubToken:         os.Getenv("GITHUB_TOKEN"),
		RepositoryDirectory: getEnvWithDefault("REPOSITORY_DIRECTORY", "."),
		DryRun:              getBoolEnvWithDefault("DRY_RUN", true),
		MaxFailures:         getIntEnvWithDefault("MAX_FAILURES", 3),
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
