# Changelog

## [Unreleased]

### Added

- Initial implementation of flaky test analysis action
- Loki integration for fetching test failure logs
- Git history analysis to find test authors
- Comprehensive test suite with golden file testing

### Features

- **Loki Log Analysis**: Fetches and parses test failure logs using LogQL
- **Flaky Test Detection**: Identifies tests that fail inconsistently across branches
- **Git Author Tracking**: Finds recent commits that modified flaky tests
- **Configurable Limits**: Top-K filtering to focus on most problematic tests
