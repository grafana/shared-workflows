# Dependabot GAR Login

Updates the token for Google Artifact Registry and sets it as a secret for Dependabot at the repository level.

The token can be refreshed and set to the `DEPENDABOT_GAR_TOKEN` secret every 50 minutes to ensure Dependabot can access Google Artifact Registry without interruptions.

## Example

```yaml
name: Update Google Artifact Registry Token

on:
  schedule:
    # Update every 50 minutes
    - cron: "*/50 * * * *"
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    steps:
      - name: Update Artifact Registry Token
        uses: ./.github/actions/update-artifact-registry-token
```

## Configuring Dependabot

To use the secret in your dependabot.yml, add the following configuration:

```yaml
registries:
  google-artifact-registry:
    type: "docker-registry" # Replace with your registry type
    url: "us-docker.pkg.dev"
    token: "${{ secrets.DEPENDABOT_GAR_TOKEN }}"
```
