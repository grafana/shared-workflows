# component-change-detection

This is a composite GitHub Action that determines which components need rebuilding by analyzing git history, dependency relationships, and file changes. It enables selective deployment by comparing the current commit against a previous deployment.

<!-- x-release-please-start-version -->

```yaml
name: Build with selective deployment
on:
  push:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
    steps:
      - name: Checkout with history
        uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
        with:
          fetch-depth: 100

      - name: Detect changed components
        id: detect
        uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
        with:
          config-file: ".component-deps.yaml"
          previous-tags-source: "deploy-prod.yml"

      - name: Build apiserver (if changed)
        if: fromJSON(steps.detect.outputs.changes_json).apiserver == true
        run: make build-apiserver

      - name: Build controller (if changed)
        if: fromJSON(steps.detect.outputs.changes_json).controller == true
        run: make build-controller
```

<!-- x-release-please-end-version -->

## Inputs

| Name                   | Type    | Description                                                     | Default |
| ---------------------- | ------- | --------------------------------------------------------------- | ------- |
| `config-file`          | String  | Path to component dependencies YAML file                        |         |
| `previous-tags-source` | String  | Workflow name to download previous component-tags artifact from |         |
| `target-ref`           | String  | Git ref to compare against                                      | `HEAD`  |
| `force-rebuild-all`    | Boolean | Force rebuild all components                                    | `false` |
| `force-components`     | String  | Force rebuild specific components (comma-separated)             |         |

## Outputs

| Name              | Type   | Description                                                                             |
| ----------------- | ------ | --------------------------------------------------------------------------------------- |
| `changes_json`    | String | All component changes as JSON object (e.g., `{"apiserver": true, "controller": false}`) |
| `components_json` | String | List of all components as JSON array (e.g., `["apiserver", "controller"]`)              |

## Configuration File

Create a `.component-deps.yaml` file in your repository root:

```yaml
# Global exclusions applied to all components
global_excludes:
  - "**/*.test.ts"
  - "docs/**"
  - "**/*.md"

components:
  migrator:
    paths:
      - "database/migrations/**"
    excludes: []
    dependencies: []

  apiserver:
    paths:
      - "**" # Watch all code
    excludes:
      - "database/migrations/**"
    dependencies:
      - migrator # If migrator changes, apiserver rebuilds too
```

### Path Patterns

Supports glob patterns with `**` for recursive matching:

```yaml
paths:
  - "**/*.go" # All Go files
  - "cmd/apiserver/**" # Everything under cmd/apiserver
  - "go.mod" # Specific files
  - "**" # Watch everything (use with excludes)
```

### Dependencies

When a component changes, all components that depend on it are automatically marked as changed:

```yaml
components:
  shared_lib:
    paths:
      - "lib/**"
    dependencies: []

  apiserver:
    paths:
      - "cmd/apiserver/**"
    dependencies:
      - shared_lib # Rebuilds when shared_lib changes
```

## Component Tags Artifact

Your deployment workflow must upload a `component-tags` artifact after successful deployment:

```yaml
# In your deploy workflow
- name: Upload component tags
  uses: actions/upload-artifact@b7c566a772e6b6bfb58ed0dc250532a479d7789f # v6.0.0
  with:
    name: component-tags
    path: component-tags.json
    retention-days: 90
```

The `component-tags.json` format:

```json
{
  "apiserver": {
    "commitSHA": "abc123def",
    "digest": "abc123d@sha256:..."
  },
  "controller": {
    "commitSHA": "abc123def",
    "digest": "abc123d@sha256:..."
  }
}
```

## How It Works

1. Downloads `component-tags.json` from the last successful deployment via the `previous-tags-source` workflow
2. Extracts component list from `.component-deps.yaml`
3. Compiles the Go-based change detection tool
4. Compares each component's paths against git diff between current commit and previous deployment
5. Applies exclusion patterns (global and component-specific)
6. Builds dependency graph and propagates changes through dependencies
7. Sets GitHub Actions outputs for each component

## Advanced Examples

### Force Rebuild Specific Components

```yaml
- uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: ".component-deps.yaml"
    previous-tags-source: "deploy-prod.yml"
    force-components: "apiserver,controller"
```

### Force Rebuild All

```yaml
- uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: ".component-deps.yaml"
    previous-tags-source: "deploy-prod.yml"
    force-rebuild-all: "true"
```

### Compare Against Specific Ref

```yaml
- uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: ".component-deps.yaml"
    previous-tags-source: "deploy-prod.yml"
    target-ref: "v1.0.0"
```

## Requirements

- **Git history**: Checkout with `fetch-depth: 100` or more to ensure commits are available for comparison
- **Deployment workflow**: Must upload `component-tags` artifact after deployment
- **yq**: Pre-installed on GitHub-hosted runners
- **jq**: Pre-installed on GitHub-hosted runners
- **Go**: Automatically installed by the action via `actions/setup-go`

## Permissions

This action requires the following GitHub token permissions:

```yaml
permissions:
  actions: read # Download artifacts from previous workflow runs
  contents: read # Checkout code and read git history
```

## Troubleshooting

| Issue                               | Cause                                            | Solution                                                                         |
| ----------------------------------- | ------------------------------------------------ | -------------------------------------------------------------------------------- |
| "No previous build state found"     | First deployment or artifact expired             | All components will be marked as changed (safe default)                          |
| Components always marked as changed | Previous tags not being saved/uploaded correctly | Ensure deployment workflow uploads `component-tags` artifact with correct format |
| "Git operation failed"              | Not enough git history fetched                   | Increase `fetch-depth` in checkout step (e.g., `fetch-depth: 200`)               |
| Circular dependency detected        | Component A depends on B, and B depends on A     | Remove circular dependency in `.component-deps.yaml`                             |
