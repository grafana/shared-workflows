# Changelog

## [0.4.0](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.3.0...login-to-gar-v0.4.0) (2025-04-30)


### 🎉 Features

* use `docker-credential-gcr` instead of `auth_token` for `login-to-gar` action ([#921](https://github.com/grafana/shared-workflows/issues/921)) ([cac9a09](https://github.com/grafana/shared-workflows/commit/cac9a09f00dfb7c7743500f1986d8faebca72f9f))


### 🐛 Bug Fixes

* **everything:** fix all things for zizmor ([af9b0c5](https://github.com/grafana/shared-workflows/commit/af9b0c52635d39023136fb9312a354f91d9b2bfd))
* make default `delete_credentials_file` value false ([#950](https://github.com/grafana/shared-workflows/issues/950)) ([71ec5a1](https://github.com/grafana/shared-workflows/commit/71ec5a1861019932272c4ec12a8d7903049797c5))


### 🤖 Continuous Integration

* remove gcp credentials after composite action finishes ([#925](https://github.com/grafana/shared-workflows/issues/925)) ([62f8dda](https://github.com/grafana/shared-workflows/commit/62f8ddaa78b23147b22ba6a38df2b97963dab4b3))


### 🔧 Miscellaneous Chores

* **deps:** update google-github-actions/auth action to v2.1.10 ([#926](https://github.com/grafana/shared-workflows/issues/926)) ([fa48192](https://github.com/grafana/shared-workflows/commit/fa48192dac470ae356b3f7007229f3ac28c48a25))
* **deps:** update google-github-actions/auth action to v2.1.9 ([#924](https://github.com/grafana/shared-workflows/issues/924)) ([2774f26](https://github.com/grafana/shared-workflows/commit/2774f26e2321f825e20c85e424a1c6fa8298d820))

## [0.3.0](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.2.2...login-to-gar-v0.3.0) (2025-04-23)


### 🎉 Features

* use auth_token in login-to-gar action ([#846](https://github.com/grafana/shared-workflows/issues/846)) ([e65ba18](https://github.com/grafana/shared-workflows/commit/e65ba18704a12d05c4c5ad00439c31d5861ba9a1))


### 🤖 Continuous Integration

* make configure-docker less verbose ([#824](https://github.com/grafana/shared-workflows/issues/824)) ([623010a](https://github.com/grafana/shared-workflows/commit/623010ae889725b324e1ae1b3572d1be621b76b9))
* stop persisting credentials in google auth steps ([#916](https://github.com/grafana/shared-workflows/issues/916)) ([4d185da](https://github.com/grafana/shared-workflows/commit/4d185da792dd4520730b3b60ceedb1c9cb16cb6c))


### 🔧 Miscellaneous Chores

* **deps:** update docker/login-action action to v3.4.0 ([#848](https://github.com/grafana/shared-workflows/issues/848)) ([117d851](https://github.com/grafana/shared-workflows/commit/117d8511cbc5da0337972deeb400c4298b057af3))

## [0.2.2](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.2.1...login-to-gar-v0.2.2) (2025-02-26)


### 🐛 Bug Fixes

* install gcloud in login-to-gar action ([#813](https://github.com/grafana/shared-workflows/issues/813)) ([935970b](https://github.com/grafana/shared-workflows/commit/935970b13327698aa89e768f511a45432285f5cd))


### 🔧 Miscellaneous Chores

* **main:** release push-to-gar-docker 0.3.0 ([#794](https://github.com/grafana/shared-workflows/issues/794)) ([a7bc536](https://github.com/grafana/shared-workflows/commit/a7bc5367c4a91c389526d58839d8f6224dba4dcc))

## [0.2.1](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.2.0...login-to-gar-v0.2.1) (2025-02-17)


### ♻️ Code Refactoring

* simplify login-to-gar, removes describe service account ([#781](https://github.com/grafana/shared-workflows/issues/781)) ([4e593d1](https://github.com/grafana/shared-workflows/commit/4e593d17433d7b3968ae727e0dc509b77a074ebe))

## [0.2.0](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.1.1...login-to-gar-v0.2.0) (2025-02-12)


### 🎉 Features

* update login-to-gar action to include direct wif ([#772](https://github.com/grafana/shared-workflows/issues/772)) ([ed6261d](https://github.com/grafana/shared-workflows/commit/ed6261dda7dd83c57740658f195030be6e9723e8))


### 🔧 Miscellaneous Chores

* **deps:** pin google-github-actions/setup-gcloud action to 6189d56 ([#774](https://github.com/grafana/shared-workflows/issues/774)) ([315dfc8](https://github.com/grafana/shared-workflows/commit/315dfc8f3d82295337d2032840f9c22848868296))
* **deps:** update google-github-actions/auth action to v2.1.8 ([#740](https://github.com/grafana/shared-workflows/issues/740)) ([f75f620](https://github.com/grafana/shared-workflows/commit/f75f620c6800b60d1a31262154e90b5c7a3ee955))
* **deps:** update google-github-actions/auth action to v2.1.8 ([#775](https://github.com/grafana/shared-workflows/issues/775)) ([c773be9](https://github.com/grafana/shared-workflows/commit/c773be9039d28ffb2cf9740e39789eccc1c701e3))

## [0.1.1](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.1.0...login-to-gar-v0.1.1) (2025-01-29)


### 🔧 Miscellaneous Chores

* update readme when a new release is available ([#548](https://github.com/grafana/shared-workflows/issues/548)) ([9bf9163](https://github.com/grafana/shared-workflows/commit/9bf9163126c44247bcee6b6b9390eb488f9ead53))

## 0.1.0 (2024-11-28)


### 🔧 Miscellaneous Chores

* **deps:** update docker/login-action action to v3.3.0 ([#254](https://github.com/grafana/shared-workflows/issues/254)) ([a678ac5](https://github.com/grafana/shared-workflows/commit/a678ac51c04a71178b65744276e210a6ad61b096))
* **deps:** update google-github-actions/auth action to v2.1.5 ([#248](https://github.com/grafana/shared-workflows/issues/248)) ([a5d1613](https://github.com/grafana/shared-workflows/commit/a5d1613fba998ba9b99b7267b6f9b915562da962))
* **deps:** update google-github-actions/auth action to v2.1.6 ([#436](https://github.com/grafana/shared-workflows/issues/436)) ([a275eef](https://github.com/grafana/shared-workflows/commit/a275eefa9f63e3bec05bd90ea77cfbbc9879afe8))
* **deps:** update google-github-actions/auth action to v2.1.7 ([#509](https://github.com/grafana/shared-workflows/issues/509)) ([41774d7](https://github.com/grafana/shared-workflows/commit/41774d7ebb3ca78e05aa6d2007e5e98c7a2fcf4f))
