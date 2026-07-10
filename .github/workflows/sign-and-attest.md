# sign-and-attest

This is a reusable workflow that signs and attests a pushed, digest-pinned
container image:

1. **Keyless cosign signature**, signed via GitHub Actions OIDC through
   Sigstore.
2. **SLSA build provenance attestation**, generated with
   [`actions/attest-build-provenance`], pushed both to the registry as an OCI
   referrer and to the calling repository's GitHub attestation store. It ties
   the image digest to the source commit, workflow, and run that built it.

Both are verified in CI immediately after being produced, using the same
claims consumers are documented to check below, so a broken signature or
attestation fails the run before anyone relies on it.

Registry authentication uses the calling repository's Workload Identity
Federation identity (via [`login-to-gar`]), so the workflow works unchanged
for any `grafana` repository that can push to GAR. There are no secrets to
pass; the job in the `grafana/shared-workflows` repository is guarded with
`if: github.repository_owner == 'grafana'` and skips cleanly elsewhere.

[`actions/attest-build-provenance`]: https://github.com/actions/attest-build-provenance
[`login-to-gar`]: ../../actions/login-to-gar/README.md

## Usage

Call it after the job that pushes the image, passing the digest-pinned
reference of the (multi-arch index) manifest:

```yaml
jobs:
  build:
    # ... builds and pushes the image, exposes the digest-pinned ref as an output ...

  sign:
    needs: build
    permissions:
      contents: read
      id-token: write
      attestations: write
    uses: grafana/shared-workflows/.github/workflows/sign-and-attest.yml@<sha> # main
    with:
      image: ${{ needs.build.outputs.image }} # e.g. us-docker.pkg.dev/…/my-image@sha256:…
```

Pass a digest (`@sha256:…`), not a tag. Signatures attach to the manifest
digest; the tag is only resolved to a digest at verify time. This has two
consequences: one signing call covers every tag pointing at that manifest
(`latest`, `vX.Y.Z`, …), and passing a tag here would be ambiguous, since the
tag could have been repointed since your build pushed it. The workflow
therefore rejects refs without `@sha256:`.

## Inputs

| Name             | Type   | Description                                                                                       |
| ---------------- | ------ | ------------------------------------------------------------------------------------------------- |
| `image`          | string | **Required.** Digest-pinned image reference under `registry`, e.g. `us-docker.pkg.dev/…@sha256:…` |
| `registry`       | string | GAR hostname (`*.pkg.dev`) to authenticate against and sign in. Default: `us-docker.pkg.dev`      |
| `cosign-version` | string | Version of cosign to use for signing and attesting.                                               |

## Required caller permissions

The calling job's permissions must include (called-workflow permissions can
never exceed the caller's; check the whole `workflow_call` chain if the
calling workflow is itself reusable):

```yaml
permissions:
  contents: read
  id-token: write # keyless cosign OIDC, provenance signing, WIF auth to GAR
  attestations: write # write to the GitHub attestations store
```

## Verification identity contract

The Fulcio certificate identity for every image signed by this workflow is:

```text
https://github.com/grafana/shared-workflows/.github/workflows/sign-and-attest.yml@<ref>
issuer: https://token.actions.githubusercontent.com
```

> [!WARNING]
> This path/name is a public verification contract shared by **every**
> consuming repository. Renaming or moving this workflow breaks verification
> for all of them and for all previously published verification docs. Don't.

Because the identity is shared org-wide, it only proves "signed by Grafana's
shared signer workflow", not which repo the image came from.
Verifiers bind the signature to a specific source repository with the
certificate's repository claim:

```bash
# Requires cosign >= v3.0
cosign verify \
  --certificate-identity-regexp '^https://github\.com/grafana/shared-workflows/\.github/workflows/sign-and-attest\.yml@' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-github-workflow-repository grafana/<repo> \
  <image>@sha256:...
```

And for the provenance attestation (`--repo` binds to the repository;
`--signer-workflow` is the valid companion flag; it is mutually exclusive
with `--cert-identity`/`--signer-repo`):

```bash
gh attestation verify oci://<image>@sha256:... \
  --repo grafana/<repo> \
  --signer-workflow grafana/shared-workflows/.github/workflows/sign-and-attest.yml
```

This cross-repository signer pattern is [GitHub's documented architecture for
artifact attestations with reusable workflows](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-and-reusable-workflows-to-achieve-slsa-v1-build-level-3).

## SLSA level

This is **SLSA Build L2**: the provenance is authentic and signed, but the
image is built in the _calling_ repository's job, not inside this trusted
workflow, so the build itself is not isolated from the caller. Reaching L3
would require moving the build into the reusable workflow. Don't claim L3
when documenting images signed this way.

## Why a reusable workflow and not a composite action?

A composite action runs inside the caller's job, so the Fulcio certificate
identity (`job_workflow_ref`) stays the _caller's_ workflow: every repo gets
a different verification identity, there is no shared contract to document,
and no centrally patchable signer. Only a reusable workflow puts
`grafana/shared-workflows` in the certificate, which is exactly the
architecture GitHub documents for cross-repository signing (link above).
