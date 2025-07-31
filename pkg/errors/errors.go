// Package errors provides structured error types for the memory calculator.
package errors

import (
	"fmt"
)

// ErrorCode represents different types of errors that can occur.
type ErrorCode string

const (
	// ErrInvalidMemoryFormat indicates an invalid memory format string.
	ErrInvalidMemoryFormat ErrorCode = "INVALID_MEMORY_FORMAT"
	// ErrCgroupsAccess indicates problems accessing cgroups filesystem.
	ErrCgroupsAccess ErrorCode = "CGROUPS_ACCESS_ERROR"
	// ErrMemoryCalculation indicates errors in JVM memory calculation.
	ErrMemoryCalculation ErrorCode = "MEMORY_CALCULATION_ERROR"
	// ErrInvalidConfiguration indicates invalid configuration parameters.
	ErrInvalidConfiguration ErrorCode = "INVALID_CONFIGURATION"
	// ErrSystemError indicates system-level errors.
	ErrSystemError ErrorCode = "SYSTEM_ERROR"
)

// MemoryCalculatorError represents a structured error with context.
type MemoryCalculatorError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface.
func (e *MemoryCalculatorError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error.
func (e *MemoryCalculatorError) Unwrap() error {
	return e.Cause
}

// NewMemoryFormatError creates a new memory format error.
func NewMemoryFormatError(input string, cause error) *MemoryCalculatorError {
	return &MemoryCalculatorError{
		Code:    ErrInvalidMemoryFormat,
		Message: fmt.Sprintf("invalid memory format: %s", input),
		Cause:   cause,
		Context: map[string]interface{}{
			"input": input,
		},
	}
}

// NewCgroupsError creates a new cgroups access error.
func NewCgroupsError(path string, cause error) *MemoryCalculatorError {
	return &MemoryCalculatorError{
		Code:    ErrCgroupsAccess,
		Message: fmt.Sprintf("failed to read cgroups at %s", path),
		Cause:   cause,
		Context: map[string]interface{}{
			"path": path,
		},
	}
}

// NewCalculationError creates a new memory calculation error.
func NewCalculationError(message string, cause error) *MemoryCalculatorError {
	return &MemoryCalculatorError{
		Code:    ErrMemoryCalculation,
		Message: message,
		Cause:   cause,
	}
}

// NewConfigurationError creates a new configuration error.
func NewConfigurationError(parameter string, value interface{}, message string) *MemoryCalculatorError {
	return &MemoryCalculatorError{
		Code:    ErrInvalidConfiguration,
		Message: fmt.Sprintf("invalid configuration for %s: %s", parameter, message),
		Context: map[string]interface{}{
			"parameter": parameter,
			"value":     value,
		},
	}
}

// NewSystemError creates a new system error.
func NewSystemError(message string, cause error) *MemoryCalculatorError {
	return &MemoryCalculatorError{
		Code:    ErrSystemError,
		Message: message,
		Cause:   cause,
	}
}
