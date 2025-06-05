# Changelog

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
