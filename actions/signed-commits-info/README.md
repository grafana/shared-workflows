# Signed Commits Info Action

A GitHub Action that runs on pull requests and reports any commit (introduced
by the PR over its base ref) that does not have a fully verified signature
according to GitHub's [commit signature verification][verification].

The result is written to the workflow's `GITHUB_STEP_SUMMARY` and also
published as comment to the pull-request. The action does not fail the job if a
commit could not be verified — it is informational only.

## Usage

<!-- x-release-please-start-version -->

```yaml
name: Signed commits

on:
  pull_request:
  # zizmor: ignore[dangerous-triggers] The underlying action does not interact in any way with the source code of the fork. It requires the scope of the original repository in order to post a comment to the pull request itself.
  pull_request_target:

jobs:
  signed-commits:
    permissions:
      contents: read
      pull-requests: write
    if: >
      (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository) ||
      (github.event_name == 'pull_request_target' && github.event.pull_request.head.repo.full_name != github.repository)
    runs-on: ubuntu-latest
    steps:
      - uses: grafana/shared-workflows/actions/signed-commits-info@signed-commits-info/v0.1.0
        continue-on-error: true
```

<!-- x-release-please-end-version -->

## Development

This project uses the [bun](https://bun.sh) toolchain.

```sh
bun install
bun run build       # bundles src/main.ts to dist/index.js
bun run typecheck
```

The bundled `dist/index.js` must be committed — GitHub Actions execute it
directly.

[verification]: https://docs.github.com/rest/commits/commits#list-commits
