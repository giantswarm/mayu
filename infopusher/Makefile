PROJECT=infopusher
ORGANIZATION=giantswarm

SOURCE := $(shell find . -name '*.go')
VERSION := $(shell cat ../VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

.PHONY: all clean

all: $(PROJECT)

clean:
	rm -rf infopusher bindata.go

../.gobuild/bin/go-bindata:
	docker run \
	    --rm \
	    -v $(shell pwd)/../.gobuild:/go \
	    -e GOOS=linux \
	    -e GOARCH=amd64 \
	    golang:1.5 \
	    go get github.com/jteeuwen/go-bindata/go-bindata

bindata.go: ../.gobuild/bin/go-bindata
	docker run \
	    --rm \
	    -v $(shell pwd)/../.gobuild:/go \
	    -v $(shell pwd):/usr/code \
	    -e GOOS=linux \
	    -e GOARCH=amd64 \
	    -w /usr/code \
	    golang:1.5 \
	    /go/bin/go-bindata embedded

$(PROJECT): bindata.go $(SOURCE) ../VERSION
	@echo Building for linux/amd64

	docker run \
	    --rm \
	    -v $(shell pwd)/../.gobuild:/go \
	    golang:1.5 \
	    go get github.com/golang/glog

	docker run \
	    --rm \
	    -v $(shell pwd)/../.gobuild:/go \
	    -v $(shell pwd):/go/src/github.com/giantswarm/mayu/infopusher \
	    -v $(shell pwd):/go/out \
	    -e GOOS=linux \
	    -e GOARCH=amd64 \
	    golang:1.5 \
	    go build -a -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" \
	       -o /go/out/$(PROJECT) github.com/giantswarm/mayu/infopusher
