# run-shellcheck

Finds and lints shell scripts using [ShellCheck](https://www.shellcheck.net/). A lightweight replacement for `ludeeus/action-shellcheck` that uses the runner's pre-installed `shellcheck` instead of downloading a binary.

The action:

1. Finds shell scripts by known extensions (`.sh`, `.bash`, `.ksh`, `.zsh`, `.shlib`, plus dotfiles like `.bashrc`, `.zshrc`, `profile`, etc.)
2. Finds extensionless executables with shell shebangs
3. Runs `shellcheck` on each file (or all at once with `check-together`)
4. Exits with non-zero if any file has issues

## Inputs

| Name               | Type     | Description                                                      | Default | Required |
| ------------------ | -------- | ---------------------------------------------------------------- | ------- | -------- |
| `scandir`          | `string` | Directory to scan for shell scripts                              | `.`     | false    |
| `ignore-paths`     | `string` | Space-separated paths or directories to exclude                  |         | false    |
| `ignore-names`     | `string` | Space-separated file names to exclude                            |         | false    |
| `severity`         | `string` | Minimum severity: `error`, `warning`, `info`, `style`            |         | false    |
| `format`           | `string` | Output format: `gcc`, `tty`, `json`, `json1`, `checkstyle`, `diff`, `quiet` | `gcc` | false |
| `additional-files` | `string` | Space-separated extra file names to scan for                     |         | false    |
| `check-together`   | `string` | Run shellcheck on all files at once (`true`/`false`)             | `false` | false    |

## Outputs

| Name    | Description                                      |
| ------- | ------------------------------------------------ |
| `files` | Space-separated list of files that were checked   |

## ShellCheck options

You can pass any supported ShellCheck option via the `SHELLCHECK_OPTS` environment variable:

```yaml
- uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
  env:
    SHELLCHECK_OPTS: -e SC2059 -e SC2034 -e SC1090
```

## Examples

### Basic usage

<!-- x-release-please-start-version -->

```yaml
name: Lint
on: [push, pull_request]

jobs:
  shellcheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
```

<!-- x-release-please-end-version -->

### Ignore paths

```yaml
- uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
  with:
    ignore-paths: "vendor node_modules third_party"
```

### Ignore specific file names

```yaml
- uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
  with:
    ignore-names: "generated.sh legacy-script.sh"
```

### Only fail on errors (ignore warnings, info, style)

```yaml
- uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
  with:
    severity: error
```

### Scan a specific directory

```yaml
- uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
  with:
    scandir: ./scripts
```

### Scan additional file names

```yaml
- uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
  with:
    additional-files: "run finish"
```

### Check all files at once

Useful for resolving SC1090/SC1091 (can't follow sourced files):

```yaml
- uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
  with:
    check-together: "true"
```

### Multi-line TTY output

```yaml
- uses: grafana/shared-workflows/actions/run-shellcheck@run-shellcheck/v0.1.0
  with:
    format: tty
```

## Migration from ludeeus/action-shellcheck

| ludeeus input        | This action's input  | Notes                                           |
| -------------------- | -------------------- | ----------------------------------------------- |
| `scandir`            | `scandir`            | Same behavior                                   |
| `ignore_paths`       | `ignore-paths`       | Same behavior (uses kebab-case)                 |
| `ignore_names`       | `ignore-names`       | Same behavior (uses kebab-case)                 |
| `severity`           | `severity`           | Same behavior                                   |
| `format`             | `format`             | Same behavior, same default (`gcc`)             |
| `additional_files`   | `additional-files`   | Same behavior (uses kebab-case)                 |
| `check_together`     | `check-together`     | Same behavior (uses kebab-case)                 |
| `version`            | *(not supported)*    | Uses runner's pre-installed shellcheck          |
| `SHELLCHECK_OPTS`    | `SHELLCHECK_OPTS`    | Same -- set as env var on the step              |
