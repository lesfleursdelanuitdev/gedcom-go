# Duplicate Detection Documentation

## Overview

The duplicate detection system identifies potential duplicate individuals within a single GEDCOM file or across multiple files. It uses weighted similarity scoring based on names, dates, places, sex, and relationships to determine the likelihood that two records represent the same person.

## Features

- **Weighted Similarity Scoring**: Combines multiple metrics (name, date, place, sex, relationships)
- **Phonetic Matching**: Soundex algorithm for name variations
- **Semantic Date Comparison**: Handles imprecise dates (ABT, BEF, AFT, BETWEEN)
- **Relationship Matching**: Uses family relationships (parents, spouses, children)
- **Parallel Processing**: Multi-threaded comparison for large files
- **Performance Optimizations**: Indexing, pre-filtering, memory pooling
- **Configurable Thresholds**: Adjustable sensitivity levels

## Installation

The duplicate detection system is part of the `pkg/gedcom/duplicate` package:

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/duplicate"
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/duplicate"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
)

func main() {
    // Parse GEDCOM file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Create duplicate detector
    detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())

    // Find duplicates
    result, err := detector.FindDuplicates(tree)
    if err != nil {
        panic(err)
    }

    // Process matches
    for _, match := range result.Matches {
        fmt.Printf("Potential duplicate: %s and %s\n",
            match.Individual1.XrefID(),
            match.Individual2.XrefID())
        fmt.Printf("  Similarity: %.2f%%\n", match.SimilarityScore*100)
        fmt.Printf("  Confidence: %s\n", match.Confidence)
        fmt.Printf("  Matching fields: %v\n", match.MatchingFields)
    }
}
```

### Cross-File Comparison

```go
// Parse two files
tree1, _ := p.Parse("file1.ged")
tree2, _ := p.Parse("file2.ged")

// Find duplicates between files
result, err := detector.FindDuplicatesBetween(tree1, tree2)
```

### Find Matches for Specific Individual

```go
// Get an individual
individual := tree.GetIndividual("@I1@").(*gedcom.IndividualRecord)

// Find potential matches
matches, err := detector.FindMatches(individual, tree)
```

## Configuration

### Default Configuration

```go
config := duplicate.DefaultConfig()
// Returns:
//   MinThreshold: 0.60
//   HighConfidenceThreshold: 0.85
//   ExactMatchThreshold: 0.95
//   NameWeight: 0.40
//   DateWeight: 0.30
//   PlaceWeight: 0.15
//   SexWeight: 0.05
//   RelationshipWeight: 0.10
//   UsePhoneticMatching: true
//   UseRelationshipData: true
//   UseParallelProcessing: true
//   DateTolerance: 2
```

### Custom Configuration

```go
config := &duplicate.DuplicateConfig{
    // Thresholds
    MinThreshold:          0.70,  // Minimum similarity to report
    HighConfidenceThreshold: 0.85,  // High confidence threshold
    ExactMatchThreshold:   0.95,  // Exact match threshold

    // Weights
    NameWeight:        0.40,  // Name similarity weight
    DateWeight:        0.30,  // Date similarity weight
    PlaceWeight:       0.15,  // Place similarity weight
    SexWeight:         0.05,  // Sex match weight
    RelationshipWeight: 0.10,  // Relationship weight

    // Options
    UsePhoneticMatching:    true,  // Enable Soundex matching
    UseRelationshipData:   true,  // Use family relationships
    UseParallelProcessing: true,  // Enable parallel processing
    DateTolerance:         2,     // Years tolerance for dates
    NumWorkers:            0,     // Auto-detect worker count
}

detector := duplicate.NewDuplicateDetector(config)
```

### Configuration Presets

#### Strict Configuration (Merging)

```go
config := &duplicate.DuplicateConfig{
    MinThreshold:          0.90,
    HighConfidenceThreshold: 0.95,
    ExactMatchThreshold:   0.98,
    NameWeight:            0.50,
    DateWeight:            0.30,
    PlaceWeight:           0.15,
    UseRelationshipData:   true,
    DateTolerance:         1,  // Stricter
}
```

#### Loose Configuration (Research)

```go
config := &duplicate.DuplicateConfig{
    MinThreshold:          0.60,
    HighConfidenceThreshold: 0.75,
    ExactMatchThreshold:   0.90,
    NameWeight:            0.35,
    DateWeight:            0.30,
    PlaceWeight:           0.20,
    UseRelationshipData:   false,
    DateTolerance:         5,  // More lenient
}
```

## Similarity Metrics

### Name Similarity (40% weight)

The name similarity algorithm uses multiple strategies:

1. **Exact Match**: Identical normalized names
2. **Normalized Match**: Same name after normalization (removes slashes, case)
3. **Component Match**: Compares given name and surname separately
4. **Phonetic Match**: Soundex algorithm for similar-sounding names
5. **Fuzzy Match**: Levenshtein distance for typos and variations

**Example:**
```go
// These would match with high similarity:
"John /Doe/" vs "John Doe"           // Normalized match
"Smith" vs "Smyth"                   // Phonetic match
"John" vs "Jon"                      // Fuzzy match
```

### Date Similarity (30% weight)

Date comparison handles:
- Exact year matches
- Year differences with tolerance
- Imprecise dates (ABT, BEF, AFT, BETWEEN)
- Date range overlap calculation

**Scoring:**
- Exact match: 1.0
- Within 1 year: 0.9
- Within 2 years: 0.8
- Within 5 years: 0.7
- Within 10 years: 0.5

**Example:**
```go
// These would be semantically equivalent:
"1800" vs "ABT 1800"     // Within tolerance
"BEF 1850" vs "AFT 1840" // Overlapping ranges
```

### Place Similarity (15% weight)

Place comparison supports:
- Exact place matches
- Component matching (city, state, country)
- Hierarchy matching
- Abbreviation handling

**Example:**
```go
// These would match:
"New York" vs "New York, NY"
"New York, NY" vs "New York, New York, USA"
```

### Sex Match (5% weight)

- Match: 1.0
- Mismatch: 0.0 (strong negative indicator)
- Unknown (U): 0.5 (neutral)

### Relationship Similarity (10% weight)

Uses family relationships:
- Common parents: +0.2 bonus
- Common spouse: +0.2 bonus
- Common children: +0.1 per child (max +0.3)

**Note:** Requires `SetTree()` to be called for relationship matching.

## Confidence Levels

The system categorizes matches by confidence:

| Level | Score Range | Meaning | Action |
|-------|-------------|---------|--------|
| **Exact** | 0.95 - 1.0 | Almost certainly the same | Auto-merge candidate |
| **High** | 0.85 - 0.94 | Very likely the same | Manual review recommended |
| **Medium** | 0.70 - 0.84 | Possibly the same | Manual review required |
| **Low** | 0.60 - 0.69 | Unlikely but possible | Review if other indicators |

## API Reference

### Types

#### DuplicateDetector

Main detector struct for finding duplicates.

```go
type DuplicateDetector struct {
    config *DuplicateConfig
    tree   *gedcom.GedcomTree
}
```

#### DuplicateConfig

Configuration for duplicate detection.

```go
type DuplicateConfig struct {
    // Thresholds
    MinThreshold            float64
    HighConfidenceThreshold float64
    ExactMatchThreshold     float64

    // Weights
    NameWeight         float64
    DateWeight         float64
    PlaceWeight        float64
    SexWeight          float64
    RelationshipWeight float64

    // Options
    UsePhoneticMatching   bool
    UseRelationshipData   bool
    UseParallelProcessing bool
    DateTolerance         int
    NumWorkers            int
}
```

#### DuplicateMatch

Represents a potential duplicate match.

```go
type DuplicateMatch struct {
    Individual1      *gedcom.IndividualRecord
    Individual2      *gedcom.IndividualRecord
    SimilarityScore  float64
    Confidence       string
    MatchingFields   []string
    Differences      []string
    NameScore        float64
    DateScore        float64
    PlaceScore       float64
    SexScore         float64
    RelationshipScore float64
}
```

#### DuplicateResult

Contains all duplicate detection results.

```go
type DuplicateResult struct {
    Matches          []DuplicateMatch
    TotalComparisons int
    ProcessingTime   time.Duration
    Metrics          *PerformanceMetrics
}
```

### Methods

#### NewDuplicateDetector

Creates a new duplicate detector.

```go
func NewDuplicateDetector(config *DuplicateConfig) *DuplicateDetector
```

#### FindDuplicates

Finds duplicates within a single GEDCOM tree.

```go
func (dd *DuplicateDetector) FindDuplicates(tree *gedcom.GedcomTree) (*DuplicateResult, error)
```

#### FindDuplicatesBetween

Finds duplicates between two GEDCOM trees.

```go
func (dd *DuplicateDetector) FindDuplicatesBetween(tree1, tree2 *gedcom.GedcomTree) (*DuplicateResult, error)
```

#### FindMatches

Finds potential matches for a specific individual.

```go
func (dd *DuplicateDetector) FindMatches(individual *gedcom.IndividualRecord, tree *gedcom.GedcomTree) ([]DuplicateMatch, error)
```

#### Compare

Compares two individuals directly and returns similarity score.

```go
func (dd *DuplicateDetector) Compare(indi1, indi2 *gedcom.IndividualRecord) (float64, error)
```

#### SetTree

Sets the GEDCOM tree for relationship matching.

```go
func (dd *DuplicateDetector) SetTree(tree *gedcom.GedcomTree)
```

## Examples

### Example 1: Find All Duplicates

```go
detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())
result, err := detector.FindDuplicates(tree)

fmt.Printf("Found %d potential duplicates\n", len(result.Matches))
fmt.Printf("Compared %d pairs in %v\n", 
    result.TotalComparisons, 
    result.ProcessingTime)

for _, match := range result.Matches {
    if match.Confidence == "exact" || match.Confidence == "high" {
        fmt.Printf("High confidence match: %s ↔ %s (%.2f%%)\n",
            match.Individual1.XrefID(),
            match.Individual2.XrefID(),
            match.SimilarityScore*100)
    }
}
```

### Example 2: Filter by Confidence

```go
result, _ := detector.FindDuplicates(tree)

// Get only high-confidence matches
highConfidence := []duplicate.DuplicateMatch{}
for _, match := range result.Matches {
    if match.Confidence == "exact" || match.Confidence == "high" {
        highConfidence = append(highConfidence, match)
    }
}

fmt.Printf("Found %d high-confidence duplicates\n", len(highConfidence))
```

### Example 3: Compare Two Files

```go
tree1, _ := parser.Parse("file1.ged")
tree2, _ := parser.Parse("file2.ged")

detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())
result, _ := detector.FindDuplicatesBetween(tree1, tree2)

for _, match := range result.Matches {
    fmt.Printf("Match: %s (file1) ↔ %s (file2)\n",
        match.Individual1.XrefID(),
        match.Individual2.XrefID())
}
```

### Example 4: Custom Configuration

```go
config := &duplicate.DuplicateConfig{
    MinThreshold:          0.80,  // Only high-confidence matches
    UsePhoneticMatching:   true,
    UseRelationshipData:   true,
    DateTolerance:         1,     // Stricter date matching
    NumWorkers:            8,     // Use 8 workers
}

detector := duplicate.NewDuplicateDetector(config)
result, _ := detector.FindDuplicates(tree)
```

### Example 5: Performance Metrics

```go
result, _ := detector.FindDuplicates(tree)

if result.Metrics != nil {
    fmt.Printf("Processing time: %v\n", result.Metrics.ProcessingTime)
    fmt.Printf("Total comparisons: %d\n", result.Metrics.TotalComparisons)
    fmt.Printf("Throughput: %.2f comparisons/sec\n", result.Metrics.Throughput)
    fmt.Printf("Parallel workers: %d\n", result.Metrics.ParallelWorkers)
}
```

## Performance

### Optimization Features

1. **Pre-filtering**: Indexes by surname, birth year, place
2. **Early Termination**: Skips low-probability matches
3. **Parallel Processing**: Multi-threaded comparison
4. **Memory Pooling**: Reduces allocations

### Expected Performance

| File Size | Processing Time | Comparisons |
|-----------|------------------|-------------|
| 100 individuals | < 1 second | ~5,000 |
| 1,000 individuals | ~5 seconds | ~500,000 |
| 10,000 individuals | ~1 minute | ~50,000,000 |
| 100,000 individuals | ~10 minutes | ~5,000,000,000 |

**Note:** Performance varies based on:
- Number of potential matches
- Pre-filtering effectiveness
- Parallel processing enabled
- System resources

## Best Practices

### 1. Choose Appropriate Thresholds

- **Merging**: Use high threshold (0.85+) to avoid false positives
- **Research**: Use lower threshold (0.60-0.70) to find all possibilities
- **Quality Check**: Use medium threshold (0.70-0.80) for balanced results

### 2. Enable Relationship Matching

Relationship data is a strong indicator:
```go
config.UseRelationshipData = true
detector.SetTree(tree)  // Required for relationship matching
```

### 3. Use Parallel Processing for Large Files

```go
config.UseParallelProcessing = true
config.NumWorkers = 0  // Auto-detect, or set manually
```

### 4. Review High-Confidence Matches First

```go
for _, match := range result.Matches {
    if match.Confidence == "exact" || match.Confidence == "high" {
        // Review these first
    }
}
```

### 5. Check Matching Fields

```go
for _, match := range result.Matches {
    fmt.Printf("Matching: %v\n", match.MatchingFields)
    fmt.Printf("Different: %v\n", match.Differences)
}
```

## Troubleshooting

### No Matches Found

**Possible causes:**
- Threshold too high
- No actual duplicates in file
- Missing data (names, dates)

**Solutions:**
- Lower `MinThreshold`
- Check data quality
- Enable phonetic matching

### Too Many False Positives

**Possible causes:**
- Threshold too low
- Common names (John Smith)
- Missing relationship data

**Solutions:**
- Raise `MinThreshold`
- Increase `NameWeight` or `DateWeight`
- Enable relationship matching

### Slow Performance

**Possible causes:**
- Very large files
- Parallel processing disabled
- Too many comparisons

**Solutions:**
- Enable parallel processing
- Set `MaxComparisons` limit
- Use stricter pre-filtering

## Integration

### With Query API

```go
// Future integration
q.Duplicates().MinThreshold(0.85).Execute()
```

### With CLI

```bash
# Future CLI command
gedcom duplicates family.ged --threshold 0.85
```

## See Also

- [Duplicate Detection Design](DUPLICATE_DETECTION_DESIGN.md) - Detailed design document
- [Query API Documentation](query-api.md) - Graph-based querying
- [Types Documentation](types.md) - Core GEDCOM types
