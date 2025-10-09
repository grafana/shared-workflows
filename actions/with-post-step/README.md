# with-post-step

Action to set a command as a post step

Source Code reference: https://github.com/pyTooling/Actions/tree/main/with-post-step

<!-- x-release-please-start-version -->

```yaml
name: CI
on:
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read

    steps:
      - name: Checkout
        uses: actions/checkout@08c6903cd8c0fde910a37f88322edcfb5dd907a8 # v5.0.0
        with:
          persist-credentials: false

      - name: Run Create GitHub App Token action
        id: command
        run: |
          echo "running command"
          echo "test_output=test_output_content" >> "${GITHUB_OUTPUT}"

      - name: Skip invalid instance
        uses: ./actions/with-post-step
        with:
          main: echo "with-post-step run"
          post: echo ${{ steps.command.outputs.test_output }}
```

<!-- x-release-please-end-version -->
