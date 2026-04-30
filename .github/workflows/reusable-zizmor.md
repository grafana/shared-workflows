# Reusable workflow: zizmor

This is a [reusable workflow] which runs the [`zizmor`][zizmor] static analysis
tool on a repo's GitHub Actions workflow files. This will report things such as
whether there is potential for untrusted code to be injected via a template. See
a full list of checks in [the documentation][zizmor-checks].

This workflow will run zizmor and upload results to GitHub's code scanning
service. Findings are surfaced as inline annotations on pull requests via
GitHub's Code Scanning integration. For private repositories without Advanced
Security, the workflow falls back to posting a PR comment with the results.

[reusable workflow]: https://docs.github.com/en/actions/using-workflows/reusing-workflows
[zizmor]: https://woodruffw.github.io/zizmor/
[zizmor-checks]: https://woodruffw.github.io/zizmor/audits/

## Examples

### Online Checks

```yaml
name: Zizmor GitHub Actions static analysis
on:
  pull_request:
    paths:
      - ".github/**"
  push:
    branches:
      - main
    paths:
      - ".github/**"

jobs:
  scorecard:
    name: Analyse

    permissions:
      actions: read
      contents: read

      # used in the `job-workflow-ref` job to fetch an OIDC token, which
      # allows the run to determine its ref. That's used to find the default
      # configuration file. This doesn't work from forks. In that case,
      # Zizmor's default config behaviour will be used.
      id-token: write

      # fallback: comment on PR when code-scanning upload is unavailable
      # (private repos without Advanced Security)
      pull-requests: write
      # required to upload the results to GitHub's code scanning service. This
      # doesn't work if the repo doesn't have Advanced Security enabled. In that
      # case the workflow falls back to posting a PR comment.
      security-events: write

    uses: grafana/shared-workflows/.github/workflows/reusable-zizmor.yml@<some sha>
    with:
      # example: fail if there are any findings
      fail-severity: any
```

### Faster Offline Checks

```yaml
name: Zizmor GitHub Actions static analysis (online checks)
on:
  pull_request:
    paths:
      - ".github/**"
  push:
    branches:
      - main
    paths:
      - ".github/**"

jobs:
  scorecard:
    name: Analyse

    permissions:
      actions: read
      contents: read

      # used in the `job-workflow-ref` job to fetch an OIDC token, which
      # allows the run to determine its ref. That's used to find the default
      # configuration file. This doesn't work from forks. In that case,
      # Zizmor's default config behaviour will be used.
      id-token: write

      # fallback: comment on PR when code-scanning upload is unavailable
      # (private repos without Advanced Security)
      pull-requests: write
      # required to upload the results to GitHub's code scanning service. This
      # doesn't work if the repo doesn't have Advanced Security enabled. In that
      # case the workflow falls back to posting a PR comment.
      security-events: write

    uses: grafana/shared-workflows/.github/workflows/reusable-zizmor.yml@<some sha>
    with:
      # example: fail if there are any findings
      fail-severity: any
      extra-args: "--offline"
```

## Inputs

| Name                      | Type    | Description                                                                                                                                                                  | Default Value   | Required |
| ------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------- | -------- |
| min-severity              | string  | Only show results at or above this severity [possible values: unknown, informational, low, medium, high]                                                                     | medium          | false    |
| min-confidence            | string  | Only show results at or above this confidence level [possible values: unknown, low, medium, high]                                                                            | low             | false    |
| fail-severity             | string  | Fail the build if any result is at or above this severity [possible values: never, any, informational, low, medium, high]                                                    | high            | false    |
| runs-on                   | string  | The runner to use for jobs. Configure this to use self-hosted runners.                                                                                                       | ubuntu-latest   | false    |
| always-use-default-config | boolean | Whether to always use the [default configuration]. When `false`, `.zizmor.yml` or `.github/zizmor.yml` will be used, if present.                                             | false           | false    |
| github-token              | string  | The GitHub token to use when authenticating with the GitHub API                                                                                                              | ${github.token} | false    |
| extra-args                | string  | Extra arguments to pass into zizmor                                                                                                                                          | ""              | false    |
| send-bench-metrics        | boolean | If true, run Grafana Bench after analysis to send zizmor metrics to Prometheus. Uses shared Vault secrets (grafana-bench); no caller secrets required. Set to false to skip. | true            | false    |

[default configuration]: ../zizmor.yml

## Grafana Bench (Prometheus metrics)

When `send-bench-metrics` is true (default), the workflow runs a second job after zizmor analysis that:

1. Fetches **Prometheus credentials** from Vault (shared `grafana-bench` common secrets), same as the [setup-grafana-bench](https://github.com/grafana/grafana-bench/blob/main/.github/actions/setup-grafana-bench/action.yml) action.
2. Downloads the **SARIF file** artifact produced by the analysis job.
3. Runs the [Grafana Bench](https://github.com/grafana/grafana-bench) Docker image with `report --report-input zizmor` and `--prometheus-metrics`, sending metrics to Grafana’s Prometheus (ops) endpoint.

**No caller configuration needed:** You do not need to set `PROMETHEUS_URL` or pass `secrets: inherit` for metrics; the workflow uses Vault.

**To disable:** Pass `send-bench-metrics: false` so the Grafana Bench job is skipped.

**If only `bench_*` metrics appear in Prometheus and `zizmor_*` metrics are missing:**

1. **Docker image version:** The image `grafana-bench:v1.0.2` must be built from a grafana-bench commit that includes the zizmor parser and the Prometheus reporter logic that pushes `summary.Metrics`. If the image was built before that code was merged, build and push a new image from current grafana-bench main (e.g. tag v1.0.3) and update the workflow to use that tag.
2. **Workflow log:** The job runs bench with `--log-level debug`. In the "Run Grafana Bench (Docker image)" step, check for Prometheus push errors or missing env (e.g. "PROMETHEUS_URL not set").
3. **Remote write:** Confirm the Prometheus/Mimir remote-write endpoint accepts and retains the `zizmor_*` metric names (no filtering or drop rules).

## Getting started

This workflow uses quite strict settings by default. It isn't always practical
to introduce this workflow and fix all of the issues at once. There are a few
ways to get started if this is the case:

1. Set `fail-severity: never` to run the check without failing the build.
   Results will still be visible via Code Scanning annotations but won't be blocking.
2. Adopt an incremental approach to fixing issues. For example, start with
   `min-severity: high`. Once all high severity issues are resolved, lower the
   severity to `medium` and then onwards to `low`.

After the initial setup, we recommend running with the default settings.

## Ignore findings

Findings can be ignored by [adding a comment to the line with the finding][zizmor-ignore-comment].

```yaml
uses: actions/checkout@v3 # zizmor: ignore[artipacked]
```

[zizmor-ignore-comment]: https://woodruffw.github.io/zizmor/usage/#with-comments

## Repo-local `zizmor.yml` policy gate

When `always-use-default-config` is `false` and the calling repository uses a **repo-local** `zizmor.yml` or `.github/zizmor.yml` (so zizmor discovers that file instead of the Grafana default from `shared-workflows`), the workflow **validates that file before running zizmor**. If validation fails, the job stops and zizmor is not executed. Validation uses the composite action [`actions/validate-zizmor-config`](../../actions/validate-zizmor-config) from this repository, **hash-pinned** in `reusable-zizmor.yml` to satisfy pinning checks (bump that SHA when you change the action). **Set up Zizmor configuration** writes the path to validate as step output `repo-local-zizmor-config` when a repo-local file is used; that is separate from `zizmor-config`, which is only set when the run uses **`--config`** with the downloaded Grafana default.

The gate is **skipped** when:

- `always-use-default-config` is `true` (only the Grafana default from this repository is used), or
- There is no repo-local config file, or
- The run uses the fetched default via `--config` (no repo-local file in play).

The policy rejects configs that:

- Define any of these audit blocks under `rules` (no `disable`, `ignore`, or `config` is allowed — remove the block entirely): `insecure-commands`, `template-injection`, `impostor-commit`, `known-vulnerable-actions`, and `ref-confusion`.
- Set `rules.unpinned-uses.disable`.
- Set `rules.unpinned-uses.config.policies` with a universal [`"*": any`](https://docs.zizmor.sh/audits/#unpinned-uses) entry (all matching `uses:` clauses may stay unpinned). Scoped policies such as `actions/*: any` or `grafana/*: any` remain valid.

Inline `# zizmor: ignore[...]` comments in workflow files are unchanged; this gate applies only to the repo-local YAML config file.

## Configuration

zizmor [can be configured][zizmor-config] with a `zizmor.yml` or
`.github/zizmor.yml` file in the repository. With this, [findings or entire
files can be ignored][zizmor-ignore-config].

[zizmor-config]: https://woodruffw.github.io/zizmor/configuration/
[zizmor-ignore-config]: https://woodruffw.github.io/zizmor/usage/#with-zizmoryml
