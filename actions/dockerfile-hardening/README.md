# Dockerfile Hardening

Shared action which runs hardening tools against one or more Dockerfiles.

## Conftest

This action runs [conftest](https://www.conftest.dev/) using Rego policies under [`conftest/policy`](./conftest/policy) to enforce hardening requirements for Dockerfiles committed in Grafana Labs repositories.

These policies include (but are not limited to) the following checks:

- Final image base must utilize a `scratch` or distroless image
- All image references (`FROM` and `COPY --from=<image>`) must be pinned to a digest (`@sha256:`, `@sha512:`, or `@blake3:`).
- `ADD` instructions must not fetch from a remote URL (`http://`, `https://`, `ftp://`) — supply chain risk; use `COPY` with a verified local file or `RUN curl` with explicit `sha256sum -c` verification
- `RUN` must not pipe remote content directly into a shell (`curl ... | bash`, `wget ... | sh`, etc.) — remote code execution at build time with no verification
- Container runtime must:
  - have a `USER` instruction in the final stage
  - not run as root (UID `0` or name `root`)
  - not invoke `sudo`
  - not `EXPOSE` a privileged port (`<1024`)
  - not `chmod` to world-writable (octal `777`/`0777` or any symbolic form granting write to `other`)
  - use exec form (JSON array) for `ENTRYPOINT` and `CMD` so the binary becomes PID 1 and receives signals

Advisory checks (warnings, not failures):

- `COPY .` in the final stage — may ship the source tree to the runtime image. Confirm the build context only contains intended artifacts (typical pattern: rely on `.dockerignore` or a curated build-context directory produced by an earlier CI step).

### Example

<!-- x-release-please-start-version -->

```yaml
uses: grafana/shared-workflows/actions/dockerfile-hardening@dockerfile-hardening/v0.1.0
with:
  dockerfiles: |
    Dockerfile
    images/api/Dockerfile
```

<!-- x-release-please-end-version -->

## Inputs

| Name               | Required | Default  | Description                                                                            |
| ------------------ | -------- | -------- | -------------------------------------------------------------------------------------- |
| `dockerfiles`      | yes      | —        | Newline- or space-separated list of Dockerfile paths, relative to `$GITHUB_WORKSPACE`. |
| `conftest-version` | no       | `0.55.0` | Version of conftest to install via `setup-conftest`.                                   |

## Local development

For testing the policies against a `Dockerfile` or testing the policies themselves, the `run-conftest.sh` script must be run from the action directory (`actions/dockerfile-hardening/`).

### Verify Dockerfiles

Run the `conftest` policy check against one or more Dockerfiles. Paths are resolved relative to `$GITHUB_WORKSPACE` (or `$PWD` if that variable is unset).

Single file:

```sh
DOCKERFILES="path/to/Dockerfile" bash ./run-conftest.sh verify-dockerfiles
```

Multiple files (use `$'...'` so `\n` becomes a real newline):

```sh
DOCKERFILES=$'Dockerfile\nimages/api/Dockerfile' bash ./run-conftest.sh verify-dockerfiles
```

Or pass them with repeated `-f`:

```sh
bash ./run-conftest.sh verify-dockerfiles -f Dockerfile -f images/api/Dockerfile
```

### Test Policies

Run the `conftest` policy tests to verify the rego policies:

```sh
bash ./run-conftest.sh test-policies
```

## Adding a policy

1. Add a `.rego` file under `conftest/policy/`. Declare `package main` and `import rego.v1`.
2. Add a matching `<name>_test.rego` file under `conftest/tests/` covering positive (rule fires) and negative (rule does not fire) cases via `with input as ...`.
3. Run `bash ./run-conftest.sh test-policies` from the action directory and confirm both the smoke test and the rego unit suite pass.
4. Update the policy list in this README to reflect the new check.

Shared helpers (USER-instruction parsing, stage-alias detection, image-ref extraction, etc.) live in [`conftest/policy/common.rego`](./conftest/policy/common.rego) and are visible to every policy and test in `package main` without explicit imports.

For negative test cases, assert against the specific rule under test rather than `not deny`:

```rego
test_x_not_denied if {
    msgs := deny with input as [...]
    every msg in msgs {
        not contains(msg, "<substring unique to your rule>")
    }
}
```

`not deny` asserts that **no** policy fires, which fails whenever the synthetic input trips a different rule (e.g., missing `USER`, missing digest pin). Scoping the assertion to the specific message keeps each test isolated.
