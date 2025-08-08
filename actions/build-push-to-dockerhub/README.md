# build-push-to-dockerhub

> [!NOTE]
> If you are at Grafana Labs:
>
> - A docker mirror is available on our self-hosted runners, see [the internal
>   documentation](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/continuous-integration/#docker-caching-in-github-actions)
>   for more info.

This is a composite GitHub Action, used to build Docker images and push them to
DockerHub. It uses `get-vault-secrets` action to get the DockerHub username and
password from Vault.

Example of how to use this action in a repository:

<!-- x-release-please-start-version -->

```yaml
name: Push to DockerHub
on:
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
    steps:
      - id: checkout
        uses: actions/checkout@v4
        with:
          persist-credentials: false

      - id: push-to-dockerhub
        uses: grafana/shared-workflows/actions/build-push-to-dockerhub@build-push-to-dockerhub/v0.3.0
        with:
          repository: ${{ github.repository }} # or any other dockerhub repository
          context: .
          tags: |-
            "2024-04-01-abcd1234"
            "latest"
```

<!-- x-release-please-end-version -->

## Inputs

| Name                   | Type   | Description                                                                                                                                                                                      |
| ---------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `context`              | String | Path to the Dockerfile (default: `.`)                                                                                                                                                            |
| `platforms`            | List   | List of platforms the image should be built for (e.g. `linux/amd64,linux/arm64`)                                                                                                                 |
| `push`                 | Bool   | Push the generated image (default: `false`)                                                                                                                                                      |
| `repository`           | String | Docker repository name (**required**)                                                                                                                                                            |
| `tags`                 | List   | Tags that should be used for the image (see the [metadata-action][mda] for details)                                                                                                              |
| `file`                 | String | Path and filename of the dockerfile to build from. (Default: `{context}/Dockerfile`)                                                                                                             |
| `build-args`           | String | List of arguments necessary for the Docker image to be built.                                                                                                                                    |
| `target`               | String | Sets the target stage to build                                                                                                                                                                   |
| `cache-from`           | String | Where cache should be fetched from ([more about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/))                                   |
| `cache-to`             | String | Where cache should be stored to ([more about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/))                                      |
| `docker-buildx-driver` | String | The [driver](https://github.com/docker/setup-buildx-action/tree/v3/?tab=readme-ov-file#customizing) to use for Docker Buildx                                                                     |
| `secrets`              | List   | Secrets to [expose to the build](https://github.com/docker/build-push-action). Only needed when authenticating to private repositories outside the repository in which the image is being built. |

[mda]: https://github.com/docker/metadata-action?tab=readme-ov-file#tags-input

## Notes

- If you specify `platforms` then the action will use buildx to build the image.
- You must create a Dockerhub repo before you are able to push to it.
- Most projects at Grafana Labs should be using Google Artifact Registry instead
  of Dockerhub to store their images. You can see more about that in the
  [push-to-gar-docker] shared workflow.

[push-to-gar-docker]: ../push-to-gar-docker/README.md
