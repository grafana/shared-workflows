# Setup Argo

> If you are at Grafana Labs, follow these steps in the [internal documentation](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/continuous-delivery/argo-workflows/) first.

Setup Argo cli and add it to the PATH, this action will pull the binary from GitHub releases and store it in cache for the next run.

## Example

<!-- x-release-please-start-version -->

```
uses: grafana/shared-workflows/actions/setup-argo@setup-argo-v1.0.0
with:
  version: 3.5.1 # Version of the Argo CLI to install.

```

<!-- x-release-please-end-version -->
