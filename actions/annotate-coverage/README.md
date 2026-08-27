# annotate-coverage

Highlights uncovered lines in a PR diff using Go coverage data. Parses Go
coverage files from a directory, intersects them with a git diff, and prints
the uncovered lines as text, Markdown, or GitHub Actions workflow commands
(`::notice file=...,line=...::...`) that GitHub Actions renders as PR
annotations on the changed lines.

<!-- x-release-please-start-version -->

```yaml
name: Coverage
on:
  pull_request:

permissions:
  contents: read

jobs:
  annotate-coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v6
        with:
          go-version: "1.25"

      - name: Run tests with coverage
        run: |
          mkdir -p .coverage
          go test ./... -coverprofile=.coverage/coverage.out

      - name: Annotate uncovered lines in PR diff
        uses: grafana/shared-workflows/actions/annotate-coverage@annotate-coverage/v0.2.1
        with:
          coverage-path: .coverage
          base-ref: ${{ github.event.pull_request.base.sha }}
          commit-sha: ${{ github.event.pull_request.head.sha }}
```

<!-- x-release-please-end-version -->

## Inputs

| Input                  | Description                                                                                                                                                | Required | Default                   |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| `coverage-path`        | Directory containing Go coverage files (`*.out`)                                                                                                           | No       | `.coverage`               |
| `format`               | Output format: `Text`, `Markdown`, or `GitHubAnnotations`                                                                                                  | No       | `GitHubAnnotations`       |
| `base-ref`             | Base ref to compare against (e.g., the PR base SHA). When set, diff is `<base-ref>..<commit-sha or HEAD>`.                                                 | No       | -                         |
| `commit-sha`           | Commit ref to compare to. With `base-ref`, diff is `<base-ref>..<commit-sha>`. Without `base-ref`, diff is the changes introduced by `<commit-sha>` alone. | No       | -                         |
| `repository-directory` | Path to the git repository to analyze                                                                                                                      | No       | `${{ github.workspace }}` |
| `go-version`           | Go version used to build the action binary                                                                                                                 | No       | `1.25`                    |

## Diff modes

- **PR diff (recommended in CI):** set `base-ref` to the PR base SHA and
  `commit-sha` to the head SHA. The diff is `<base-ref>..<commit-sha>`.
- **Branch vs HEAD:** set only `base-ref`. The diff is `<base-ref>..HEAD`.
- **Single commit:** set only `commit-sha`. The diff is the changes introduced
  by that commit.
- **Working tree:** leave both empty. The diff is `git diff` against the
  working tree (useful for local runs).

## Output formats

- `Text` — human-readable output for logs.
- `Markdown` — table-formatted output for PR comments or summaries.
- `GitHubAnnotations` — `::notice file=...,line=...::...` workflow commands
  that GitHub Actions renders as PR annotations on the changed lines. No
  GitHub API client is involved — annotations are emitted via workflow
  commands on stdout.

## Notes

- Coverage files are merged at the block level using the `gocovmerge`
  algorithm, so multiple `*.out` files in `coverage-path` are combined before
  analysis.
- Only Go files (`.go`) are considered. Binary files, deleted files, and
  non-Go files are skipped.
- Lines that are not in any coverage block (comments, blank lines, package
  declarations, etc.) are excluded from the uncovered count.
- For PR runs, check out the repository with `fetch-depth: 0` so the base ref
  is available locally.
