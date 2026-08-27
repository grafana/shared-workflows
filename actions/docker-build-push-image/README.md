# docker-build-push-image

This is a composite GitHub Action, used to build and push docker images to private Grafana registries.
It builds registry URLs for Grafana's registries, authenticates to them, and then
uses [docker/build-push-action] to build and push the image(s).

This action can work 1 of 2 ways:

1. It can be run on a single runner, and if multiple `platforms` are configured then buildx/QEMU emulation is used.
2. It can be used in conjunction with [docker-export-digest] and [docker-import-digests-push-manifest] to push untagged
   images whose digests are later exported and merged into a tagged docker manifest. For true multi-arch builds.

This can push to the following registries:

1. Google Artifact Registry
2. DockerHub

> [!WARNING]
> There is a [bug with Google Artifact Registry](https://issuetracker.google.com/issues/390719013?pli=1) that prevents docker images from pushing successfully if the
> branch name is too long. This is due a max length with Workload Identity Federation claims on Google's side.
>
> Should you use this action and get a `400` error when pushing the image... try shortening the branch name.

[docker/build-push-action]: https://github.com/docker/build-push-action
[docker-build-push-image]: ../docker-build-push-image/README.md
[docker-export-digest]: ../docker-export-digest/README.md
[docker-import-digests-push-manifest]: ../docker-import-digests-push-manifest/README.md

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
      - uses: grafana/shared-workflows/actions/docker-build-push-image@docker-build-push-image/v0.3.4
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

<!-- BEGIN_INPUTS -->

| Name                          | Type    | Required | Default                    | Description                                                                                                                                                                                                                                                                              |
| ----------------------------- | ------- | -------- | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `annotations`                 | String  | No       |                            | List of custom annotations to add to the image as metadata (passed to `docker/build-push-action`). Passed to `docker/build-push-action`.                                                                                                                                                 |
| `build-args`                  | String  | No       |                            | List of arguments necessary for the Docker image to be built. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                               |
| `build-contexts`              | String  | No       |                            | List of additional build contexts (e.g., name=path). Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                                        |
| `builder`                     | String  | No       |                            | Name of the buildx builder to use. If not specified, a new builder will be created. This is useful when you need to reuse a builder, for example with buildkit-cache-dance.                                                                                                              |
| `buildkitd-config`            | String  | No       |                            | The buildkitd config file to use. Defaults to `/etc/buildkitd.toml` if you're using Grafana's self-hosted runners. Passed to [docker/setup-buildx-action](https://github.com/docker/setup-buildx-action?tab=readme-ov-file#inputs).                                                      |
| `buildkitd-config-inline`     | String  | No       |                            | The buildkitd inline config to use. Passed to [docker/setup-buildx-action](https://github.com/docker/setup-buildx-action?tab=readme-ov-file#inputs).                                                                                                                                     |
| `cache-from`                  | String  | No       | `type=gha`                 | Where cache should be fetched from. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs). [More about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/)               |
| `cache-to`                    | String  | No       | `type=gha,mode=max`        | Where cache should be stored to. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs). [More about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/)                  |
| `context`                     | String  | No       | `.`                        | Path to the Docker build context. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                                                           |
| `docker-buildx-driver`        | String  | No       | `docker-container`         | The driver to use for Docker Buildx. Passed to [docker/setup-buildx-action](https://github.com/docker/setup-buildx-action?tab=readme-ov-file#inputs).                                                                                                                                    |
| `dockerhub-registry`          | String  | No       | `docker.io`                | DockerHub Registry to store docker images in.                                                                                                                                                                                                                                            |
| `dockerhub-repository`        | String  | No       | `${{ github.repository }}` | DockerHub Repository to store docker images in. Default: github.repository                                                                                                                                                                                                               |
| `file`                        | String  | No       |                            | The dockerfile to use. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                                                                      |
| `gar-delete-credentials-file` | Boolean | No       | `true`                     | Delete the Google credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false.                                                                                                                                           |
| `gar-environment`             | String  | No       | `dev`                      | Environment for pushing artifacts (can be either dev or prod). This sets the GAR Project (gar-project) to either `grafanalabs-dev` or `grafanalabs-global`.                                                                                                                              |
| `gar-image`                   | String  | No       |                            | Name of the image to build. Default: `${GitHub Repo Name}`.                                                                                                                                                                                                                              |
| `gar-registry`                | String  | No       | `us-docker.pkg.dev`        | Google Artifact Registry to store docker images in.                                                                                                                                                                                                                                      |
| `gar-repository`              | String  | No       |                            | Override the 'repo_name' used to construct the GAR repository name. Only necessary when the GAR includes a repo name that doesn't match the GitHub repo name. Default: `docker-${GitHub Repo Name}-${gar-environment}`                                                                   |
| `include-tags-in-push`        | Boolean | No       | `true`                     | Disables the pushing of tags, and instead includes just a list of images as docker tags. Used when pushing docker digests instead of docker tags.                                                                                                                                        |
| `labels`                      | String  | No       |                            | List of custom labels to add to the image as metadata (Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs)). Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).        |
| `load`                        | Boolean | No       | `false`                    | Whether to load the built image into the local docker daemon (Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs)). Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs). |
| `outputs`                     | String  | No       |                            | List of docker output destinations. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                                                         |
| `platforms`                   | String  | No       |                            | List of platforms to build the image for. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                                                   |
| `push`                        | String  | No       |                            | Whether to push the image to the configured registries. Defaults to false (matching Docker's official action). Set to true to push images. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                  |
| `registries`                  | String  | No       |                            | CSV list of registries to build images for. Accepted registries are "gar" and "dockerhub".                                                                                                                                                                                               |
| `secrets`                     | String  | No       |                            | Secrets to expose to the build. Only needed when authenticating to private repositories outside the repository in which the image is being built. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                           |
| `ssh`                         | String  | No       |                            | List of SSH agent socket or keys to expose to the build Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                                     |
| `tags`                        | String  | Yes      |                            | List of Docker tags to be pushed. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                                                           |
| `target`                      | String  | No       |                            | Sets the target stage to build. Passed to [docker/build-push-action](https://github.com/docker/build-push-action?tab=readme-ov-file#inputs).                                                                                                                                             |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name           | Description                                                  |
| -------------- | ------------------------------------------------------------ |
| `annotations`  | Generated annotations (from docker/metadata-action)          |
| `digest`       | Image digest (from docker/build-push-action)                 |
| `imageid`      | Image ID (from docker/build-push-action)                     |
| `images`       | Comma separated list of the images that were built           |
| `json`         | JSON output of tags and labels (from docker/metadata-action) |
| `labels`       | Generated Docker labels (from docker/metadata-action)        |
| `metadata`     | Build result metadata (from docker/build-push-action)        |
| `metadatajson` | Metadata JSON (from docker/metadata)                         |
| `tags`         | Generated Docker tags (from docker/metadata-action)          |
| `version`      | Generated Docker image version (from docker/metadata-action) |

<!-- END_OUTPUTS -->

## How we construct Google Artifact Registry Images

The full GAR image is constructed as follows, where `gar-project` is determined by `inputs.gar-environment`.

"${{ inputs.gar-registry }}/${{ gar-project }}/${{ inputs.gar-repository }}/${{ inputs.gar-image }}"

## How we construct DockerHub Images

The full DockerHub image is constructed as follows:

"${{ inputs.dockerhub-registry }}/${{ inputs.dockerhub-repository }}"

## Adding New Registries

Each registry is setup as follows:

- All inputs for a registry share the same prefix (ex: `gar-image`, `gar-repository`).
- Inputs that are used for a specific registry are _not_ required by the workflow. Instead, validation is done in a step
  specific to that registry.
- To calculate which registries have been configured, we loop through `inputs.registries`, and for each registry
  configured we set the outputs `include-<registry>`. Those flags can be used to create steps that only execute when X
  registry is configured.
- Each registry has a Setup step. This step takes the inputs specific to that registry and generates an untagged,
  `image` name for that specific registry.
- The `setup-vars` step then loops through each configured image and creates a full list of images to push.
- That's it! That list of images to push is fed to `docker/build-push-action` along with the configured tags, and each
  tagged image is pushed to each registry.

So then the full checklist of work to do to implement a new registry is:

- [ ] Add (and document) any inputs that you need to capture. Use the same prefix for all inputs, and all inputs must
      _not_ be required.
- [ ] Add a step before `setup-vars` that takes those input values and constructs a valid untagged image name for the
      registry you'll be pushing to. Then set that as an output.
      Ex: `echo "image=${DOCKERHUB_REGISTRY}/${DOCKERHUB_IMAGE}" | tee -a "${GITHUB_OUTPUT}"`
- [ ] Add your image into the `setup-vars` step by passing the output image into an env variable, and adding it to the
      list of images to be parsed. Use the existing repos as examples.
- [ ] Add a login step that depends
      on `${{ inputs.push == 'true' && steps.registries.outputs.include-<yourRegistry> == 'true' }}`, where yourRegistry is
      the value that will be passed into the `registries` input. Again, use existing repos as examples.
- [ ] Celebrate

## Migrating

This action is intended to replace `build-push-to-dockerhub` and `build-push-to-dockerhub`.

### Migrating from `build-push-to-dockerhub`

> [!IMPORTANT]
> **Breaking Change:** The `push` input now defaults to `false` (matching Docker's official `build-push-action`).
> You must explicitly add `push: true` to maintain the previous behavior of pushing images to the registry.

1. Use the new action
2. Rename dockerhub specific settings
3. Add `registries: dockerhub`
4. **Add `push: true`** to maintain previous behavior

```bash
# old
  - id: push-to-dockerhub
    uses: grafana/shared-workflows/actions/build-push-to-dockerhub@build-push-to-dockerhub/v0.4.0
    with:
      repository: ${{ github.repository }} # or any other dockerhub repository
      context: .
      tags: |-
        "2024-04-01-abcd1234"
        "latest"

# new
  - id: push-to-dockerhub
    # CHANGE: use the new action
    uses: grafana/shared-workflows/actions/docker-build-push-image@docker-build-push-image/v0.1.0
    with:
      # CHANGE: dockerhub specific configs
      dockerhub-repository: ${{ github.repository }} # or any other dockerhub repository
      context: .
      tags: |-
        "2024-04-01-abcd1234"
        "latest"
      # ADD: registry
      registries: dockerhub
      # ADD: push (was implicit before, now explicit)
      push: true
```

### Migrating from `push-to-gar-docker`

> [!IMPORTANT]
> **Breaking Change:** The `push` input now defaults to `false` (matching Docker's official `build-push-action`).
> The old `push-to-gar-docker` action defaulted to `push: true`. You must explicitly add `push: true` to maintain the previous behavior.

1. Use the new action
2. Rename gar specific settings
3. Add `registries: gar`
4. **Add `push: true`** to maintain previous behavior

```bash
# old
  - id: push-to-gar
    uses: grafana/shared-workflows/actions/push-to-gar-docker@push-to-gar-docker/v0.6.1
    with:
      registry: "<YOUR-GAR>" # e.g. us-docker.pkg.dev, optional
      image_name: "backstage" # name of the image to be published, required
      environment: "dev" # can be either dev/prod
      tags: |-
        "<IMAGE_TAG>"
        "latest"
      context: "<YOUR_CONTEXT>" # e.g. "." - where the Dockerfile is

# new
  - id: push-to-gar
    # CHANGE: use the new action
    uses: grafana/shared-workflows/actions/docker-build-push-image@docker-build-push-image/v0.1.0
    with:
      # CHANGE: gar specific configs
      gar-registry: "<YOUR-GAR>" # e.g. us-docker.pkg.dev, optional
      gar-image: "backstage" # name of the image to be published, required
      gar-environment: "dev" # can be either dev/prod
      tags: |-
        "<IMAGE_TAG>"
        "latest"
      context: "<YOUR_CONTEXT>" # e.g. "." - where the Dockerfile is
      # ADD: registry
      registries: gar
      # ADD: push (old action defaulted to true)
      push: true
```
