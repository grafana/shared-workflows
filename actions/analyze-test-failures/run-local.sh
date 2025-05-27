#!/bin/bash

# Script to run the analyze-test-failures action locally for development/testing

set -e

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    echo "Please copy .env.example to .env and fill in your values:"
    echo "  cp .env.example .env"
    echo "  # Edit .env with your configuration"
    exit 1
fi

# Load environment variables from .env file
set -a
# shellcheck disable=SC1091 # We check if the file exists above and fail the script if it doesn't
source .env
set +a

# Check required environment variables
required_vars=("LOKI_URL" "LOKI_USERNAME" "LOKI_PASSWORD" "REPOSITORY" "GITHUB_TOKEN")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "Error: Required environment variable $var is not set in .env file"
        exit 1
    fi
done

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is required but not installed."
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Run the analyzer
echo "Running test failure analysis..."
echo "Repository: $REPOSITORY"
echo "Time range: ${TIME_RANGE:-1h}"
echo "Loki URL: $LOKI_URL"
echo ""

cd cmd/analyze-test-failures
go run .
cd ../..

echo ""
echo "Analysis complete! Check the generated report and outputs above."
