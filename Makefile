.PHONY: all build test lint clean install help

# Variables
BINARY_NAME=pgedge-docloader
BINARY_PATH=bin/$(BINARY_NAME)
GO=go
GOFLAGS=
LDFLAGS=
PREFIX ?= /usr/local

all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./cmd/pgedge-docloader

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete."

## lint: Run linters
lint:
	@echo "Running linters..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...
	@echo "Linting complete."

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_PATH)
	rm -f coverage.out
	$(GO) clean
	@echo "Clean complete."

## install: Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(PREFIX)/bin..."
	install -d $(PREFIX)/bin
	install -m 755 $(BINARY_PATH) $(PREFIX)/bin/$(BINARY_NAME)
	@echo "Installation complete."

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies downloaded."

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
