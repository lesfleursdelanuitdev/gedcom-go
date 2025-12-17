# GEDCOM Validator Documentation

Complete reference guide for validating GEDCOM files.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Validator Types](#validator-types)
  - [GedcomValidator](#gedcomvalidator)
  - [ParallelGedcomValidator](#parallelgedcomvalidator)
  - [AdvancedValidator](#advancedvalidator)
- [Specialized Validators](#specialized-validators)
  - [IndividualValidator](#individualvalidator)
  - [FamilyValidator](#familyvalidator)
  - [CrossReferenceValidator](#crossreferencevalidator)
  - [HeaderValidator](#headervalidator)
  - [DateConsistencyValidator](#dateconsistencyvalidator)
- [Basic Usage](#basic-usage)
- [Advanced Validation](#advanced-validation)
- [Error Severity Levels](#error-severity-levels)
- [Validation Rules](#validation-rules)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Performance Considerations](#performance-considerations)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The validator package provides comprehensive validation functionality for GEDCOM trees based on the GEDCOM 5.5.1 specification. It checks record structure, cross-references, required fields, data format compliance, and advanced data quality rules.

### Features

- **Comprehensive Validation**: Checks all GEDCOM record types
- **Cross-Reference Validation**: Validates all xref links
- **Structure Validation**: Validates tag structure and hierarchy
- **Advanced Validation**: Date consistency, relationship logic, data quality
- **Error Severity Levels**: Categorizes errors by severity (Severe, Warning, Info, Hint)
- **Parallel Validation**: Optional parallel validators for performance
- **Pluggable Rules**: Extensible rule system for advanced validation

---

## Installation

The validator package is part of the GEDCOM Go library:

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/internal/validator"
```

---

## Validator Types

### GedcomValidator

The **GedcomValidator** orchestrates all validators and validates the entire GEDCOM tree. This is the recommended validator for most use cases.

#### Features

- Runs all specialized validators
- Validates header, individuals, families, cross-references
- Optional advanced validation
- Collects all errors without stopping

#### Usage

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/validator"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func main() {
    // Parse file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Validate
    errorManager := gedcom.NewErrorManager()
    v := validator.NewGedcomValidator(errorManager)
    err = v.Validate(tree)
    if err != nil {
        panic(err)
    }

    // Check errors
    errors := errorManager.Errors()
    if len(errors) > 0 {
        fmt.Printf("Found %d validation errors:\n", len(errors))
        for _, err := range errors {
            fmt.Printf("  [%s] %s (Line %d)\n", 
                err.Severity, err.Message, err.LineNumber)
        }
    } else {
        fmt.Println("✓ Validation passed")
    }
}
```

#### Enable Advanced Validation

```go
v := validator.NewGedcomValidator(errorManager)
v.EnableAdvancedValidation() // Enable with default rules
err := v.Validate(tree)
```

---

### ParallelGedcomValidator

The **ParallelGedcomValidator** runs validators in parallel for better performance on large files.

#### Features

- Parallel execution of validators
- Thread-safe error collection
- Same validation rules as GedcomValidator
- Better performance for large files

#### Usage

```go
errorManager := gedcom.NewErrorManager()
parallelValidator := validator.NewParallelGedcomValidator(errorManager)
err := parallelValidator.Validate(tree)
```

**When to Use:**
- Large files (>10,000 records)
- Performance is critical
- Multiple CPU cores available

---

### AdvancedValidator

The **AdvancedValidator** provides a pluggable rule system for advanced data quality and consistency checks beyond basic GEDCOM compliance.

#### Features

- Pluggable validation rules
- Configurable thresholds
- Severity filtering
- Extensible rule system

#### Usage

```go
errorManager := gedcom.NewErrorManager()
advancedValidator := validator.NewAdvancedValidator(errorManager)

// Add validation rules
advancedValidator.AddRule(validator.NewDateConsistencyValidator(errorManager))

// Validate
err := advancedValidator.Validate(tree)
```

#### Custom Configuration

```go
config := validator.NewValidationConfig()
config.MinParentAge = 12
config.MaxParentAge = 80
config.MinSeverity = gedcom.SeverityWarning

advancedValidator := validator.NewAdvancedValidatorWithConfig(errorManager, config)
advancedValidator.AddRule(validator.NewDateConsistencyValidator(errorManager))
err := advancedValidator.Validate(tree)
```

---

## Specialized Validators

### IndividualValidator

Validates Individual (INDI) records.

#### Validation Rules

- **Required Tags**: NAME (required)
- **Valid Tags**: All valid GEDCOM tags for individuals
- **Sex Validation**: Valid values (M, F, U, X, N)
- **Event Validation**: Valid event structure and subtags
- **Name Validation**: Valid name component structure
- **Cross-References**: FAMS, FAMC references

#### Valid Tags

Includes: RESN, NAME, SEX, ALIA, ASSO, BIRT, DEAT, BURI, CREM, ADOP, and all standard GEDCOM individual tags.

#### Valid Sex Values

- `M`: Male
- `F`: Female
- `U`: Unknown
- `X`: Other
- `N`: Not applicable

#### Example

```go
errorManager := gedcom.NewErrorManager()
individualValidator := validator.NewIndividualValidator(errorManager)
err := individualValidator.Validate(tree)
```

---

### FamilyValidator

Validates Family (FAM) records.

#### Validation Rules

- **Required Tags**: At least one of HUSB or WIFE (required)
- **Valid Tags**: All valid GEDCOM tags for families
- **Event Validation**: Valid event structure (MARR, DIV, etc.)
- **Cross-References**: HUSB, WIFE, CHIL references

#### Valid Tags

Includes: RESN, HUSB, WIFE, CHIL, NCHI, MARR, DIV, ANUL, and all standard GEDCOM family tags.

#### Example

```go
errorManager := gedcom.NewErrorManager()
familyValidator := validator.NewFamilyValidator(errorManager)
err := familyValidator.Validate(tree)
```

---

### CrossReferenceValidator

Validates cross-references between records.

#### Validation Rules

- **Xref Format**: Valid xref ID format (@XREF@)
- **Reference Resolution**: All references point to existing records
- **Bidirectional Links**: Family-individual links are consistent

#### Validated References

- **Individual → Family**: FAMS, FAMC references
- **Family → Individual**: HUSB, WIFE, CHIL references

#### Example

```go
errorManager := gedcom.NewErrorManager()
crossRefValidator := validator.NewCrossReferenceValidator(errorManager)
err := crossRefValidator.Validate(tree)
```

---

### HeaderValidator

Validates Header (HEAD) record.

#### Validation Rules

- **Header Exists**: HEAD record is present
- **Valid Tags**: All tags are valid GEDCOM header tags
- **GEDC Structure**: GEDC tag is present
- **GEDC.VERS**: Version tag is recommended

#### Valid Tags

Includes: GEDC, CHAR, SOUR, DATE, TIME, FILE, LANG, SUBM, SUBN, COPR, DEST, NOTE.

#### Example

```go
errorManager := gedcom.NewErrorManager()
headerValidator := validator.NewHeaderValidator(errorManager)
err := headerValidator.Validate(tree)
```

---

### DateConsistencyValidator

Validates date consistency across records (advanced validation).

#### Validation Rules

- **Birth Before Death**: Death date must be after birth date
- **Reasonable Ages**: Age at death, marriage, etc. within reasonable limits
- **Parent-Child Gaps**: Parents must be old enough when child is born
- **Spouse Age Gaps**: Spouses have reasonable age differences
- **Marriage Before Birth**: Marriage must be after birth
- **Historical Dates**: Dates are within reasonable historical range

#### Configuration

```go
config := validator.NewValidationConfig()
config.MinParentAge = 10      // Minimum age for parent at child's birth
config.MaxParentAge = 80      // Maximum age for parent at child's birth
config.MinMarriageAge = 12    // Minimum age for marriage
config.MaxMarriageAge = 100   // Maximum age for marriage
config.MaxDeathAge = 120      // Maximum reasonable age at death
config.SpouseAgeGapWarn = 30 // Age gap to trigger warning
config.SpouseAgeGapHint = 40 // Age gap to trigger hint
```

#### Example

```go
errorManager := gedcom.NewErrorManager()
dateValidator := validator.NewDateConsistencyValidator(errorManager)

config := validator.NewValidationConfig()
config.MinParentAge = 12
config.MaxParentAge = 80

advancedValidator := validator.NewAdvancedValidatorWithConfig(errorManager, config)
advancedValidator.AddRule(dateValidator)
err := advancedValidator.Validate(tree)
```

---

## Basic Usage

### Simple Validation

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/validator"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func main() {
    // Parse file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Validate
    errorManager := gedcom.NewErrorManager()
    v := validator.NewGedcomValidator(errorManager)
    err = v.Validate(tree)
    if err != nil {
        panic(err)
    }

    // Check errors
    errors := errorManager.Errors()
    if len(errors) > 0 {
        fmt.Printf("Found %d validation errors\n", len(errors))
    }
}
```

### Validation with Error Filtering

```go
v := validator.NewGedcomValidator(errorManager)
err := v.Validate(tree)

// Filter errors by severity
allErrors := errorManager.Errors()
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

## Advanced Validation

### Enable Advanced Validation

```go
v := validator.NewGedcomValidator(errorManager)
v.EnableAdvancedValidation() // Enable with defaults
err := v.Validate(tree)
```

### Custom Advanced Validation

```go
config := validator.NewValidationConfig()
config.MinParentAge = 12
config.MaxParentAge = 80
config.MinMarriageAge = 14
config.MaxMarriageAge = 100
config.MaxDeathAge = 120
config.MinSeverity = gedcom.SeverityWarning // Only show warnings and severe

v := validator.NewGedcomValidator(errorManager)
v.EnableAdvancedValidationWithConfig(config)
err := v.Validate(tree)
```

### Create Custom Validation Rule

```go
type CustomRule struct {
    *validator.BaseValidator
}

func (cr *CustomRule) Name() string {
    return "Custom Rule"
}

func (cr *CustomRule) Description() string {
    return "Validates custom business logic"
}

func (cr *CustomRule) Validate(tree *gedcom.GedcomTree, config *validator.ValidationConfig) []*gedcom.GedcomError {
    errors := []*gedcom.GedcomError{}
    
    // Your validation logic here
    individuals := tree.GetAllIndividuals()
    for xref, indi := range individuals {
        // Check something
        if /* condition */ {
            errors = append(errors, &gedcom.GedcomError{
                Severity:   gedcom.SeverityWarning,
                Message:    fmt.Sprintf("Custom validation failed for %s", xref),
                LineNumber: indi.FirstLine().LineNumber,
                Context:    "Custom Rule",
            })
        }
    }
    
    return errors
}

// Usage
errorManager := gedcom.NewErrorManager()
advancedValidator := validator.NewAdvancedValidator(errorManager)
advancedValidator.AddRule(&CustomRule{
    BaseValidator: validator.NewBaseValidator(errorManager),
})
err := advancedValidator.Validate(tree)
```

---

## Error Severity Levels

Errors are categorized by severity to help prioritize fixes:

### Severe

**Critical errors that must be fixed:**
- Missing required tags
- Invalid cross-references
- Invalid tag structure
- Death before birth
- Marriage before birth

**Example:**
```
[SEVERE] INDI @I1@: Missing required tag NAME (Line 5)
[SEVERE] FAM @F1@: Invalid cross-reference: @I99@ (Line 10)
```

### Warning

**Issues that should be reviewed:**
- Invalid tag values
- Multiple events of same type
- Unusual but possible situations
- Age at death exceeds maximum

**Example:**
```
[WARNING] INDI @I1@: Invalid SEX value 'X' (Line 7)
[WARNING] INDI @I2@: Multiple BIRT events found (Line 12)
```

### Info

**Data quality issues and suggestions:**
- Missing optional but recommended data
- Missing death date for very old individuals
- Missing birth date

**Example:**
```
[INFO] INDI @I1@: Missing birth date (Line 5)
[INFO] INDI @I2@: Missing death date (individual would be 105 years old) (Line 8)
```

### Hint

**Best practices and optimizations:**
- Suggestions for improvement
- Optional enhancements
- Data quality hints

**Example:**
```
[HINT] INDI @I1@: Consider adding place information for birth event (Line 6)
[HINT] FAM @F1@: Large age gap between spouses (40 years) (Line 15)
```

---

## Validation Rules

### Individual Validation Rules

| Rule | Severity | Description |
|------|----------|-------------|
| Missing NAME | Severe | Individual must have at least one NAME tag |
| Invalid SEX value | Warning | SEX must be M, F, U, X, or N |
| Invalid tag | Severe | Tag is not valid for individual records |
| Multiple BIRT events | Warning | Should have only one birth event |
| Multiple DEAT events | Warning | Should have only one death event |
| Invalid FAMS reference | Severe | FAMS xref does not exist |
| Invalid FAMC reference | Severe | FAMC xref does not exist |
| Invalid event structure | Warning | Event subtags are invalid |

### Family Validation Rules

| Rule | Severity | Description |
|------|----------|-------------|
| Missing HUSB/WIFE | Severe | Family must have at least one spouse |
| Invalid HUSB reference | Severe | HUSB xref does not exist |
| Invalid WIFE reference | Severe | WIFE xref does not exist |
| Invalid CHIL reference | Severe | CHIL xref does not exist |
| Multiple MARR events | Warning | Should have only one marriage event |
| Multiple DIV events | Warning | Should have only one divorce event |
| Invalid tag | Severe | Tag is not valid for family records |

### Cross-Reference Validation Rules

| Rule | Severity | Description |
|------|----------|-------------|
| Invalid xref format | Severe | Xref ID must be @XREF@ format |
| Broken reference | Severe | Reference points to non-existent record |
| Missing bidirectional link | Warning | Family-individual links are inconsistent |

### Header Validation Rules

| Rule | Severity | Description |
|------|----------|-------------|
| Missing HEAD | Severe | File must have HEAD record |
| Missing GEDC | Severe | HEAD must have GEDC tag |
| Missing GEDC.VERS | Warning | GEDC.VERS is recommended |
| Invalid tag | Warning | Tag is not valid for header |

### Date Consistency Rules (Advanced)

| Rule | Severity | Description |
|------|----------|-------------|
| Death before birth | Severe | Death date must be after birth date |
| Marriage before birth | Severe | Marriage date must be after birth date |
| Age at death too high | Warning | Age at death exceeds maximum (default: 120) |
| Parent too young | Warning | Parent age at child's birth too low (default: <10) |
| Parent too old | Warning | Parent age at child's birth too high (default: >80) |
| Spouse age gap large | Warning/Hint | Large age gap between spouses (configurable) |
| Missing birth date | Info | Birth date is missing |
| Missing death date (old) | Info | Death date missing for very old individual |

---

## API Reference

### GedcomValidator

#### Constructor

```go
func NewGedcomValidator(errorManager *gedcom.ErrorManager) *GedcomValidator
```

#### Methods

##### Validate

```go
func (gv *GedcomValidator) Validate(tree *gedcom.GedcomTree) error
```

Runs all validators on the tree. Returns error only for fatal issues.

##### EnableAdvancedValidation

```go
func (gv *GedcomValidator) EnableAdvancedValidation()
```

Enables advanced validation with default configuration.

##### EnableAdvancedValidationWithConfig

```go
func (gv *GedcomValidator) EnableAdvancedValidationWithConfig(config *ValidationConfig)
```

Enables advanced validation with custom configuration.

##### GetErrorManager

```go
func (gv *GedcomValidator) GetErrorManager() *gedcom.ErrorManager
```

Returns the error manager.

##### GetAdvancedValidator

```go
func (gv *GedcomValidator) GetAdvancedValidator() *AdvancedValidator
```

Returns the advanced validator (if enabled).

### ValidationConfig

```go
type ValidationConfig struct {
    // Age thresholds
    MinParentAge     int
    MaxParentAge     int
    MinMarriageAge   int
    MaxMarriageAge   int
    MaxDeathAge      int
    SpouseAgeGapWarn int
    SpouseAgeGapHint int

    // Date thresholds
    MinHistoricalDate int
    MaxFutureDate     int
    DateRangeWarn     int

    // Duplicate detection
    NameSimilarity float64
    DateSimilarity int

    // Severity filtering
    MinSeverity gedcom.ErrorSeverity
}
```

---

## Examples

### Complete Validation Example

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/validator"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func main() {
    // Parse file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Validate with advanced rules
    errorManager := gedcom.NewErrorManager()
    v := validator.NewGedcomValidator(errorManager)
    v.EnableAdvancedValidation()
    
    err = v.Validate(tree)
    if err != nil {
        panic(err)
    }

    // Report errors by severity
    errors := errorManager.Errors()
    if len(errors) == 0 {
        fmt.Println("✓ Validation passed")
        return
    }

    // Group by severity
    severe := []*gedcom.GedcomError{}
    warnings := []*gedcom.GedcomError{}
    info := []*gedcom.GedcomError{}
    hints := []*gedcom.GedcomError{}

    for _, err := range errors {
        switch err.Severity {
        case gedcom.SeveritySevere:
            severe = append(severe, err)
        case gedcom.SeverityWarning:
            warnings = append(warnings, err)
        case gedcom.SeverityInfo:
            info = append(info, err)
        case gedcom.SeverityHint:
            hints = append(hints, err)
        }
    }

    fmt.Printf("Validation Results:\n")
    fmt.Printf("  Severe: %d\n", len(severe))
    fmt.Printf("  Warnings: %d\n", len(warnings))
    fmt.Printf("  Info: %d\n", len(info))
    fmt.Printf("  Hints: %d\n", len(hints))

    // Print severe errors
    if len(severe) > 0 {
        fmt.Printf("\nSevere Errors:\n")
        for _, err := range severe {
            fmt.Printf("  [Line %d] %s\n", err.LineNumber, err.Message)
        }
    }
}
```

### Custom Validation Configuration

```go
config := validator.NewValidationConfig()

// Age thresholds
config.MinParentAge = 12      // Minimum age for parent
config.MaxParentAge = 80      // Maximum age for parent
config.MinMarriageAge = 14    // Minimum marriage age
config.MaxMarriageAge = 100   // Maximum marriage age
config.MaxDeathAge = 120      // Maximum age at death

// Spouse age gaps
config.SpouseAgeGapWarn = 30  // Warning at 30 years
config.SpouseAgeGapHint = 40  // Hint at 40 years

// Date thresholds
config.MinHistoricalDate = 500 // Minimum historical date
config.MaxFutureDate = 2026   // Maximum future date

// Severity filtering
config.MinSeverity = gedcom.SeverityWarning // Only warnings and severe

// Use configuration
errorManager := gedcom.NewErrorManager()
v := validator.NewGedcomValidator(errorManager)
v.EnableAdvancedValidationWithConfig(config)
err := v.Validate(tree)
```

### Parallel Validation

```go
errorManager := gedcom.NewErrorManager()
parallelValidator := validator.NewParallelGedcomValidator(errorManager)
err := parallelValidator.Validate(tree)

if err != nil {
    panic(err)
}

errors := errorManager.Errors()
fmt.Printf("Found %d validation issues\n", len(errors))
```

---

## Performance Considerations

### Validator Selection

| Validator | Best For | Performance |
|-----------|----------|-------------|
| **GedcomValidator** | Most files | Fast |
| **ParallelGedcomValidator** | Large files (>10K records) | Very Fast |
| **AdvancedValidator** | Data quality checks | Slower (more checks) |

### Performance Tips

1. **Use parallel validator for large files** (>10,000 records)
2. **Disable advanced validation** if not needed (faster)
3. **Filter by severity** to reduce error processing
4. **Reuse error managers** when validating multiple files

---

## Best Practices

### Error Handling

Always check for errors after validation:

```go
v := validator.NewGedcomValidator(errorManager)
err := v.Validate(tree)
if err != nil {
    // Handle fatal error
    log.Fatalf("Validation failed: %v", err)
}

// Check for non-fatal errors
errors := errorManager.Errors()
if len(errors) > 0 {
    // Process errors
}
```

### Severity Filtering

Filter errors by severity to focus on important issues:

```go
errors := errorManager.Errors()
severeErrors := []*gedcom.GedcomError{}

for _, err := range errors {
    if err.Severity == gedcom.SeveritySevere {
        severeErrors = append(severeErrors, err)
    }
}

// Fix severe errors first
```

### Configuration

Use appropriate configuration for your data:

```go
config := validator.NewValidationConfig()

// Adjust for historical data
config.MinHistoricalDate = 0  // Allow very old dates
config.MaxParentAge = 100     // Higher for historical data

// Adjust for different cultures
config.MinMarriageAge = 10    // Historical minimum
config.SpouseAgeGapWarn = 50  // Larger gaps acceptable
```

---

## Troubleshooting

### Common Issues

#### 1. "Missing required tag" Errors

**Problem:** Required tags are missing.

**Solutions:**
- Add missing tags (e.g., NAME for individuals)
- Check if tags are at correct level
- Verify GEDCOM file structure

#### 2. "Invalid cross-reference" Errors

**Problem:** References point to non-existent records.

**Solutions:**
- Check xref IDs are correct
- Verify referenced records exist
- Check for typos in xref IDs

#### 3. "Death before birth" Errors

**Problem:** Date consistency validation fails.

**Solutions:**
- Check date values are correct
- Verify date parsing is correct
- Review date format

#### 4. Too Many Warnings/Info/Hints

**Problem:** Too many low-severity errors.

**Solutions:**
- Filter by severity: `config.MinSeverity = gedcom.SeverityWarning`
- Disable advanced validation if not needed
- Focus on severe errors first

### Debugging

Enable verbose error reporting:

```go
errors := errorManager.Errors()
for _, err := range errors {
    fmt.Printf("[%s] Line %d: %s (Context: %s)\n",
        err.Severity, err.LineNumber, err.Message, err.Context)
}
```

---

## See Also

- [CLI Documentation](cli.md) - Command-line validation
- [Parser Documentation](parser.md) - Parsing GEDCOM files
- [GEDCOM Specification](https://www.gedcom.org/) - Official GEDCOM 5.5.1 specification

---

**Last Updated:** 2025-01-27
