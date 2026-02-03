.PHONY: build build-all run clean install test

# Binary name
BINARY_NAME=noa

# Version (can be overridden)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .

# Build for all platforms
build-all: clean-dist
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p dist
	
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-linux-amd64 .
	
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-linux-arm64 .
	
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-darwin-amd64 .
	
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-darwin-arm64 .
	
	@echo "Done! Binaries in dist/"
	@ls -la dist/


# Install the binary to GOPATH/bin or GOBIN
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install .

# Clean build artifacts
clean: clean-dist
	@echo "Cleaning..."
	@go clean
	@rm -f $(BINARY_NAME)

# Clean dist folder
clean-dist:
	@rm -rf dist/

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go mod tidy
tidy:
	@echo "Running go mod tidy..."
	@go mod tidy

