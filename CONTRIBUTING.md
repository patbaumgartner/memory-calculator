# Contributing

Contributions are welcome. This project is small enough that every change should be easy to reason
about, test locally and review as a focused diff.

## Set up

Requirements:

- Go 1.25 or later
- Git
- GNU Make
- Optional: Docker, `golangci-lint`, `gosec`, `govulncheck`

```bash
git clone https://github.com/patbaumgartner/memory-calculator.git
cd memory-calculator
go mod download   # there are currently no external modules
make test
make build
./memory-calculator --total-memory=2G --loaded-class-count=5000
```

## Project map

| Path | Responsibility |
|------|----------------|
| `cmd/memory-calculator` | CLI flags, exits, error reporting |
| `internal/config` | flags and environment configuration |
| `internal/calculator` | orchestration and typed result |
| `internal/calc` | memory algorithm and JVM option grammar |
| `internal/cgroups`, `internal/host` | memory-limit detection |
| `internal/count` | class counting |
| `internal/parser`, `internal/memory` | option and size parsing |
| `internal/display` | human and quiet output |
| `pkg/errors` | importable structured errors |

Read [ARCHITECTURE.md](ARCHITECTURE.md) before changing the algorithm or detection order.

## Workflow

1. Create a branch from `main`.
2. Reproduce the problem with a failing test.
3. Make the smallest change that fixes the root cause.
4. Run the focused test, then the full suite with the race detector.
5. Run the smoke test through the real binary.
6. Update README/API/architecture docs when the public contract changes.
7. Keep commits small, green and independently reviewable.

```bash
go test -count=1 ./internal/calc/ -run TestName
go test -count=1 -race ./...
./test-local.sh
make quality
```

The repository normalizes text files to LF through `.gitattributes`. `gofmt -l .` must produce no
output.

## Make targets

Run `make help` for the authoritative list. The most useful targets are:

```bash
make build          # current platform
make build-all      # linux/darwin, amd64/arm64
make test           # race-enabled package tests
make integration    # real-binary subprocess tests
make coverage       # coverage profile and report
make quality        # format, lint, security and vulnerability checks
make docker-build   # build the release image
make release-check  # pre-release checks
```

## Testing standards

A meaningful change needs a test at the right level:

- **Parser or calculation change**: table-driven unit tests for accepted and rejected input.
- **Bug fix**: a regression test reproducing the original failure.
- **Memory arithmetic**: assert the invariant (regions fit the budget), not just one expected number.
- **Cgroup behavior**: build an on-disk fixture tree; do not make tests depend on the host cgroup.
- **CLI behavior**: extend `integration_test.go` or `test-local.sh` and assert exit status, stdout and
  stderr separately.
- **Documentation-only change**: verify every command and numeric claim against the repository.

Use `t.Setenv`, `t.TempDir` and subtests. Do not call `t.Parallel` in tests that mutate package globals
or process-wide state. Never delete a failing test or reduce its assertion to make CI green.

Current total statement coverage is 84.2%. Coverage is a signal, not the objective: `internal/count`
has the lowest percentage because archive and I/O error paths are expensive to enumerate, while the
safety-critical calculator and detectors are above 85% and carry invariant tests.

## Code style

- Follow idiomatic Go and the existing package boundaries.
- Return errors with context and preserve causes with `%w`.
- Prefer typed structs over `map[string]string` when values have a known schema.
- Avoid global mutable state; environment variables belong at the application edge.
- Validate untrusted input at that edge and validate domain invariants in the domain package too.
- Keep comments for decisions and invariants the code cannot express; do not narrate obvious code.
- Do not add a dependency for behavior the standard library implements clearly.
- Do not introduce a second implementation of the calculation behind a build tag.

The configured linters include `govet`, `staticcheck`, `gosec`, `errcheck`, `revive`, `gocyclo`,
`goconst`, `misspell`, `unused`, `ineffassign` and a 120-character line-length check.

## Commit messages

Use conventional commits:

```
fix(cgroups): honour ancestor memory limits
feat(cli): add a machine-readable output format
test(calc): cover size overflow
docs: correct memory allocation formula
refactor(calculator): return a typed result
```

A good commit explains why the behavior changes, includes its tests and leaves the repository green.
Do not mix an unrelated cleanup into a bug fix.

## Pull requests

A pull request should state:

- the problem and its impact;
- the root cause;
- the chosen fix and trade-offs;
- how it was tested;
- any compatibility or release note implications.

Before opening one:

```bash
gofmt -l .                 # no output
make test
./test-local.sh
make quality
```

CI must pass on the supported Go version and all release targets.

## Compatibility

The public contract is the CLI, its environment variables, stdout/stderr behavior, exit codes and
release artifact names. Packages under `internal/` are intentionally not importable.

Behavior that can prevent an OOM kill may justify a compatibility break, but call it out prominently
in the changelog. For an artifact rename, preserve an alias for a transition release when practical.

## Security

Do not report vulnerabilities in a public issue. Follow [SECURITY.md](SECURITY.md). For routine code,
consider:

- integer overflow in size arithmetic;
- archive decompression limits;
- paths derived from user input;
- shell interpolation in examples and CI;
- whether a failure defaults safely or silently continues with a wrong budget.

## License

By contributing, you agree that your work is licensed under the repository's MIT license.
