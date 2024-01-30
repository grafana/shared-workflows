# push-to-gar-docker

This is a composite GitHub Action, used to push docker objects to a bucket in Google Cloud Storage (GCS).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can 
trigger these composite workflows.

```yaml
name: CI
on: 
  pull_request:
    
env:
  ENVIRONMENT: "dev" # can be either dev/prod

# These permissions are needed to assume roles from Github's OIDC.
permissions:
  contents: read
  id-token: write

jobs:
  upload-object:
    runs-on: ubuntu-latest

    steps:
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        id: push-to-gcs
        with:
          object: path/to/object
          destination: folder/object  # Optional: defaults to the object input
          environment: ${{ env.ENVIRONMENT }}
```
