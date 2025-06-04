# Changelog

## [0.2.1](https://github.com/grafana/shared-workflows/compare/push-to-gcs-v0.2.0...push-to-gcs/v0.2.1) (2025-06-04)


### 🐛 Bug Fixes

* ensure every action disables git credential persistence ([#821](https://github.com/grafana/shared-workflows/issues/821)) ([31ebf3f](https://github.com/grafana/shared-workflows/commit/31ebf3f8e5d0f8709e6ec4ef73b39dd2bd08f959))
* **everything:** fix all things for zizmor ([af9b0c5](https://github.com/grafana/shared-workflows/commit/af9b0c52635d39023136fb9312a354f91d9b2bfd))


### 📝 Documentation

* **multiple-actions:** move permissions to job level in workflow examples ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))
* **multiple-actions:** move permissions to job level in workflows ([#969](https://github.com/grafana/shared-workflows/issues/969)) ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))
* update all readmes to replace hyphen with slash ([#1008](https://github.com/grafana/shared-workflows/issues/1008)) ([472df76](https://github.com/grafana/shared-workflows/commit/472df76fb1cbb92a17fb9e055bdf0d1399109ee3))


### 🤖 Continuous Integration

* don't persist shared workflows folder after action is done ([#905](https://github.com/grafana/shared-workflows/issues/905)) ([9a34c93](https://github.com/grafana/shared-workflows/commit/9a34c9302d2064c48e03cf7c4c7cd45998c4615e))
* remove gcp credentials after composite action finishes ([#925](https://github.com/grafana/shared-workflows/issues/925)) ([62f8dda](https://github.com/grafana/shared-workflows/commit/62f8ddaa78b23147b22ba6a38df2b97963dab4b3))


### 🔧 Miscellaneous Chores

* **deps:** update google-github-actions/upload-cloud-storage action to v2.2.2 ([#741](https://github.com/grafana/shared-workflows/issues/741)) ([5f7a536](https://github.com/grafana/shared-workflows/commit/5f7a5361daa274f9a1994893a4c21a8967cf2a24))
* **deps:** update google-github-actions/upload-cloud-storage action to v2.2.2 ([#749](https://github.com/grafana/shared-workflows/issues/749)) ([948bff0](https://github.com/grafana/shared-workflows/commit/948bff0b53f9d51876b8bca2cb1408384b4ce3b5))
* **main:** release push-to-gar-docker 0.3.0 ([#794](https://github.com/grafana/shared-workflows/issues/794)) ([a7bc536](https://github.com/grafana/shared-workflows/commit/a7bc5367c4a91c389526d58839d8f6224dba4dcc))

## [0.2.0](https://github.com/grafana/shared-workflows/compare/push-to-gcs-v0.1.0...push-to-gcs-v0.2.0) (2025-01-28)


### 🎉 Features

* add `service-account` input to `*-to-gcs` actions ([#720](https://github.com/grafana/shared-workflows/issues/720)) ([b4868e3](https://github.com/grafana/shared-workflows/commit/b4868e355b1e41a3ea54a272aa9970a809ec7ef1))
* **docs:** added EngHub doc links to corresponding actions readmes ([#635](https://github.com/grafana/shared-workflows/issues/635)) ([a7d04c1](https://github.com/grafana/shared-workflows/commit/a7d04c1e98496dbf07f8e44602933af07ba62f9f))


### 🔧 Miscellaneous Chores

* update readme when a new release is available ([#548](https://github.com/grafana/shared-workflows/issues/548)) ([9bf9163](https://github.com/grafana/shared-workflows/commit/9bf9163126c44247bcee6b6b9390eb488f9ead53))

## 0.1.0 (2024-11-28)


### 🎉 Features

* **push-to-gcs:** allow setting 'predefinedAcl' on objects when uploading ([#193](https://github.com/grafana/shared-workflows/issues/193)) ([97e6191](https://github.com/grafana/shared-workflows/commit/97e6191605de61d528f08aa85fa2f9ee2dfac355))


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v4.1.7 ([#244](https://github.com/grafana/shared-workflows/issues/244)) ([1d5fba5](https://github.com/grafana/shared-workflows/commit/1d5fba52e7cb2780dfd1af758e1d84e35ce6e8f7))
* **deps:** update actions/checkout action to v4.2.0 ([#313](https://github.com/grafana/shared-workflows/issues/313)) ([ba6268c](https://github.com/grafana/shared-workflows/commit/ba6268c6beef0ab5b461f45eef4cfe1b4e6d6013))
* **deps:** update actions/checkout action to v4.2.1 ([#445](https://github.com/grafana/shared-workflows/issues/445)) ([c72e039](https://github.com/grafana/shared-workflows/commit/c72e039d656ea7db5cbcfd98dffd0f8554e1f029))
* **deps:** update actions/checkout action to v4.2.2 ([#498](https://github.com/grafana/shared-workflows/issues/498)) ([7c6dbe2](https://github.com/grafana/shared-workflows/commit/7c6dbe23c5fd8f3ab5863fb0e3f9d95de621b746))
* **deps:** update google-github-actions/upload-cloud-storage action to v2.2.0 ([#261](https://github.com/grafana/shared-workflows/issues/261)) ([fc430d2](https://github.com/grafana/shared-workflows/commit/fc430d28d426938f51cf2cb575f775a1b7b55d40))
* **deps:** update google-github-actions/upload-cloud-storage action to v2.2.1 ([#510](https://github.com/grafana/shared-workflows/issues/510)) ([8545475](https://github.com/grafana/shared-workflows/commit/85454759d8be4364b64b3a409e640fbec5797dcc))
