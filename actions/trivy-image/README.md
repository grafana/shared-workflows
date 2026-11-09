# trivy-image

Scans a container image built earlier in the same job with
[Trivy](https://github.com/aquasecurity/trivy) and posts an informational
comment on the pull request. It never fails the job because of findings.

The comment shows counts by severity and a table of the findings at the requested
severities (expanded when short, collapsed when long), with a link to the workflow run.
The same Markdown is written to the job summary, so fork PRs (which cannot post comments)
still get the results. Scan metadata that is not in the comment (detected OS, whether the
SBOM came from cache) is printed in the `Build report` step log.

For why Trivy rather than Grype or OSV-Scanner, see [scanner-evaluation.md](scanner-evaluation.md).

## Requirements

Linux runner (x86_64 or arm64) with `curl`, `jq`, `sha256sum` and `tar`, plus `docker`
when the image comes from the daemon. GitHub-hosted Ubuntu runners provide all of these;
the action checks for them up front.

## Example workflow

The image can come from the local Docker daemon (`image-ref`) or from an archive
(`image-tar`). Exactly one must be set.

<!-- x-release-please-start-version -->

```yaml
name: Build

on:
  pull_request:

permissions: {}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write # PR comment
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          persist-credentials: false

      - uses: docker/build-push-action@53b7df96c91f9c12dcc8a07bcb9ccacbed38856a # v7.3.0
        with:
          context: .
          load: true # puts the image in the daemon (single platform only)
          push: false
          tags: my-app:pr-${{ github.sha }}

      - uses: grafana/shared-workflows/actions/trivy-image@trivy-image/v0.1.0
        with:
          image-ref: my-app:pr-${{ github.sha }}
          image-name: my-app
```

<!-- x-release-please-end-version -->

When the build uses the `docker-container` buildx driver without `load: true` (the
default for multi-platform builds, and what `docker-build-push-image` does), the image
is not in the daemon. Export it as an OCI archive and pass that instead:

```yaml
- run: docker buildx build --output "type=oci,dest=${RUNNER_TEMP}/image.tar" .

- uses: grafana/shared-workflows/actions/trivy-image@<sha>
  with:
    image-tar: ${{ runner.temp }}/image.tar
    image-name: my-app
```

`image-tar` accepts an OCI layout (tar or directory) or a `docker save` archive.

`image-name` is the stable identity used for cache keys and the comment id, so a
workflow that builds several images gets one comment per image. It is derived from
`image-ref` when not set and required with `image-tar`.

## How caching works

- The image's **SBOM** (CycloneDX) is cached, keyed on the image's layer digests and
  the Trivy version. Layer digests are content-addressed, so an unchanged image hits the
  cache even though its image ID changes with every build timestamp. On a hit, the
  expensive image analysis is skipped. A cached SBOM that fails validation is discarded
  and regenerated. Set `cache-sbom: "false"` for workflows that scan a stream of
  distinct images where the cache would never hit. Daemon images and archives of the
  same image produce different keys (uncompressed vs compressed layer digests), so
  switching input modes for an image costs one cache miss.
- Vulnerability matching always runs against the **current** Trivy DB, so a PR that
  has been open for a week still shows newly published CVEs. The DB is cached in
  six-hour buckets to match its publishing cadence; on a miss the newest previous
  bucket is restored and Trivy only downloads a fresh copy if that one has expired.
  Caches are per repository, so each consumer pays roughly one download per six hours.

## Failure behaviour

Findings never fail the job. Nor does an unreadable cached SBOM or a failed PR comment
(a warning is emitted and the report is still in the job summary). The action does fail
when it cannot do its core job: the image is not in the local daemon or the archive is
not an image, Trivy cannot be installed or its checksum does not match, the
vulnerability DB cannot be downloaded after three attempts, or the scan itself errors.

## Supply chain

Trivy is installed as a pinned release binary (`TRIVY_BIN_VERSION` in `action.yaml`,
updated by Renovate) with SHA-256 verification against the release checksums file. The
`aquasecurity/trivy-action` and `setup-trivy` actions are not used; they were the vector
of the March 2026 tag-hijacking incident
([GHSA-69fq-xp46-6x23](https://github.com/aquasecurity/trivy/security/advisories/GHSA-69fq-xp46-6x23)).
Third-party actions used here are pinned to commit SHAs. Package names, versions and
URLs in the comment come from the image and the vulnerability DB and are escaped before
rendering; only plain `https://` advisory URLs are linked.

## Suppressing findings

A `.trivyignore` at the consumer repo root is picked up automatically; use
`trivyignores` to point at other files instead. Entries are CVE IDs, one per line,
optionally with an `exp:YYYY-MM-DD` expiry; see the
[Trivy docs](https://trivy.dev/latest/docs/configuration/filtering/#trivyignore).

## Inputs

<!-- BEGIN_INPUTS -->

| Name             | Type    | Required | Default               | Description                                                                                                                                                                                                                                |
| ---------------- | ------- | -------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `cache-sbom`     | Boolean | No       | `true`                | Cache the image SBOM keyed on layer digests. Set to false for workflows that scan a stream of distinct images where the cache would never hit.                                                                                             |
| `comment`        | Boolean | No       | `true`                | Post or update an informational comment on the pull request.                                                                                                                                                                               |
| `github-token`   | String  | No       | `${{ github.token }}` | Token used to post the PR comment. Needs `pull-requests: write`.                                                                                                                                                                           |
| `ignore-unfixed` | Boolean | No       | `false`               | Only report vulnerabilities that have a fixed version.                                                                                                                                                                                     |
| `image-name`     | String  | No       |                       | Stable name for this image used in cache keys and the comment id. Defaults to `image-ref` without tag/digest; required with `image-tar`.                                                                                                   |
| `image-ref`      | String  | No       |                       | Image reference present in the local Docker daemon, e.g. `my-app:pr-<sha>`. With docker/build-push-action set `load: true` so the image is loaded. Exactly one of `image-ref` and `image-tar` must be set.                                 |
| `image-tar`      | String  | No       |                       | Path to the image as an OCI layout (tar or directory, e.g. the output of `docker buildx build --output type=oci,dest=image.tar`) or a `docker save` archive. Use this when the image is not loaded into the daemon. Requires `image-name`. |
| `max-rows`       | String  | No       | `100`                 | Maximum number of rows in the findings table.                                                                                                                                                                                              |
| `severity`       | String  | No       | `CRITICAL,HIGH`       | Comma-separated severities to report.                                                                                                                                                                                                      |
| `timeout`        | String  | No       | `10m`                 | Trivy timeout applied to each Trivy invocation.                                                                                                                                                                                            |
| `trivyignores`   | String  | No       |                       | Comma-separated workspace-relative paths to .trivyignore files (concatenated).                                                                                                                                                             |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name             | Description                                                          |
| ---------------- | -------------------------------------------------------------------- |
| `critical`       | Number of CRITICAL findings.                                         |
| `fixable`        | Findings that have a fixed version.                                  |
| `high`           | Number of HIGH findings.                                             |
| `low`            | Number of LOW findings.                                              |
| `medium`         | Number of MEDIUM findings.                                           |
| `report-path`    | Path to the Trivy JSON vulnerability report.                         |
| `sbom-cache-hit` | true when a valid SBOM was restored from cache instead of generated. |
| `sbom-path`      | Path to the CycloneDX SBOM of the image.                             |
| `total`          | Total findings at the requested severities.                          |

<!-- END_OUTPUTS -->

Use the outputs if a consumer wants its own gate, for example:

```yaml
- if: steps.scan.outputs.critical != '0'
  run: echo "::warning::This image has ${{ steps.scan.outputs.critical }} critical vulnerabilities"
```
