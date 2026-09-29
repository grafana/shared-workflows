# download-branch-workflow-artifact

Downloads the newest artifact with a given name produced by a workflow on a specific branch. This is useful for retrieving build state from previous deployments (e.g., component tags, digests) without relying on third-party actions.

Uses only the first-party `actions/github-script` action. It lists the repository's artifacts by name (retrying a few times to work around the eventually-consistent artifacts API), keeps only those produced by successful runs of the requested workflow and branch, and downloads the most recently created one. Selection is based on the artifact creation time rather than the order in which workflow runs are returned, which is not stable and could otherwise cause an older artifact to be downloaded at random.

## Inputs

<!-- BEGIN_INPUTS -->

| Name            | Type   | Required | Default               | Description                                                        |
| --------------- | ------ | -------- | --------------------- | ------------------------------------------------------------------ |
| `artifact-name` | String | Yes      |                       | Name of the artifact to download                                   |
| `branch`        | String | No       | `main`                | Branch to filter workflow runs by                                  |
| `github-token`  | String | No       | `${{ github.token }}` | GitHub token with actions:read permission                          |
| `path`          | String | No       | `.`                   | Directory to download the artifact to                              |
| `workflow`      | String | Yes      |                       | Workflow filename to download artifact from (e.g. deploy-prod.yml) |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name            | Description                                                               |
| --------------- | ------------------------------------------------------------------------- |
| `download-path` | Path where the artifact was downloaded                                    |
| `found`         | Whether the artifact was found and downloaded (true/false)                |
| `run-id`        | The workflow run ID the artifact was downloaded from (empty if not found) |

<!-- END_OUTPUTS -->

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
