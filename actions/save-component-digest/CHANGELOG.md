# Changelog

## 1.0.0 (2026-02-20)

### 🎉 Features

* Initial release of save-component-digest action
* Saves newly built component image digests to disk for use by component-selective-deploy
* Automatically strips `_digest` suffix from component name when deriving filename
* Configurable output directory via `component-digests-dir` input
