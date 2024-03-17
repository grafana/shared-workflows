# bump-and-release-cloud-plugin

This composite action will help automate the release process of a cloud plugin.

Through it the following steps will be performed:

1. Bump the version of the plugin in the `package.json` file through the use of either `yarn version <level>`
   or `npm version <level>`
2. Optionally update the changelog file by inserting the new version after the `## Unreleased` header
3. Lint, test, and build the plugin
4. Package, sign, and validate the plugin
5. If the plugin is valid, it will commit the version bump, create a tag and release, and push the artifacts to GCS

## Inputs

| Name                        | Description                                                                                                      | Required | Default      |
|-----------------------------|------------------------------------------------------------------------------------------------------------------|----------|--------------|
| `default-branch`            | The default branch of the repository                                                                             | `false`  | `main`       |
| `gcs-bucket`                | The GCS bucket to upload the artifacts to                                                                        | `true`   |              |
| `gcs-service-account-creds` | The GCS service account credentials json to log into GCP                                                         | `true`   |              |
| `github-token`              | The GitHub token to use for the action, requires the permissions to commit to main and create tags and releases. | `true`   |              |
| `package-manager`           | The package manager used for the plugin                                                                          | `false`  | `yarn`       |
| `release-level`             | The level of the release                                                                                         | `false`  | `prerelease` |
| `signing-token`             | The access policy token to use for signing the plugin                                                            | `true`   |              |
| `update-changelog`          | Whether to update the changelog file                                                                             | `false`  | `false`      |

## Note on the artifacts

The zip file name with be `{{ plugin-id }}-{{ version }}.zip` E.g. `grafana-example-app-1.0.0.zip`

There will also be a `*.zip.md5` checksum calculated and uploaded to the same GCS bucket.

If the release level is anything other than `prerelease` an additional zip file will be uploaded as `*-latest.zip` (and checksum)

## Example workflow

```yaml
name: Release

on:
  workflow_dispatch:
    inputs:
      level:
        description: 'Release level'
        required: true
        default: 'prerelease'
        type: choice
        options:
          - major
          - minor
          - patch
          - prerelease

jobs:
  bump-and-release:
    runs-on: ubuntu-latest
    steps:
      # ... Steps to get necessary secrets
      - uses: grafana/shared-workflows/actions/bump-and-release-cloud-plugin@main
        with:
          github-token: ${{ env.mygithub-token }} # Needs the ability to commit to main and create tags and releases
          gcs-bucket: ${{ env.GCS_BUCKET }}
          gcs-service-account-creds: ${{ env.GCS_SERVICE_ACCOUNT_CREDS }}
          release-level: 'minor' # 'major', 'minor', 'patch', 'prerelease'
          signing-token: ${{ env.SIGNING_TOKEN }}
```

