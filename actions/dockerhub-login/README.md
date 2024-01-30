# dockerhub-login

This is a composite GitHub Action, used to login to the grafanabot DockerHub account
It uses `get-vault-secrets` action to get the DockerHub username and password from Vault.

Example of how to use this action in a repository:

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
        uses: grafana/shared-workflows/actions/dockerhub-login@main
      - name: Build and push
        run: make build && make push
```
