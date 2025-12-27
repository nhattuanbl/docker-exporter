# Docker Exporter Makefile

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION ?= $(shell go version | cut -d' ' -f3)

# Build flags
LDFLAGS := -s -w \
	-X github.com/nhattuanbl/docker-exporter/internal/config.Version=$(VERSION) \
	-X github.com/nhattuanbl/docker-exporter/internal/config.GitCommit=$(GIT_COMMIT) \
	-X github.com/nhattuanbl/docker-exporter/internal/config.BuildDate=$(BUILD_DATE) \
	-X github.com/nhattuanbl/docker-exporter/internal/config.GoVersion=$(GO_VERSION)

# Binary name
BINARY := docker-exporter

# Platforms for cross-compilation
PLATFORMS := linux/amd64 linux/arm64 linux/arm darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build clean test fmt lint release help

## Build for current platform
build:
	@echo "Building $(BINARY) $(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/docker-exporter

## Build for all platforms
release: clean
	@echo "Building releases..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output="dist/$(BINARY)_$${os}_$${arch}"; \
		if [ "$$os" = "windows" ]; then output="$$output.exe"; fi; \
		echo "  Building $$output..."; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o $$output ./cmd/docker-exporter; \
	done
	@echo "Done! Binaries in dist/"

## Run tests
test:
	go test -v ./...

## Format code
fmt:
	go fmt ./...

## Run linter
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed"; \
	fi

## Clean build artifacts
clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf dist/

## Run the exporter (development)
run:
	go run ./cmd/docker-exporter -l debug

## Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"

## Install dependencies
deps:
	go mod download
	go mod tidy

## Show help
help:
	@echo "Docker Exporter Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build    Build for current platform"
	@echo "  release  Build for all platforms"
	@echo "  test     Run tests"
	@echo "  fmt      Format code"
	@echo "  lint     Run linter"
	@echo "  clean    Clean build artifacts"
	@echo "  run      Run in development mode"
	@echo "  version  Show version info"
	@echo "  deps     Install dependencies"
	@echo "  help     Show this help"
