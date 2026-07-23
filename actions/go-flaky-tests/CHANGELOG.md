# Changelog

## [0.1.2](https://github.com/grafana/shared-workflows/compare/go-flaky-tests/v0.1.1...go-flaky-tests/v0.1.2) (2026-07-23)


### 🔧 Miscellaneous Chores

* **deps:** update actions/setup-go action to v6.5.0 ([#2145](https://github.com/grafana/shared-workflows/issues/2145)) ([67f1fea](https://github.com/grafana/shared-workflows/commit/67f1fea99eb1c85cc97bb621d3b3130381f623a9))
* **deps:** update module github.com/rogpeppe/go-internal to v1.15.0 ([#2189](https://github.com/grafana/shared-workflows/issues/2189)) ([dc18119](https://github.com/grafana/shared-workflows/commit/dc1811986e78895b6e2951be3dbbf0a14cf7f2ac))

## [0.1.1](https://github.com/grafana/shared-workflows/compare/go-flaky-tests/v0.1.0...go-flaky-tests/v0.1.1) (2026-06-10)


### 🐛 Bug Fixes

* **create-github-app-token:** trigger release-please for reverted gatb change ([#1988](https://github.com/grafana/shared-workflows/issues/1988)) ([e6c8753](https://github.com/grafana/shared-workflows/commit/e6c875364b041be8288bcb1bee15f79cea31ffb1))

## 0.1.0 (2026-05-04)


### 🎉 Features

* add go-flaky-tests action to release-please config ([#1416](https://github.com/grafana/shared-workflows/issues/1416)) ([eed6978](https://github.com/grafana/shared-workflows/commit/eed69781bdf177e1c126172712690dae79b0bc11))


### 🔧 Miscellaneous Chores

* **deps:** update actions/setup-go action to v6.2.0 ([#1657](https://github.com/grafana/shared-workflows/issues/1657)) ([d29b916](https://github.com/grafana/shared-workflows/commit/d29b9161f1803baed4a7305c85cb5a3018bc3c3e))
* **deps:** update actions/setup-go action to v6.3.0 ([#1756](https://github.com/grafana/shared-workflows/issues/1756)) ([c6b0752](https://github.com/grafana/shared-workflows/commit/c6b07529443393154d824d1ad0e707f4b3d090f6))
* **deps:** update actions/setup-go action to v6.4.0 ([#1837](https://github.com/grafana/shared-workflows/issues/1837)) ([170bd5b](https://github.com/grafana/shared-workflows/commit/170bd5b0ba3c2414519216fd5d7f0fe5a40e3f40))
* **deps:** update actions/setup-go digest to 4dc6199 ([#1551](https://github.com/grafana/shared-workflows/issues/1551)) ([444d030](https://github.com/grafana/shared-workflows/commit/444d030c6286b899fda8821d4108d8b74cc1c3b8))

## [Unreleased]

### Added

- Initial implementation of flaky test analysis action
- Loki integration for fetching test failure logs
- Git history analysis to find test authors
- GitHub issue creation and management for flaky tests
- Dry run mode for testing without creating issues
- Comprehensive test suite with golden file testing

### Features

- **Loki Log Analysis**: Fetches and parses test failure logs using LogQL
- **Flaky Test Detection**: Identifies tests that fail inconsistently across branches
- **Git Author Tracking**: Finds recent commits that modified flaky tests
- **GitHub Integration**: Creates and updates issues with detailed test information
- **Configurable Limits**: Top-K filtering to focus on most problematic tests
- **Rich Issue Templates**: Detailed issue descriptions with investigation guidance
