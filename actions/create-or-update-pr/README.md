# create-or-update-pr

Creates or updates a pull request from file changes in the working tree. A lightweight replacement for third-party PR creation actions using only `git` and `gh` CLI.

The action:

1. Stages the specified file paths
2. Exits cleanly if no changes are detected
3. Creates a branch, commits, and pushes
4. Creates a new PR or updates the existing one on that branch
5. Outputs the PR number, URL, and operation performed

Assumes `actions/checkout` has already been called.

## Inputs

<!-- BEGIN_INPUTS -->

| Name             | Type    | Required | Default                                                 | Description                                                                                                                         |
| ---------------- | ------- | -------- | ------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `add-paths`      | String  | Yes      |                                                         | A comma, newline, or space-separated list of file paths to commit. Paths follow git's pathspec syntax (e.g. '_.json', 'docs/_.md'). |
| `base-branch`    | String  | No       | `main`                                                  | Base branch for the PR                                                                                                              |
| `body`           | String  | Yes      |                                                         | Pull request body (markdown)                                                                                                        |
| `branch`         | String  | Yes      |                                                         | Branch name to push to (e.g. updater/my-update)                                                                                     |
| `commit-message` | String  | Yes      |                                                         | Commit message for the change                                                                                                       |
| `draft`          | Boolean | No       | `false`                                                 | Create a draft pull request (true/false)                                                                                            |
| `git-user-email` | String  | No       | `41898282+github-actions[bot]@users.noreply.github.com` | Git user email for the commit (used as both author and committer)                                                                   |
| `git-user-name`  | String  | No       | `github-actions[bot]`                                   | Git user name for the commit (used as both author and committer)                                                                    |
| `labels`         | String  | No       |                                                         | A comma or newline-separated list of labels to add to the pull request                                                              |
| `reviewers`      | String  | No       |                                                         | A comma or newline-separated list of GitHub usernames to request review from                                                        |
| `title`          | String  | Yes      |                                                         | Pull request title                                                                                                                  |
| `token`          | String  | No       | `${{ github.token }}`                                   | GitHub token for push and PR creation (contents: write, pull-requests: write)                                                       |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name                     | Description                                                      |
| ------------------------ | ---------------------------------------------------------------- |
| `pull-request-number`    | The pull request number                                          |
| `pull-request-operation` | The operation performed by the action: created, updated, or none |
| `pull-request-url`       | The URL of the pull request                                      |

<!-- END_OUTPUTS -->

## Permissions

The token requires:

- `contents: write` -- to push the branch
- `pull-requests: write` -- to create/update the PR

## Examples

### Auto-update an OpenAPI spec

<!-- x-release-please-start-version -->

```yaml
name: Update OpenAPI spec
on:
  repository_dispatch:
    types: [spec-updated]

jobs:
  update:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
      - name: Download latest spec
        run: curl -o api/spec.json https://example.com/openapi.json
      - name: Create or update PR
        id: cpr
        uses: grafana/shared-workflows/actions/create-or-update-pr@create-or-update-pr/v0.1.0
        with:
          branch: updater/openapi-spec
          commit-message: "Update OpenAPI spec"
          title: "Update OpenAPI spec"
          body: "Automated update of the OpenAPI specification."
          add-paths: api/spec.json
      - name: Check outputs
        if: ${{ steps.cpr.outputs.pull-request-number }}
        run: |
          echo "PR #${{ steps.cpr.outputs.pull-request-number }}"
          echo "URL: ${{ steps.cpr.outputs.pull-request-url }}"
          echo "Operation: ${{ steps.cpr.outputs.pull-request-operation }}"
```

<!-- x-release-please-end-version -->

### With labels, reviewers, and draft

```yaml
- uses: grafana/shared-workflows/actions/create-or-update-pr@create-or-update-pr/v0.1.0
  with:
    branch: updater/my-update
    commit-message: "Automated update"
    title: "Automated update"
    body: "This PR was created automatically."
    add-paths: "path/to/file1.json, path/to/file2.json"
    labels: "automated, dependencies"
    reviewers: "octocat, hubot"
    draft: "true"
```

### Custom git identity

```yaml
- uses: grafana/shared-workflows/actions/create-or-update-pr@create-or-update-pr/v0.1.0
  with:
    branch: updater/my-update
    commit-message: "Automated update"
    title: "Automated update"
    body: "This PR was created automatically."
    add-paths: path/to/file.json
    git-user-name: "my-bot"
    git-user-email: "my-bot@example.com"
```

### Using a GitHub App token

```yaml
- uses: grafana/shared-workflows/actions/create-or-update-pr@create-or-update-pr/v0.1.0
  with:
    branch: updater/my-update
    commit-message: "Automated update"
    title: "Automated update"
    body: "This PR was created automatically."
    add-paths: path/to/file.json
    token: ${{ steps.app-token.outputs.token }}
```

### Multiple pathspecs

The `add-paths` input supports git pathspec syntax:

```yaml
- uses: grafana/shared-workflows/actions/create-or-update-pr@create-or-update-pr/v0.1.0
  with:
    branch: updater/docs
    commit-message: "Update generated docs"
    title: "Update generated docs"
    body: "Automated docs update."
    add-paths: |
      docs/*.md
      api/spec.json
```

### Local dry-run

The underlying script supports a `DRY_RUN` mode for local testing:

```bash
DRY_RUN=1 \
PR_BRANCH=updater/test \
COMMIT_MSG="Test commit" \
PR_TITLE="Test PR" \
PR_BODY="Testing." \
ADD_PATHS="path/to/file" \
./actions/create-or-update-pr/create-or-update-pr.sh
```
