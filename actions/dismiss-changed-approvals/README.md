# Dismiss Changed Approvals Action

A GitHub Action that replaces the native "Dismiss stale pull request
approvals" branch protection with a content-based equivalent: approvals are
dismissed only when the content the reviewer approved actually changed.

For every active approval, the action compares the SHA-256 hash of the PR's
three-dot diff (`base...head`) at the approved commit against the current one.
Identical hashes mean the reviewed content is unchanged — clean merges from
the base branch and clean rebases keep their approvals. On top of diff
equality, every commit pushed after the approval must have a verified
signature and be authored by the PR author (GitHub-generated base-branch merge
commits are exempt from the author rule).

The action fails closed:

- A changed diff, an unsigned or foreign commit after approval, or an
  approved commit whose diff can no longer be computed (rewritten history)
  dismisses the approval.
- An API or evaluation error fails the job **without** dismissing anything, so
  a required-check failure blocks merging while preserving reviews.
- The action can only ever dismiss reviews. Keep the org/repo setting
  "Allow GitHub Actions to create and approve pull requests" disabled so the
  workflow token is platform-blocked from approving.

Repositories using this action should disable the native "Dismiss stale pull
request approvals" setting, keep required approvals and CODEOWNERS review
enabled, and must not rely on fork contributions (fork PRs are skipped because
the workflow token is read-only there).

Every run writes its verdicts and diff hashes to `GITHUB_STEP_SUMMARY` as an
audit trail.

## Usage

<!-- x-release-please-start-version -->

```yaml
name: Dismiss changed approvals

on:
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]

jobs:
  dismiss-changed-approvals:
    permissions:
      contents: read
      pull-requests: write
    runs-on: ubuntu-latest
    steps:
      - uses: grafana/shared-workflows/actions/dismiss-changed-approvals@dismiss-changed-approvals/v0.1.0
        with:
          # Set to true to log verdicts without dismissing (shadow mode).
          dry-run: false
```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name           | Type    | Required | Default               | Description                                   |
| -------------- | ------- | -------- | --------------------- | --------------------------------------------- |
| `dry-run`      | Boolean | No       | `false`               | Log verdicts without dismissing any approvals |
| `github-token` | String  | No       | `${{ github.token }}` | Token used to call the GitHub API             |

<!-- END_INPUTS -->

## Development

This project uses the [bun](https://bun.sh) toolchain.

```sh
bun install
bun run build       # bundles src/main.ts to dist/index.js
bun run typecheck
bun test
```

The bundled `dist/index.js` must be committed — GitHub Actions execute it
directly.
