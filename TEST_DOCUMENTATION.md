# Test Suite

The test suite covers every package used by the shipped binary. Current total statement coverage is
**84.8%**; CI also runs the suite with the race detector.

## Run it

```bash
make test                    # go test -count=1 -v -race -timeout=10m ./...
make integration             # root-package binary integration tests
make coverage                # coverage.out + per-function report
./test-local.sh              # fast behaviour smoke test

go test -count=1 ./...       # bypass the test cache
```

`make quality` adds formatting, `golangci-lint`, `gosec` and `govulncheck`.

## Layers

### Unit tests

Each internal package is tested at its own boundary:

| Package | Coverage | What is exercised |
|---------|----------|-------------------|
| `internal/calc` | 92.1% | allocation, option grammar, overflow, invariants |
| `internal/calculator` | 85.0% | orchestration, environment inputs, end-to-end results |
| `internal/cgroups` | 96.6% | v1/v2, hybrid, nested limits, unlimited sentinels |
| `internal/config` | 100.0% | defaults, validation, environment variables |
| `internal/count` | 63.0% | files, JARs, nested archives, malformed archives |
| `internal/display` | 100.0% | reports, quiet output, help and version |
| `internal/host` | 90.9% | `MemAvailable`, old-kernel fallback, malformed files |
| `internal/logger` | 100.0% | quiet and non-quiet logging |
| `internal/memory` | 98.2% | human-facing size grammar and formatting |
| `internal/parser` | 100.0% | JVM option splitting, quoting and escapes |
| `pkg/errors` | 100.0% | structured errors and unwrapping |

The root `integration_test.go` and the `cmd` package do not contribute statement coverage, because
the integration test builds and runs the real binary as a subprocess. Their behaviour is still
covered.

### Invariant tests

Examples are not enough for a memory calculator. Two tests sweep values to lock the properties the
tool depends on:

- `TestCalculateNeverExceedsTotalMemory` varies total memory, threads, classes and head room, and
  asserts every successful calculation has a positive heap and regions whose sum fits the budget.
- `TestSizeStringNeverExceedsBudget` asserts an emitted JVM size re-parses to no more than the value
  it represents, so unit rounding never grows a maximum.

### Cgroup fixture tests

Tests create temporary directory trees that look like cgroup filesystems. This is deterministic,
does not require root or Docker, and covers scenarios that are difficult to arrange on the CI host:

- finite cgroup v1 and v2 limits;
- hybrid systems, with v2 taking precedence;
- an unlimited leaf constrained by a Kubernetes-style ancestor;
- the tightest finite limit in a hierarchy;
- `max`, `-1`, 4 KiB-page and 64 KiB-page v1 sentinels, and unsigned overflow;
- malformed, empty and missing files;
- a legitimate limit above 1 TiB;
- host `MemAvailable` fallback.

`internal/calculator/detection_test.go` exercises the complete path from those fixture files to the
emitted `-Xmx`. A 512 MiB limit must produce the same heap whether it is supplied by cgroup v2,
cgroup v1, an ancestor cgroup or host memory.

### Binary integration tests

`integration_test.go` builds `./cmd/memory-calculator` into a temporary directory and runs it with
real arguments and environment variables. It verifies:

- help and version output;
- normal and quiet output;
- memory units, decimal total memory and explicit settings;
- invalid values exit non-zero and explain themselves;
- standard-output decoration does not leak into quiet mode.

Use `go test -count=1 -run TestMainIntegration .` to run only this layer.

### Local smoke test

`./test-local.sh` is the shortest pre-push check. It builds the binary and verifies:

- quiet output contains JVM options and no report decoration;
- verbose output reports the memory budget actually used;
- a user-supplied `-Xmx` is preserved and not duplicated;
- invalid total memory, JVM options and head room fail non-zero;
- errors are visible on stderr even under `--quiet`.

## Adding a test

1. Put a unit test beside the package it exercises.
2. Prefer a table when the behavior has multiple input classes.
3. For a bug, first encode the exact regression, then add a broader invariant if the root cause
   represents a family of failures.
4. Use `t.Setenv` for environment variables; it restores state automatically.
5. Use `t.TempDir` for filesystem fixtures.
6. Run `go test -count=1 -race ./...` before committing.

Do not weaken or delete a failing test to make the suite pass. A test that exposes a real mismatch
between documentation and behavior is evidence that one of them must be corrected.

## CI

`.github/workflows/build.yml` runs:

1. `gofmt` verification;
2. native and cross-platform builds;
3. `go test -race` with coverage;
4. `test-local.sh`;
5. `golangci-lint`;
6. `gosec`;
7. `govulncheck`;
8. release builds for linux/darwin on amd64/arm64;
9. Docker image smoke tests, including automatic detection in a 512 MiB container.

The legacy release filenames containing `minimal` are compatibility aliases of the tested standard
binary, not a separate build. They are deprecated and will be removed in the next major release.
