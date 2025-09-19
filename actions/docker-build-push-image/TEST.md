Design decisions

docker-build-push-image

1. This should handle both gar and dockerhub logins, with a framework in place that makes it easy to add other
   registries in the future. Which registries we build images for and login to should be configurable.
2. This should handle both pushing tags and digests (for multiarch builds). The idea is that we can use this as a direct
   replacement for both `push-to-gar-docker` and `build-push-to-dockerhub`, while also making it flexible enough that we
   can use it for true multiarch builds.
2. To configure which registries we build and push to, an input is provided that takes a CSV of
   registries (`gar,dockerhub`).
3. For each registry that we configure, a script is included that takes the inputs for that registry and outputs an
   image to the
   GITHUB_OUTPUT. Each image (without tags) is then appended to the `images` output. Inputs for each registry begin with
   the same prefix (ex: `gar-config-value`)
4. For each registry that we configure, we only login to that registry if it's in the `registries` input list.
5. This action outputs both a tagged image list and an image list without tags. This is because when you're pushing a
   manifest, tags cannot be
   included; however when you're pushing tags, you have to include them. So we include both.
6. If `include-tags-in-push=true`, then we push the tag list, if `include-tags-in-push=false` then we push the image
   list.
7. To provide access to Grafana's Docker mirror, if a workflow is running on a self-hosted runner, and no
   buildkitd-config is provided, then we default to `/etc/buildkitd.toml`. This means if you're running this action from
   inside a container you need to either mount that directory, or explicitly override it by setting `buidkitd-config`
   or `buidkitd-config-inline`.
5.
6.
7. (Ex: images=us-docker.pkg.dev/grafanalabs-dev/gar-registry/image-name,docker.io/grafana/dockerhub-image)
5.
3. We only authenticate to registries that are included in the ``
