# JVM Memory Calculator - Interface Reference

`memory-calculator` is a CLI tool. Its public contract is the command line: flags,
environment variables, what it writes to stdout and stderr, and its exit code.

Everything under `internal/` is private to this module. Go forbids importing those
packages from another module, so they are not an API. The only importable package is
[`pkg/errors`](#importable-package-pkgerrors).

## Invocation

```bash
memory-calculator [flags]
```

Typical use is command substitution:

```bash
export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)"
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--total-memory` | string | auto-detected | Memory budget, e.g. `2G`, `512M`, `1.5GB`, `1073741824` |
| `--thread-count` | int | 250 | Threads to reserve stack space for |
| `--loaded-class-count` | int | counted from `--path` | Classes to size metaspace for |
| `--head-room` | int | 0 | Percentage of total memory to leave unallocated (0 to 99) |
| `--path` | string | `/app` | Directory scanned for JARs and classes |
| `--quiet` | bool | false | Print only the JVM options |
| `--version` | bool | false | Print version, build time, commit and Go version |
| `--help` | bool | false | Print usage |

Sizes accept `B`, `K`/`KB`, `M`/`MB`, `G`/`GB`, `T`/`TB`, case-insensitive, with
decimals such as `1.5G`. Units are binary (1 KB = 1024 bytes). A bare number is bytes.

## Environment variables

Every flag is initialised from its environment variable, then overwritten by an
explicitly supplied flag. Flags therefore win.

| Variable | Equivalent flag | Notes |
|----------|-----------------|-------|
| `BPL_JVM_TOTAL_MEMORY` | `--total-memory` | |
| `BPL_JVM_THREAD_COUNT` | `--thread-count` | Default 250 |
| `BPL_JVM_LOADED_CLASS_COUNT` | `--loaded-class-count` | Skips class counting when set |
| `BPL_JVM_HEAD_ROOM` | `--head-room` | Default 0 |
| `BPL_JVM_HEADROOM` |, | Deprecated. Warns, and is ignored when `BPL_JVM_HEAD_ROOM` is also set |
| `BPI_APPLICATION_PATH` | `--path` | Default `/app` |
| `BPI_JVM_CLASS_COUNT` |, | Classes attributed to the JVM itself, default 1000 |
| `BPI_CLASS_ADJUSTMENT_FACTOR` |, | Percentage applied to the total class count, default 100 |
| `BPI_CLASS_STATIC_ADJUSTMENT` |, | Classes added before the adjustment factor, default 0 |
| `JAVA_TOOL_OPTIONS` |, | Read as input. Options already present are preserved and their regions are not recalculated |

## Output

Errors and warnings always go to **stderr**, including under `--quiet`. So
`export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)"` can never silently yield
empty options: a failure is both visible and non-zero.

### `--quiet`

stdout is exactly the options string, with no trailing newline:

```
-XX:MaxDirectMemorySize=10M -Xmx1543159K -XX:MaxMetaspaceSize=41992K -XX:ReservedCodeCacheSize=240M -Xss1M
```

### Default

```
==================================================
JVM Memory Configuration
==================================================
Total Memory:     2.00 GB
Thread Count:     250
Loaded Classes:   5000
Head Room:        0%
Application Path: /app

Calculated JVM Arguments:
------------------------------
Max Heap Size:         1543159K
Thread Stack Size:     1M
Max Metaspace Size:    41992K
Code Cache Size:       240M
Direct Memory Size:    10M

Complete JVM Options:
------------------------------
JAVA_TOOL_OPTIONS=-XX:MaxDirectMemorySize=10M -Xmx1543159K -XX:MaxMetaspaceSize=41992K -XX:ReservedCodeCacheSize=240M -Xss1M
```

Progress and warning lines are logged to stderr before this block.

### Exit codes

| Code | Meaning |
|------|---------|
| 0 | Options were calculated and printed |
| 1 | Any failure: invalid flag or environment value, unusable `JAVA_TOOL_OPTIONS`, class counting failure, or a budget too small for the fixed regions |

Invalid input is fatal, not a warning. An unparseable `--total-memory` fails instead
of falling back to auto-detection, because sizing the JVM for the host rather than the
container reliably ends in an OOM kill.

A JVM memory option the tool recognises but cannot parse, `-Xmx1.5G` for example,
which HotSpot itself rejects, is also fatal. It previously emitted a second,
conflicting `-Xmx`.

`--head-room` accepts 0 to 99. 100 was once accepted and could never produce a valid
configuration.

## Memory detection

The budget is resolved in this order, stopping at the first that yields a limit:

1. `--total-memory` / `BPL_JVM_TOTAL_MEMORY`
2. cgroup v2: `/sys/fs/cgroup/memory.max`
3. cgroup v1: `/sys/fs/cgroup/memory/memory.limit_in_bytes`
4. Host memory: `MemAvailable` from `/proc/meminfo`, falling back to `MemTotal` on
   kernels without it. Linux only; other platforms skip this step
5. 1 GiB, with a warning on stderr

For both cgroup versions the detector starts at the process's own cgroup, resolved via
`/proc/self/cgroup`, and walks up to the mount root, using the **smallest** finite
limit it finds. An ancestor limit such as a Kubernetes pod cgroup is therefore
honoured.

Values meaning "no limit" are ignored: the literal `max`, values `<= 0`, values at or
above 4 EiB, and values too large for a signed 64-bit integer. A limit above
`MaxJVMSize` (64 TiB) is clamped to 64 TiB.

## Allocation algorithm

| Constant | Value |
|----------|-------|
| `ClassSize` (bytes per loaded class) | 5,800 |
| `ClassOverhead` (base metaspace) | 14,000,000 |
| `ClassLoadFactor` | 0.35 |
| Reserved code cache | 240 MiB |
| Direct memory | 10 MiB |
| Stack per thread | 1 MiB |
| Default thread count | 250 |
| Default head room | 0% |
| `MaxHeadRoom` | 99 |
| Default total memory | 1 GiB |
| `MaxJVMSize` | 64 TiB |

```
metaspace = ClassOverhead + LoadedClassCount * ClassSize
heap      = total - (head room + metaspace + code cache + direct memory + stack * threads)
```

Everything except the heap is sized first; the heap gets the remainder. If the fixed
regions do not fit, the calculation fails with exit code 1 rather than emitting an
unusable configuration.

Without `--loaded-class-count`, the count is derived from the classes found under
`--path` plus `BPI_JVM_CLASS_COUNT` and `BPI_CLASS_STATIC_ADJUSTMENT`, scaled by
`BPI_CLASS_ADJUSTMENT_FACTOR` percent and then by `ClassLoadFactor`.

A region already pinned in `JAVA_TOOL_OPTIONS` keeps the user's value and is left out
of the recalculated options.

## Importable package: `pkg/errors`

The only package outside `internal/`. It defines the structured error type used
throughout the tool.

```go
import "github.com/patbaumgartner/memory-calculator/pkg/errors"

type ErrorCode string

const (
    ErrInvalidMemoryFormat  ErrorCode = "INVALID_MEMORY_FORMAT"
    ErrCgroupsAccess        ErrorCode = "CGROUPS_ACCESS_ERROR"
    ErrMemoryCalculation    ErrorCode = "MEMORY_CALCULATION_ERROR"
    ErrInvalidConfiguration ErrorCode = "INVALID_CONFIGURATION"
    ErrSystemError          ErrorCode = "SYSTEM_ERROR"
)

type MemoryCalculatorError struct {
    Code    ErrorCode
    Message string
    Cause   error
    Context map[string]interface{}
}

func (e *MemoryCalculatorError) Error() string
func (e *MemoryCalculatorError) Unwrap() error

func NewMemoryFormatError(input string, cause error) *MemoryCalculatorError
func NewCgroupsError(path string, cause error) *MemoryCalculatorError
func NewCalculationError(message string, cause error) *MemoryCalculatorError
func NewConfigurationError(parameter string, value interface{}, message string) *MemoryCalculatorError
func NewSystemError(message string, cause error) *MemoryCalculatorError
```

`Error()` renders as `[CODE] message` or `[CODE] message: cause`, which is the shape
you see on stderr:

```
memory-calculator: [INVALID_CONFIGURATION] invalid configuration for total-memory: must be a memory size such as 2G, 512M, or 1073741824
```

`Unwrap()` makes the type work with `errors.Is` and `errors.As`.

---

For integration patterns and troubleshooting, see [USAGE_GUIDE.md](USAGE_GUIDE.md).
