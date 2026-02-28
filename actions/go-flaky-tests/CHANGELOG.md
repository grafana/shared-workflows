# Changelog

## 0.1.0 (2026-02-28)


### 🎉 Features

* add git history analysis to go flaky tests ([#1029](https://github.com/grafana/shared-workflows/issues/1029)) ([d005a71](https://github.com/grafana/shared-workflows/commit/d005a712405c56f649db427886a189e353322811))
* add go-flaky-tests action to release-please config ([#1416](https://github.com/grafana/shared-workflows/issues/1416)) ([eed6978](https://github.com/grafana/shared-workflows/commit/eed69781bdf177e1c126172712690dae79b0bc11))
* **go-flaky-tests:** add github issue management ([#1276](https://github.com/grafana/shared-workflows/issues/1276)) ([f2c0b02](https://github.com/grafana/shared-workflows/commit/f2c0b0223f8e954cdc034aed2de5be6b7a90df3d))
* **go-flaky-tests:** add ignored-tests option to filter out specific test failures ([#1359](https://github.com/grafana/shared-workflows/issues/1359)) ([1f61f1d](https://github.com/grafana/shared-workflows/commit/1f61f1dfffd8226c0a403db65c5c9a87fd80e6cb))


### 🐛 Bug Fixes

* **deps:** update module github.com/stretchr/testify to v1.11.0 ([#1263](https://github.com/grafana/shared-workflows/issues/1263)) ([92d0612](https://github.com/grafana/shared-workflows/commit/92d06123e73d57688a53671d0239197efb06cc60))
* **deps:** update module github.com/stretchr/testify to v1.11.1 ([#1279](https://github.com/grafana/shared-workflows/issues/1279)) ([6901f03](https://github.com/grafana/shared-workflows/commit/6901f036a3aa16cfaaba7020f3515c31eaa2f999))


### 🔧 Miscellaneous Chores

* **deps:** update actions/setup-go action to v6 ([#1299](https://github.com/grafana/shared-workflows/issues/1299)) ([6262c5e](https://github.com/grafana/shared-workflows/commit/6262c5e47024d01fd9a114356509ceb9872072b4))
* **deps:** update actions/setup-go action to v6.2.0 ([#1657](https://github.com/grafana/shared-workflows/issues/1657)) ([d29b916](https://github.com/grafana/shared-workflows/commit/d29b9161f1803baed4a7305c85cb5a3018bc3c3e))
* **deps:** update actions/setup-go digest to 4dc6199 ([#1551](https://github.com/grafana/shared-workflows/issues/1551)) ([444d030](https://github.com/grafana/shared-workflows/commit/444d030c6286b899fda8821d4108d8b74cc1c3b8))
* **deps:** update dependency go to v1.25.0 ([#1251](https://github.com/grafana/shared-workflows/issues/1251)) ([5b400bc](https://github.com/grafana/shared-workflows/commit/5b400bcc8e746df9281ac11d28c10e4d26a20c9e))

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
