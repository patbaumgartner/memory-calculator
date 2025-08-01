package calc

import (
	"testing"
)

func TestDirectMemoryString(t *testing.T) {
	dm := DirectMemory{Value: 10 * Mebi}
	expected := "-XX:MaxDirectMemorySize=10M"
	result := dm.String()
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestMatchDirectMemory(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"-XX:MaxDirectMemorySize=10M", true},
		{"-XX:MaxDirectMemorySize=1K", true},
		{"-Xmx1G", false},
		{"invalid", false},
	}

	for _, test := range tests {
		result := MatchDirectMemory(test.input)
		if result != test.expected {
			t.Errorf("For input %q, expected %t, got %t", test.input, test.expected, result)
		}
	}
}

func TestParseDirectMemory(t *testing.T) {
	input := "-XX:MaxDirectMemorySize=10M"
	result, err := ParseDirectMemory(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := DirectMemory{Value: 10 * Mebi}
	if result.Value != expected.Value {
		t.Errorf("Expected value %d, got %d", expected.Value, result.Value)
	}
}

func TestHeapString(t *testing.T) {
	heap := Heap{Value: 1 * Gibi}
	expected := "-Xmx1G"
	result := heap.String()
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestMetaspaceString(t *testing.T) {
	ms := Metaspace{Value: 128 * Mebi}
	expected := "-XX:MaxMetaspaceSize=128M"
	result := ms.String()
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestReservedCodeCacheString(t *testing.T) {
	rcc := ReservedCodeCache{Value: 240 * Mebi}
	expected := "-XX:ReservedCodeCacheSize=240M"
	result := rcc.String()
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestStackString(t *testing.T) {
	stack := Stack{Value: 1 * Mebi}
	expected := "-Xss1M"
	result := stack.String()
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
