# cleanup-branches

Composite action (step) to query for branches that are not in an open PR, and delete them if 'dry-run' is 'false'. Protected branches are excluded as well.

## Inputs

<!-- BEGIN_INPUTS -->

| Name               | Type    | Required | Default               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| ------------------ | ------- | -------- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `dry-run`          | Boolean | No       | `false`               | If 'true', then the action will print branches to be deleted, but will not delete them                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `exclude-patterns` | String  | No       |                       | Optional list of glob patterns. Branches whose names match any pattern are excluded from deletion. Patterns use bash glob syntax (e.g. `release/*`).                                                                                                                                                                                                                                                                                                                                                                                     |
| `max-date`         | String  | No       | `3 months ago`        | Value provided to `date -d={}. From `man date`: "The --date=STRING is a mostly free format human readable date string such as "Sun, 29 Feb 2004 16:21:42 -0800" or "2004-02-29 16:21:42" or even "next Thursday". A date string may contain items indicating calendar date, time of day, time zone, day of week, relative time, relative date, and numbers. An empty string indicates the beginning of the day. The date string format is more complex than is easily documented here but is fully described in the info documentation." |
| `token`            | String  | No       | `${{ github.token }}` | GitHub token used to authenticate with `gh`. Requires permission to query for protected branches and delete branches (contents: write) and pull requests (pull_requests: read)                                                                                                                                                                                                                                                                                                                                                           |

<!-- END_INPUTS -->

## Examples

### Clean up branches on a weekly cron schedule

<!-- x-release-please-start-version -->

```yaml
name: Clean up orphaned branches
on:
  schedule:
    - cron: "0 9 * * 1"

jobs:
  cleanup-branches:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: read
    steps:
      - uses: actions/checkout@v5
      - uses: grafana/shared-workflows/actions/cleanup-branches@cleanup-branches/v0.3.1
        with:
          dry-run: false
          exclude-patterns: |
            release/*
```

<!-- x-release-please-end-version -->
