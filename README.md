# shared-workflows

[![OpenSSF Scorecard][scorecard image]][scorecard]

A public-facing, centralized place to store reusable GitHub workflows and action
used by Grafana Labs. See the `actions/` directory for the individual actions
themselves.

[scorecard]: https://scorecard.dev/viewer/?uri=github.com/grafana/shared-workflows
[scorecard image]: https://api.scorecard.dev/projects/github.com/grafana/shared-workflows/badge

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

### Refer to other `shared-workflows` actions using relative paths

When referencing other actions in this repository, use a relative path. This
will ensure actions in this repo are always used at the same commit. To do this:

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
    # substitute your-action with a unique name (within `shared-repos` for your
    # action), so if multiple actions check `shared-workflows` out, they don't
    # overwrite each other
    path: _shared-workflows-your-action
    persist-credentials: false

- name: Use another action
  uses: ./_shared-workflows-your-action/actions/some-action
  with:
    some-input: some-value

# This step ensures the checkout directory is cleaned up even if previous steps fail
  - name: Cleanup checkout directory
    if: ${{ !cancelled() }}
    shell: bash
    run: |
      # Check that the directory looks OK before removing it
      if ! [ -d "_shared-workflows-push-to-gar/.git" ]; then
      echo "::warning Not removing shared workflows directory: doesn't look like a git repository"
      exit 0
      fi
      rm -rf _shared-workflows-push-to-gar
```

### Use separate files for shell scripts so they're linted

Instead of embedding a shell script in the `run` string, write a separate script and refer to that.

For example, don't use the step:

```yaml
id: echo-success
shell: bash
run: |
  echo "Success!"
```

Instead, create the file `echo-success.bash` in the same directory and use the step:

```yaml
id: echo-success
shell: bash
run: ./echo-success.bash
```

### Releasing a version of a component in shared-workflows

When working with `shared-workflows`, it's essential to avoid breaking backwards compatibility. To ensure this, we must provide releasable actions for engineers to review incoming changes. This also helps automated update tools like `dependabot` and `renovate` to work effectively.

Upon push to main, a draft PR with updates in the CHANGELOG.md will be updated or created. This can be undrafted and merged at any time to create the next tagged release. Since we're a monorepo, one PR will be created for each action/reusable workflow that has been updated. They can be released individually and tags will be of the form `<name>-<semver version>`.

In order for the release action to work properly, which means to generate a CHANGELOG for the current release, the pull request titles need to follow the [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/). This means that the PR should start with a `type` followed by a colon, and then a `subject` - all in lowercase.

Minor version bumps are indicated by new features: `feat: add support for eating lollipops`. Major version bumps are indicated by an `!` after the type in the commit message/description, for example: `feat!: rename foo input to bar`.

Also, the PR description needs to be filled and should never be empty.

Failing to follow any of the aforementioned necessary steps, will lead to CI failing on your pull request.

More about how the upstream action works can be found [here](https://github.com/googleapis/release-please-action).

### Add new components to Release Please config file

In order for components to be released, they must be in the [release-please-config.json](./release-please-config.json) file. Always ensure new components are added to this file.

`README` files for each component should have embedded versions updated every time there is a new release.

Add a new entry which looks like this:

```json
  "packages": {
    "actions/my-new-action": {
      "package-name": "my-new-action",
      "extra-files": ["README.md"]
    },
  }
```

Also, the following block should be added in the README file which will be responsible for updating the version with a new one:

`README.md`:

````markdown
# my-new-action

This is my new action which does awesome stuff!

<!-- x-release-please-start-version -->

```yaml
name: My new action
on:
  pull_request:

jobs:
  my-new-action:
    runs-on: ubuntu-latest

    steps:
      - id: do-stuff
        uses: grafana/shared-workflows/actions/my-new-action@my-new-action/v1.0.0
```

<!-- x-release-please-end-version -->
````

Every semver-like string between the `x-release-please-start-version` and `x-release-please-end-version` comments will be updated with the new version whenever the component is released.
