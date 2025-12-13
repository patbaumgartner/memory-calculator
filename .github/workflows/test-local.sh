#!/bin/bash
# Local GitHub Actions test simulation
# This script simulates the GitHub Actions environment locally

set -e

echo "🚀 Simulating GitHub Actions environment locally..."

# Set environment variables similar to GitHub Actions
export CGO_ENABLED=1
export GO_VERSION="1.25.5"

echo "📋 Environment Information:"
echo "Go version: $(go version)"
echo "Working directory: $(pwd)"
echo "CGO_ENABLED: $CGO_ENABLED"

echo ""
echo "📁 Directory structure:"
ls -la

echo ""
echo "🔍 Verifying module structure..."
go list -m
go list ./...

echo ""
echo "🏗️ Testing build process..."
CGO_ENABLED=0 go build -o test-binary ./cmd/memory-calculator
./test-binary --version
rm test-binary

echo ""
echo "🌍 Testing cross-compilation..."
echo "Building for linux/arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o test-binary-arm64 ./cmd/memory-calculator
echo "✅ ARM64 build successful"
rm test-binary-arm64

echo "Building for darwin/amd64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o test-binary-darwin ./cmd/memory-calculator
echo "✅ Darwin build successful"
rm test-binary-darwin

echo ""
echo "📦 Download dependencies..."
go mod download

echo ""
echo "🧪 Running tests with race detection..."
# Set up test environment with a mock application directory
mkdir -p testdata/app
echo "mock.jar" > testdata/app/mock.jar

# Run tests with proper environment
BPI_APPLICATION_PATH=$(pwd)/testdata/app go test -v -race -timeout=10m -coverprofile=coverage.out ./...

echo ""
echo "📊 Generate coverage report..."
go tool cover -func=coverage.out
echo "Coverage Summary:"
go tool cover -func=coverage.out | tail -1

echo ""
echo "🔧 Testing both build variants..."
echo "Building standard variant..."
go build -o memory-calculator-test-standard ./cmd/memory-calculator
echo "Building minimal variant..."
go build -tags minimal -o memory-calculator-test-minimal ./cmd/memory-calculator

echo ""
echo "Testing standard build variant:"
BPI_APPLICATION_PATH=$(pwd)/testdata/app ./memory-calculator-test-standard --total-memory 1G --thread-count 50 --quiet

echo ""
echo "Testing minimal build variant:"
BPI_APPLICATION_PATH=$(pwd)/testdata/app ./memory-calculator-test-minimal --total-memory 1G --thread-count 50 --quiet

echo ""
echo "Comparing binary sizes:"
ls -lh memory-calculator-test-* | awk '{print $5 "\t" $9}'

# Clean up test binaries
rm -f memory-calculator-test-standard memory-calculator-test-minimal

echo ""
echo "✅ All steps completed successfully!"
