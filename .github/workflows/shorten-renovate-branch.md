# shorten-renovate-branch

This is a reusable workflow that creates mirror PRs with short branch names when Renovate generates branches that are
too long for Google Workload Identity Federation (WIF).

> [!NOTE]
> There is a [bug with Google Artifact Registry](https://issuetracker.google.com/issues/390719013) where WIF's
> `assertion.sub` claim has a ~127 byte limit. The claim format is `repo:<owner>/<repo>:ref:refs/heads/<branch>`, so
> long Renovate branch names combined with long repo names can exceed this limit and break GAR authentication.

## How it works

When a Renovate PR is opened or updated, this workflow checks if the combined length of the repository name and branch
name exceeds a configurable threshold (default: 100 characters). If it does, the workflow:

1. Force-pushes the Renovate branch content to a short branch named `<original-pr-number>`
2. Creates a mirror PR targeting the default branch
3. Copies labels from the original PR to the mirror
4. Comments on the original PR linking to the mirror

On subsequent pushes (`synchronize`), the short branch is force-pushed to stay in sync.

When the original PR is closed, the mirror PR is closed and the short branch is deleted.

```yaml
name: Shorten Renovate branches

on:
  pull_request:
    types: [opened, synchronize, closed]

permissions:
  contents: write
  pull-requests: write

jobs:
  shorten:
    uses: grafana/shared-workflows/.github/workflows/shorten-renovate-branch.yml@main
```

## Inputs

| Name                  | Type   | Description                                                                                                                                                  |
| --------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `max-combined-length` | number | Maximum allowed length of repo name + branch name combined. WIF `assertion.sub` has a ~127 byte limit; subtract 21 chars overhead to get the default of 100. |
