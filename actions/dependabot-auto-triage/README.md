# Auto Dismiss Dependabot Alerts

A GitHub composite action to automatically dismiss Dependabot alerts based on manifest paths using glob patterns.

## Usage

This action can be used in a workflow to automatically dismiss Dependabot alerts for specific manifest paths. This is particularly useful for repositories with dependencies that are not directly used in production or for which vulnerabilities may not be relevant.

### Example Workflow

Create a workflow file (e.g., `.github/workflows/auto-dismiss-dependabot-alerts.yml`) with the following content:

<!-- x-release-please-start-version -->

```yaml
name: Auto Dismiss Dependabot Alerts

on:
  # Run daily to dismiss new alerts
  schedule:
    - cron: "0 0 * * *"

  # Allow manual triggering
  workflow_dispatch:

jobs:
  auto-dismiss:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    steps:
      - name: Checkout repository
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1

      # Get a GitHub App token with Dependabot alerts permissions via the GitHub App Token Broker
      - name: Generate token
        id: generate-token
        uses: grafana/shared-workflows/actions/create-github-app-token@5bbb526b1728ca57fbe1750c4c27ab6ab06171c2
        with:
          github_app: dependabot-auto-triage-app

      # Use the token with the auto-triage action
      - name: Auto Dismiss Dependabot Alerts
        uses: grafana/shared-workflows/actions/dependabot-auto-triage@dependabot-auto-triage/v1.1.3
        with:
          token: ${{ steps.generate-token.outputs.token }}
          paths: |
            terraform/modules/**/*.json
            docker/vendor/**
            ksonnet/lib/argo-workflows/charts/**/*.json
          dismissal-reason: "not_used"
          dismissal-comment: "These dependencies are not used in production and pose no risk"
          close-prs: "true" # Optional: close associated Dependabot PRs
```

<!-- x-release-please-end-version -->

### Inputs

<!-- BEGIN_INPUTS -->

| Name                | Type    | Required | Default                                               | Description                                                                                                       |
| ------------------- | ------- | -------- | ----------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `alert-types`       | String  | No       | `dependency`                                          | Comma-separated list of alert types to dismiss (default: "dependency")                                            |
| `close-prs`         | Boolean | No       | `false`                                               | Whether to close associated pull requests when dismissing alerts                                                  |
| `dismissal-comment` | String  | No       | `Auto-dismissed based on manifest path configuration` | Default comment to add when dismissing alerts                                                                     |
| `dismissal-reason`  | String  | No       | `not_used`                                            | Default reason for dismissal. One of `fix_started`, `inaccurate`, `no_bandwidth`, `not_used` or `tolerable_risk`. |
| `paths`             | String  | Yes      |                                                       | Multi-line list of glob patterns to match manifest paths to dismiss                                               |
| `token`             | String  | Yes      |                                                       | GitHub token with permissions to dismiss alerts                                                                   |

<!-- END_INPUTS -->

### How It Works

1. The action fetches all open Dependabot alerts for the repository
2. For each alert, it checks if the manifest path matches any of the provided glob patterns
3. If `close-prs` is enabled, it fetches associated pull requests for matching alerts
4. For each matching alert, it optionally closes the associated pull request first (if `close-prs` is true)
5. It then dismisses the alert with the specified reason and comment

### Glob Pattern Syntax

The action uses [minimatch](https://github.com/isaacs/minimatch) for glob pattern matching. Some common patterns:

- `**/*.json` - Match all JSON files in any directory
- `terraform/modules/**` - Match all files in terraform/modules and subdirectories
- `docker/vendor/**/package-lock.json` - Match all package-lock.json files in docker/vendor and subdirectories
- `ksonnet/lib/*/charts/**` - Match all files in any charts subdirectory under ksonnet/lib/\*/

### Permissions

Due to API limitations, accessing and dismissing Dependabot alerts requires a GitHub App token with specific permissions. The standard `GITHUB_TOKEN` does not have sufficient access to the Dependabot API, even when `security-events: write` permissions are specified.

#### GitHub App Requirements

To use this action, you need:

1. A GitHub App with the following permissions:
   - Repository permissions:
     - **Dependabot alerts**: Read & Write
     - **Pull requests**: Read & Write (only required if `close-prs` is set to `true`)

2. The GitHub App needs to be installed on your repository or organization

3. The GitHub App needs to be configured in the GitHub App Token Broker for your repository

The example workflow above demonstrates using the [`create-github-app-token`](../create-github-app-token/README.md) action to get a token for the `dependabot-auto-triage-app` GitHub App from the GitHub App Token Broker, so the App's private key is never exposed to the workflow.

If you're experiencing "Resource not accessible by integration" errors, this indicates that the token being used doesn't have the necessary permissions to access the Dependabot API.

## License

This action is licensed under the same license as the parent repository.
