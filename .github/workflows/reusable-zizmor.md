# Reusable workflow: zizmor

This is a [reusable workflow] which runs the [`zizmor`][zizmor] static analysis
tool on a repo's GitHub Actions workflow files. This will report things such as
whether there is potential for untrusted code to be injected via a template. See
a full list of checks in [the documentation][zizmor-checks].

This workflow will run zizmor, upload results to GitHub's code scanning service
(requires an Advanced Security subscription for private repositories), and
comment on the pull request with the results. The comment will be re-posted on
each run - and previous comments hidden - so the most recent comment will always
show the current results.

[reusable workflow]: https://docs.github.com/en/actions/using-workflows/reusing-workflows
[zizmor]: https://woodruffw.github.io/zizmor/
[zizmor-checks]: https://woodruffw.github.io/zizmor/audits/

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

      # required to comment on pull requests with the results of the check
      pull-requests: write
      # required to upload the results to GitHub's code scanning service
      security-events: write

    uses: grafana/shared-workflows/.github/workflows/reusable-zizmor.yml@<some sha>
    with:
      # example: fail if there are any findings
      fail-severity: any
```

## Inputs

| Name                      | Type    | Description                                                                                                                                                                                                                        | Default Value   | Required |
| ------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------- | -------- |
| min-severity              | string  | Only show results at or above this severity [possible values: unknown, informational, low, medium, high]                                                                                                                           | medium          | false    |
| min-confidence            | string  | Only show results at or above this confidence level [possible values: unknown, low, medium, high]                                                                                                                                  | low             | false    |
| fail-severity             | string  | Fail the build if any result is at or above this severity [possible values: never, any, informational, low, medium, high]                                                                                                          | high            | false    |
| runs-on                   | string  | The runner to use for jobs. Configure this to use self-hosted runners.                                                                                                                                                             | ubuntu-latest   | false    |
| default-config            | boolean | The default Zizmor configuration to use. If `always-use-default-config` is `true`, this configuration will always be used. Otherwise, it will be used if the repository does not have a `.github/zizmor.yml` or `zizmor.yml` file. | true            | false    |
| always-use-default-config | boolean | Whether to always use `default-config`.                                                                                                                                                                                            | false           | false    |
| github-token              | string  | The GitHub token to use when authenticating with the GitHub API                                                                                                                                                                    | ${github.token} | false    |

## Getting started

This workflow uses quite strict settings by default. It isn't always practical
to introduce this workflow and fix all of the issues at once. There are a few
ways to get started if this is the case:

1. Set `fail-severity: never` to run the check without failing the build.
   Results will still be posted to pull requests but they won't be blocking.
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

## Configuration

zizmor [can be configured][zizmor-config] with a `zizmor.yml` or
`.github/zizmor.yml` file in the repository. With this, [findings or entire
files can be ignored][zizmor-ignore-config].

[zizmor-config]: https://woodruffw.github.io/zizmor/configuration/
[zizmor-ignore-config]: https://woodruffw.github.io/zizmor/usage/#with-zizmoryml
