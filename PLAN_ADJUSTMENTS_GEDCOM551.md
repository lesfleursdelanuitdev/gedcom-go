# Plan Adjustments Based on GEDCOM 5.5.1 Specification

## Overview

This document outlines adjustments needed to our implementation plan based on the GEDCOM 5.5.1 specification. Note: GEDCOM 7.0 exists but we're targeting 5.5.1 for compatibility with existing files.

## Key Findings from GEDCOM 5.5.1 Specification

### 1. Line Format - CRITICAL ADJUSTMENT

**Current Plan**: Assumes format `LEVEL [XREF_ID] TAG [VALUE]`

**GEDCOM 5.5.1 Specification**:
- Format: `LEVEL [XREF_ID] TAG [VALUE]`
- **XREF_ID position**: XREF_ID comes BEFORE the tag, not after
- **Order matters**: `LEVEL XREF_ID TAG` or `LEVEL TAG VALUE`

**Examples from spec**:
```
0 HEAD                    # Level 0, tag HEAD
0 @I1@ INDI              # Level 0, xref @I1@, tag INDI
1 NAME John /Doe/         # Level 1, tag NAME, value "John /Doe/"
2 DATE 1 Jan 1900         # Level 2, tag DATE, value "1 Jan 1900"
```

**Adjustment Needed**:
- ✅ Our parsing algorithm is correct (we already handle this)
- Verify: XREF_ID always comes immediately after level number
- Verify: XREF_ID format is `@[A-Z0-9_]+@` (alphanumeric + underscore)

### 2. Character Encoding - IMPORTANT ADJUSTMENT

**Current Plan**: Assumes UTF-8 with BOM detection

**GEDCOM 5.5.1 Specification**:
- **Primary encoding**: ANSEL (American National Standard for Extended Latin)
- **Also supports**: ASCII, UTF-8, UNICODE (UTF-16)
- **CHAR tag**: Specifies encoding in header
  - `ANSI` = Windows ANSI (code page 1252)
  - `ANSEL` = American National Standard Extended Latin
  - `ASCII` = 7-bit ASCII
  - `UTF-8` = UTF-8
  - `UNICODE` = UTF-16

**Adjustment Needed**:
```go
// Add ANSEL support (complex encoding)
type Encoding string

const (
    EncodingANSEL Encoding = "ANSEL"  // Primary for 5.5.1
    EncodingASCII Encoding = "ASCII"
    EncodingUTF8  Encoding = "UTF-8"
    EncodingUTF16 Encoding = "UNICODE"
    EncodingANSI  Encoding = "ANSI"
)

// ANSEL requires special handling - may need library
// For now, support UTF-8 and UTF-16, warn on ANSEL
```

**Implementation Strategy**:
1. **Phase 1**: Support UTF-8 and UTF-16 (most common)
2. **Phase 2**: Add ANSEL support (requires encoding library)
3. **Phase 3**: Add ASCII and ANSI support

### 3. Level Constraints - VERIFICATION NEEDED

**Current Plan**: Levels can be 0, 1, 2, 3, etc.

**GEDCOM 5.5.1 Specification**:
- **Maximum level**: Typically 0-99, but specification may limit
- **Level 0**: Only for top-level records (HEAD, TRLR, INDI, FAM, etc.)
- **Level validation**: Each tag has valid level ranges

**Adjustment Needed**:
```go
// Add level validation
const (
    MaxLevel = 99  // Verify from spec
    MinLevel = 0
)

func validateLevel(level int, tag string) error {
    if level < MinLevel || level > MaxLevel {
        return fmt.Errorf("level %d out of range [%d-%d]", level, MinLevel, MaxLevel)
    }
    // Tag-specific level validation (e.g., HEAD must be level 0)
    return nil
}
```

### 4. Tag Format - VERIFICATION NEEDED

**Current Plan**: Tags are strings

**GEDCOM 5.5.1 Specification**:
- **Tag format**: 1-31 characters, uppercase letters, numbers, underscore
- **User-defined tags**: Must start with underscore `_`
- **Tag validation**: Each record type has valid child tags

**Adjustment Needed**:
```go
// Add tag validation
func isValidTag(tag string) bool {
    if len(tag) == 0 || len(tag) > 31 {
        return false
    }
    // Check format: uppercase letters, numbers, underscore
    for _, r := range tag {
        if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
            return false
        }
    }
    return true
}

func isUserDefinedTag(tag string) bool {
    return strings.HasPrefix(tag, "_")
}
```

### 5. XREF_ID Format - VERIFICATION NEEDED

**Current Plan**: XREF_ID is string starting with `@`

**GEDCOM 5.5.1 Specification**:
- **Format**: `@[A-Z0-9_]+@`
- **Length**: 1-22 characters between @ symbols
- **Case**: Typically uppercase
- **Uniqueness**: Must be unique within file

**Adjustment Needed**:
```go
// Add XREF validation
func isValidXrefID(xref string) bool {
    if !strings.HasPrefix(xref, "@") || !strings.HasSuffix(xref, "@") {
        return false
    }
    inner := xref[1 : len(xref)-1]
    if len(inner) == 0 || len(inner) > 22 {
        return false
    }
    // Check alphanumeric + underscore
    for _, r := range inner {
        if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
            return false
        }
    }
    return true
}
```

### 6. CONC/CONT Handling - VERIFICATION NEEDED

**Current Plan**: CONC concatenates, CONT adds newline

**GEDCOM 5.5.1 Specification**:
- **CONC**: Concatenates value to previous line's value (no space, no newline)
- **CONT**: Concatenates value to previous line's value (with newline)
- **Level**: CONC/CONT must be at same or higher level than previous line
- **Restriction**: CONC/CONT cannot be subordinate to another CONC/CONT

**Adjustment Needed**:
```go
// Verify CONC/CONT algorithm
func handleContinuation(tag string, level int, value string, 
    lastTag *tagInfo, currentValue *strings.Builder) error {
    
    if tag == "CONC" {
        // No space, no newline - direct concatenation
        currentValue.WriteString(value)
    } else if tag == "CONT" {
        // Add newline before value
        currentValue.WriteString("\n")
        currentValue.WriteString(value)
    }
    
    // Verify level constraint
    if lastTag != nil && lastTag.tag == "CONC" || lastTag.tag == "CONT" {
        if level < lastTag.level {
            return fmt.Errorf("CONC/CONT cannot be subordinate to CONC/CONT")
        }
    }
    
    return nil
}
```

### 7. Value Format - VERIFICATION NEEDED

**Current Plan**: Value is free-form string

**GEDCOM 5.5.1 Specification**:
- **Length**: Typically up to 255 characters per line (before CONC/CONT)
- **Special characters**: 
  - `/` around surnames: `John /Doe/`
  - `@` for xref references: `@I1@`
- **Whitespace**: Preserved (except leading/trailing on line)
- **Line breaks**: Only via CONT tag

**Adjustment Needed**:
- ✅ Our plan already handles this correctly
- Add: Maximum line length validation (255 chars before continuation)
- Add: Surname extraction utility (text between `/`)

### 8. Record Structure - VERIFICATION NEEDED

**Current Plan**: Records are hierarchical trees

**GEDCOM 5.5.1 Specification**:
- **Required records**: HEAD (required), TRLR (required)
- **Optional records**: All others
- **Record order**: HEAD first, TRLR last, others in between
- **XREF requirements**: 
  - HEAD: No xref
  - TRLR: No xref
  - INDI, FAM, SOUR, REPO, NOTE, SUBM, OBJE: Must have xref

**Adjustment Needed**:
```go
// Add record validation
func validateRecordStructure(record Record) error {
    switch record.Type() {
    case RecordTypeHEAD:
        if record.XrefID() != "" {
            return fmt.Errorf("HEAD record must not have xref")
        }
    case RecordTypeTRLR:
        if record.XrefID() != "" {
            return fmt.Errorf("TRLR record must not have xref")
        }
    case RecordTypeINDI, RecordTypeFAM, RecordTypeSOUR, 
         RecordTypeREPO, RecordTypeNOTE, RecordTypeSUBM, RecordTypeOBJE:
        if record.XrefID() == "" {
            return fmt.Errorf("%s record must have xref", record.Type())
        }
    }
    return nil
}
```

### 9. Date Format - NEW REQUIREMENT

**Current Plan**: Dates are strings

**GEDCOM 5.5.1 Specification**:
- **Formats**: 
  - Exact: `1 JAN 1900`
  - Approximate: `ABT 1900`, `CAL 1900`, `EST 1900`
  - Before/After: `BEF 1900`, `AFT 1900`
  - Between: `BET 1900 AND 1901`
  - Period: `FROM 1900 TO 1905`
- **Month abbreviations**: JAN, FEB, MAR, APR, MAY, JUN, JUL, AUG, SEP, OCT, NOV, DEC

**Adjustment Needed**:
```go
// Add date parsing (separate package)
package date

type Date struct {
    Type    DateType  // Exact, Approximate, Before, After, Between, Period
    Year    int
    Month   int       // 1-12, 0 if not specified
    Day     int       // 1-31, 0 if not specified
    Year2   int       // For ranges
    Month2  int       // For ranges
    Day2    int       // For ranges
}

type DateType string

const (
    DateTypeExact      DateType = "EXACT"
    DateTypeApproximate DateType = "ABT"
    DateTypeCalculated DateType = "CAL"
    DateTypeEstimated DateType = "EST"
    DateTypeBefore     DateType = "BEF"
    DateTypeAfter      DateType = "AFT"
    DateTypeBetween    DateType = "BET"
    DateTypePeriod     DateType = "FROM"
)

func ParseDate(dateStr string) (*Date, error) {
    // Complex parsing logic
}
```

### 10. Place Format - NEW REQUIREMENT

**Current Plan**: Places are strings

**GEDCOM 5.5.1 Specification**:
- **Format**: Hierarchical (most specific to least specific)
- **Example**: `Weston, Madison, Connecticut, United States of America`
- **Components**: Can be separated by commas
- **FORM tag**: Can specify place format

**Adjustment Needed**:
```go
// Add place parsing
type Place struct {
    Value     string   // Full place string
    Components []string // Parsed components
    Form      string   // Place format (if specified)
}

func ParsePlace(placeStr string) *Place {
    components := strings.Split(placeStr, ",")
    for i := range components {
        components[i] = strings.TrimSpace(components[i])
    }
    return &Place{
        Value:      placeStr,
        Components: components,
    }
}
```

## Updated File Structure

```
gedcom-go/
├── pkg/
│   ├── line.go              # GedcomLine
│   ├── record.go            # Record interface
│   ├── tree.go              # GedcomTree
│   ├── error.go             # Error types
│   ├── types.go             # RecordType, constants
│   ├── selector.go          # Selector resolution
│   ├── date/                # NEW: Date parsing
│   │   ├── date.go
│   │   └── parser.go
│   └── place/               # NEW: Place parsing
│       └── place.go
├── internal/
│   ├── parser/
│   │   ├── line.go          # Line parsing (with validation)
│   │   ├── gedcom.go        # GEDCOM parser
│   │   └── encoding.go      # NEW: Encoding detection & conversion
│   ├── validator/
│   │   ├── tag.go           # NEW: Tag validation
│   │   ├── xref.go          # NEW: XREF validation
│   │   └── level.go         # NEW: Level validation
│   └── ...
```

## Implementation Priority Adjustments

### Phase 1 (Core Types) - ADD:
1. ✅ GedcomLine (existing)
2. ✅ Record interface (existing)
3. ✅ Error types (existing)
4. **NEW**: Tag validation utilities
5. **NEW**: XREF validation utilities
6. **NEW**: Level validation utilities

### Phase 2 (Parser) - ADD:
1. ✅ Line parsing (existing)
2. **ENHANCE**: Encoding detection (add ANSEL support plan)
3. **ENHANCE**: CONC/CONT handling (verify spec compliance)
4. **NEW**: Line length validation (255 char limit)
5. **NEW**: Record structure validation

### Phase 3 (Records) - ADD:
1. ✅ Base records (existing)
2. **NEW**: Date parsing (separate package)
3. **NEW**: Place parsing (separate package)
4. **ENHANCE**: Specialized records with date/place helpers

### Phase 4 (Validators) - ADD:
1. ✅ Basic validators (existing)
2. **NEW**: Tag format validator
3. **NEW**: XREF uniqueness validator
4. **NEW**: Level range validator
5. **NEW**: Record structure validator

## Critical Verification Points

From the GEDCOM 5.5.1 PDF, verify:

1. **Line Format**:
   - [ ] Exact XREF_ID position (before or after tag?)
   - [ ] Maximum line length (255 chars?)
   - [ ] Whitespace handling rules

2. **Encoding**:
   - [ ] ANSEL character mapping table
   - [ ] Encoding detection priority
   - [ ] Fallback behavior

3. **Levels**:
   - [ ] Maximum level value
   - [ ] Tag-specific level constraints
   - [ ] Level 0 restrictions

4. **Tags**:
   - [ ] Maximum tag length (31 chars?)
   - [ ] Valid characters
   - [ ] User-defined tag rules

5. **XREF**:
   - [ ] Maximum length (22 chars?)
   - [ ] Valid characters
   - [ ] Uniqueness scope (file-wide?)

6. **CONC/CONT**:
   - [ ] Exact concatenation rules
   - [ ] Level constraints
   - [ ] Nesting restrictions

7. **Dates**:
   - [ ] All date format variations
   - [ ] Month abbreviations
   - [ ] Year-only dates

8. **Places**:
   - [ ] Hierarchical structure rules
   - [ ] Component separation
   - [ ] FORM tag usage

## Recommendations

1. **Start with UTF-8**: Implement UTF-8 support first (most common)
2. **Defer ANSEL**: Add ANSEL support in Phase 2 (requires library)
3. **Add validation early**: Tag/XREF/Level validation in Phase 1
4. **Date/Place parsing**: Separate packages, implement in Phase 3
5. **Spec compliance**: Create test cases from official spec examples
6. **Backward compatibility**: Support common variations (case-insensitive tags?)

## Next Steps

1. ✅ Review this document
2. ⏳ Verify specific points from GEDCOM 5.5.1 PDF
3. ⏳ Update IMPLEMENTATION_PLAN.md with these adjustments
4. ⏳ Create validation utilities
5. ⏳ Start Phase 1 implementation

