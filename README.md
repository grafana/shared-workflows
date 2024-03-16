# shared-workflows

A public-facing, centralized place to store reusable GitHub workflows and action
used by Grafana Labs. See the `actions/` directory for the individual actions
themselves.

## Notes

### Pin versions

When referencing third-party actions, [always pin the version to a specific
commit hash][hardening]. This ensures that the workflow will always use the same
version of the action, even if the action's maintainers release a new version or
if the tag itself is updated.

Dependabot can update these SHA references when there are new versions. If you
include a tag in a comment after the SHA, it can update the comment too. For
example:

```yaml
- uses: action/foo@abcdef0123456789abcdef0123456789 # v1.2.3
```

[hardening]: https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions
