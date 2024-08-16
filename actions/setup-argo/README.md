# Setup Argo

Setup Argo cli and add it to the PATH, this action will pull the binary from GitHub releases and store it in cache for the next run.

## Example

```
uses: grafana/shared-workflows/actions/setup-argo@main
with:
  version: 3.5.1 # Version of the Argo CLI to install.

```
