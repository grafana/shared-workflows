# Changelog

## [1.1.0](https://github.com/grafana/shared-workflows/compare/component-change-detection/v1.0.2...component-change-detection/v1.1.0) (2026-07-23)


### 🎉 Features

* replace dawidd6/action-download-artifact with first-party GitHub actions ([#1955](https://github.com/grafana/shared-workflows/issues/1955)) ([e378ca0](https://github.com/grafana/shared-workflows/commit/e378ca000c6166a44060ed3a3c987a53da14128d))


### 🔧 Miscellaneous Chores

* **deps:** update actions/setup-go action to v6.5.0 ([#2145](https://github.com/grafana/shared-workflows/issues/2145)) ([67f1fea](https://github.com/grafana/shared-workflows/commit/67f1fea99eb1c85cc97bb621d3b3130381f623a9))
* **deps:** update module github.com/go-logfmt/logfmt to v0.6.1 ([#2185](https://github.com/grafana/shared-workflows/issues/2185)) ([0e76b00](https://github.com/grafana/shared-workflows/commit/0e76b0012f40d2ce5798f06eb14824b95276a183))

## [1.0.2](https://github.com/grafana/shared-workflows/compare/component-change-detection/v1.0.1...component-change-detection/v1.0.2) (2026-06-10)


### 🐛 Bug Fixes

* **create-github-app-token:** trigger release-please for reverted gatb change ([#1988](https://github.com/grafana/shared-workflows/issues/1988)) ([e6c8753](https://github.com/grafana/shared-workflows/commit/e6c875364b041be8288bcb1bee15f79cea31ffb1))

## [1.0.1](https://github.com/grafana/shared-workflows/compare/component-change-detection/v1.0.0...component-change-detection/v1.0.1) (2026-05-05)


### 🔧 Miscellaneous Chores

* **deps:** update actions/setup-go action to v6.3.0 ([#1756](https://github.com/grafana/shared-workflows/issues/1756)) ([c6b0752](https://github.com/grafana/shared-workflows/commit/c6b07529443393154d824d1ad0e707f4b3d090f6))
* **deps:** update actions/setup-go action to v6.4.0 ([#1837](https://github.com/grafana/shared-workflows/issues/1837)) ([170bd5b](https://github.com/grafana/shared-workflows/commit/170bd5b0ba3c2414519216fd5d7f0fe5a40e3f40))
* **deps:** update dawidd6/action-download-artifact action to v15 ([#1738](https://github.com/grafana/shared-workflows/issues/1738)) ([36005ff](https://github.com/grafana/shared-workflows/commit/36005ff276117278dcb4498b82975048530ef069))
* **deps:** update dawidd6/action-download-artifact action to v16 ([#1753](https://github.com/grafana/shared-workflows/issues/1753)) ([b71be18](https://github.com/grafana/shared-workflows/commit/b71be180de45d83fe3f1641ff8ced2b2d967c155))
* **deps:** update dawidd6/action-download-artifact action to v18 ([#1812](https://github.com/grafana/shared-workflows/issues/1812)) ([e40c7fe](https://github.com/grafana/shared-workflows/commit/e40c7fecb195dd2644cd3991abf756782cbd4143))
* **deps:** update dawidd6/action-download-artifact action to v19 ([#1822](https://github.com/grafana/shared-workflows/issues/1822)) ([4d6ff3e](https://github.com/grafana/shared-workflows/commit/4d6ff3ed5e93b9b5f229ca66ead5430fe9b421c7))
* **deps:** update dawidd6/action-download-artifact action to v20 ([#1850](https://github.com/grafana/shared-workflows/issues/1850)) ([8027637](https://github.com/grafana/shared-workflows/commit/80276376018d7be3c868bc7ab37711db9ca5d7c2))
* **deps:** update dawidd6/action-download-artifact action to v21 ([#1914](https://github.com/grafana/shared-workflows/issues/1914)) ([c083482](https://github.com/grafana/shared-workflows/commit/c083482fa9e363e80b376d0317172bc3f55ddb3d))

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
