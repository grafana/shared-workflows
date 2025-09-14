## Input sections

# GAR CONFIGS

# The GAR image to push is configured as follows:

# ${gar-registry}/${gar-project}/${gar-repository}/${gar-image}

# Ex: us-docker.pkg.dev/grafanalabs-dev/docker-cicd-test-dev/cicd-test

#

# Note that gar-project is determined by the value of gar-environment.

# DOCKERHUB CONFIGS

# REGISTRY/IMAGE CONFIG

dockerhub-repository:
description: |
Ipsum dockerhubium
default: "${{ github.repository }}"
gar-registry:
description: |
Google Artifact Registry to store docker images in.
default: "us-docker.pkg.dev"
gar-repository:
description: |
Override the 'repo_name' used to construct the GAR repository name.
Only necessary when the GAR includes a repo name that doesn't match the GitHub repo name.
Default: `docker-${GitHub Repo Name}-${gar-environment}`
default: ""
gar-environment:
description: |
Environment for pushing artifacts (can be either dev or prod).
default: "dev"
gar-image:
description: |
Name of the image to build.
Default: `${GitHub Repo Name}`
default: ""
registries:
description: |
List of registries to build images for.
default: ""

# AUTH

delete-credentials-file: #TODO: Expand on this
description: |
Delete the credentials file after the action is finished.
If you want to keep the credentials file for a later step, set this to false.
default: "true"

# DOCKER/METADATA-ACTION

tags:
description: |
List of Docker tags to be pushed.
required: true

# DOCKER/SETUP-BUILDX-ACTION

docker-buildx-driver:
description: |
The driver to use for Docker Buildx
required: false
default: "docker-container"
buildkitd-config:
description: |
Configum buildkitium descriptium
default: ""
buildkitd-config-inline:
description: |
Inliumium configutorium buildkitium descriptium
default: ""

# DOCKER/BUILD-PUSH-ACTION

build-args:
description: |
List of arguments necessary for the Docker image to be built.
default: ""
build-contexts:
description: |
List of additional build contexts (e.g., name=path)
required: false
cache-from:
description: |
Where cache should be fetched from
required: false
default: "type=gha"
cache-to:
description: |
Where cache should be stored to
required: false
default: "type=gha,mode=max"
context:
description: |
Path to the Docker build context.
default: "."
file:
description: |
The dockerfile to use.
required: false
labels:
description: |
List of custom labels to add to the image as metadata.
required: false
load:
description: |
Whether to load the built image into the local docker daemon.
required: false
default: "false"
outputs:
description: | # TODO Desc
Ipsum factum explainum.
required: false
default: ""
platforms:
description: |
List of platforms to build the image for
required: false
push:
description: |
Whether to push the image to the configured registries.
required: false
secrets:
description: |
Secrets to expose to the build. Only needed when authenticating to private repositories outside the repository in which the image is being built.
required: false
ssh:
description: |
List of SSH agent socket or keys to expose to the build
target:
description: |
Sets the target stage to build
required: false
