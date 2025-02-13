# Changelog

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
