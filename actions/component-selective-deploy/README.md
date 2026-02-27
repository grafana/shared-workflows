# component-selective-deploy

Aggregates newly built image digests, selects between new and previous digests
based on change detection output, validates all digests, and updates the
`component-tags` artifact that tracks what was last deployed.

Designed to run in the deploy job, immediately after
`component-change-detection` determines what changed and after component digest
artifacts from the build job have been downloaded.

<!-- x-release-please-start-version -->

```yaml
- name: Detect changed components
  id: detect
  uses: grafana/shared-workflows/actions/component-change-detection@component-change-detection/v1.0.0
  with:
    config-file: ".component-deps.yaml"
    previous-tags-source: "deploy-prod.yml"

- name: Selective deploy
  id: deploy
  uses: grafana/shared-workflows/actions/component-selective-deploy@component-selective-deploy/v1.0.0
  with:
    components-json: ${{ steps.detect.outputs.components_json }}
    changes-json: ${{ steps.detect.outputs.changes_json }}
    commit-sha: ${{ github.sha }}

- name: Trigger deployment
  uses: grafana/shared-workflows/actions/trigger-argo-workflow@...
  with:
    parameters: |
      apiserver_digest=${{ fromJson(steps.deploy.outputs.selected-digests-json).apiserver_digest }}
      controller_digest=${{ fromJson(steps.deploy.outputs.selected-digests-json).controller_digest }}
```

<!-- x-release-please-end-version -->

## Inputs

| Name                    | Type   | Description                                                                                                   | Default               |
| ----------------------- | ------ | ------------------------------------------------------------------------------------------------------------- | --------------------- |
| `components-json`       | String | JSON array of component names, from `component-change-detection` `outputs.components_json`                    |                       |
| `changes-json`          | String | JSON object of component → changed (`true`/`false`), from `component-change-detection` `outputs.changes_json` |                       |
| `commit-sha`            | String | Git commit SHA of this deployment                                                                             |                       |
| `component-digests-dir` | String | Directory containing per-component digest `.txt` files from `save-component-digest`                           | `component-digests`   |
| `component-tags-file`   | String | Path to `component-tags.json` containing the previous deployment state                                        | `component-tags.json` |

## Outputs

| Name                    | Type   | Description                                                             |
| ----------------------- | ------ | ----------------------------------------------------------------------- |
| `selected-digests-json` | String | JSON object of selected component digests keyed as `<component>_digest` |
| `dockertag`             | String | Docker tag (short SHA) used for this build                              |

### Using `selected-digests-json`

Use `fromJson()` to extract individual component digests in downstream steps:

```yaml
grafana_com_api_digest=${{ fromJson(steps.deploy.outputs.selected-digests-json).grafana_com_api_digest }}
```

## Files Written to the Workspace

| File                    | Description                                                         |
| ----------------------- | ------------------------------------------------------------------- |
| `new-digests.json`      | All newly built digests aggregated from `component-digests/`        |
| `selected-digests.json` | Final selected digests (new for changed, old for unchanged)         |
| `component-tags.json`   | Updated with the new deployment state (ready to upload as artifact) |

After running this action, upload `component-tags.json` as an artifact so future
deployments can compare against it:

```yaml
- name: Upload component tags
  uses: actions/upload-artifact@v6
  with:
    name: component-tags
    path: component-tags.json
    retention-days: 90
```

## How It Works

1. **Aggregate** — reads per-component digest files from `component-digests/` (written by `save-component-digest`) and builds `new-digests.json`
2. **Select** — for each component, picks the new digest if changed or the previous digest from `component-tags.json` if unchanged
3. **Validate** — asserts every digest is non-empty and matches the `<tag>@sha256:<64-hex>` format before triggering deployment
4. **Update** — rewrites `component-tags.json` with the new commitSHA and digest for changed components, leaving unchanged entries intact

## Requirements

- `component-digests/` must be populated (via `save-component-digest` + artifact download) before calling this action
- `component-tags.json` must be present (via `download-previous-component-tags` or similar). If missing, all components are treated as first-time deployments
- `jq` — pre-installed on GitHub-hosted runners

## Permissions

This action requires no additional GitHub token permissions beyond what is
needed for the surrounding workflow.

## Relation to Other Actions

```
save-component-digest      → called once per component in the build job
                             writes component-digests/<name>.txt

component-change-detection → called in the deploy job
                             outputs changes_json and components_json

component-selective-deploy → called in the deploy job after detection
                             reads component-digests/ and component-tags.json
                             outputs selected-digests-json and dockertag
```

## Troubleshooting

| Issue                               | Cause                                                                                  | Solution                                                                                                     |
| ----------------------------------- | -------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| "Missing digest file for component" | `save-component-digest` didn't run for that component, or artifacts weren't downloaded | Ensure every component in `components-json` has a corresponding artifact uploaded and downloaded             |
| "Digest validation failed"          | Digest format is incorrect                                                             | Check that `new-digest` passed to `save-component-digest` is the raw `sha256:...` value from the Docker push |
| All components use new digests      | No `component-tags.json` found                                                         | Expected on first deployment; upload the file after each run so future deploys can use it                    |
| `component-tags.json` not updating  | File not uploaded as artifact after deployment                                         | Add an upload step after the action runs                                                                     |
