# Changelog

## [1.0.1](https://github.com/grafana/shared-workflows/compare/component-change-detection/v1.0.0...component-change-detection/v1.0.1) (2026-03-20)


### 🔧 Miscellaneous Chores

* **deps:** update actions/setup-go action to v6.3.0 ([#1756](https://github.com/grafana/shared-workflows/issues/1756)) ([c6b0752](https://github.com/grafana/shared-workflows/commit/c6b07529443393154d824d1ad0e707f4b3d090f6))
* **deps:** update dawidd6/action-download-artifact action to v15 ([#1738](https://github.com/grafana/shared-workflows/issues/1738)) ([36005ff](https://github.com/grafana/shared-workflows/commit/36005ff276117278dcb4498b82975048530ef069))
* **deps:** update dawidd6/action-download-artifact action to v16 ([#1753](https://github.com/grafana/shared-workflows/issues/1753)) ([b71be18](https://github.com/grafana/shared-workflows/commit/b71be180de45d83fe3f1641ff8ced2b2d967c155))
* **deps:** update dawidd6/action-download-artifact action to v18 ([#1812](https://github.com/grafana/shared-workflows/issues/1812)) ([e40c7fe](https://github.com/grafana/shared-workflows/commit/e40c7fecb195dd2644cd3991abf756782cbd4143))

## 1.0.0 (2026-02-19)


### 🎉 Features

* add component-change-detection action ([#1711](https://github.com/grafana/shared-workflows/issues/1711)) ([ef4f6da](https://github.com/grafana/shared-workflows/commit/ef4f6dac37ac4e312a6282e112a0726888bee36e))

## 1.0.0 (2026-02-06)


### 🎉 Features

* Initial release of component-change-detection action
* Go-based change detection tool for analyzing git history
* Support for glob patterns in path matching
* Dependency graph with cycle detection
* Automatic propagation of changes through dependencies
* Force rebuild options (all or specific components)
* Comprehensive documentation and examples
* Integration with GitHub Actions artifacts for state tracking
