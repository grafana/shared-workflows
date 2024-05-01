# shared-workflows

A public-facing, centralized place to store reusable GitHub workflows and action
used by Grafana Labs. See the `actions/` directory for the individual actions
themselves.

## Notes

### Configure your IDE to run Prettier

[Prettier] will run in CI to ensure that files are formatted correctly. To ensure
that your code is formatted correctly before you commit, set up your IDE to run
Prettier on save.

Or from the commandline, you can run Prettier using [`npx`][npx]:

```sh
npx prettier --check .
```

Or, of course, install it in any other way you want.

[npx]: https://www.npmjs.com/package/npx
[prettier]: https://prettier.io/

### Pin versions

When referencing third-party actions, [always pin the version to a specific
commit hash][hardening]. This ensures that the workflow will always use the same
version of the action, even if the action's maintainers release a new version or
if the tag itself is updated.

Dependabot can update these SHA references when there are new versions. If you
include a tag in a commend after the SHA, it can update the comment too. For
example:

```yaml
- uses: action/foo@abcdef0123456789abcdef0123456789 # v1.2.3
```

[hardening]: https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions
