.PHONY: build run clean install test

# Binary name
BINARY_NAME=noa

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .

# Run the application
run: build
	@./$(BINARY_NAME) address test-address

# Install the binary to GOPATH/bin or GOBIN
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@go clean
	@rm -f $(BINARY_NAME)

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

