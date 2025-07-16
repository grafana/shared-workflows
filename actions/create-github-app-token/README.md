# get-vault-secrets

From a `grafana/` org repository, get a ephemeral github API token from a Github App using vault.

## Inputs

| Name             | Type   | Description                 | Default Value | Required |
| ---------------- | ------ | --------------------------- | ------------- | -------- |
| `permission_set` | String | The required permission set | `default`     | Yes      |
| `github-app`     | String | The required github app     |               | Yes      |
| `vault_instance` | String | Vault instance to point     | `ops`         | No       |

## Outputs

| Name           | Type   | Description                |
| -------------- | ------ | -------------------------- |
| `github_token` | String | The generated github token |

## Examples

### Using Environment Variables (default)

<!-- x-release-please-start-version -->

```yaml
name: CI
on:
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest

    # These permissions are needed to assume roles from Github's OIDC.
    permissions:
      contents: read
      id-token: write

    steps:
      - id: get-github-token
        uses: grafana/shared-workflows/actions/create-github-app-token@create-github-app-token/v1.2.1
        with:
          github_app: github-app-name

      # Use the secrets
      - name: list issues assignees
        run: |
          curl -L \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${{ steps.get-github-token.outputs.github_token }}" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            https://api.github.com/repos/grafana/grafana/assignees
```
