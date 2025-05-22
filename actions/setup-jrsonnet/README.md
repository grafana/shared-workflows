# Setup jrsonnet

Setup jrsonnet CLI and add it to the PATH, this action will pull the binary from GitHub releases and store it in cache for the next run.

## Example

<!-- x-release-please-start-version -->

```yaml
uses: grafana/shared-workflows/actions/setup-jrsonnet@setup-jrsonnet-v1.1.0
with:
  version: 1.1.0 # Version of the jrsonnet CLI to install.
```

<!-- x-release-please-end-version -->
