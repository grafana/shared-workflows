# Setup jrsonnet

Setup jrsonnet CLI and add it to the PATH, this action will pull the binary from GitHub releases and store it in cache for the next run.

## Example

<!-- x-release-please-start-version -->

```yaml
uses: grafana/shared-workflows/actions/setup-jrsonnet@setup-jrsonnet/v1.0.1
with:
  version: 1.0.1 # Version of the jrsonnet CLI to install.
```

<!-- x-release-please-end-version -->
