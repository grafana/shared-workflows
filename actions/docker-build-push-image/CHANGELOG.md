# Changelog

## [0.3.1](https://github.com/grafana/shared-workflows/compare/docker-build-push-image/v0.3.0...docker-build-push-image/v0.3.1) (2026-02-18)


### 📝 Documentation

* **docker-build-push-image:** clarify push input defaults ([#1719](https://github.com/grafana/shared-workflows/issues/1719)) ([256dfbb](https://github.com/grafana/shared-workflows/commit/256dfbbabfebc3e21d15113e9f5e876ba2fe7b48))


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v6.0.2 ([#1672](https://github.com/grafana/shared-workflows/issues/1672)) ([3105e25](https://github.com/grafana/shared-workflows/commit/3105e251e687194e9b2b4b456cb2846a761e0df0))
* **deps:** update docker/build-push-action action to v6.19.0 ([#1717](https://github.com/grafana/shared-workflows/issues/1717)) ([37c176c](https://github.com/grafana/shared-workflows/commit/37c176c8a62d075692cbafeacbc4ea31613a33c5))
* **deps:** update docker/build-push-action action to v6.19.2 ([#1721](https://github.com/grafana/shared-workflows/issues/1721)) ([bf35378](https://github.com/grafana/shared-workflows/commit/bf35378348bf907badcc18ddf782201261c62d8c))
* **deps:** update docker/setup-buildx-action action to v3.12.0 ([#1633](https://github.com/grafana/shared-workflows/issues/1633)) ([9c7001f](https://github.com/grafana/shared-workflows/commit/9c7001fb8ab6048113c07b6454aea78589e0e6b9))

## [0.3.0](https://github.com/grafana/shared-workflows/compare/docker-build-push-image/v0.2.0...docker-build-push-image/v0.3.0) (2025-12-11)


### 🎉 Features

* **docker-build-push-image:** add builder input ([#1605](https://github.com/grafana/shared-workflows/issues/1605)) ([e39baa6](https://github.com/grafana/shared-workflows/commit/e39baa6584886aef49fe533215defc5627ffbd3f))


### 🐛 Bug Fixes

* **docker-build-push-image:** pass builder name to docker/build-push-action ([#1602](https://github.com/grafana/shared-workflows/issues/1602)) ([59bb68c](https://github.com/grafana/shared-workflows/commit/59bb68c6a0a1bc5701261915eaa04199916be716))


### 📝 Documentation

* **multiple:** add warning about push to GAR failure ([#1555](https://github.com/grafana/shared-workflows/issues/1555)) ([eb33f84](https://github.com/grafana/shared-workflows/commit/eb33f84481d38701f4d2c587a4817ce332784f5f))


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v5.0.1 ([#1541](https://github.com/grafana/shared-workflows/issues/1541)) ([773f5b1](https://github.com/grafana/shared-workflows/commit/773f5b1eb7b717c5c89a2718c1c4322a45f2ed7f))
* **deps:** update actions/checkout action to v6 ([#1570](https://github.com/grafana/shared-workflows/issues/1570)) ([af4d9df](https://github.com/grafana/shared-workflows/commit/af4d9dfcfa9da2582544cd2a6e6dcf06e516f9ea))
* **deps:** update actions/checkout action to v6.0.1 ([#1590](https://github.com/grafana/shared-workflows/issues/1590)) ([2425a5f](https://github.com/grafana/shared-workflows/commit/2425a5fe46fb39d1d282caad59150165323e29a6))
* **deps:** update docker/metadata-action action to v5.10.0 ([#1582](https://github.com/grafana/shared-workflows/issues/1582)) ([d80ddba](https://github.com/grafana/shared-workflows/commit/d80ddba3b588ad911410ce91c599bbbe513196b0))
* **multiple:** deprecate old docker actions and add migration guide ([#1606](https://github.com/grafana/shared-workflows/issues/1606)) ([b6c252d](https://github.com/grafana/shared-workflows/commit/b6c252dc86cb65eaf2d8344d6d51ca07436214a2))

## [0.2.0](https://github.com/grafana/shared-workflows/compare/docker-build-push-image/v0.1.1...docker-build-push-image/v0.2.0) (2025-11-11)


### 🎉 Features

* **docker-build-push-image:** support annotations ([#1513](https://github.com/grafana/shared-workflows/issues/1513)) ([62333c1](https://github.com/grafana/shared-workflows/commit/62333c16b0e89cea15c9c4de726271fa4e638f96))


### 📝 Documentation

* improve docker build action docs ([#1486](https://github.com/grafana/shared-workflows/issues/1486)) ([2dd0b03](https://github.com/grafana/shared-workflows/commit/2dd0b0349e130ca5ccf86b3a61250589a840bdb2))


### 🔧 Miscellaneous Chores

* **deps:** update docker/metadata-action action to v5.9.0 ([#1501](https://github.com/grafana/shared-workflows/issues/1501)) ([2d5a067](https://github.com/grafana/shared-workflows/commit/2d5a0678eb32b0fd6655b4f7a3a7ec72eaf530ca))
* **deps:** update docker/setup-qemu-action action to v3.7.0 ([#1505](https://github.com/grafana/shared-workflows/issues/1505)) ([559f4cd](https://github.com/grafana/shared-workflows/commit/559f4cd66a01e0ebe256defa20964721f25bad1f))

## [0.1.1](https://github.com/grafana/shared-workflows/compare/docker-build-push-image/v0.1.0...docker-build-push-image/v0.1.1) (2025-11-07)


### 🐛 Bug Fixes

* **docker-build-push-image:** use labels from docker/metadata-action ([#1502](https://github.com/grafana/shared-workflows/issues/1502)) ([4d65325](https://github.com/grafana/shared-workflows/commit/4d65325ccd5a45fae20b0821f4ab15e2451d92ea))

## 0.1.0 (2025-10-15)


### 🎉 Features

* docker multi-arch composite actions ([#1347](https://github.com/grafana/shared-workflows/issues/1347)) ([3df0c01](https://github.com/grafana/shared-workflows/commit/3df0c015c8b528638bdbbccc6326b7c2edc79ae1))


### 🔧 Miscellaneous Chores

* **docker actions:** prep for releases ([#1404](https://github.com/grafana/shared-workflows/issues/1404)) ([b5b2517](https://github.com/grafana/shared-workflows/commit/b5b25178a74fbef4cc2db2252ac322729ab76e91))

## Changelog
