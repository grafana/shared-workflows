# docker-build-push-image

This is a composite GitHub Action, used to build and push docker images to private Grafana registries.
It builds registry URLs for Grafana's registries, authenticates to them, and then
uses [docker/build-push-action](https://github.com/docker/build-push-action) to build and push the image(s).

# TODO: do we need QEMU?

<!-- x-release-please-start-version -->

```yaml
name: Build a Docker Image

on:
  push:
    branches:
      - main

jobs:
  build-push-image:
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: grafana/shared-workflows/actions/docker-build-push-image@main # TODO: Fix version once released
        with:
          platforms: linux/arm64,linux/amd64
          tags: |
            ${{ github.sha }}
            main
          push: true
          registries: "gar,dockerhub"
```

<!-- x-release-please-end-version -->

## Inputs

| Name                      | Type    | Description                                                                                                                                                                                                            |
|---------------------------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `build-args`              | String  | List of arguments necessary for the Docker image to be built.                                                                                                                                                          |
| `build-contexts`          | String  | List of additional build contexts (e.g., name=path)                                                                                                                                                                    |
| `buildkitd-config`        | String  | Configum buildkitium descriptium                                                                                                                                                                                       |
| `buildkitd-config-inline` | String  | Inliumium configutorium buildkitium descriptium                                                                                                                                                                        |
| `cache-from`              | String  | Where cache should be fetched from                                                                                                                                                                                     |
| `cache-to`                | String  | Where cache should be stored to                                                                                                                                                                                        |
| `context`                 | String  | Path to the Docker build context.                                                                                                                                                                                      |
| `delete-credentials-file` | Boolean | Delete the credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false.                                                                                |
| `dockerhub-repository`    | String  | Ipsum dockerhubium                                                                                                                                                                                                     |
| `docker-buildx-driver`    | String  | The driver to use for Docker Buildx                                                                                                                                                                                    |
| `file`                    | String  | The dockerfile to use.                                                                                                                                                                                                 |
| `gar-registry`            | String  | Google Artifact Registry to store docker images in.                                                                                                                                                                    |
| `gar-repository`          | String  | Override the 'repo_name' used to construct the GAR repository name. Only necessary when the GAR includes a repo name that doesn't match the GitHub repo name. Default: `docker-${GitHub Repo Name}-${gar-environment}` |
| `gar-environment`         | String  | Environment for pushing artifacts (can be either dev or prod).                                                                                                                                                         |
| `gar-image`               | String  | Name of the image to build. Default: `${GitHub Repo Name}`                                                                                                                                                             |
| `include-tags-in-push`    | Boolean | Can't push tags with a digest, need a way to disable                                                                                                                                                                   |
| `labels`                  | String  | List of custom labels to add to the image as metadata.                                                                                                                                                                 |
| `load`                    | Boolean | Whether to load the built image into the local docker daemon.                                                                                                                                                          |
| `outputs`                 | String  | Ipsum factum explainum.                                                                                                                                                                                                |
| `platforms`               | String  | List of platforms to build the image for                                                                                                                                                                               |
| `push`                    | String  | Whether to push the image to the configured registries.                                                                                                                                                                |
| `registries`              | String  | List of registries to build images for.                                                                                                                                                                                |
| `secrets`                 | String  | Secrets to expose to the build. Only needed when authenticating to private repositories outside the repository in which the image is being built.                                                                      |
| `ssh`                     | String  | List of SSH agent socket or keys to expose to the build                                                                                                                                                                |
| `tags`                    | String  | List of Docker tags to be pushed.                                                                                                                                                                                      |
| `target`                  | String  | Sets the target stage to build                                                                                                                                                                                         |

## Outputs

| Name           | Type   | Description                                                  |
|----------------|--------|--------------------------------------------------------------|
| `annotations`  | String | Generated annotations (from docker/metadata-action)          |
| `digest`       | String | Image digest (from docker/build-push-action)                 |
| `imageid`      | String | Image ID (from docker/build-push-action)                     |
| `images`       | String | Comma separated list of the images that were built           |
| `json`         | String | JSON output of tags and labels (from docker/metadata-action) |
| `labels`       | String | Generated Docker labels (from docker/metadata-action)        |
| `metadata`     | String | Build result metadata (from docker/build-push-action)        |
| `metadatajson` | String | Metadata JSON (from docker/metadata)                         |
| `version`      | String | Generated Docker image version (from docker/metadata-action) |
| `tags`         | String | Generated Docker tags (from docker/metadata-action)          |
