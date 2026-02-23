# save-component-digest

Saves a newly built component image digest to the `component-digests/` directory
immediately after a Docker push. Used in the build job to record what was built,
so the `component-selective-deploy` action can later decide which image to deploy
for each component.

This action is designed to be called once per component in a matrix build job.

<!-- x-release-please-start-version -->

```yaml
- name: Save component digest
  uses: grafana/shared-workflows/actions/save-component-digest@save-component-digest/v1.0.0
  with:
    component: ${{ matrix.argowf_parameter_name }}
    new-dockertag: ${{ env.TAG }}
    new-digest: ${{ steps.push.outputs.digest }}
```

<!-- x-release-please-end-version -->

Then upload the directory as an artifact so the deploy job can access it:

```yaml
- name: Upload component digest
  uses: actions/upload-artifact@v6
  with:
    name: component-digest-${{ matrix.argowf_parameter_name }}
    path: component-digests/
    retention-days: 90
```

## Inputs

| Name                    | Type   | Description                                                                               | Default             |
| ----------------------- | ------ | ----------------------------------------------------------------------------------------- | ------------------- |
| `component`             | String | Component identifier, typically the argowf parameter name (e.g. `grafana_com_api_digest`) |                     |
| `new-dockertag`         | String | Short git SHA used as the Docker tag                                                      |                     |
| `new-digest`            | String | Full image digest from the Docker push (e.g. `sha256:abc123...`)                          |                     |
| `component-digests-dir` | String | Directory to write digest files into                                                      | `component-digests` |

## Outputs

This action produces no step outputs. It writes two files to disk:

| File                                           | Contents                                  |
| ---------------------------------------------- | ----------------------------------------- |
| `<component-digests-dir>/<component_name>.txt` | Full digest string `<dockertag>@<digest>` |
| `<component-digests-dir>/dockertag.txt`        | The Docker tag (short SHA) for this build |

The `_digest` suffix is automatically stripped from the `component` input when forming the filename, so `grafana_com_api_digest` becomes `grafana_com_api.txt`.

## How It Works

1. Concatenates `new-dockertag` and `new-digest` into a full digest string (`tag@sha256:...`)
2. Strips the `_digest` suffix from `component` to derive the filename
3. Writes the digest to `<component-digests-dir>/<component_name>.txt`
4. Writes the docker tag to `<component-digests-dir>/dockertag.txt`

## Requirements

- The directory specified by `component-digests-dir` is created automatically if it does not exist.

## Permissions

This action requires no additional GitHub token permissions.

## Working with `component-selective-deploy`

This action is the build-time counterpart to `component-selective-deploy`. The
typical full flow across two jobs is:

```yaml
jobs:
  build:
    steps:
      - name: Push image
        id: push
        uses: grafana/shared-workflows/actions/push-to-gar-docker@...

      - name: Save digest
        uses: grafana/shared-workflows/actions/save-component-digest@save-component-digest/v1.0.0
        with:
          component: ${{ matrix.argowf_parameter_name }}
          new-dockertag: ${{ env.TAG }}
          new-digest: ${{ steps.push.outputs.digest }}

      - name: Upload digest artifact
        uses: actions/upload-artifact@v6
        with:
          name: component-digest-${{ matrix.argowf_parameter_name }}
          path: component-digests/

  deploy:
    needs: build
    steps:
      - name: Download all component digests
        uses: actions/download-artifact@v7
        with:
          pattern: component-digest-*
          path: component-digests/
          merge-multiple: true

      - name: Detect changed components
        id: detect
        uses: grafana/shared-workflows/actions/component-change-detection@...

      - name: Selective deploy
        uses: grafana/shared-workflows/actions/component-selective-deploy@...
        with:
          components-json: ${{ steps.detect.outputs.components_json }}
          changes-json: ${{ steps.detect.outputs.changes_json }}
          commit-sha: ${{ github.sha }}
```
