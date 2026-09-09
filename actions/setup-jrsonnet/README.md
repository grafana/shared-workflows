# Setup jrsonnet

Setup jrsonnet CLI and add it to the PATH, this action will pull the binary from GitHub releases and store it in cache for the next run.

## Example

<!-- x-release-please-start-version -->

```yaml
uses: grafana/shared-workflows/actions/setup-jrsonnet@setup-jrsonnet/v1.0.3
with:
  version: 1.0.3 # Version of the jrsonnet CLI to install.
```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name           | Type   | Required | Default            | Description                             |
| -------------- | ------ | -------- | ------------------ | --------------------------------------- |
| `cache-prefix` | String | No       | `jrsonnet`         | Prefix for the cache key.               |
| `version`      | String | No       | `0.5.0-pre96-test` | Version of the jrsonnet CLI to install. |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name        | Description                       |
| ----------- | --------------------------------- |
| `cache-hit` | Whether the cache was hit or not. |

<!-- END_OUTPUTS -->
