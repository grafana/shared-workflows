# Changelog

## [0.7.0](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker/v0.6.1...push-to-gar-docker/v0.7.0) (2025-12-11)


### 🎉 Features

* **push-to-gar-docker:** parse registry as input in login-to-gar ([#1492](https://github.com/grafana/shared-workflows/issues/1492)) ([b545d8a](https://github.com/grafana/shared-workflows/commit/b545d8afe61d00248004bf0f5fe076d31037290e))
* support passing 'builder' parameter in push-to-gar-docker ([#1568](https://github.com/grafana/shared-workflows/issues/1568)) ([68818a1](https://github.com/grafana/shared-workflows/commit/68818a1d7ffad7f276b89ef3e1835054d4106d46))


### 📝 Documentation

* improve docker build action docs ([#1486](https://github.com/grafana/shared-workflows/issues/1486)) ([2dd0b03](https://github.com/grafana/shared-workflows/commit/2dd0b0349e130ca5ccf86b3a61250589a840bdb2))


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v5.0.1 ([#1541](https://github.com/grafana/shared-workflows/issues/1541)) ([773f5b1](https://github.com/grafana/shared-workflows/commit/773f5b1eb7b717c5c89a2718c1c4322a45f2ed7f))
* **deps:** update actions/checkout action to v6 ([#1570](https://github.com/grafana/shared-workflows/issues/1570)) ([af4d9df](https://github.com/grafana/shared-workflows/commit/af4d9dfcfa9da2582544cd2a6e6dcf06e516f9ea))
* **deps:** update actions/checkout action to v6.0.1 ([#1590](https://github.com/grafana/shared-workflows/issues/1590)) ([2425a5f](https://github.com/grafana/shared-workflows/commit/2425a5fe46fb39d1d282caad59150165323e29a6))
* **deps:** update docker/metadata-action action to v5.10.0 ([#1582](https://github.com/grafana/shared-workflows/issues/1582)) ([d80ddba](https://github.com/grafana/shared-workflows/commit/d80ddba3b588ad911410ce91c599bbbe513196b0))
* **deps:** update docker/metadata-action action to v5.9.0 ([#1501](https://github.com/grafana/shared-workflows/issues/1501)) ([2d5a067](https://github.com/grafana/shared-workflows/commit/2d5a0678eb32b0fd6655b4f7a3a7ec72eaf530ca))
* **multiple:** deprecate old docker actions and add migration guide ([#1606](https://github.com/grafana/shared-workflows/issues/1606)) ([b6c252d](https://github.com/grafana/shared-workflows/commit/b6c252dc86cb65eaf2d8344d6d51ca07436214a2))

## [0.6.1](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker/v0.6.0...push-to-gar-docker/v0.6.1) (2025-10-06)


### 🐛 Bug Fixes

* push-to-gar-docker pin buildx to 0.28.0 ([#1378](https://github.com/grafana/shared-workflows/issues/1378)) ([60eb720](https://github.com/grafana/shared-workflows/commit/60eb7207fa9ce578ff6e52c68f9e13542499f41b))

## [0.6.0](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker/v0.5.2...push-to-gar-docker/v0.6.0) (2025-10-03)


### 🎉 Features

* support load parameter in push-to-gar-docker ([#1190](https://github.com/grafana/shared-workflows/issues/1190)) ([bc06cdd](https://github.com/grafana/shared-workflows/commit/bc06cdd4721071acd1beb13f34a737801e52357f))


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v4.3.0 ([#1221](https://github.com/grafana/shared-workflows/issues/1221)) ([17ab531](https://github.com/grafana/shared-workflows/commit/17ab531bf2c16c79af38988e7caf7a3d8a37634b))
* **deps:** update actions/checkout action to v5 ([#1227](https://github.com/grafana/shared-workflows/issues/1227)) ([fd79c02](https://github.com/grafana/shared-workflows/commit/fd79c02730e0629f728e2f5c3d614545269208a9))
* **deps:** update docker/metadata-action action to v5.8.0 ([#1182](https://github.com/grafana/shared-workflows/issues/1182)) ([315e38a](https://github.com/grafana/shared-workflows/commit/315e38a03f442c39bd82e902b88d8ba6ff8879b7))

## [0.5.2](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker/v0.5.1...push-to-gar-docker/v0.5.2) (2025-07-23)


### 🐛 Bug Fixes

* remediate latest zizmor findings, fix supplying zizmor config ([#1101](https://github.com/grafana/shared-workflows/issues/1101)) ([712c599](https://github.com/grafana/shared-workflows/commit/712c59975bc0de22124b866153826f04023f18fd))


### 📝 Documentation

* update all readmes to replace hyphen with slash ([#1008](https://github.com/grafana/shared-workflows/issues/1008)) ([472df76](https://github.com/grafana/shared-workflows/commit/472df76fb1cbb92a17fb9e055bdf0d1399109ee3))


### 🔧 Miscellaneous Chores

* **deps:** update docker/build-push-action action to v6.18.0 ([#1065](https://github.com/grafana/shared-workflows/issues/1065)) ([5b5ee4c](https://github.com/grafana/shared-workflows/commit/5b5ee4cf0a527daf5e32b7f968637b8a8ed7efcb))
* **deps:** update docker/setup-buildx-action action to v3.11.1 ([#1068](https://github.com/grafana/shared-workflows/issues/1068)) ([5233cbc](https://github.com/grafana/shared-workflows/commit/5233cbc5d62242fb17b2259c2c4bd2a628af5528))

## [0.5.1](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker-v0.5.0...push-to-gar-docker/v0.5.1) (2025-06-02)


### 📝 Documentation

* **multiple:** add docker cache notes to build-push-to-dockerhub and push-to-gar-docker ([#1003](https://github.com/grafana/shared-workflows/issues/1003)) ([e5377f9](https://github.com/grafana/shared-workflows/commit/e5377f9c2aee143ccf63001896fa59eef7bea1d5))

## [0.5.0](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker-v0.4.1...push-to-gar-docker-v0.5.0) (2025-05-28)


### 🎉 Features

* **push-to-gar-docker,build-push-to-dockerhub:** add support for buildkit secrets ([#990](https://github.com/grafana/shared-workflows/issues/990)) ([bfed586](https://github.com/grafana/shared-workflows/commit/bfed586d71f4799f2506878776b481d00ca84bda))
* **push-to-gar-docker:** enable docker mirror for buildx on self-hosted runners ([#1000](https://github.com/grafana/shared-workflows/issues/1000)) ([77d2ce5](https://github.com/grafana/shared-workflows/commit/77d2ce511c62e35630fdef86985e6faf4a28afcc))


### 📝 Documentation

* **multiple-actions:** move permissions to job level in workflow examples ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))
* **multiple-actions:** move permissions to job level in workflows ([#969](https://github.com/grafana/shared-workflows/issues/969)) ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))

## [0.4.1](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker-v0.4.0...push-to-gar-docker-v0.4.1) (2025-05-07)


### 🤖 Continuous Integration

* remove gcp credentials after composite action finishes ([#925](https://github.com/grafana/shared-workflows/issues/925)) ([62f8dda](https://github.com/grafana/shared-workflows/commit/62f8ddaa78b23147b22ba6a38df2b97963dab4b3))

## [0.4.0](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker-v0.3.1...push-to-gar-docker-v0.4.0) (2025-04-29)


### 🎉 Features

* **push-to-gar-docker:** add target input for multi-stage builds ([#904](https://github.com/grafana/shared-workflows/issues/904)) ([fd2e2da](https://github.com/grafana/shared-workflows/commit/fd2e2da52d1a729ae0985fdf6ff85b33710381f9))


### 🐛 Bug Fixes

* ensure every action disables git credential persistence ([#821](https://github.com/grafana/shared-workflows/issues/821)) ([31ebf3f](https://github.com/grafana/shared-workflows/commit/31ebf3f8e5d0f8709e6ec4ef73b39dd2bd08f959))
* **everything:** fix all things for zizmor ([af9b0c5](https://github.com/grafana/shared-workflows/commit/af9b0c52635d39023136fb9312a354f91d9b2bfd))
* update buildx version to `latest` ([#895](https://github.com/grafana/shared-workflows/issues/895)) ([f366250](https://github.com/grafana/shared-workflows/commit/f366250e45bf8aadca4bb5e00802eb3854fb111d))


### 🤖 Continuous Integration

* don't persist shared workflows folder after action is done ([#905](https://github.com/grafana/shared-workflows/issues/905)) ([9a34c93](https://github.com/grafana/shared-workflows/commit/9a34c9302d2064c48e03cf7c4c7cd45998c4615e))


### 🔧 Miscellaneous Chores

* **deps:** update docker/build-push-action action to v6.15.0 ([#816](https://github.com/grafana/shared-workflows/issues/816)) ([0ae253d](https://github.com/grafana/shared-workflows/commit/0ae253d4a198408407a161de482680eddf2dfa42))
* **deps:** update docker/build-push-action action to v6.16.0 ([#923](https://github.com/grafana/shared-workflows/issues/923)) ([a301072](https://github.com/grafana/shared-workflows/commit/a30107276148b4f29eaeaef05a3f9173d1aa0ad9))
* **deps:** update docker/metadata-action action to v5.7.0 ([#818](https://github.com/grafana/shared-workflows/issues/818)) ([9f9b2eb](https://github.com/grafana/shared-workflows/commit/9f9b2eb3897a39fd65e5b92f17a60704925f94c4))
* **deps:** update docker/setup-buildx-action action to v3.10.0 ([#819](https://github.com/grafana/shared-workflows/issues/819)) ([09fb633](https://github.com/grafana/shared-workflows/commit/09fb633eb9f6c77153fa941e662be7cd418ca1fb))

## [0.3.1](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker-v0.2.1...push-to-gar-docker-v0.3.1) (2025-02-26)


### 🎉 Features

* **push-to-gar-docker:** add custom label input ([#795](https://github.com/grafana/shared-workflows/issues/795)) ([b6aa0b6](https://github.com/grafana/shared-workflows/commit/b6aa0b6312f7cd58416885007204ac9a4a71c094))


### 🔧 Miscellaneous Chores

* **deps:** update docker/build-push-action action to v6.14.0 ([#793](https://github.com/grafana/shared-workflows/issues/793)) ([d750654](https://github.com/grafana/shared-workflows/commit/d750654d770aefa0516e11735cdfdb89b7a380a1))

## [0.2.1](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker-v0.2.0...push-to-gar-docker-v0.2.1) (2025-02-13)


### 🔧 Miscellaneous Chores

* **deps:** update docker/setup-buildx-action action to v3.9.0 ([#755](https://github.com/grafana/shared-workflows/issues/755)) ([8dd62e3](https://github.com/grafana/shared-workflows/commit/8dd62e320f60df7426d30b67c9b26f17af352ed7))

## [0.2.0](https://github.com/grafana/shared-workflows/compare/push-to-gar-docker-v0.1.0...push-to-gar-docker-v0.2.0) (2025-01-29)


### 🎉 Features

* **docs:** added EngHub doc links to corresponding actions readmes ([#635](https://github.com/grafana/shared-workflows/issues/635)) ([a7d04c1](https://github.com/grafana/shared-workflows/commit/a7d04c1e98496dbf07f8e44602933af07ba62f9f))


### 🐛 Bug Fixes

* login to GAR only on push events ([#670](https://github.com/grafana/shared-workflows/issues/670)) ([c1714b0](https://github.com/grafana/shared-workflows/commit/c1714b03ca3d5cb08308ffb857e615cb9b6d439d))


### 🔧 Miscellaneous Chores

* **deps:** update docker/build-push-action action to v6.11.0 ([#679](https://github.com/grafana/shared-workflows/issues/679)) ([e1b07ec](https://github.com/grafana/shared-workflows/commit/e1b07ec29d283a54c100628a646a8077ac2477ad))
* **deps:** update docker/build-push-action action to v6.12.0 ([#698](https://github.com/grafana/shared-workflows/issues/698)) ([3b08e31](https://github.com/grafana/shared-workflows/commit/3b08e3185a075be3d294bb070cf3e9729312b4af))
* **deps:** update docker/build-push-action action to v6.13.0 ([#715](https://github.com/grafana/shared-workflows/issues/715)) ([4b971c2](https://github.com/grafana/shared-workflows/commit/4b971c2583aa388393ad4da89a79b86379fd9197))
* **deps:** update docker/setup-buildx-action action to v3.8.0 ([#654](https://github.com/grafana/shared-workflows/issues/654)) ([d55f5e9](https://github.com/grafana/shared-workflows/commit/d55f5e910f5f76c0b23ba86ef590e2939c475899))
* update readme when a new release is available ([#548](https://github.com/grafana/shared-workflows/issues/548)) ([9bf9163](https://github.com/grafana/shared-workflows/commit/9bf9163126c44247bcee6b6b9390eb488f9ead53))

## 0.1.0 (2024-11-28)


### 🎉 Features

* **build-push-to-gar:** Expose platforms parameter ([#78](https://github.com/grafana/shared-workflows/issues/78)) ([f86c2ca](https://github.com/grafana/shared-workflows/commit/f86c2cae0a68db2803adc0006fe5919483d861dc))
* push-to-gar-docker - add outputs for ease-of-use ([#89](https://github.com/grafana/shared-workflows/issues/89)) ([6e9b07a](https://github.com/grafana/shared-workflows/commit/6e9b07a8ad263b99c027843ec520969c14852d30))
* **push-to-gar-docker:** replace underscores with hyphens in repo names ([#199](https://github.com/grafana/shared-workflows/issues/199)) ([a67842b](https://github.com/grafana/shared-workflows/commit/a67842be4f21319c80f40041d7bc02a26d8722bc))


### 🐛 Bug Fixes

* add repository_name input to push-to-gar-docker ([#198](https://github.com/grafana/shared-workflows/issues/198)) ([264a3f2](https://github.com/grafana/shared-workflows/commit/264a3f2a5d4f756715d5c1f3b37f627689e70ab1))
* Make file argument optional in docker build actions ([#50](https://github.com/grafana/shared-workflows/issues/50)) ([b2c2806](https://github.com/grafana/shared-workflows/commit/b2c2806d455f6cbe4086fb0df849083ef48fd01c))
* **push-gar-doc:** fix typo ([#200](https://github.com/grafana/shared-workflows/issues/200)) ([5d89d95](https://github.com/grafana/shared-workflows/commit/5d89d954c8bc3d7664e576b86bfdbaa1302a1ca5))


### 📝 Documentation

* **push-to-gar-docker:** Fix metadata link ([#159](https://github.com/grafana/shared-workflows/issues/159)) ([1a9e4bc](https://github.com/grafana/shared-workflows/commit/1a9e4bc0ccbb0bff51f47a275e6a93f5509384f3))


### 🔧 Miscellaneous Chores

* Bump upstream docker actions ([#51](https://github.com/grafana/shared-workflows/issues/51)) ([f33ebd9](https://github.com/grafana/shared-workflows/commit/f33ebd946aa2bcd994fb26afdedb575131a5b0b3))
* **deps:** update actions/checkout action to v4.1.7 ([#244](https://github.com/grafana/shared-workflows/issues/244)) ([1d5fba5](https://github.com/grafana/shared-workflows/commit/1d5fba52e7cb2780dfd1af758e1d84e35ce6e8f7))
* **deps:** update actions/checkout action to v4.2.0 ([#313](https://github.com/grafana/shared-workflows/issues/313)) ([ba6268c](https://github.com/grafana/shared-workflows/commit/ba6268c6beef0ab5b461f45eef4cfe1b4e6d6013))
* **deps:** update actions/checkout action to v4.2.1 ([#445](https://github.com/grafana/shared-workflows/issues/445)) ([c72e039](https://github.com/grafana/shared-workflows/commit/c72e039d656ea7db5cbcfd98dffd0f8554e1f029))
* **deps:** update actions/checkout action to v4.2.2 ([#498](https://github.com/grafana/shared-workflows/issues/498)) ([7c6dbe2](https://github.com/grafana/shared-workflows/commit/7c6dbe23c5fd8f3ab5863fb0e3f9d95de621b746))
* **deps:** update docker/build-push-action action to v5.4.0 ([#253](https://github.com/grafana/shared-workflows/issues/253)) ([30f2a90](https://github.com/grafana/shared-workflows/commit/30f2a90675be35c05810244a374dda92ca4cc813))
* **deps:** update docker/build-push-action action to v6 ([#265](https://github.com/grafana/shared-workflows/issues/265)) ([7a88455](https://github.com/grafana/shared-workflows/commit/7a884559706c0b959e39cd82a6baa6c2b771f1a2))
* **deps:** update docker/build-push-action action to v6.10.0 ([#547](https://github.com/grafana/shared-workflows/issues/547)) ([ae58551](https://github.com/grafana/shared-workflows/commit/ae585512b1988ff838ee02c4c2433693701c5d14))
* **deps:** update docker/build-push-action action to v6.8.0 ([#413](https://github.com/grafana/shared-workflows/issues/413)) ([1994696](https://github.com/grafana/shared-workflows/commit/1994696f5a63ba7308496d2bae1d98b29f8965e3))
* **deps:** update docker/build-push-action action to v6.9.0 ([#431](https://github.com/grafana/shared-workflows/issues/431)) ([51c9559](https://github.com/grafana/shared-workflows/commit/51c9559f727b006be385d4383df75212d4eee894))
* **deps:** update docker/metadata-action action to v5.6.0 ([#539](https://github.com/grafana/shared-workflows/issues/539)) ([1bdb3a4](https://github.com/grafana/shared-workflows/commit/1bdb3a48906e610f13acdf4a1990dca485c85497))
* **deps:** update docker/metadata-action action to v5.6.1 ([1bdb3a4](https://github.com/grafana/shared-workflows/commit/1bdb3a48906e610f13acdf4a1990dca485c85497))
* **deps:** update docker/setup-buildx-action action to v3.6.1 ([#257](https://github.com/grafana/shared-workflows/issues/257)) ([46bb727](https://github.com/grafana/shared-workflows/commit/46bb727fff56784c6f157d03e1a77b1ac84636f2))
* **deps:** update docker/setup-buildx-action action to v3.7.0 ([#438](https://github.com/grafana/shared-workflows/issues/438)) ([c546bd5](https://github.com/grafana/shared-workflows/commit/c546bd5895ab8ca039394f7aeca414243c6108c7))
* **deps:** update docker/setup-buildx-action action to v3.7.1 ([#440](https://github.com/grafana/shared-workflows/issues/440)) ([680fa60](https://github.com/grafana/shared-workflows/commit/680fa602301c5650881d920fb094604c6586ac7d))
