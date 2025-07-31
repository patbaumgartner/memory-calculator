# JVM Memory Calculator Makefile

# Build variables
BINARY_NAME=memory-calculator
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH)"

# Output directories
DIST_DIR=dist
COVERAGE_DIR=coverage

.PHONY: all build build-all clean test test-coverage coverage coverage-html deps help vulncheck

all: clean deps test build

## Build commands
build: ## Build binary for current platform
	@echo "Building $(BINARY_NAME) for current platform..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/memory-calculator
	@echo "Build complete: $(BINARY_NAME)"

build-all: ## Build binaries for all platforms
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/memory-calculator
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/memory-calculator
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/memory-calculator
	
	# macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/memory-calculator
	
	@echo "Cross-platform build complete. Binaries in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

## Test commands
test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out

coverage: ## Run tests with coverage (alias for test-coverage)
	@$(MAKE) test-coverage

coverage-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

benchmark-compare: ## Run benchmarks and save results for comparison
	@echo "Running benchmarks with comparison data..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -bench=. -benchmem ./... > $(COVERAGE_DIR)/benchmark.txt
	@echo "Benchmark results saved to $(COVERAGE_DIR)/benchmark.txt"

## Dependency commands
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## Security commands
security: ## Run security checks
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

security-install: ## Install security tools
	@echo "Installing security tools..."
	go install github.com/securego/gosec/v2/cmd/gosec@latest

vuln-check: ## Check for known vulnerabilities
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	elif [ -f ~/go/bin/govulncheck ]; then \
		~/go/bin/govulncheck ./...; \
	else \
		echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

vulncheck: ## Check for known vulnerabilities (alias for vuln-check)
	@$(MAKE) vuln-check

vuln-install: ## Install vulnerability checker
	@echo "Installing vulnerability checker..."
	go install golang.org/x/vuln/cmd/govulncheck@latest

## Utility commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BINARY_NAME)
	rm -rf $(DIST_DIR)
	rm -rf $(COVERAGE_DIR)

format: ## Format Go code
	@echo "Formatting code..."
	gofmt -s -w .

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

lint-install: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Development commands
dev: ## Run in development mode
	@echo "Running in development mode..."
	$(GOCMD) run ./cmd/memory-calculator --help

dev-test: ## Run with test parameters
	@echo "Running with test parameters..."
	$(GOCMD) run ./cmd/memory-calculator --total-memory 2G --thread-count 250

install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	cp $(BINARY_NAME) $(GOPATH)/bin/

## Release commands
release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@git status --porcelain
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Working directory is not clean"; \
		exit 1; \
	fi
	@echo "✓ Working directory is clean"
	@$(GOTEST) ./...
	@echo "✓ All tests pass"
	@echo "Ready for release"

## Help
help: ## Show this help message
	@echo "JVM Memory Calculator - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Default target
.DEFAULT_GOAL := help
