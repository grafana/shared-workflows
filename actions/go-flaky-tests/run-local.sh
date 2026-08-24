#!/bin/bash

# Local development script for go-flaky-tests action
# This script allows you to run the action locally for testing and development

set -e

# Default values
DEFAULT_TIME_RANGE="24h"
DEFAULT_TOP_K="3"
DEFAULT_SKIP_POSTING_ISSUES="true"

# Function to print usage
usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help                    Show this help message"
    echo "  --loki-url URL                Loki endpoint URL (required)"
    echo "  --loki-username USER          Username for Loki authentication"
    echo "  --loki-password PASS          Password for Loki authentication"
    echo "  --repository REPO             Repository name in 'owner/repo' format (required)"
    echo "  --time-range RANGE            Time range for query (default: ${DEFAULT_TIME_RANGE})"
    echo "  --github-token TOKEN          GitHub token for API access"
    echo "  --repository-directory DIR    Repository directory (default: current directory)"
    echo "  --skip-posting-issues BOOL    Skip creating GitHub issues (default: ${DEFAULT_SKIP_POSTING_ISSUES})"
    echo "  --top-k NUM                   Number of top flaky tests to analyze (default: ${DEFAULT_TOP_K})"
    echo "  --ignored-tests TESTS         Comma-delimited test names to skip failures for"
    echo ""
    echo "Environment variables:"
    echo "  LOKI_URL, LOKI_USERNAME, LOKI_PASSWORD, REPOSITORY, TIME_RANGE,"
    echo "  REPOSITORY_DIRECTORY, GITHUB_TOKEN, SKIP_POSTING_ISSUES, TOP_K, IGNORED_TESTS"
    echo ""
    echo "Example:"
    echo "  $0 --loki-url http://localhost:3100 --repository myorg/myrepo --time-range 7d"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        --loki-url)
            LOKI_URL="$2"
            shift 2
            ;;
        --loki-username)
            LOKI_USERNAME="$2"
            shift 2
            ;;
        --loki-password)
            LOKI_PASSWORD="$2"
            shift 2
            ;;
        --repository)
            REPOSITORY="$2"
            shift 2
            ;;
        --time-range)
            TIME_RANGE="$2"
            shift 2
            ;;
        --github-token)
            GITHUB_TOKEN="$2"
            shift 2
            ;;
        --repository-directory)
            REPOSITORY_DIRECTORY="$2"
            shift 2
            ;;
        --skip-posting-issues)
            SKIP_POSTING_ISSUES="$2"
            shift 2
            ;;
        --top-k)
            TOP_K="$2"
            shift 2
            ;;
        --ignored-tests)
            IGNORED_TESTS="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Set defaults for optional parameters
TIME_RANGE="${TIME_RANGE:-$DEFAULT_TIME_RANGE}"
TOP_K="${TOP_K:-$DEFAULT_TOP_K}"
SKIP_POSTING_ISSUES="${SKIP_POSTING_ISSUES:-$DEFAULT_SKIP_POSTING_ISSUES}"
REPOSITORY_DIRECTORY="${REPOSITORY_DIRECTORY:-$(pwd)}"

# Validate required parameters
if [[ -z "$LOKI_URL" ]]; then
    echo "Error: --loki-url is required"
    usage
fi

if [[ -z "$REPOSITORY" ]]; then
    echo "Error: --repository is required"
    usage
fi

# Export environment variables
export LOKI_URL
export LOKI_USERNAME
export LOKI_PASSWORD
export REPOSITORY
export TIME_RANGE
export GITHUB_TOKEN
export REPOSITORY_DIRECTORY
export SKIP_POSTING_ISSUES
export TOP_K
export IGNORED_TESTS

echo "🔧 Running go-flaky-tests locally..."
echo "📊 Repository: $REPOSITORY"
echo "⏰ Time range: $TIME_RANGE"
echo "📁 Repository directory: $REPOSITORY_DIRECTORY"
echo "🔝 Top K tests: $TOP_K"
echo "🏃 Dry run mode: $SKIP_POSTING_ISSUES"
echo ""

# Build the application
echo "🔨 Building application..."
go build -o analyzer ./cmd/go-flaky-tests

# Run the application
echo "🚀 Running analysis..."
./analyzer

echo "✅ Analysis complete!"

# Clean up
rm -f analyzer
