PROJECT=mayu
ORGANISATION=giantswarm

.PHONY: all mayu

GO_VERSION=$(shell go version | cut -d ' ' -f 3)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

all: infopusher/infopusher helpers/infopusher bin/mayu bin/mayuctl

infopusher/infopusher:
	cd infopusher ; make

helpers/infopusher: infopusher/infopusher
	cp ./infopusher/infopusher helpers/infopusher

bin/mayu:
	docker run \
		--rm \
		-v $(shell pwd):/go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-e GOPATH=/go \
		-e CGOENABLED=0 \
		-w /go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		golang:1.7.5 \
		go build -a -v -tags netgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(GIT_COMMIT)" -o bin/mayu github.com/$(ORGANISATION)/$(PROJECT)

bin/mayuctl:
	docker run \
		--rm \
		-v $(shell pwd):/go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-e GOPATH=/go \
		-e CGOENABLED=0 \
		-w /go/src/github.com/$(ORGANISATION)/$(PROJECT) \
		golang:1.7.5 \
		go build -a -v -tags netgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(GIT_COMMIT)" -o bin/mayuctl github.com/$(ORGANISATION)/$(PROJECT)/mayuctl

bin-dist: all
		mkdir -p bin-dist/tftproot
		mkdir -p bin-dist/static_html
		mkdir -p bin-dist/templates
		mkdir -p bin-dist/template_snippets
		mkdir -p bin-dist/images
		cp helpers/undionly.kpxe bin-dist/tftproot
		cp infopusher/infopusher bin-dist/static_html
		cp bin/mayuctl bin-dist/static_html
		cp -f bin/mayu bin-dist
		cp -f bin/mayuctl bin-dist
		cp config.yaml.dist bin-dist
		cp Dockerfile.dist bin-dist/Dockerfile
		cp .dockerignore.dist bin-dist/.dockerignore
		cp -a templates/* bin-dist/templates
		cp -a template_snippets/* bin-dist/template_snippets
		cp scripts/fetch-coreos-image bin-dist/fetch-coreos-image
		cp scripts/fetch-coreos-qemu-image bin-dist/fetch-coreos-qemu-image
		cp scripts/fetch-yochu-assets bin-dist/fetch-yochu-assets
		cd bin-dist && rm -f $(PROJECT).*.tar.gz && tar czf $(PROJECT).$(VERSION)-linux-amd64.tar.gz *

release386: bin-dist
		rm -rf bin
		@GOOS=linux; GOARCH=386 $(MAKE) bin/mayu
		@GOOS=linux; GOARCH=386 $(MAKE) bin/mayuctl
		cp bin/mayu* bin-dist
		cd bin-dist && rm -f $(PROJECT).*-linux-i386.tar.gz && tar czf $(PROJECT).$(VERSION)-linux-i386.tar.gz --exclude='*.tar.gz' *

clean:
		rm -rf .gobuild bin helpers/infopusher
		cd infopusher ; make clean
