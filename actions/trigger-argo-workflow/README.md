# Trigger Argo Workflow Action

> [!NOTE]
> If you are at Grafana Labs, see the [internal documentation](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/continuous-delivery/argo-workflows/#triggering-a-workflow-from-github-actions) for information on how to set up Argo Workflows and configure them to be triggerable by this action.

This GitHub action triggers an Argo workflow in one of the Grafana Labs Argo
Workflows instances. It contains a small wrapper around the `argo` CLI, which is
downloaded as part of the action.

As our Argo Workflows instances require authentication, this workflow will only
work when triggered from a `grafana`-owned repository. We would welcome
contributions to extend this action to work with any instance, or you are free
to fork and modify to run on your own instances. See [#21][issue-21].

[issue-21]: https://github.com/grafana/shared-workflows/issues/21

## How to use

## Inputs

- `instance`: The instance to use (`dev` or `ops`). Defaults to `ops`.
- `namespace`: Required. The namespace to trigger the workflow in.
- `parameters`: The newline-separated parameters to pass to the Argo workflow. Example:

```yaml
parameters: |
  param1=value1
  param2=value2
```

- `workflow_template`: The workflow template to use. Required if `command` is `submit` (the default).
- `extra_args`: Extra arguments to pass to the Argo CLI. Example: `--generate-name foo-`
- `log_level`: The log level to use. Choose from `debug`, `info`, `warn` or `error`. Defaults to `info`.

## Outputs

- `uri`: The URI of the workflow that was created.

## Required permissions

This action needs a couple of explicit `GITHUB_TOKEN` scopes because it:

- authenticates to Vault via GitHub OIDC (needs **`id-token: write`**)
- checks out / reads Go files from the repo (needs **`contents: read`**)

Ideally, place these permissions at the job level to avoid zizmor flagging them as [excessive permissions](https://woodruffw.github.io/zizmor/audits/#excessive-permissions).

```yaml
permissions:
  contents: read # allows actions/checkout and setup-go to read the repo
  id-token: write # allows get-vault-secrets to create an OIDC token for Vault
```

## Usage

Here is an example of how to use this action:

<!-- x-release-please-start-version -->

```yaml
name: Trigger Argo Workflow
on:
  pull_request:

jobs:
  trigger-argo-workflow:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
    steps:
      - name: Trigger Argo Workflow
        uses: grafana/shared-workflows/actions/trigger-argo-workflow@0f705663f602e305aa22034489f351dc7022d8ce # trigger-argo-workflow-v1.3.0
        with:
          instance: "ops"
          namespace: "mynamespace"
          workflow_template: "hello"
          parameters: |
            message=world
          extra_args: "--generate-name hello-world-"
          log_level: "debug"
```

<!-- x-release-please-end-version -->
