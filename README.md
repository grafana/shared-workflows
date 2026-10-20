# shared-workflows

[![OpenSSF Scorecard][scorecard image]][scorecard]

A centralized place to store reusable workflows and GitHub Actions used by Grafana Labs.
Refer to the [`actions/`](./actions) directory for the individual actions themselves.

> **Note:** As of May 4th 2026, all action releases are immutable. Once a version tag is created, it will not be moved or overwritten.

[scorecard]: https://scorecard.dev/viewer/?uri=github.com/grafana/shared-workflows
[scorecard image]: https://api.scorecard.dev/projects/github.com/grafana/shared-workflows/badge

## Custom Renovate config

This is a monorepo containing several Actions. When we release a workflow, we create a tag `<workflow name>/v<workflow version>`.

While Dependabot can update references to these actions, Renovate can't do it out of the box. It can, however, be configured to do so:

```json
{
  "packageRules": [
    {
      "matchPackageNames": ["grafana/shared-workflows"],
      "versioning": "regex:^(?<compatibility>.*)[-/]v?(?<major>\\d+)\\.(?<minor>\\d+)\\.(?<patch>\\d+)?$",
      "additionalBranchPrefix": "{{ lookup (split newVersion \"/\") 0 }}-",
      "commitMessagePrefix": "chore(deps):",
      "commitMessageAction": "update",
      "commitMessageTopic": "{{depName}}/{{ lookup (split newVersion \"/\") 0 }} action",
      "commitMessageExtra": "to {{ lookup (split newVersion \"/\") 1 }}"
    }
  ]
}
```

## Releasing actions and reusable workflows

Each action and reusable workflow is versioned and released individually, using tags of the form `<name>/v<semver version>`.
This lets consumers review changes before updating and allows Dependabot and Renovate to track new releases.

### Prepare your changes

Pull request titles must follow the [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/): a type, an optional scope, a colon, and a subject, all in lowercase.
Release Please uses these titles to generate the changelog and determine version bumps:

- `feat: add support for eating lollipops` causes a minor version bump.
- `feat!: rename foo input to bar` marks a breaking change and causes a major version bump.

Each pull request must also have a description explaining the change.
CI enforces conventional pull request titles and non-empty descriptions.

### Publish the release and update consumers

1. Merge all the changes you want to release into `main`.
2. Find the release PR for each action or reusable workflow you changed.
   On every push to `main`, our Release Please GitHub Action creates or updates a draft PR for each changed component, including its `CHANGELOG.md` updates.
3. Review the release PR, mark it ready for review, and merge it to create the next tagged release.
   This requires write access to the repository.
4. Update consumers to reference the new release.
   Publishing a release does not update their pinned versions.
   For org-wide checks, look for consumers in this repository and [security-github-actions](https://github.com/grafana/security-github-actions); updates may be needed in either or both.

For more information about Release Please, refer to the [release-please-action repository](https://github.com/googleapis/release-please-action).

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

Dependabot can update these SHA references when there are new versions.
If you include the complete tag name in a comment after the SHA, it can update the comment too.
For example:

```yaml
- uses: grafana/shared-workflows/actions/foo@0123456789abcdef0123456789abcdef01234567 # foo/v1.2.3
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

### Add new components to the Release Please configuration file

In order for components to be released, they must be in the [`release-please-config.json`](./release-please-config.json) file.
Always ensure new components are added to this file.

`README` files for each component should have embedded versions updated every time there is a new release.

Add a new entry that looks like this:

```json
  "packages": {
    "actions/my-new-action": {
      "package-name": "my-new-action",
      "extra-files": ["README.md"]
    },
  }
```

Also add the following block in the README file to update the embedded version:

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

Every semver-like string between the `x-release-please-start-version` and `x-release-please-end-version` comments is updated with the new version whenever the component is released.

### Deprecating shared workflows

When deprecating a shared workflow, follow this procedure:

1. Post a deprecation notice and warning in the affected action.
   Also provide a migration guide.
   Use [this commit](https://github.com/grafana/shared-workflows/commit/b6c252dc86cb65eaf2d8344d6d51ca07436214a2) as an example.
2. Once step 1 is merged, release the affected action so the latest version includes the deprecation notice.
   Renovate will automatically start to roll this version out.
3. On the agreed date, delete the action from `main`.
