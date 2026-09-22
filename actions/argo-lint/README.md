# Lint Argo

Shared workflow to lint Argo workflow files.

## Example

<!-- x-release-please-start-version -->

```
uses: grafana/shared-workflows/actions/argo-lint@argo-lint/v1.1.1
with:
  path: /path/to/files # Paths to files for linting

```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name           | Type   | Required | Default  | Description                                 |
| -------------- | ------ | -------- | -------- | ------------------------------------------- |
| `argo-version` | String | No       | `3.7.10` | Version of the Argo CLI to use for linting. |
| `path`         | String | Yes      |          | Path to files for linting.                  |

<!-- END_INPUTS -->
