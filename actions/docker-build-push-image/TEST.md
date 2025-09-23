Design decisions

## Whole workflow

1. The guide I used to build multi-arch images is
   here: https://docs.docker.com/build/ci/github-actions/multi-platform/#distribute-build-across-multiple-runners
1. The names of all the new actions are prefixed with `docker-`, to improve discoverability.
1. I've tried to use docker's official actions ecosystem as much as possible. Most inputs are directly passed
   through to the underlying docker actions, unless some sort of custom logic is necessary.
1. I've avoided repeating actions in places where I can. An example of this is one area where I differentiated from the
   official guide (linked in item #1). Docker shows the metadata action being used twice (once when building the
   image and once when pushing the manifest). Instead of setting up the metadata action a second time, I capture the
   output from the original build, and use it as an input when pushing the manifests.
1. I tried to make this as reusable as possible. An example of this is GAR images: previously we made some assumptions
   about the naming convention for GAR repos. Now you can now configure the entire image URL/path, so if your GAR bucket
   doesn't match our naming convention this will (probably) still work.

# docker-build-push-image

1. This handles both gar and dockerhub logins, with a framework in place that makes it easy to add other
   registries in the future. How to add new registries is documented in the README.
2. This handles both pushing tags and digests (for multiarch builds). The idea is that we can use this as a direct
   replacement for both `push-to-gar-docker` and `build-push-to-dockerhub`, while also making it flexible enough that we
   can use it for true multiarch builds, and good ole-fashioned emulation builds.
3. To configure which registries we build and push to, an input is provided that takes a CSV of
   registries (`gar,dockerhub`).
4. For each registry that we configure, a script is included that takes the inputs for that registry and outputs an
   image to the GITHUB_OUTPUT. Each image (without tags) is then appended to the `images` output. Inputs for each
   registry begin with the same prefix (ex: `gar-config-value`)
5. For each registry that we configure, we only login to that registry if it's in the `registries` input list.
6. This action outputs both a tagged image list and an image list without tags. This is because when you're pushing a
   manifest, tags cannot be included; however when you're pushing tags, you have to include them. So we offer the
   ability to do either.
7. If `include-tags-in-push=true`, then we push the tags list, if `include-tags-in-push=false` then we push the untagged
   images. When untagged images are pushed you must collect the docker digests and merge them into a manifest in a
   following step/job.
8. To provide access to Grafana's Docker mirror, if a workflow is running on a self-hosted runner, and no
   buildkitd-config is provided, then we set the default to `/etc/buildkitd.toml` if it exists.
9. In the event `/etc/buildkitd.toml` does not exist then we will not set a default buildkitd-config. This allows the
   action to work by default when run inside a container.
10. Most inputs are passed directly to the underlying `docker/build-push-action`, with the exception of tags. The tags
    are built depending on the input values of `include-tags-in-push` and `registries`.
11. If this action is run with `push=false` then we still successfully build images. However a warning is logged saying
    that we're not pushing anything.
12. If this action is run without any registries configured then we still successfully build an image. We construct a
    fake image name and log a warning that we're not pushing anything.

## docker-export-digest

1. The `digest` input can be collected from `docker-build-push-image` or `docker/build-push-image`; both of which
   output a docker digest.

## docker-import-digests-push-manifest

1. The `docker-metadata-json` and `images` inputs can be captured from `docker-build-push-image`.
2. The `gar-environment` and `push` match inputs that are also used by `docker-build-push-image`.
3. For simplicity, I did not replicate the `setup-vars` logic from `docker-build-push-image`, and instead we log in to
   both GAR and DockerHub when `push=true`.

## Potential Follow-up Items

1. Should we convert naming to match new standards?
   - `actions/dockerhub-login` -> `actions/login-to-dockerhub`
2. Deprecate `actions/push-to-gar-docker` and `actions/build-push-to-dockerhub` in favor
   of `actions/docker-build-push-image`
3. Integrated GHCR registry into `docker-build-push-image`.
