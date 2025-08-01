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

# Go tool paths
GOPATH=$(shell $(GOCMD) env GOPATH)
GOBIN=$(shell $(GOCMD) env GOBIN)
ifeq ($(GOBIN),)
GOBIN=$(GOPATH)/bin
endif

# Tool binaries (with fallback paths)
GOLANGCI_LINT=$(shell command -v golangci-lint 2>/dev/null || echo "$(GOBIN)/golangci-lint")
GOSEC=$(shell command -v gosec 2>/dev/null || echo "$(GOBIN)/gosec")
GOVULNCHECK=$(shell command -v govulncheck 2>/dev/null || echo "$(GOBIN)/govulncheck")
GOFUMPT=$(shell command -v gofumpt 2>/dev/null || echo "$(GOBIN)/gofumpt")

# Build flags for optimized binaries
# -s: Strip symbol table and debug info (reduces size)
# -w: Strip DWARF debug info (reduces size further)  
# -trimpath: Remove file system paths from executable (reproducible builds)
# -a: Force rebuilding of packages (ensures clean build)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH) -s -w"
BUILD_FLAGS=-trimpath -a

# Output directories
DIST_DIR=dist
COVERAGE_DIR=coverage

.PHONY: all build build-all build-compressed build-minimal build-size-comparison build-ultimate-comparison clean test test-all integration test-coverage coverage coverage-html benchmark benchmark-compare deps tools tools-check security security-install vuln-check vulncheck vuln-install format quality lint lint-install dev dev-test install release-check docker-build docker-run docker-test help

all: clean deps test build ## Build everything (clean, deps, test, build)

## Build commands
build: ## Build binary for current platform
	@echo "Building $(BINARY_NAME) for current platform..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/memory-calculator
	@echo "Build complete: $(BINARY_NAME)"

build-all: ## Build binaries for all platforms
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/memory-calculator
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/memory-calculator
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/memory-calculator
	
	# macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/memory-calculator
	
	@echo "Cross-platform build complete. Binaries in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

build-compressed: ## Build ultra-compressed binary (requires upx)
	@echo "Building ultra-compressed $(BINARY_NAME)..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/memory-calculator
	@if command -v upx >/dev/null 2>&1; then \
		echo "Compressing binary with UPX..."; \
		upx --best --lzma $(BINARY_NAME); \
		echo "Ultra-compressed build complete: $(BINARY_NAME)"; \
	else \
		echo "Warning: UPX not found. Install with: sudo apt install upx-ucl (Ubuntu/Debian) or brew install upx (macOS)"; \
		echo "Regular optimized build complete: $(BINARY_NAME)"; \
	fi

build-minimal: ## Build minimal binary without optional features
	@echo "Building minimal $(BINARY_NAME) (excludes ZIP processing, uses estimates)..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -tags minimal -o $(BINARY_NAME)-minimal ./cmd/memory-calculator
	@echo "Minimal build complete: $(BINARY_NAME)-minimal"

build-size-comparison: ## Compare binary sizes with and without optimization
	@echo "Building size comparison..."
	@echo "Building without optimization..."
	$(GOBUILD) -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH)" -o $(BINARY_NAME)-unoptimized ./cmd/memory-calculator
	@echo "Building with optimization..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-optimized ./cmd/memory-calculator
	@echo ""
	@echo "Size comparison:"
	@echo "Unoptimized: $$(du -h $(BINARY_NAME)-unoptimized | cut -f1)"
	@echo "Optimized:   $$(du -h $(BINARY_NAME)-optimized | cut -f1)"
	@echo "Savings:     $$(echo "scale=1; (($$(stat -c%s $(BINARY_NAME)-unoptimized) - $$(stat -c%s $(BINARY_NAME)-optimized)) / $$(stat -c%s $(BINARY_NAME)-unoptimized)) * 100" | bc)%"
	@rm -f $(BINARY_NAME)-unoptimized $(BINARY_NAME)-optimized

build-ultimate-comparison: ## Compare all build variants (standard, minimal, compressed)
	@echo "Building ultimate size comparison..."
	@echo "1. Building standard optimized..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-standard ./cmd/memory-calculator
	@echo "2. Building minimal..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -tags minimal -o $(BINARY_NAME)-minimal ./cmd/memory-calculator
	@echo "3. Building unoptimized (for comparison)..."
	$(GOBUILD) -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH)" -o $(BINARY_NAME)-unoptimized ./cmd/memory-calculator
	@echo ""
	@echo "Ultimate size comparison:"
	@echo "Unoptimized: $$(du -h $(BINARY_NAME)-unoptimized | cut -f1)"
	@echo "Standard:    $$(du -h $(BINARY_NAME)-standard | cut -f1)"
	@echo "Minimal:     $$(du -h $(BINARY_NAME)-minimal | cut -f1)"
	@echo ""
	@echo "Savings (vs unoptimized):"
	@echo "Standard: $$(echo "scale=1; (($$(stat -c%s $(BINARY_NAME)-unoptimized) - $$(stat -c%s $(BINARY_NAME)-standard)) / $$(stat -c%s $(BINARY_NAME)-unoptimized)) * 100" | bc)%"
	@echo "Minimal:  $$(echo "scale=1; (($$(stat -c%s $(BINARY_NAME)-unoptimized) - $$(stat -c%s $(BINARY_NAME)-minimal)) / $$(stat -c%s $(BINARY_NAME)-unoptimized)) * 100" | bc)%"
	@rm -f $(BINARY_NAME)-unoptimized $(BINARY_NAME)-standard $(BINARY_NAME)-minimal

## Test commands
test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

integration: ## Run integration tests only
	@echo "Running integration tests..."
	$(GOTEST) -v -run "TestMain" .

test-all: test integration ## Run all tests including integration tests
	@echo "All tests completed"

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

tools: ## Install all required development tools
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing gosec..."
	$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Installing govulncheck..."
	$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Installing gofumpt..."
	$(GOCMD) install mvdan.cc/gofumpt@latest
	@echo "All tools installed to $(GOBIN)/"
	@echo "Make sure $(GOBIN) is in your PATH"

tools-check: ## Check if all tools are available
	@echo "Checking development tools..."
	@echo -n "golangci-lint: "; if [ -x "$(GOLANGCI_LINT)" ]; then echo "✓ found at $(GOLANGCI_LINT)"; else echo "✗ not found"; fi
	@echo -n "gosec: "; if [ -x "$(GOSEC)" ]; then echo "✓ found at $(GOSEC)"; else echo "✗ not found"; fi
	@echo -n "govulncheck: "; if [ -x "$(GOVULNCHECK)" ]; then echo "✓ found at $(GOVULNCHECK)"; else echo "✗ not found"; fi
	@echo -n "gofumpt: "; if [ -x "$(GOFUMPT)" ]; then echo "✓ found at $(GOFUMPT)"; else echo "✗ not found"; fi

## Security commands
security: ## Run security checks
	@echo "Running security checks..."
	@if [ -x "$(GOSEC)" ]; then \
		$(GOSEC) ./...; \
	else \
		echo "gosec not found. Installing..."; \
		$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest; \
		$(GOBIN)/gosec ./...; \
	fi

security-install: ## Install security tools
	@echo "Installing security tools..."
	$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest

vuln-check: ## Check for known vulnerabilities
	@echo "Checking for vulnerabilities..."
	@if [ -x "$(GOVULNCHECK)" ]; then \
		$(GOVULNCHECK) ./...; \
	else \
		echo "govulncheck not found. Installing..."; \
		$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest; \
		$(GOBIN)/govulncheck ./...; \
	fi

vulncheck: ## Check for known vulnerabilities (alias for vuln-check)
	@$(MAKE) vuln-check

vuln-install: ## Install vulnerability checker
	@echo "Installing vulnerability checker..."
	$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest

## Utility commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BINARY_NAME)
	rm -rf $(DIST_DIR)
	rm -rf $(COVERAGE_DIR)

format: ## Format Go code
	@echo "Formatting code..."
	@if [ -x "$(GOFUMPT)" ]; then \
		$(GOFUMPT) -w .; \
	else \
		echo "gofumpt not found. Installing..."; \
		$(GOCMD) install mvdan.cc/gofumpt@latest; \
		$(GOBIN)/gofumpt -w .; \
	fi

quality: ## Run all quality checks (format, lint, security, vulnerabilities)
	@echo "Running comprehensive quality checks..."
	@$(MAKE) format
	@$(MAKE) lint
	@$(MAKE) security
	@$(MAKE) vulncheck
	@echo "All quality checks completed ✓"

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@if [ -x "$(GOLANGCI_LINT)" ]; then \
		$(GOLANGCI_LINT) run; \
	else \
		echo "golangci-lint not found. Installing..."; \
		$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		$(GOBIN)/golangci-lint run; \
	fi

lint-install: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Development commands
dev: ## Run in development mode
	@echo "Running in development mode..."
	$(GOCMD) run ./cmd/memory-calculator --help

dev-test: ## Run with test parameters
	@echo "Running with test parameters..."
	$(GOCMD) run ./cmd/memory-calculator --total-memory 2G --thread-count 250

install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(GOBIN)..."
	cp $(BINARY_NAME) $(GOBIN)/

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

## Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm $(BINARY_NAME):latest

docker-test: ## Test Docker container with memory limit
	@echo "Testing Docker container with 2G memory limit..."
	docker run --rm --memory=2g $(BINARY_NAME):latest

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
