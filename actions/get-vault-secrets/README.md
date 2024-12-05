# get-vault-secrets

> > If you are at Grafana Labs, follow these steps in the [internal documentation](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/vault/) first.

From a `grafana/` org repository, get a secret from the Grafana vault instance.
The secret format is defined here: <https://github.com/hashicorp/vault-action>

Example workflow:

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
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets-v1.0.1
        with:
          # Secrets placed in the ci/common/<path> path in Vault
          common_secrets: |
            ENVVAR1=test-secret:testing
          # Secrets placed in the ci/repo/grafana/<repo>/<path> path in Vault
          repo_secrets: |
            ENVVAR2=test-secret:key1

      # Use the secrets
      # You can use the envvars directly in scripts or use the `${{ env.* }}` accessor in the workflow
      - name: echo
        run: |
          echo "$ENVVAR1"
          echo "${{ env.ENVVAR2 }}"
```

<!-- x-release-please-end-version -->
