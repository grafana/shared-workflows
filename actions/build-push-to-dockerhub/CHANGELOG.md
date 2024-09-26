# Changelog

## 1.0.0 (2024-09-26)


### 🎉 Features

* **build-push-to-dockerhub:** Expose platforms parameter ([#37](https://github.com/grafana/shared-workflows/issues/37)) ([bb37651](https://github.com/grafana/shared-workflows/commit/bb376519aa50489c7c5cb51c22830f804b0b176f))
* Rename push-to-dockerhub-action ([#33](https://github.com/grafana/shared-workflows/issues/33)) ([6730582](https://github.com/grafana/shared-workflows/commit/673058269d2bc16224e7ee844037a794765e432e))


### 🐛 Bug Fixes

* Make file argument optional in docker build actions ([#50](https://github.com/grafana/shared-workflows/issues/50)) ([b2c2806](https://github.com/grafana/shared-workflows/commit/b2c2806d455f6cbe4086fb0df849083ef48fd01c))


### 🔧 Miscellaneous Chores

* Bump upstream docker actions ([#51](https://github.com/grafana/shared-workflows/issues/51)) ([f33ebd9](https://github.com/grafana/shared-workflows/commit/f33ebd946aa2bcd994fb26afdedb575131a5b0b3))
* **deps:** update actions/checkout action to v4.1.7 ([#244](https://github.com/grafana/shared-workflows/issues/244)) ([1d5fba5](https://github.com/grafana/shared-workflows/commit/1d5fba52e7cb2780dfd1af758e1d84e35ce6e8f7))
* **deps:** update docker/build-push-action action to v5.4.0 ([#253](https://github.com/grafana/shared-workflows/issues/253)) ([30f2a90](https://github.com/grafana/shared-workflows/commit/30f2a90675be35c05810244a374dda92ca4cc813))
* **deps:** update docker/build-push-action action to v6 ([#265](https://github.com/grafana/shared-workflows/issues/265)) ([7a88455](https://github.com/grafana/shared-workflows/commit/7a884559706c0b959e39cd82a6baa6c2b771f1a2))
* **deps:** update docker/setup-buildx-action action to v3.6.1 ([#257](https://github.com/grafana/shared-workflows/issues/257)) ([46bb727](https://github.com/grafana/shared-workflows/commit/46bb727fff56784c6f157d03e1a77b1ac84636f2))
* **deps:** update docker/setup-qemu-action action to v3.2.0 ([#258](https://github.com/grafana/shared-workflows/issues/258)) ([9d623d7](https://github.com/grafana/shared-workflows/commit/9d623d79425ca7088f7570b4a5862847950a5425))
* Skip log in to DockerHub when not publishing, fail if trying to publish on anything but `push` events ([#126](https://github.com/grafana/shared-workflows/issues/126)) ([0f7721c](https://github.com/grafana/shared-workflows/commit/0f7721c56e0cc8b8b1dcfd17a44808aca4a9cc96))
