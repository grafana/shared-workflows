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
        uses: grafana/shared-workflows/actions/annotate-coverage@annotate-coverage/v0.2.0
        with:
          coverage-path: .coverage
          base-ref: ${{ github.event.pull_request.base.sha }}
          commit-sha: ${{ github.event.pull_request.head.sha }}
```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name                   | Type   | Required | Default                   | Description                                                                                                                                                                                     |
| ---------------------- | ------ | -------- | ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `base-ref`             | String | No       |                           | Base ref to compare against (e.g., the PR base branch). When set, diff is computed as &lt;base-ref>..&lt;commit-sha or HEAD>.                                                                   |
| `commit-sha`           | String | No       |                           | Commit ref to compare to. With base-ref, diff is &lt;base-ref>..&lt;commit-sha>. Without base-ref, diff is the changes introduced by &lt;commit-sha>. Defaults to HEAD when used with base-ref. |
| `coverage-path`        | String | No       | `.coverage`               | Directory containing Go coverage files (*.out)                                                                                                                                                  |
| `format`               | String | No       | `GitHubAnnotations`       | Output format: Text, Markdown, or GitHubAnnotations                                                                                                                                             |
| `go-version`           | String | No       | `1.25`                    | Go version to use when building the binary                                                                                                                                                      |
| `repository-directory` | String | No       | `${{ github.workspace }}` | Path to the git repository to analyze                                                                                                                                                           |

<!-- END_INPUTS -->

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
