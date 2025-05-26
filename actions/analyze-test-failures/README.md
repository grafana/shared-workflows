# Analyze Test Failures

This action fetches logs from Loki using LogQL queries, analyzes test failures, and identifies the authors of failing tests using git blame.

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `loki-url` | Loki endpoint URL | Yes | |
| `loki-username` | Username for Loki authentication | No | |
| `loki-password` | Password for Loki authentication | No | |
| `repository` | Repository name to analyze test failures for | Yes | |
| `time-range` | Time range for the query (e.g., '1h', '24h', '7d') | No | `1h` |
| `github-token` | GitHub token for repository access | No | `${{ github.token }}` |
| `working-directory` | Working directory to analyze | No | `${{ github.workspace }}` |

## Outputs

| Output | Description |
|--------|-------------|
| `failure-count` | Number of test failures found |
| `affected-authors` | JSON array of authors who wrote the failing tests |
| `analysis-summary` | Summary of the analysis results |
| `report-path` | Path to the generated analysis report |

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
  "analysis_summary": "Found 3 flaky test failures affecting 2 authors...",
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
      "is_flaky": true
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
      "is_flaky": true
    }
  ]
}
```

## LogQL Query

The action uses a predefined LogQL query that analyzes test failures from CI/CD logs:

```logql
sum by (parent_test_name, resources_ci_github_workflow_run_head_branch) (
    count_over_time({service_name="$repo", service_namespace="cicd-o11y"} 
    |= "--- FAIL: Test" 
    | json 
    | __error__="" 
    | resources_ci_github_workflow_run_conclusion!="cancelled" 
    | line_format "{{.body}}" 
    | regexp "--- FAIL: (?P<test_name>.*) \\(\\d" 
    | line_format "{{.test_name}}" 
    | regexp `(?P<parent_test_name>Test[a-z0-9A-Z_]+)`[7d])
)
```

This query:
- Filters logs from the specified repository in the `cicd-o11y` namespace
- Looks for Go test failures (`--- FAIL: Test`)
- Parses JSON logs and filters out error logs
- Excludes cancelled workflow runs using `resources_ci_github_workflow_run_conclusion`
- Extracts test content from the `body` field
- Groups results by both test name AND branch
- Counts failures over the last 7 days

## Flaky Test Detection

The action identifies tests as "flaky" based on the following criteria:
- **Failed on main branch**: Any test that has failed on the `main` branch is considered flaky
- **Failed on multiple branches**: Any test that has failed on more than one branch is considered flaky

Only tests meeting these criteria are included in the analysis results.

## Workflow Integration

This action is designed to be run periodically via a workflow. Example workflow:

```yaml
name: Test Failure Analysis
on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours
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