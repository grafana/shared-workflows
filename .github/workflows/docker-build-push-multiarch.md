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

<!-- BEGIN_INPUTS -->

| Name                          | Type   | Required | Default                    | Description                                                                                                                                                                                                            |
| ----------------------------- | ------ | -------- | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `build-args`                  | string | No       |                            | List of arguments necessary for the Docker image to be built. Passed to `docker/build-push-action`.                                                                                                                    |
| `build-contexts`              | string | No       |                            | List of additional build contexts (e.g., name=path). Passed to `docker/build-push-action`.                                                                                                                             |
| `buildkitd-config`            | string | No       |                            | The buildkitd config file to use. Defaults to `/etc/buildkitd.toml` if you're using Grafana's self-hosted runners. Passed to `docker/setup-buildx-action`.                                                             |
| `buildkitd-config-inline`     | string | No       |                            | The buildkitd inline config to use. Passed to `docker/setup-buildx-action`.                                                                                                                                            |
| `cache-from`                  | string | No       | `type=gha`                 | Where cache should be fetched from. Passed to `docker/build-push-action`.                                                                                                                                              |
| `cache-to`                    | string | No       | `type=gha,mode=max`        | Where cache should be stored to. Passed to `docker/build-push-action`.                                                                                                                                                 |
| `context`                     | string | No       | `.`                        | Path to the Docker build context. Passed to `docker/build-push-action`.                                                                                                                                                |
| `docker-buildx-driver`        | string | No       | `docker-container`         | The driver to use for Docker Buildx. Passed to `docker/setup-buildx-action`.                                                                                                                                           |
| `dockerhub-registry`          | string | No       | `docker.io`                | DockerHub Registry to store docker images in.                                                                                                                                                                          |
| `dockerhub-repository`        | string | No       | `${{ github.repository }}` | DockerHub Repository to store docker images in. Default: github.repository                                                                                                                                             |
| `file`                        | string | No       |                            | The dockerfile to use. Passed to `docker/build-push-action`.                                                                                                                                                           |
| `gar-delete-credentials-file` | string | No       | `true`                     | Delete the Google credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false.                                                                         |
| `gar-environment`             | string | No       | `dev`                      | Environment for pushing artifacts (can be either dev or prod). This sets the GAR Project (gar-project) to either `grafanalabs-dev` or `grafanalabs-global`.                                                            |
| `gar-image`                   | string | No       |                            | Name of the image to build. Default: `${GitHub Repo Name}`.                                                                                                                                                            |
| `gar-registry`                | string | No       | `us-docker.pkg.dev`        | Google Artifact Registry to store docker images in.                                                                                                                                                                    |
| `gar-repository`              | string | No       |                            | Override the 'repo_name' used to construct the GAR repository name. Only necessary when the GAR includes a repo name that doesn't match the GitHub repo name. Default: `docker-${GitHub Repo Name}-${gar-environment}` |
| `generate-summary`            | string | No       | `false`                    | Generates a markdown step summary and sets the OCI_MANIFEST_OUTPUT_JSON env variable and image-digests output after pushing the manifest.                                                                              |
| `include-tags-in-push`        | string | No       | `true`                     | Disables the pushing of tags, and instead includes just a list of images as docker tags. Used when pushing docker digests instead of docker tags.                                                                      |
| `labels`                      | string | No       |                            | List of custom labels to add to the image as metadata (passed to `docker/build-push-action`). Passed to `docker/build-push-action`.                                                                                    |
| `load`                        | string | No       | `false`                    | Whether to load the built image into the local docker daemon (passed to `docker/build-push-action`). Passed to `docker/build-push-action`.                                                                             |
| `outputs`                     | string | No       |                            | List of docker output destinations. Passed to `docker/build-push-action`.                                                                                                                                              |
| `platforms`                   | string | No       |                            | List of platforms to build the image for. Passed to `docker/build-push-action`.                                                                                                                                        |
| `push`                        | string | No       |                            | Whether to push the image to the configured registries. Passed to `docker/build-push-action`.                                                                                                                          |
| `registries`                  | string | No       |                            | CSV list of registries to build images for. Accepted registries are "gar" and "dockerhub".                                                                                                                             |
| `runner-type`                 | string | No       | `self-hosted`              | Setting this flag will dictate the default instance types to use. Allowed values are 'self-hosted' or 'github'.                                                                                                        |
| `runner-type-arm64`           | string | No       |                            | The instance type to use for arm64 builds.                                                                                                                                                                             |
| `runner-type-manifest`        | string | No       |                            | The instance type to use when building and pushing the manifest.                                                                                                                                                       |
| `runner-type-x64`             | string | No       |                            | The instance type to use for x64 builds.                                                                                                                                                                               |
| `secrets`                     | string | No       |                            | Secrets to expose to the build. Only needed when authenticating to private repositories outside the repository in which the image is being built. Passed to `docker/build-push-action`.                                |
| `ssh`                         | string | No       |                            | List of SSH agent socket or keys to expose to the build Passed to `docker/build-push-action`.                                                                                                                          |
| `tags`                        | string | Yes      |                            | List of Docker tags to be pushed. Passed to `docker/build-push-action`.                                                                                                                                                |
| `target`                      | string | No       |                            | Sets the target stage to build. Passed to `docker/build-push-action`.                                                                                                                                                  |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name                       | Description                                                                                                                      |
| -------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `annotations`              | Generated annotations (from docker/metadata-action)                                                                              |
| `digest`                   | Image digest (from docker/build-push-action)                                                                                     |
| `image-digests`            | Newline-separated list of image digests in the format &lt;image>:&lt;tag>@&lt;digest> (from docker-import-digests-push-manifest) |
| `imageid`                  | Image ID (from docker/build-push-action)                                                                                         |
| `images`                   | Comma separated list of the images that were built                                                                               |
| `json`                     | JSON output of tags and labels (from docker/metadata-action)                                                                     |
| `labels`                   | Generated Docker labels (from docker/metadata-action)                                                                            |
| `metadata`                 | Build result metadata (from docker/build-push-action)                                                                            |
| `oci-manifest-output-json` | JSON array of manifests with tag, indexDigest, and per-platform digest information (from docker-import-digests-push-manifest)    |
| `runner-arches`            | The list of OS used to build images (for mapping to self hosted runners)                                                         |
| `tags`                     | Generated Docker tags (from docker/metadata-action)                                                                              |
| `version`                  | Generated Docker image version (from docker/metadata-action)                                                                     |

<!-- END_OUTPUTS -->
