# Check Renovate enablement

Lightweight repository policy checker for the
[organization workflow](../../.github/workflows/global-renovate-policy.md).
It implements the three explicit disablement checks proposed in
[shared-workflows#2352](https://github.com/grafana/shared-workflows/pull/2352)
without installing Renovate or changing existing schema-validator callers.

## Run locally

With Node 24, from this directory:

```sh
npm ci --ignore-scripts --no-audit --no-fund
npm test
node check.mjs /absolute/path/to/repository
```

Exit codes: `0` passes, `1` reports explicit repository-wide disablement, `2`
reports invalid/unreadable configuration or an invocation error. The CLI writes
GitHub annotations and a job summary, without printing configuration contents.

## Configuration discovery

The first present configuration file wins, using Renovate's default order:
`renovate.json`, `renovate.jsonc`, `renovate.json5`, the same filenames under
`.github/`, then `.gitlab/`, then `.renovaterc`, `.renovaterc.json`,
`.renovaterc.jsonc`, `.renovaterc.json5`, and finally the `renovate` property of
`package.json`. An ordinary `package.json` is not a Renovate config.

The self-hosted deployment currently uses default filenames. Custom
`configFileNames` and inherited presets are not resolved here. Keep discovery
aligned with the deployed Renovate version; changes to that setting require
updating this checker and the workflow's sparse checkout together.

Standalone configuration is parsed with lockfile-pinned JSON5; `package.json`
uses JSON. Configurations must be objects. The checker validates the types of the
three fields it inspects, but does not replace Renovate's schema validator. It
does not execute repository JavaScript, install target repository dependencies,
or evaluate package rules. Symlinks and files larger than 1 MiB fail explicitly.

The workflow installs only `json5` from this trusted directory, with lifecycle
scripts disabled and the npm download cache keyed by the lockfile. The target
repository is checked out separately and used only as data. Include cold-cache
install and checkout times when measuring the brownout's runner cost.
