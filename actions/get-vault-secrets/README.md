# get-vault-secrets

> [!NOTE]
> If you are at Grafana Labs, follow these steps in the [internal documentation](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/continuous-integration/#vault-storing-your-secrets) for Vault instructions and best practices.

From a `grafana/` org repository, get a secret from the Grafana vault instance.
The secret format is defined here: <https://github.com/hashicorp/vault-action>

## Examples

To access the secrets, you need to read from the JSON `secrets` output of the action: `${{ fromJSON(steps.get-secrets.outputs.secrets).SECRET1 }}`.

_Secrets are no longer automatically exposed as environment variables, as these are accessible to all steps and can easily be leaked by malicious code. Instead, secrets should be exposed as environment variables only to the trusted steps that require them._

<!-- x-release-please-start-version -->

```yaml
name: CI
on:
  pull_request:

# These permissions are needed to assume roles from Github's OIDC.
permissions:
  contents: read
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - id: get-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets/v2.0.1
        with:
          # Secrets placed in the ci/common/<path> path in Vault
          common_secrets: |
            SECRET1=test-secret:testing
          # Secrets placed in the ci/repo/grafana/<repo>/<path> path in Vault
          repo_secrets: |
            SECRET2=test-secret:key1

      # Use the secrets from the JSON output in the env block
      - name: echo
        env:
          ENVVAR1: ${{ fromJSON(steps.get-secrets.outputs.secrets).SECRET1 }}
          ENVVAR2: ${{ fromJSON(steps.get-secrets.outputs.secrets).SECRET2 }}
        run: |
          echo "$ENVVAR1"
          echo "${ENVVAR2}"
```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name             | Type    | Required | Default | Description                                                                                                                                                                                                                                                            |
| ---------------- | ------- | -------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `common_secrets` | String  | No       |         | Common secrets mapping. Only Grafana Labs employees can list these secrets in Vault. These secrets are stored in the `ci/data/common/<secret path>` path in Vault. Ex: `repo_secrets: \| ENVVAR1=secretpath:key ENVVAR2=secretpath:key2`                               |
| `ignore_missing` | Boolean | No       | `false` | When set to true, prevents the action from failing when a secret does not exist.                                                                                                                                                                                       |
| `repo_secrets`   | String  | No       |         | Repository-specific secret mapping. Only Grafana Labs employees can list these secrets in Vault. These secrets are stored in the `ci/data/repo/grafana/<repo name>/<secret path>` path in Vault. Ex: `repo_secrets: \| ENVVAR1=secretpath:key ENVVAR2=secretpath:key2` |
| `vault_instance` | String  | No       | `ops`   | The Vault instance to use (`dev` or `ops`). Defaults to `ops`.                                                                                                                                                                                                         |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name      | Description                            |
| --------- | -------------------------------------- |
| `secrets` | JSON object containing all the secrets |

<!-- END_OUTPUTS -->
