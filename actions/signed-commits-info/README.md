# Signed Commits Info Action

A GitHub Action that runs on pull requests and reports any commit (introduced
by the PR over its base ref) that does not have a fully verified signature
according to GitHub's [commit signature verification][verification].

The result is written to the workflow's `GITHUB_STEP_SUMMARY`. The action does
not fail the job — it is informational only.

## Usage

```yaml
name: Signed commits

on:
  pull_request:
  pull_request_target:

jobs:
  signed-commits:
    runs-on: ubuntu-latest
    steps:
      - uses: grafana/shared-workflows/actions/signed-commits-info@main
```

The default `${{ github.token }}` is sufficient; no extra scopes are required.

## Development

This project uses the [bun](https://bun.sh) toolchain.

```sh
bun install
bun run build       # bundles src/main.ts to dist/index.js
bun run typecheck
```

The bundled `dist/index.js` must be committed — GitHub Actions execute it
directly.

[verification]: https://docs.github.com/en/rest/commits/commits#list-commits
