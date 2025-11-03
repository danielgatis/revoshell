.PHONY: build clean test fmt lint help version

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS := -X github.com/danielgaits/revoshell/pkg/version.Version=$(VERSION) \
           -X github.com/danielgaits/revoshell/pkg/version.GitCommit=$(GIT_COMMIT) \
           -X github.com/danielgaits/revoshell/pkg/version.BuildDate=$(BUILD_DATE)

build:
	@echo "Building revoshell..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@mkdir -p bin
	@go build -ldflags "$(LDFLAGS)" -o bin/revoshell .

clean:
	@rm -rf bin

test:
	@go test -v -race ./...

fmt:
	@gofmt -s -w .
	@go mod tidy

lint:
	@go tool golangci-lint run --fix --timeout 10m0s

deps:
	@go mod download
	@go mod tidy

version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
