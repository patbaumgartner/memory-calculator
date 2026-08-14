package calculator

import (
	"archive/zip"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patbaumgartner/memory-calculator/internal/calc"
)

func intPtr(value int) *int       { return &value }
func int64Ptr(value int64) *int64 { return &value }

func basicInput(path string) Input {
	return Input{
		TotalMemory:      int64Ptr(2 * calc.Gibi),
		ThreadCount:      250,
		LoadedClassCount: intPtr(5000),
		ApplicationPath:  path,
		JVMClassCount:    1000,
		AdjustmentFactor: 100,
		StaticAdjustment: 0,
	}
}

func TestExecuteWithTypedInput(t *testing.T) {
	input := basicInput(t.TempDir())
	input.HeadRoom = 10

	result, err := Create(true).Execute(input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.JavaToolOptions, "-Xmx") {
		t.Errorf("JavaToolOptions = %q, want -Xmx", result.JavaToolOptions)
	}
	if result.TotalMemory.Value != 2*calc.Gibi {
		t.Errorf("TotalMemory = %d, want %d", result.TotalMemory.Value, 2*calc.Gibi)
	}
	if result.ThreadCount != 250 || result.LoadedClassCount != 5000 || result.HeadRoom != 10 {
		t.Errorf("Result = %+v, want effective input values", result)
	}
	if result.Environment()["JAVA_TOOL_OPTIONS"] != result.JavaToolOptions {
		t.Errorf("Environment() did not render JavaToolOptions")
	}
}

func TestExecutePreservesUserOptions(t *testing.T) {
	input := basicInput(t.TempDir())
	input.JavaToolOptions = "-Xmx1G -Dapp.name=test"

	result, err := Create(true).Execute(input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got := strings.Count(result.JavaToolOptions, "-Xmx"); got != 1 {
		t.Errorf("JavaToolOptions contains %d heaps, want 1: %q", got, result.JavaToolOptions)
	}
	if !strings.HasPrefix(result.JavaToolOptions, input.JavaToolOptions) {
		t.Errorf("JavaToolOptions = %q, want prefix %q", result.JavaToolOptions, input.JavaToolOptions)
	}
}

func TestExecuteCalculatesClassCount(t *testing.T) {
	app := t.TempDir()
	if err := os.WriteFile(filepath.Join(app, "Application.class"), []byte("class"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	input := basicInput(app)
	input.LoadedClassCount = nil
	input.JVMClassCount = 1000
	input.StaticAdjustment = 100
	input.AdjustmentFactor = 200

	result, err := Create(true).Execute(input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// int((1000 JVM + 1 application + 100 static) * 200% * 35% load factor)
	if want := 770; result.LoadedClassCount != want {
		t.Errorf("LoadedClassCount = %d, want %d", result.LoadedClassCount, want)
	}
}

func TestCountAgentClasses(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "agent.jar")
	file, err := os.Create(jar) // #nosec G304 - test fixture
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	writer := zip.NewWriter(file)
	if _, err := writer.Create("agent/One.class"); err != nil {
		t.Fatalf("Create entry: %v", err)
	}
	if _, err := writer.Create("agent/Two.class"); err != nil {
		t.Fatalf("Create entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}

	tests := []struct {
		name string
		opts string
		want int
	}{
		{"no agent", "-Xmx1G", 0},
		{"agent path", "-javaagent:" + jar, 2},
		{"agent options are not part of the path", "-javaagent:" + jar + "=debug=true", 2},
		{"quoted agent path", `-javaagent:"` + jar + `"`, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Create(true).CountAgentClasses(tt.opts)
			if err != nil {
				t.Fatalf("CountAgentClasses() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("CountAgentClasses() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountAgentClassesRejectsMalformedOptions(t *testing.T) {
	if _, err := Create(true).CountAgentClasses(`-javaagent:"unterminated`); err == nil {
		t.Error("CountAgentClasses() succeeded, want an error")
	}
}

func TestExecuteRejectsInvalidDomainInput(t *testing.T) {
	input := basicInput(t.TempDir())
	input.ThreadCount = -1

	if _, err := Create(true).Execute(input); err == nil {
		t.Error("Execute() succeeded, want an error")
	}
}

func TestAdjustedClassCount(t *testing.T) {
	tests := []struct {
		name                    string
		jvm, app, agent, static int
		factor                  int
		want                    int
		wantError               bool
	}{
		{"normal", 1000, 1, 0, 100, 200, 770, false},
		{"zero factor", 1000, 1, 0, 0, 0, 0, false},
		{"negative adjusted total", 0, 0, 0, -1, 100, 0, true},
		{"sum overflow", math.MaxInt, math.MaxInt, 0, 0, 100, 0, true},
		{"scale overflow", math.MaxInt, 0, 0, 0, math.MaxInt, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adjustedClassCount(tt.jvm, tt.app, tt.agent, tt.static, tt.factor)
			if tt.wantError {
				if err == nil {
					t.Fatalf("adjustedClassCount() = %d, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("adjustedClassCount() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("adjustedClassCount() = %d, want %d", got, tt.want)
			}
		})
	}
}
