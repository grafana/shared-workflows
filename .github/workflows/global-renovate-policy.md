# Renovate repository policy

This workflow checks for explicit repository-wide Renovate disablement. Register
`global-renovate-policy.yml` from this public repository as an organization ruleset
workflow. Repositories do not need to add their own caller workflow. The existing
`validate-renovate-config` action and its callers are unchanged.

The ruleset controls enforcement: use **Evaluate** to collect results, **Active**
for the announced brownout, and return to **Evaluate** at its fixed end time.
Merging this workflow does not activate an organization ruleset. A failing check
in Evaluate mode remains non-blocking unless it is independently required by
another rule; do not add a separate required status check for this experiment.

The job reads the proposed PR merge commit or merge-group commit. It checks every
PR, including changes unrelated to Renovate. A PR removing the disablement can
pass. The checker comes from `grafana/shared-workflows` at `github.workflow_sha`,
the commit containing the required workflow, rather than from the target repo.
Pin the workflow revision in the ruleset and verify both checkout SHAs during the
canary before widening scope.

| Configuration                                                      | Result                                  |
| ------------------------------------------------------------------ | --------------------------------------- |
| Top-level `enabled: false`                                         | Fail                                    |
| Top-level `force.enabled: false`                                   | Fail                                    |
| Direct `extends: [":disableRenovate"]`                             | Fail                                    |
| Scoped `enabled: false` in package rules, managers or update types | Pass                                    |
| No local configuration                                             | Pass; self-hosted global defaults apply |
| Malformed, unreadable, oversized or symlinked configuration        | Error                                   |
| Disablement inherited through another preset                       | Not detected                            |
| Catch-all package rules or combinations that suppress updates      | Not detected                            |

The [checker README](../../scripts/check-renovate-enabled/README.md) documents
configuration discovery, dependency installation and local tests. This is a
policy check, not full Renovate schema validation or proof of successful scanning.
Repositories excluded from discovery or lacking App access still need a separate
coverage audit.

Use the [brownout runbook](https://github.com/grafana/deployment_tools/blob/master/docs/platform/renovate/enablement-brownout.md)
for scope, owner communications, activation, observations and rollback. Background:
[deployment_tools#791627](https://github.com/grafana/deployment_tools/issues/791627)
and the original checker proposal in [#2352](https://github.com/grafana/shared-workflows/pull/2352).

GitHub ruleset workflows ignore event path filters and do not automatically run
for all PRs already open when the rule is added. The canary must cover an existing
PR, a new PR, a configuration-fix PR, a fork PR and a merge queue where used. See
[GitHub's ruleset workflow troubleshooting](https://docs.github.com/en/enterprise-cloud@latest/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/troubleshooting-rules).
