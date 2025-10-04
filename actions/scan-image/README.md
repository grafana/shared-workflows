# scan-image

This is a composite GitHub Action used to scan your images in search of vulnerabilities.

The goal is to provide developers a way to check if their PR changes, without having to
wait for periodic scans (Faster feedback loop). This can also be used as part of deployment
CI/CD jobs as a way to verify things before shipping to production environments.

<!-- x-release-please-start-version -->

```yaml
name: Scan image for vulnerabilities
jobs:
  scan-image:
    name: Scan image for vulnerabilities
    steps:
      - name: Scan image for vulnerabilities
        id: scan-image
        uses: grafana/shared-workflows/actions/scan-image@scan-image/v0.1.0
        with:
          image_name: docker.io/hello-world
          fail_on: critical
          fail_on_threshold: 1
```

<!-- x-release-please-end-version -->
