# build-push-to-dockerhub

> [!WARNING]
> This GitHub Action is deprecated and will be removed from main after Jan 31, 2026.
> Please migrate to [docker-build-push-image](https://github.com/grafana/shared-workflows/tree/main/actions/docker-build-push-image#migrating).

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
        uses: grafana/shared-workflows/actions/build-push-to-dockerhub@build-push-to-dockerhub/v0.4.1
        with:
          repository: ${{ github.repository }} # or any other dockerhub repository
          context: .
          tags: |-
            "2024-04-01-abcd1234"
            "latest"
```

<!-- x-release-please-end-version -->

## Inputs

| Name                   | Type    | Description                                                                                                                                                                                      | Default           |
| ---------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------- |
| `build-args`           | String  | List of arguments necessary for the Docker image to be built.                                                                                                                                    |                   |
| `cache-from`           | String  | Where cache should be fetched from ([more about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/))                                   | type=gha          |
| `cache-to`             | String  | Where cache should be stored to ([more about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/))                                      | type=gha,mode=max |
| `context`              | String  | Path to the Docker build context.                                                                                                                                                                | .                 |
| `docker-buildx-driver` | String  | The [driver](https://github.com/docker/setup-buildx-action/tree/v3/?tab=readme-ov-file#customizing) to use for Docker Buildx                                                                     | docker-container  |
| `file`                 | String  | Path and filename of the dockerfile to build from. (Default: `{context}/Dockerfile`)                                                                                                             |                   |
| `labels`               | List    | Labels that should be used for the image (see the [metadata-action][mda] for details)                                                                                                            |                   |
| `load`                 | Boolean | Whether to load the built image into the local docker daemon.                                                                                                                                    | false             |
| `platforms`            | String  | List of platforms the image should be built for (e.g. `linux/amd64,linux/arm64`)                                                                                                                 |                   |
| `push`                 | Boolean | Push the generated image                                                                                                                                                                         | false             |
| `repository`           | String  | Docker repository name (**required**)                                                                                                                                                            |                   |
| `secrets`              | String  | Secrets to [expose to the build](https://github.com/docker/build-push-action). Only needed when authenticating to private repositories outside the repository in which the image is being built. |                   |
| `tags`                 | String  | Tags that should be used for the image (see the [metadata-action][mda] for details)                                                                                                              |                   |
| `target`               | String  | Sets the target stage to build                                                                                                                                                                   |                   |

[mda]: https://github.com/docker/metadata-action?tab=readme-ov-file#tags-input

## Notes

- If you specify `platforms` then the action will use buildx to build the image.
- You must create a Dockerhub repo before you are able to push to it.
- Most projects at Grafana Labs should be using Google Artifact Registry instead
  of Dockerhub to store their images. You can see more about that in the
  [push-to-gar-docker] shared workflow.

[push-to-gar-docker]: ../push-to-gar-docker/README.md
