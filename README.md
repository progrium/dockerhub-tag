Did you ever find yourself creating automated builds by hand on http://hub.docker.com for each new git tag?
Why is it called *automated* build, if you cann’t create a new one automatically? This project helps to
do the missing step.


## Usage

```
Usage:
  dockerhub-tag create <dockerRepo> <dockerTag> <gitTag> <location>   [--verbose|-v]
```

It will create a new 

## Example

```
dockerhub-tag create lalyos/dockerhub-tag 0.1.0 0.1.0 /
```

## Single Dockerhub tag policy

Docker images are binary artifacts, so whenever you create a docker image tag based on
a github tag, it should be built only once.

Think about how you should never change a git tag, once you pushed it to the central repo.

It also creates unnecessary load on hub.dockler.com server if you rebuild the same 
old (0.1.0, 0.1.1 0.1.2) docker images without any change.

## Makefile

You can trigger dockerhub tag creation by `make docker-tag`, with a couple of lines:

```
DOCKERHUB_REPO=lalyos/dockerhub-tag
DOCKERFILE_LOCATION=/
GIT_TAG=$(shell git describe --tags)

deps:
        go get github.com/lalyos/dockerhub-tag

docker-tag:
        dockerhub-tag create $(DOCKERHUB_REPO) $(GIT_TAG) $(GIT_TAG) $(DOCKERFILE_LOCATION)
```
## CircleCI

???

## tl;dr

Unfortunately non of the registry api versions gives you access to automated builds.§
- [registry api v1](https://docs.docker.com/reference/api/registry_api/)
- [registry api v2](https://docs.docker.com/registry/spec/api/)

You can create docker image tags for **non**automated builds only:
- see [CenturyLinkLabs/docker-reg-client PR#3](https://github.com/CenturyLinkLabs/docker-reg-client/pull/3)
- see [godoc](https://github.com/CenturyLinkLabs/docker-reg-client/blob/master/registry/doc.go#L48-L51)

