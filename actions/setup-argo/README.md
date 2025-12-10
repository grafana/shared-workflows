# Setup Argo

Setup Argo cli and add it to the PATH, this action will pull the binary from GitHub releases and store it in cache for the next run.

## Example

<!-- x-release-please-start-version -->

```yaml
uses: grafana/shared-workflows/actions/setup-argo@setup-argo/v1.0.3
with:
  version: 1.0.3 # Version of the Argo CLI to install.
```

<!-- x-release-please-end-version -->
