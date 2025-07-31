# Contributing to JVM Memory Calculator

Thank you for your interest in contributing to the JVM Memory Calculator! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

This project follows a simple code of conduct:

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Assume good intentions

## Getting Started

### Prerequisites

- Go 1.24.5 or later
- Git
- Make (optional, but recommended)

### Setting up the Development Environment

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/yourusername/memory-calculator.git
   cd memory-calculator
   ```

3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/patbaumgartner/memory-calculator.git
   ```

4. Install dependencies:
   ```bash
   make deps
   # or
   go mod download
   ```

5. Run tests to ensure everything works:
   ```bash
   make test
   # or
   go test -v ./...
   ```

## Development Workflow

### Branch Naming

Use descriptive branch names:
- `feature/add-new-memory-units` - for new features
- `fix/cgroups-detection-bug` - for bug fixes
- `docs/update-readme` - for documentation changes
- `refactor/cleanup-parser` - for code improvements

### Making Changes

1. Create a new branch from `main`:
   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the [code style guidelines](#code-style)

3. Add tests for new functionality:
   ```bash
   # Run tests frequently during development
   make test
   
   # Check coverage
   make coverage
   ```

4. Update documentation if needed

5. Commit your changes with descriptive commit messages:
   ```bash
   git add .
   git commit -m "feat: add support for PB/petabyte memory units"
   ```

### Commit Message Format

We follow the [Conventional Commits](https://conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting (no logic changes)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```bash
feat: add support for petabyte memory units
fix: resolve cgroups v2 detection on newer kernels  
docs: update installation instructions
test: add integration tests for quiet mode
refactor: simplify memory parsing logic
```

## Testing

### Test Types

- **Unit Tests**: Test individual functions and components
- **Integration Tests**: Test the full binary with various inputs
- **Benchmark Tests**: Performance testing for critical paths

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Generate HTML coverage report
make coverage-html

# Run specific test files
go test -v ./memory_test.go
go test -run TestMemoryParsing -v
```

### Writing Tests

1. **Test file naming**: `*_test.go`
2. **Test function naming**: `TestFunctionName`
3. **Use table-driven tests** for multiple test cases:

```go
func TestParseMemoryString(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int64
        hasError bool
    }{
        {"Valid GB", "2G", 2147483648, false},
        {"Invalid unit", "2X", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := parseMemoryString(tt.input)
            if tt.hasError && err == nil {
                t.Errorf("Expected error but got none")
            }
            if !tt.hasError && result != tt.expected {
                t.Errorf("Expected %d, got %d", tt.expected, result)
            }
        })
    }
}
```

### Test Coverage Requirements

- New features must include tests
- Aim for >90% coverage on new code
- Critical paths must have comprehensive test coverage
- Integration tests for CLI functionality

## Code Style

### Go Standards

- Follow standard Go conventions
- Use `gofmt` for formatting
- Use `golint` for linting
- Write clear, self-documenting code

### Formatting

```bash
# Format code
make format
# or
gofmt -s -w .

# Run linter (if installed)
make lint
# or  
golangci-lint run
```

### Documentation

- Add comments for exported functions
- Include usage examples in complex functions
- Update README.md for new features
- Use godoc-style comments:

```go
// ParseMemoryString converts a memory string (e.g., "2G", "512M") to bytes.
// Supported units: B, K, KB, M, MB, G, GB, T, TB (case insensitive).
// Returns the memory in bytes and an error if the format is invalid.
func ParseMemoryString(memory string) (int64, error) {
    // implementation
}
```

## Submitting Changes

### Pull Request Process

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub:
   - GitHub will automatically use our pull request template
   - Fill out all sections of the template
   - Use a descriptive title following [Conventional Commits](https://conventionalcommits.org/)
   - Reference any related issues using keywords (e.g., "Fixes #123")
   - Provide a clear description of changes
   - Include testing information and screenshots if applicable

3. **Automated GitHub Actions** will run:
   - ✅ **Test Suite**: Complete test suite with race detection
   - ✅ **Coverage**: Coverage analysis and reporting
   - ✅ **Linting**: golangci-lint with project configuration
   - ✅ **Security**: gosec security scanning
   - ✅ **Vulnerabilities**: govulncheck vulnerability scanning
   - ✅ **Cross-Platform**: Build verification for all platforms

4. **Pull Request Template** includes:
   - Description and type of change checkboxes
   - Testing checklist (unit tests, integration tests, manual testing)
   - Documentation update confirmation
   - Breaking changes description (if applicable)
   - Performance impact assessment

### Review Process

1. **Automated Checks**: All GitHub Actions must pass ✅
   - Test suite must have 100% pass rate
   - Coverage should not decrease significantly
   - No linting errors or security issues
   - All platforms must build successfully

2. **Code Review**: Maintainers will review:
   - Code quality and adherence to Go standards
   - Test coverage for new functionality
   - Documentation updates
   - Breaking change compatibility

3. **Feedback**: Address any requested changes:
   - Push additional commits to your branch
   - GitHub Actions will re-run automatically
   - Respond to review comments

4. **Approval & Merge**: Once approved:
   - Squash and merge is preferred for clean history
   - Commit message should follow [Conventional Commits](https://conventionalcommits.org/)
   - PR will be automatically closed

### Issue Reporting

Use GitHub Issues with our templates:

- **🐛 Bug Report**: Use the bug report template with:
  - Environment details (OS, architecture, version)
  - Steps to reproduce
  - Expected vs actual behavior
  - Command used and output

- **✨ Feature Request**: Use the feature request template with:
  - Clear use case description
  - Proposed implementation approach
  - Alternatives considered

- **❓ Question**: Use the question template for:
  - Usage questions
  - Clarifications about behavior
  - General support requests

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
- Major: Breaking changes
- Minor: New features (backward compatible)  
- Patch: Bug fixes (backward compatible)

### Automated Release Workflow (Maintainers Only)

Releases are **fully automated** using GitHub Actions:

1. **Prepare Release**:
   ```bash
   # Ensure all changes are merged to main
   git checkout main
   git pull origin main
   
   # Verify everything is ready
   make release-check
   make test
   make quality
   ```

2. **Create Release Tag**:
   ```bash
   # Create and push version tag
   git tag v1.2.0
   git push origin v1.2.0
   ```

3. **Automated GitHub Actions**:
   - ✅ Runs complete test suite
   - ✅ Performs security and vulnerability scans
   - ✅ Builds binaries for all platforms (Linux/macOS, amd64/arm64)
   - ✅ Creates GitHub release with auto-generated notes
   - ✅ Uploads all artifacts with SHA256 checksums
   - ✅ Builds and pushes multi-arch Docker images
   - ✅ Updates package registries

4. **Release Artifacts**:
   - `memory-calculator-linux-amd64` - Linux x86_64 binary
   - `memory-calculator-linux-arm64` - Linux ARM64 binary
   - `memory-calculator-darwin-amd64` - macOS Intel binary
   - `memory-calculator-darwin-arm64` - macOS Apple Silicon binary
   - `checksums.txt` - SHA256 checksums
   - Docker images on Docker Hub

### Pre-Release Checklist

Before creating a release tag:

- [ ] All PRs merged and main branch updated
- [ ] `make test` passes
- [ ] `make quality` passes (format, lint, security, vulnerabilities)
- [ ] CHANGELOG.md updated with new version
- [ ] Version numbers updated in relevant files
- [ ] Breaking changes documented
- [ ] Release notes prepared (or rely on auto-generation)

### Emergency Releases

For critical bug fixes:

1. Create hotfix branch from latest release tag
2. Apply minimal fix
3. Create new patch version tag
4. GitHub Actions will handle the rest

### Release Communication

Releases are automatically announced via:
- GitHub Releases page with detailed notes
- Docker Hub with updated images
- Package registries (when configured)

## Getting Help

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/patbaumgartner/memory-calculator/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/patbaumgartner/memory-calculator/discussions)  
- **Documentation**: Check the [README](README.md) and inline code documentation
- **Contact**: Patrick Baumgartner <contact@patbaumgartner.com>

## Recognition

Contributors will be:
- Listed in the project's contributors
- Mentioned in release notes for significant contributions
- Invited to be maintainers for sustained, quality contributions

Thank you for contributing to the JVM Memory Calculator! 🎉
