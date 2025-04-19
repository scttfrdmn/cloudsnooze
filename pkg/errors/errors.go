package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType represents the type of an error
type ErrorType uint

const (
	// ErrorTypeUnknown is the default error type
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation
	// ErrorTypePermission represents permission errors
	ErrorTypePermission
	// ErrorTypeCloud represents cloud provider errors
	ErrorTypeCloud
	// ErrorTypeConfiguration represents configuration errors
	ErrorTypeConfiguration
	// ErrorTypeNetwork represents network errors
	ErrorTypeNetwork
	// ErrorTypeInternal represents internal errors
	ErrorTypeInternal
)

// CloudSnoozeError is a custom error type with context
type CloudSnoozeError struct {
	Type    ErrorType
	Message string
	Err     error
	Stack   string
}

// Error implements the error interface
func (e *CloudSnoozeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap implements the errors.Unwrap interface
func (e *CloudSnoozeError) Unwrap() error {
	return e.Err
}

// WithStack adds the stack trace to the error
func (e *CloudSnoozeError) WithStack() *CloudSnoozeError {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[0:n])

	var builder strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&builder, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	e.Stack = builder.String()
	return e
}

// New creates a new CloudSnoozeError
func New(errorType ErrorType, message string) *CloudSnoozeError {
	return &CloudSnoozeError{
		Type:    errorType,
		Message: message,
	}.WithStack()
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, message string) *CloudSnoozeError {
	return &CloudSnoozeError{
		Type:    errorType,
		Message: message,
		Err:     err,
	}.WithStack()
}

// ValidationError creates a new validation error
func ValidationError(message string) *CloudSnoozeError {
	return New(ErrorTypeValidation, message)
}

// PermissionError creates a new permission error
func PermissionError(message string) *CloudSnoozeError {
	return New(ErrorTypePermission, message)
}

// CloudError creates a new cloud provider error
func CloudError(message string) *CloudSnoozeError {
	return New(ErrorTypeCloud, message)
}

// ConfigurationError creates a new configuration error
func ConfigurationError(message string) *CloudSnoozeError {
	return New(ErrorTypeConfiguration, message)
}

// NetworkError creates a new network error
func NetworkError(message string) *CloudSnoozeError {
	return New(ErrorTypeNetwork, message)
}

// InternalError creates a new internal error
func InternalError(message string) *CloudSnoozeError {
	return New(ErrorTypeInternal, message)
}

// IsType checks if an error is of a specific type
func IsType(err error, errorType ErrorType) bool {
	if csErr, ok := err.(*CloudSnoozeError); ok {
		return csErr.Type == errorType
	}
	return false
}