# shared-workflows

A public-facing, centralized place to store reusable workflows and GitHub Actions used by Grafana Labs.
Refer to the [`actions/`](./actions) directory for the individual actions themselves.

## Notes

### Configure your IDE to run Prettier

[Prettier][] runs in CI to ensure that files are formatted correctly.
To format your code correctly before you commit, set up your IDE to run Prettier on save.

Or from the command line, you can run Prettier using [`npx`][npx]:

```sh
npx prettier --check .
```

Or, of course, install it in any other way you want.

[npx]: https://www.npmjs.com/package/npx
[prettier]: https://prettier.io/

### Pin versions

When using third-party actions, [always pin the version to a specific commit hash][hardening].
This ensures that the workflow always uses the same version of the action, even if the action's maintainers release a new version or update the Git tag.

Dependabot updates the SHA references when there are new versions.
If you include a tag in a comment after the SHA, it updates the comment too.
For example:

```yaml
- uses: action/foo@abcdef0123456789abcdef0123456789 # v1.2.3
```

[hardening]: https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#using-third-party-actions

### Use other `shared-workflows` actions with relative paths

When using other actions in this repository, use a relative path.
This means that workflows always use actions at the same commit.
To do this:

```yaml
- name: Checkout
  env:
    # In a composite action, these two need to be indirected via the
    # environment, as per the GitHub actions documentation:
    # https://docs.github.com/en/actions/learn-github-actions/contexts
    action_repo: ${{ github.action_repository }}
    action_ref: ${{ github.action_ref }}
  uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
  with:
    repository: ${{ env.action_repo }}
    ref: ${{ env.action_ref }}
    # Substitute your-action with a unique name (within `shared-repos` for your
    # action), so if multiple actions check `shared-workflows` out, they don't
    # overwrite each other.
    path: _shared-workflows-your-action

- name: Use another action
  uses: ./_shared-workflows-your-action/actions/some-action
  with:
    some-input: some-value
```

### Version actions and reusable workflows

To avoid breaking compatibility, each action or reusable workflow is versioned so that engineers consuming the component can review incoming changes.
This also helps automated update tools like Dependabot and Renovate to work effectively.

For every push to the `main` branch, Release Please creates or updates a draft PR with updates in the `CHANGELOG.md`.
Users with write access to the repository can mark a draft pull request as ready for review and then merge the pull request to create the next tagged release.

Release Please creates a pull request for every updated action or reusable workflow.
Each action released individually and use tags of the form `<NAME>-<SEMANTIC VERSION>`.

To generate the CHANGELOG for the current release, all pull request titles need to follow the [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/).
This means that the PR should start with a _type_ followed by a colon, and then a _subject_, all in lowercase.

Pull request titles with the `feat` type, like `feat: add support for eating lollipops`, cause minor version bumps.

Pull request titles that include an exclamation mark (`!`) after the type, like `feat!: rename foo input to bar`, cause major version bumps.

Each pull request must also have a description that explains the change.

CI enforces the use of conventional pull request titles and non-empty pull request descriptions.

For more information about Release Please, refer to their [GitHub repository](https://github.com/googleapis/release-please-action).

### Add new components to Release Please configuration file

For components to be released, they must be in the [`release-please-config.json`](./release-please-config.json) file.
