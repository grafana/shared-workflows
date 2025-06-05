# Changelog

## [Unreleased]

### Added
- Initial implementation of flaky test analysis action
- Loki integration for fetching test failure logs
- Comprehensive test suite with golden file testing
- Progressive PR structure for better code review

### Features
- **Loki Log Analysis**: Fetches and parses test failure logs using LogQL
- **Flaky Test Detection**: Identifies tests that fail inconsistently across branches
- **Configurable Limits**: Top-K filtering to focus on most problematic tests

### Technical Details
- Built with Go 1.22
- Implements interface-based architecture for testability
- Includes comprehensive mocking and golden file tests