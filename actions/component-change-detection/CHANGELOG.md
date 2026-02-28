# Changelog

## [1.0.1](https://github.com/grafana/shared-workflows/compare/component-change-detection/v1.0.0...component-change-detection/v1.0.1) (2026-02-28)


### 🔧 Miscellaneous Chores

* **deps:** update dawidd6/action-download-artifact action to v15 ([#1738](https://github.com/grafana/shared-workflows/issues/1738)) ([36005ff](https://github.com/grafana/shared-workflows/commit/36005ff276117278dcb4498b82975048530ef069))

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
