# download-branch-workflow-artifact

Downloads an artifact from the last successful run of a workflow on a specific branch. This is useful for retrieving build state from previous deployments (e.g., component tags, digests) without relying on third-party actions.

Uses only first-party GitHub actions (`actions/github-script` and `actions/download-artifact`) to query for the most recent successful workflow run and download the specified artifact.

## Inputs

| Name            | Type     | Description                                 | Default Value         | Required |
| --------------- | -------- | ------------------------------------------- | --------------------- | -------- |
| `workflow`      | `string` | Workflow filename to download artifact from |                       | true     |
| `artifact-name` | `string` | Name of the artifact to download            |                       | true     |
| `branch`        | `string` | Branch to filter workflow runs by           | `main`                | false    |
| `path`          | `string` | Directory to download the artifact to       | `.`                   | false    |
| `github-token`  | `string` | GitHub token with `actions:read` permission | `${{ github.token }}` | false    |

## Outputs

| Name            | Type     | Description                                                    |
| --------------- | -------- | -------------------------------------------------------------- |
| `found`         | `string` | Whether the artifact was found and downloaded (`true`/`false`) |
| `run-id`        | `string` | The workflow run ID the artifact was downloaded from           |
| `download-path` | `string` | Path where the artifact was downloaded                         |

## Permissions

The calling workflow must have `actions: read` permission for the GitHub token to query workflow runs and download cross-run artifacts.

## Examples

### Download component tags from the last successful deployment

<!-- x-release-please-start-version -->

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
    steps:
      - uses: actions/checkout@v4

      - name: Download previous deployment state
        id: previous
        uses: grafana/shared-workflows/actions/download-branch-workflow-artifact@download-branch-workflow-artifact/v0.1.0
        with:
          workflow: deploy-prod.yml
          artifact-name: component-tags

      - name: Use previous state
        if: steps.previous.outputs.found == 'true'
        run: cat component-tags.json
```

<!-- x-release-please-end-version -->
