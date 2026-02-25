# docker-build-push-multiarch

This is a reusable workflow that uses Grafana's hosted runners to natively build and push multi-architecture docker
images.

Right now this supports pushing images to:

- Google Artifact Registry
- DockerHub

And supports building the following image types:

- linux/arm64
- linux/amd64

## How it works

This generates a matrix based off of the `platforms` input, then creates a job per platform that runs the composite
actions [docker-build-push-image] and [docker-export-digest] to build and push docker images, and capture their digests.
There is then a final job that runs the composite action [docker-import-digests-push-manifest] to push the docker
manifest.

[docker/build-push-action]: https://github.com/docker/build-push-action
[docker-build-push-image]: ../../docker-build-push-image/README.md
[docker-export-digest]: ../../docker-export-digest/README.md
[docker-import-digests-push-manifest]: ../../docker-import-digests-push-manifest/README.md

```yaml
name: Build and Push and Push MultiArch

on: push

jobs:
  build-push-multiarch:
    uses: grafana/shared-workflows/.github/workflows/docker-build-push-multiarch@6b59374893555bf476179dfeb96013b80406102f # main
    with:
      platforms: linux/arm64,linux/amd64
      tags: |
        ${{ github.sha }}
        latest
      push: true
      registries: "gar,dockerhub"
```

## Inputs

| Name                          | Type   | Description                                                                                                                                                                                                            |
| ----------------------------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `build-args`                  | string | List of arguments necessary for the Docker image to be built. Passed to `docker/build-push-action`.                                                                                                                    |
| `build-contexts`              | string | List of additional build contexts (e.g., name=path). Passed to `docker/build-push-action`.                                                                                                                             |
| `buildkitd-config`            | string | The buildkitd config file to use. Defaults to `/etc/buildkitd.toml` if you're using Grafana's self-hosted runners. Passed to `docker/setup-buildx-action`.                                                             |
| `buildkitd-config-inline`     | string | The buildkitd inline config to use. Passed to `docker/setup-buildx-action`.                                                                                                                                            |
| `cache-from`                  | string | Where cache should be fetched from. Passed to `docker/build-push-action`.                                                                                                                                              |
| `cache-to`                    | string | Where cache should be stored to. Passed to `docker/build-push-action`.                                                                                                                                                 |
| `context`                     | string | Path to the Docker build context. Passed to `docker/build-push-action`.                                                                                                                                                |
| `docker-buildx-driver`        | string | The driver to use for Docker Buildx. Passed to `docker/setup-buildx-action`.                                                                                                                                           |
| `dockerhub-registry`          | string | DockerHub Registry to store docker images in.                                                                                                                                                                          |
| `dockerhub-repository`        | string | DockerHub Repository to store docker images in. Default: github.repository                                                                                                                                             |
| `file`                        | string | The dockerfile to use. Passed to `docker/build-push-action`.                                                                                                                                                           |
| `gar-delete-credentials-file` | string | Delete the Google credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false.                                                                         |
| `gar-environment`             | string | Environment for pushing artifacts (can be either dev or prod). This sets the GAR Project (gar-project) to either `grafanalabs-dev` or `grafanalabs-global`.                                                            |
| `gar-image`                   | string | Name of the image to build. Default: `${GitHub Repo Name}`.                                                                                                                                                            |
| `gar-registry`                | string | Google Artifact Registry to store docker images in.                                                                                                                                                                    |
| `gar-repository`              | string | Override the 'repo_name' used to construct the GAR repository name. Only necessary when the GAR includes a repo name that doesn't match the GitHub repo name. Default: `docker-${GitHub Repo Name}-${gar-environment}` |
| `include-tags-in-push`        | string | Disables the pushing of tags, and instead includes just a list of images as docker tags. Used when pushing docker digests instead of docker tags.                                                                      |
| `labels`                      | string | List of custom labels to add to the image as metadata (passed to `docker/build-push-action`). Passed to `docker/build-push-action`.                                                                                    |
| `load`                        | string | Whether to load the built image into the local docker daemon (passed to `docker/build-push-action`). Passed to `docker/build-push-action`.                                                                             |
| `outputs`                     | string | List of docker output destinations. Passed to `docker/build-push-action`.                                                                                                                                              |
| `platforms`                   | string | List of platforms to build the image for. Passed to `docker/build-push-action`.                                                                                                                                        |
| `push`                        | string | Whether to push the image to the configured registries. Passed to `docker/build-push-action`.                                                                                                                          |
| `registries`                  | string | CSV list of registries to build images for. Accepted registries are "gar" and "dockerhub".                                                                                                                             |
| `runner-type`                 | string | Setting this flag will dictate the default instance types to use. If runner-type-x64, runner-type-arm64, and runner-type-manifest are all set then this value is superseded because no defaults will be used.          |
| `runner-type-arm64`           | string | The instance type to use for arm64 builds.                                                                                                                                                                             |
| `runner-type-manifest`        | string | The instance type to use when building and pushing the manifest.                                                                                                                                                       |
| `runner-type-x64`             | string | The instance type to use for x64 builds.                                                                                                                                                                               |
| `generate-summary`            | string | Generates a markdown step summary and sets the `OCI_MANIFEST_OUTPUT_JSON` env variable and `image-digests` output after pushing the manifest. Default: `false`.                                                        |
| `secrets`                     | string | Secrets to expose to the build. Only needed when authenticating to private repositories outside the repository in which the image is being built. Passed to `docker/build-push-action`.                                |
| `ssh`                         | string | List of SSH agent socket or keys to expose to the build Passed to `docker/build-push-action`.                                                                                                                          |
| `tags`                        | string | List of Docker tags to be pushed. Passed to `docker/build-push-action`.                                                                                                                                                |
| `target`                      | string | Sets the target stage to build. Passed to `docker/build-push-action`.                                                                                                                                                  |

## Outputs

| Name            | Type   | Description                                                                                                        |
| --------------- | ------ | ------------------------------------------------------------------------------------------------------------------ |
| `annotations`   | String | Generated annotations (from docker/metadata-action)                                                                |
| `digest`        | String | Image digest (from docker/build-push-action)                                                                       |
| `imageid`       | String | Image ID (from docker/build-push-action)                                                                           |
| `images`        | String | Comma separated list of the images that were built                                                                 |
| `json`          | String | JSON output of tags and labels (from docker/metadata-action)                                                       |
| `labels`        | String | Generated Docker labels (from docker/metadata-action)                                                              |
| `metadata`      | String | Build result metadata (from docker/build-push-action)                                                              |
| `runner_arches` | String | The list of OS used to build images (for mapping to self hosted runners)                                           |
| `image-digests` | String | Newline-separated list of image digests in the format `<image>:<tag>@<digest>` (requires `generate-summary: true`) |
| `tags`          | String | Generated Docker tags (from docker/metadata-action)                                                                |
| `version`       | String | Generated Docker image version (from docker/metadata-action)                                                       |
