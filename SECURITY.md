# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

The JVM Memory Calculator team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings, and will make every effort to acknowledge your contributions.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

#### Email
Send an email to: **contact@patbaumgartner.com**

Please include:
- Description of the vulnerability
- Steps to reproduce the issue
- Possible impact
- Any suggested fixes (if available)

#### Response Time
- We will acknowledge receipt of your vulnerability report within **48 hours**
- We will provide a more detailed response within **72 hours** indicating the next steps
- We will keep you informed of the progress towards a fix and full announcement

### What to Expect

After submitting a report, you can expect:

1. **Confirmation** - We'll confirm receipt of your report
2. **Assessment** - We'll assess the vulnerability and determine severity
3. **Fix Development** - We'll work on a fix (if needed)
4. **Release** - We'll release the fix and credit you (if desired)
5. **Disclosure** - We'll coordinate public disclosure

### Security Update Process

1. **Vulnerability Assessment** - Determine severity and impact
2. **Fix Development** - Create and test security patch
3. **Release Preparation** - Prepare new version with security fix
4. **Coordinated Disclosure** - Release fix and security advisory
5. **User Notification** - Notify users via GitHub releases and documentation

## Security Best Practices

When using the JVM Memory Calculator:

### Container Security
- Run with non-root user (our Docker image does this by default)
- Use read-only file systems where possible
- Limit container capabilities to minimum required
- Regular container base image updates

### Input Validation
- The calculator validates all memory input formats
- Invalid inputs are rejected with clear error messages
- No user input is executed as shell commands

### Memory Safety
- Written in Go with memory safety guarantees
- No buffer overflows or memory corruption vulnerabilities
- Static binary with minimal attack surface

## Dependencies

We use minimal dependencies and keep them updated:
- Primary dependency: `github.com/paketo-buildpacks/libjvm`
- Dependabot automatically creates PRs for dependency updates
- All dependencies are reviewed for security issues

## Security Features

- **Input Sanitization**: All user inputs are validated and sanitized
- **No Code Execution**: Calculator doesn't execute user-provided code
- **Minimal Privileges**: Runs with minimal required permissions
- **Container Ready**: Secure container deployment patterns
- **Dependency Scanning**: Automated dependency vulnerability scanning

## Disclosure Policy

- **Private Disclosure**: Security issues are first disclosed privately to maintainers
- **Fix Development**: Security fixes are developed privately
- **Coordinated Release**: Public disclosure happens with fix release
- **Credit**: Security researchers receive credit (if desired)

## Contact

For security-related questions or concerns:
- **Email**: contact@patbaumgartner.com
- **Scope**: JVM Memory Calculator security issues only

Thank you for helping keep JVM Memory Calculator and our users safe!
