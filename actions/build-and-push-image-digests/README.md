# build-and-push-image-digests

This is a composite GitHub Action, used to build and push image digests to GitHub, so then they can be picked up by the `push-to-gar-docker-multiarch` action, to create a manifest and push the actual image to GAR.

The user is able to see digests appearing on the github action level after every successful run.

```yaml
name: CI
on:
  pull_request:

# These permissions are needed to assume roles from Github's OIDC.
permissions:
  contents: read
  id-token: write

jobs:
  build-and-push:
    runs-on: ubuntu-latest

    steps:
      - id: checkout
        uses: actions/checkout@v4

      - id: build-and-push-image-digests
        uses: grafana/shared-workflows/actions/build-and-push-image-digests@main
        with:
          registry: "<YOUR-GAR>" # e.g. us-docker.pkg.dev, optional
          tags: "<IMAGE_TAG>"
          context: "<YOUR_CONTEXT>" # e.g. "." - where the Dockerfile is
          image_name: "backstage" # name of the image to be published, required
          environment: "dev" # can be either dev/prod
```

## Inputs

| Name                   | Type    | Description                                                                                                                                                                    |
| ---------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `registry`             | String  | Google Artifact Registry to store docker images in.                                                                                                                            |
| `tags`                 | List    | Tags that should be used for the image (see the [metadata-action][mda] for details)                                                                                            |
| `context`              | List    | Path to the Docker build context.                                                                                                                                              |
| `environment`          | Bool    | Environment for pushing artifacts (can be either dev or prod).                                                                                                                 |
| `image_name`           | String  | Name of the image to be pushed to GAR.                                                                                                                                         |
| `build-args`           | String  | List of arguments necessary for the Docker image to be built.                                                                                                                  |
| `push`                 | Boolean | Whether to push the image to the registry.                                                                                                                                     |
| `file`                 | String  | Path and filename of the dockerfile to build from. (Default: `{context}/Dockerfile`)                                                                                           |
| `platforms`            | List    | List of platforms the image should be built for (e.g. `linux/amd64,linux/arm64`)                                                                                               |
| `cache-from`           | String  | Where cache should be fetched from ([more about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/))                 |
| `cache-to`             | String  | Where cache should be stored to ([more about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/))                    |
| `ssh`                  | List    | List of SSH agent socket or keys to expose to the build ([more about ssh for docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs)) |
| `build-contexts`       | List    | List of additional [build contexts](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs) (e.g., `name=path`)                                                 |
| `docker-buildx-driver` | String  | The [driver](https://github.com/docker/setup-buildx-action/tree/v3/?tab=readme-ov-file#customizing) to use for Docker Buildx                                                   |

[mda]: https://github.com/docker/metadata-action?tab=readme-ov-file#tags-input
