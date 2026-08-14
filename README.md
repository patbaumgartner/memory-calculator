# JVM Memory Calculator

[![Build Status](https://github.com/patbaumgartner/memory-calculator/actions/workflows/build.yml/badge.svg)](https://github.com/patbaumgartner/memory-calculator/actions/workflows/build.yml)
[![Coverage](https://codecov.io/gh/patbaumgartner/memory-calculator/branch/main/graph/badge.svg)](https://codecov.io/gh/patbaumgartner/memory-calculator)
[![Go Report Card](https://goreportcard.com/badge/github.com/patbaumgartner/memory-calculator)](https://goreportcard.com/report/github.com/patbaumgartner/memory-calculator)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/patbaumgartner/memory-calculator)](https://github.com/patbaumgartner/memory-calculator/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/patbaumgartner/memory-calculator)](https://go.dev/)
[![License](https://img.shields.io/github/license/patbaumgartner/memory-calculator)](LICENSE)

Sizes a JVM to the container it runs in. It reads the container's memory limit, divides that budget
across the JVM's memory regions, and prints the resulting options as a `JAVA_TOOL_OPTIONS` string.

Compatible with the Paketo buildpack environment contract (`BPL_JVM_*` / `BPI_*`), so it can stand in
for the buildpack memory calculator outside a buildpack. Written in Go with **no external
dependencies**.

## Why

A JVM that cannot see its container limit sizes itself against the host. In a 512 MiB container on a
64 GiB host, the default heap is chosen from 64 GiB and the kernel kills the process. This tool
computes a heap that fits the limit actually in force, including limits set on an ancestor cgroup.

## Quick start

```bash
# Download
curl -L https://github.com/patbaumgartner/memory-calculator/releases/latest/download/memory-calculator-linux-amd64 -o memory-calculator
chmod +x memory-calculator

# Let it detect the container limit
export JAVA_TOOL_OPTIONS="$(./memory-calculator --quiet)"
java -jar app.jar
```

A child process cannot change its parent's environment, so the calculator prints the options and you
assign them. `--quiet` writes exactly the options to stdout, with no trailing newline and no
decoration, so it is safe to capture. Diagnostics always go to stderr, so a failure never leaves you
with a silently empty variable.

### Build from source

```bash
git clone https://github.com/patbaumgartner/memory-calculator.git
cd memory-calculator
make build
```

Requires Go 1.25 or later. `make build-all` cross-compiles for linux and darwin on amd64 and arm64.

## Usage

```bash
# Detect the limit automatically
./memory-calculator

# State the budget explicitly
./memory-calculator --total-memory 2G --thread-count 300

# Only the options, for scripting
./memory-calculator --total-memory 1G --quiet

# Skip JAR scanning by giving the class count directly
./memory-calculator --total-memory 4G --loaded-class-count 50000 --head-room 15
```

### Output

```
==================================================
JVM Memory Configuration
==================================================
Total Memory:     2.00 GB
Thread Count:     250
Loaded Classes:   35000
Head Room:        5%
Application Path: /app

Calculated JVM Arguments:
------------------------------
Max Heap Size:         1268380K
Thread Stack Size:     1M
Max Metaspace Size:    211914K
Code Cache Size:       240M
Direct Memory Size:    10M

Complete JVM Options:
------------------------------
JAVA_TOOL_OPTIONS=-XX:MaxDirectMemorySize=10M -Xmx1268380K -XX:MaxMetaspaceSize=211914K -XX:ReservedCodeCacheSize=240M -Xss1M
```

With `--quiet`, only the final options line is printed.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--total-memory` | string | auto-detect | Memory budget, e.g. `1G`, `512M`, `2.5GB`, `1073741824` |
| `--thread-count` | int | 250 | Threads to reserve stack space for |
| `--loaded-class-count` | int | counted from `--path` | Classes to size metaspace for |
| `--head-room` | int | 0 | Percentage of total memory to leave unallocated (0–99) |
| `--path` | string | `/app` | Directory scanned for JARs to estimate the class count || `--quiet` | bool | false | Print only the JVM options |
| `--version` | bool | false | Print version information |
| `--help` | bool | false | Print usage |

Sizes accept `B`, `K`/`KB`, `M`/`MB`, `G`/`GB`, `T`/`TB`, case-insensitive, with decimals such as
`1.5G`. Units are binary (1 KB = 1024 bytes). A value that cannot be parsed is a fatal error, not a
fallback to auto-detection.

### Environment variables

The Paketo buildpack contract is supported, so the tool can be dropped into an existing buildpack
environment. Command-line flags take precedence.

| Variable | Equivalent flag |
|----------|-----------------|
| `BPL_JVM_TOTAL_MEMORY` | `--total-memory` |
| `BPL_JVM_THREAD_COUNT` | `--thread-count` |
| `BPL_JVM_LOADED_CLASS_COUNT` | `--loaded-class-count` |
| `BPL_JVM_HEAD_ROOM` | `--head-room` |
| `BPI_APPLICATION_PATH` | `--path` |
| `BPI_JVM_CLASS_COUNT` | base JVM class count, default 1000 |
| `BPI_CLASS_ADJUSTMENT_FACTOR` | percentage applied to the counted classes, default 100 |
| `BPI_CLASS_STATIC_ADJUSTMENT` | classes added before the adjustment factor, default 0 |

`BPL_JVM_HEADROOM` is accepted for backwards compatibility, warns, and is ignored when
`BPL_JVM_HEAD_ROOM` is also set.

Options already present in `JAVA_TOOL_OPTIONS` are preserved: the calculator will not compute a
region the caller has already set. If it recognises one of its options but cannot parse the value —
`-Xmx1.5G`, for example, which HotSpot itself rejects — it fails rather than emitting a second,
conflicting `-Xmx`.

## How the budget is divided

```
┌─────────────────────────────────────┐
│           Total Memory              │
├─────────────────────────────────────┤
│ Head Room        configurable %     │
├─────────────────────────────────────┤
│ Thread Stacks    threads × 1 MiB    │
├─────────────────────────────────────┤
│ Metaspace        14 MB + 5800 B/cls │
├─────────────────────────────────────┤
│ Code Cache       240 MiB            │
├─────────────────────────────────────┤
│ Direct Memory    10 MiB             │
├─────────────────────────────────────┤
│ Heap             what remains       │
└─────────────────────────────────────┘
```

Everything except the heap is sized first; the heap receives the remainder. If the fixed regions do
not fit in the budget, the calculator fails with a breakdown rather than emitting an unusable
configuration.

### Finding the memory limit

In order, stopping at the first that yields a limit:

1. `--total-memory` / `BPL_JVM_TOTAL_MEMORY`
2. cgroup v2 — `/sys/fs/cgroup/memory.max`
3. cgroup v1 — `/sys/fs/cgroup/memory/memory.limit_in_bytes`
4. Host memory — `MemAvailable` from `/proc/meminfo`, Linux only
5. 1 GiB, with a warning

For both cgroup versions the detector resolves the process's own cgroup from `/proc/self/cgroup` and
walks up to the mount root, taking the **smallest** limit it finds. A container's own cgroup is often
unlimited while its pod cgroup is capped, and only the walk finds that cap.

Values meaning "no limit" are ignored rather than treated as enormous budgets: `max`, zero or
negative, anything at or above 4 EiB, and values too large for a signed 64-bit integer. Limits above
64 TiB are clamped to 64 TiB.

Host memory uses `MemAvailable` rather than `MemTotal`: with no cgroup limit the process shares the
machine with everything else on it, and sizing a heap against total RAM overcommits a busy host.

### Estimating the class count

Metaspace is sized from the number of classes. Unless `--loaded-class-count` is given, the calculator
walks `--path`, counts entries ending in `.class`, `.classdata`, `.clj`, `.groovy` and `.kts` in JARs
(including one level of nested JARs) and on disk, adds `BPI_JVM_CLASS_COUNT` for the JVM's own
classes, applies `BPI_CLASS_STATIC_ADJUSTMENT` and `BPI_CLASS_ADJUSTMENT_FACTOR`, and multiplies by a
0.35 load factor. Nested JAR extraction is capped at 100 MB to bound decompression.

If `--path` is not given and the default `/app` does not exist — the usual case outside a buildpack
image — the calculator warns and sizes metaspace from `BPI_JVM_CLASS_COUNT` alone rather than
failing. A path you name explicitly must exist; a typo there is an error, not a warning. Point
`--path` at the directory holding your JARs to get metaspace sized for your application.

## Deployment

### Docker

```dockerfile
FROM eclipse-temurin:25-jre-alpine
COPY --from=patbaumgartner/memory-calculator:alpine /usr/local/bin/memory-calculator /usr/local/bin/
COPY app.jar /app/app.jar
CMD export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet --path /app)" && exec java -jar /app/app.jar
```

Building in a multi-stage image instead:

```dockerfile
FROM golang:1.25-alpine AS builder
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o memory-calculator ./cmd/memory-calculator

FROM eclipse-temurin:25-jre-alpine
COPY --from=builder /build/memory-calculator /usr/local/bin/
COPY app.jar /app/app.jar
CMD export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet --path /app)" && exec java -jar /app/app.jar
```

### Kubernetes

The calculator reads the limit from the container's cgroup, so it needs no configuration beyond the
limit itself:

```yaml
spec:
  containers:
  - name: app
    image: myapp:latest
    resources:
      limits:
        memory: "1Gi"
    command: ["/bin/sh", "-c"]
    args:
    - export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)" && exec java -jar /app/app.jar
```

More patterns, including shell helpers and troubleshooting, are in [USAGE_GUIDE.md](USAGE_GUIDE.md)
and [examples/](examples/).

## Exit codes

`0` on success, `1` on any failure, `2` on a usage error such as an unknown flag. Failures print a
diagnostic to stderr naming the offending setting. Invalid configuration, a memory budget too small
for the fixed regions, an application path you named that does not exist, and an unparseable JVM
option are all failures.

Warnings are different from failures: when the memory limit cannot be detected, when it is clamped,
or when the default application path does not exist, the calculator prints a warning to stderr and
still emits usable options. Warnings are printed even under `--quiet`, because they mean the options
are not the ones you asked for.

## Development

```bash
make test        # run the test suite
make coverage    # run tests with a coverage report
make quality     # format, lint, security scan, vulnerability check
make build       # build for the current platform
make help        # list all targets
./test-local.sh  # build the binary and smoke-test its behaviour
```

Total statement coverage is 84.2%. CI runs the suite with the race detector, `golangci-lint`,
`gosec` and `govulncheck` on every push and pull request.

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines and
[ARCHITECTURE.md](ARCHITECTURE.md) for how the packages fit together.

## Documentation

- **[USAGE_GUIDE.md](USAGE_GUIDE.md)** — integration patterns and troubleshooting
- **[API.md](API.md)** — full interface reference
- **[ARCHITECTURE.md](ARCHITECTURE.md)** — design and package layout
- **[CONTRIBUTING.md](CONTRIBUTING.md)** — development workflow
- **[SECURITY.md](SECURITY.md)** — reporting vulnerabilities
- **[CHANGELOG.md](CHANGELOG.md)** — release history

## License

MIT — see [LICENSE](LICENSE).

## Acknowledgments

- [Paketo Buildpacks](https://paketo.io/) for the original memory calculator this tool mirrors
- [Temurin](https://adoptium.net/) and [Liberica](https://bell-sw.com/) JDK teams
