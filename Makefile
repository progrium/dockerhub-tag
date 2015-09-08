deps:
	go get github.com/gliderlabs/glu
	
build:
	glu build darwin,linux . dockerhub-tag

release: build
	glu release
