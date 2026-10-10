# get-vault-blessed-operations-secrets

> [!NOTE]
> If you are at Grafana Labs, see the [CI gates documentation](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/continuous-integration/ci-gates/) for how operations, blessed workflows and adopters fit together.

Read the secrets of a CI gates operation from Vault, from inside that operation's blessed workflow.

Secrets live under `ci/data/operations/<operation>/` in Vault. Unlike [`get-vault-secrets`](../get-vault-secrets), this action does not decide who may read them: it logs in with the operation's Vault role, and Vault only issues a token when the run's OIDC claims match it. That means the run must come from a repository that has adopted the operation, through the pinned blessed workflow, in the job that declares the operation's GitHub Environment, on an allowed protected ref.

The action registers a post-job step that revokes the Vault token when the job finishes.

## Inputs

<!-- BEGIN_INPUTS -->

| Name             | Type   | Required | Default | Description                                                                                                                                                     |
| ---------------- | ------ | -------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `operation`      | String | Yes      |         | CI gates operation name, as in the catalog (`terraform/ci-gates/catalog/<operation>.yaml` in deployment_tools), e.g. `publish_winget`.                          |
| `secrets`        | String | Yes      |         | Secrets to read, one per line, as `NAME=subpath:key`. The subpath is relative to `ci/data/operations/<operation>/`. Ex: `secrets: \| WINGET_TOKEN=winget:token` |
| `vault_instance` | String | No       | `ops`   | The Vault instance to use (`dev` or `ops`). Defaults to `ops`. A catalog binding with `instance: prod` is served by `ops`.                                      |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name      | Description                                                  |
| --------- | ------------------------------------------------------------ |
| `secrets` | JSON object mapping each requested name to its secret value. |

<!-- END_OUTPUTS -->

## Action Permissions

The job that uses this action needs:

```yaml
permissions:
  id-token: write
```

## Examples

To access the secrets, read the JSON `secrets` output: `${{ fromJSON(steps.secrets.outputs.secrets).WINGET_TOKEN }}`. Expose a secret as an environment variable only on the step that needs it.

<!-- x-release-please-start-version -->

```yaml
# A blessed workflow for the `publish_winget` operation.
name: publish-winget

on:
  workflow_call:
    inputs:
      package-id:
        type: string
        required: true

permissions: {}

jobs:
  publish:
    runs-on: windows-2025
    # Gate 1: GitHub checks this environment in the caller's repository
    # (allowed refs, reviewers) before the job starts. Must match the
    # operation's environment in the CI gates catalog.
    environment: publish_winget
    permissions:
      contents: read
      id-token: write
    steps:
      # Gate 2: Vault only issues a token if the run's OIDC claims match the
      # operation's role (adopted repository, this blessed workflow at its
      # pinned tag, allowed protected ref, the environment above).
      - id: secrets
        uses: grafana/shared-workflows/actions/get-vault-blessed-operations-secrets@get-vault-blessed-operations-secrets/v0.1.0
        with:
          operation: publish_winget
          # Reads key `token` of ci/data/operations/publish_winget/winget
          secrets: |
            WINGET_TOKEN=winget:token

      - name: Submit WinGet manifest
        shell: pwsh
        env:
          WINGET_CREATE_GITHUB_TOKEN: ${{ fromJSON(steps.secrets.outputs.secrets).WINGET_TOKEN }}
          PACKAGE_ID: ${{ inputs.package-id }}
        run: wingetcreate update $env:PACKAGE_ID --submit
```

<!-- x-release-please-end-version -->

Blessed workflows should pin this action by commit SHA. A blessed workflow is trusted at its pinned tag, but that pin does not cover the actions it calls.

## Troubleshooting

- **`Vault auth failed (HTTP 400)`**: the run's claims do not match the operation's role. Check that the calling repository has adopted the operation in CI gates, that this step runs in the job with `environment: <operation>`, and that the ref is allowed and protected.
- **`Vault read ... failed (HTTP 403)`**: the path is outside the operation's `read_paths` in the CI gates catalog.
