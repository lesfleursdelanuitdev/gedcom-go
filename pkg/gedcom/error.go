package gedcom

import (
	"fmt"
	"sync"
)

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
	SeverityHint    ErrorSeverity = "hint"    // Best practices, optimizations
	SeverityInfo    ErrorSeverity = "info"    // Data quality issues, suggestions
	SeverityWarning ErrorSeverity = "warning" // Unlikely but possible situations
	SeveritySevere  ErrorSeverity = "severe"  // Impossible situations, must fix
)

// GedcomError represents a GEDCOM parsing/validation error
type GedcomError struct {
	Severity   ErrorSeverity
	Message    string
	LineNumber int
	Context    string
}

// Error implements the error interface
func (e *GedcomError) Error() string {
	if e.LineNumber > 0 {
		return fmt.Sprintf("%s: %s (Line %d)", e.Severity, e.Message, e.LineNumber)
	}
	return fmt.Sprintf("%s: %s", e.Severity, e.Message)
}

// String returns a string representation of the error
func (e *GedcomError) String() string {
	return e.Error()
}

// ErrorManager manages collection of errors during parsing
type ErrorManager struct {
	mu     sync.RWMutex
	errors []*GedcomError
}

// NewErrorManager creates a new ErrorManager
func NewErrorManager() *ErrorManager {
	return &ErrorManager{
		errors: make([]*GedcomError, 0),
	}
}

// AddError adds an error to the collection
func (em *ErrorManager) AddError(severity ErrorSeverity, message string, lineNumber int, context string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.errors = append(em.errors, &GedcomError{
		Severity:   severity,
		Message:    message,
		LineNumber: lineNumber,
		Context:    context,
	})
}

// Errors returns a copy of all errors
func (em *ErrorManager) Errors() []*GedcomError {
	em.mu.RLock()
	defer em.mu.RUnlock()
	result := make([]*GedcomError, len(em.errors))
	copy(result, em.errors)
	return result
}

// HasErrors returns true if there are any errors
func (em *ErrorManager) HasErrors() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return len(em.errors) > 0
}

// HasSevereErrors returns true if there are any severe errors
func (em *ErrorManager) HasSevereErrors() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	for _, err := range em.errors {
		if err.Severity == SeveritySevere {
			return true
		}
	}
	return false
}

// GetErrorsBySeverity returns all errors of the specified severity
func (em *ErrorManager) GetErrorsBySeverity(severity ErrorSeverity) []*GedcomError {
	em.mu.RLock()
	defer em.mu.RUnlock()
	result := make([]*GedcomError, 0)
	for _, err := range em.errors {
		if err.Severity == severity {
			result = append(result, err)
		}
	}
	return result
}

// GetErrorSummary returns a summary of errors by severity
func (em *ErrorManager) GetErrorSummary() map[ErrorSeverity]int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	summary := make(map[ErrorSeverity]int)
	for _, err := range em.errors {
		summary[err.Severity]++
	}
	return summary
}

// Clear removes all errors
func (em *ErrorManager) Clear() {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.errors = em.errors[:0]
}

// Count returns the total number of errors
func (em *ErrorManager) Count() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return len(em.errors)
}
