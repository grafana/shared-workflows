# shared-workflows

Welcome to Platform Productivity's shared workflows repository!

This public-facing repository contains a collection of GitHub Actions workflows designed to streamline and automate common tasks across Grafana projects. These shared workflows are meant to be invoked from multiple repositories, promoting consistency and efficiency in our development processes.

Workflows in this repository fall into two categories:

1. **Common Workflows:** These are highly generic actions that any Grafana repository might need in its CI pipeline. Examples include logging into Docker, pushing Docker images, tagging Docker images, logging into AWS, publishing RPMs, and more. These actions are stored in the `actions/` folder and are specifically maintained by the platform-productivity team.

2. **Group-Specific Workflows:** These are workflows tailored to specific groups but used by multiple repositories. Examples include tests specific to a project that need to be called from multiple repos or steps to publish a plugin for a repo. Workflows in this category are owned by the team that contributed them and are maintained accordingly. These actions should be stored in a folder named after the group that's responsible for them (Ex: `grafana/`).

**Contributor Requirements:**
- **Updating Contributor Information:** Contributors are expected to update the `CODEOWNERS` file with any new workflows they contribute to this repository.
- **Ownership:** Workflows in category 2 are owned by the team that contributed them, ensuring accountability and maintenance by the relevant stakeholders.

By centralizing these workflows here, we ensure that all our projects can easily leverage these functionalities without duplicating effort or reinventing the wheel. Contributions that enhance the usability, reliability, and security of these shared workflows are highly encouraged.

Thank you for contributing to our collaborative development efforts!


## Notes

### Pin versions

When referring to third-party actions, it's important to [pin the version to a specific commit hash][hardening]. This guarantees that the workflow consistently uses the intended version of the action, unaffected by future releases or updates to the tag itself.

Dependabot is capable of updating these SHA references automatically when new versions are available. If you add a tag in a comment following the SHA, Dependabot can also update the comment accordingly. For instance:

```yaml
- uses: action/foo@abcdef0123456789abcdef0123456789 # v1.2.3
```

[hardening]: https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions
