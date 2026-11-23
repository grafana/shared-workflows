# Renovate policy regression tests

The production checker is embedded in
[`reusable-zizmor.yml`](../../.github/workflows/reusable-zizmor.yml). Keeping it in
the trusted reusable workflow avoids executing a script from the calling
repository's PR checkout or introducing a second checkout. The tests extract
and execute that exact Python code against temporary repository fixtures.

Run with the versions pinned in the existing
[`test-validate-zizmor-config.yml`](../../.github/workflows/test-validate-zizmor-config.yml)
test job:

```sh
uv run --no-project --no-config --no-env-file --isolated \
  --with pyyaml==6.0.3 --with json5==0.12.1 \
  -- python3 -I scripts/renovate-policy/test_renovate_policy.py -v
```

Coverage includes direct disablement, allowed scoped exclusions, JSON5, every
supported config location, first-file precedence, missing config, malformed
config, warning/enforcement modes, symlinks, file-size limits, untrusted input,
and PR/merge-queue event handling. These tests run as an extra step in the
existing Zizmor policy test job; they add no CI job.

See [the workflow documentation](../../.github/workflows/reusable-zizmor.md#renovate-enablement-policy-and-brownouts)
for the policy's limits and the rollout/rollback procedure.
