package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestMemoryCalculatorError(t *testing.T) {
	tests := []struct {
		name     string
		error    *MemoryCalculatorError
		expected string
	}{
		{
			name: "Error with cause",
			error: &MemoryCalculatorError{
				Code:    ErrInvalidMemoryFormat,
				Message: "invalid memory format",
				Cause:   errors.New("parse error"),
			},
			expected: "[INVALID_MEMORY_FORMAT] invalid memory format: parse error",
		},
		{
			name: "Error without cause",
			error: &MemoryCalculatorError{
				Code:    ErrSystemError,
				Message: "system error occurred",
			},
			expected: "[SYSTEM_ERROR] system error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.error.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewMemoryFormatError(t *testing.T) {
	input := "invalid_input"
	cause := errors.New("parse error")
	err := NewMemoryFormatError(input, cause)

	if err.Code != ErrInvalidMemoryFormat {
		t.Errorf("Expected code %v, got %v", ErrInvalidMemoryFormat, err.Code)
	}

	if err.Context["input"] != input {
		t.Errorf("Expected input %v in context, got %v", input, err.Context["input"])
	}

	if err.Unwrap() != cause {
		t.Errorf("Expected cause %v, got %v", cause, err.Unwrap())
	}
}

func TestNewCgroupsError(t *testing.T) {
	path := "/sys/fs/cgroup/memory.max"
	cause := errors.New("file not found")
	err := NewCgroupsError(path, cause)

	if err.Code != ErrCgroupsAccess {
		t.Errorf("Expected code %v, got %v", ErrCgroupsAccess, err.Code)
	}

	if err.Context["path"] != path {
		t.Errorf("Expected path %v in context, got %v", path, err.Context["path"])
	}
}

func TestNewCalculationError(t *testing.T) {
	message := "calculation failed"
	cause := errors.New("helper error")
	err := NewCalculationError(message, cause)

	if err.Code != ErrMemoryCalculation {
		t.Errorf("Expected code %v, got %v", ErrMemoryCalculation, err.Code)
	}

	if err.Message != message {
		t.Errorf("Expected message %v, got %v", message, err.Message)
	}
}

func TestNewConfigurationError(t *testing.T) {
	parameter := "thread-count"
	value := "-1"
	message := "must be positive"
	err := NewConfigurationError(parameter, value, message)

	if err.Code != ErrInvalidConfiguration {
		t.Errorf("Expected code %v, got %v", ErrInvalidConfiguration, err.Code)
	}

	if err.Context["parameter"] != parameter {
		t.Errorf("Expected parameter %v in context, got %v", parameter, err.Context["parameter"])
	}

	if err.Context["value"] != value {
		t.Errorf("Expected value %v in context, got %v", value, err.Context["value"])
	}
}

func TestNewSystemError(t *testing.T) {
	message := "system failed"
	cause := errors.New("underlying error")
	err := NewSystemError(message, cause)

	if err.Code != ErrSystemError {
		t.Errorf("Expected code %v, got %v", ErrSystemError, err.Code)
	}

	if err.Message != message {
		t.Errorf("Expected message %v, got %v", message, err.Message)
	}
}

func TestErrorUnwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &MemoryCalculatorError{
		Code:    ErrSystemError,
		Message: "system error",
		Cause:   cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("Expected unwrapped error %v, got %v", cause, unwrapped)
	}

	// Test error without cause
	errNoCause := &MemoryCalculatorError{
		Code:    ErrSystemError,
		Message: "system error",
	}

	if unwrapped := errNoCause.Unwrap(); unwrapped != nil {
		t.Errorf("Expected nil unwrapped error, got %v", unwrapped)
	}
}

func TestErrorChaining(t *testing.T) {
	rootCause := errors.New("root cause")
	intermediateErr := NewSystemError("intermediate error", rootCause)
	finalErr := NewCalculationError("final error", intermediateErr)

	// Test error unwrapping chain
	unwrapped := finalErr.Unwrap()
	if unwrapped != intermediateErr {
		t.Errorf("Expected intermediate error, got %v", unwrapped)
	}

	// Test using errors.Unwrap from standard library
	deepUnwrapped := errors.Unwrap(unwrapped)
	if deepUnwrapped != rootCause {
		t.Errorf("Expected root cause, got %v", deepUnwrapped)
	}

	// Test errors.Is functionality
	if !errors.Is(finalErr, intermediateErr) {
		t.Error("errors.Is should find intermediate error in chain")
	}

	if !errors.Is(finalErr, rootCause) {
		t.Error("errors.Is should find root cause in chain")
	}
}

func TestErrorCodeConstants(t *testing.T) {
	expectedCodes := map[ErrorCode]string{
		ErrInvalidMemoryFormat:  "INVALID_MEMORY_FORMAT",
		ErrCgroupsAccess:        "CGROUPS_ACCESS_ERROR",
		ErrMemoryCalculation:    "MEMORY_CALCULATION_ERROR",
		ErrInvalidConfiguration: "INVALID_CONFIGURATION",
		ErrSystemError:          "SYSTEM_ERROR",
	}

	for code, expectedString := range expectedCodes {
		if string(code) != expectedString {
			t.Errorf("ErrorCode %v should equal %s, got %s", code, expectedString, string(code))
		}
	}
}

func TestContextPreservation(t *testing.T) {
	tests := []struct {
		name        string
		createError func() *MemoryCalculatorError
		checkFunc   func(*testing.T, *MemoryCalculatorError)
	}{
		{
			name: "Memory format error preserves input",
			createError: func() *MemoryCalculatorError {
				return NewMemoryFormatError("2XGB", errors.New("invalid unit"))
			},
			checkFunc: func(t *testing.T, err *MemoryCalculatorError) {
				if err.Context["input"] != "2XGB" {
					t.Errorf("Expected input context '2XGB', got %v", err.Context["input"])
				}
			},
		},
		{
			name: "Cgroups error preserves path",
			createError: func() *MemoryCalculatorError {
				return NewCgroupsError("/invalid/path", errors.New("not found"))
			},
			checkFunc: func(t *testing.T, err *MemoryCalculatorError) {
				if err.Context["path"] != "/invalid/path" {
					t.Errorf("Expected path context '/invalid/path', got %v", err.Context["path"])
				}
			},
		},
		{
			name: "Configuration error preserves parameter and value",
			createError: func() *MemoryCalculatorError {
				return NewConfigurationError("heap-size", "invalid", "not a number")
			},
			checkFunc: func(t *testing.T, err *MemoryCalculatorError) {
				if err.Context["parameter"] != "heap-size" {
					t.Errorf("Expected parameter context 'heap-size', got %v", err.Context["parameter"])
				}
				if err.Context["value"] != "invalid" {
					t.Errorf("Expected value context 'invalid', got %v", err.Context["value"])
				}
			},
		},
		{
			name: "Calculation error has no context",
			createError: func() *MemoryCalculatorError {
				return NewCalculationError("failed to calculate", errors.New("math error"))
			},
			checkFunc: func(t *testing.T, err *MemoryCalculatorError) {
				if err.Context != nil {
					t.Errorf("Expected no context for calculation error, got %v", err.Context)
				}
			},
		},
		{
			name: "System error has no context",
			createError: func() *MemoryCalculatorError {
				return NewSystemError("system failure", errors.New("os error"))
			},
			checkFunc: func(t *testing.T, err *MemoryCalculatorError) {
				if err.Context != nil {
					t.Errorf("Expected no context for system error, got %v", err.Context)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()
			tt.checkFunc(t, err)
		})
	}
}

func TestErrorFormatting(t *testing.T) {
	tests := []struct {
		name          string
		error         *MemoryCalculatorError
		expectedParts []string
	}{
		{
			name:          "Simple error",
			error:         NewSystemError("disk full", nil),
			expectedParts: []string{"[SYSTEM_ERROR]", "disk full"},
		},
		{
			name:          "Error with cause",
			error:         NewMemoryFormatError("bad", errors.New("parse failed")),
			expectedParts: []string{"[INVALID_MEMORY_FORMAT]", "invalid memory format: bad", "parse failed"},
		},
		{
			name:          "Complex nested error",
			error:         NewCalculationError("outer", NewSystemError("inner", errors.New("root"))),
			expectedParts: []string{"[MEMORY_CALCULATION_ERROR]", "outer", "[SYSTEM_ERROR]", "inner", "root"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorStr := tt.error.Error()
			for _, part := range tt.expectedParts {
				if !strings.Contains(errorStr, part) {
					t.Errorf("Error string '%s' should contain '%s'", errorStr, part)
				}
			}
		})
	}
}

func TestErrorEquality(t *testing.T) {
	// Test that same error types with same content are logically equivalent
	err1 := NewMemoryFormatError("1GB", errors.New("test"))
	err2 := NewMemoryFormatError("1GB", errors.New("test"))

	if err1 == err2 {
		t.Error("Two different error instances should not be pointer-equal")
	}

	if err1.Code != err2.Code {
		t.Error("Same error type should have same code")
	}

	if err1.Error() != err2.Error() {
		t.Error("Same error content should produce same error string")
	}
}

func TestNilContextHandling(t *testing.T) {
	// Test that errors handle nil context gracefully
	err := &MemoryCalculatorError{
		Code:    ErrSystemError,
		Message: "test error",
		Context: nil, // Explicitly nil
	}

	// Should not panic when accessing or printing
	errorStr := err.Error()
	if !strings.Contains(errorStr, "test error") {
		t.Errorf("Error with nil context should still format correctly, got: %s", errorStr)
	}
}

func TestErrorWithComplexContext(t *testing.T) {
	// Test errors with complex context data
	complexValue := map[string]interface{}{
		"nested": map[string]int{"count": 42},
		"array":  []string{"a", "b", "c"},
	}

	err := &MemoryCalculatorError{
		Code:    ErrInvalidConfiguration,
		Message: "complex config error",
		Context: map[string]interface{}{
			"simple":  "value",
			"complex": complexValue,
			"number":  123,
		},
	}

	// Should handle complex context without panicking
	errorStr := err.Error()
	if !strings.Contains(errorStr, "complex config error") {
		t.Errorf("Error with complex context should format message correctly, got: %s", errorStr)
	}

	// Context should be preserved
	if err.Context["simple"] != "value" {
		t.Error("Simple context value should be preserved")
	}
	if err.Context["number"] != 123 {
		t.Error("Numeric context value should be preserved")
	}
}

func TestErrorImplementsErrorInterface(t *testing.T) {
	var err error = &MemoryCalculatorError{
		Code:    ErrSystemError,
		Message: "test",
	}

	// Should be usable as standard error
	if err.Error() == "" {
		t.Error("Error should implement error interface properly")
	}

	// Should work with fmt verbs
	formatted := fmt.Sprintf("Got error: %v", err)
	if !strings.Contains(formatted, "test") {
		t.Errorf("Error should format with %%v, got: %s", formatted)
	}
}
