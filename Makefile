# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMODTIDY=$(GOCMD) mod tidy
GOLINT=golangci-lint

# Project details
BINARY_NAME=mail-analyzer
DIST_DIR=./dist

# Get the latest git tag for versioning
GIT_TAG=$(shell git describe --tags --always --dirty --match "v*" 2>/dev/null || echo "dev")
# LDFLAGS for setting the version
LDFLAGS=-ldflags "-X main.version=$(GIT_TAG)"

.PHONY: all build clean test lint tidy vulncheck help

all: build

# Build for the current OS/Arch
build: $(DIST_DIR)/$(shell go env GOOS)/$(shell go env GOARCH)/$(BINARY_NAME)

$(DIST_DIR)/$(shell go env GOOS)/$(shell go env GOARCH)/$(BINARY_NAME):
	@echo "Building for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@mkdir -p $(@D) # Ensure output directory exists
	$(GOBUILD) $(LDFLAGS) -o $@ main.go

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	@rm -rf $(DIST_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -timeout 10s ./...

# Run linter
lint:
	@echo "Running linter..."
	@which $(GOLINT) > /dev/null || (echo "golangci-lint not found, installing..."; $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	$(GOLINT) run

# Tidy modules
tidy:
	@echo "Tidying go modules..."
	$(GOMODTIDY)

# Check for vulnerabilities in dependencies
vulncheck:
	@echo "Checking for vulnerabilities..."
	@which govulncheck > /dev/null || (echo "govulncheck not found, installing..."; $(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

# Display help
help:
	@echo "Available commands:"
	@echo "  build      - Build the application for the current OS/Arch"
	@echo "  clean      - Clean all build artifacts"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linter"
	@echo "  tidy       - Tidy go modules"
	@echo "  vulncheck  - Check for vulnerabilities in dependencies"
	@echo "  help       - Display this help message"