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
   make test-coverage
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
make test-coverage

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
   - Use a descriptive title
   - Reference any related issues
   - Provide a clear description of changes
   - Include testing information

3. **Pull Request Template**:
   ```markdown
   ## Description
   Brief description of the changes

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature  
   - [ ] Documentation update
   - [ ] Performance improvement
   - [ ] Code refactoring

   ## Testing
   - [ ] Tests pass locally
   - [ ] Added tests for new functionality
   - [ ] Updated documentation

   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-review completed
   - [ ] Comments added to complex code
   - [ ] No breaking changes (or marked as such)
   ```

### Review Process

1. **Automated Checks**: GitHub Actions will run tests and builds
2. **Code Review**: Maintainers will review your code
3. **Feedback**: Address any requested changes
4. **Approval**: Once approved, your PR will be merged

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
- Major: Breaking changes
- Minor: New features (backward compatible)  
- Patch: Bug fixes (backward compatible)

### Release Workflow

1. **Prepare Release**:
   ```bash
   # Ensure main is up to date
   git checkout main
   git pull upstream main
   
   # Check release readiness
   make release-check
   ```

2. **Create Release**:
   - Create a new tag: `git tag v1.2.3`
   - Push tag: `git push upstream v1.2.3`
   - GitHub Actions will automatically build and create release

3. **Post-Release**:
   - Update documentation if needed
   - Announce in discussions/issues if significant

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
