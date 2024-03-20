# build-push-to-dockerhub

This is a composite GitHub Action, used to build Docker images and push them to DockerHub.
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
        uses: grafana/shared-workflows/actions/build-push-to-dockerhub@main
        with:
          repository: ${{ github.repository }} # or any other dockerhub repository
          context: .
          tags: |-
            "2024-04-01-abcd1234"
            "latest"
```

## Inputs

| Name         | Type   | Description                                                                          |
|--------------|--------|--------------------------------------------------------------------------------------|
| `context`    | String | Path to the Dockerfile (default: `.`)                                                |
| `platforms`  | List   | List of platforms the image should be built for (e.g. `linux/amd64,linux/arm64`)     |
| `push`       | Bool   | Push the generated image (default: `false`)                                          |
| `repository` | String | Docker repository name                                                               |
| `tags`       | List   | Tags that should be used for the image (see the [metadata-action][mda] for details)  |
| `file`       | String | Path and filename of the dockerfile to build from. (Default: `{context}/Dockerfile`) |

[mda]: https://github.com/docker/metadata-action?tab=readme-ov-file#tags-input


## Notes

- If you specify `platforms` then the action will use buildx to build the image.
- You must create a Dockerhub repo before you are able to push to it.
- Most projects should be using Google Artifact Registry (instead of Dockerhub) to store their images. You can see more about that in the push-to-gar-docker shared workflow.
