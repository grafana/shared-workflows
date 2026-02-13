# Component Change Detection Tool

A reusable Go-based CLI tool that determines which Docker images need rebuilding by analyzing git
history, dependency relationships, and file changes.

## Motivation

In a multi-component repository, we need to efficiently determine which components have changed and
need rebuilding. By excluding non-functional changes (like documentation and tests), we maximize
CI/CD efficiency. This is critical for:

- **CI/CD Optimization**: Avoid rebuilding unchanged components
- **Build Time Reduction**: Skip unnecessary Docker builds in large repositories
- **Transitive Dependencies**: Automatically detect when shared code changes affect multiple
  components
- **Deployment Efficiency**: Only deploy components that have actually changed

**Designed for CI/CD**: This tool is optimized for GitHub Actions and other CI/CD environments where
git is always available. It uses the system git command for maximum performance and compatibility.

## How It Works

1. Load component configuration from YAML
2. Compare files changed between git tags and target ref
3. Match changed files against component path patterns (with glob support)
4. Apply exclusion patterns to filter out test files, docs, etc.
5. Build dependency graph and detect circular dependencies
6. Propagate changes through transitive dependencies
7. Output JSON map of components → changed (true/false)

## Quick Start

```bash
# 1. Build the tool
make build

# 2. Create configuration file (.component-deps.yaml)
# 3. Create tags file (current-tags.json)

# 4. Detect changes
./bin/changed-components \
  --config .component-deps.yaml \
  --tags current-tags.json \
  --verbose
```

## Configuration

Create a `.component-deps.yaml` file defining your components:

```yaml
# Global exclusions applied to all components
global_excludes:
  - "**/*_test.go"
  - "docs/**"
  - "**/*.md"

components:
  migrator:
    paths:
      - "migrations/**"
    excludes: []
    dependencies: []

  apiserver:
    paths:
      - "pkg/**"
      - "cmd/apiserver/**"
      - "go.mod"
      - "go.sum"
    excludes: [] # Uses global_excludes
    dependencies:
      - migrator # If migrator changes, apiserver must rebuild
```

### Path Patterns

- **Glob support**: Use `**` for recursive matching (e.g., `pkg/**/*.go`)
- **Specific files**: `go.mod`, `Dockerfile`, `Makefile`
- **Exclusions**: Per-component or global patterns to filter test files, docs, etc.

### Dependency Propagation

When a component changes, **all components that depend on it** are automatically marked as changed.

**Example:**

```
migrations/ changes
  ↓
migrator → changed (direct file match)
  ↓
apiserver → changed (depends on migrator)
  ↓
controller → changed (depends on migrator)

templatewatcher → unchanged (no dependency)
```

This ensures that downstream components are rebuilt when their dependencies change, even if their
own files haven't changed.

## Architecture

### Package Structure

```
pkg/changedetector/
  ├── types.go       # Core types (Config, ComponentConfig, Tags)
  ├── config.go      # YAML config loading and validation
  ├── git.go         # Git operations (diff, ref checks)
  ├── matcher.go     # Glob pattern matching with exclusions
  ├── graph.go       # Dependency graph and cycle detection
  └── detector.go    # Main detection orchestrator
```

### Key Features

- ✅ **Glob patterns** with `**` support via `doublestar` library
- ✅ **Cycle detection** prevents circular dependencies
- ✅ **Graceful defaults** - missing tags → mark as changed
- ✅ **Git integration** - null-terminated output for filenames with spaces
- ✅ **Validation** - validates config on load with clear error messages

## Usage

### Output to File

```bash
# Save results to a file
./bin/changed-components \
  --config .component-deps.yaml \
  --tags current-tags.json \
  --output changes.json
```

### Compare Against Specific Commit

```bash
# Compare against a specific git ref
./bin/changed-components \
  --config .component-deps.yaml \
  --tags current-tags.json \
  --target abc123def
```

## Input Format

### Tags File (JSON)

The tags file should be a JSON object mapping component names to image dockertags:

```json
{
  "migrator": "b893b86",
  "apiserver": "6f2c1a9",
  "controller": "6f2c1a9",
  "templatewatcher": "b893b86"
}
```

**Special values:**

- `"none"` - Component has never been built (will be marked as changed)
- Empty string `""` - Same as "none"
- Missing key - Component will be marked as changed

## Output Format

### Default Output

The tool outputs JSON indicating which components changed:

```json
{
  "migrator": true,
  "apiserver": false,
  "controller": false,
  "templatewatcher": false
}
```

- `true` = Component needs to be rebuilt
- `false` = Component is unchanged

### Verbose Output (--verbose flag)

With the `--verbose` flag, the tool outputs detailed information showing which files and commits
triggered changes:

```json
{
  "migrator": [{ "file": "migrations/001_add_table.sql", "commit": "abc123" }],
  "apiserver": [],
  "controller": [],
  "templatewatcher": []
}
```

This shows the specific files that matched each component's path patterns along with the commits
they were changed in. Useful for understanding exactly what triggered a rebuild.

## Integration with CI/CD

### GitHub Actions

This tool is designed to run natively in GitHub Actions. Git is pre-installed on all GitHub Actions
runners, so no additional setup is required.

```yaml
- name: Checkout with full history
  uses: actions/checkout@v3
  with:
    fetch-depth: 0 # Required for git diff to work

- name: Build change detector
  run: make build

- name: Detect changes
  id: detect
  run: |
    ./bin/changed-components \
      --config .component-deps.yaml \
      --tags current-tags.json \
      --output changes.json
    echo "changes=$(cat changes.json)" >> $GITHUB_OUTPUT

- name: Build apiserver
  if: fromJSON(steps.detect.outputs.changes).apiserver == true
  run: make apiserver.tag
```

**Important**: Use `fetch-depth: 0` in your checkout action to ensure full git history is available
for comparison.

### Makefile Integration

The tool is already integrated into the Makefile:

```bash
# Run change detection
make detect-changes

# This will use:
# - Config: .component-deps.yaml
# - Tags: current-tags.json
# - Output: changes.json (boolean map) and changes-detailed.json (files/commits per component)
```

## Troubleshooting

### "Tag not found" Error

If you see an error like `tag "v1.0.0" does not exist in git repository`:

1. Verify the tag exists: `git tag -l`
2. Check the exact tag name in your tags JSON file
3. Ensure you've fetched all tags: `git fetch --tags`

### No Changes Detected When Expected

Run with `--verbose` to get detailed JSON output showing which files and commits triggered changes
for each component.

**Normal output (boolean map):**

```json
{
  "grafana_com_api": true,
  "dashboards_api": false,
  "marketplaces_api": true,
  "plugins_api": false
}
```

**Verbose output (files and commits per component):**

```json
{
  "grafana_com_api": [
    {
      "file": "packages/grafana-com-base/api/routes/users.ts",
      "commit": "def456"
    },
    { "file": "packages/grafana-com-base/package.json", "commit": "def456" }
  ],
  "dashboards_api": [],
  "marketplaces_api": [
    {
      "file": "packages/grafana-com-marketplaces-api/src/index.ts",
      "commit": "abc123"
    }
  ],
  "plugins_api": []
}
```

This makes it easy to understand exactly which commits and files triggered a rebuild, useful for
debugging and retroactive analysis when multiple commits are batched together.

### All Components Marked as Changed

This happens when:

- Tags are set to "none" or empty
- Tags point to invalid git refs
- The target comparison is the same as the tags (no diff)
