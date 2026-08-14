#!/bin/bash
# Local smoke test: builds the binary and exercises the behaviour a user actually depends on.
# CI runs the full Go test suite; this is the fast pre-push check.
set -euo pipefail

BINARY=./memory-calculator
FIXTURE=$(mktemp -d)
trap 'rm -rf "$FIXTURE" "$BINARY"' EXIT

pass=0
fail=0

check() {
  local name=$1 expected=$2 actual=$3
  if [[ "$actual" == *"$expected"* ]]; then
    echo "  PASS  $name"
    pass=$((pass + 1))
  else
    echo "  FAIL  $name"
    echo "        expected to contain: $expected"
    echo "        actual:              $actual"
    fail=$((fail + 1))
  fi
}

check_exit() {
  local name=$1 expected=$2
  shift 2
  local rc=0
  "$@" >/dev/null 2>&1 || rc=$?
  if [[ "$rc" == "$expected" ]]; then
    echo "  PASS  $name"
    pass=$((pass + 1))
  else
    echo "  FAIL  $name (exit $rc, want $expected)"
    fail=$((fail + 1))
  fi
}

echo "Building..."
go build -o "$BINARY" ./cmd/memory-calculator

echo
echo "Output contract:"
check "quiet mode emits JVM flags" "-Xmx" \
  "$($BINARY --total-memory 2G --thread-count 100 --loaded-class-count 1000 --quiet --path "$FIXTURE")"
check "quiet mode emits no decoration" "0" \
  "$($BINARY --total-memory 2G --loaded-class-count 1000 --quiet --path "$FIXTURE" | grep -c 'JVM Memory Configuration' || true)"
check "verbose mode reports the memory it used" "Total Memory:     2.00 GB" \
  "$($BINARY --total-memory 2G --loaded-class-count 1000 --path "$FIXTURE")"
check "help lists the flags" "--total-memory" "$($BINARY --help)"
check "version reports build info" "Version:" "$($BINARY --version)"

echo
echo "User-supplied JVM flags are respected:"
check "explicit heap is not overridden" "-Xmx1G" \
  "$(JAVA_TOOL_OPTIONS='-Xmx1G' $BINARY --total-memory 4G --loaded-class-count 1000 --quiet --path "$FIXTURE")"
check "explicit heap is not duplicated" "1" \
  "$(JAVA_TOOL_OPTIONS='-Xmx1G' $BINARY --total-memory 4G --loaded-class-count 1000 --quiet --path "$FIXTURE" | grep -o '\-Xmx' | wc -l | tr -d ' ')"

echo
echo "Invalid input fails loudly:"
check_exit "unparseable total memory is fatal" 1 $BINARY --total-memory banana --quiet --path "$FIXTURE"
check_exit "unparseable heap flag is fatal" 1 env JAVA_TOOL_OPTIONS='-Xmxbogus' $BINARY --total-memory 2G --quiet --path "$FIXTURE"
check_exit "head room above 99 is fatal" 1 $BINARY --total-memory 2G --head-room 100 --quiet --path "$FIXTURE"
check_exit "memory too small to satisfy is fatal" 1 $BINARY --total-memory 1M --quiet --path "$FIXTURE"
check "failures explain themselves on stderr" "total-memory" \
  "$($BINARY --total-memory banana --quiet --path "$FIXTURE" 2>&1 >/dev/null || true)"

echo
echo "Passed: $pass  Failed: $fail"
[[ "$fail" -eq 0 ]]
