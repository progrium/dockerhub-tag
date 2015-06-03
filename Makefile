DOCKERHUB_REPO=lalyos/dockerhub-tag
DOCKERFILE_LOCATION=/
GIT_TAG=$(shell git describe --tags)

deps:
	go get github.com/lalyos/dockerhub-tag

docker-tag:
	dockerhub-tag create $(DOCKERHUB_REPO) $(GIT_TAG) $(GIT_TAG) $(DOCKERFILE_LOCATION)
