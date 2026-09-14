# Global dismiss-changed-approvals workflow

This workflow triggers the dismiss-changed-approvals action, which dismisses
pull-request approvals only when the reviewed content actually changed (see
the action's README for the exact rules). It is intended to be enforced as an
org-wide required workflow (repository ruleset, "require workflows to pass")
on repositories that opt into content-based approval dismissal instead of the
native "Dismiss stale pull request approvals" setting.

Enrollment requirements for a repository:

- Native "Dismiss stale pull request approvals" disabled (this workflow is
  the dismissal authority).
- Required approvals, CODEOWNERS review, signed-commit enforcement and
  "restrict who can dismiss reviews" stay enabled.
- "Allow GitHub Actions to create and approve pull requests" stays disabled,
  so the workflow token can dismiss but never approve.
- No reliance on fork contributions (fork PRs are skipped: read-only token).

The `dry-run` input is set to `true` during rollout (shadow mode): verdicts
are logged and written to the job summary without dismissing anything. Flip it
to `false` to enforce.
