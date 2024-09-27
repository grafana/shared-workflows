# Changelog

## [1.1.0](https://github.com/grafana/shared-workflows/compare/v1.0.0...v1.1.0) (2024-09-27)


### Features

* **actions:** Create `aws-auth` composite action ([#67](https://github.com/grafana/shared-workflows/issues/67)) ([49b9885](https://github.com/grafana/shared-workflows/commit/49b9885e467b0544c76602d4e8b8ee342f6ea96b))
* add argo-lint and install-argo-cli action ([#171](https://github.com/grafana/shared-workflows/issues/171)) ([d848da2](https://github.com/grafana/shared-workflows/commit/d848da21d310b2a847a73457059b5a2d93d9f154))
* Add new techdocs-rewrite-relative-links-action ([#99](https://github.com/grafana/shared-workflows/issues/99)) ([93c8404](https://github.com/grafana/shared-workflows/commit/93c84040a318ceb535ed130b9b75c76eb68b0a06))
* add release-please action ([#183](https://github.com/grafana/shared-workflows/issues/183)) ([0c6afbf](https://github.com/grafana/shared-workflows/commit/0c6afbfb9e7f4af01cf3cfed7535eae33943fe46))
* add setup-conftest action ([#212](https://github.com/grafana/shared-workflows/issues/212)) ([6a252ee](https://github.com/grafana/shared-workflows/commit/6a252ee32cc3109533ce51789842d3ed78e6abf2))
* **aws-auth:** add workflow_ref claim ([#227](https://github.com/grafana/shared-workflows/issues/227)) ([c0e3298](https://github.com/grafana/shared-workflows/commit/c0e329819eb62c2cfb5611a56289a2017066b1e7))
* **build-push-to-dockerhub:** Expose platforms parameter ([#37](https://github.com/grafana/shared-workflows/issues/37)) ([bb37651](https://github.com/grafana/shared-workflows/commit/bb376519aa50489c7c5cb51c22830f804b0b176f))
* **build-push-to-gar:** Expose platforms parameter ([#78](https://github.com/grafana/shared-workflows/issues/78)) ([f86c2ca](https://github.com/grafana/shared-workflows/commit/f86c2cae0a68db2803adc0006fe5919483d861dc))
* **deps:** Use Renovate to manage Actions dependncies ([#240](https://github.com/grafana/shared-workflows/issues/240)) ([cd594c3](https://github.com/grafana/shared-workflows/commit/cd594c37d6f39fd9fb74d4abe8055d36a262c677))
* **lint-pr-title:** rewrite in Typescript and handle `merge_group` events ([#233](https://github.com/grafana/shared-workflows/issues/233)) ([82f051e](https://github.com/grafana/shared-workflows/commit/82f051e035ffb6f74dfdc2ce3a0d2eded327b0b0))
* modify openapi spec at client generation time ([#204](https://github.com/grafana/shared-workflows/issues/204)) ([fc84de9](https://github.com/grafana/shared-workflows/commit/fc84de984d84586aaa4c05c88620553d1473f735))
* **pub-techdocs:** add input for running `actions/checkout` ([#84](https://github.com/grafana/shared-workflows/issues/84)) ([d393a41](https://github.com/grafana/shared-workflows/commit/d393a4176d28e9e357a2781cb225603ed839ebbf))
* **pub-techdocs:** checkout submodules ([#128](https://github.com/grafana/shared-workflows/issues/128)) ([e809fd2](https://github.com/grafana/shared-workflows/commit/e809fd2353a58174b5e634e813ce244abfaa52ac))
* **pub-techdocs:** make bucket an input ([#165](https://github.com/grafana/shared-workflows/issues/165)) ([3c8b714](https://github.com/grafana/shared-workflows/commit/3c8b714cda46503c7934a610d78a73b6c02811c0))
* **pub-techdocs:** make checking out `submodules` configurable ([#132](https://github.com/grafana/shared-workflows/issues/132)) ([3ac29a6](https://github.com/grafana/shared-workflows/commit/3ac29a66ab91084d07be10f0bbf35f572cb763f5))
* push-to-gar-docker - add outputs for ease-of-use ([#89](https://github.com/grafana/shared-workflows/issues/89)) ([6e9b07a](https://github.com/grafana/shared-workflows/commit/6e9b07a8ad263b99c027843ec520969c14852d30))
* **push-to-gar-docker:** replace underscores with hyphens in repo names ([#199](https://github.com/grafana/shared-workflows/issues/199)) ([a67842b](https://github.com/grafana/shared-workflows/commit/a67842be4f21319c80f40041d7bc02a26d8722bc))
* **push-to-gcs:** allow setting 'predefinedAcl' on objects when uploading ([#193](https://github.com/grafana/shared-workflows/issues/193)) ([97e6191](https://github.com/grafana/shared-workflows/commit/97e6191605de61d528f08aa85fa2f9ee2dfac355))
* Rename push-to-dockerhub-action ([#33](https://github.com/grafana/shared-workflows/issues/33)) ([6730582](https://github.com/grafana/shared-workflows/commit/673058269d2bc16224e7ee844037a794765e432e))
* **trigger-argo-workflow:** support {dev,ops}-aws instances and get repo-specifc paths ([#269](https://github.com/grafana/shared-workflows/issues/269)) ([968bd76](https://github.com/grafana/shared-workflows/commit/968bd76796b6eccd56f66c713fc0f07bf34824a2))


### Bug Fixes

* add repository_name input to push-to-gar-docker ([#198](https://github.com/grafana/shared-workflows/issues/198)) ([264a3f2](https://github.com/grafana/shared-workflows/commit/264a3f2a5d4f756715d5c1f3b37f627689e70ab1))
* **argo-lint:** fallback to hardcode action repo ([#213](https://github.com/grafana/shared-workflows/issues/213)) ([c663230](https://github.com/grafana/shared-workflows/commit/c6632305ef48112fe6b1aad26ecf2b32a743bda9))
* generate go client in a subdir ([#208](https://github.com/grafana/shared-workflows/issues/208)) ([335e261](https://github.com/grafana/shared-workflows/commit/335e261108a1299ee06227acad2e487118e3110e))
* **lint-shared-workflows:** correct tag for `prettier_action` ([#255](https://github.com/grafana/shared-workflows/issues/255)) ([2586d40](https://github.com/grafana/shared-workflows/commit/2586d409a48b5218db8449ce07c86f3347bf8d05))
* Make file argument optional in docker build actions ([#50](https://github.com/grafana/shared-workflows/issues/50)) ([b2c2806](https://github.com/grafana/shared-workflows/commit/b2c2806d455f6cbe4086fb0df849083ef48fd01c))
* **pub-techdocs:** pin versions to fix regression ([#160](https://github.com/grafana/shared-workflows/issues/160)) ([3d798d5](https://github.com/grafana/shared-workflows/commit/3d798d546fc4aab6ecd4f370fb73ecdda78e3c1c))
* **publish-techdocs:** remove `cache` from `setup python` step name ([#5](https://github.com/grafana/shared-workflows/issues/5)) ([cee0668](https://github.com/grafana/shared-workflows/commit/cee06689c88bf5ab35e7047faacc86f4b47ece09))
* **publish-techdocs:** remove cache from setup-python ([#4](https://github.com/grafana/shared-workflows/issues/4)) ([040c7ef](https://github.com/grafana/shared-workflows/commit/040c7ef79b820444cca5bd940663fefef753b651))
* **push-gar-doc:** fix typo ([#200](https://github.com/grafana/shared-workflows/issues/200)) ([5d89d95](https://github.com/grafana/shared-workflows/commit/5d89d954c8bc3d7664e576b86bfdbaa1302a1ca5))
* **renovate:** fix action versions, add more triggers ([#249](https://github.com/grafana/shared-workflows/issues/249)) ([3b1e65f](https://github.com/grafana/shared-workflows/commit/3b1e65f1b3b563a309b4aa5f888d916ad389cec3))
* **renovate:** fix dry-run mode for manual trigger ([#252](https://github.com/grafana/shared-workflows/issues/252)) ([c1dacd0](https://github.com/grafana/shared-workflows/commit/c1dacd07e56da089d92489e0ac81147a24da7544))
* **setup-argo:** less fragile OS/ARCH selection ([#222](https://github.com/grafana/shared-workflows/issues/222)) ([e9775f3](https://github.com/grafana/shared-workflows/commit/e9775f3ace2ef954b81548720476fb42ebde52e8))
* **setup-argo:** use amd64 as arch ([#217](https://github.com/grafana/shared-workflows/issues/217)) ([7a0ba14](https://github.com/grafana/shared-workflows/commit/7a0ba14ec0596297d38441c7829cbe8eb30fb036))
* **setup-argo:** use bash, fix idempotency, only restore if needed ([#221](https://github.com/grafana/shared-workflows/issues/221)) ([1e75822](https://github.com/grafana/shared-workflows/commit/1e75822620b1413e97deb7d60b10cad9ebf0fdeb))
* **setup-conftest:** normalize runner.os ([#216](https://github.com/grafana/shared-workflows/issues/216)) ([318d33e](https://github.com/grafana/shared-workflows/commit/318d33e1f443b2f511a21593f1945e9e026c86d0))
* **trigger-argo-workflow:** use setup-argo action ([#219](https://github.com/grafana/shared-workflows/issues/219)) ([47a1c7f](https://github.com/grafana/shared-workflows/commit/47a1c7f387daf4ef593b82cb6ac2abca0cd7cf73))

## 1.0.0 (2024-07-15)

### Features

- **actions:** Create `aws-auth` composite action ([#67](https://github.com/grafana/shared-workflows/issues/67)) ([49b9885](https://github.com/grafana/shared-workflows/commit/49b9885e467b0544c76602d4e8b8ee342f6ea96b))
- add argo-lint and install-argo-cli action ([#171](https://github.com/grafana/shared-workflows/issues/171)) ([d848da2](https://github.com/grafana/shared-workflows/commit/d848da21d310b2a847a73457059b5a2d93d9f154))
- Add new techdocs-rewrite-relative-links-action ([#99](https://github.com/grafana/shared-workflows/issues/99)) ([93c8404](https://github.com/grafana/shared-workflows/commit/93c84040a318ceb535ed130b9b75c76eb68b0a06))
- **build-push-to-dockerhub:** Expose platforms parameter ([#37](https://github.com/grafana/shared-workflows/issues/37)) ([bb37651](https://github.com/grafana/shared-workflows/commit/bb376519aa50489c7c5cb51c22830f804b0b176f))
- **build-push-to-gar:** Expose platforms parameter ([#78](https://github.com/grafana/shared-workflows/issues/78)) ([f86c2ca](https://github.com/grafana/shared-workflows/commit/f86c2cae0a68db2803adc0006fe5919483d861dc))
- **pub-techdocs:** add input for running `actions/checkout` ([#84](https://github.com/grafana/shared-workflows/issues/84)) ([d393a41](https://github.com/grafana/shared-workflows/commit/d393a4176d28e9e357a2781cb225603ed839ebbf))
- **pub-techdocs:** checkout submodules ([#128](https://github.com/grafana/shared-workflows/issues/128)) ([e809fd2](https://github.com/grafana/shared-workflows/commit/e809fd2353a58174b5e634e813ce244abfaa52ac))
- **pub-techdocs:** make bucket an input ([#165](https://github.com/grafana/shared-workflows/issues/165)) ([3c8b714](https://github.com/grafana/shared-workflows/commit/3c8b714cda46503c7934a610d78a73b6c02811c0))
- **pub-techdocs:** make checking out `submodules` configurable ([#132](https://github.com/grafana/shared-workflows/issues/132)) ([3ac29a6](https://github.com/grafana/shared-workflows/commit/3ac29a66ab91084d07be10f0bbf35f572cb763f5))
- push-to-gar-docker - add outputs for ease-of-use ([#89](https://github.com/grafana/shared-workflows/issues/89)) ([6e9b07a](https://github.com/grafana/shared-workflows/commit/6e9b07a8ad263b99c027843ec520969c14852d30))
- Rename push-to-dockerhub-action ([#33](https://github.com/grafana/shared-workflows/issues/33)) ([6730582](https://github.com/grafana/shared-workflows/commit/673058269d2bc16224e7ee844037a794765e432e))

### Bug Fixes

- Make file argument optional in docker build actions ([#50](https://github.com/grafana/shared-workflows/issues/50)) ([b2c2806](https://github.com/grafana/shared-workflows/commit/b2c2806d455f6cbe4086fb0df849083ef48fd01c))
- **pub-techdocs:** pin versions to fix regression ([#160](https://github.com/grafana/shared-workflows/issues/160)) ([3d798d5](https://github.com/grafana/shared-workflows/commit/3d798d546fc4aab6ecd4f370fb73ecdda78e3c1c))
- **publish-techdocs:** remove `cache` from `setup python` step name ([#5](https://github.com/grafana/shared-workflows/issues/5)) ([cee0668](https://github.com/grafana/shared-workflows/commit/cee06689c88bf5ab35e7047faacc86f4b47ece09))
- **publish-techdocs:** remove cache from setup-python ([#4](https://github.com/grafana/shared-workflows/issues/4)) ([040c7ef](https://github.com/grafana/shared-workflows/commit/040c7ef79b820444cca5bd940663fefef753b651))
