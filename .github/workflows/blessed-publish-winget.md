# blessed-publish-winget

This is the blessed reusable workflow for the `publish_winget` CI gates
operation. It submits a new version of a Grafana package to
[WinGet](https://learn.microsoft.com/windows/package-manager/winget/) by opening
a pull request on [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs)
with Microsoft's `wingetcreate` tool.

The WinGet token is a CI gates operation secret. Only repositories that have
adopted `publish_winget` in CI gates can read it, from an allowed protected
ref, and only inside this workflow's `publish` job. The calling repository
needs its `ci-gates/config.yaml` and the `publish_winget` GitHub Environment
before it can use this workflow.

```yaml
name: Submit WinGet manifest

on:
  release:
    types: [released]

jobs:
  submit-winget-manifest:
    permissions:
      contents: read
      id-token: write
    uses: grafana/shared-workflows/.github/workflows/blessed-publish-winget.yml@main
    with:
      release-tag: ${{ github.event.release.tag_name }}
      package-id: GrafanaLabs.Alloy
      urls: https://github.com/grafana/alloy/releases/download/${{ github.event.release.tag_name }}/alloy-installer-windows-amd64.exe
```

## Inputs

<!-- BEGIN_INPUTS -->

| Name          | Type   | Required | Default | Description                              |
| ------------- | ------ | -------- | ------- | ---------------------------------------- |
| `package-id`  | string | Yes      |         | WinGet package ID (e.g., GrafanaLabs.k6) |
| `release-tag` | string | Yes      |         | Release tag (e.g., v1.0.0)               |
| `urls`        | string | Yes      |         | Space-separated installer URLs           |

<!-- END_INPUTS -->
