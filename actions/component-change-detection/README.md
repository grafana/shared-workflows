# Component Change Detection Action

A reusable GitHub Action that determines which components need rebuilding by analyzing git history, dependency relationships, and file changes.

## Overview

This action enables **selective deployment** by comparing the current commit against a previous deployment to determine which components need to be rebuilt/redeployed.

**Benefits:**
- ✅ Skip deployments when only docs/tests changed
- ✅ Reuse unchanged component images (faster deployments)
- ✅ Reduce cloud costs (fewer builds)
- ✅ Automatic transitive dependency handling
- ✅ Configurable via YAML (no code changes needed)

## Usage

<!-- x-release-please-start-version -->

```yaml
- name: Detect changed components
  id: detect
  uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: '.component-deps.yaml'
    previous-tags-source: 'deploy-prod.yml'

- name: Build component if changed
  if: fromJSON(steps.detect.outputs.changes_json).apiserver == true
  run: make build-apiserver
```

<!-- x-release-please-end-version -->

## Quick Start

### 1. Create Configuration File

Create a `.component-deps.yaml` file in your repository root:

```yaml
# Global exclusions applied to all components
global_excludes:
  - '**/*.test.ts'
  - 'docs/**'
  - '**/*.md'

components:
  migrator:
    paths:
      - 'database/migrations/**'
    excludes: []
    dependencies: []

  apiserver:
    paths:
      - '**'  # Watch all code
    excludes:
      - 'database/migrations/**'
    dependencies:
      - migrator  # If migrator changes, apiserver rebuilds too
```

### 2. Use the Action in Your Workflow

```yaml
- name: Checkout with history
  uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
  with:
    fetch-depth: 100  # Fetch recent history for comparison

- name: Detect changed components
  id: detect
  uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: '.component-deps.yaml'
    previous-tags-source: 'deploy-prod.yml'  # Workflow that uploads component-tags

- name: Build apiserver (if changed)
  if: fromJSON(steps.detect.outputs.changes_json).apiserver == true
  run: make build-apiserver

- name: Build controller (if changed)
  if: fromJSON(steps.detect.outputs.changes_json).controller == true
  run: make build-controller
```

### 3. Upload Component Tags After Deployment

Your deployment workflow needs to upload the `component-tags` artifact:

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

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `config-file` | Path to component dependencies YAML file | Yes | - |
| `previous-tags-source` | Workflow name to get previous tags from | Yes | - |
| `target-ref` | Git ref to compare against | No | `HEAD` |
| `force-rebuild-all` | Force rebuild all components | No | `false` |
| `force-components` | Force rebuild specific components (comma-separated) | No | `''` |

## Outputs

| Output | Description | Example |
|--------|-------------|---------|
| `changes_json` | All component changes as JSON object | `{"apiserver": true, "controller": false}` |
| `components_json` | List of all components as JSON array | `["apiserver", "controller"]` |

## Configuration File Format

### Basic Structure

```yaml
global_excludes:
  - 'pattern1'
  - 'pattern2'

components:
  component_name:
    paths:
      - 'path/to/watch/**'
    excludes:
      - 'path/to/ignore/**'
    dependencies:
      - 'other_component'
```

### Path Patterns

Supports **glob patterns** with `**` for recursive matching:

```yaml
paths:
  - '**/*.go'          # All Go files
  - 'cmd/apiserver/**' # Everything under cmd/apiserver
  - 'go.mod'           # Specific files
  - '**'               # Watch everything (use with excludes)
```

### Exclusions

**Global exclusions** apply to all components:

```yaml
global_excludes:
  - '**/*.test.go'
  - 'docs/**'
  - '**/*.md'
  - 'scripts/**'
```

**Component-specific exclusions** apply only to that component:

```yaml
components:
  apiserver:
    paths:
      - '**'
    excludes:
      - 'database/migrations/**'  # Don't trigger on migrations
```

### Dependencies

When a component changes, **all components that depend on it** are automatically marked as changed:

```yaml
components:
  shared_lib:
    paths:
      - 'lib/**'
    dependencies: []

  apiserver:
    paths:
      - 'cmd/apiserver/**'
    dependencies:
      - shared_lib  # Rebuilds when shared_lib changes
```

**Dependency chain example:**

```
lib/utils.go changes
  ↓
shared_lib → changed (direct file match)
  ↓
apiserver → changed (depends on shared_lib)
  ↓
admin_ui → changed (depends on apiserver)
```

## How It Works

1. **Download Previous Tags**: Gets `component-tags.json` from last successful deployment
2. **Extract Components**: Reads component list from `.component-deps.yaml`
3. **Build Change Detector**: Compiles the Go-based change detection tool
4. **Detect Changes**: 
   - Compares each component's paths against git diff
   - Applies exclusion patterns
   - Builds dependency graph
   - Propagates changes through dependencies
5. **Output Results**: Sets GitHub Actions outputs for each component

## Advanced Usage

### Force Rebuild Specific Components

```yaml
- name: Detect changes
  uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: '.component-deps.yaml'
    previous-tags-source: 'deploy-prod.yml'
    force-components: 'apiserver,controller'  # Force these to rebuild
```

### Force Rebuild All

```yaml
- name: Detect changes
  uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: '.component-deps.yaml'
    previous-tags-source: 'deploy-prod.yml'
    force-rebuild-all: 'true'
```

### Compare Against Specific Ref

```yaml
- name: Detect changes since v1.0.0
  uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: '.component-deps.yaml'
    previous-tags-source: 'deploy-prod.yml'
    target-ref: 'v1.0.0'
```

## Requirements

### Dependencies

- **Git history**: Checkout with `fetch-depth: 100` (or more) to ensure commits are available
- **Deployment workflow**: Must upload `component-tags` artifact
- **yq**: Pre-installed on GitHub-hosted runners
- **jq**: Pre-installed on GitHub-hosted runners
- **Go**: Automatically installed by the action via `actions/setup-go`

### Permissions

This action requires the following GitHub token permissions:

```yaml
permissions:
  actions: read  # Required to download artifacts from previous workflow runs
  contents: read # Required to checkout code and read git history
```

**Minimal Example:**

```yaml
jobs:
  detect-changes:
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
        with:
          fetch-depth: 100
      
      - uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
        with:
          config-file: '.component-deps.yaml'
          previous-tags-source: 'deploy-prod.yml'
```

## Troubleshooting

### "No previous build state found"

**Cause**: First deployment or artifact expired.
**Solution**: All components will be marked as changed (safe default).

### Components always marked as changed

**Cause**: Previous tags not being saved/uploaded correctly.
**Solution**: Ensure deployment workflow uploads `component-tags` artifact with correct format.

### "Git operation failed"

**Cause**: Not enough git history fetched.
**Solution**: Increase `fetch-depth` in checkout step (e.g., `fetch-depth: 200`).

### Circular dependency detected

**Cause**: Component A depends on B, and B depends on A.
**Solution**: Remove circular dependency in `.component-deps.yaml`.

## Examples

See the [grafana-com repository](https://github.com/grafana/grafana-com) for a complete working example:

- [.component-deps.yaml](https://github.com/grafana/grafana-com/blob/main/.component-deps.yaml)
- [build.yml workflow](https://github.com/grafana/grafana-com/blob/main/.github/workflows/build.yml)
- [deploy-prod.yml workflow](https://github.com/grafana/grafana-com/blob/main/.github/workflows/deploy-prod.yml)

## License

Same as parent repository.

## Support

For issues or questions, please file an issue in the [shared-workflows repository](https://github.com/grafana/shared-workflows/issues).
