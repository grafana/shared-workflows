# pr-retarget

Route feature pull requests from one base branch to another without changing the
repository's default branch. Defaults to `main → dev`; branch names and the
persistent override label are configurable.

## Behavior

| Event or condition                                               | Result                                      |
| ---------------------------------------------------------------- | ------------------------------------------- |
| Open, reopen, edit, label or unlabel an open PR targeting `main` | Change its base to `dev`, unless exempt     |
| Same-repository `dev → main` PR                                  | Leave it alone (promotion/rollup exemption) |
| PR carrying `allow-main`                                         | Leave its base alone                        |
| Maintainer comments `/target main`                               | Add `allow-main`, then restore `main`       |
| Maintainer comments `/target dev`                                | Remove `allow-main`, then target `dev`      |
| Closed PR, unrelated base branch, ordinary comment               | No change                                   |

Commands must be the entire comment (surrounding whitespace is allowed). Only
users whose **current repository permission** is write, maintain or admin can
issue commands. PR authorship or organization membership alone is insufficient.
For custom branches, the commands use those names, e.g. `/target trunk` and
`/target develop`. A fork branch named `dev` does not qualify for the rollup
exemption.

Removing `allow-main` from a PR still targeting `main` resumes automatic routing
when the caller subscribes to `unlabeled`. Adding the label does not itself
restore `main`; use the comment command for that.

## Setup

1. Ensure both branches exist in each consuming repository.
2. Create the configured override label once, for example:

   ```sh
   gh label create allow-main --repo grafana/YOUR-REPO --color FEF2C0 \
     --description 'Maintainer-approved exception to development-branch routing'
   ```

3. Install the caller workflow below on the repository's default branch. On
   GitHub.com, both `issue_comment` and `pull_request_target` run from the default
   branch, not the PR's head. For older GitHub Enterprise Server versions where
   `pull_request_target` uses the base branch, install it on both base branches
   too.
4. Configure a GitHub App/Vault permission set scoped to the consuming repository
   with **Pull requests: write**, **Issues: write** and **Contents: read**. The
   token also needs access to the collaborator-permission endpoint (see
   [GitHub's permission lookup documentation](https://docs.github.com/en/rest/collaborators/collaborators#get-repository-permissions-for-a-user)).
5. Make sure PR CI includes `edited` events so it runs against a changed base.
   Existing branch protection and required checks still apply. This action never
   merges a PR, approves it or enables auto-merge.

### Caller workflow

Save as `.github/workflows/pr-retarget.yml`. Replace the App and permission-set
names with your repository's provisioned Vault configuration. The `pr-retarget`
version below is the planned initial release; until published, use the reviewed
commit SHA. Pin all action references to reviewed commit SHAs for production.

<!-- x-release-please-start-version -->

```yaml
name: Route PRs to development

on:
  pull_request_target:
    types: [opened, reopened, edited, labeled, unlabeled]
  issue_comment:
    types: [created]

permissions: {}

# Serialize both event types for the same PR. Do not cancel a run halfway through
# persisting an override and changing the base. The action reads fresh PR state.
concurrency:
  group: pr-retarget-${{ github.repository }}-${{ github.event.pull_request.number || github.event.issue.number }}
  cancel-in-progress: false

jobs:
  retarget:
    if: >-
      github.event_name == 'pull_request_target' ||
      (github.event.issue.pull_request &&
       (github.event.comment.body == '/target main' ||
        github.event.comment.body == '/target dev'))
    runs-on: ubuntu-latest
    timeout-minutes: 5
    permissions:
      id-token: write
    steps:
      - name: Get routing token
        id: token
        uses: grafana/shared-workflows/actions/create-github-app-token@create-github-app-token/v1.0.0
        with:
          github_app: YOUR-GITHUB-APP
          permission_set: YOUR-PR-ROUTING-PERMISSION-SET

      - name: Route PR
        uses: grafana/shared-workflows/actions/pr-retarget@pr-retarget/v0.1.0
        with:
          github-token: ${{ steps.token.outputs.token }}
```

<!-- x-release-please-end-version -->

The caller's comment filter is intentionally exact to avoid issuing Vault tokens
for ordinary comments. Remove the filter if you need commands with surrounding
whitespace, or change the two command strings when configuring different branch
names. The action independently validates commands and permissions.

### CI after retargeting

Prefer an **App installation token** for `github-token`. GitHub does not start
Actions workflows for `edited` events generated using `GITHUB_TOKEN`; using it
can leave CI missing or checking the old base. App tokens allow the base-change
event to start normal CI, provided the CI workflow subscribes to `edited`:

```yaml
on:
  pull_request:
    types: [opened, synchronize, reopened, edited]
```

See [GitHub's workflow-trigger rules](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/trigger-a-workflow).

## Security and operational notes

- Metadata-only: the action does **not** check out, build or execute PR code.
  Do not add a PR-head checkout to this privileged workflow.
- Comments are read as data, never interpolated into shell or JavaScript source.
- A source-branch override is persisted before the base changes, so subsequent
  events cannot bounce an intentional `main` PR back to `dev`.
- The action fetches current PR state instead of trusting queued event snapshots.
  Keep the caller's per-PR concurrency group when adapting the example.
- Base changes can alter the diff, invalidate review context and change required
  checks. Apply the same review/protection policies to intentional `main` PRs.
- API failures fail the job. Missing destination branches and override labels
  fail before base changes; provisioning is not silently skipped.
- A routing explanation is posted only after an actual base change. If posting
  it fails, the base change is not rolled back; inspect the failed run.

## Inputs

<!-- BEGIN_INPUTS -->

| Name             | Type   | Required | Default      | Description                                                                                       |
| ---------------- | ------ | -------- | ------------ | ------------------------------------------------------------------------------------------------- |
| `github-token`   | String | Yes      |              | GitHub token with pull requests write permission. Prefer an App token so base changes trigger CI. |
| `override-label` | String | No       | `allow-main` | Persistent opt-out label applied by the source-branch comment command.                            |
| `source-branch`  | String | No       | `main`       | Branch from which newly opened or edited pull requests are redirected.                            |
| `target-branch`  | String | No       | `dev`        | Development branch to which pull requests are redirected.                                         |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name          | Description                                                                     |
| ------------- | ------------------------------------------------------------------------------- |
| `base-branch` | Current base branch of the processed pull request, or empty for ignored events. |
| `changed`     | Whether the action changed the pull request base branch.                        |

<!-- END_OUTPUTS -->

## Development

No bundled code or runtime dependency installation is needed. The composite
action uses pinned `actions/github-script` and loads its own JavaScript module
from the runner's action cache. Unit tests use mocked GitHub APIs and never
mutate a real repository:

```sh
node --test actions/pr-retarget/retarget.test.mjs
```
