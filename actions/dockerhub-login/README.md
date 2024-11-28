# dockerhub-login

This is a composite GitHub Action, used to login to the grafanabot DockerHub account
It uses `get-vault-secrets` action to get the DockerHub username and password from Vault.

Example of how to use this action in a repository:

<!-- x-release-please-start-version -->

```yaml
name: Push to DockerHub
on:
  pull_request:

permissions:
  contents: read
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - name: Login to DockerHub
        uses: grafana/shared-workflows/actions/dockerhub-login@dockerhub-login-v1.0.0
      - name: Build and push
        run: make build && make push
```

<!-- x-release-please-end-version -->
