# issues-update-project-status

Updates a GitHub Project (v2) issue status field when triggered by an issue event.

The calling job must have `id-token: write` permission for Vault authentication.

## Inputs

<!-- BEGIN_INPUTS -->

| Name                      | Type   | Required | Default   | Description                                                                                                                                                                |
| ------------------------- | ------ | -------- | --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `github-app`              | String | Yes      |           | GitHub app name in Vault.                                                                                                                                                  |
| `permission-set`          | String | No       | `default` | Permission set name                                                                                                                                                        |
| `project-id`              | String | Yes      |           | The node ID of the GitHub Project (v2) board. Retrieve with: gh project view &lt;number> --owner &lt;org> --format json                                                    |
| `status-field-id`         | String | Yes      |           | The node ID of the Status field on the project board. Retrieve with: gh project field-list &lt;number> --owner &lt;org> --format json                                      |
| `target-status-option-id` | String | Yes      |           | The ID of the status option to set on the project item (e.g. the "In Progress" option ID). Retrieve with: gh project field-list &lt;number> --owner &lt;org> --format json |

<!-- END_INPUTS -->

## Filtering by label

To limit which issues trigger a status update, use an `if:` condition on the step in the calling workflow:

```yaml
- uses: grafana/shared-workflows/actions/issues-update-project-status@issues-update-project-status/v0.1.0
  if: contains(github.event.issue.labels.*.name, 'area/federal')
  with:
    github-app: grafana-federal-app
    project-id: PVT_kwDOAG3Mbc4AfbLH
    status-field-id: PVTSSF_lADOAG3Mbc4AfbLHzgUxglk
    target-status-option-id: 47fc9ee4
```

## Examples

<!-- x-release-please-start-version -->

```yaml
name: Update project status on assignment
on:
  issues:
    types:
      - assigned

jobs:
  update-project-status:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
    steps:
      - uses: grafana/shared-workflows/actions/issues-update-project-status@issues-update-project-status/v0.2.1
        if: contains(github.event.issue.labels.*.name, 'area/federal')
        with:
          github-app: grafana-federal-app
          project-id: PVT_kwDOAG3Mbc4AfbLH
          status-field-id: PVTSSF_lADOAG3Mbc4AfbLHzgUxglk
          target-status-option-id: 47fc9ee4
```

<!-- x-release-please-end-version -->
