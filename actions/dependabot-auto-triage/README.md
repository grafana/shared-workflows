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
        uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2

      # Get GitHub App token with Dependabot alerts permissions
      - name: Retrieve GitHub App secrets
        id: get-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets-v1.1.0
        with:
          common_secrets: |
            DEPENDABOT_AUTO_TRIAGE_APP_ID=dependabot-auto-triage:app-id
            DEPENDABOT_AUTO_TRIAGE_APP_PRIVATE_KEY=dependabot-auto-triage:private-key

      - name: Generate token
        id: generate-token
        uses: actions/create-github-app-token@3ff1caaa28b64c9cc276ce0a02e2ff584f3900c5 # v2.0.2
        with:
          app-id: ${{ env.DEPENDABOT_AUTO_TRIAGE_APP_ID }}
          private-key: ${{ env.DEPENDABOT_AUTO_TRIAGE_APP_PRIVATE_KEY }}

      # Use the token with the auto-triage action
      - name: Auto Dismiss Dependabot Alerts
        uses: grafana/shared-workflows/actions/dependabot-auto-triage@dependabot-auto-triage-v0.1.0
        with:
          token: ${{ steps.generate-token.outputs.token }}
          paths: |
            terraform/modules/**/*.json
            docker/vendor/**
            ksonnet/lib/argo-workflows/charts/**/*.json
          dismissal-reason: "not_used"
          dismissal-comment: "These dependencies are not used in production and pose no risk"
```

<!-- x-release-please-end-version -->

### Inputs

| Name                | Description                                                                                                       | Required | Default                                               |
| ------------------- | ----------------------------------------------------------------------------------------------------------------- | -------- | ----------------------------------------------------- |
| `token`             | GitHub token with permissions to dismiss alerts                                                                   | Yes      | N/A                                                   |
| `alert-types`       | Comma-separated list of alert types to dismiss                                                                    | No       | `dependency`                                          |
| `paths`             | Multi-line list of glob patterns to match manifest paths to dismiss                                               | Yes      | N/A                                                   |
| `dismissal-comment` | Default comment to add when dismissing alerts                                                                     | No       | `Auto-dismissed based on manifest path configuration` |
| `dismissal-reason`  | Default reason for dismissal (options: `fix_started`, `inaccurate`, `no_bandwidth`, `not_used`, `tolerable_risk`) | No       | `not_used`                                            |

### How It Works

1. The action fetches all open Dependabot alerts for the repository
2. For each alert, it checks if the manifest path matches any of the provided glob patterns
3. If the path matches a pattern, it dismisses the alert with the specified reason and comment

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

2. The GitHub App needs to be installed on your repository or organization

3. The App ID and private key should be stored securely (e.g., in Vault)

The example workflow above demonstrates using the `actions/create-github-app-token` action to generate a token with the required permissions.

If you're experiencing "Resource not accessible by integration" errors, this indicates that the token being used doesn't have the necessary permissions to access the Dependabot API.

## License

This action is licensed under the same license as the parent repository.
