#!/bin/bash
# Local GitHub Actions test simulation
# This script simulates the GitHub Actions environment locally

set -e

echo "🚀 Simulating GitHub Actions environment locally..."

# Set environment variables similar to GitHub Actions
export CGO_ENABLED=1
export GO_VERSION="1.24.5"

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
go build -o test-binary ./cmd/memory-calculator
./test-binary --version
rm test-binary

echo ""
echo "📦 Download dependencies..."
go mod download

echo ""
echo "🧪 Running tests with race detection..."
go test -v -race -timeout=10m -coverprofile=coverage.out ./...

echo ""
echo "📊 Generate coverage report..."
go tool cover -func=coverage.out
echo "Coverage Summary:"
go tool cover -func=coverage.out | tail -1

echo ""
echo "✅ All steps completed successfully!"
