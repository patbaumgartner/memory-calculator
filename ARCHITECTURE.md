# Architecture

How the JVM Memory Calculator is put together, and why.

## Table of contents

- [The problem](#the-problem)
- [Flow](#flow)
- [Packages](#packages)
- [The calculation](#the-calculation)
- [Detecting the memory limit](#detecting-the-memory-limit)
- [Counting classes](#counting-classes)
- [Design decisions](#design-decisions)
- [Testing](#testing)
- [Security](#security)

## The problem

A JVM sizes itself from what it believes the machine has. Inside a container that belief is wrong
unless something tells it otherwise: the JVM may see the host's memory while the kernel enforces a
much smaller cgroup limit. The gap between those two numbers is where OOM kills happen.

This tool closes the gap. It determines the memory budget actually in force, divides it across the
JVM's memory regions, and emits the corresponding options for `JAVA_TOOL_OPTIONS`.

Two properties matter more than anything else:

1. **The budget must be the real one.** Reading the host's memory when a cgroup limit applies, or
   reading a "no limit" sentinel as a literal number, produces a confidently wrong configuration.
2. **The regions must fit the budget.** The sum of every region must never exceed total memory. This
   is asserted directly by a test that sweeps combinations of inputs.

Where the tool cannot be sure, it fails loudly. A wrong answer here is discovered in production, at
which point the process is already dead; an error at startup is discovered immediately.

## Flow

```
                    ┌──────────────────────┐
   flags + env ───► │  internal/config     │  parse, apply defaults, validate
                    └──────────┬───────────┘
                               │ Config
                    ┌──────────▼───────────┐
                    │ internal/calculator  │  orchestration
                    └──────────┬───────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
     ┌────────────────┐ ┌────────────┐ ┌────────────────┐
     │internal/cgroups│ │internal/   │ │internal/parser │
     │  + host        │ │  count     │ │                │
     │ memory budget  │ │class count │ │ split existing │
     │                │ │            │ │ JVM options    │
     └────────┬───────┘ └─────┬──────┘ └───────┬────────┘
              └───────────────┼────────────────┘
                              ▼
                    ┌──────────────────────┐
                    │   internal/calc      │  divide the budget
                    └──────────┬───────────┘
                               │ Result
                    ┌──────────▼───────────┐
                    │  internal/display    │  render
                    └──────────────────────┘
```

Configuration is read once, at the edge, in `internal/config`. Everything downstream receives typed
values. `calculator.Execute` returns a `Result` carrying the options string, the budget used, the
calculated regions and the effective thread, class and head room values, so the display layer reads
fields rather than re-parsing the string it was just handed.

## Packages

| Package | Responsibility |
|---------|----------------|
| `cmd/memory-calculator` | Flag definitions, exit codes, error reporting |
| `internal/config` | Loads flags and environment variables, applies defaults, validates |
| `internal/calculator` | Orchestrates detection, counting and calculation |
| `internal/calc` | The memory algorithm, JVM option grammar, `Size` type |
| `internal/cgroups` | Finds the memory limit in force |
| `internal/host` | Reads host memory from `/proc/meminfo` |
| `internal/count` | Counts classes in JARs and on disk |
| `internal/parser` | Splits a JVM option string into arguments |
| `internal/memory` | Parses and formats human-facing sizes |
| `internal/display` | Renders the report and the quiet output |
| `internal/logger` | Quiet-aware informational logging to stderr |
| `pkg/errors` | Structured error types; the only importable package |

Dependencies point inward and form no cycles. `internal/calc` depends only on `internal/parser`;
nothing in the domain reaches back out to configuration or display.

### Two size grammars, deliberately

There are two size parsers and the split is intentional:

- **`internal/memory`** parses what a *human* writes: `1.5GB`, `512m`, `2147483648`. It accepts
  decimals and two-letter units. It is used for `--total-memory` and its environment equivalent.
- **`calc.ParseSize`** parses what the *JVM* accepts inside an option: unsigned digits and at most
  one unit letter. It rejects decimals and signs, because HotSpot rejects them too.

Merging them would mean either rejecting `1.5GB` on the command line, which is convenient and
unambiguous, or accepting `-Xmx1.5G`, which the JVM refuses to start with. Each grammar matches the
contract it serves.

## The calculation

Total memory is divided in a fixed order. Everything except the heap is sized first; the heap
receives what remains.

```
head room  = total × headRoomPercent / 100
stacks     = threadCount × stackSize            (default 1 MiB each)
metaspace  = 14,000,000 + loadedClassCount × 5,800
code cache = 240 MiB
direct     = 10 MiB
heap       = total − (head room + stacks + metaspace + code cache + direct)
```

Any region the caller already set in `JAVA_TOOL_OPTIONS` is marked `UserConfigured` and used as
given, rather than calculated and re-emitted.

Three checks guard the result: the fixed regions must fit the budget, the non-heap regions must fit
the budget, and the sum of all regions must fit the budget. Failing any of them is an error carrying
the breakdown, because a configuration that does not fit cannot be silently trimmed into one that
does.

Rendered sizes always round **down** to a whole unit, so an emitted maximum can never exceed the
value that was checked against the budget.

## Detecting the memory limit

In order, stopping at the first that yields a limit:

1. `--total-memory` / `BPL_JVM_TOTAL_MEMORY`
2. cgroup v2 — `/sys/fs/cgroup/memory.max`
3. cgroup v1 — `/sys/fs/cgroup/memory/memory.limit_in_bytes`
4. Host `MemAvailable` from `/proc/meminfo`, Linux only
5. 1 GiB, with a warning

Three details carry most of the correctness:

**v2 before v1.** On a hybrid host both hierarchies are mounted, but the unified hierarchy is the one
the container runtime configures. Reading v1 first can return a stale limit.

**Walk the hierarchy.** The process's own cgroup is not necessarily the one that constrains it. A
container's cgroup is frequently unlimited while its pod cgroup is capped. The detector resolves
membership from `/proc/self/cgroup` and walks up to the mount root, taking the smallest finite limit.
Inside a cgroup namespace the container's cgroup is mounted as the root, so the walk collapses to a
single read and stays correct.

**"No limit" is not a number.** An unset v1 limit is reported as the kernel's page-counter maximum:
int64 max rounded down to a page boundary, which differs between a 4 KiB and a 64 KiB page kernel.
Matching one literal misses the other and turns "unlimited" into a 64 TiB budget. Anything at or
above 4 EiB, at or below zero, or too large for int64 is treated as no limit.

Host memory uses `MemAvailable` rather than `MemTotal`. Without a cgroup limit the process shares the
machine, and sizing a heap against total RAM overcommits a busy host. Non-Linux platforms report
nothing rather than guess, so the caller applies the documented default.

## Counting classes

Metaspace scales with loaded classes, so the class count must be estimated when not supplied.
`internal/count` walks the application path and counts entries ending in `.class`, `.classdata`,
`.clj`, `.groovy` or `.kts`, both inside JARs and on disk, descending one level into nested JARs.

The count is then adjusted: the JVM's own classes are added (`BPI_JVM_CLASS_COUNT`, default 1000),
`BPI_CLASS_STATIC_ADJUSTMENT` is applied, the total is scaled by `BPI_CLASS_ADJUSTMENT_FACTOR`, and a
0.35 load factor accounts for the fraction of shipped classes a process actually loads.

Nested JAR extraction is bounded at 100 MB so a crafted archive cannot exhaust memory.

## Design decisions

**Environment variables are an input format, not an internal transport.** The buildpack contract
(`BPL_JVM_*`, `BPI_*`) is part of the public interface and is preserved. It is read once in
`internal/config`; the calculator receives typed values rather than re-reading process state.

**One implementation of the algorithm.** A `minimal` build tag once selected alternative parsing and
class-counting implementations. They disagreed with the standard ones on real input, were never
exercised by CI, and shipped in every release. Two implementations of a correctness-critical
calculation are not worth an 8% binary saving. Removing the last regular expression instead cut 11%
from the binary with one implementation.

**Failures are loud.** Invalid input, an unusable JVM option, and a budget too small for the fixed
regions all exit non-zero with a diagnostic on stderr — including under `--quiet`, whose contract is
a clean *stdout*, not silence. The alternative is `export JAVA_TOOL_OPTIONS="$(...)"` quietly
producing an empty string.

**No dependencies.** The module has an empty `go.sum`. For a binary that runs at container startup,
in the path of every JVM launch, the supply-chain surface is worth the small amount of parsing code.

## Testing

- **Unit tests** cover each package; total statement coverage is 84.8%.
- **Property sweeps** assert the invariants: regions never exceed the budget, and a rendered size
  never re-parses to more than the value it was rendered from.
- **Fixture trees** stand in for cgroup filesystems, so v1, v2, hybrid, nested ancestors, every
  "unlimited" sentinel and malformed files are all tested without a container.
- **End-to-end tests** in `internal/calculator` assert that a 512 MiB limit yields the same heap
  whether it arrives via v2, v1, an ancestor cgroup or host memory.
- **Integration tests** in `integration_test.go` build the real binary and run it as a subprocess.
- **`test-local.sh`** builds the binary and checks the output contract, user-flag handling and that
  invalid input fails loudly.

CI runs the suite with the race detector, plus `golangci-lint`, `gosec` and `govulncheck`.

## Security

The tool reads files and writes to stdout; it opens no network connections and executes no
subprocesses.

- Cgroup and `/proc` paths are fixed constants, configurable only from tests.
- Nested JAR decompression is capped at 100 MB, bounding decompression-bomb exposure.
- Sizes are checked for 64-bit overflow before the unit multiplier is applied, so a large input
  cannot wrap into a negative budget.
- Untrusted values reach `strconv` and path joins only; nothing is interpolated into a shell.
- The release Docker image runs as a non-root user.

See [SECURITY.md](SECURITY.md) for how to report a vulnerability.
