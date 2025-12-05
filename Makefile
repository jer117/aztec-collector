CI_COMMIT_REF_NAME ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
CI_PROJECT_NAME ?= $(notdir $(CURDIR))
UNAME ?= $(shell uname -s | tr A-Z a-z)
GOARCH ?= $(shell go env GOARCH)

NAME := $(CI_PROJECT_NAME)
VERSION ?= $(CI_COMMIT_REF_NAME)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS := -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

.PHONY: all
all: build

.PHONY: build
build:
	@mkdir -p build
	CGO_ENABLED=0 GOOS=$(UNAME) GOARCH=$(GOARCH) go build $(LDFLAGS) -o build/aztec-collector-$(UNAME)-$(GOARCH) -a -installsuffix cgo ./cmd/collector/main.go

.PHONY: build-linux
build-linux:
	@mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/aztec-collector-linux-amd64 -a -installsuffix cgo ./cmd/collector/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o build/aztec-collector-linux-arm64 -a -installsuffix cgo ./cmd/collector/main.go

.PHONY: build-darwin
build-darwin:
	@mkdir -p build
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o build/aztec-collector-darwin-amd64 -a -installsuffix cgo ./cmd/collector/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o build/aztec-collector-darwin-arm64 -a -installsuffix cgo ./cmd/collector/main.go

.PHONY: build-all
build-all: build-linux build-darwin

.PHONY: check
check: lint test

.PHONY: test
test:
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --enable gofmt ./...; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: deps
deps:
	go mod download

.PHONY: run
run: build
	./build/aztec-collector-$(UNAME)-$(GOARCH) --config example-config.yaml

.PHONY: clean
clean:
	rm -rf build/
	rm -f coverage.out coverage.html
	rm -f aztec-collector.json

.PHONY: docker-build
docker-build:
	docker build -t aztec-collector:$(VERSION) .

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build for current platform"
	@echo "  build-linux  - Build for Linux (amd64 and arm64)"
	@echo "  build-darwin - Build for macOS (amd64 and arm64)"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  tidy         - Tidy go modules"
	@echo "  deps         - Download dependencies"
	@echo "  run          - Build and run with example config"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"

