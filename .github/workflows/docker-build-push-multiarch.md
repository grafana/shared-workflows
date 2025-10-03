# docker-build-push-multiaarch

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
There is a then a final job that runs the composite action [docker-import-digests-push-manifest] to push the final
docker manifest.

[docker/build-push-action]: https://github.com/docker/build-push-action
[docker-build-push-image]: ../../docker-build-push-image/README.md
[docker-export-digest]: ../../docker-export-digest/README.md
[docker-import-digests-push-manifest]: ../../docker-import-digests-push-manifest/README.md

<!-- x-release-please-start-version -->

```yaml
name: Build and Push and Push MultiArch

on: push

jobs:
  build-push-multiarch:
    uses: grafana/shared-workflows/.github/workflows/build-and-push-docker-multiarch.yml@rwhitaker/multi-arch-builds # TODO: Pin to version
    with:
      platforms: linux/arm64,linux/amd64
      tags: |
        ${{ github.sha }}
        rickytest
      push: true
      registries: "gar,dockerhub"
      pre-build-script: scripts/ci-build.sh

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
| `docker-buildx-driver`    | String  | The driver to use for Docker Buildx                                                                                                                                                                                    |
| `dockerhub-repository`    | String  | Ipsum dockerhubium                                                                                                                                                                                                     |
| `file`                    | String  | The dockerfile to use.                                                                                                                                                                                                 |
| `gar-environment`         | String  | Environment for pushing artifacts (can be either dev or prod).                                                                                                                                                         |
| `gar-image`               | String  | Name of the image to build. Default: `${GitHub Repo Name}`                                                                                                                                                             |
| `gar-registry`            | String  | Google Artifact Registry to store docker images in.                                                                                                                                                                    |
| `gar-repository`          | String  | Override the 'repo_name' used to construct the GAR repository name. Only necessary when the GAR includes a repo name that doesn't match the GitHub repo name. Default: `docker-${GitHub Repo Name}-${gar-environment}` |
| `labels`                  | String  | List of custom labels to add to the image as metadata.                                                                                                                                                                 |
| `load`                    | Boolean | Whether to load the built image into the local docker daemon.                                                                                                                                                          |
| `outputs`                 | String  | Ipsum factum explainum.                                                                                                                                                                                                |
| `platforms`               | String  | List of platforms to build the image for                                                                                                                                                                               |
| `post-build-script`       | String  | A script to run after docker build                                                                                                                                                                                     |
| `pre-build-script`        | String  | A script to run before docker build                                                                                                                                                                                    |
| `push`                    | String  | Whether to push the image to the configured registries.                                                                                                                                                                |
| `registries`              | String  | List of registries to build images for.                                                                                                                                                                                |
| `secrets`                 | String  | Secrets to expose to the build. Only needed when authenticating to private repositories outside the repository in which the image is being built.                                                                      |
| `server-size`             | String  | Size of the hosted runner                                                                                                                                                                                              |
| `ssh`                     | String  | List of SSH agent socket or keys to expose to the build                                                                                                                                                                |
| `tags`                    | String  | List of Docker tags to be pushed.                                                                                                                                                                                      |
| `target`                  | String  | Sets the target stage to build                                                                                                                                                                                         |

## Outputs

| Name            | Type   | Description                                                              |
|-----------------|--------|--------------------------------------------------------------------------|
| `annotations`   | String | Generated annotations (from docker/metadata-action)                      |
| `digest`        | String | Image digest (from docker/build-push-action)                             |
| `imageid`       | String | Image ID (from docker/build-push-action)                                 |
| `images`        | String | Comma separated list of the images that were built                       |
| `json`          | String | JSON output of tags and labels (from docker/metadata-action)             |
| `labels`        | String | Generated Docker labels (from docker/metadata-action)                    |
| `metadata`      | String | Build result metadata (from docker/build-push-action)                    |
| `metadatajson`  | String | Metadata JSON (from docker/metadata)                                     |
| `runner_arches` | String | The list of OS used to build images (for mapping to self hosted runners) |
| `tags`          | String | Generated Docker tags (from docker/metadata-action)                      |
| `version`       | String | Generated Docker image version (from docker/metadata-action)             |
