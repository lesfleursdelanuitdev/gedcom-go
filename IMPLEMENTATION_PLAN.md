# GEDCOM Go Implementation - Detailed Plan

## Overview

This document provides a comprehensive plan for implementing the GEDCOM parser in Go, including data structures, algorithms, and file organization.

## GEDCOM Format Specification

### Line Format
A GEDCOM line follows this structure:
```
LEVEL [XREF_ID] TAG [VALUE]
```

**Examples:**
- `0 HEAD` - Level 0, tag HEAD, no xref, no value
- `0 @I1@ INDI` - Level 0, xref @I1@, tag INDI, no value
- `1 NAME John /Doe/` - Level 1, tag NAME, value "John /Doe/"
- `2 DATE 1 Jan 1900` - Level 2, tag DATE, value "1 Jan 1900"

### Special Tags
- **CONC**: Continuation - concatenates value to previous line (no newline)
- **CONT**: Continuation - concatenates value to previous line (with newline)

### Record Types
- **HEAD**: Header record (single, no xref)
- **TRLR**: Trailer record (single, no xref)
- **INDI**: Individual record (has xref)
- **FAM**: Family record (has xref)
- **SOUR**: Source record (has xref)
- **REPO**: Repository record (has xref)
- **NOTE**: Note record (has xref)
- **SUBM**: Submitter record (has xref)
- **OBJE**: Multimedia record (has xref)

## Core Data Structures

### 1. GedcomLine

**Purpose**: Represents a single line in a GEDCOM file with hierarchical structure.

**Fields**:
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

**Key Methods**:
- `AddChild(child *GedcomLine)` - Add a child line
- `GetValue(selector string) string` - Get value using dot notation (e.g., "BIRT.DATE")
- `GetValues(selector string) []string` - Get all values matching selector
- `GetLines(selector string) []*GedcomLine` - Get all lines matching selector
- `ToGED() []string` - Convert to GEDCOM format
- `ToJSON() interface{}` - Convert to JSON

**Design Decisions**:
- `Children` is a map of `tag -> []*GedcomLine` to allow multiple children with same tag
- `Parent` pointer for upward traversal
- `LineNumber` for error reporting
- `XrefID` only set for level 0 records with xref

### 2. GedcomRecord

**Purpose**: Wraps a GedcomLine to represent a complete record (INDI, FAM, etc.)

**Interface**:
```go
type Record interface {
    Type() RecordType
    XrefID() string
    FirstLine() *GedcomLine
    GetValue(selector string) string
    GetValues(selector string) []string
    GetLines(selector string) []*GedcomLine
    ToGED() []string
    ToJSON() interface{}
}
```

**Base Implementation**:
```go
type BaseRecord struct {
    firstLine  *GedcomLine
    recordType RecordType
}

func (br *BaseRecord) Type() RecordType
func (br *BaseRecord) XrefID() string
func (br *BaseRecord) FirstLine() *GedcomLine
// ... other methods delegate to firstLine
```

**Specialized Records**:
- `IndividualRecord` - Extends BaseRecord with methods like `GetName()`, `GetBirthDate()`
- `FamilyRecord` - Extends BaseRecord with methods like `GetHusband()`, `GetWife()`
- `HeaderRecord` - Extends BaseRecord
- `SourceRecord`, `RepositoryRecord`, `NoteRecord`, `SubmitterRecord`, `MultimediaRecord`

**Design Decisions**:
- Interface-based design for extensibility
- BaseRecord provides default implementation
- Specialized records add domain-specific methods
- All records wrap a GedcomLine (the first line at level 0)

### 3. GedcomTree (Main Container)

**Purpose**: Represents the entire GEDCOM file structure.

**Structure**:
```go
type GedcomTree struct {
    mu sync.RWMutex
    
    // Records organized by type
    header      Record
    individuals map[string]Record  // key: xref_id
    families    map[string]Record
    notes       map[string]Record
    sources     map[string]Record
    repositories map[string]Record
    submitters  map[string]Record
    multimedia  map[string]Record
    
    // Cross-reference index (all records by xref_id)
    xrefIndex map[string]Record
    
    // Metadata
    encoding string
    version  string
    
    // Components
    errorManager *ErrorManager
    validator    Validator
    parser       Parser
    exporter     Exporter
    
    // Record counts for ID generation
    recordCounts map[RecordType]int
}
```

**Design Decisions**:
- Thread-safe with mutex (for concurrent access)
- Separate maps for each record type (type-safe access)
- Unified xrefIndex for fast cross-reference lookup
- Components (parser, validator, exporter) are interfaces

## Algorithms

### 1. Line Parsing Algorithm

**Input**: Raw line string (e.g., `"1 NAME John /Doe/"`)

**Steps**:
1. Split by whitespace (max 3 parts: level, tag/xref, value)
2. Parse level (must be integer, non-negative)
3. Check if second part is xref (starts with `@`)
4. Extract tag and value accordingly

**Pseudocode**:
```
parseLine(line string) -> (level, tag, value, xrefID, error)
    parts = split(line, whitespace, max=3)
    if len(parts) < 2:
        return error("insufficient parts")
    
    level = parseInt(parts[0])
    if level < 0:
        return error("negative level")
    
    if len(parts) == 3 && parts[1].startsWith("@"):
        // Format: level xref tag
        return (level, parts[2], "", parts[1], nil)
    else if len(parts) == 3:
        // Format: level tag value
        return (level, parts[1], parts[2], "", nil)
    else:
        // Format: level tag
        return (level, parts[1], "", "", nil)
```

**Error Handling**:
- Invalid level format → return error
- Negative level → return error
- Insufficient parts → return error
- All errors are explicit (no panics)

### 2. Tree Building Algorithm (Stack-Based)

**Input**: Stream of parsed lines

**Data Structures**:
- `parentsStack []*GedcomLine` - Stack of parent lines
- `currentValue strings.Builder` - Accumulated CONC/CONT value
- `lastTag *tagInfo` - Last processed tag info

**Steps**:
```
for each line in file:
    1. Parse line → (level, tag, value, xrefID)
    
    2. Handle CONC/CONT:
       if tag == "CONC" or tag == "CONT":
           if tag == "CONC":
               currentValue += value
           else:
               currentValue += "\n" + value
           continue
    
    3. Apply accumulated value:
       if currentValue.Len() > 0:
           parentsStack[-1].Value = currentValue.String()
           currentValue.Reset()
    
    4. Handle level 0 (top-level record):
       if level == 0:
           line = NewGedcomLine(level, tag, value, xrefID)
           record = CreateRecord(line)
           tree.AddRecord(record)
           parentsStack = [line]  // Reset stack
           continue
    
    5. Find parent level:
       while len(parentsStack) > 0 && parentsStack[-1].Level >= level:
           parentsStack = parentsStack[:len(parentsStack)-1]
       
       if len(parentsStack) == 0:
           // Orphaned line - log warning, skip
           continue
    
    6. Add as child:
       parent = parentsStack[len(parentsStack)-1]
       line = NewGedcomLine(level, tag, value, "")
       parent.AddChild(line)
       parentsStack = append(parentsStack, line)
```

**Key Points**:
- Stack maintains current parent chain
- When level decreases, pop stack until parent level < current level
- Orphaned lines (no parent) are logged but don't stop parsing
- CONC/CONT handled separately before tree building

### 3. Cross-Reference Resolution

**Algorithm**:
```
buildXrefIndex():
    for each record type:
        for each record in type:
            if record.XrefID() != "":
                xrefIndex[record.XrefID()] = record
```

**Timing**: After parsing, before validation

**Purpose**: Fast lookup of records by xref_id for validation and navigation

### 4. Selector Resolution (Dot Notation)

**Input**: Selector string (e.g., `"BIRT.DATE"`)

**Algorithm**:
```
GetValue(selector string) string:
    if selector == "":
        return this.Value
    
    parts = split(selector, ".")
    currentTag = parts[0]
    remaining = join(parts[1:], ".")
    
    if children, ok := this.Children[currentTag]; ok:
        for each child in children:
            result = child.GetValue(remaining)
            if result != "":
                return result
    
    return ""
```

**Examples**:
- `"NAME"` → returns value of first NAME child
- `"BIRT.DATE"` → returns DATE value of first BIRT child
- `"BIRT.PLAC"` → returns PLAC value of first BIRT child

## File Structure

```
gedcom-go/
├── cmd/
│   └── gedcom-cli/
│       └── main.go              # CLI entry point
├── internal/                    # Internal packages
│   ├── parser/
│   │   ├── parser.go            # Parser interface
│   │   ├── gedcom.go            # GEDCOM parser implementation
│   │   ├── json.go              # JSON parser implementation
│   │   └── line.go              # Line parsing utilities
│   ├── exporter/
│   │   ├── exporter.go          # Exporter interface
│   │   ├── gedcom.go             # GEDCOM exporter
│   │   └── json.go               # JSON exporter
│   ├── validator/
│   │   ├── validator.go         # Validator interface
│   │   ├── individual.go
│   │   ├── family.go
│   │   ├── crossref.go
│   │   └── ...
│   └── record/
│       ├── factory.go            # Record factory
│       ├── individual.go
│       ├── family.go
│       └── ...
├── pkg/                         # Public API
│   ├── line.go                  # GedcomLine
│   ├── record.go                # Record interface and BaseRecord
│   ├── tree.go                  # GedcomTree
│   ├── error.go                 # Error types and ErrorManager
│   ├── types.go                 # RecordType, constants
│   └── selector.go              # Selector resolution utilities
├── go.mod
├── go.sum
└── README.md
```

## Implementation Order

### Phase 1: Core Types (Week 1)
1. **pkg/types.go** - RecordType, constants
2. **pkg/error.go** - Error types, ErrorManager
3. **pkg/line.go** - GedcomLine struct and methods
4. **pkg/record.go** - Record interface, BaseRecord
5. **Tests** for all core types

### Phase 2: Parser (Week 2)
1. **internal/parser/line.go** - Line parsing
2. **internal/parser/gedcom.go** - GEDCOM parser
3. **pkg/tree.go** - GedcomTree with AddRecord
4. **internal/record/factory.go** - Record factory
5. **Tests** for parser

### Phase 3: Records (Week 2-3)
1. **internal/record/individual.go**
2. **internal/record/family.go**
3. **internal/record/header.go**
4. Other record types
5. **Tests** for records

### Phase 4: Validators (Week 3)
1. **internal/validator/validator.go** - Interface
2. **internal/validator/individual.go**
3. **internal/validator/family.go**
4. **internal/validator/crossref.go**
5. **Tests** for validators

### Phase 5: Exporters (Week 4)
1. **internal/exporter/exporter.go** - Interface
2. **internal/exporter/gedcom.go**
3. **internal/exporter/json.go**
4. **Tests** for exporters

### Phase 6: CLI & Integration (Week 5)
1. **cmd/gedcom-cli/main.go**
2. Integration tests
3. Documentation
4. Performance benchmarks

## Key Design Decisions

### 1. Immutability vs Mutability
- **Decision**: GedcomLine and Records are mutable after creation
- **Rationale**: Need to update values (CONC/CONT), add children during parsing
- **Alternative Considered**: Immutable with builder pattern (more complex)

### 2. Parent Pointer vs Parent ID
- **Decision**: Parent pointer in GedcomLine
- **Rationale**: Direct navigation, no lookup needed
- **Trade-off**: Circular references possible (but not problematic in tree structure)

### 3. Children as Map vs Slice
- **Decision**: `map[string][]*GedcomLine` (tag → list of children)
- **Rationale**: Fast lookup by tag, allows multiple children with same tag
- **Alternative**: Slice with linear search (slower for large trees)

### 4. Thread Safety
- **Decision**: Mutex in GedcomTree for concurrent access
- **Rationale**: May be used in web servers or concurrent processing
- **Trade-off**: Slight performance cost, but safety is worth it

### 5. Error Handling
- **Decision**: Explicit error returns, ErrorManager for collection
- **Rationale**: Go idiom, no exceptions, clear error flow
- **Benefit**: All errors are explicit and traceable

### 6. Selector Resolution
- **Decision**: Recursive dot-notation resolution
- **Rationale**: Matches Python implementation, intuitive API
- **Performance**: O(n) where n is depth of selector (acceptable)

## Performance Considerations

### Memory
- **Streaming Parser**: Parse line-by-line, don't load entire file
- **Pre-allocated Slices**: Use `make([]T, 0, capacity)` when size known
- **String Pooling**: Reuse strings where possible (Go's string interning helps)

### CPU
- **Map Lookups**: O(1) for tag-based child lookup
- **Stack Operations**: O(1) for parent stack management
- **Selector Resolution**: O(n) where n is tree depth (acceptable)

### Large Files
- **Streaming**: Process file line-by-line
- **Lazy Validation**: Validate on-demand, not all at once
- **Indexing**: Build xref index once, reuse for lookups

## Testing Strategy

### Unit Tests
- Line parsing (valid and invalid inputs)
- Tree building (various hierarchies)
- Selector resolution (simple and nested)
- Record creation and access

### Integration Tests
- Parse sample.ged → verify structure
- Parse → Export → Parse cycle
- Error handling (malformed files)
- CONC/CONT handling

### Fuzz Tests
- Random malformed lines
- Invalid levels
- Missing parts

### Benchmark Tests
- Parsing speed (lines/second)
- Memory usage
- Selector resolution speed

## Example Usage

```go
// Create tree
tree := gedcom.NewTree()

// Parse file
if err := tree.Parse("ged", "sample.ged"); err != nil {
    log.Fatal(err)
}

// Check errors
if tree.ErrorManager().HasErrors() {
    for _, err := range tree.ErrorManager().Errors() {
        fmt.Printf("Error: %s\n", err)
    }
}

// Access records
individuals := tree.Individuals()
for xrefID, indi := range individuals {
    name := indi.GetValue("NAME")
    birthDate := indi.GetValue("BIRT.DATE")
    fmt.Printf("%s: %s (born %s)\n", xrefID, name, birthDate)
}

// Export
if err := tree.Export("json", "output.json"); err != nil {
    log.Fatal(err)
}
```

## Next Steps

1. Review and approve this plan
2. Start Phase 1: Core types
3. Implement with tests from the start
4. Iterate based on findings

