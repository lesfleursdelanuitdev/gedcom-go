package gedcom

import (
	"fmt"
)

// ErrorType represents the type/category of an error
type ErrorType string

const (
	// ErrorTypeParse represents parsing errors
	ErrorTypeParse ErrorType = "parse"
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeQuery represents query errors
	ErrorTypeQuery ErrorType = "query"
	// ErrorTypeStorage represents storage errors
	ErrorTypeStorage ErrorType = "storage"
	// ErrorTypeIO represents I/O errors
	ErrorTypeIO ErrorType = "io"
	// ErrorTypeInternal represents internal errors
	ErrorTypeInternal ErrorType = "internal"
)

// StandardError provides a standardized error structure across packages
type StandardError struct {
	// Type categorizes the error
	Type ErrorType
	// Severity indicates error severity
	Severity ErrorSeverity
	// Message is the error message
	Message string
	// Context provides additional context (component, operation, etc.)
	Context string
	// LineNumber is the line number (if applicable)
	LineNumber int
	// Xref is the record XREF (if applicable)
	Xref string
	// Cause is the underlying error (if any)
	Cause error
}

// Error implements the error interface
func (e *StandardError) Error() string {
	msg := e.Message
	if e.Context != "" {
		msg = fmt.Sprintf("[%s] %s", e.Context, msg)
	}
	if e.LineNumber > 0 {
		msg = fmt.Sprintf("%s (line %d)", msg, e.LineNumber)
	}
	if e.Xref != "" {
		msg = fmt.Sprintf("%s (record %s)", msg, e.Xref)
	}
	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// Unwrap returns the underlying error for error unwrapping
func (e *StandardError) Unwrap() error {
	return e.Cause
}

// NewStandardError creates a new StandardError
func NewStandardError(errType ErrorType, severity ErrorSeverity, message string) *StandardError {
	return &StandardError{
		Type:     errType,
		Severity: severity,
		Message:  message,
	}
}

// NewStandardErrorWithContext creates a new StandardError with context
func NewStandardErrorWithContext(errType ErrorType, severity ErrorSeverity, message, context string) *StandardError {
	return &StandardError{
		Type:     errType,
		Severity: severity,
		Message:  message,
		Context:  context,
	}
}

// NewStandardErrorWithCause creates a new StandardError with a cause
func NewStandardErrorWithCause(errType ErrorType, severity ErrorSeverity, message string, cause error) *StandardError {
	return &StandardError{
		Type:     errType,
		Severity: severity,
		Message:  message,
		Cause:    cause,
	}
}

// WrapError wraps an existing error as a StandardError
func WrapError(errType ErrorType, severity ErrorSeverity, err error, context string) *StandardError {
	return &StandardError{
		Type:     errType,
		Severity: severity,
		Message:  err.Error(),
		Context:  context,
		Cause:    err,
	}
}

// IsParseError checks if an error is a parse error
func IsParseError(err error) bool {
	if se, ok := err.(*StandardError); ok {
		return se.Type == ErrorTypeParse
	}
	return false
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if se, ok := err.(*StandardError); ok {
		return se.Type == ErrorTypeValidation
	}
	return false
}

// IsQueryError checks if an error is a query error
func IsQueryError(err error) bool {
	if se, ok := err.(*StandardError); ok {
		return se.Type == ErrorTypeQuery
	}
	return false
}

// IsStorageError checks if an error is a storage error
func IsStorageError(err error) bool {
	if se, ok := err.(*StandardError); ok {
		return se.Type == ErrorTypeStorage
	}
	return false
}

// GetErrorSeverity extracts the severity from an error
func GetErrorSeverity(err error) ErrorSeverity {
	if se, ok := err.(*StandardError); ok {
		return se.Severity
	}
	if ge, ok := err.(*GedcomError); ok {
		return ge.Severity
	}
	return SeverityWarning // Default severity
}

// GetErrorType extracts the type from an error
func GetErrorType(err error) ErrorType {
	if se, ok := err.(*StandardError); ok {
		return se.Type
	}
	return ErrorTypeInternal // Default type
}

