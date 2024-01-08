# push-to-dockerhub

This is a composite GitHub Action, used to push docker images to DockerHub.
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
      - id: push-to-dockerhub
        uses: grafana/shared-workflows/actions/push-to-dockerhub@main
        with:
          repository: ${{ github.repository }} # or any other dockerhub repository
          build_path: .
          tags: |-
            "2024-04-01-abcd1234"
            "latest"
```
