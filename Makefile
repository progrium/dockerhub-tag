DOCKERHUB_REPO=lalyos/dockerhub-tag
GITHUB_REPO=progrium/dockerhub-tag
DOCKERFILE_LOCATION=/
VERSION=$(shell cat VERSION)

deps:
	go get github.com/progrium/gh-release/...
	go get github.com/progrium/dockerhub-tag

release:
	gh-release create $(GITHUB_REPO) $(VERSION)
	dockerhub-tag create $(DOCKERHUB_REPO) $(VERSION) $(VERSION) $(DOCKERFILE_LOCATION)
