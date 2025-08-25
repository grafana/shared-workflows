# create-github-app-token

From a `grafana/` org repository, get a ephemeral GitHub API token from a GitHub App using Vault.

## Inputs

| Name             | Type   | Description                 | Default Value | Required |
| ---------------- | ------ | --------------------------- | ------------- | -------- |
| `permission_set` | String | The required permission set | `default`     | Yes      |
| `github-app`     | String | The required GitHub app     |               | Yes      |
| `vault_instance` | String | Vault instance to point     | `ops`         | No       |

## Outputs

| Name           | Type   | Description                |
| -------------- | ------ | -------------------------- |
| `github_token` | String | The generated GitHub token |

## Examples

### Using Environment Variables (default)

<!-- x-release-please-start-version -->

#### Using default permission set

```yaml
name: CI
on:
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest

    # These permissions are needed to assume roles from GitHub's OIDC.
    permissions:
      contents: read
      id-token: write

    steps:
      - id: get-github-token
        uses: grafana/shared-workflows/actions/create-github-app-token@create-github-app-token/v0.2.0
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

#### Using multiple permissions sets

```yaml
name: CI
on:
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest

    # These permissions are needed to assume roles from GitHub's OIDC.
    permissions:
      contents: read
      id-token: write

    steps:
      - id: get-github-token-read
        uses: grafana/shared-workflows/actions/create-github-app-token@create-github-app-token/v0.2.0
        with:
          github_app: github-app-name
          permissions-set: read-only-on-foo-repository

      # Use the secrets
      - name: list issues assignees
        run: |
          curl -L \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${{ steps.get-github-token-read.outputs.github_token }}" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            https://api.github.com/repos/grafana/foo-repository/assignees

      - id: get-github-token-write
        uses: grafana/shared-workflows/actions/create-github-app-token@create-github-app-token/v0.2.0
        with:
          github_app: github-app-name
          permissions-set: write-on-bar-repository

      # Use the secrets
      - name: create a pull request
        run: |
          curl -L \
            -X POST \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${{ steps.get-github-token-write.outputs.github_token }}" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            https://api.github.com/repos/grafana/bar-repository/pulls \
            -d '{"title":"Amazing new feature","body":"Please pull these awesome changes in!","head":"octocat:new-feature","base":"master"}'
```

<!-- x-release-please-end-version -->
