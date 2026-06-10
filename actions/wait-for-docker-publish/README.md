# wait-for-docker-publish

Polls an OCI registry until the given image reference is published, or fails
after a timeout. The load-bearing use case is Grafana's async GAR→DockerHub
mirror: after pushing to GAR with a short-lived OIDC token, downstream
consumers use this action to wait for the mirror to replicate the image to
DockerHub before pulling. Mechanism is `docker manifest inspect` in a retry
loop, so the action works against any OCI-conformant registry.

## Inputs

| Name               | Required | Default | Description                                                                        |
| ------------------ | -------- | ------- | ---------------------------------------------------------------------------------- |
| `image`            | yes      | —       | OCI image reference. **Must include a `:tag` or an `@sha256:…` digest** (or both). |
| `timeout`          | no       | `10m`   | Total wall-clock budget. Accepts `s`/`m`/`h` suffixes.                             |
| `initial-interval` | no       | `5s`    | First sleep after a miss.                                                          |
| `max-interval`     | no       | `60s`   | Upper bound on the exponential backoff.                                            |

## Outputs

None. The step exits 0 on success and 1 on timeout.

## Behaviour

- Polling cadence: `initial-interval`, then doubling each miss up to `max-interval` (default `5s, 10s, 20s, 40s, 60s, 60s, …`).
- Auth is delegated to the Docker CLI's normal credential lookup (`~/.docker/config.json` + credential helpers). Run `grafana/shared-workflows/actions/dockerhub-login` or `docker/login-action` before this step for private images.

## Usage

Primary use case — waiting for a GAR-pushed image to be mirrored to DockerHub:

<!-- x-release-please-start-version -->

```yaml
jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
    outputs:
      digest: ${{ steps.push.outputs.digest }}
    steps:
      - id: push
        uses: grafana/shared-workflows/actions/docker-build-push-image@docker-build-push-image/v0.2.0
        with:
          registries: "gar"
          push: true
          tags: |
            ${{ github.sha }}

  wait-for-mirror:
    needs: publish
    runs-on: ubuntu-latest
    steps:
      - name: Wait for DockerHub mirror
        uses: grafana/shared-workflows/actions/wait-for-docker-publish@wait-for-docker-publish/v0.2.0
        with:
          image: grafana/myrepo@${{ needs.publish.outputs.digest }}
```

Tag-only example (sufficient when tags are not reused):

```yaml
- name: Wait for DockerHub mirror
  uses: grafana/shared-workflows/actions/wait-for-docker-publish@wait-for-docker-publish/v0.2.0
  with:
    image: grafana/myrepo:${{ github.sha }}
    timeout: 15m
```

<!-- x-release-please-end-version -->

## Non-goals

- Verifying that a moving tag now points to a specific digest. Pinning by digest in the `image` input checks digest _reachability_, but not that tag now points at a specific digest.
- Verifying that a manifest list contains all expected platforms.
