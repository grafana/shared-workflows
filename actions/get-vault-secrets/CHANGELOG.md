# Changelog

## [1.3.1](https://github.com/grafana/shared-workflows/compare/get-vault-secrets/v1.3.0...get-vault-secrets/v1.3.1) (2026-02-19)


### 🔧 Miscellaneous Chores

* **deps:** update actions/github-script action to v7.1.0 ([#1306](https://github.com/grafana/shared-workflows/issues/1306)) ([31b0c57](https://github.com/grafana/shared-workflows/commit/31b0c573abbbd9b56060318f7327ae8bb3ec041e))
* **deps:** update actions/github-script action to v8 ([#1307](https://github.com/grafana/shared-workflows/issues/1307)) ([078c4a8](https://github.com/grafana/shared-workflows/commit/078c4a8af09e06d646077550f9e0f68171d5881e))

## [1.3.0](https://github.com/grafana/shared-workflows/compare/get-vault-secrets/v1.2.1...get-vault-secrets/v1.3.0) (2025-08-19)


### 🎉 Features

* add helpful error handling for OIDC token failures ([#1171](https://github.com/grafana/shared-workflows/issues/1171)) ([6272233](https://github.com/grafana/shared-workflows/commit/62722333225a1fae03ae27a63d638f9bc2176edb))
* add support for ignoring missing secrets to get-vault-secrets action ([#1236](https://github.com/grafana/shared-workflows/issues/1236)) ([49ec8e2](https://github.com/grafana/shared-workflows/commit/49ec8e26626286f514faebe62b7eafcbe034fe30))


### 🤖 Continuous Integration

* enhance, fix, run pre-commit ([#1033](https://github.com/grafana/shared-workflows/issues/1033)) ([9ffb9ce](https://github.com/grafana/shared-workflows/commit/9ffb9cec67a7712b4247e4ac37eb69946d802aed))


### 🔧 Miscellaneous Chores

* **deps:** update hashicorp/vault-action action to v3.4.0 ([#1072](https://github.com/grafana/shared-workflows/issues/1072)) ([b1c5ce9](https://github.com/grafana/shared-workflows/commit/b1c5ce97ab9234836b62604ae52a90736f9a2f43))

## [1.2.1](https://github.com/grafana/shared-workflows/compare/get-vault-secrets-v1.2.0...get-vault-secrets/v1.2.1) (2025-06-04)


### 📝 Documentation

* **get-vault-secrets:** move permissions to job-level in example wor… ([#968](https://github.com/grafana/shared-workflows/issues/968)) ([b42327d](https://github.com/grafana/shared-workflows/commit/b42327db568c2c02b18bc53768aeb603de8c4a81))
* **get-vault-secrets:** move permissions to job-level in example workflow in readme ([b42327d](https://github.com/grafana/shared-workflows/commit/b42327db568c2c02b18bc53768aeb603de8c4a81))
* **get-vault-secrets:** update enghub link to point to new vault docs ([#992](https://github.com/grafana/shared-workflows/issues/992)) ([5e3deaf](https://github.com/grafana/shared-workflows/commit/5e3deaf6734ec48f298adadad5fb2d12a2139907))
* update all readmes to replace hyphen with slash ([#1008](https://github.com/grafana/shared-workflows/issues/1008)) ([472df76](https://github.com/grafana/shared-workflows/commit/472df76fb1cbb92a17fb9e055bdf0d1399109ee3))

## [1.2.0](https://github.com/grafana/shared-workflows/compare/get-vault-secrets-v1.1.0...get-vault-secrets-v1.2.0) (2025-04-29)


### 🎉 Features

* expose exportenv in get-vault-secrets ([4ea1476](https://github.com/grafana/shared-workflows/commit/4ea1476b297f17f388a7d9003ae28216c05bdb59))
* expose exportEnv in get-vault-secrets ([#903](https://github.com/grafana/shared-workflows/issues/903)) ([4ea1476](https://github.com/grafana/shared-workflows/commit/4ea1476b297f17f388a7d9003ae28216c05bdb59))


### 🐛 Bug Fixes

* **everything:** fix all things for zizmor ([af9b0c5](https://github.com/grafana/shared-workflows/commit/af9b0c52635d39023136fb9312a354f91d9b2bfd))
* **get-vault-secrets:** add outputs for export_env false ([#931](https://github.com/grafana/shared-workflows/issues/931)) ([7580496](https://github.com/grafana/shared-workflows/commit/75804962c1ba608148988c1e2dc35fbb0ee21746))


### 🔧 Miscellaneous Chores

* **deps:** update hashicorp/vault-action action to v3.3.0 ([#831](https://github.com/grafana/shared-workflows/issues/831)) ([98384a8](https://github.com/grafana/shared-workflows/commit/98384a8bf33e1bea6957186fa78b999da95dd657))
* **main:** release push-to-gar-docker 0.3.0 ([#794](https://github.com/grafana/shared-workflows/issues/794)) ([a7bc536](https://github.com/grafana/shared-workflows/commit/a7bc5367c4a91c389526d58839d8f6224dba4dcc))

## [1.1.0](https://github.com/grafana/shared-workflows/compare/get-vault-secrets-v1.0.1...get-vault-secrets-v1.1.0) (2025-01-28)


### 🎉 Features

* **docs:** added EngHub doc links to corresponding actions readmes ([#635](https://github.com/grafana/shared-workflows/issues/635)) ([a7d04c1](https://github.com/grafana/shared-workflows/commit/a7d04c1e98496dbf07f8e44602933af07ba62f9f))


### 🔧 Miscellaneous Chores

* **deps:** update hashicorp/vault-action action to v3.1.0 ([#684](https://github.com/grafana/shared-workflows/issues/684)) ([4cf72da](https://github.com/grafana/shared-workflows/commit/4cf72da04db103f024b145c360f0743c298740b5))
* update readme when a new release is available ([#548](https://github.com/grafana/shared-workflows/issues/548)) ([9bf9163](https://github.com/grafana/shared-workflows/commit/9bf9163126c44247bcee6b6b9390eb488f9ead53))

## [1.0.1](https://github.com/grafana/shared-workflows/compare/get-vault-secrets-v1.0.0...get-vault-secrets-v1.0.1) (2024-11-28)


### 🔧 Miscellaneous Chores

* **deps:** update google-github-actions/auth action to v2.1.7 ([#509](https://github.com/grafana/shared-workflows/issues/509)) ([41774d7](https://github.com/grafana/shared-workflows/commit/41774d7ebb3ca78e05aa6d2007e5e98c7a2fcf4f))
* **get-vault-secrets:** remove the iap-auth step from get-vault-secrets ([#520](https://github.com/grafana/shared-workflows/issues/520)) ([cbdc528](https://github.com/grafana/shared-workflows/commit/cbdc528c28586253be7c33b531916b2bd7431324))

## 1.0.0 (2024-10-16)


### 🎉 Features

* added github oidc token as a header ([#471](https://github.com/grafana/shared-workflows/issues/471)) ([6426ecd](https://github.com/grafana/shared-workflows/commit/6426ecdc24bc202b5d0ac928f1bbede1ebcc4298))


### 🐛 Bug Fixes

* **renovate:** fix action versions, add more triggers ([#249](https://github.com/grafana/shared-workflows/issues/249)) ([3b1e65f](https://github.com/grafana/shared-workflows/commit/3b1e65f1b3b563a309b4aa5f888d916ad389cec3))


### 🔧 Miscellaneous Chores

* **deps:** update google-github-actions/auth action to v2.1.5 ([#248](https://github.com/grafana/shared-workflows/issues/248)) ([a5d1613](https://github.com/grafana/shared-workflows/commit/a5d1613fba998ba9b99b7267b6f9b915562da962))
* **deps:** update google-github-actions/auth action to v2.1.6 ([#436](https://github.com/grafana/shared-workflows/issues/436)) ([a275eef](https://github.com/grafana/shared-workflows/commit/a275eefa9f63e3bec05bd90ea77cfbbc9879afe8))
* **deps:** update hashicorp/vault-action action to v3 ([#267](https://github.com/grafana/shared-workflows/issues/267)) ([553b76e](https://github.com/grafana/shared-workflows/commit/553b76e74be26afbed46de7df5393459deccacad))
* Update action dependencies ([#39](https://github.com/grafana/shared-workflows/issues/39)) ([b271a8b](https://github.com/grafana/shared-workflows/commit/b271a8b01e61d00dc987dbb77744bd9e01fe862d))
