Did you ever find yourself creating automated builds by hand on http://hub.docker.com for each new git tag?
Why is it called *automated* build, if you can’t create a new one automatically? This project helps to
do the missing step.

There are plenty of articles describing a process where a central continuous build server (jenkins/travis/circleCI) builds
the image locally, and than **pushes** it to DockerHub. The end result of such a process **is not an automated build**. 
Docker images which aren’t automated builds, considered a security risk and therefore should be voided.

Please note that the main purpose of this tool is: to be able to create a new **automated** DockerHub build from cli,
so it can be built easily integrated into any CI server/process.

## Usage

```
Usage:
  dockerhub-tag list   <dockerRepo>                                   [--verbose|-v]
  dockerhub-tag add    <dockerRepo> <dockerTag> <gitTag> <location>   [--verbose|-v]
  dockerhub-tag set    <dockerRepo> <dockerTag> <gitTag> <location>   [--verbose|-v]
  dockerhub-tag delete <dockerRepo> <dockerTag>                       [--verbose|-v]
```

- **list** : Lists all automated buils in a table format.
- **add** : Creates a new automated build pointing to a git **Tag** reference
- **set** : Creates a new automated build pointing to a git **Tag** reference, while deletes all other **Tag**. (Branches are untouched)
- **delete** : Deletes a Tag by name.

## Authentication

Via environment variables
```
export DOCKERHUB_USERNAME=yourname
export DOCKERHUB_PASSWORD=Id0ntt3lU
```

## Example

```
dockerhub-tag create gliderlabs/registrator 0.4.0 0.4.0 /
```

## Single Dockerhub tag policy

Docker images are binary artifacts, so whenever you create a docker image tag based on
a github tag, it should be built only once.

Think about how you should never change a git tag, once you pushed it to the central repo.

It also creates unnecessary load on hub.dockler.com server if you rebuild the same
old (0.1.0, 0.1.1 0.1.2) docker images without any change.

## Makefile

You can trigger dockerhub tag creation by `make release`, with a couple of lines:

```
DOCKERHUB_REPO=progrium/dockerhub-tag
GITHUB_REPO=progrium/dockerhub-tag
DOCKERFILE_LOCATION=/
VERSION=$(shell cat VERSION)

deps:
	go get github.com/progrium/gh-release/...
	go get github.com/progrium/dockerhub-tag

release:
	gh-release create $(GITHUB_REPO) $(VERSION)
	dockerhub-tag set $(DOCKERHUB_REPO) $(VERSION) $(VERSION) $(DOCKERFILE_LOCATION)
```
## CircleCI

```
deployment:
  release:
    branch: release
    commands:
      - make release
```

Please remember to set env vars at **Project Settings / Environment variables**:
- DOCKERHUB_USERNAME: authentication on hub.docker.com
- DOCKERHUB_PASSWORD: credential on hub.docker.com
- GITHUB_ACCESS_TOKEN: for creating the tag on github.com

## tl;dr

Unfortunately non of the registry api versions gives you access to automated builds.
- [registry api v1](https://docs.docker.com/reference/api/registry_api/)
- [registry api v2](https://docs.docker.com/registry/spec/api/)

You can create docker image tags for **non**automated builds only:
- see [CenturyLinkLabs/docker-reg-client PR#3](https://github.com/CenturyLinkLabs/docker-reg-client/pull/3)
- see [godoc](https://github.com/CenturyLinkLabs/docker-reg-client/blob/master/registry/doc.go#L48-L51)

