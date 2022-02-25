# DO NOT EDIT. Generated with:
#
#    devctl@4.19.0
#

PACKAGE_DIR    := ./bin-dist

APPLICATION    := $(shell go list -m | cut -d '/' -f 3)
BUILDTIMESTAMP := $(shell date -u '+%FT%TZ')
GITSHA1        := $(shell git rev-parse --verify HEAD)
MODULE         := $(shell go list -m)
OS             := $(shell go env GOOS)
SOURCES        := $(shell find . -name '*.go')
VERSION        := $(shell architect project version)
ifeq ($(OS), linux)
EXTLDFLAGS := -static
endif
LDFLAGS        ?= -w -linkmode 'auto' -extldflags '$(EXTLDFLAGS)' \
  -X '$(shell go list -m)/pkg/project.buildTimestamp=${BUILDTIMESTAMP}' \
  -X '$(shell go list -m)/pkg/project.gitSHA=${GITSHA1}'

.DEFAULT_GOAL := build

##@ Go

.PHONY: build build-darwin build-darwin-64 build-linux build-linux-arm64
build: $(APPLICATION) ## Builds a local binary.
	@echo "====> $@"
build-darwin: $(APPLICATION)-darwin ## Builds a local binary for darwin/amd64.
	@echo "====> $@"
build-darwin-arm64: $(APPLICATION)-darwin-arm64 ## Builds a local binary for darwin/arm64.
	@echo "====> $@"
build-linux: $(APPLICATION)-linux ## Builds a local binary for linux/amd64.
	@echo "====> $@"
build-linux-arm64: $(APPLICATION)-linux-arm64 ## Builds a local binary for linux/arm64.
	@echo "====> $@"

$(APPLICATION): $(APPLICATION)-v$(VERSION)-$(OS)-amd64
	@echo "====> $@"
	cp -a $< $@

$(APPLICATION)-darwin: $(APPLICATION)-v$(VERSION)-darwin-amd64
	@echo "====> $@"
	cp -a $< $@

$(APPLICATION)-darwin-arm64: $(APPLICATION)-v$(VERSION)-darwin-arm64
	@echo "====> $@"
	cp -a $< $@

$(APPLICATION)-linux: $(APPLICATION)-v$(VERSION)-linux-amd64
	@echo "====> $@"
	cp -a $< $@

$(APPLICATION)-linux-arm64: $(APPLICATION)-v$(VERSION)-linux-arm64
	@echo "====> $@"
	cp -a $< $@

$(APPLICATION)-v$(VERSION)-%-amd64: $(SOURCES)
	@echo "====> $@"
	CGO_ENABLED=0 GOOS=$* GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $@ .

$(APPLICATION)-v$(VERSION)-%-arm64: $(SOURCES)
	@echo "====> $@"
	CGO_ENABLED=0 GOOS=$* GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $@ .

.PHONY: package-darwin-amd64 package-darwin-arm64 package-linux-amd64 package-linux-arm64
package-darwin-amd64: $(PACKAGE_DIR)/$(APPLICATION)-v$(VERSION)-darwin-amd64.tar.gz ## Prepares a packaged darwin/amd64 version.
	@echo "====> $@"
package-darwin-arm64: $(PACKAGE_DIR)/$(APPLICATION)-v$(VERSION)-darwin-arm64.tar.gz ## Prepares a packaged darwin/arm64 version.
	@echo "====> $@"
package-linux-amd64: $(PACKAGE_DIR)/$(APPLICATION)-v$(VERSION)-linux-amd64.tar.gz ## Prepares a packaged linux/amd64 version.
	@echo "====> $@"
package-linux-arm64: $(PACKAGE_DIR)/$(APPLICATION)-v$(VERSION)-linux-arm64.tar.gz ## Prepares a packaged linux/arm64 version.
	@echo "====> $@"

$(PACKAGE_DIR)/$(APPLICATION)-v$(VERSION)-%-amd64.tar.gz: DIR=$(PACKAGE_DIR)/$<
$(PACKAGE_DIR)/$(APPLICATION)-v$(VERSION)-%-amd64.tar.gz: $(APPLICATION)-v$(VERSION)-%-amd64
	@echo "====> $@"
	mkdir -p $(DIR)
	cp $< $(DIR)/$(APPLICATION)
	cp README.md LICENSE $(DIR)
	tar -C $(PACKAGE_DIR) -cvzf $(PACKAGE_DIR)/$<.tar.gz $<
	rm -rf $(DIR)
	rm -rf $<

$(PACKAGE_DIR)/$(APPLICATION)-v$(VERSION)-%-arm64.tar.gz: DIR=$(PACKAGE_DIR)/$<
$(PACKAGE_DIR)/$(APPLICATION)-v$(VERSION)-%-arm64.tar.gz: $(APPLICATION)-v$(VERSION)-%-arm64
	@echo "====> $@"
	mkdir -p $(DIR)
	cp $< $(DIR)/$(APPLICATION)
	cp README.md LICENSE $(DIR)
	tar -C $(PACKAGE_DIR) -cvzf $(PACKAGE_DIR)/$<.tar.gz $<
	rm -rf $(DIR)
	rm -rf $<

.PHONY: install
install: ## Install the application.
	@echo "====> $@"
	go install -ldflags "$(LDFLAGS)" .

.PHONY: run
run: ## Runs go run main.go.
	@echo "====> $@"
	go run -ldflags "$(LDFLAGS)" -race .

.PHONY: clean
clean: ## Cleans the binary.
	@echo "====> $@"
	rm -f $(APPLICATION)*
	go clean

.PHONY: imports
imports: ## Runs goimports.
	@echo "====> $@"
	goimports -local $(MODULE) -w .

.PHONY: lint
lint: ## Runs golangci-lint.
	@echo "====> $@"
	golangci-lint run -E gosec -E goconst --timeout=15m ./...

.PHONY: test
test: ## Runs go test with default values.
	@echo "====> $@"
	go test -ldflags "$(LDFLAGS)" -race ./...

.PHONY: build-docker
build-docker: build-linux ## Builds docker image to registry.
	@echo "====> $@"
	cp -a $(APPLICATION)-linux $(APPLICATION)
	docker build -t ${APPLICATION}:${VERSION} .
