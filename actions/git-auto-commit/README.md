# git-auto-commit

Commits changes in the working tree and pushes to the current branch. A lightweight replacement for `stefanzweifel/git-auto-commit-action` using only `git` and pre-installed tools.

Typically used in `pull_request` workflows to auto-commit generated files (formatters, code generators, etc.) back to the PR branch.

The action:

1. Stages files matching the file pattern
2. Checks for real changes (ignores CRLF-only differences)
3. Commits with the configured identity and message
4. Pushes to the target branch

## Inputs

| Name              | Type     | Description                                      | Default                                                                       | Required |
| ----------------- | -------- | ------------------------------------------------ | ----------------------------------------------------------------------------- | -------- |
| `commit-message`  | `string` | Commit message                                   | `Apply automatic changes`                                                     | false    |
| `branch`          | `string` | Branch to push to                                | `${{ github.head_ref }}`                                                      | false    |
| `file-pattern`    | `string` | Space-separated file patterns for `git add`      | `.`                                                                           | false    |
| `git-user-name`   | `string` | Git user name for the commit                     | `github-actions[bot]`                                                         | false    |
| `git-user-email`  | `string` | Git user email for the commit                    | `41898282+github-actions[bot]@users.noreply.github.com`                       | false    |
| `token`           | `string` | GitHub token for push authentication             | `${{ github.token }}`                                                         | false    |
| `commit-options`  | `string` | Additional flags for `git commit`                |                                                                               | false    |
| `push-options`    | `string` | Additional flags for `git push`                  |                                                                               | false    |
| `skip-push`       | `string` | Skip the push step (`true`/`false`)              | `false`                                                                       | false    |

## Outputs

| Name               | Description                                                      |
| ------------------ | ---------------------------------------------------------------- |
| `changes-detected` | `true` if changes were committed, `false` if working tree clean  |
| `commit-hash`      | Full SHA of the created commit (empty if no changes)             |

## Permissions

The token requires:

- `contents: write` -- to push the commit

## Examples

### Basic usage

<!-- x-release-please-start-version -->

```yaml
name: Format and commit
on:
  pull_request:

permissions:
  contents: write

jobs:
  format:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.head_ref }}
      - name: Run formatter
        run: ./format.sh
      - uses: grafana/shared-workflows/actions/git-auto-commit@git-auto-commit/v0.1.0
        with:
          commit-message: "Apply formatting"
```

<!-- x-release-please-end-version -->

### Commit specific files only

```yaml
- uses: grafana/shared-workflows/actions/git-auto-commit@git-auto-commit/v0.1.0
  with:
    commit-message: "Update generated clients"
    file-pattern: "src/generated/ docs/api/"
```

### Check if changes were committed

```yaml
- uses: grafana/shared-workflows/actions/git-auto-commit@git-auto-commit/v0.1.0
  id: auto-commit
  with:
    commit-message: "Auto-format"
- name: Show result
  run: |
    echo "Changes detected: ${{ steps.auto-commit.outputs.changes-detected }}"
    echo "Commit hash: ${{ steps.auto-commit.outputs.commit-hash }}"
```

### Custom git identity

```yaml
- uses: grafana/shared-workflows/actions/git-auto-commit@git-auto-commit/v0.1.0
  with:
    commit-message: "Update generated files"
    git-user-name: "my-bot"
    git-user-email: "my-bot@example.com"
```

### Using a GitHub App token

```yaml
- uses: grafana/shared-workflows/actions/git-auto-commit@git-auto-commit/v0.1.0
  with:
    commit-message: "Auto-format"
    token: ${{ steps.app-token.outputs.token }}
```

### Skip push (commit only)

```yaml
- uses: grafana/shared-workflows/actions/git-auto-commit@git-auto-commit/v0.1.0
  with:
    commit-message: "Local commit"
    skip-push: "true"
```

### With commit and push options

```yaml
- uses: grafana/shared-workflows/actions/git-auto-commit@git-auto-commit/v0.1.0
  with:
    commit-message: "Auto-format"
    commit-options: "--no-verify"
    push-options: "--force"
```

### On push events

When triggered by `push` events, `github.head_ref` is empty. The action detects this and pushes to the current branch:

```yaml
on:
  push:
    branches: [main]

jobs:
  update:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
      - run: ./generate-docs.sh
      - uses: grafana/shared-workflows/actions/git-auto-commit@git-auto-commit/v0.1.0
        with:
          commit-message: "Update generated docs"
```

## Migration from stefanzweifel/git-auto-commit-action

| stefanzweifel input    | This action's input  | Notes                                           |
| ---------------------- | -------------------- | ----------------------------------------------- |
| `commit_message`       | `commit-message`     | Same behavior (uses kebab-case)                 |
| `branch`               | `branch`             | Same default (`github.head_ref`)                |
| `file_pattern`         | `file-pattern`       | Same behavior (uses kebab-case)                 |
| `commit_user_name`     | `git-user-name`      | Same behavior                                   |
| `commit_user_email`    | `git-user-email`     | Same behavior                                   |
| `commit_author`        | *(not supported)*    | Uses same identity for author and committer     |
| `commit_options`       | `commit-options`     | Same behavior (uses kebab-case)                 |
| `push_options`         | `push-options`       | Same behavior (uses kebab-case)                 |
| `skip_push`            | `skip-push`          | Same behavior (uses kebab-case)                 |
| `repository`           | *(not supported)*    | Use `working-directory` on the step instead      |
| `add_options`          | *(not supported)*    | Use `file-pattern` for most cases               |
| `status_options`       | *(not supported)*    | Rarely needed                                   |
| `skip_dirty_check`     | *(not supported)*    | Rarely needed                                   |
| `skip_fetch`           | *(not applicable)*   | Action does not fetch                           |
| `skip_checkout`        | *(not supported)*    | Rarely needed                                   |
| `disable_globbing`     | *(not applicable)*   | Handled internally                              |
| `create_branch`        | *(not supported)*    | Use `create-or-update-pr` action instead        |
| `tag_name`             | *(not supported)*    | Tagging is a separate concern                   |
| `tagging_message`      | *(not supported)*    | Tagging is a separate concern                   |
| `create_git_tag_only`  | *(not supported)*    | Tagging is a separate concern                   |
