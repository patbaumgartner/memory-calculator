package display

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/patbaumgartner/memory-calculator/internal/calc"
	"github.com/patbaumgartner/memory-calculator/internal/calculator"
	"github.com/patbaumgartner/memory-calculator/internal/config"
)

// capture redirects stdout for the duration of fn and returns what was written.
func capture(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	defer func() { os.Stdout = original }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy: %v", err)
	}

	return buf.String()
}

func sampleResult() calculator.Result {
	heap := calc.Heap{Value: 1678125 * calc.Kibi}
	metaspace := calc.Metaspace{Value: 15654 * calc.Kibi}

	return calculator.Result{
		JavaToolOptions: "-XX:MaxDirectMemorySize=10M -Xmx1678125K -XX:MaxMetaspaceSize=15654K " +
			"-XX:ReservedCodeCacheSize=240M -Xss1M",
		TotalMemory:      calc.Size{Value: 2 * calc.Gibi},
		ThreadCount:      250,
		LoadedClassCount: 35000,
		HeadRoom:         5,
		Regions: calc.MemoryRegions{
			Heap:              &heap,
			Metaspace:         &metaspace,
			Stack:             calc.DefaultStack,
			ReservedCodeCache: calc.DefaultReservedCodeCache,
			DirectMemory:      calc.DefaultDirectMemory,
		},
	}
}

func TestCreateFormatter(t *testing.T) {
	if CreateFormatter() == nil {
		t.Fatal("CreateFormatter() = nil")
	}
}

// TestDisplayResultsReportsTheMemoryItUsed locks the fix for a report that always read
// "Total Memory: Unknown", because the value was never passed to the formatter.
func TestDisplayResultsReportsTheMemoryItUsed(t *testing.T) {
	cfg := &config.Config{Path: "/app"}

	output := capture(t, func() { CreateFormatter().DisplayResults(sampleResult(), cfg) })

	for _, want := range []string{
		"JVM Memory Configuration",
		"Total Memory:     2.00 GB",
		"Thread Count:     250",
		"Loaded Classes:   35000",
		"Head Room:        5%",
		"Application Path: /app",
		"Max Heap Size:         1678125K",
		"Thread Stack Size:     1M",
		"Max Metaspace Size:    15654K",
		"Code Cache Size:       240M",
		"Direct Memory Size:    10M",
		"JAVA_TOOL_OPTIONS=-XX:MaxDirectMemorySize=10M",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, output)
		}
	}

	if strings.Contains(output, "Unknown") {
		t.Errorf("output reports Unknown memory:\n%s", output)
	}
}

func TestDisplayResultsToleratesMissingRegions(t *testing.T) {
	result := calculator.Result{TotalMemory: calc.Size{Value: calc.Gibi}}

	output := capture(t, func() { CreateFormatter().DisplayResults(result, &config.Config{Path: "/app"}) })

	if !strings.Contains(output, "JVM Memory Configuration") {
		t.Errorf("expected a report even with empty regions, got:\n%s", output)
	}
}

// TestDisplayQuietResultsEmitsOnlyOptions protects the documented contract that
// `export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)"` captures exactly the options.
func TestDisplayQuietResultsEmitsOnlyOptions(t *testing.T) {
	result := sampleResult()

	output := capture(t, func() { CreateFormatter().DisplayQuietResults(result) })

	if output != result.JavaToolOptions {
		t.Errorf("DisplayQuietResults() = %q, want exactly %q", output, result.JavaToolOptions)
	}
	if strings.Contains(output, "\n") {
		t.Errorf("quiet output contains a newline, which would be captured into the variable: %q", output)
	}
}

func TestDisplayVersion(t *testing.T) {
	cfg := &config.Config{BuildVersion: "1.2.3", BuildTime: "2026-01-01", CommitHash: "abc1234"}

	output := capture(t, func() { CreateFormatter().DisplayVersion(cfg) })

	for _, want := range []string{"1.2.3", "2026-01-01", "abc1234", "Go Version: go1."} {
		if !strings.Contains(output, want) {
			t.Errorf("version output missing %q\ngot:\n%s", want, output)
		}
	}
}

// TestDisplayHelpDocumentsEveryFlag guards against help text drifting from the real flags.
func TestDisplayHelpDocumentsEveryFlag(t *testing.T) {
	output := capture(t, func() { CreateFormatter().DisplayHelp(&config.Config{BuildVersion: "dev"}) })

	for _, flag := range []string{
		"--total-memory", "--thread-count", "--loaded-class-count",
		"--head-room", "--path", "--quiet", "--version", "--help",
	} {
		if !strings.Contains(output, flag) {
			t.Errorf("help output missing %q\ngot:\n%s", flag, output)
		}
	}
}
