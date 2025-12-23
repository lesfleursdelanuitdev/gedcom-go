# Error Handling Guide

Complete guide to error handling in GEDCOM Go, including error types, best practices, and examples.

## Table of Contents

- [Overview](#overview)
- [Error Types](#error-types)
- [StandardError](#standarderror)
- [ErrorManager](#errormanager)
- [Error Severity Levels](#error-severity-levels)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

GEDCOM Go provides a comprehensive error handling system with:

- **Structured Errors**: `StandardError` with type, severity, context, and line numbers
- **Error Collection**: `ErrorManager` for collecting multiple errors during operations
- **Error Types**: Categorized errors (Parse, Validation, Query, Storage, IO, Internal)
- **Severity Levels**: hint, info, warning, severe
- **Error Wrapping**: Support for error chains and unwrapping

---

## Error Types

### ErrorType Constants

```go
const (
    ErrorTypeParse      ErrorType = "parse"      // Parsing errors
    ErrorTypeValidation ErrorType = "validation" // Validation errors
    ErrorTypeQuery      ErrorType = "query"       // Query errors
    ErrorTypeStorage    ErrorType = "storage"     // Storage errors
    ErrorTypeIO         ErrorType = "io"          // I/O errors
    ErrorTypeInternal   ErrorType = "internal"    // Internal errors
)
```

### Error Severity Levels

```go
const (
    SeverityHint    ErrorSeverity = "hint"    // Best practices, optimizations
    SeverityInfo    ErrorSeverity = "info"    // Data quality issues, suggestions
    SeverityWarning ErrorSeverity = "warning" // Unlikely but possible situations
    SeveritySevere  ErrorSeverity = "severe"  // Impossible situations, must fix
)
```

---

## StandardError

`StandardError` provides a standardized error structure across all packages.

### Structure

```go
type StandardError struct {
    Type       ErrorType      // Error category
    Severity   ErrorSeverity  // Error severity
    Message    string         // Error message
    Context    string         // Additional context (component, operation)
    LineNumber int            // Line number (if applicable)
    Xref       string         // Record XREF (if applicable)
    Cause      error          // Underlying error (if any)
}
```

### Creating StandardErrors

#### Basic Error

```go
err := gedcom.NewStandardError(
    gedcom.ErrorTypeQuery,
    gedcom.SeverityWarning,
    "Individual not found",
)
```

#### Error with Context

```go
err := gedcom.NewStandardErrorWithContext(
    gedcom.ErrorTypeQuery,
    gedcom.SeverityWarning,
    "Individual not found",
    "Query.Individual",
)
```

#### Error with Cause

```go
err := gedcom.NewStandardErrorWithCause(
    gedcom.ErrorTypeStorage,
    gedcom.SeveritySevere,
    "Failed to load node from database",
    originalErr,
)
```

#### Wrapping Existing Errors

```go
err := gedcom.WrapError(
    gedcom.ErrorTypeStorage,
    gedcom.SeveritySevere,
    originalErr,
    "HybridStorage.LoadNode",
)
```

### Checking Error Types

```go
// Check if error is a specific type
if gedcom.IsParseError(err) {
    // Handle parse error
}

if gedcom.IsValidationError(err) {
    // Handle validation error
}

if gedcom.IsQueryError(err) {
    // Handle query error
}

if gedcom.IsStorageError(err) {
    // Handle storage error
}

// Get error type
errorType := gedcom.GetErrorType(err)

// Get error severity
severity := gedcom.GetErrorSeverity(err)
```

### Error Unwrapping

```go
// Unwrap to get underlying error
if se, ok := err.(*gedcom.StandardError); ok {
    underlyingErr := se.Unwrap()
    // Handle underlying error
}
```

---

## ErrorManager

`ErrorManager` collects multiple errors during operations (parsing, validation, etc.).

### Basic Usage

```go
errorManager := gedcom.NewErrorManager()

// Add errors
errorManager.AddError(
    gedcom.SeverityWarning,
    "Missing birth date",
    123,
    "Individual Validation",
)

errorManager.AddError(
    gedcom.SeveritySevere,
    "Invalid XREF format",
    456,
    "Cross-Reference Validation",
)

// Check for errors
if errorManager.HasErrors() {
    errors := errorManager.Errors()
    for _, err := range errors {
        fmt.Printf("[%s] %s (line %d)\n", 
            err.Severity, err.Message, err.LineNumber)
    }
}

// Check for severe errors
if errorManager.HasSevereErrors() {
    fmt.Println("Severe errors found!")
}

// Get error summary
summary := errorManager.GetErrorSummary()
fmt.Printf("Severe: %d, Warning: %d, Info: %d, Hint: %d\n",
    summary[gedcom.SeveritySevere],
    summary[gedcom.SeverityWarning],
    summary[gedcom.SeverityInfo],
    summary[gedcom.SeverityHint])

// Get errors by severity
warnings := errorManager.GetErrorsBySeverity(gedcom.SeverityWarning)

// Clear errors
errorManager.Clear()

// Get error count
count := errorManager.Count()
```

---

## Examples

### Example 1: Parsing with Error Collection

```go
p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")
if err != nil {
    log.Fatal(err)
}

// Get parsing errors
errors := p.GetErrors()
if len(errors) > 0 {
    fmt.Printf("Found %d parsing issues:\n", len(errors))
    for _, err := range errors {
        switch err.Severity {
        case gedcom.SeveritySevere:
            fmt.Printf("  ✗ [SEVERE] %s (line %d)\n", 
                err.Message, err.LineNumber)
        case gedcom.SeverityWarning:
            fmt.Printf("  ⚠ [WARNING] %s (line %d)\n", 
                err.Message, err.LineNumber)
        case gedcom.SeverityInfo:
            fmt.Printf("  ℹ [INFO] %s (line %d)\n", 
                err.Message, err.LineNumber)
        case gedcom.SeverityHint:
            fmt.Printf("  💡 [HINT] %s (line %d)\n", 
                err.Message, err.LineNumber)
        }
    }
}
```

### Example 2: Validation with Error Filtering

```go
errorManager := gedcom.NewErrorManager()
validator := validator.NewGedcomValidator(errorManager)
validator.EnableAdvancedValidation()

err := validator.Validate(tree)
if err != nil {
    log.Fatal(err)
}

// Filter errors by severity
allErrors := errorManager.Errors()
severeErrors := errorManager.GetErrorsBySeverity(gedcom.SeveritySevere)

if len(severeErrors) > 0 {
    fmt.Printf("Found %d severe errors:\n", len(severeErrors))
    for _, err := range severeErrors {
        fmt.Printf("  ✗ %s (line %d) [%s]\n", 
            err.Message, err.LineNumber, err.Context)
    }
}
```

### Example 3: Query Error Handling

```go
q, err := query.NewQuery(tree)
if err != nil {
    // Check error type
    if gedcom.IsQueryError(err) {
        fmt.Println("Query initialization error")
    }
    log.Fatal(err)
}

// Execute query with error handling
results, err := q.Filter().ByName("John").Execute()
if err != nil {
    // Check error type
    if gedcom.IsQueryError(err) {
        if se, ok := err.(*gedcom.StandardError); ok {
            fmt.Printf("Query error: %s (context: %s)\n", 
                se.Message, se.Context)
        }
    }
    log.Fatal(err)
}
```

### Example 4: Storage Error Handling

```go
graph, err := query.BuildGraphHybrid(tree, "indexes.db", "graph_data", config)
if err != nil {
    // Check if it's a storage error
    if gedcom.IsStorageError(err) {
        if se, ok := err.(*gedcom.StandardError); ok {
            fmt.Printf("Storage error: %s\n", se.Message)
            if se.Cause != nil {
                fmt.Printf("  Cause: %v\n", se.Cause)
            }
        }
    }
    log.Fatal(err)
}
```

### Example 5: Custom Error Creation

```go
// Create custom error for missing individual
func findIndividual(graph *query.Graph, xref string) (*gedcom.IndividualRecord, error) {
    node := graph.GetIndividual(xref)
    if node == nil {
        return nil, gedcom.NewStandardErrorWithContext(
            gedcom.ErrorTypeQuery,
            gedcom.SeverityWarning,
            fmt.Sprintf("Individual %s not found", xref),
            "Graph.GetIndividual",
        )
    }
    return node.Individual, nil
}

// Usage
indi, err := findIndividual(graph, "@I123@")
if err != nil {
    if gedcom.IsQueryError(err) {
        // Handle query error
        fmt.Printf("Error: %s\n", err.Error())
    }
}
```

### Example 6: Error Wrapping

```go
func loadNodeFromDatabase(db *sql.DB, nodeID uint32) (*query.IndividualNode, error) {
    // Database operation that might fail
    row := db.QueryRow("SELECT data FROM nodes WHERE id = ?", nodeID)
    var data []byte
    if err := row.Scan(&data); err != nil {
        // Wrap the database error
        return nil, gedcom.WrapError(
            gedcom.ErrorTypeStorage,
            gedcom.SeveritySevere,
            err,
            "Database.LoadNode",
        )
    }
    // ... deserialize node
    return node, nil
}

// Usage
node, err := loadNodeFromDatabase(db, 123)
if err != nil {
    // Check error type
    if gedcom.IsStorageError(err) {
        if se, ok := err.(*gedcom.StandardError); ok {
            fmt.Printf("Storage error: %s\n", se.Message)
            fmt.Printf("Context: %s\n", se.Context)
            // Unwrap to get original database error
            if underlyingErr := se.Unwrap(); underlyingErr != nil {
                fmt.Printf("Original error: %v\n", underlyingErr)
            }
        }
    }
}
```

### Example 7: Error Aggregation

```go
func processMultipleFiles(files []string) error {
    var allErrors []error
    
    for _, file := range files {
        p := parser.NewHierarchicalParser()
        tree, err := p.Parse(file)
        if err != nil {
            allErrors = append(allErrors, gedcom.WrapError(
                gedcom.ErrorTypeParse,
                gedcom.SeveritySevere,
                err,
                fmt.Sprintf("Parser.Parse(%s)", file),
            ))
            continue
        }
        
        // Process tree...
    }
    
    // Aggregate errors
    if len(allErrors) > 0 {
        return gedcom.NewStandardErrorWithCause(
            gedcom.ErrorTypeParse,
            gedcom.SeveritySevere,
            fmt.Sprintf("Failed to process %d files", len(allErrors)),
            fmt.Errorf("%d errors occurred", len(allErrors)),
        )
    }
    
    return nil
}
```

---

## Best Practices

### 1. Always Check Errors

```go
// ❌ Bad
tree, _ := p.Parse("family.ged")

// ✅ Good
tree, err := p.Parse("family.ged")
if err != nil {
    return fmt.Errorf("failed to parse file: %w", err)
}
```

### 2. Provide Context

```go
// ❌ Bad
return fmt.Errorf("not found")

// ✅ Good
return gedcom.NewStandardErrorWithContext(
    gedcom.ErrorTypeQuery,
    gedcom.SeverityWarning,
    "Individual not found",
    "Query.Individual",
)
```

### 3. Use Appropriate Severity

```go
// Severe: Must fix, operation cannot continue
errorManager.AddError(
    gedcom.SeveritySevere,
    "Invalid XREF format",
    lineNumber,
    "Parser",
)

// Warning: Unlikely but possible, should review
errorManager.AddError(
    gedcom.SeverityWarning,
    "Missing birth date",
    lineNumber,
    "Validator",
)

// Info: Data quality issue, suggestion
errorManager.AddError(
    gedcom.SeverityInfo,
    "Consider adding birth place",
    lineNumber,
    "Validator",
)

// Hint: Best practice, optimization
errorManager.AddError(
    gedcom.SeverityHint,
    "Consider using lazy loading for large datasets",
    0,
    "Performance",
)
```

### 4. Wrap Errors with Context

```go
// ❌ Bad
if err != nil {
    return err
}

// ✅ Good
if err != nil {
    return gedcom.WrapError(
        gedcom.ErrorTypeStorage,
        gedcom.SeveritySevere,
        err,
        "HybridStorage.LoadNode",
    )
}
```

### 5. Check Error Types

```go
// Use type checking for different error handling
if gedcom.IsQueryError(err) {
    // Handle query errors
} else if gedcom.IsStorageError(err) {
    // Handle storage errors
} else if gedcom.IsValidationError(err) {
    // Handle validation errors
}
```

### 6. Log Errors Appropriately

```go
if err != nil {
    severity := gedcom.GetErrorSeverity(err)
    switch severity {
    case gedcom.SeveritySevere:
        log.Error(err)
    case gedcom.SeverityWarning:
        log.Warn(err)
    case gedcom.SeverityInfo:
        log.Info(err)
    case gedcom.SeverityHint:
        log.Debug(err)
    }
}
```

---

## Troubleshooting

### Common Error Patterns

#### Pattern 1: Silent Failures

**Problem**: Errors are ignored

```go
// ❌ Bad
tree, _ := p.Parse("family.ged")
```

**Solution**: Always check errors

```go
// ✅ Good
tree, err := p.Parse("family.ged")
if err != nil {
    return fmt.Errorf("parse failed: %w", err)
}
```

#### Pattern 2: Lost Context

**Problem**: Error doesn't indicate where it occurred

```go
// ❌ Bad
return fmt.Errorf("not found")
```

**Solution**: Add context

```go
// ✅ Good
return gedcom.NewStandardErrorWithContext(
    gedcom.ErrorTypeQuery,
    gedcom.SeverityWarning,
    "Individual not found",
    "Query.Individual",
)
```

#### Pattern 3: Error Swallowing

**Problem**: Original error is lost

```go
// ❌ Bad
if err != nil {
    return fmt.Errorf("operation failed")
}
```

**Solution**: Wrap original error

```go
// ✅ Good
if err != nil {
    return gedcom.WrapError(
        gedcom.ErrorTypeStorage,
        gedcom.SeveritySevere,
        err,
        "Storage.Operation",
    )
}
```

---

## Summary

- Use `StandardError` for structured errors with type, severity, and context
- Use `ErrorManager` to collect multiple errors during operations
- Always check errors and provide appropriate context
- Use error type checking for different handling strategies
- Wrap errors to preserve error chains
- Use appropriate severity levels (severe, warning, info, hint)

For more examples, see [API_EXAMPLES.md](API_EXAMPLES.md).

