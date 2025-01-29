# Changelog

## [1.1.0](https://github.com/grafana/shared-workflows/compare/trigger-argo-workflow-v1.0.0...trigger-argo-workflow-v1.1.0) (2025-01-29)


### 🎉 Features

* **docs:** added EngHub doc links to corresponding actions readmes ([#635](https://github.com/grafana/shared-workflows/issues/635)) ([a7d04c1](https://github.com/grafana/shared-workflows/commit/a7d04c1e98496dbf07f8e44602933af07ba62f9f))
* print workflow uri immediately even when using --wait flag ([#663](https://github.com/grafana/shared-workflows/issues/663)) ([32bb517](https://github.com/grafana/shared-workflows/commit/32bb517d371b3f8349345cc16365e859be76c323))


### 🐛 Bug Fixes

* **trigger-argo-workflow:** prevent unnecessary retries if a permanent error is encountered ([#631](https://github.com/grafana/shared-workflows/issues/631)) ([81c3771](https://github.com/grafana/shared-workflows/commit/81c377191b9f604bc5f2c64cc2258dfe4bc5ea9c))


### 🏗️ Build System

* **deps:** bump github.com/lmittmann/tint to 1.0.6 in trigger-argo-workflow ([#665](https://github.com/grafana/shared-workflows/issues/665)) ([a094a39](https://github.com/grafana/shared-workflows/commit/a094a395da63897275978d860fb1c79dc45d8895))
* **deps:** bump golang.org/x/term from 0.26.0 to 0.27.0 in /actions/trigger-argo-workflow ([#639](https://github.com/grafana/shared-workflows/issues/639)) ([648f46e](https://github.com/grafana/shared-workflows/commit/648f46efa76a0370d1e0f25c8b81c2f4c7214f0c))
* **deps:** bump golang.org/x/term from 0.27.0 to 0.28.0 in /actions/trigger-argo-workflow ([#673](https://github.com/grafana/shared-workflows/issues/673)) ([ee7ca4e](https://github.com/grafana/shared-workflows/commit/ee7ca4ed19ef4f64d0a42a22685a83666da5a99f))


### 🔧 Miscellaneous Chores

* **deps:** update actions/setup-go action to v5.2.0 ([#646](https://github.com/grafana/shared-workflows/issues/646)) ([bf4b9d4](https://github.com/grafana/shared-workflows/commit/bf4b9d4275d219cda56ae308981df427575b880e))
* **deps:** update actions/setup-go action to v5.3.0 ([#707](https://github.com/grafana/shared-workflows/issues/707)) ([42df8ce](https://github.com/grafana/shared-workflows/commit/42df8cefcbb9c0a25cf060c7566c96eab5d5de69))

## 1.0.0 (2024-11-29)


### 🎉 Features

* add argo-lint and install-argo-cli action ([#171](https://github.com/grafana/shared-workflows/issues/171)) ([d848da2](https://github.com/grafana/shared-workflows/commit/d848da21d310b2a847a73457059b5a2d93d9f154))
* add workflow to check for non-releasable actions ([#588](https://github.com/grafana/shared-workflows/issues/588)) ([e16bf1a](https://github.com/grafana/shared-workflows/commit/e16bf1ac180d7b6c9c13a6e556b24e0f7dc0d57c))
* **trigger-argo-workflow:** support {dev,ops}-aws instances and get repo-specifc paths ([#269](https://github.com/grafana/shared-workflows/issues/269)) ([968bd76](https://github.com/grafana/shared-workflows/commit/968bd76796b6eccd56f66c713fc0f07bf34824a2))


### 🐛 Bug Fixes

* **trigger-argo-workflow:** treat AWS dev instance as Argo Workflows dev ([#465](https://github.com/grafana/shared-workflows/issues/465)) ([a2b807c](https://github.com/grafana/shared-workflows/commit/a2b807c8fdb4be6f2a8236578ab904ad6f0f072e))
* **trigger-argo-workflow:** treat AWS ops instance as Argo Workflows ops ([#504](https://github.com/grafana/shared-workflows/issues/504)) ([0991901](https://github.com/grafana/shared-workflows/commit/099190181e72dac02e346c9167166410b58bcc6f))
* **trigger-argo-workflow:** use `argo-workflows-dev-aws.grafana.net` instead for dev-aws ([#447](https://github.com/grafana/shared-workflows/issues/447)) ([2ae6644](https://github.com/grafana/shared-workflows/commit/2ae66445c4d18cb653f5236f14e7f9d28ce64a99))
* **trigger-argo-workflow:** use correct url for ops on AWS ([#444](https://github.com/grafana/shared-workflows/issues/444)) ([afeb9d6](https://github.com/grafana/shared-workflows/commit/afeb9d6495057ef0046dc76a5fd97202d746b5e3))
* **trigger-argo-workflow:** use setup-argo action ([#219](https://github.com/grafana/shared-workflows/issues/219)) ([47a1c7f](https://github.com/grafana/shared-workflows/commit/47a1c7f387daf4ef593b82cb6ac2abca0cd7cf73))


### 🏗️ Build System

* **deps:** bump github.com/stretchr/testify to 1.10.0 in /actions/trigger-argo-workflow ([#543](https://github.com/grafana/shared-workflows/issues/543)) ([31233e0](https://github.com/grafana/shared-workflows/commit/31233e0888680aac0606ca9999345ae71830149b))
* **deps:** bump github.com/urfave/cli/v2 ([#202](https://github.com/grafana/shared-workflows/issues/202)) ([8eb6b11](https://github.com/grafana/shared-workflows/commit/8eb6b118d95f7098645f3bd9be7b5c0ff69e60a7))
* **deps:** bump github.com/urfave/cli/v2 ([#461](https://github.com/grafana/shared-workflows/issues/461)) ([13cc39c](https://github.com/grafana/shared-workflows/commit/13cc39c275a7c0c6c791b73dbe2d56e6b953a20c))
* **deps:** bump golang.org/x/term ([#237](https://github.com/grafana/shared-workflows/issues/237)) ([8752a98](https://github.com/grafana/shared-workflows/commit/8752a983ed0c01b7ca7d93ee2b245d51212610a0))
* **deps:** bump golang.org/x/term ([#443](https://github.com/grafana/shared-workflows/issues/443)) ([08e0541](https://github.com/grafana/shared-workflows/commit/08e05415ed9f52fbe19b7ba9365bc24b7474631a))
* **deps:** bump golang.org/x/term from 0.25.0 to 0.26.0 in /actions/trigger-argo-workflow ([#526](https://github.com/grafana/shared-workflows/issues/526)) ([597cb17](https://github.com/grafana/shared-workflows/commit/597cb17fd3131ad57abd41a46b0bc0febcfa12e5))
* **deps:** bump the go group ([#224](https://github.com/grafana/shared-workflows/issues/224)) ([3075c51](https://github.com/grafana/shared-workflows/commit/3075c5147e45a81e60f0c4f39b50307524e3fff2))
* **deps:** bump the go group across 1 directory with 2 updates ([#178](https://github.com/grafana/shared-workflows/issues/178)) ([f88c6e2](https://github.com/grafana/shared-workflows/commit/f88c6e250f169b0123f90052844f633f0e7df081))
* **deps:** bump the go group in /actions/trigger-argo-workflow with 5 updates ([#150](https://github.com/grafana/shared-workflows/issues/150)) ([6058627](https://github.com/grafana/shared-workflows/commit/60586273f16369c4abd4e626de271785c3e87401))


### 🔧 Miscellaneous Chores

* **deps:** update actions/checkout action to v4.1.7 ([#244](https://github.com/grafana/shared-workflows/issues/244)) ([1d5fba5](https://github.com/grafana/shared-workflows/commit/1d5fba52e7cb2780dfd1af758e1d84e35ce6e8f7))
* **deps:** update actions/checkout action to v4.2.0 ([#313](https://github.com/grafana/shared-workflows/issues/313)) ([ba6268c](https://github.com/grafana/shared-workflows/commit/ba6268c6beef0ab5b461f45eef4cfe1b4e6d6013))
* **deps:** update actions/checkout action to v4.2.1 ([#445](https://github.com/grafana/shared-workflows/issues/445)) ([c72e039](https://github.com/grafana/shared-workflows/commit/c72e039d656ea7db5cbcfd98dffd0f8554e1f029))
* **deps:** update actions/checkout action to v4.2.2 ([#498](https://github.com/grafana/shared-workflows/issues/498)) ([7c6dbe2](https://github.com/grafana/shared-workflows/commit/7c6dbe23c5fd8f3ab5863fb0e3f9d95de621b746))
* **deps:** update actions/setup-go action to v5.0.2 ([#245](https://github.com/grafana/shared-workflows/issues/245)) ([47c75fd](https://github.com/grafana/shared-workflows/commit/47c75fd2f3c1bb6d1a1b7e21c3dabbb24081f56d))
* **deps:** update actions/setup-go action to v5.1.0 ([#501](https://github.com/grafana/shared-workflows/issues/501)) ([afcd2c5](https://github.com/grafana/shared-workflows/commit/afcd2c517a07f844b271fa82982f96ed436216d2))
* Update action dependencies ([#39](https://github.com/grafana/shared-workflows/issues/39)) ([b271a8b](https://github.com/grafana/shared-workflows/commit/b271a8b01e61d00dc987dbb77744bd9e01fe862d))
