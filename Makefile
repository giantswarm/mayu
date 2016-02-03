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

COREOS_VERSION := 681.2.0
ETCD_VERSION := v2.2.1-gs-1
FLEET_VERSION := v0.11.3-gs-2
YOCHU_VERSION := 0.15.1

.PHONY: all clean bin-dist clean-bin-dist publish vendor-clean vendor-update

all: .gobuild infopusher/infopusher helpers/infopusher $(BINARY_SERVER) $(BINARY_CTL) cache/coreos_production_pxe.vmlinuz cache/coreos_production_pxe_image.cpio.gz cache/coreos_production_image.bin.bz2 cache/yochu/$(YOCHU_VERSION) cache/fleet/$(FLEET_VERSION) cache/etcd/$(ETCD_VERSION)

.gobuild:
	mkdir -p $(PROJECT_PATH)
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

distclean: clean clean-cache clean-bin-dist

clean:
	rm -rf .gobuild bin helpers/infopusher
	cd infopusher ; make clean

clean-cache:
	rm -f cache/coreos_production_pxe.vmlinuz
	rm -f cache/coreos_pxe_image.cpio.gz
	rm -f cache/coreos_production_image.bin.bz2
	rm -f cache/coreos_production_pxe_image.cpio.gz
	rm -rf cache/yochu
	rm -rf cache/etcd
	rm -rf cache/fleet

cache/coreos_production_pxe.vmlinuz:
	wget -O cache/coreos_production_pxe.vmlinuz  http://stable.release.core-os.net/amd64-usr/${COREOS_VERSION}/coreos_production_pxe.vmlinuz

cache/coreos_pxe_image.cpio.gz:
	wget -O cache/coreos_pxe_image.cpio.gz  http://stable.release.core-os.net/amd64-usr/${COREOS_VERSION}/coreos_production_pxe_image.cpio.gz

cache/coreos_production_image.bin.bz2:
	wget -O cache/coreos_production_image.bin.bz2 http://stable.release.core-os.net/amd64-usr/${COREOS_VERSION}/coreos_production_image.bin.bz2

cache/coreos_production_pxe_image.cpio.gz: cache/coreos_pxe_image.cpio.gz
	docker run --rm -v $(shell pwd)/cache:/usr/code/cache \
			debian:jessie /bin/bash -c "apt-get update -y && apt-get install cpio && \
			zcat /usr/code/cache/coreos_pxe_image.cpio.gz > /usr/code/cache/coreos_production_pxe_image.cpio && \
			cd /usr/code/cache && find usr | cpio -o -A -H newc -O coreos_production_pxe_image.cpio && \
			gzip coreos_production_pxe_image.cpio"

cache/yochu/$(YOCHU_VERSION):
	mkdir -p cache/yochu/${YOCHU_VERSION}
	wget -O cache/yochu/${YOCHU_VERSION}/yochu https://bootstrap.giantswarm.io/yochu/${YOCHU_VERSION}/yochu

cache/etcd/$(ETCD_VERSION):
	mkdir -p cache/etcd/${ETCD_VERSION}
	wget -O cache/etcd/${ETCD_VERSION}/etcd https://bootstrap.giantswarm.io/etcd/${ETCD_VERSION}/etcd
	wget -O cache/etcd/${ETCD_VERSION}/etcdctl https://bootstrap.giantswarm.io/etcd/${ETCD_VERSION}/etcdctl

cache/fleet/$(FLEET_VERSION):
	mkdir -p cache/fleet/${FLEET_VERSION}
	wget -O cache/fleet/${FLEET_VERSION}/fleetd https://bootstrap.giantswarm.io/fleet/${FLEET_VERSION}/fleetd
	wget -O cache/fleet/${FLEET_VERSION}/fleetctl https://bootstrap.giantswarm.io/fleet/${FLEET_VERSION}/fleetctl

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
	cp -R cache/yochu bin-dist/static_html
	cp -R cache/etcd bin-dist/static_html
	cp -R cache/fleet bin-dist/static_html
	cp cache/coreos_production_pxe_image.cpio.gz cache/coreos_production_pxe.vmlinuz bin-dist/images
	cp cache/coreos_production_image.bin.bz2 bin-dist/images
	cp -f $(BINARY_SERVER) bin-dist
	cp -f $(BINARY_CTL) bin-dist
	cp config.yaml.dist bin-dist
	cp -a templates/* bin-dist/templates
	cp -a template_snippets/* bin-dist/template_snippets
	cd bin-dist && tar czf $(PROJECT).$(VERSION).tar.gz *

vendor-clean:
	rm -rf vendor/

vendor-update: vendor-clean
	rm -rf glide.lock
	GO15VENDOREXPERIMENT=1 glide install
	find vendor/ -name .git -type d -prune | xargs rm -rf

install: $(BINARY_SERVER) $(BINARY_CTL)
	cp $(BINARY_SERVER) $(BINARY_CTL) /usr/local/bin/
