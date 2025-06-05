# Go Flaky Tests

A GitHub Action that detects and analyzes flaky Go tests by fetching logs from Loki and finding their authors.

## Features

- **Loki Integration**: Fetches test failure logs from Loki using LogQL queries
- **Flaky Test Detection**: Identifies tests that fail inconsistently across different branches
- **Git History Analysis**: Finds test files and extracts recent commit authors

## Usage

```yaml
name: Go Flaky Tests
on:
  schedule:
    - cron: "0 9 * * 1" # Run every Monday at 9 AM
  workflow_dispatch:

jobs:
  analyze-failures:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Go Flaky Tests
        uses: grafana/shared-workflows/actions/go-flaky-tests@main
        with:
          loki-url: ${{ secrets.LOKI_URL }}
          loki-username: ${{ secrets.LOKI_USERNAME }}
          loki-password: ${{ secrets.LOKI_PASSWORD }}
          repository: ${{ github.repository }}
          time-range: "7d"
          top-k: "5"
```

## Inputs

| Input                  | Description                                                                                                                  | Required | Default                   |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| `loki-url`             | Loki endpoint URL                                                                                                            | ✅       | -                         |
| `loki-username`        | Username for Loki authentication                                                                                             | ❌       | -                         |
| `loki-password`        | Password for Loki authentication. If using Grafana Cloud, then the access policy for this token needs the `logs:read` scope. | ❌       | -                         |
| `repository`           | Repository name in 'owner/repo' format                                                                                       | ✅       | -                         |
| `time-range`           | Time range for the query (e.g., '1h', '24h', '7d')                                                                           | ❌       | `1h`                      |
| `repository-directory` | Relative path to the directory with a git repository          | ❌       | `${{ github.workspace }}` |
| `top-k`                | Include only the top K flaky tests by distinct branches count                                                                | ❌       | `3`                       |

## Outputs

| Output             | Description                                     |
| ------------------ | ----------------------------------------------- |
| `test-count`       | Number of flaky tests found                     |
| `analysis-summary` | Summary of the analysis results                 |
| `report-path`      | Path to the generated analysis report JSON file |

## How It Works

1. **Fetch Logs**: Queries Loki for test failure logs within the specified time range
2. **Parse Failures**: Extracts test names, branches, and workflow URLs from logs
3. **Detect Flaky Tests**: Identifies tests that fail on multiple branches or multiple times on main/master
4. **Find Test Files**: Locates test files in the repository using grep
5. **Extract Authors**: Uses `git log -L` to find recent commits that modified each test

## Flaky Test Detection Logic

A test is considered "flaky" if:

- It fails on the main or master branch, OR
- It fails on multiple different branches

Tests that only fail on feature branches are not considered flaky, as they likely indicate legitimate test failures for that specific feature.

## Local Development

Run the analysis locally using the provided script:

```bash
# Set required environment variables
export LOKI_URL="your-loki-url"
export REPOSITORY="owner/repo"
export TIME_RANGE="24h"
export REPOSITORY_DIRECTORY="."

# Run the analysis
./run-local.sh
```

## Requirements

- Go 1.22 or later
- Git repository with test files
- Access to Loki instance with test failure logs

## Output Format

The action generates a JSON report with the following structure:

```json
{
  "test_count": 2,
  "analysis_summary": "Found 2 flaky tests. Most common tests: TestUserLogin (3 total failures; recently changed by alice), TestPayment (1 total failures; recently changed by bob)",
  "report_path": "/path/to/test-failure-analysis.json",
  "flaky_tests": [
    {
      "test_name": "TestUserLogin",
      "file_path": "handlers/auth_test.go",
      "total_failures": 3,
      "branch_counts": {
        "main": 2,
        "feature-branch": 1
      },
      "example_workflows": [
        "https://github.com/owner/repo/actions/runs/123",
        "https://github.com/owner/repo/actions/runs/124"
      ],
      "recent_commits": [
        {
          "hash": "abc123",
          "author": "alice",
          "timestamp": "2024-01-15T10:30:00Z",
          "title": "Fix authentication flow"
        }
      ]
    }
  ]
}
```
