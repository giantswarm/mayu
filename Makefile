PROJECT=mayu
ORGANIZATION=giantswarm

BINARY_SERVER=bin/$(PROJECT)
BINARY_CTL=bin/$(PROJECT)ctl

SOURCE := $(shell find . -name '*.go')
VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)
GOPATH := $(shell pwd)/.gobuild
PROJECT_PATH := $(GOPATH)/src/github.com/$(ORGANIZATION)

ifndef GOOS
  GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
  GOARCH := $(shell go env GOARCH)
endif

.PHONY: all clean bin-dist clean-bin-dist publish vendor-clean vendor-update release386

all: .gobuild infopusher/infopusher helpers/infopusher $(BINARY_SERVER) $(BINARY_CTL)

.gobuild:
	mkdir -p $(PROJECT_PATH)
	mkdir -p $(GOPATH)/doc
	cd $(PROJECT_PATH) && ln -s ../../../.. $(PROJECT)

infopusher/infopusher:
	cd infopusher ; make

helpers/infopusher: infopusher/infopusher
	cp ./infopusher/infopusher helpers/infopusher

test: .gobuild
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -e GO15VENDOREXPERIMENT=1 \
	    -w /usr/code/ \
		golang:1.5 \
	    bash -c 'cd .gobuild/src/github.com/giantswarm/mayu && go test $$(go list ./... | grep -v vendor)'

$(BINARY_SERVER): $(SOURCE) VERSION .gobuild
	@echo Building for $(GOOS)/$(GOARCH)
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -e GO15VENDOREXPERIMENT=1 \
	    -w /usr/code \
      golang:1.5 \
	    go build -a -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o $(BINARY_SERVER) github.com/$(ORGANIZATION)/$(PROJECT)

$(BINARY_CTL): $(SOURCE) VERSION .gobuild
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -e GO15VENDOREXPERIMENT=1 \
	    -w /usr/code \
      golang:1.5 \
	    go build -a -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o $(BINARY_CTL) github.com/$(ORGANIZATION)/$(PROJECT)/mayuctl

distclean: clean clean-bin-dist

clean:
	rm -rf .gobuild bin helpers/infopusher
	cd infopusher ; make clean

clean-bin-dist:
	rm -fr bin-dist

bin-dist: all
	mkdir -p bin-dist/tftproot
	mkdir -p bin-dist/static_html
	mkdir -p bin-dist/templates
	mkdir -p bin-dist/template_snippets
	mkdir -p bin-dist/images
	cp helpers/undionly.kpxe bin-dist/tftproot
	cp infopusher/infopusher bin-dist/static_html
	cp $(BINARY_CTL) bin-dist/static_html
	cp -f $(BINARY_SERVER) bin-dist
	cp -f $(BINARY_CTL) bin-dist
	cp config.yaml.dist bin-dist
	cp Dockerfile.dist bin-dist/Dockerfile
	cp .dockerignore.dist bin-dist/.dockerignore
	cp -a templates/* bin-dist/templates
	cp -a template_snippets/* bin-dist/template_snippets
	cp scripts/fetch-coreos-image bin-dist/fetch-coreos-image
	cp scripts/fetch-coreos-qemu-image bin-dist/fetch-coreos-qemu-image
	cp scripts/fetch-yochu-assets bin-dist/fetch-yochu-assets
	cd bin-dist && rm -f $(PROJECT).*.tar.gz && tar czf $(PROJECT).$(VERSION)-linux-amd64.tar.gz *

vendor-clean:
	rm -rf vendor/

vendor-update: vendor-clean
	rm -rf glide.lock
	GO15VENDOREXPERIMENT=1 glide install
	find vendor/ -name .git -type d -prune | xargs rm -rf

install: $(BINARY_SERVER) $(BINARY_CTL)
	cp $(BINARY_SERVER) $(BINARY_CTL) /usr/local/bin/

release386: bin-dist
	rm -rf bin
	@GOARCH=386 $(MAKE) $(BINARY_SERVER)
	@GOARCH=386 $(MAKE) $(BINARY_CTL)
	cp bin/mayu* bin-dist
	cd bin-dist && rm -f $(PROJECT).*-linux-i386.tar.gz && tar czf $(PROJECT).$(VERSION)-linux-i386.tar.gz --exclude='*.tar.gz' *

godoc: all
	@echo Opening godoc server at http://localhost:6060/pkg/github.com/$(ORGANIZATION)/$(PROJECT)/
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOROOT=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -e GO15VENDOREXPERIMENT=1 \
	    -w /usr/code \
      -p 6060:6060 \
		golang:1.5 \
		godoc -http=:6060
