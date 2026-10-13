BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
COMMIT := $(shell git log -1 --format='%H' 2>/dev/null)
APPNAME := atlas

# Build output location
BUILD_DIR := $(CURDIR)/build
BUILD_TARGET := $(BUILD_DIR)/$(APPNAME)d

# Go toolchain settings
GO ?= go
GOFLAGS ?= -mod=readonly
CGO_ENABLED ?= 0
BUILD_TAGS ?= ""

# Fallbacks so the build works outside a git repo (e.g. a tarball or CI cache)
ifeq (,$(BRANCH))
  BRANCH := unknown
endif
ifeq (,$(COMMIT))
  COMMIT := unknown
endif

# do not override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

# Update the ldflags with the app, client & server names
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(APPNAME) \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(APPNAME)d \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -tags $(BUILD_TAGS) -ldflags '$(ldflags)'


#####  Build #####

# Build the chain binary into ./build for the host platform.
build:
	@echo "--> Building $(APPNAME)d ($(VERSION)) for $(shell $(GO) env GOOS)/$(shell $(GO) env GOARCH)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) $(BUILD_FLAGS) -o $(BUILD_TARGET) ./cmd/$(APPNAME)d
	@echo "--> Built: $(BUILD_TARGET)"

# Cross-compile a static Linux (amd64) binary
build-linux:
	@echo "--> Building $(APPNAME)d for linux/amd64"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(BUILD_FLAGS) -o $(BUILD_TARGET)-linux-amd64 ./cmd/$(APPNAME)d
	@echo "--> Built: $(BUILD_TARGET)-linux-amd64"

# Cross-compile a static Linux arm64 binary.
build-linux-arm64:
	@echo "--> Building $(APPNAME)d for linux/arm64"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(BUILD_FLAGS) -o $(BUILD_TARGET)-linux-arm64 ./cmd/$(APPNAME)d
	@echo "--> Built: $(BUILD_TARGET)-linux-arm64"

clean:
	@echo "--> Cleaning build directory"
	@rm -rf $(BUILD_DIR)

.PHONY: build build-linux build-linux-arm64 clean


#####  Install  #####

all: build

install:
	@echo "--> ensure dependencies have not been modified"
	$(GO) mod verify
	@echo "--> installing $(APPNAME)d"
	$(GO) install $(BUILD_FLAGS) $(GOFLAGS) ./cmd/$(APPNAME)d

.PHONY: all install


#####  Test  #####

test-unit:
	@echo Running unit tests...
	@go test -mod=readonly -v -timeout 30m ./...

test-race:
	@echo Running unit tests with race condition reporting...
	@go test -mod=readonly -v -race -timeout 30m ./...

test-cover:
	@echo Running unit tests and creating coverage report...
	@go test -mod=readonly -v -timeout 30m -coverprofile=$(COVER_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVER_FILE) -o $(COVER_HTML_FILE)
	@rm $(COVER_FILE)

bench:
	@echo Running unit tests with benchmarking...
	@go test -mod=readonly -v -timeout 30m -bench=. ./...

test: govet govulncheck test-unit

.PHONY: test test-unit test-race test-cover bench


#####  Linting  #####

lint:
	@echo "--> Running linter"
	@go tool github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --timeout 15m

lint-fix:
	@echo "--> Running linter and fixing issues"
	@go tool github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --fix --timeout 15m

.PHONY: lint lint-fix


##### Development #####

govet:
	@echo Running go vet...
	@go vet ./...

govulncheck:
	@echo Running govulncheck...
	@go tool golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

.PHONY: govet govulncheck
