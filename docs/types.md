# GEDCOM Types Documentation

Complete reference guide for core GEDCOM data types and structures.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Core Types](#core-types)
  - [GedcomTree](#gedcomtree)
  - [GedcomLine](#gedcomline)
  - [Record](#record)
  - [BaseRecord](#baserecord)
- [Record Types](#record-types)
  - [IndividualRecord](#individualrecord)
  - [FamilyRecord](#familyrecord)
  - [HeaderRecord](#headerrecord)
  - [NoteRecord](#noterecord)
  - [SourceRecord](#sourcerecord)
  - [RepositoryRecord](#repositoryrecord)
  - [SubmitterRecord](#submitterrecord)
  - [MultimediaRecord](#multimediarecord)
- [Date and Place Types](#date-and-place-types)
  - [GedcomDate](#gedcomdate)
  - [GedcomPlace](#gedcomplace)
- [Error Types](#error-types)
  - [GedcomError](#gedcomerror)
  - [ErrorManager](#errormanager)
- [Record Factory](#record-factory)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)

---

## Overview

The `gedcom` package provides core data structures and types for working with GEDCOM files. It implements the GEDCOM 5.5.1 specification with type-safe structures and thread-safe operations.

### Features

- **Type Safety**: Strong typing throughout
- **Thread Safety**: Mutex-protected concurrent access
- **Hierarchical Structure**: Complete parent-child relationships
- **Date/Place Parsing**: Structured date and place objects
- **Error Management**: Centralized error collection

---

## Installation

The gedcom package is part of the GEDCOM Go library:

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
```

---

## Core Types

### GedcomTree

The root container for all parsed GEDCOM records.

#### Features

- Thread-safe access with `sync.RWMutex`
- Separate maps for each record type
- Cross-reference index for fast lookups
- Metadata storage (encoding, version)

#### Structure

```go
type GedcomTree struct {
    mu sync.RWMutex
    
    // Records organized by type
    header       Record
    individuals  map[string]Record
    families     map[string]Record
    notes        map[string]Record
    sources      map[string]Record
    repositories map[string]Record
    submitters   map[string]Record
    multimedia   map[string]Record
    
    // Cross-reference index
    xrefIndex map[string]Record
    
    // Metadata
    encoding string
    version  string
}
```

#### Methods

```go
// Create new tree
func NewGedcomTree() *GedcomTree

// Add record
func (gt *GedcomTree) AddRecord(record Record)

// Get records
func (gt *GedcomTree) GetHeader() Record
func (gt *GedcomTree) GetIndividual(xrefID string) Record
func (gt *GedcomTree) GetFamily(xrefID string) Record
func (gt *GedcomTree) GetAllIndividuals() map[string]Record
func (gt *GedcomTree) GetAllFamilies() map[string]Record
func (gt *GedcomTree) GetAllNotes() map[string]Record
func (gt *GedcomTree) GetAllSources() map[string]Record
func (gt *GedcomTree) GetAllRepositories() map[string]Record
func (gt *GedcomTree) GetAllSubmitters() map[string]Record
func (gt *GedcomTree) GetAllMultimedia() map[string]Record

// Metadata
func (gt *GedcomTree) SetEncoding(encoding string)
func (gt *GedcomTree) GetEncoding() string
func (gt *GedcomTree) SetVersion(version string)
func (gt *GedcomTree) GetVersion() string
```

#### Example

```go
tree := gedcom.NewGedcomTree()
tree.AddRecord(individualRecord)

individuals := tree.GetAllIndividuals()
for xref, indi := range individuals {
    fmt.Printf("%s: %s\n", xref, indi.(*gedcom.IndividualRecord).GetName())
}
```

---

### GedcomLine

Represents a single line in a GEDCOM file with hierarchical structure.

#### Structure

```go
type GedcomLine struct {
    Level      int                    // 0, 1, 2, etc.
    Tag        string                 // TAG name (e.g., "NAME", "BIRT")
    Value      string                 // Value after tag
    XrefID     string                 // Cross-reference ID (e.g., "@I1@")
    LineNumber int                    // Original line number in file
    Parent     *GedcomLine            // Parent line (nil for level 0)
    Children   map[string][]*GedcomLine // Children grouped by tag
}
```

#### Methods

```go
// Create new line
func NewGedcomLine(level int, tag, value, xrefID string) *GedcomLine

// Add child
func (gl *GedcomLine) AddChild(child *GedcomLine)

// Get value using dot notation
func (gl *GedcomLine) GetValue(selector string) string

// Get lines using dot notation
func (gl *GedcomLine) GetLines(selector string) []*GedcomLine

// Convert to GEDCOM format
func (gl *GedcomLine) ToGED() string
```

#### Example

```go
line := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
child := gedcom.NewGedcomLine(2, "GIVN", "John", "")
line.AddChild(child)

// Get value using dot notation
givenName := line.GetValue("GIVN")  // "John"
```

---

### Record

Interface for all GEDCOM record types.

#### Interface

```go
type Record interface {
    Type() RecordType
    XrefID() string
    FirstLine() *GedcomLine
    GetValue(selector string) string
    GetValues(selector string) []string
    GetLines(selector string) []*GedcomLine
}
```

#### Record Types

```go
const (
    RecordTypeHEAD RecordType = "HEAD"
    RecordTypeINDI RecordType = "INDI"
    RecordTypeFAM  RecordType = "FAM"
    RecordTypeNOTE RecordType = "NOTE"
    RecordTypeSOUR RecordType = "SOUR"
    RecordTypeREPO RecordType = "REPO"
    RecordTypeSUBM RecordType = "SUBM"
    RecordTypeOBJE RecordType = "OBJE"
    RecordTypeTRLR RecordType = "TRLR"
)
```

#### Example

```go
record := tree.GetIndividual("@I1@")
if record != nil {
    fmt.Printf("Type: %s\n", record.Type())
    fmt.Printf("Xref: %s\n", record.XrefID())
    fmt.Printf("Name: %s\n", record.GetValue("NAME"))
}
```

---

### BaseRecord

Basic implementation of the Record interface.

#### Structure

```go
type BaseRecord struct {
    firstLine  *GedcomLine
    recordType RecordType
}
```

#### Methods

```go
func NewBaseRecord(line *GedcomLine) *BaseRecord
func (br *BaseRecord) Type() RecordType
func (br *BaseRecord) XrefID() string
func (br *BaseRecord) FirstLine() *GedcomLine
func (br *BaseRecord) GetValue(selector string) string
func (br *BaseRecord) GetValues(selector string) []string
func (br *BaseRecord) GetLines(selector string) []*GedcomLine
```

---

## Record Types

### IndividualRecord

Represents an Individual (INDI) record with domain-specific methods.

#### Structure

```go
type IndividualRecord struct {
    *BaseRecord
}
```

#### Methods

```go
// Name methods
func (ir *IndividualRecord) GetName() string
func (ir *IndividualRecord) GetNames() []string
func (ir *IndividualRecord) GetGivenName() string
func (ir *IndividualRecord) GetSurname() string

// Demographics
func (ir *IndividualRecord) GetSex() string

// Birth information
func (ir *IndividualRecord) GetBirthDate() string
func (ir *IndividualRecord) GetBirthPlace() string
func (ir *IndividualRecord) GetBirthData() map[string]interface{}
func (ir *IndividualRecord) GetBirthDateParsed() (*GedcomDate, error)
func (ir *IndividualRecord) GetBirthPlaceParsed() (*GedcomPlace, error)

// Death information
func (ir *IndividualRecord) GetDeathDate() string
func (ir *IndividualRecord) GetDeathPlace() string
func (ir *IndividualRecord) GetDeathData() map[string]interface{}
func (ir *IndividualRecord) GetDeathDateParsed() (*GedcomDate, error)
func (ir *IndividualRecord) GetDeathPlaceParsed() (*GedcomPlace, error)

// Relationships
func (ir *IndividualRecord) GetFamiliesAsSpouse() []string
func (ir *IndividualRecord) GetFamiliesAsChild() []string

// Events and attributes
func (ir *IndividualRecord) GetEvents() []map[string]interface{}
func (ir *IndividualRecord) GetAttributes() []map[string]interface{}

// Other
func (ir *IndividualRecord) GetOccupation() string
func (ir *IndividualRecord) GetNotes() []string
func (ir *IndividualRecord) GetSources() []string
```

#### Example

```go
indi := tree.GetIndividual("@I1@").(*gedcom.IndividualRecord)

fmt.Printf("Name: %s\n", indi.GetName())
fmt.Printf("Sex: %s\n", indi.GetSex())
fmt.Printf("Birth: %s in %s\n", indi.GetBirthDate(), indi.GetBirthPlace())
fmt.Printf("Death: %s in %s\n", indi.GetDeathDate(), indi.GetDeathPlace())

// Parsed dates
birthDate, _ := indi.GetBirthDateParsed()
if birthDate != nil && birthDate.IsValid() {
    fmt.Printf("Birth Year: %d\n", birthDate.Year)
}
```

---

### FamilyRecord

Represents a Family (FAM) record with domain-specific methods.

#### Structure

```go
type FamilyRecord struct {
    *BaseRecord
}
```

#### Methods

```go
// Relationships
func (fr *FamilyRecord) GetHusband() string
func (fr *FamilyRecord) GetWife() string
func (fr *FamilyRecord) GetChildren() []string

// Marriage information
func (fr *FamilyRecord) GetMarriageDate() string
func (fr *FamilyRecord) GetMarriagePlace() string
func (fr *FamilyRecord) GetMarriageData() map[string]interface{}
func (fr *FamilyRecord) GetMarriageDateParsed() (*GedcomDate, error)
func (fr *FamilyRecord) GetMarriagePlaceParsed() (*GedcomPlace, error)

// Divorce information
func (fr *FamilyRecord) GetDivorceDate() string
func (fr *FamilyRecord) GetDivorcePlace() string
func (fr *FamilyRecord) GetDivorceData() map[string]interface{}
func (fr *FamilyRecord) GetDivorceDateParsed() (*GedcomDate, error)
func (fr *FamilyRecord) GetDivorcePlaceParsed() (*GedcomPlace, error)

// Events
func (fr *FamilyRecord) GetEvents() []map[string]interface{}

// Other
func (fr *FamilyRecord) GetNotes() []string
func (fr *FamilyRecord) GetSources() []string
```

#### Example

```go
fam := tree.GetFamily("@F1@").(*gedcom.FamilyRecord)

fmt.Printf("Husband: %s\n", fam.GetHusband())
fmt.Printf("Wife: %s\n", fam.GetWife())
fmt.Printf("Children: %v\n", fam.GetChildren())
fmt.Printf("Marriage: %s in %s\n", fam.GetMarriageDate(), fam.GetMarriagePlace())
```

---

### HeaderRecord

Represents a Header (HEAD) record with metadata methods.

#### Structure

```go
type HeaderRecord struct {
    *BaseRecord
}
```

#### Methods

```go
func (hr *HeaderRecord) GetGedcomVersion() string
func (hr *HeaderRecord) GetGedcomForm() string
func (hr *HeaderRecord) GetCharacterEncoding() string
func (hr *HeaderRecord) GetSourceName() string
func (hr *HeaderRecord) GetSourceVersion() string
func (hr *HeaderRecord) GetSourceCorporation() string
func (hr *HeaderRecord) GetSubmissionXref() string
func (hr *HeaderRecord) GetFile() string
func (hr *HeaderRecord) GetLanguage() string
func (hr *HeaderRecord) GetDate() string
func (hr *HeaderRecord) GetTime() string
```

#### Example

```go
header := tree.GetHeader().(*gedcom.HeaderRecord)

fmt.Printf("GEDCOM Version: %s\n", header.GetGedcomVersion())
fmt.Printf("Character Encoding: %s\n", header.GetCharacterEncoding())
fmt.Printf("Source: %s %s\n", header.GetSourceName(), header.GetSourceVersion())
```

---

### NoteRecord

Represents a Note (NOTE) record.

#### Structure

```go
type NoteRecord struct {
    *BaseRecord
}
```

#### Methods

```go
func (nr *NoteRecord) GetText() string
```

#### Example

```go
note := tree.GetNote("@N1@").(*gedcom.NoteRecord)
fmt.Printf("Note: %s\n", note.GetText())
```

---

### SourceRecord

Represents a Source (SOUR) record.

#### Structure

```go
type SourceRecord struct {
    *BaseRecord
}
```

#### Methods

```go
func (sr *SourceRecord) GetTitle() string
func (sr *SourceRecord) GetAuthor() string
func (sr *SourceRecord) GetPublication() string
func (sr *SourceRecord) GetRepository() string
func (sr *SourceRecord) GetNotes() []string
```

---

### RepositoryRecord

Represents a Repository (REPO) record.

#### Structure

```go
type RepositoryRecord struct {
    *BaseRecord
}
```

#### Methods

```go
func (rr *RepositoryRecord) GetName() string
func (rr *RepositoryRecord) GetAddress() map[string]string
func (rr *RepositoryRecord) GetNotes() []string
```

---

### SubmitterRecord

Represents a Submitter (SUBM) record.

#### Structure

```go
type SubmitterRecord struct {
    *BaseRecord
}
```

#### Methods

```go
func (sr *SubmitterRecord) GetName() string
func (sr *SubmitterRecord) GetAddress() map[string]string
func (sr *SubmitterRecord) GetPhone() string
func (sr *SubmitterRecord) GetEmail() string
```

---

### MultimediaRecord

Represents a Multimedia (OBJE) record.

#### Structure

```go
type MultimediaRecord struct {
    *BaseRecord
}
```

#### Methods

```go
func (mr *MultimediaRecord) GetFile() string
func (mr *MultimediaRecord) GetFormat() string
func (mr *MultimediaRecord) GetTitle() string
func (mr *MultimediaRecord) GetNotes() []string
```

---

## Date and Place Types

### GedcomDate

Represents a parsed GEDCOM date with structured components.

#### Structure

```go
type GedcomDate struct {
    Original string   // Original GEDCOM date string
    Type     DateType // EXACT, ABOUT, BEFORE, AFTER, BETWEEN, etc.
    Calendar Calendar // GREGORIAN, JULIAN, HEBREW, etc.

    // Exact date components
    Year  int
    Month int
    Day   int

    // Range date components
    StartYear  int
    StartMonth int
    StartDay   int
    EndYear    int
    EndMonth   int
    EndDay     int

    // Parsed status
    IsParsed   bool
    ParseError error
}
```

#### Date Types

```go
const (
    DateTypeExact   DateType = "EXACT"
    DateTypeAbout   DateType = "ABOUT"
    DateTypeBefore  DateType = "BEFORE"
    DateTypeAfter   DateType = "AFTER"
    DateTypeBetween DateType = "BETWEEN"
    DateTypeFrom    DateType = "FROM"
    DateTypeTo      DateType = "TO"
    DateTypeFromTo  DateType = "FROM_TO"
    DateTypeUnknown DateType = "UNKNOWN"
)
```

#### Calendar Types

```go
const (
    CalendarGregorian Calendar = "GREGORIAN"
    CalendarJulian    Calendar = "JULIAN"
    CalendarHebrew    Calendar = "HEBREW"
    CalendarFrench    Calendar = "FRENCH"
    CalendarUnknown   Calendar = "UNKNOWN"
)
```

#### Methods

```go
// Parse date string
func ParseDate(dateStr string) (*GedcomDate, error)

// Date methods
func (gd *GedcomDate) IsValid() bool
func (gd *GedcomDate) Earliest() time.Time
func (gd *GedcomDate) Latest() time.Time
func (gd *GedcomDate) String() string
```

#### Example

```go
date, err := gedcom.ParseDate("15 JAN 1800")
if err != nil {
    panic(err)
}

fmt.Printf("Type: %s\n", date.Type)
fmt.Printf("Year: %d\n", date.Year)
fmt.Printf("Month: %d\n", date.Month)
fmt.Printf("Day: %d\n", date.Day)

// Convert to time.Time
earliest := date.Earliest()
fmt.Printf("Earliest: %s\n", earliest.Format("2006-01-02"))
```

#### Supported Formats

- `"15 JAN 1800"` - Exact date
- `"JAN 1800"` - Month-year
- `"1800"` - Year only
- `"ABT 1850"` - About
- `"BEF 1900"` - Before
- `"AFT 1900"` - After
- `"BET 1800 AND 1850"` - Between
- `"FROM 1800 TO 1850"` - Range

---

### GedcomPlace

Represents a parsed GEDCOM place with hierarchical components.

#### Structure

```go
type GedcomPlace struct {
    Original   string   // Original GEDCOM place string
    Components []string // Parsed components (from most specific to least specific)

    // Hierarchical components
    City       string
    County     string
    State      string
    Country    string
    PostalCode string

    // Geographic data (optional)
    Latitude  float64
    Longitude float64

    // Parsed status
    IsParsed   bool
    ParseError error
}
```

#### Methods

```go
// Parse place string
func ParsePlace(placeStr string) (*GedcomPlace, error)

// Place methods
func (gp *GedcomPlace) String() string
func (gp *GedcomPlace) GetComponent(level int) string
```

#### Example

```go
place, err := gedcom.ParsePlace("Rapid City, Pennington, South Dakota, USA")
if err != nil {
    panic(err)
}

fmt.Printf("City: %s\n", place.City)
fmt.Printf("County: %s\n", place.County)
fmt.Printf("State: %s\n", place.State)
fmt.Printf("Country: %s\n", place.Country)
```

#### Supported Formats

- `"Rapid City"` - Simple
- `"Rapid City, South Dakota"` - City, State
- `"Rapid City, Pennington, South Dakota, USA"` - Full hierarchy
- `"New York, NY, USA"` - With abbreviations

---

## Error Types

### GedcomError

Represents a GEDCOM parsing/validation error.

#### Structure

```go
type GedcomError struct {
    Severity   ErrorSeverity
    Message    string
    LineNumber int
    Context    string
}
```

#### Error Severity

```go
const (
    SeverityHint    ErrorSeverity = "hint"
    SeverityInfo    ErrorSeverity = "info"
    SeverityWarning ErrorSeverity = "warning"
    SeveritySevere  ErrorSeverity = "severe"
)
```

#### Methods

```go
func (e *GedcomError) Error() string
func (e *GedcomError) String() string
```

#### Example

```go
err := &gedcom.GedcomError{
    Severity:   gedcom.SeverityWarning,
    Message:    "Missing birth date",
    LineNumber: 10,
    Context:    "Individual Validation",
}
fmt.Printf("Error: %s\n", err.Error())
```

---

### ErrorManager

Manages collection of errors during parsing and validation.

#### Structure

```go
type ErrorManager struct {
    mu     sync.RWMutex
    errors []*GedcomError
}
```

#### Methods

```go
// Create new error manager
func NewErrorManager() *ErrorManager

// Add error
func (em *ErrorManager) AddError(severity ErrorSeverity, message string, lineNumber int, context string)

// Get errors
func (em *ErrorManager) Errors() []*GedcomError
func (em *ErrorManager) HasErrors() bool
func (em *ErrorManager) HasSevereErrors() bool
func (em *ErrorManager) GetErrorsBySeverity(severity ErrorSeverity) []*GedcomError
```

#### Example

```go
errorManager := gedcom.NewErrorManager()

errorManager.AddError(
    gedcom.SeverityWarning,
    "Missing birth date",
    10,
    "Individual Validation",
)

if errorManager.HasErrors() {
    errors := errorManager.Errors()
    for _, err := range errors {
        fmt.Printf("[%s] %s (Line %d)\n", 
            err.Severity, err.Message, err.LineNumber)
    }
}
```

---

## Record Factory

Creates record instances from GedcomLine.

#### Structure

```go
type RecordFactory struct{}
```

#### Methods

```go
func NewRecordFactory() *RecordFactory
func (rf *RecordFactory) CreateRecord(line *GedcomLine) Record
```

#### Example

```go
factory := gedcom.NewRecordFactory()
line := gedcom.NewGedcomLine(0, "@I1@", "INDI", "")
record := factory.CreateRecord(line)

switch record.Type() {
case gedcom.RecordTypeINDI:
    indi := record.(*gedcom.IndividualRecord)
    fmt.Printf("Individual: %s\n", indi.XrefID())
}
```

---

## API Reference

### Core Types

| Type | Description |
|------|-------------|
| `GedcomTree` | Root container for all records |
| `GedcomLine` | Single line with hierarchical structure |
| `Record` | Interface for all record types |
| `BaseRecord` | Base implementation of Record |

### Record Types

| Type | GEDCOM Tag | Description |
|------|------------|-------------|
| `IndividualRecord` | INDI | Individual person |
| `FamilyRecord` | FAM | Family unit |
| `HeaderRecord` | HEAD | File header |
| `NoteRecord` | NOTE | Note |
| `SourceRecord` | SOUR | Source citation |
| `RepositoryRecord` | REPO | Repository |
| `SubmitterRecord` | SUBM | Submitter |
| `MultimediaRecord` | OBJE | Multimedia object |

### Utility Types

| Type | Description |
|------|-------------|
| `GedcomDate` | Parsed date with components |
| `GedcomPlace` | Parsed place with hierarchy |
| `GedcomError` | Error with severity |
| `ErrorManager` | Error collection manager |
| `RecordFactory` | Record creation factory |

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
    // Parse file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Access individuals
    individuals := tree.GetAllIndividuals()
    for xref, record := range individuals {
        indi := record.(*gedcom.IndividualRecord)
        
        fmt.Printf("\n%s: %s\n", xref, indi.GetName())
        fmt.Printf("  Sex: %s\n", indi.GetSex())
        
        // Birth information
        if birthDate := indi.GetBirthDate(); birthDate != "" {
            fmt.Printf("  Birth: %s", birthDate)
            if birthPlace := indi.GetBirthPlace(); birthPlace != "" {
                fmt.Printf(" in %s", birthPlace)
            }
            fmt.Println()
            
            // Parsed date
            if parsedDate, err := indi.GetBirthDateParsed(); err == nil && parsedDate.IsValid() {
                fmt.Printf("  Birth Year: %d\n", parsedDate.Year)
            }
        }
        
        // Death information
        if deathDate := indi.GetDeathDate(); deathDate != "" {
            fmt.Printf("  Death: %s", deathDate)
            if deathPlace := indi.GetDeathPlace(); deathPlace != "" {
                fmt.Printf(" in %s", deathPlace)
            }
            fmt.Println()
        }
        
        // Relationships
        familiesAsChild := indi.GetFamiliesAsChild()
        familiesAsSpouse := indi.GetFamiliesAsSpouse()
        fmt.Printf("  Families as child: %v\n", familiesAsChild)
        fmt.Printf("  Families as spouse: %v\n", familiesAsSpouse)
    }
    
    // Access families
    families := tree.GetAllFamilies()
    for xref, record := range families {
        fam := record.(*gedcom.FamilyRecord)
        
        fmt.Printf("\nFamily %s:\n", xref)
        fmt.Printf("  Husband: %s\n", fam.GetHusband())
        fmt.Printf("  Wife: %s\n", fam.GetWife())
        fmt.Printf("  Children: %v\n", fam.GetChildren())
        
        if marriageDate := fam.GetMarriageDate(); marriageDate != "" {
            fmt.Printf("  Marriage: %s", marriageDate)
            if marriagePlace := fam.GetMarriagePlace(); marriagePlace != "" {
                fmt.Printf(" in %s", marriagePlace)
            }
            fmt.Println()
        }
    }
    
    // Access header
    if header := tree.GetHeader(); header != nil {
        hr := header.(*gedcom.HeaderRecord)
        fmt.Printf("\nGEDCOM Version: %s\n", hr.GetGedcomVersion())
        fmt.Printf("Character Encoding: %s\n", hr.GetCharacterEncoding())
    }
}
```

### Date Parsing Example

```go
// Parse various date formats
dates := []string{
    "15 JAN 1800",
    "JAN 1800",
    "1800",
    "ABT 1850",
    "BEF 1900",
    "AFT 1900",
    "BET 1800 AND 1850",
}

for _, dateStr := range dates {
    date, err := gedcom.ParseDate(dateStr)
    if err != nil {
        fmt.Printf("Error parsing %s: %v\n", dateStr, err)
        continue
    }
    
    fmt.Printf("%s -> Type: %s, Year: %d\n", dateStr, date.Type, date.Year)
    if date.IsValid() {
        fmt.Printf("  Earliest: %s\n", date.Earliest().Format("2006-01-02"))
    }
}
```

### Place Parsing Example

```go
// Parse various place formats
places := []string{
    "Rapid City",
    "Rapid City, South Dakota",
    "Rapid City, Pennington, South Dakota, USA",
    "New York, NY, USA",
}

for _, placeStr := range places {
    place, err := gedcom.ParsePlace(placeStr)
    if err != nil {
        fmt.Printf("Error parsing %s: %v\n", placeStr, err)
        continue
    }
    
    fmt.Printf("%s ->\n", placeStr)
    fmt.Printf("  City: %s\n", place.City)
    fmt.Printf("  State: %s\n", place.State)
    fmt.Printf("  Country: %s\n", place.Country)
}
```

---

## Best Practices

### Type Assertions

Always check type assertions:

```go
record := tree.GetIndividual("@I1@")
if record == nil {
    return
}

indi, ok := record.(*gedcom.IndividualRecord)
if !ok {
    return
}

// Use indi safely
fmt.Printf("Name: %s\n", indi.GetName())
```

### Error Handling

Check for errors when parsing dates/places:

```go
birthDate, err := indi.GetBirthDateParsed()
if err != nil {
    // Handle error
    return
}

if birthDate != nil && birthDate.IsValid() {
    // Use parsed date
    fmt.Printf("Year: %d\n", birthDate.Year)
}
```

### Thread Safety

GedcomTree is thread-safe, but individual records are not:

```go
// Safe: Multiple goroutines can read from tree
go func() {
    individuals := tree.GetAllIndividuals()
    // Process...
}()

// Safe: Multiple goroutines can read different records
go func() {
    indi := tree.GetIndividual("@I1@")
    // Process...
}()
```

### Dot Notation

Use dot notation for nested values:

```go
// Good: Dot notation
birthDate := indi.GetValue("BIRT.DATE")
birthPlace := indi.GetValue("BIRT.PLAC")

// Less efficient: Manual traversal
birthLines := indi.GetLines("BIRT")
for _, line := range birthLines {
    date := line.GetValue("DATE")
    // Process...
}
```

---

## See Also

- [Query API Documentation](query-api.md) - Graph-based query API
- [CLI Documentation](cli.md) - Command-line interface
- [Parser Documentation](parser.md) - Parsing GEDCOM files
- [Validator Documentation](validator.md) - Validating GEDCOM files

---

**Last Updated:** 2025-01-27
