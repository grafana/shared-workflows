# Analyze Test Failures

This action fetches logs from Loki using LogQL queries, analyzes test failures, and identifies the authors of failing tests using git blame.

## Inputs

| Input               | Description                                        | Required | Default                   |
| ------------------- | -------------------------------------------------- | -------- | ------------------------- |
| `loki-url`          | Loki endpoint URL                                  | Yes      |                           |
| `loki-username`     | Username for Loki authentication                   | No       |                           |
| `loki-password`     | Password for Loki authentication                   | No       |                           |
| `repository`        | Repository name to analyze test failures for       | Yes      |                           |
| `time-range`        | Time range for the query (e.g., '1h', '24h', '7d') | No       | `1h`                      |
| `github-token`      | GitHub token for repository access                 | No       | `${{ github.token }}`     |
| `working-directory` | Working directory to analyze                       | No       | `${{ github.workspace }}` |

## Outputs

| Output             | Description                                       |
| ------------------ | ------------------------------------------------- |
| `failure-count`    | Number of test failures found                     |
| `affected-authors` | JSON array of authors who wrote the failing tests |
| `analysis-summary` | Summary of the analysis results                   |
| `report-path`      | Path to the generated analysis report             |

## Example Usage

```yaml
- name: Analyze test failures
  uses: grafana/shared-workflows/actions/analyze-test-failures@main
  with:
    loki-url: https://logs.grafana.com
    loki-username: ${{ secrets.LOKI_USERNAME }}
    loki-password: ${{ secrets.LOKI_PASSWORD }}
    repository: "my-repo"
    time-range: "24h"
  id: analyze

- name: Comment on PR with results
  if: steps.analyze.outputs.failure-count > 0
  uses: actions/github-script@v7
  with:
    script: |
      const failureCount = '${{ steps.analyze.outputs.failure-count }}';
      const authors = JSON.parse('${{ steps.analyze.outputs.affected-authors }}');
      const summary = '${{ steps.analyze.outputs.analysis-summary }}';

      github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: `## Test Failure Analysis
        
        Found ${failureCount} test failures.
        
        **Affected authors:** ${authors.join(', ')}
        
        **Summary:** ${summary}
        
        Full report: ${{ steps.analyze.outputs.report-path }}`
      });
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

- Filters logs from the specified repository in the `cicd-o11y` namespace
- Looks for Go test failures (`--- FAIL: Test`)
- Parses JSON logs and filters out error logs
- Excludes cancelled workflow runs using `resources_ci_github_workflow_run_conclusion`
- Extracts test content from the `body` field
- Extracts test names using regex patterns
- Returns raw log entries for analysis in the application

## Flaky Test Detection

The action identifies tests as "flaky" based on the following criteria:

- **Failed on main branch**: Any test that has failed on the `main` branch is considered flaky
- **Failed on multiple branches**: Any test that has failed on more than one branch is considered flaky

Only tests meeting these criteria are included in the analysis results.

## Example PR Links

For each flaky test, the action captures up to 3 example Pull Request URLs where the test failed. This provides concrete examples of when and where the test exhibited flaky behavior, making it easier to investigate the root cause.

## Workflow Integration

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
          repository: ${{ github.event.repository.name }}
          time-range: "6h"
```

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
   REPOSITORY=your-repo-name
   GITHUB_TOKEN=ghp_your_github_token_here

   # Optional
   TIME_RANGE=24h
   WORKING_DIRECTORY=.
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
