
# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMODTIDY=$(GOCMD) mod tidy
GOLINT=golangci-lint

# Binary name
BINARY_NAME=mail-analyzer

.PHONY: all build clean test lint tidy vulncheck help

all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) main.go

# Clean the binary
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

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
	@echo "  build      - Build the application"
	@echo "  clean      - Clean the binary"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linter"
	@echo "  tidy       - Tidy go modules"
	@echo "  vulncheck  - Check for vulnerabilities in dependencies"
	@echo "  help       - Display this help message"

