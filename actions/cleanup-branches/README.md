# cleanup-branches

Composite action (step) to query for branches that are not in an open PR, and delete them if 'dry-run' is 'false'. Protected branches are excluded as well.

## Inputs

| Name       | Type     | Description                                                                                                                                                                        | Default Value         | Required |
| ---------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------- | -------- |
| `token`    | `string` | GitHub token used to authenticate with `gh`. Requires permission to query for protected branches and delete branches (`contents: write`) and pull requests (`pull_requests: read`) | `${{ github.token }}` | true     |
| `dry-run`  | `bool`   | If `'true'`, then the action will print branches to be deleted, but will not delete them                                                                                           | `'true'`              | true     |
| `max-date` | `string` | Value passed to `date -d`; a human readable date string. Maximum date of the head ref of a branch in order to be deleted.                                                          | `"2 weeks ago"`       | false    |

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
      - uses: grafana/shared-workflows/actions/cleanup-branches@cleanup-branches/v0.2.0
        with:
          dry-run: false
```

<!-- x-release-please-end-version -->
