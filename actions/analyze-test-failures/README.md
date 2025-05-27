# Analyze Test Failures

This action fetches logs from Loki using LogQL queries, analyzes test failures, and identifies the authors of failing tests using git blame.

## Inputs

| Input                  | Description                                                               | Required | Default                   |
| ---------------------- | ------------------------------------------------------------------------- | -------- | ------------------------- |
| `loki-url`             | Loki endpoint URL                                                         | Yes      |                           |
| `loki-username`        | Username for Loki authentication                                          | No       |                           |
| `loki-password`        | Password for Loki authentication                                          | No       |                           |
| `repository`           | Repository name in 'owner/repo' format (e.g., 'grafana/grafana')          | Yes      |                           |
| `time-range`           | Time range for the query (e.g., '1h', '24h', '7d')                        | No       | `1h`                      |
| `github-token`         | GitHub token for repository access                                        | No       | `${{ github.token }}`     |
| `repository-directory` | Repository directory to analyze                                           | No       | `${{ github.workspace }}` |
| `skip-posting-issues`  | Skip creating/updating GitHub issues (dry-run mode)                       | No       | `true`                    |
| `top-k`                | Include only the top K flaky tests by distinct branches count in analysis | No       | `3`                       |

## Outputs

| Output             | Description                                       |
| ------------------ | ------------------------------------------------- |
| `failure-count`    | Number of test failures found                     |
| `affected-authors` | JSON array of authors who wrote the failing tests |
| `analysis-summary` | Summary of the analysis results                   |
| `report-path`      | Path to the generated analysis report             |

## Example Workflow

This action is designed to be run periodically via a workflow. Example workflow:

```yaml
name: Test Failure Analysis
on:
  schedule:
    - cron: "0 */6 * * *" # Every 6 hours
  workflow_dispatch:

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Analyze recent test failures
        uses: grafana/shared-workflows/actions/analyze-test-failures@main
        with:
          loki-url: ${{ vars.LOKI_URL }}
          loki-username: ${{ secrets.LOKI_USERNAME }}
          loki-password: ${{ secrets.LOKI_PASSWORD }}
          repository: ${{ github.repository }}
          time-range: "6h"
```

## Report Format

The action generates a JSON report with the following structure:

```json
{
  "failure_count": 3,
  "affected_authors": ["user1@example.com", "user2@example.com"],
  "analysis_summary": "Found 3 flaky test failures affecting 2 authors. Most common failures: TestUserLogin (8 failures), TestDatabaseConnection (12 failures)",
  "report_path": "/path/to/test-failure-analysis.json",
  "test_failures": [
    {
      "test_name": "TestUserLogin",
      "file_path": "user_test.go",
      "timestamp": "1640995200000000000",
      "message": "Flaky test TestUserLogin failed 8 times across branches: main:3, feature-branch:5",
      "branch_counts": {
        "main": 3,
        "feature-branch": 5
      },
      "is_flaky": true,
      "example_prs": [
        "https://github.com/owner/repo/pull/123",
        "https://github.com/owner/repo/pull/124",
        "https://github.com/owner/repo/pull/125"
      ]
    },
    {
      "test_name": "TestDatabaseConnection",
      "file_path": "database_test.go",
      "timestamp": "1640995300000000000",
      "message": "Flaky test TestDatabaseConnection failed 12 times across branches: main:2, dev:4, staging:6",
      "branch_counts": {
        "main": 2,
        "dev": 4,
        "staging": 6
      },
      "is_flaky": true,
      "example_prs": [
        "https://github.com/owner/repo/pull/456",
        "https://github.com/owner/repo/pull/457"
      ]
    }
  ]
}
```

## LogQL Query

The action uses a predefined LogQL query that analyzes test failures from CI/CD logs:

```logql
{service_name="$repo", service_namespace="cicd-o11y"}
|= "--- FAIL: Test"
| json
| __error__=""
| resources_ci_github_workflow_run_conclusion!="cancelled"
| line_format "{{.body}}"
| regexp "--- FAIL: (?P<test_name>.*) \\(\\d"
| line_format "{{.test_name}}"
| regexp `(?P<parent_test_name>Test[a-z0-9A-Z_]+)`
```

This query:

- Filters logs from the specified repository (in owner-repo format, **not** owner/repo format) in the `cicd-o11y` namespace
- Looks for Go test failures (`--- FAIL: Test`)
- Parses JSON logs and filters out error logs
- Excludes cancelled workflow runs using `resources_ci_github_workflow_run_conclusion`. A failure during a cancelled workflow may not necessarily indicate a flaky test.
- Extracts test content from the `body` field
- Extracts test names using regex patterns
- Returns raw log entries for analysis in the GitHub action

## Flaky Test Detection

The action identifies tests as "flaky" based on the following criteria:

- **Failed on main branch**: Any test that has failed on the `main` branch is considered flaky
- **Failed on multiple branches**: Any test that has failed on more than one branch is considered flaky

Only tests meeting these criteria are included in the analysis results.

## GitHub Issue Management

The action automatically creates and maintains GitHub issues for detected flaky tests:

- **Creates new issues**: For newly detected flaky tests with detailed investigation guides
- **Updates existing issues**: Adds new failure data and mentions recent contributors
- **Reopens closed issues**: When previously resolved flaky tests fail again
- **Author mentions**: Uses git blame to identify recent contributors to the tests and mentions them in issues
- **Dry-run mode**: Set `skip-posting-issues: false` to enable issue creation (disabled by default)

### Issue Content

Each issue includes:

- Test name and failure summary
- Links to example PRs where the test failed
- Recent commit authors who modified the test
- Investigation steps and debugging guidance
- Failure count and branch breakdown

## Example PR Links

For each flaky test, the action captures up to 3 example Pull Request URLs where the test failed. This provides concrete examples of when and where the test exhibited flaky behavior, making it easier to investigate the root cause.

## Local Development

You can run this action locally for testing and development:

### Prerequisites

1. **Go 1.22+** installed
2. **GitHub CLI (gh)** installed and authenticated
3. Access to a Loki instance
4. GitHub token with appropriate permissions

### Setup

1. **Clone the repository and navigate to the action directory:**

   ```bash
   git clone https://github.com/grafana/shared-workflows.git
   cd shared-workflows/actions/analyze-test-failures
   ```

2. **Create environment configuration:**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Configure your .env file:**

   ```bash
   # Required
   LOKI_URL=https://your-loki-instance.com
   LOKI_USERNAME=your_username
   LOKI_PASSWORD=your_password
   REPOSITORY=owner/repo-name
   GITHUB_TOKEN=ghp_your_github_token_here

   # Optional
   TIME_RANGE=24h
   REPOSITORY_DIRECTORY=.
   ```

### Running

Execute the local run script:

```bash
./run-local.sh
```

This will:

- Validate your configuration
- Run the Go application with `go run`
- Execute analysis against your Loki instance
- Generate a local `test-failure-analysis.json` report
- Display results to stdout

### Example Output

```
Running test failure analysis...
Repository: my-repo
Time range: 24h
Loki URL: https://logs.grafana.com

::set-output name=failure-count::3
::set-output name=affected-authors::["user1","user2","user3"]
::set-output name=analysis-summary::Found 3 flaky test failures affecting 3 authors. Most common failures: TestUserLogin (8 failures), TestAuth (5 failures)
::set-output name=report-path::/path/to/test-failure-analysis.json

Analysis complete! Check the generated report and outputs above.
```

### Troubleshooting

- **"gh CLI not found"**: Install from https://cli.github.com/
- **"gh not authenticated"**: Run `gh auth login`
- **"Loki authentication failed"**: Check your LOKI_USERNAME/LOKI_PASSWORD
- **"No test failures found"**: Verify your repository name and time range
- **"Test function not found"**: Make sure you're running from a repository root with Go test files
