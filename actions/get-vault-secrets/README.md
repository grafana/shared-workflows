# get-vault-secrets

> [!NOTE]
> If you are at Grafana Labs, follow these steps in the [internal documentation](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/vault/#ci-secrets) for the paths where this workflow can read secrets from.

From a `grafana/` org repository, get a secret from the Grafana vault instance.
The secret format is defined here: <https://github.com/hashicorp/vault-action>

## Examples

### Using Environment Variables (default)

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
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets-v1.1.0
        with:
          # Secrets placed in the ci/common/<path> path in Vault
          common_secrets: |
            ENVVAR1=test-secret:testing
          # Secrets placed in the ci/repo/grafana/<repo>/<path> path in Vault
          repo_secrets: |
            ENVVAR2=test-secret:key1

      # Use the secrets
      # You can use the envvars directly in scripts
      - name: echo
        run: |
          echo "$ENVVAR1"
          echo "${ENVVAR2}"
```

<!-- x-release-please-end-version -->

### Using Outputs

You can also use the action with `export_env: false` to get secrets as outputs instead of environment variables:

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
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets-v1.1.0
        with:
          # Secrets placed in the ci/common/<path> path in Vault
          common_secrets: |
            SECRET1=test-secret:testing
          # Secrets placed in the ci/repo/grafana/<repo>/<path> path in Vault
          repo_secrets: |
            SECRET2=test-secret:key1
          # Set to false to get secrets as outputs instead of environment variables
          export_env: false

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

This approach is useful when you need to pass secrets to other actions or reusable workflows as inputs, while keeping them secure. It's also beneficial when you want to limit which steps have access to the secrets, as environment variables are available to all subsequent steps in a job, whereas outputs require explicit passing to each step that needs them.
