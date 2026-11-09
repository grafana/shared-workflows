# trivy-image

Scans a container image built earlier in the same job with
[Trivy](https://github.com/aquasecurity/trivy) and posts an informational
comment on the pull request. It never fails the job because of findings.

What the comment shows:

- Counts by severity.
- A red/green/blue callout with the vulnerabilities **introduced** and **resolved**
  compared to a baseline, when one is available (see [Baseline](#baseline)).
- The full list of findings at the requested severities, collapsed, with a link to the
  workflow run.

The same Markdown is written to the job summary, so fork PRs (which cannot post
comments) still get the results. Scan metadata that is not in the comment (detected OS,
whether the SBOM came from cache, which baseline source was used) is printed in the
`Build report` step log.

For why Trivy rather than Grype or OSV-Scanner, see [scanner-evaluation.md](scanner-evaluation.md).

## Requirements

Linux runner (x86_64 or arm64) with `curl`, `jq`, `sha256sum` and `tar`, plus `docker`
when the image or the baseline image comes from the daemon. GitHub-hosted Ubuntu
runners provide all of these; the action checks for them up front.

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
      actions: read # baseline artifact lookup (optional, see Baseline)
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

`image-name` is the stable identity used for cache keys, the baseline, the artifact
name and the comment id. Keep it identical between the PR and push workflows for the
same image. It is derived from `image-ref` when not set and required with `image-tar`.

## Baseline

The introduced/resolved section needs an SBOM of the image as it was before the
change. The action looks for one in this order and uses the first that works:

1. **`baseline-image-ref`**: an image you name explicitly, e.g. the tag your default
   branch last pushed to a registry. It is pulled if not present locally (log in to the
   registry first) and its SBOM is cached by layer digest like the main image.
2. **Cached baseline**: the SBOM saved to the Actions cache by a previous run on the
   base branch (`baseline-ref`, default: the PR base branch).
3. **Baseline artifact**: the `trivy-image-baseline-<image-name>` workflow artifact
   from the most recent run on the base branch. Survives cache eviction (90-day
   retention) but needs `actions: read` on the token; without it this step is skipped
   with a notice.

Options 2 and 3 are populated by running the action on pushes to the base branch. On
`push` events (`save-baseline: auto`) the image's SBOM is saved both to the cache (under
the branch name) and as an artifact. If the merge workflow pushes to a registry instead
of loading locally, pull or load the image first, then point `image-ref` at it:

```yaml
# push-to-main workflow
- uses: grafana/shared-workflows/actions/trivy-image@<sha>
  with:
    image-ref: us-docker.pkg.dev/.../my-app:${{ steps.meta.outputs.tag }}
    image-name: my-app
    comment: "false"
```

Alternatively, if the merge workflow always pushes a `latest` (or `main`) tag, skip the
push-side run entirely and diff against the registry image directly:

```yaml
- uses: grafana/shared-workflows/actions/trivy-image@<sha>
  with:
    image-ref: my-app:pr-${{ github.sha }}
    image-name: my-app
    baseline-image-ref: us-docker.pkg.dev/<project>/<repo>/my-app:latest
```

Until a baseline exists, PR comments say so and show the full list only. The callout
names the baseline it compared against, and the `baseline-source` output exposes the
same.

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
  six-hour buckets to match its publishing cadence, with a prefix fallback so a stale
  cache is updated incrementally rather than downloaded from scratch.
- The baseline SBOM is re-matched at the same moment with the same DB, so the
  introduced/resolved diff reflects only what the change did to the image, not DB drift
  between two scan times.

## Failure behaviour

Findings never fail the job. Nor does a missing or broken baseline, an unreadable
cached SBOM, or a failed PR comment (a warning is emitted and the report is still in
the job summary). The action does fail when it cannot do its core job: the image is not
in the local daemon or the archive is not an image, Trivy cannot be installed or its
checksum does not match, the vulnerability DB cannot be downloaded after three
attempts, or the scan itself errors.

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

| Name                 | Type    | Required | Default               | Description                                                                                                                                                                                                                                                  |
| -------------------- | ------- | -------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `baseline-image-ref` | String  | No       |                       | Image to diff against instead of the cached/artifact baseline, e.g. the image your default branch last pushed to a registry. Pulled if not present locally; log in to the registry before this step. Its SBOM is cached by layer digest like the main image. |
| `baseline-ref`       | String  | No       |                       | Branch whose baseline to diff against. Defaults to the PR base branch.                                                                                                                                                                                       |
| `cache-sbom`         | Boolean | No       | `true`                | Cache the image SBOM keyed on layer digests. Set to false for workflows that scan a stream of distinct images where the cache would never hit.                                                                                                               |
| `comment`            | Boolean | No       | `true`                | Post or update an informational comment on the pull request.                                                                                                                                                                                                 |
| `github-token`       | String  | No       | `${{ github.token }}` | Token used to post the PR comment and to look up baseline artifacts. Needs `pull-requests: write` for comments and `actions: read` for the artifact fallback.                                                                                                |
| `ignore-unfixed`     | Boolean | No       | `false`               | Only report vulnerabilities that have a fixed version.                                                                                                                                                                                                       |
| `image-name`         | String  | No       |                       | Stable name for this image used in cache keys, baseline keys, artifact names and the comment id. Must be the same on pull_request and push runs. Defaults to `image-ref` without tag/digest; required with `image-tar`.                                      |
| `image-ref`          | String  | No       |                       | Image reference present in the local Docker daemon, e.g. `my-app:pr-<sha>`. With docker/build-push-action set `load: true` so the image is loaded. Exactly one of `image-ref` and `image-tar` must be set.                                                   |
| `image-tar`          | String  | No       |                       | Path to the image as an OCI layout (tar or directory, e.g. the output of `docker buildx build --output type=oci,dest=image.tar`) or a `docker save` archive. Use this when the image is not loaded into the daemon. Requires `image-name`.                   |
| `max-rows`           | String  | No       | `100`                 | Maximum number of rows in each findings table.                                                                                                                                                                                                               |
| `save-baseline`      | String  | No       | `auto`                | Save this image's SBOM as the baseline for the current branch (Actions cache and a 90-day workflow artifact) so pull requests targeting it can report introduced/resolved vulnerabilities. `auto` saves on push events only.                                 |
| `severity`           | String  | No       | `CRITICAL,HIGH`       | Comma-separated severities to report.                                                                                                                                                                                                                        |
| `timeout`            | String  | No       | `10m`                 | Trivy timeout applied to each Trivy invocation.                                                                                                                                                                                                              |
| `trivyignores`       | String  | No       |                       | Comma-separated workspace-relative paths to .trivyignore files (concatenated).                                                                                                                                                                               |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name              | Description                                                                      |
| ----------------- | -------------------------------------------------------------------------------- |
| `baseline-found`  | true when a baseline SBOM was available.                                         |
| `baseline-source` | Where the baseline came from: image, cache, artifact or none.                    |
| `critical`        | Number of CRITICAL findings.                                                     |
| `fixable`         | Findings that have a fixed version.                                              |
| `high`            | Number of HIGH findings.                                                         |
| `introduced`      | Findings present in this image but not in the baseline (empty without baseline). |
| `low`             | Number of LOW findings.                                                          |
| `medium`          | Number of MEDIUM findings.                                                       |
| `report-path`     | Path to the Trivy JSON vulnerability report.                                     |
| `resolved`        | Findings present in the baseline but not in this image (empty without baseline). |
| `sbom-cache-hit`  | true when a valid SBOM was restored from cache instead of generated.             |
| `sbom-path`       | Path to the CycloneDX SBOM of the image.                                         |
| `total`           | Total findings at the requested severities.                                      |

<!-- END_OUTPUTS -->

Use the outputs if a consumer wants its own gate, for example:

```yaml
- if: steps.scan.outputs.introduced != '' && steps.scan.outputs.introduced != '0'
  run: echo "::warning::This change introduces ${{ steps.scan.outputs.introduced }} new vulnerabilities"
```
