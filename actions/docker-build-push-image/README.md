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

| Name                          | Type    | Description                                                                                                                                                                                                            |
|-------------------------------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `build-args`                  | String  | List of arguments necessary for the Docker image to be built. Passed to `docker/build-push-action`.                                                                                                                    |
| `build-contexts`              | String  | List of additional build contexts (e.g., name=path). Passed to `docker/build-push-action`.                                                                                                                             |
| `buildkitd-config`            | String  | The buildkitd config file to use. Defaults to `/etc/buildkitd.toml` if you're using Grafana's self-hosted runners. Passed to `docker/setup-buildx-action`.                                                             |
| `buildkitd-config-inline`     | String  | The buildkitd inline config to use. Passed to `docker/setup-buildx-action`.                                                                                                                                            |
| `cache-from`                  | String  | Where cache should be fetched from. Passed to `docker/build-push-action`.                                                                                                                                              |
| `cache-to`                    | String  | Where cache should be stored to. Passed to `docker/build-push-action`.                                                                                                                                                 |
| `context`                     | String  | Path to the Docker build context. Passed to `docker/build-push-action`.                                                                                                                                                |
| `docker-buildx-driver`        | String  | The driver to use for Docker Buildx. Passed to `docker/setup-buildx-action`.                                                                                                                                           |
| `dockerhub-repository`        | String  | Ipsum dockerhubium                                                                                                                                                                                                     |
| `file`                        | String  | The dockerfile to use. Passed to `docker/build-push-action`.                                                                                                                                                           |
| `gar-delete-credentials-file` | Boolean | Delete the Google credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false.                                                                         |
| `gar-environment`             | String  | Environment for pushing artifacts (can be either dev or prod). This sets the GAR Project (gar-project) to either `grafanalabs-dev` or `grafanalabs-global`.                                                            |
| `gar-image`                   | String  | Name of the image to build. Default: `${GitHub Repo Name}`.                                                                                                                                                            |
| `gar-registry`                | String  | Google Artifact Registry to store docker images in.                                                                                                                                                                    |
| `gar-repository`              | String  | Override the 'repo_name' used to construct the GAR repository name. Only necessary when the GAR includes a repo name that doesn't match the GitHub repo name. Default: `docker-${GitHub Repo Name}-${gar-environment}` |
| `include-tags-in-push`        | Boolean | Disables the pushing of tags, and instead includes just a list of images as docker tags. Used when pushing docker digests instead of docker tags.                                                                      |
| `labels`                      | String  | List of custom labels to add to the image as metadata (passed to `docker/build-push-action`). Passed to `docker/build-push-action`.                                                                                    |
| `load`                        | Boolean | Whether to load the built image into the local docker daemon (passed to `docker/build-push-action`). Passed to `docker/build-push-action`.                                                                             |
| `outputs`                     | String  | List of docker output destinations. Passed to `docker/build-push-action`.                                                                                                                                              |
| `platforms`                   | String  | List of platforms to build the image for. Passed to `docker/build-push-action`.                                                                                                                                        |
| `push`                        | String  | Whether to push the image to the configured registries. Passed to `docker/build-push-action`.                                                                                                                          |
| `registries`                  | String  | CSV list of registries to build images for. Accepted registries are "gar" and "dockerhub".                                                                                                                             |
| `secrets`                     | String  | Secrets to expose to the build. Only needed when authenticating to private repositories outside the repository in which the image is being built. Passed to `docker/build-push-action`.                                |
| `ssh`                         | String  | List of SSH agent socket or keys to expose to the build Passed to `docker/build-push-action`.                                                                                                                          |
| `tags`                        | String  | List of Docker tags to be pushed. Passed to `docker/build-push-action`.                                                                                                                                                |
| `target`                      | String  | Sets the target stage to build. Passed to `docker/build-push-action`.                                                                                                                                                  |

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
| `tags`         | String | Generated Docker tags (from docker/metadata-action)          |
| `version`      | String | Generated Docker image version (from docker/metadata-action) |


## How we construct Google Artifact Registry Images

The full GAR image is constructed as follows, where `gar-project` is determined by `inputs.gar-environment`.

"${{ inputs.gar-registry }}/${{ gar-project }}/${{ inputs.gar-repository }}/${{ inputs.gar-image }}"

## How we construct DockerHub Images

The full DockerHub image is constructed as follows:

