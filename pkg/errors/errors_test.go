package errors

import (
	"errors"
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
