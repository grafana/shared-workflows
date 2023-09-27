# get-vault-secrets

From a `grafana/` org repository, get a secret from the Grafana vault instance.

Example workflow:

```yaml
name: CI
on: 
  pull_request:

# These permissions are needed to assume roles from Github's OIDC.
permissions:
  contents: read
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - id: get-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@main
        with:
          workload_identity_provider: '${{ secrets.WORKLOAD_IDENTITY_POOL_PROVIDER }}'
          iap_service_account: '${{ secrets.VAULT_IAP_SA_EMAIL }}'
          iap_audience: '${{ secrets.VAULT_IAP_OAUTH_CLIENT_ID }}'
          vault_url: '${{ secrets.VAULT_URL }}'
          secrets: |
              ci/data/repo/grafana/<repo>/test-secret my-key | TEST_KEY;
              ci/data/common/test-secret my-key | TEST_KEY_2;

    # Use the secrets
    # You can use the envvars directly in scripts or use the `${{ env.* }}` accessor in the workflow
      - name: echo
        run: |
          echo "$TEST_KEY"
          echo "${{ env.TEST_KEY_2 }}"

```

<details>
<summary>Implementation Details</summary>

- This is an action, and not a shared workflow, because when secrets are read, they need to be shared as env variables, or at least in any way that is strictly in-memory. Shared workflows cannot be called as steps and workflows can only share data through external storage (caches, buckets, etc).
- Secrets need to be injected because actions don't have access to the `secrets` item.
</details>
