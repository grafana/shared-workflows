# Changelog

## [2.0.1](https://github.com/grafana/shared-workflows/compare/dockerhub-login/v2.0.0...dockerhub-login/v2.0.1) (2026-07-23)


### 🔧 Miscellaneous Chores

* **deps:** update docker/login-action action to v4.4.0 ([#2168](https://github.com/grafana/shared-workflows/issues/2168)) ([21536e4](https://github.com/grafana/shared-workflows/commit/21536e4603262dd7882584949c2da3c8d01cba45))

## [2.0.0](https://github.com/grafana/shared-workflows/compare/dockerhub-login/v1.0.4...dockerhub-login/v2.0.0) (2026-06-10)


### ⚠ BREAKING CHANGES

* **get-vault-secrets:** remove export_env option, use JSON output always ([#1957](https://github.com/grafana/shared-workflows/issues/1957))

### 🎉 Features

* **get-vault-secrets:** remove export_env option, use JSON output always ([#1957](https://github.com/grafana/shared-workflows/issues/1957)) ([84e8abf](https://github.com/grafana/shared-workflows/commit/84e8abf0d3cd31cc8fa01d0e2c629a96864a108a))


### 🐛 Bug Fixes

* **create-github-app-token:** trigger release-please for reverted gatb change ([#1988](https://github.com/grafana/shared-workflows/issues/1988)) ([e6c8753](https://github.com/grafana/shared-workflows/commit/e6c875364b041be8288bcb1bee15f79cea31ffb1))
* reference sibling actions directly instead of checkout ([#2032](https://github.com/grafana/shared-workflows/issues/2032)) ([614ae58](https://github.com/grafana/shared-workflows/commit/614ae58b964b32c08d190dde334a583cc8373723))


### 🔧 Miscellaneous Chores

* **deps:** update docker/login-action action to v4.2.0 ([#2020](https://github.com/grafana/shared-workflows/issues/2020)) ([b2f47cf](https://github.com/grafana/shared-workflows/commit/b2f47cf9a22221fcb6fb816ffea2fd382971b218))

## [1.0.4](https://github.com/grafana/shared-workflows/compare/dockerhub-login/v1.0.3...dockerhub-login/v1.0.4) (2026-05-05)


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v6.0.2 ([#1672](https://github.com/grafana/shared-workflows/issues/1672)) ([3105e25](https://github.com/grafana/shared-workflows/commit/3105e251e687194e9b2b4b456cb2846a761e0df0))
* **deps:** update docker/login-action action to v3.7.0 ([#1684](https://github.com/grafana/shared-workflows/issues/1684)) ([60e2e92](https://github.com/grafana/shared-workflows/commit/60e2e922ca94bc80093982a1a26d4b7059874f6e))
* **deps:** update docker/login-action action to v4 ([#1786](https://github.com/grafana/shared-workflows/issues/1786)) ([abfcbbb](https://github.com/grafana/shared-workflows/commit/abfcbbbaf2b9d5811d44fd7260533b3c2e54ad79))
* **deps:** update docker/login-action action to v4.1.0 ([#1849](https://github.com/grafana/shared-workflows/issues/1849)) ([249c27a](https://github.com/grafana/shared-workflows/commit/249c27a2470e9c9eb6e1dd01177484b44aad9b5c))

## [1.0.3](https://github.com/grafana/shared-workflows/compare/dockerhub-login/v1.0.2...dockerhub-login/v1.0.3) (2025-12-08)


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v4.3.0 ([#1221](https://github.com/grafana/shared-workflows/issues/1221)) ([17ab531](https://github.com/grafana/shared-workflows/commit/17ab531bf2c16c79af38988e7caf7a3d8a37634b))
* **deps:** update actions/checkout action to v5 ([#1227](https://github.com/grafana/shared-workflows/issues/1227)) ([fd79c02](https://github.com/grafana/shared-workflows/commit/fd79c02730e0629f728e2f5c3d614545269208a9))
* **deps:** update actions/checkout action to v5.0.1 ([#1541](https://github.com/grafana/shared-workflows/issues/1541)) ([773f5b1](https://github.com/grafana/shared-workflows/commit/773f5b1eb7b717c5c89a2718c1c4322a45f2ed7f))
* **deps:** update actions/checkout action to v6 ([#1570](https://github.com/grafana/shared-workflows/issues/1570)) ([af4d9df](https://github.com/grafana/shared-workflows/commit/af4d9dfcfa9da2582544cd2a6e6dcf06e516f9ea))
* **deps:** update actions/checkout action to v6.0.1 ([#1590](https://github.com/grafana/shared-workflows/issues/1590)) ([2425a5f](https://github.com/grafana/shared-workflows/commit/2425a5fe46fb39d1d282caad59150165323e29a6))
* **deps:** update docker/login-action action to v3.5.0 ([#1187](https://github.com/grafana/shared-workflows/issues/1187)) ([d8060f2](https://github.com/grafana/shared-workflows/commit/d8060f2893039ae143f0d9c26f766565fb334c6a))
* **deps:** update docker/login-action action to v3.6.0 ([#1360](https://github.com/grafana/shared-workflows/issues/1360)) ([fcd9134](https://github.com/grafana/shared-workflows/commit/fcd9134b43d790d16878200b31a80f925b7a802c))

## [1.0.2](https://github.com/grafana/shared-workflows/compare/dockerhub-login-v1.0.1...dockerhub-login/v1.0.2) (2025-06-04)


### 🐛 Bug Fixes

* ensure every action disables git credential persistence ([#821](https://github.com/grafana/shared-workflows/issues/821)) ([31ebf3f](https://github.com/grafana/shared-workflows/commit/31ebf3f8e5d0f8709e6ec4ef73b39dd2bd08f959))


### 📝 Documentation

* **multiple-actions:** move permissions to job level in workflow examples ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))
* **multiple-actions:** move permissions to job level in workflows ([#969](https://github.com/grafana/shared-workflows/issues/969)) ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))
* update all readmes to replace hyphen with slash ([#1008](https://github.com/grafana/shared-workflows/issues/1008)) ([472df76](https://github.com/grafana/shared-workflows/commit/472df76fb1cbb92a17fb9e055bdf0d1399109ee3))


### 🔧 Miscellaneous Chores

* **deps:** update docker/login-action action to v3.4.0 ([#848](https://github.com/grafana/shared-workflows/issues/848)) ([117d851](https://github.com/grafana/shared-workflows/commit/117d8511cbc5da0337972deeb400c4298b057af3))
* **main:** release push-to-gar-docker 0.3.0 ([#794](https://github.com/grafana/shared-workflows/issues/794)) ([a7bc536](https://github.com/grafana/shared-workflows/commit/a7bc5367c4a91c389526d58839d8f6224dba4dcc))

## [1.0.1](https://github.com/grafana/shared-workflows/compare/dockerhub-login-v1.0.0...dockerhub-login-v1.0.1) (2025-01-28)


### 🔧 Miscellaneous Chores

* update readme when a new release is available ([#548](https://github.com/grafana/shared-workflows/issues/548)) ([9bf9163](https://github.com/grafana/shared-workflows/commit/9bf9163126c44247bcee6b6b9390eb488f9ead53))

## 1.0.0 (2024-10-16)


### 🔧 Miscellaneous Chores

* Bump upstream docker actions ([#51](https://github.com/grafana/shared-workflows/issues/51)) ([f33ebd9](https://github.com/grafana/shared-workflows/commit/f33ebd946aa2bcd994fb26afdedb575131a5b0b3))
* **deps:** update actions/checkout action to v4.1.7 ([#244](https://github.com/grafana/shared-workflows/issues/244)) ([1d5fba5](https://github.com/grafana/shared-workflows/commit/1d5fba52e7cb2780dfd1af758e1d84e35ce6e8f7))
* **deps:** update actions/checkout action to v4.2.0 ([#313](https://github.com/grafana/shared-workflows/issues/313)) ([ba6268c](https://github.com/grafana/shared-workflows/commit/ba6268c6beef0ab5b461f45eef4cfe1b4e6d6013))
* **deps:** update actions/checkout action to v4.2.1 ([#445](https://github.com/grafana/shared-workflows/issues/445)) ([c72e039](https://github.com/grafana/shared-workflows/commit/c72e039d656ea7db5cbcfd98dffd0f8554e1f029))
* **deps:** update docker/login-action action to v3.3.0 ([#254](https://github.com/grafana/shared-workflows/issues/254)) ([a678ac5](https://github.com/grafana/shared-workflows/commit/a678ac51c04a71178b65744276e210a6ad61b096))
