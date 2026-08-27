# create-github-app-token

From a `grafana/` org repository, get a ephemeral GitHub API token from a GitHub App using Vault.

## Inputs

<!-- BEGIN_INPUTS -->

| Name             | Type   | Required | Default   | Description                                                                                             |
| ---------------- | ------ | -------- | --------- | ------------------------------------------------------------------------------------------------------- |
| `github_app`     | String | No       |           | GitHub app name in Vault. You can define mutiple app to do a loadbalancing in a comma separated format. |
| `permission_set` | String | No       | `default` | Permission set name                                                                                     |
| `vault_instance` | String | No       | `ops`     | The Vault instance to use (`dev` or `ops`). Defaults to `ops`.                                          |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name    | Description                      |
| ------- | -------------------------------- |
| `token` | GitHub installation access token |

<!-- END_OUTPUTS -->

## Action Permissions

This action will need the following permissions in your workflow file to generate github OIDC token:

```yaml
permissions:
  id-token: write
```

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
      id-token: write

    steps:
      - id: get-github-token
        uses: grafana/shared-workflows/actions/create-github-app-token@create-github-app-token/v0.3.1
        with:
          github_app: github-app-name

      # Use the secrets
      - name: list issues assignees
        run: |
          curl -L \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${{ steps.get-github-token.outputs.token }}" \
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
      id-token: write

    steps:
      - id: get-github-token-read
        uses: grafana/shared-workflows/actions/create-github-app-token@create-github-app-token/v0.3.1
        with:
          github_app: github-app-name
          permission_set: read-only-on-foo-repository

      # Use the secrets
      - name: list issues assignees
        run: |
          curl -L \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${{ steps.get-github-token-read.outputs.token }}" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            https://api.github.com/repos/grafana/foo-repository/assignees

      - id: get-github-token-write
        uses: grafana/shared-workflows/actions/create-github-app-token@create-github-app-token/v0.3.1
        with:
          github_app: github-app-name
          permission_set: write-on-bar-repository

      # Use the secrets
      - name: create a pull request
        run: |
          curl -L \
            -X POST \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${{ steps.get-github-token-write.outputs.token }}" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            https://api.github.com/repos/grafana/bar-repository/pulls \
            -d '{"title":"Amazing new feature","body":"Please pull these awesome changes in!","head":"octocat:new-feature","base":"master"}'
```

<!-- x-release-please-end-version -->
