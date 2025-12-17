# GEDCOM Parser Documentation

Complete reference guide for parsing GEDCOM files.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Parser Types](#parser-types)
  - [HierarchicalParser](#hierarchicalparser)
  - [BasicParser](#basicparser)
  - [ParallelHierarchicalParser](#parallelhierarchicalparser)
  - [TwoPhaseParser](#twophaseparser)
  - [StreamingHierarchicalParser](#streaminghierarchicalparser)
- [Basic Usage](#basic-usage)
- [API Reference](#api-reference)
- [Parser Components](#parser-components)
  - [Line Parsing](#line-parsing)
  - [Stack-Based Hierarchy](#stack-based-hierarchy)
  - [Continuation Handling](#continuation-handling)
  - [Encoding Detection](#encoding-detection)
  - [File Validation](#file-validation)
- [Error Handling](#error-handling)
- [Examples](#examples)
- [Performance Considerations](#performance-considerations)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The parser package provides functionality for parsing GEDCOM (Genealogical Data Communication) files. It implements a hierarchical parsing algorithm that builds a complete tree structure from GEDCOM files, handling nested levels, continuation lines (CONC/CONT), and various character encodings.

### Features

- **Hierarchical Parsing**: Builds complete parent-child relationships
- **Multiple Parser Types**: Sequential, parallel, streaming, and two-phase parsers
- **Encoding Detection**: Supports UTF-8, UTF-16 (with BOM), and ANSEL
- **Continuation Handling**: Processes CONC (concatenate) and CONT (continue) lines
- **Error Recovery**: Continues parsing after non-fatal errors
- **Line Validation**: Validates line format and structure
- **Memory Efficient**: Streaming parser for very large files

---

## Installation

The parser package is part of the GEDCOM Go library:

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
```

---

## Parser Types

### HierarchicalParser

The **HierarchicalParser** is the recommended parser for most use cases. It builds a complete hierarchical tree structure from GEDCOM files.

#### Features

- Complete tree structure with all relationships
- Error collection without stopping parsing
- Handles all GEDCOM record types
- Validates file and encoding
- Processes continuation lines (CONC/CONT)

#### Usage

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
)

func main() {
    // Create parser
    p := parser.NewHierarchicalParser()

    // Parse file
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Check for errors
    if p.HasErrors() {
        errors := p.GetErrors()
        for _, err := range errors {
            fmt.Printf("Error: %s (Line %d)\n", err.Message, err.LineNumber)
        }
    }

    // Use tree
    individuals := tree.GetAllIndividuals()
    fmt.Printf("Found %d individuals\n", len(individuals))
}
```

---

### BasicParser

The **BasicParser** is a backward-compatible wrapper around HierarchicalParser. It maintains the API from earlier versions.

#### Usage

```go
p := parser.NewBasicParser()
tree, err := p.Parse("family.ged")
```

**Note:** This parser internally uses `HierarchicalParser`, so functionality is identical.

---

### ParallelHierarchicalParser

The **ParallelHierarchicalParser** is an experimental parser that attempts to parallelize record creation while maintaining sequential parsing for hierarchy.

#### Features

- Parallel record creation in separate goroutine
- Sequential parsing (required for hierarchy)
- Experimental - use with caution

#### Usage

```go
p := parser.NewParallelHierarchicalParser()
tree, err := p.Parse("family.ged")
```

**Note:** This is experimental. The sequential parser is recommended for most use cases.

---

### TwoPhaseParser

The **TwoPhaseParser** implements a two-phase parsing approach:
1. **Phase 1**: Collect records sequentially (identify boundaries)
2. **Phase 2**: Parse record children in parallel

#### Features

- Sequential record collection
- Parallel child parsing
- Useful for very large files with many records

#### Usage

```go
p := parser.NewTwoPhaseParser()
tree, err := p.Parse("large_family.ged")
```

---

### StreamingHierarchicalParser

The **StreamingHierarchicalParser** processes records incrementally without loading the entire tree into memory. Ideal for very large files (>100MB).

#### Features

- Memory efficient (processes records one at a time)
- Callback-based processing
- No full tree in memory
- Suitable for files >100MB

#### Usage with Handler

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func main() {
    // Create streaming parser
    sp := parser.NewStreamingHierarchicalParser()

    // Parse with callback
    err := sp.ParseWithHandler("large.ged", func(record gedcom.Record) error {
        // Process record immediately without storing entire tree
        fmt.Printf("Found record: %s (%s)\n", record.Type(), record.XrefID())
        
        // Example: Only process individuals
        if record.Type() == "INDI" {
            indi := record.(*gedcom.IndividualRecord)
            fmt.Printf("  Name: %s\n", indi.GetName())
        }
        
        return nil // Continue parsing
    })
    
    if err != nil {
        panic(err)
    }
}
```

#### When to Use Streaming Parser

- Files larger than 100MB
- Limited memory available
- Processing records one at a time is sufficient
- Don't need full tree structure in memory

---

## Basic Usage

### Simple Parse

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
)

func main() {
    // Create parser
    p := parser.NewHierarchicalParser()

    // Parse file
    tree, err := p.Parse("family.ged")
    if err != nil {
        fmt.Printf("Parse failed: %v\n", err)
        return
    }

    // Access parsed data
    individuals := tree.GetAllIndividuals()
    families := tree.GetAllFamilies()

    fmt.Printf("Parsed %d individuals and %d families\n", 
        len(individuals), len(families))
}
```

### Parse with Error Checking

```go
p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")
if err != nil {
    // Fatal error (file not found, encoding error, etc.)
    panic(err)
}

// Check for non-fatal errors
if p.HasErrors() {
    errors := p.GetErrors()
    for _, err := range errors {
        switch err.Severity {
        case gedcom.SeveritySevere:
            fmt.Printf("SEVERE: %s (Line %d)\n", err.Message, err.LineNumber)
        case gedcom.SeverityWarning:
            fmt.Printf("WARNING: %s (Line %d)\n", err.Message, err.LineNumber)
        case gedcom.SeverityInfo:
            fmt.Printf("INFO: %s (Line %d)\n", err.Message, err.LineNumber)
        }
    }
}
```

### Access Parsed Records

```go
tree, err := p.Parse("family.ged")
if err != nil {
    panic(err)
}

// Get all individuals
individuals := tree.GetAllIndividuals()
for xref, indi := range individuals {
    fmt.Printf("%s: %s\n", xref, indi.GetName())
}

// Get all families
families := tree.GetAllFamilies()
for xref, fam := range families {
    fmt.Printf("Family %s: Husband=%s, Wife=%s\n", 
        xref, fam.GetHusband(), fam.GetWife())
}

// Get header
header := tree.GetHeader()
if header != nil {
    fmt.Printf("Source: %s\n", header.GetValue("SOUR"))
}
```

---

## API Reference

### HierarchicalParser

#### Constructor

```go
func NewHierarchicalParser() *HierarchicalParser
```

Creates a new hierarchical parser instance.

#### Methods

##### Parse

```go
func (hp *HierarchicalParser) Parse(filePath string) (*gedcom.GedcomTree, error)
```

Parses a GEDCOM file and returns the tree structure. Returns an error only for fatal issues (file not found, encoding errors). Non-fatal errors are collected and available via `GetErrors()`.

**Parameters:**
- `filePath`: Path to the GEDCOM file

**Returns:**
- `*gedcom.GedcomTree`: The parsed tree structure
- `error`: Fatal error if parsing cannot continue

##### GetErrors

```go
func (hp *HierarchicalParser) GetErrors() []*gedcom.GedcomError
```

Returns all errors collected during parsing, including warnings and info messages.

##### HasErrors

```go
func (hp *HierarchicalParser) HasErrors() bool
```

Returns `true` if any errors were encountered during parsing.

##### HasSevereErrors

```go
func (hp *HierarchicalParser) HasSevereErrors() bool
```

Returns `true` if any severe errors were encountered.

##### GetErrorManager

```go
func (hp *HierarchicalParser) GetErrorManager() *gedcom.ErrorManager
```

Returns the error manager for advanced error handling.

##### GetTree

```go
func (hp *HierarchicalParser) GetTree() *gedcom.GedcomTree
```

Returns the parsed tree (same as return value from `Parse`).

---

## Parser Components

### Line Parsing

The parser uses `ParseLine` to extract components from each GEDCOM line.

#### GEDCOM Line Format

```
LEVEL [XREF_ID] TAG [VALUE]
```

**Examples:**
- `0 HEAD` → Level 0, Tag "HEAD", No value, No xref
- `0 @I1@ INDI` → Level 0, Tag "INDI", No value, Xref "@I1@"
- `1 NAME John /Doe/` → Level 1, Tag "NAME", Value "John /Doe/", No xref
- `2 DATE 1 Jan 1900` → Level 2, Tag "DATE", Value "1 Jan 1900", No xref

#### ParseLine Function

```go
func ParseLine(line string) (level int, tag string, value string, xrefID string, err error)
```

Parses a single GEDCOM line into its components.

**Returns:**
- `level`: The level number (0, 1, 2, etc.)
- `tag`: The tag name (HEAD, INDI, NAME, etc.)
- `value`: The value after the tag (empty if no value)
- `xrefID`: The cross-reference ID if present (empty if not)
- `err`: Error if line format is invalid

---

### Stack-Based Hierarchy

The parser uses a stack-based algorithm to build hierarchical relationships.

#### Algorithm

1. **Level 0 lines**: Create record and add to tree, reset stack, push to stack
2. **Level > 0 lines**: Find parent using stack, add as child, push to stack
3. **Stack management**: Pop lines when level decreases

#### LineStack

The `LineStack` maintains the current parent chain during parsing.

```go
type LineStack struct {
    lines []*gedcom.GedcomLine
}
```

**Methods:**
- `Push(line)`: Add line to top of stack
- `Pop()`: Remove and return top line
- `Peek()`: Return top line without removing
- `FindParent(level)`: Find appropriate parent for given level
- `IsEmpty()`: Check if stack is empty
- `Clear()`: Remove all lines

#### Example Stack Behavior

```
Line: 0 @I1@ INDI
  → Stack: [INDI@0]

Line: 1 NAME
  → Stack: [INDI@0, NAME@1]

Line: 2 GIVN John
  → Stack: [INDI@0, NAME@1, GIVN@2]

Line: 2 SURN Doe
  → Stack: [INDI@0, NAME@1, SURN@2]  (GIVN popped, SURN pushed)

Line: 1 SEX M
  → Stack: [INDI@0, SEX@1]  (NAME and SURN popped, SEX pushed)
```

---

### Continuation Handling

GEDCOM allows long lines to be split using continuation tags:
- **CONC** (Concatenate): Appends value directly (no space, no newline)
- **CONT** (Continue): Appends value with newline

#### ContinuationHandler

The `ContinuationHandler` manages CONC/CONT lines.

```go
type ContinuationHandler struct {
    currentValue strings.Builder
    lastTag      *TagInfo
}
```

**Methods:**
- `HandleContinuation(tag, level, value)`: Process CONC/CONT line
- `HasAccumulatedValue()`: Check if value is accumulated
- `GetAccumulatedValue()`: Get and reset accumulated value
- `SetLastTag(tag, level)`: Set last processed tag

#### Example

```
1 NOTE This is a long note that continues
2 CONC on the same line without space
2 CONT and continues on a new line
```

Results in:
```
NOTE This is a long note that continues on the same line without space
and continues on a new line
```

---

### Encoding Detection

The parser automatically detects file encoding by reading the BOM (Byte Order Mark).

#### Supported Encodings

- **UTF-8**: Most common (default if no BOM)
- **UTF-16**: Big-endian (FE FF) or Little-endian (FF FE)
- **ANSEL**: Detected from CHAR tag in header
- **ASCII/ANSI**: Fallback options

#### DetectEncoding Function

```go
func DetectEncoding(filePath string) (Encoding, error)
```

**Detection Order:**
1. Check for UTF-8 BOM (EF BB BF)
2. Check for UTF-16 BE BOM (FE FF)
3. Check for UTF-16 LE BOM (FF FE)
4. Default to UTF-8 if no BOM found

#### GetReader Function

```go
func GetReader(file *os.File, encoding Encoding) (io.Reader, error)
```

Returns an appropriate reader for the given encoding, handling BOM skipping.

---

### File Validation

The parser validates files before parsing.

#### ValidateFile Function

```go
func ValidateFile(filePath string) error
```

**Checks Performed:**
1. File exists
2. Path is a file (not a directory)
3. File is readable
4. File is not empty

**Returns:** Error if validation fails, nil otherwise.

#### Additional File Functions

```go
// Get file information
func GetFileInfo(filePath string) (*FileInfo, error)

// Check if file is readable
func IsReadable(filePath string) bool

// Check if file exists
func FileExists(filePath string) bool

// Check if path is a file
func IsFile(filePath string) bool
```

---

## Error Handling

### Error Severity Levels

The parser uses severity levels for errors:

- **Severe**: Critical errors that may affect parsing
- **Warning**: Issues that don't stop parsing but may indicate problems
- **Info**: Informational messages
- **Hint**: Suggestions for improvement

### Error Collection

Errors are collected during parsing but don't stop the process (except for fatal errors).

```go
p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")

// Fatal error (file not found, encoding error, etc.)
if err != nil {
    panic(err)
}

// Non-fatal errors
if p.HasErrors() {
    errors := p.GetErrors()
    for _, err := range errors {
        fmt.Printf("[%s] %s (Line %d)\n", 
            err.Severity, err.Message, err.LineNumber)
    }
}
```

### Common Errors

#### File Errors

- **File not found**: File path doesn't exist
- **Permission denied**: File is not readable
- **Is directory**: Path points to a directory
- **File is empty**: File has no content

#### Parsing Errors

- **Malformed line**: Line doesn't match GEDCOM format
- **Invalid level**: Level is negative or invalid
- **Orphaned line**: Line has no parent (hierarchy issue)
- **Invalid continuation**: CONC/CONT used incorrectly

#### Encoding Errors

- **Encoding detection failed**: Cannot determine encoding
- **Reader creation failed**: Cannot create reader for encoding

---

## Examples

### Complete Example

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func main() {
    // Create parser
    p := parser.NewHierarchicalParser()

    // Parse file
    tree, err := p.Parse("family.ged")
    if err != nil {
        fmt.Printf("Fatal error: %v\n", err)
        return
    }

    // Check for errors
    if p.HasErrors() {
        fmt.Printf("\nParsing completed with %d issues:\n", len(p.GetErrors()))
        for _, err := range p.GetErrors() {
            fmt.Printf("  [%s] %s (Line %d)\n", 
                err.Severity, err.Message, err.LineNumber)
        }
    } else {
        fmt.Println("✓ Parsing completed without errors")
    }

    // Access data
    individuals := tree.GetAllIndividuals()
    families := tree.GetAllFamilies()

    fmt.Printf("\nStatistics:\n")
    fmt.Printf("  Individuals: %d\n", len(individuals))
    fmt.Printf("  Families: %d\n", len(families))

    // Process individuals
    fmt.Printf("\nIndividuals:\n")
    for xref, indi := range individuals {
        fmt.Printf("  %s: %s\n", xref, indi.GetName())
        if sex := indi.GetValue("SEX"); sex != "" {
            fmt.Printf("    Sex: %s\n", sex)
        }
    }
}
```

### Streaming Example

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func main() {
    sp := parser.NewStreamingHierarchicalParser()

    individualCount := 0
    familyCount := 0

    err := sp.ParseWithHandler("large.ged", func(record gedcom.Record) error {
        switch record.Type() {
        case "INDI":
            individualCount++
            if individualCount%1000 == 0 {
                fmt.Printf("Processed %d individuals...\n", individualCount)
            }
        case "FAM":
            familyCount++
        }
        return nil // Continue parsing
    })

    if err != nil {
        panic(err)
    }

    fmt.Printf("Processed %d individuals and %d families\n", 
        individualCount, familyCount)
}
```

### Error Filtering Example

```go
p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")
if err != nil {
    panic(err)
}

// Filter errors by severity
allErrors := p.GetErrors()
severeErrors := []*gedcom.GedcomError{}
warnings := []*gedcom.GedcomError{}

for _, err := range allErrors {
    switch err.Severity {
    case gedcom.SeveritySevere:
        severeErrors = append(severeErrors, err)
    case gedcom.SeverityWarning:
        warnings = append(warnings, err)
    }
}

fmt.Printf("Severe errors: %d\n", len(severeErrors))
fmt.Printf("Warnings: %d\n", len(warnings))
```

---

## Performance Considerations

### Parser Selection

| Parser Type | Best For | Memory Usage | Speed |
|------------|----------|--------------|-------|
| **HierarchicalParser** | Most files (<100MB) | Medium | Fast |
| **StreamingHierarchicalParser** | Large files (>100MB) | Low | Fast |
| **TwoPhaseParser** | Very large files with many records | Medium | Very Fast |
| **ParallelHierarchicalParser** | Experimental use | Medium | Variable |

### Memory Usage

- **HierarchicalParser**: Loads entire tree into memory
- **StreamingHierarchicalParser**: Processes records one at a time
- **TwoPhaseParser**: Loads records but parses in parallel

### Performance Tips

1. **Use streaming parser for large files** (>100MB)
2. **Reuse parser instances** when parsing multiple files
3. **Check errors after parsing** (don't check during parsing)
4. **Use appropriate parser** for your file size

---

## Best Practices

### Error Handling

Always check for errors after parsing:

```go
p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")
if err != nil {
    // Handle fatal error
    log.Fatalf("Parse failed: %v", err)
}

// Check for non-fatal errors
if p.HasErrors() {
    // Log or handle errors
    for _, err := range p.GetErrors() {
        log.Printf("Parse issue: %s", err.Message)
    }
}
```

### File Validation

Validate files before parsing when possible:

```go
if err := parser.ValidateFile("family.ged"); err != nil {
    log.Fatalf("File validation failed: %v", err)
}

p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")
```

### Memory Management

For large files, use streaming parser:

```go
// Good for large files
sp := parser.NewStreamingHierarchicalParser()
err := sp.ParseWithHandler("large.ged", func(record gedcom.Record) error {
    // Process record
    return nil
})
```

### Parser Reuse

Reuse parser instances when parsing multiple files:

```go
p := parser.NewHierarchicalParser()

for _, file := range files {
    tree, err := p.Parse(file)
    // Process tree
}
```

---

## Troubleshooting

### Common Issues

#### 1. "File not found" Error

**Problem:** File path is incorrect or file doesn't exist.

**Solutions:**
- Check file path is correct
- Use absolute path if needed
- Verify file exists: `parser.FileExists(filePath)`

#### 2. "Encoding detection failed" Error

**Problem:** Cannot determine file encoding.

**Solutions:**
- Check file has valid BOM
- Try specifying encoding manually (if supported)
- Verify file is valid GEDCOM format

#### 3. "Malformed line" Warnings

**Problem:** Lines don't match GEDCOM format.

**Solutions:**
- Check line format: `LEVEL [XREF] TAG [VALUE]`
- Verify no extra spaces or characters
- Check for encoding issues

#### 4. "Orphaned line" Warnings

**Problem:** Line has no parent in hierarchy.

**Solutions:**
- Check GEDCOM file structure
- Verify level numbers are correct
- May indicate corrupted file

#### 5. Memory Issues with Large Files

**Problem:** Parser runs out of memory.

**Solutions:**
- Use `StreamingHierarchicalParser` for large files
- Process records incrementally
- Increase available memory

### Debugging

Enable verbose error reporting:

```go
p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")
if err != nil {
    fmt.Printf("Fatal: %v\n", err)
    return
}

// Print all errors
for _, err := range p.GetErrors() {
    fmt.Printf("[%s] Line %d: %s\n", 
        err.Severity, err.LineNumber, err.Message)
}
```

### Performance Debugging

Measure parsing time:

```go
import "time"

start := time.Now()
p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")
duration := time.Since(start)

fmt.Printf("Parsed in %v\n", duration)
fmt.Printf("Records: %d\n", len(tree.GetAllIndividuals()))
```

---

## See Also

- [CLI Documentation](cli.md) - Command-line interface for parsing
- [Exporter Documentation](exporter.md) - Exporting parsed data
- [GEDCOM Specification](https://www.gedcom.org/) - Official GEDCOM 5.5.1 specification

---

**Last Updated:** 2025-01-27
