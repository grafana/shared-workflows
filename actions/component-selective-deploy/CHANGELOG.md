# Changelog

## 1.0.0 (2026-02-20)

### 🎉 Features

* Initial release of component-selective-deploy action
* Aggregates newly built component digests from per-component artifact files
* Selects new or previous digest per component based on change detection output
* Validates all digests against `<tag>@sha256:<64-hex>` format before deployment
* Updates `component-tags.json` artifact to track deployed state for future runs
* Configurable digest directory and component tags file paths
* Designed to pair with `component-change-detection` and `save-component-digest`
