# GEDCOM Diff Documentation

## Overview

The GEDCOM diff system performs semantic comparison of two GEDCOM files, identifying differences at a meaningful level rather than simple line-by-line text comparison. It understands GEDCOM structure and reports changes in terms of added/removed records, modified fields, relationship changes, and structural differences.

The system also tracks change history, recording when, what, and optionally who made each change.

## Features

- **Semantic Understanding**: Understands GEDCOM structure, not just text
- **Multiple Matching Strategies**: XREF-based, content-based, or hybrid
- **Field-Level Comparison**: Detailed field-by-field changes
- **Semantic Equivalence**: Recognizes equivalent values (e.g., "1800" ≈ "ABT 1800")
- **Change History Tracking**: Records timestamp, field, and values for each change
- **Multiple Output Formats**: Text, JSON (coming soon), HTML (coming soon)
- **Performance Optimized**: Fast comparison with indexing

## Installation

The diff system is part of the `pkg/gedcom/diff` package:

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/diff"
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/diff"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
)

func main() {
    // Parse two GEDCOM files
    p := parser.NewHierarchicalParser()
    tree1, _ := p.Parse("family_v1.ged")
    tree2, _ := p.Parse("family_v2.ged")

    // Create differ
    differ := diff.NewGedcomDiffer(diff.DefaultConfig())

    // Compare files
    result, err := differ.Compare(tree1, tree2)
    if err != nil {
        panic(err)
    }

    // Generate report
    report, err := differ.GenerateReport(result)
    if err != nil {
        panic(err)
    }

    fmt.Println(report)
}
```

### Compare Specific Records

```go
// Compare two individuals with same XREF
indi1 := tree1.GetIndividual("@I1@").(*gedcom.IndividualRecord)
indi2 := tree2.GetIndividual("@I1@").(*gedcom.IndividualRecord)

// The differ will automatically detect this as a modification
result, _ := differ.Compare(tree1, tree2)

// Find modifications for @I1@
for _, mod := range result.Changes.Modified {
    if mod.Xref == "@I1@" {
        fmt.Printf("Changes to @I1@:\n")
        for _, change := range mod.Changes {
            fmt.Printf("  %s: %v → %v\n", 
                change.Path, change.OldValue, change.NewValue)
        }
    }
}
```

## Configuration

### Default Configuration

```go
config := diff.DefaultConfig()
// Returns:
//   MatchingStrategy: "xref"
//   SimilarityThreshold: 0.85
//   DateTolerance: 2
//   IncludeUnchanged: false
//   DetailLevel: "field"
//   OutputFormat: "text"
//   TrackHistory: true
```

### Custom Configuration

```go
config := &diff.DiffConfig{
    MatchingStrategy:   "hybrid",  // "xref", "content", or "hybrid"
    SimilarityThreshold: 0.85,     // For content matching
    DateTolerance:      2,         // Years tolerance for date equivalence
    IncludeUnchanged:   false,     // Include unchanged records
    DetailLevel:        "field",   // "summary", "field", or "full"
    OutputFormat:       "text",    // "text", "json", "html", "unified"
    TrackHistory:       true,      // Track change history
}

differ := diff.NewGedcomDiffer(config)
```

## Matching Strategies

### XREF-Based Matching (Default)

Matches records by their XREF IDs. Fast and accurate for same-file versions.

**Use Case:** Version comparison, same-file edits

```go
config.MatchingStrategy = "xref"
```

**Pros:**
- Very fast (O(n) complexity)
- Accurate for same-file versions
- Preserves record identity

**Cons:**
- Doesn't work for cross-file comparison
- Fails when XREFs differ

### Content-Based Matching

Uses duplicate detection to match records by content similarity.

**Use Case:** Cross-file comparison, merge preparation

```go
config.MatchingStrategy = "content"
```

**Pros:**
- Works across different files
- Handles XREF mismatches
- Finds semantic matches

**Cons:**
- Slower (requires similarity calculations)
- May have false positives
- Requires threshold configuration

### Hybrid Matching (Recommended)

Tries XREF first, falls back to content matching for unmatched records.

**Use Case:** General-purpose comparison

```go
config.MatchingStrategy = "hybrid"
```

**Pros:**
- Fast for same-file, accurate for cross-file
- Best of both worlds
- Configurable behavior

**Cons:**
- More complex implementation

## Change Types

### Added Records

Records that exist in file 2 but not in file 1.

```go
for _, added := range result.Changes.Added {
    fmt.Printf("Added: %s (%s)\n", added.Xref, added.Type)
}
```

### Removed Records

Records that exist in file 1 but not in file 2.

```go
for _, removed := range result.Changes.Removed {
    fmt.Printf("Removed: %s (%s)\n", removed.Xref, removed.Type)
}
```

### Modified Records

Records with the same XREF but different content.

```go
for _, mod := range result.Changes.Modified {
    fmt.Printf("Modified: %s\n", mod.Xref)
    for _, change := range mod.Changes {
        switch change.Type {
        case diff.ChangeTypeModified:
            fmt.Printf("  %s: %v → %v\n", 
                change.Path, change.OldValue, change.NewValue)
        case diff.ChangeTypeAdded:
            fmt.Printf("  %s: Added (%v)\n", change.Path, change.NewValue)
        case diff.ChangeTypeRemoved:
            fmt.Printf("  %s: Removed (%v)\n", change.Path, change.OldValue)
        case diff.ChangeTypeSemanticallyEquivalent:
            fmt.Printf("  %s: %v → %v (equivalent)\n",
                change.Path, change.OldValue, change.NewValue)
        }
    }
}
```

## Field Comparison

### Individual Record Fields

The differ compares:
- **Name**: Full name, given name, surname
- **Sex**: M, F, U
- **Birth Date**: With semantic equivalence
- **Birth Place**: With hierarchy matching
- **Death Date**: Added/removed/modified
- **Death Place**: Added/removed/modified

### Family Record Fields

The differ compares:
- **Husband**: XREF changes
- **Wife**: XREF changes
- **Children**: Added/removed children
- **Marriage Date**: With semantic equivalence
- **Marriage Place**: With hierarchy matching

### Semantic Equivalence

The differ recognizes equivalent values:

**Dates:**
```go
"1800" ≈ "ABT 1800"        // Within tolerance
"BEF 1850" ≈ "AFT 1840"    // Overlapping ranges
```

**Places:**
```go
"New York" ≈ "New York, NY"
"New York, NY" ≈ "New York, New York, USA"
```

## Change History

### Overview

When `TrackHistory` is enabled, the system records:
- **Timestamp**: When the change was detected
- **Field**: Which field changed
- **Old/New Values**: Previous and current values
- **Change Type**: Added, removed, modified, or semantically equivalent

### Accessing History

```go
// Global change history
for _, entry := range result.History {
    fmt.Printf("[%s] %s: %s\n",
        entry.Timestamp.Format(time.RFC3339),
        entry.ChangeType,
        entry.Field)
    fmt.Printf("  %s → %s\n", entry.OldValue, entry.NewValue)
}

// Record-level history
for _, mod := range result.Changes.Modified {
    for _, entry := range mod.History {
        fmt.Printf("Record %s: %s\n", mod.Xref, entry.Field)
    }
}

// Field-level history
for _, mod := range result.Changes.Modified {
    for _, change := range mod.Changes {
        for _, entry := range change.History {
            fmt.Printf("Field %s: %s\n", change.Path, entry.Field)
        }
    }
}
```

### Future: Author and Reason

The `ChangeHistory` structure includes fields for:
- `Author`: Who made the change (optional)
- `Reason`: Why the change was made (optional)

These can be populated when integrating with version control or user tracking systems.

## Output Formats

### Text Format (Default)

Human-readable text report with sections for:
- Summary statistics
- Added records
- Removed records
- Modified records with field changes
- Change history
- Performance statistics

```go
report, err := differ.GenerateReport(result)
fmt.Println(report)
```

**Example Output:**
```
GEDCOM Diff Report
==================================================

Summary:
  File 1: 100 individuals, 50 families, 20 notes, 15 sources
  File 2: 105 individuals, 52 families, 22 notes, 16 sources

  Changes:
    Added:     5 individuals, 2 families
    Removed:   2 individuals, 0 families
    Modified:  3 individuals, 0 families
    Unchanged: 95 individuals, 50 families

Added Records:
--------------------------------------------------
  @I105@: INDI
    Name: John /Smith/
    Birth: 1850, New York

Modified Records:
--------------------------------------------------
  @I1@: INDI
    NAME: John /Doe/ → John /Doe Jr/
    BIRT.DATE: 1800 → ABT 1800 (semantically equivalent)
    DEAT.DATE: Added (1870)
```

### JSON Format (Coming Soon)

Structured JSON for programmatic processing.

### HTML Format (Coming Soon)

Visual HTML diff with color coding.

### Unified Diff Format (Coming Soon)

Git-style unified diff for version control.

## API Reference

### Types

#### GedcomDiffer

Main differ struct for comparing GEDCOM files.

```go
type GedcomDiffer struct {
    config *DiffConfig
}
```

#### DiffConfig

Configuration for GEDCOM comparison.

```go
type DiffConfig struct {
    MatchingStrategy   string
    SimilarityThreshold float64
    DateTolerance      int
    IncludeUnchanged   bool
    DetailLevel        string
    OutputFormat       string
    TrackHistory       bool
}
```

#### DiffResult

Contains complete diff results.

```go
type DiffResult struct {
    Summary    DiffSummary
    Changes    DiffChanges
    Statistics DiffStatistics
    History    []ChangeHistory
}
```

#### ChangeHistory

Tracks individual changes.

```go
type ChangeHistory struct {
    Timestamp  time.Time
    Author     string
    Reason     string
    ChangeType ChangeType
    Field      string
    OldValue   string
    NewValue   string
}
```

### Methods

#### NewGedcomDiffer

Creates a new GEDCOM differ.

```go
func NewGedcomDiffer(config *DiffConfig) *GedcomDiffer
```

#### Compare

Compares two GEDCOM trees.

```go
func (gd *GedcomDiffer) Compare(tree1, tree2 *gedcom.GedcomTree) (*DiffResult, error)
```

#### CompareFiles

Compares two GEDCOM files from disk (coming soon).

```go
func (gd *GedcomDiffer) CompareFiles(file1, file2 string) (*DiffResult, error)
```

#### GenerateReport

Generates a text report from diff results.

```go
func (gd *GedcomDiffer) GenerateReport(result *DiffResult) (string, error)
```

## Examples

### Example 1: Basic Comparison

```go
differ := diff.NewGedcomDiffer(diff.DefaultConfig())
result, _ := differ.Compare(tree1, tree2)

fmt.Printf("Summary:\n")
fmt.Printf("  Added: %d\n", len(result.Changes.Added))
fmt.Printf("  Removed: %d\n", len(result.Changes.Removed))
fmt.Printf("  Modified: %d\n", len(result.Changes.Modified))
```

### Example 2: Detailed Field Changes

```go
result, _ := differ.Compare(tree1, tree2)

for _, mod := range result.Changes.Modified {
    fmt.Printf("Record %s changes:\n", mod.Xref)
    for _, change := range mod.Changes {
        fmt.Printf("  %s: %v → %v (%s)\n",
            change.Path,
            change.OldValue,
            change.NewValue,
            change.Type)
    }
}
```

### Example 3: Change History

```go
config := diff.DefaultConfig()
config.TrackHistory = true
differ := diff.NewGedcomDiffer(config)

result, _ := differ.Compare(tree1, tree2)

// Print change history
for _, entry := range result.History {
    fmt.Printf("[%s] %s changed: %s\n",
        entry.Timestamp.Format("2006-01-02 15:04:05"),
        entry.Field,
        entry.ChangeType)
    fmt.Printf("  %s → %s\n", entry.OldValue, entry.NewValue)
}
```

### Example 4: Filter Changes

```go
result, _ := differ.Compare(tree1, tree2)

// Only show added records
for _, added := range result.Changes.Added {
    if added.Type == "INDI" {
        fmt.Printf("New individual: %s\n", added.Xref)
    }
}

// Only show modified names
for _, mod := range result.Changes.Modified {
    for _, change := range mod.Changes {
        if change.Field == "NAME" {
            fmt.Printf("Name changed: %s\n", mod.Xref)
        }
    }
}
```

### Example 5: Semantic Equivalence

```go
result, _ := differ.Compare(tree1, tree2)

// Find semantically equivalent changes
for _, mod := range result.Changes.Modified {
    for _, change := range mod.Changes {
        if change.Type == diff.ChangeTypeSemanticallyEquivalent {
            fmt.Printf("%s: %v ≈ %v (equivalent)\n",
                change.Path,
                change.OldValue,
                change.NewValue)
        }
    }
}
```

## Performance

### Optimization Features

1. **Indexing**: Fast XREF lookup (O(1))
2. **Pre-filtering**: Reduces comparison space
3. **Early Termination**: Skips identical records

### Expected Performance

| File Size | Comparison Time (XREF) | Comparison Time (Content) |
|-----------|------------------------|---------------------------|
| 100 records | < 1 second | ~5 seconds |
| 1,000 records | ~1 second | ~1 minute |
| 10,000 records | ~5 seconds | ~10 minutes |

**Note:** Content-based comparison is slower due to similarity calculations.

## Best Practices

### 1. Choose Appropriate Matching Strategy

- **Same-file versions**: Use "xref" (fastest)
- **Cross-file comparison**: Use "content" or "hybrid"
- **General use**: Use "hybrid" (balanced)

### 2. Enable Change History

```go
config.TrackHistory = true
```

Provides audit trail and detailed change tracking.

### 3. Use Semantic Equivalence

The system automatically detects equivalent values:
- Dates within tolerance
- Places with hierarchy matching

No additional configuration needed.

### 4. Review Summary First

```go
summary := result.Summary
fmt.Printf("Changes: %d added, %d removed, %d modified\n",
    len(result.Changes.Added),
    len(result.Changes.Removed),
    len(result.Changes.Modified))
```

### 5. Check Statistics

```go
fmt.Printf("Processing time: %v\n", result.Statistics.ProcessingTime)
fmt.Printf("Records compared: %d\n", result.Statistics.RecordsCompared)
```

## Troubleshooting

### No Differences Found

**Possible causes:**
- Files are identical
- Matching strategy too strict
- XREF mismatches (use content matching)

**Solutions:**
- Verify files are different
- Use "content" or "hybrid" strategy
- Check XREF consistency

### Too Many Differences

**Possible causes:**
- Files are completely different
- Semantic equivalence not working
- Date tolerance too strict

**Solutions:**
- Verify you're comparing correct files
- Increase `DateTolerance`
- Check semantic equivalence settings

### Slow Performance

**Possible causes:**
- Very large files
- Content-based matching
- No indexing

**Solutions:**
- Use "xref" strategy if possible
- Compare smaller subsets
- Wait for parallel processing (coming soon)

## Integration

### With Duplicate Detection

The diff system can use duplicate detection for content matching:

```go
config.MatchingStrategy = "content"
// Automatically uses duplicate detection internally
```

### With Query API

```go
// Future integration
q.Diff(tree1, tree2).Strategy("hybrid").Execute()
```

### With CLI

```bash
# Future CLI command
gedcom diff file1.ged file2.ged --format json -o diff.json
```

## See Also

- [GEDCOM Diff Design](GEDCOM_DIFF_DESIGN.md) - Detailed design document
- [Duplicate Detection Documentation](duplicate-detection.md) - Duplicate detection system
- [Query API Documentation](query-api.md) - Graph-based querying
- [Types Documentation](types.md) - Core GEDCOM types
