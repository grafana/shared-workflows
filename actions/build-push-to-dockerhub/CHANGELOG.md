# Changelog

## [0.3.0](https://github.com/grafana/shared-workflows/compare/build-push-to-dockerhub/v0.2.0...build-push-to-dockerhub/v0.3.0) (2025-10-03)


### 🎉 Features

* **build-push-to-dockerhub:** support load parameter ([#1193](https://github.com/grafana/shared-workflows/issues/1193)) ([58d9866](https://github.com/grafana/shared-workflows/commit/58d9866681bf38ddb7567d1283659c60149b4c99))


### 🐛 Bug Fixes

* remediate latest zizmor findings, fix supplying zizmor config ([#1101](https://github.com/grafana/shared-workflows/issues/1101)) ([712c599](https://github.com/grafana/shared-workflows/commit/712c59975bc0de22124b866153826f04023f18fd))


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v4.3.0 ([#1221](https://github.com/grafana/shared-workflows/issues/1221)) ([17ab531](https://github.com/grafana/shared-workflows/commit/17ab531bf2c16c79af38988e7caf7a3d8a37634b))
* **deps:** update actions/checkout action to v5 ([#1227](https://github.com/grafana/shared-workflows/issues/1227)) ([fd79c02](https://github.com/grafana/shared-workflows/commit/fd79c02730e0629f728e2f5c3d614545269208a9))
* **deps:** update docker/build-push-action action to v6.18.0 ([#1065](https://github.com/grafana/shared-workflows/issues/1065)) ([5b5ee4c](https://github.com/grafana/shared-workflows/commit/5b5ee4cf0a527daf5e32b7f968637b8a8ed7efcb))
* **deps:** update docker/metadata-action action to v5.8.0 ([#1182](https://github.com/grafana/shared-workflows/issues/1182)) ([315e38a](https://github.com/grafana/shared-workflows/commit/315e38a03f442c39bd82e902b88d8ba6ff8879b7))
* **deps:** update docker/setup-buildx-action action to v3.11.1 ([#1068](https://github.com/grafana/shared-workflows/issues/1068)) ([5233cbc](https://github.com/grafana/shared-workflows/commit/5233cbc5d62242fb17b2259c2c4bd2a628af5528))

## [0.2.0](https://github.com/grafana/shared-workflows/compare/build-push-to-dockerhub-v0.1.1...build-push-to-dockerhub/v0.2.0) (2025-06-04)


### 🎉 Features

* **build-push-to-dockerhub:** enable docker mirror for buildx on self-hosted runners ([#999](https://github.com/grafana/shared-workflows/issues/999)) ([f797fbd](https://github.com/grafana/shared-workflows/commit/f797fbd07354fd4727f952291bfa6b85eab568ef))
* **push-to-gar-docker,build-push-to-dockerhub:** add support for buildkit secrets ([#990](https://github.com/grafana/shared-workflows/issues/990)) ([bfed586](https://github.com/grafana/shared-workflows/commit/bfed586d71f4799f2506878776b481d00ca84bda))


### 🐛 Bug Fixes

* ensure every action disables git credential persistence ([#821](https://github.com/grafana/shared-workflows/issues/821)) ([31ebf3f](https://github.com/grafana/shared-workflows/commit/31ebf3f8e5d0f8709e6ec4ef73b39dd2bd08f959))


### 📝 Documentation

* **multiple-actions:** move permissions to job level in workflow examples ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))
* **multiple-actions:** move permissions to job level in workflows ([#969](https://github.com/grafana/shared-workflows/issues/969)) ([49c90b1](https://github.com/grafana/shared-workflows/commit/49c90b10fcbce463983bed45932cf468b8bd06ce))
* **multiple:** add docker cache notes to build-push-to-dockerhub and push-to-gar-docker ([#1003](https://github.com/grafana/shared-workflows/issues/1003)) ([e5377f9](https://github.com/grafana/shared-workflows/commit/e5377f9c2aee143ccf63001896fa59eef7bea1d5))
* update all readmes to replace hyphen with slash ([#1008](https://github.com/grafana/shared-workflows/issues/1008)) ([472df76](https://github.com/grafana/shared-workflows/commit/472df76fb1cbb92a17fb9e055bdf0d1399109ee3))


### 🔧 Miscellaneous Chores

* **deps:** update docker/build-push-action action to v6.14.0 ([#793](https://github.com/grafana/shared-workflows/issues/793)) ([d750654](https://github.com/grafana/shared-workflows/commit/d750654d770aefa0516e11735cdfdb89b7a380a1))
* **deps:** update docker/build-push-action action to v6.15.0 ([#816](https://github.com/grafana/shared-workflows/issues/816)) ([0ae253d](https://github.com/grafana/shared-workflows/commit/0ae253d4a198408407a161de482680eddf2dfa42))
* **deps:** update docker/build-push-action action to v6.16.0 ([#923](https://github.com/grafana/shared-workflows/issues/923)) ([a301072](https://github.com/grafana/shared-workflows/commit/a30107276148b4f29eaeaef05a3f9173d1aa0ad9))
* **deps:** update docker/metadata-action action to v5.7.0 ([#818](https://github.com/grafana/shared-workflows/issues/818)) ([9f9b2eb](https://github.com/grafana/shared-workflows/commit/9f9b2eb3897a39fd65e5b92f17a60704925f94c4))
* **deps:** update docker/setup-buildx-action action to v3.10.0 ([#819](https://github.com/grafana/shared-workflows/issues/819)) ([09fb633](https://github.com/grafana/shared-workflows/commit/09fb633eb9f6c77153fa941e662be7cd418ca1fb))
* **deps:** update docker/setup-buildx-action action to v3.9.0 ([#755](https://github.com/grafana/shared-workflows/issues/755)) ([8dd62e3](https://github.com/grafana/shared-workflows/commit/8dd62e320f60df7426d30b67c9b26f17af352ed7))
* **deps:** update docker/setup-qemu-action action to v3.4.0 ([#756](https://github.com/grafana/shared-workflows/issues/756)) ([753c87a](https://github.com/grafana/shared-workflows/commit/753c87a0ea97496f0088e51c025e1f4c69be6626))
* **deps:** update docker/setup-qemu-action action to v3.5.0 ([#820](https://github.com/grafana/shared-workflows/issues/820)) ([183a929](https://github.com/grafana/shared-workflows/commit/183a929dfee60c6294552ac80371153c29860c16))
* **deps:** update docker/setup-qemu-action action to v3.6.0 ([#825](https://github.com/grafana/shared-workflows/issues/825)) ([bfbcd01](https://github.com/grafana/shared-workflows/commit/bfbcd01788fe3d09fb1de307529afe2c111cbc64))
* **main:** release push-to-gar-docker 0.3.0 ([#794](https://github.com/grafana/shared-workflows/issues/794)) ([a7bc536](https://github.com/grafana/shared-workflows/commit/a7bc5367c4a91c389526d58839d8f6224dba4dcc))

## [0.1.1](https://github.com/grafana/shared-workflows/compare/build-push-to-dockerhub-v0.1.0...build-push-to-dockerhub-v0.1.1) (2025-01-29)


### 🔧 Miscellaneous Chores

* **deps:** update docker/build-push-action action to v6.11.0 ([#679](https://github.com/grafana/shared-workflows/issues/679)) ([e1b07ec](https://github.com/grafana/shared-workflows/commit/e1b07ec29d283a54c100628a646a8077ac2477ad))
* **deps:** update docker/build-push-action action to v6.12.0 ([#698](https://github.com/grafana/shared-workflows/issues/698)) ([3b08e31](https://github.com/grafana/shared-workflows/commit/3b08e3185a075be3d294bb070cf3e9729312b4af))
* **deps:** update docker/build-push-action action to v6.13.0 ([#715](https://github.com/grafana/shared-workflows/issues/715)) ([4b971c2](https://github.com/grafana/shared-workflows/commit/4b971c2583aa388393ad4da89a79b86379fd9197))
* **deps:** update docker/setup-buildx-action action to v3.8.0 ([#654](https://github.com/grafana/shared-workflows/issues/654)) ([d55f5e9](https://github.com/grafana/shared-workflows/commit/d55f5e910f5f76c0b23ba86ef590e2939c475899))
* **deps:** update docker/setup-qemu-action action to v3.3.0 ([#680](https://github.com/grafana/shared-workflows/issues/680)) ([1f47ea2](https://github.com/grafana/shared-workflows/commit/1f47ea2687b3eb8188f4c00dbdb2658cb6eb3321))
* update readme when a new release is available ([#548](https://github.com/grafana/shared-workflows/issues/548)) ([9bf9163](https://github.com/grafana/shared-workflows/commit/9bf9163126c44247bcee6b6b9390eb488f9ead53))

## 0.1.0 (2024-11-28)


### 🎉 Features

* **build-push-to-dockerhub:** Expose platforms parameter ([#37](https://github.com/grafana/shared-workflows/issues/37)) ([bb37651](https://github.com/grafana/shared-workflows/commit/bb376519aa50489c7c5cb51c22830f804b0b176f))
* Rename push-to-dockerhub-action ([#33](https://github.com/grafana/shared-workflows/issues/33)) ([6730582](https://github.com/grafana/shared-workflows/commit/673058269d2bc16224e7ee844037a794765e432e))


### 🐛 Bug Fixes

* Make file argument optional in docker build actions ([#50](https://github.com/grafana/shared-workflows/issues/50)) ([b2c2806](https://github.com/grafana/shared-workflows/commit/b2c2806d455f6cbe4086fb0df849083ef48fd01c))


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
* **deps:** update docker/setup-qemu-action action to v3.2.0 ([#258](https://github.com/grafana/shared-workflows/issues/258)) ([9d623d7](https://github.com/grafana/shared-workflows/commit/9d623d79425ca7088f7570b4a5862847950a5425))
* Skip log in to DockerHub when not publishing, fail if trying to publish on anything but `push` events ([#126](https://github.com/grafana/shared-workflows/issues/126)) ([0f7721c](https://github.com/grafana/shared-workflows/commit/0f7721c56e0cc8b8b1dcfd17a44808aca4a9cc96))
