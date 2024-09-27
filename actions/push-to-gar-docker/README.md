# push-to-gar-docker

This is a composite GitHub Action, used to push docker images to Google Artifact Registry (GAR).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can
trigger these composite workflows.

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

      - id: push-to-gar
        uses: grafana/shared-workflows/actions/push-to-gar-docker@main
        with:
          registry: "<YOUR-GAR>" # e.g. us-docker.pkg.dev, optional
          tags: |-
            "<IMAGE_TAG>"
            "latest"
          context: "<YOUR_CONTEXT>" # e.g. "." - where the Dockerfile is
          image_name: "backstage" # name of the image to be published, required
          environment: "dev" # can be either dev/prod
```

[Artifact Registry repositories can't contain underscores][underscore-issue].
As a convention, this action will replace any underscores in the repository name
with hyphens. That behaviour can be overridden using the `repository_name`
input.

[underscore-issue]: https://issuetracker.google.com/issues/229159012

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
| `repository_name`      | String  | Override the 'repo_name' which is included as part of the GAR repository name. Only necessary when the GAR includes a repo name that doesn't match the GitHub repo name.       |

[mda]: https://github.com/docker/metadata-action?tab=readme-ov-file#tags-input

## Outputs

The following outputs are exposed from [`docker/metadata-action`](https://github.com/docker/metadata-action?tab=readme-ov-file#outputs) and [`docker/build-push-action`](https://github.com/docker/build-push-action?tab=readme-ov-file#outputs):

| Name          | Type   | Description                    | From                       |
| ------------- | ------ | ------------------------------ | -------------------------- |
| `version`     | String | Generated Docker image version | `docker/metadata-action`   |
| `tags`        | String | Generated Docker tags          | `docker/metadata-action`   |
| `labels`      | String | Generated Docker labels        | `docker/metadata-action`   |
| `annotations` | String | Generated annotations          | `docker/metadata-action`   |
| `json`        | String | JSON output of tags and labels | `docker/metadata-action`   |
| `imageid`     | String | Image ID                       | `docker/build-push-action` |
| `digest`      | String | Image digest                   | `docker/build-push-action` |
| `metadata`    | String | Build result metadata          | `docker/build-push-action` |
