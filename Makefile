# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

SHELL := /bin/sh

GO ?= go
GOFLAGS ?= -trimpath
BIN_DIR ?= bin
BINARY_NAME ?= oasdiff
VERSION=$(shell git describe --always --tags | cut -d "v" -f 2)
LINKER_FLAGS=-s -w -X github.com/oasdiff/oasdiff/build.Version=${VERSION}
GOLANGCILINT_VERSION=v2.11.4
CMD_DIR := .

.PHONY: all build test vet linux darwin windows clean cross

all: test build

.PHONY: test
test: localize ## Run tests
	@echo "==> Running tests..."
	$(GO) test ./...

.PHONY: coverage
coverage: ## Run tests with coverage
	@echo "==> Running tests with coverage..."
	$(GO) test -coverprofile=coverage.out ./...

.PHONY: coverage-html
coverage-html: coverage ## Generate HTML coverage report
	@echo "==> Generating HTML coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: build
build: ## Build oasdiff binary for current platform
	@echo "==> Building oasdiff binary"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LINKER_FLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)

.PHONY: install
install: deps ## Install oasdiff binary
	@echo "==> Installing oasdiff binary..."
	$(GO) install -ldflags "$(LINKER_FLAGS)" $(CMD_DIR)

.PHONY: linux
linux: ## Cross-compile for Linux (amd64 + arm64)
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LINKER_FLAGS)' -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags '$(LINKER_FLAGS)' -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)

.PHONY: darwin
darwin: ## Cross-compile for macOS (amd64 + arm64)
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LINKER_FLAGS)' -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags '$(LINKER_FLAGS)' -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)

.PHONY: windows
windows: ## Cross-compile for Windows (amd64)
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LINKER_FLAGS)' -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

.PHONY: cross
cross: linux darwin windows ## Cross-compile for all platforms

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out coverage.html

.PHONY: deps
deps:  ## Download go module dependencies
	@echo "==> Installing go.mod dependencies..."
	$(GO) mod download
	$(GO) mod tidy

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: modernize ## Run linter
	$(GO) fmt ./...
	$(GO) vet ./...
	golangci-lint run --enable=unused

.PHONY: modernize
modernize: ## Report code that could use newer Go constructs
	$(GO) run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@v0.48.0 -test -omitzero=false $$(go list ./... | grep -v checker/localizations)
	
.PHONY: localize
localize: ## Compile localized changelog messages
	@echo "==> Compiling localized changelog messages..."
	$(GO) install github.com/m1/go-localize@latest
	go-localize -input checker/localizations_src -output checker/localizations 
	$(GO) fmt ./checker/localizations

.PHONY: devtools
devtools:  ## Install dev tools
	@echo "==> Installing dev tools..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCILINT_VERSION)

.PHONY: help
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: link-git-hooks
link-git-hooks: ## Install git hooks
	@echo "==> Installing all git hooks..."
	find .git/hooks -type l -exec rm {} \;
	find .githooks -type f -exec ln -sf ../../{} .git/hooks/ \;