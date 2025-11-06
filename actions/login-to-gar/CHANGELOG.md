# Changelog

## [1.0.1](https://github.com/grafana/shared-workflows/compare/login-to-gar/v1.0.0...login-to-gar/v1.0.1) (2025-11-05)


### 🔧 Miscellaneous Chores

* **deps:** update google-github-actions/auth action to v2.1.11 ([#1150](https://github.com/grafana/shared-workflows/issues/1150)) ([895722b](https://github.com/grafana/shared-workflows/commit/895722b12337ee97909efa8a78886ee69297ed50))
* **deps:** update google-github-actions/auth action to v2.1.12 ([#1184](https://github.com/grafana/shared-workflows/issues/1184)) ([ee464c5](https://github.com/grafana/shared-workflows/commit/ee464c522eba7d1a22b82d27739f4bf789102900))
* **deps:** update google-github-actions/auth action to v3 ([#1285](https://github.com/grafana/shared-workflows/issues/1285)) ([9f0b61a](https://github.com/grafana/shared-workflows/commit/9f0b61a459b2106ee915a3c482e493c7f659312f))
* **deps:** update google-github-actions/setup-gcloud action to v2.1.5 ([#1151](https://github.com/grafana/shared-workflows/issues/1151)) ([84f55b1](https://github.com/grafana/shared-workflows/commit/84f55b125e875869f7aead3c1ed900eaae2735bb))
* **deps:** update google-github-actions/setup-gcloud action to v2.2.0 ([#1211](https://github.com/grafana/shared-workflows/issues/1211)) ([dc9441f](https://github.com/grafana/shared-workflows/commit/dc9441f43be7baaf190c5b1b6c0fa3589a988907))
* **deps:** update google-github-actions/setup-gcloud action to v2.2.1 ([#1266](https://github.com/grafana/shared-workflows/issues/1266)) ([8bb65cb](https://github.com/grafana/shared-workflows/commit/8bb65cb7cc5b627ec350a229de34aee58872650a))
* **deps:** update google-github-actions/setup-gcloud action to v3 ([#1267](https://github.com/grafana/shared-workflows/issues/1267)) ([ac79b81](https://github.com/grafana/shared-workflows/commit/ac79b814a3c74384d24bb0431a3d99caa948e806))
* **deps:** update google-github-actions/setup-gcloud action to v3.0.1 ([#1283](https://github.com/grafana/shared-workflows/issues/1283)) ([3a233ec](https://github.com/grafana/shared-workflows/commit/3a233ece646e1a9715d9e4ff27f7d8f98ae8b232))

## [1.0.0](https://github.com/grafana/shared-workflows/compare/login-to-gar/v0.4.3...login-to-gar/v1.0.0) (2025-06-17)


### ⚠ BREAKING CHANGES

* **login-to-gar:** Update configurations which specify `delete-credentials: false` to have `workspace-credentials: true` instead. If you don't have the option, you are not affected.
* only allow direct workload identity federation in login-to-gar ([#1009](https://github.com/grafana/shared-workflows/issues/1009))

### 🎉 Features

* **login-to-gar:** store credentials in temporary location by default ([#1023](https://github.com/grafana/shared-workflows/issues/1023)) ([fe29dde](https://github.com/grafana/shared-workflows/commit/fe29dde24ab0697084e75883d351eca1c961e352))
* only allow direct workload identity federation in login-to-gar ([#1009](https://github.com/grafana/shared-workflows/issues/1009)) ([0789629](https://github.com/grafana/shared-workflows/commit/078962963e9e785bbe565287f41f96c23ba03274))


### 🐛 Bug Fixes

* **login-to-gar:** check if delete_credentials_file is set ([#1020](https://github.com/grafana/shared-workflows/issues/1020)) ([7803c2c](https://github.com/grafana/shared-workflows/commit/7803c2ce62f8d6d5da83cac0ae9af3d57b70a0ff))


### 📝 Documentation

* add warning about using `checkout` action before `login-to-gar` ([#1012](https://github.com/grafana/shared-workflows/issues/1012)) ([cb40def](https://github.com/grafana/shared-workflows/commit/cb40def95f3c449ae8c7f23fa302c22bf9355fb5))


### 🤖 Continuous Integration

* add section for gha-creds jsons and .gitignore ([#1021](https://github.com/grafana/shared-workflows/issues/1021)) ([f008500](https://github.com/grafana/shared-workflows/commit/f008500f574f01cf9fcc5054be2464d6f5d6dcec))

## [0.4.3](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.4.2...login-to-gar/v0.4.3) (2025-06-04)


### 📝 Documentation

* update all readmes to replace hyphen with slash ([#1008](https://github.com/grafana/shared-workflows/issues/1008)) ([472df76](https://github.com/grafana/shared-workflows/commit/472df76fb1cbb92a17fb9e055bdf0d1399109ee3))

## [0.4.2](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.4.1...login-to-gar-v0.4.2) (2025-05-28)


### 🐛 Bug Fixes

* **login-to-gar:** replace hardcoded opt dir with runner temp env var ([#1001](https://github.com/grafana/shared-workflows/issues/1001)) ([d03fbe2](https://github.com/grafana/shared-workflows/commit/d03fbe21194b8bae035dabfba8fdabe19c122660))

## [0.4.1](https://github.com/grafana/shared-workflows/compare/login-to-gar-v0.4.0...login-to-gar-v0.4.1) (2025-05-26)


### 🐛 Bug Fixes

* use custom step for docker-credential-gcr ([#996](https://github.com/grafana/shared-workflows/issues/996)) ([36bbb4c](https://github.com/grafana/shared-workflows/commit/36bbb4c0ab04a493b5b76ee6e00d4476a0e954f5))


### 📝 Documentation

* add inputs section in login-to-gar action ([#961](https://github.com/grafana/shared-workflows/issues/961)) ([3ce65db](https://github.com/grafana/shared-workflows/commit/3ce65db098d2e00917a8b98c49a5417dd7a8797a))
* **multiple-actions:** move permissions to job level in workflow examples ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))
* **multiple-actions:** move permissions to job level in workflows ([#969](https://github.com/grafana/shared-workflows/issues/969)) ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))

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
