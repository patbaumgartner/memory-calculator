package display

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/patbaumgartner/memory-calculator/internal/config"
)

func TestCreateFormatter(t *testing.T) {
	formatter := CreateFormatter()
	if formatter == nil {
		t.Error("CreateFormatter() returned nil")
		return
	}
	if formatter.parser == nil {
		t.Error("CreateFormatter() did not initialize parser")
	}
}

func TestDisplayVersion(t *testing.T) {
	formatter := CreateFormatter()
	cfg := &config.Config{
		BuildVersion: "1.0.0",
		BuildTime:    "2023-01-01_12:00:00",
		CommitHash:   "abc123",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.DisplayVersion(cfg)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedParts := []string{
		"JVM Memory Calculator",
		"Version: 1.0.0",
		"Build Time: 2023-01-01_12:00:00",
		"Commit: abc123",
		"Go Version: 1.24.5",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected output to contain %q, got:\n%s", part, output)
		}
	}
}

func TestDisplayHelp(t *testing.T) {
	formatter := CreateFormatter()
	cfg := &config.Config{
		BuildVersion: "1.0.0",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.DisplayHelp(cfg)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedParts := []string{
		"JVM Memory Calculator",
		"Version: 1.0.0",
		"Usage:",
		"--total-memory",
		"--thread-count",
		"--quiet",
		"Examples:",
		"memory-calculator",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected output to contain %q, got:\n%s", part, output)
		}
	}
}

func TestDisplayResults(t *testing.T) {
	formatter := CreateFormatter()
	cfg := &config.Config{
		ThreadCount:      "250",
		LoadedClassCount: "35000",
		HeadRoom:         "10",
	}

	props := map[string]string{
		"JAVA_TOOL_OPTIONS": "-Xmx1024M -Xss1M -XX:MaxMetaspaceSize=256M",
	}

	totalMemory := int64(2 * 1024 * 1024 * 1024) // 2GB

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.DisplayResults(props, totalMemory, cfg)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedParts := []string{
		"JVM Memory Configuration",
		"Total Memory:     2.00 GB",
		"Thread Count:     250",
		"Loaded Classes:   35000",
		"Head Room:        10%",
		"Calculated JVM Arguments:",
		"Complete JVM Options:",
		"JAVA_TOOL_OPTIONS=-Xmx1024M -Xss1M -XX:MaxMetaspaceSize=256M",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected output to contain %q, got:\n%s", part, output)
		}
	}
}

func TestDisplayResultsWithIndividualProps(t *testing.T) {
	formatter := CreateFormatter()
	cfg := &config.Config{
		ThreadCount:      "250",
		LoadedClassCount: "35000",
		HeadRoom:         "0",
	}

	props := map[string]string{
		"-Xmx":                      "1024M",
		"-Xss":                      "1M",
		"-XX:MaxMetaspaceSize":      "256M",
		"-XX:ReservedCodeCacheSize": "128M",
		"-XX:MaxDirectMemorySize":   "64M",
	}

	totalMemory := int64(2 * 1024 * 1024 * 1024) // 2GB

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.DisplayResults(props, totalMemory, cfg)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedParts := []string{
		"Max Heap Size:         1024M",
		"Thread Stack Size:     1M",
		"Max Metaspace Size:    256M",
		"Code Cache Size:       128M",
		"Direct Memory Size:    64M",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected output to contain %q, got:\n%s", part, output)
		}
	}
}

func TestDisplayQuietResults(t *testing.T) {
	formatter := CreateFormatter()

	tests := []struct {
		name     string
		props    map[string]string
		expected string
	}{
		{
			name: "With JAVA_TOOL_OPTIONS",
			props: map[string]string{
				"JAVA_TOOL_OPTIONS": "-Xmx1024M -Xss1M",
			},
			expected: "-Xmx1024M -Xss1M",
		},
		{
			name: "With individual flags",
			props: map[string]string{
				"-Xmx": "1024M",
				"-Xss": "1M",
			},
			expected: "-Xmx1024M -Xss1M", // Order may vary
		},
		{
			name:     "Empty props",
			props:    map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter.DisplayQuietResults(tt.props)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if tt.name == "With individual flags" {
				// For individual flags, just check that expected flags are present
				if !strings.Contains(output, "-Xmx1024M") || !strings.Contains(output, "-Xss1M") {
					t.Errorf("Expected output to contain both flags, got: %q", output)
				}
			} else {
				if output != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, output)
				}
			}
		})
	}
}

func TestExtractJVMFlag(t *testing.T) {
	formatter := CreateFormatter()

	tests := []struct {
		name            string
		javaToolOptions string
		flag            string
		expected        string
	}{
		{
			name:            "Extract Xmx flag",
			javaToolOptions: "-Xmx1024M -Xss1M",
			flag:            "-Xmx",
			expected:        "1024M",
		},
		{
			name:            "Extract XX flag with equals",
			javaToolOptions: "-XX:MaxMetaspaceSize=256M -Xmx1024M",
			flag:            "-XX:MaxMetaspaceSize",
			expected:        "256M",
		},
		{
			name:            "Flag not found",
			javaToolOptions: "-Xmx1024M -Xss1M",
			flag:            "-XX:MaxMetaspaceSize",
			expected:        "",
		},
		{
			name:            "Empty options",
			javaToolOptions: "",
			flag:            "-Xmx",
			expected:        "",
		},
		{
			name:            "Multiple similar flags",
			javaToolOptions: "-XX:MaxMetaspaceSize=256M -XX:MaxDirectMemorySize=64M",
			flag:            "-XX:MaxMetaspaceSize",
			expected:        "256M",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.extractJVMFlag(tt.javaToolOptions, tt.flag)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBuildJavaToolOptions(t *testing.T) {
	formatter := CreateFormatter()

	tests := []struct {
		name     string
		props    map[string]string
		expected string
	}{
		{
			name: "With existing JAVA_TOOL_OPTIONS",
			props: map[string]string{
				"JAVA_TOOL_OPTIONS": "-Xmx1024M -Xss1M",
			},
			expected: "-Xmx1024M -Xss1M",
		},
		{
			name: "Build from individual flags",
			props: map[string]string{
				"-Xmx": "1024M",
				"-Xss": "1M",
			},
			// Order may vary, so we'll check contents instead
		},
		{
			name:     "Empty props",
			props:    map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.buildJavaToolOptions(tt.props)

			if tt.name == "Build from individual flags" {
				// Check that result contains expected flags
				if !strings.Contains(result, "-Xmx1024M") || !strings.Contains(result, "-Xss1M") {
					t.Errorf("Expected result to contain both flags, got: %q", result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestDisplayJVMSetting(t *testing.T) {
	formatter := CreateFormatter()

	tests := []struct {
		name     string
		props    map[string]string
		flag     string
		label    string
		expected string
	}{
		{
			name: "Individual flag exists",
			props: map[string]string{
				"-Xmx": "1024M",
			},
			flag:     "-Xmx",
			label:    "Max Heap Size: ",
			expected: "Max Heap Size: 1024M",
		},
		{
			name: "Extract from JAVA_TOOL_OPTIONS",
			props: map[string]string{
				"JAVA_TOOL_OPTIONS": "-Xmx1024M -Xss1M",
			},
			flag:     "-Xmx",
			label:    "Max Heap Size: ",
			expected: "Max Heap Size: 1024M",
		},
		{
			name: "Flag not found",
			props: map[string]string{
				"JAVA_TOOL_OPTIONS": "-Xss1M",
			},
			flag:     "-Xmx",
			label:    "Max Heap Size: ",
			expected: "", // No output expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter.displayJVMSetting(tt.props, tt.flag, tt.label)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())

			if output != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}

// Benchmark tests
func BenchmarkDisplayResults(b *testing.B) {
	formatter := CreateFormatter()
	cfg := &config.Config{
		ThreadCount:      "250",
		LoadedClassCount: "35000",
		HeadRoom:         "0",
	}

	props := map[string]string{
		"JAVA_TOOL_OPTIONS": "-Xmx1024M -Xss1M -XX:MaxMetaspaceSize=256M",
	}

	totalMemory := int64(2 * 1024 * 1024 * 1024)

	// Redirect stdout to discard output during benchmark
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.DisplayResults(props, totalMemory, cfg)
	}
}

func BenchmarkExtractJVMFlag(b *testing.B) {
	formatter := CreateFormatter()
	javaToolOptions := "-Xmx1024M -Xss1M -XX:MaxMetaspaceSize=256M -XX:ReservedCodeCacheSize=128M"
	flag := "-Xmx"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatter.extractJVMFlag(javaToolOptions, flag)
	}
}

func TestEdgeCases(t *testing.T) {
	formatter := CreateFormatter()

	t.Run("Nil config", func(t *testing.T) {
		// This should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DisplayResults panicked with nil config: %v", r)
			}
		}()

		props := map[string]string{"JAVA_TOOL_OPTIONS": "-Xmx1024M"}

		// Capture and discard output
		old := os.Stdout
		os.Stdout, _ = os.Open(os.DevNull)

		// This will panic if not handled properly
		formatter.DisplayResults(props, 1024*1024*1024, &config.Config{})

		os.Stdout = old
	})

	t.Run("Very long JVM options", func(t *testing.T) {
		props := map[string]string{
			"JAVA_TOOL_OPTIONS": strings.Repeat("-Xmx1024M ", 100),
		}

		result := formatter.buildJavaToolOptions(props)
		if len(result) == 0 {
			t.Error("Expected non-empty result for long JVM options")
		}
	})
}
