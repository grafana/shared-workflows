# Setup Argo

Setup Argo cli and add it to the PATH, this action will pull the binary from GitHub releases and store it in cache for the next run.

## Example

<!-- x-release-please-start-version -->

```yaml
uses: grafana/shared-workflows/actions/setup-argo@setup-argo/v1.2.1
with:
  version: 1.2.1 # Version of the Argo CLI to install.
```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name           | Type   | Required | Default | Description                         |
| -------------- | ------ | -------- | ------- | ----------------------------------- |
| `cache-prefix` | String | No       | `argo`  | Prefix for the cache key.           |
| `version`      | String | No       | `4.0.3` | Version of the Argo CLI to install. |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name        | Description                       |
| ----------- | --------------------------------- |
| `cache-hit` | Whether the cache was hit or not. |

<!-- END_OUTPUTS -->
