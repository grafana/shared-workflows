# Scanner evaluation: why `trivy-image` uses Trivy

Evaluated 2026-11-09 with Trivy 0.76.1, Grype 0.116.0 (Syft 1.43.0) and OSV-Scanner
2.4.0, each with its database as of that day. Numbers below are a one-day snapshot;
the shape of the differences is what matters, not the exact counts.

## Requirements

The action scans an image built in a PR job and posts an informational comment. It needs:

1. OS package and language package coverage for the image shapes we build.
2. Stable, machine-readable output with vendor severity and fixed-version data.
3. An SBOM round-trip (`image → SBOM → vulnerabilities`) that loses nothing, so the
   SBOM can be cached by image layer digest and re-matched against the current DB.
4. Results that line up with the scanning that runs on the same image after it is pushed,
   so CI and the registry-side report do not disagree.

## Image shapes

The images that will adopt this action first are, by their `FROM` lines, mostly Go binaries
on `scratch` or a distroless static base, with a smaller number of Go binaries plus a vendored
CLI tool on Alpine and the occasional Node app on a distroless Node base. Go-on-minimal-base
is the large majority, so Go binary analysis is the coverage that matters most.

## Method

Three public images were used as stand-ins for the dominant shapes, scanned by all three tools:

| Stand-in        | Image                                                           | Shape                                     |
| --------------- | --------------------------------------------------------------- | ----------------------------------------- |
| `alpine-go`     | `ghcr.io/google/osv-scanner:v2.0.0`                             | Alpine 3.21 + Go binary                   |
| `distroless-go` | `registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.10.0` | distroless/static (Debian 11) + Go binary |
| `debian-node`   | `node:20.11-bookworm-slim`                                      | Debian 12 + Node + npm's `node_modules`   |

```bash
trivy image --format json --list-all-pkgs --output trivy.json "$IMG"
grype -o json --file grype.json "$IMG"
osv-scanner scan image --format json --output osv.json "$IMG"
```

Findings were reduced to unique vulnerability IDs per tool: GHSA/GO IDs were replaced by
their CVE alias when one existed, and OSV's `DEBIAN-CVE-*` / `ALPINE-CVE-*` IDs had the
distro prefix stripped. OSV-Scanner reports a CVSS score rather than a vendor severity, so
its CRITICAL/HIGH counts use the usual 9.0 / 7.0 thresholds.

## Results

Unique IDs, CRITICAL/HIGH by each tool's own severity, and IDs reported by only that tool.

**alpine-go**

| Tool        | Unique IDs | Critical | High | Only this tool | Notes                                                               |
| ----------- | ---------- | -------- | ---- | -------------- | ------------------------------------------------------------------- |
| Trivy       | 263        | 4        | 95   | 0              |                                                                     |
| Grype       | 269        | 22       | 122  | 30             | all 30 extras are `nvd:cpe` matches (busybox, c-ares, go, nghttp2…) |
| OSV-Scanner | 261        | 17       | 62   | 2              | 253 of 261 shared with Trivy                                        |

Trivy and Grype agreed on 233 IDs but disagreed on severity for 69 of them.

**distroless-go**

| Tool        | Unique IDs | Critical | High | Only this tool | Notes                                       |
| ----------- | ---------- | -------- | ---- | -------------- | ------------------------------------------- |
| Trivy       | 133        | 2        | 50   | 1              | 128 Go + 5 `tzdata` DLAs (severity UNKNOWN) |
| Grype       | 68         | 4        | 30   | 0              | matched only `stdlib` (67) and `x/net` (9)  |
| OSV-Scanner | 133        | 8        | 5    | 0              | 132 of 133 shared with Trivy                |

Syft catalogued 75 Go modules and 3 deb packages here, so Grype's gap is in matching, not
inventory: it reported nothing for `golang.org/x/crypto` where Trivy found 23 IDs and OSV
41, and 9 for `golang.org/x/net` against 23 / 30. Root cause not investigated.

**debian-node**

| Tool        | Unique IDs | Critical | High | Only this tool | Notes                                       |
| ----------- | ---------- | -------- | ---- | -------------- | ------------------------------------------- |
| Trivy       | 221        | 7        | 55   | 9              | `debian` + `node-pkg`                       |
| Grype       | 254        | 11       | 93   | 44             | all 44 extras are `nvd:cpe` matches         |
| OSV-Scanner | 190        | 10       | 72   | 12             | Debian packages only; found no npm packages |

**Timing** (local, warm DB): 1–4 s for every tool on every image. Cold start is dominated
by the DB download: Trivy ≈10 s, Grype ≈54 s in this run. In CI the action caches the
Trivy DB in six-hour buckets.

## Observations

- **Trivy and OSV-Scanner agree almost completely** on the vulnerability set for Alpine,
  Debian and Go. The two draw from the same distro and Go vulnerability feeds.
- **Grype's extras are CPE matches.** Every ID Grype reported that neither other tool did
  came from the `nvd:cpe` namespace. These are the classic false-positive source (a package
  name matching an NVD CPE regardless of ecosystem). `--only-fixed` or disabling the CPE
  matcher removes them, at which point Grype is close to Trivy.
- **Grype under-reported Go module vulnerabilities on the distroless image.** With most of
  the target images being Go on scratch/distroless, this is the most consequential finding.
- **OSV-Scanner missed npm packages** inside the Node image and only has distro support for
  Alpine, Debian and Ubuntu (it handled distroless' Debian `status.d` correctly). It also
  has no secret or misconfiguration scanning, and reports CVSS scores rather than vendor
  severities, so its critical/high counts differ structurally from the other two.
- **Severity labels are not comparable across tools.** Trivy and Grype disagreed on ~30% of
  the IDs they both reported (69/233, 17/67, 71/210). Whatever CI uses should match what
  scans the pushed image, or every PR comment becomes an argument about which number is right.
- **Trivy quirk:** on Debian LTS it reports DLA advisory IDs with severity UNKNOWN instead
  of CVEs. These are excluded by the default `CRITICAL,HIGH` filter.

## Decision

Trivy, for these reasons in order:

1. **Coverage on our dominant shape.** Trivy and OSV-Scanner found the full Go module
   vulnerability set on the distroless image; Grype found roughly half.
2. **Consistency with post-push scanning.** Pushed images are scanned with Trivy
   downstream, so a PR comment and the registry-side report for the same image show
   the same IDs and severities, and `.trivyignore` conventions carry over.
3. **SBOM round-trip verified.** `trivy image --format cyclonedx` followed by `trivy sbom`
   produced exactly the same findings as a direct image scan (146 OS + 133 Go on a test
   image), which is what makes caching the SBOM by layer digest safe.
4. **Report-only mode favours recall.** Since the action never fails the build, Trivy's
   slightly higher recall costs nothing; Grype's quieter defaults would be more attractive
   for a blocking gate.

Grype remains a reasonable second opinion (it adds EPSS and CISA KEV annotations Trivy
lacks) and OSV-Scanner remains the right tool for source and lockfile scanning.

## Supply chain

On 2026-03-19 attackers force-pushed 76 of 77 `aquasecurity/trivy-action` tags and all
`setup-trivy` tags to a credential stealer and published a malicious Trivy `v0.69.4`
([GHSA-69fq-xp46-6x23](https://github.com/aquasecurity/trivy/security/advisories/GHSA-69fq-xp46-6x23)).
This action does not use either GitHub Action. It downloads a pinned Trivy release binary
(`TRIVY_BIN_VERSION` in `action.yaml`, tracked by Renovate) and verifies its SHA-256 against the
release checksums file. The incident is a reason to pin, not a reason to prefer a different
scanner; none of the alternatives is structurally immune to the same attack.

## Revisit when

- Grype's Go module matching on distroless images is confirmed fixed (re-run the
  `distroless-go` comparison).
- The action moves from report-only to a blocking gate, where Grype's lower noise with
  CPE matching disabled may be worth the severity drift against post-push scanning.
- KEV/EPSS prioritisation in the PR comment becomes a requirement.
