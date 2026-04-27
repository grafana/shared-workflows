# issues-update-project-status

Updates a GitHub Project (v2) issue status field when triggered by an issue event.

The calling job must have `id-token: write` permission for Vault authentication.

## Inputs

| Name                       | Type   | Description                                                                                                  | Default | Required |
| -------------------------- | ------ | ------------------------------------------------------------------------------------------------------------ | ------- | -------- |
| `app-id-vault-secret`      | String | Vault common secret path for the GitHub App ID (e.g. `grafana-federal-app:app-id`)                           |         | Yes      |
| `private-key-vault-secret` | String | Vault common secret path for the GitHub App private key (e.g. `grafana-federal-app:private-key`)             |         | Yes      |
| `project-id`               | String | Node ID of the GitHub Project (v2). Retrieve with: `gh project view <projNum> --owner <org> --format json`   |         | Yes      |
| `status-field-id`          | String | Node ID of the Status field. Retrieve with: `gh project field-list <projNum> --owner <org> --format json`    |         | Yes      |
| `target-status-option-id`  | String | ID of the status option to set. Retrieve with: `gh project field-list <projNum> --owner <org> --format json` |         | Yes      |

## Filtering by label

To limit which issues trigger a status update, use an `if:` condition on the step in the calling workflow:

```yaml
- uses: grafana/shared-workflows/actions/issues-update-project-status@issues-update-project-status/v0.1.0
  if: contains(github.event.issue.labels.*.name, 'area/federal')
  with:
    app-id-vault-secret: grafana-federal-app:app-id
    private-key-vault-secret: grafana-federal-app:private-key
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
      - uses: grafana/shared-workflows/actions/issues-update-project-status@issues-update-project-status/v0.1.1
        if: contains(github.event.issue.labels.*.name, 'area/federal')
        with:
          app-id-vault-secret: grafana-federal-app:app-id
          private-key-vault-secret: grafana-federal-app:private-key
          project-id: PVT_kwDOAG3Mbc4AfbLH
          status-field-id: PVTSSF_lADOAG3Mbc4AfbLHzgUxglk
          target-status-option-id: 47fc9ee4
```

<!-- x-release-please-end-version -->
