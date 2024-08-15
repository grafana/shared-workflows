# Setup Conftest

Setup conftest and add it to the PATH, this action will pull the binary from GitHub releases and store it in cache for the next run.

## Example

```
uses: grafana/shared-workflows/actions/setup-conftest@main
with:
  version: 0.55.0 # Version of conftest to install.

```

