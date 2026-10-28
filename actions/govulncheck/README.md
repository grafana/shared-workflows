# govulncheck

Runs [govulncheck](https://go.dev/doc/security/vuln/) using
[golang/govulncheck-action](https://github.com/golang/govulncheck-action),
publishes the results to the job summary and, optionally, and optionally
notifies a Slack channel.

Only vulnerabilities in code that is actually called are reported as findings,
which is the same rule govulncheck uses to report failure (exit code 3).
Vulnerabilities in dependencies that are imported but not called are listed in
a collapsed section of the summary, and don't fail the job or trigger a Slack
message.

The action checks out the repository itself, without persisting credentials,
and installs Go and govulncheck. Set `checkout` to `false` to run on a
repository that is already checked out, and use `work-dir` if the Go module
isn't at the root of the repository.

When vulnerabilities are found, the summary and the Slack message are produced
first and the action fails afterwards, unless `fail-on-vulnerabilities` is set
to `false`.

## Example workflow

<!-- x-release-please-start-version -->

```yaml
name: govulncheck

on:
  pull_request:
  schedule:
    - cron: "0 7 * * 1-5"

permissions: {}

jobs:
  govulncheck:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write # only needed when slack-channel is set
    steps:
      - uses: grafana/shared-workflows/actions/govulncheck@govulncheck/v0.1.0
        with:
          # Optional, defaults to the version used by golang/govulncheck-action.
          go-version: stable
          # Optional, the ID of the channel (not its name). No message is sent if not set.
          slack-channel: C0123456789
```

<!-- x-release-please-end-version -->

## Minimum Go version

`golang/govulncheck-action` installs the latest release of govulncheck, which
requires a recent Go toolchain. If `go-version` is older than that, the
`Install govulncheck` step fails with an error like:

```text
go: golang.org/x/vuln/cmd/govulncheck@latest: golang.org/x/vuln@v1.9.0 requires go >= 1.26.0 (running go 1.25.14; GOTOOLCHAIN=local)
```

To find the minimum Go version currently required, run:

```sh
go list -m -f '{{.GoVersion}}' golang.org/x/vuln@latest
```

Use a `go-version` at least as recent as the result, or leave it unset to use
the default of `golang/govulncheck-action`. The module under analysis can still
declare an older Go version in its `go.mod`.

## Slack notifications

The message is sent with [send-slack-message](../send-slack-message/README.md),
which obtains its token from Vault. This means that:

- the calling job needs the `id-token: write` permission, and
- the repository must be in the `grafana` organization.

A failure to send the message is reported but doesn't fail the action, so it
can't hide the vulnerabilities themselves.

## Inputs

<!-- BEGIN_INPUTS -->

| Name                      | Type    | Required | Default | Description                                                                                                                                      |
| ------------------------- | ------- | -------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `checkout`                | Boolean | No       | `true`  | Check out the repository (without persisting credentials). Set to `false` to run on a repository that is already checked out.                    |
| `fail-on-vulnerabilities` | Boolean | No       | `true`  | Fail the action if vulnerabilities are found in code that is called. The summary and the Slack message are always produced first.                |
| `go-version`              | String  | No       |         | Go version to run govulncheck with. If not set, the default of golang/govulncheck-action is used.                                                |
| `slack-channel`           | String  | No       |         | ID of the Slack channel to notify when vulnerabilities are found. If not set, no Slack message is sent. The calling job needs `id-token: write`. |
| `work-dir`                | String  | No       | `.`     | Directory to run govulncheck in, for repositories where the Go module is not at the root.                                                        |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name                    | Description                                                                    |
| ----------------------- | ------------------------------------------------------------------------------ |
| `has-vulnerabilities`   | `true` if vulnerabilities were found in code that is called, `false` otherwise |
| `vulnerabilities-found` | Number of vulnerabilities found in code that is called                         |

<!-- END_OUTPUTS -->
