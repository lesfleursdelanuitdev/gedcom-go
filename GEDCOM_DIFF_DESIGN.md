# GEDCOM Diff Design

**Date:** 2025-01-27  
**Status:** 🎨 Design Phase

---

## Overview

Design a semantic differ that compares two GEDCOM files and identifies differences at a meaningful level, not just line-by-line text comparison. The differ should understand GEDCOM structure and report changes in terms of:
- Added/removed individuals, families, notes, sources
- Modified record data (names, dates, places, events)
- Relationship changes (family connections)
- Structural differences (hierarchical changes)

---

## Use Cases

### 1. Version Comparison

**Scenario:** Compare two versions of the same family tree file.

**Use Cases:**
- Track changes between file versions
- Review edits made by collaborators
- Audit trail for data modifications
- Merge conflict detection

**Example:**
```
File 1: family_v1.ged (100 individuals)
File 2: family_v2.ged (105 individuals)
→ 5 new individuals, 2 removed, 3 modified
```

### 2. Merge Preparation

**Scenario:** Compare two different family trees before merging.

**Use Cases:**
- Identify overlapping individuals (potential duplicates)
- Find conflicting data (same person, different information)
- Discover new relationships
- Plan merge strategy

**Example:**
```
File 1: smith_family.ged
File 2: jones_family.ged
→ 10 overlapping individuals, 50 new individuals, 5 conflicts
```

### 3. Data Quality Review

**Scenario:** Compare original file with cleaned/validated version.

**Use Cases:**
- Verify data corrections
- Review validation fixes
- Confirm data normalization
- Track data improvements

**Example:**
```
Original: raw_data.ged (with errors)
Cleaned: cleaned_data.ged (validated)
→ 15 date corrections, 8 name normalizations, 3 relationship fixes
```

### 4. Import Validation

**Scenario:** Compare imported data with source file.

**Use Cases:**
- Verify import accuracy
- Detect data loss
- Identify transformation issues
- Validate import process

---

## Semantic Comparison Levels

### Level 1: Record-Level Comparison

**What:** Compare at the record level (individuals, families, notes, sources).

**Changes Detected:**
- **Added Records**: Records in file 2 but not in file 1
- **Removed Records**: Records in file 1 but not in file 2
- **Modified Records**: Records with same XREF but different content

**Matching Strategy:**
- Primary: Match by XREF ID (e.g., `@I1@` == `@I1@`)
- Fallback: Match by content similarity (using duplicate detection)

**Example:**
```
Added: @I105@ (John /Smith/)
Removed: @I50@ (Jane /Doe/)
Modified: @I1@
  - Name: "John /Doe/" → "John /Doe Jr/"
  - Birth Date: "1800" → "ABT 1800"
```

### Level 2: Field-Level Comparison

**What:** Compare individual fields within records.

**Changes Detected:**
- **Name Changes**: Given name, surname, full name
- **Date Changes**: Birth, death, marriage dates
- **Place Changes**: Birth place, death place, residence
- **Attribute Changes**: Sex, occupation, religion, etc.
- **Event Changes**: Added/removed/modified events

**Comparison Strategy:**
- Exact match: Same value
- Semantic match: Equivalent values (e.g., "1800" vs "ABT 1800")
- Modified: Different values
- Added: Field exists in file 2 but not file 1
- Removed: Field exists in file 1 but not file 2

**Example:**
```
@I1@ Changes:
  NAME:
    - Old: "John /Doe/"
    - New: "John /Doe Jr/"
  BIRT.DATE:
    - Old: "1800"
    - New: "ABT 1800" (semantically equivalent)
  BIRT.PLAC:
    - Old: "New York"
    - New: "New York, NY, USA" (enhanced)
  DEAT:
    - Added: New death event (DATE: "1870", PLAC: "Boston")
```

### Level 3: Relationship Comparison

**What:** Compare family relationships and connections.

**Changes Detected:**
- **Parent-Child Relationships**: Added/removed children
- **Spouse Relationships**: Added/removed marriages
- **Family Structure**: Family creation/dissolution
- **Relationship Changes**: Modified family connections

**Example:**
```
Family @F1@ Changes:
  Children:
    - Added: @I25@ (new child)
    - Removed: @I10@ (child removed)
  Marriage:
    - DATE: "1850" → "1851" (corrected)
```

### Level 4: Hierarchical Comparison

**What:** Compare nested structures (events, attributes, sources, notes).

**Changes Detected:**
- **Event Modifications**: Date/place changes in events
- **Source Citations**: Added/removed/modified sources
- **Notes**: Added/removed/modified notes
- **Media**: Added/removed/modified media references

**Example:**
```
@I1@ BIRT Event:
  DATE: "1800" → "ABT 1800"
  PLAC: "New York" → "New York, NY"
  SOUR:
    - Added: @S5@ (new source citation)
    - Removed: @S1@ (source removed)
```

---

## Matching Strategies

### Strategy 1: XREF-Based Matching (Primary)

**Approach:** Match records by their XREF IDs.

**Pros:**
- Fast and accurate for same-file versions
- Preserves record identity
- Handles renames and moves

**Cons:**
- Fails when XREFs differ between files
- Doesn't work for cross-file comparison
- Requires fallback for missing XREFs

**Use Case:** Version comparison, same-file edits

### Strategy 2: Content-Based Matching (Fallback)

**Approach:** Use duplicate detection to match records by content.

**Pros:**
- Works across different files
- Handles XREF mismatches
- Finds semantic matches

**Cons:**
- Slower (requires similarity calculation)
- May have false positives
- Requires threshold configuration

**Use Case:** Cross-file comparison, merge preparation

### Strategy 3: Hybrid Matching

**Approach:** Try XREF first, fallback to content matching.

**Pros:**
- Best of both worlds
- Fast for same-file, accurate for cross-file
- Configurable matching strategy

**Cons:**
- More complex implementation
- Requires duplicate detection system

**Use Case:** General-purpose comparison

---

## Difference Types

### 1. Structural Differences

**Added Record:**
- Record exists in file 2 but not file 1
- Example: New individual `@I105@` added

**Removed Record:**
- Record exists in file 1 but not file 2
- Example: Individual `@I50@` deleted

**Moved Record:**
- Record exists in both but XREF changed
- Example: `@I1@` in file 1 → `@I101@` in file 2

### 2. Content Differences

**Modified Field:**
- Field value changed
- Example: Name "John /Doe/" → "John /Doe Jr/"

**Added Field:**
- Field exists in file 2 but not file 1
- Example: Death date added

**Removed Field:**
- Field exists in file 1 but not file 2
- Example: Occupation removed

**Semantically Equivalent:**
- Different values but same meaning
- Example: "1800" vs "ABT 1800" (within tolerance)

### 3. Relationship Differences

**Added Relationship:**
- New family connection
- Example: New child added to family

**Removed Relationship:**
- Family connection removed
- Example: Child removed from family

**Modified Relationship:**
- Relationship data changed
- Example: Marriage date corrected

### 4. Nested Differences

**Event Changes:**
- Events added/removed/modified
- Example: New death event added

**Source Changes:**
- Source citations added/removed/modified
- Example: New source citation added

**Note Changes:**
- Notes added/removed/modified
- Example: Note text updated

---

## Output Format

### 1. Text Summary

**Format:** Human-readable text report.

**Content:**
- Summary statistics
- List of changes by category
- Detailed change descriptions

**Example:**
```
GEDCOM Diff Report
==================

Summary:
  Total Records (File 1): 100
  Total Records (File 2): 105
  Added: 5
  Removed: 2
  Modified: 3
  Unchanged: 95

Added Records:
  @I105@: John /Smith/ (b. 1850)
  @I106@: Jane /Smith/ (b. 1852)
  ...

Removed Records:
  @I50@: Jane /Doe/ (b. 1800)
  ...

Modified Records:
  @I1@: John /Doe/
    NAME: "John /Doe/" → "John /Doe Jr/"
    BIRT.DATE: "1800" → "ABT 1800"
    DEAT: Added (DATE: "1870", PLAC: "Boston")
```

### 2. JSON Format

**Format:** Structured JSON for programmatic processing.

**Structure:**
```json
{
  "summary": {
    "file1": {
      "individuals": 100,
      "families": 50,
      "notes": 20,
      "sources": 15
    },
    "file2": {
      "individuals": 105,
      "families": 52,
      "notes": 22,
      "sources": 16
    },
    "changes": {
      "added": 5,
      "removed": 2,
      "modified": 3,
      "unchanged": 95
    }
  },
  "changes": {
    "added": [
      {
        "xref": "@I105@",
        "type": "INDI",
        "record": { ... }
      }
    ],
    "removed": [ ... ],
    "modified": [
      {
        "xref": "@I1@",
        "type": "INDI",
        "changes": [
          {
            "field": "NAME",
            "old": "John /Doe/",
            "new": "John /Doe Jr/",
            "type": "modified"
          },
          {
            "field": "BIRT.DATE",
            "old": "1800",
            "new": "ABT 1800",
            "type": "semantically_equivalent"
          }
        ]
      }
    ]
  }
}
```

### 3. HTML Diff View

**Format:** Visual HTML diff with color coding.

**Features:**
- Side-by-side comparison
- Color coding (green=added, red=removed, yellow=modified)
- Expandable sections
- Navigation between changes

### 4. Unified Diff Format

**Format:** Git-style unified diff.

**Use Case:** Version control integration, patch generation.

---

## Comparison Algorithms

### Algorithm 1: XREF-Based Comparison

**Steps:**
1. Build index of all records by XREF from file 1
2. Build index of all records by XREF from file 2
3. For each XREF in file 1:
   - If exists in file 2: Compare content → Modified or Unchanged
   - If not in file 2: → Removed
4. For each XREF in file 2:
   - If not in file 1: → Added

**Complexity:** O(n) where n = number of records

**Performance:** Very fast, O(1) lookup per record

### Algorithm 2: Content-Based Comparison

**Steps:**
1. Build index of all records from file 1
2. For each record in file 2:
   - Try XREF match first
   - If no XREF match, use duplicate detection to find best match
   - If similarity > threshold: → Modified
   - If similarity < threshold: → Added (new record)
3. Records in file 1 without matches → Removed

**Complexity:** O(n × m) where n = file 1 records, m = file 2 records

**Performance:** Slower, requires similarity calculations

### Algorithm 3: Hybrid Comparison

**Steps:**
1. Perform XREF-based comparison (fast path)
2. For unmatched records, use content-based matching (slow path)
3. Merge results

**Complexity:** O(n) + O(u × m) where u = unmatched records

**Performance:** Fast for same-file, accurate for cross-file

---

## Field Comparison Rules

### Name Comparison

**Exact Match:**
- Same normalized name string
- Example: "John /Doe/" == "John /Doe/"

**Semantic Match:**
- Same name components (given + surname)
- Example: "John /Doe/" == "John Doe"

**Modified:**
- Different name components
- Example: "John /Doe/" → "John /Doe Jr/"

### Date Comparison

**Exact Match:**
- Same date string
- Example: "1800" == "1800"

**Semantic Match:**
- Dates within tolerance
- Example: "1800" ≈ "ABT 1800" (within 2 years)

**Modified:**
- Dates differ beyond tolerance
- Example: "1800" → "1850"

### Place Comparison

**Exact Match:**
- Same place string
- Example: "New York" == "New York"

**Semantic Match:**
- Same place components
- Example: "New York" ≈ "New York, NY" (hierarchy match)

**Modified:**
- Different places
- Example: "New York" → "Boston"

### Relationship Comparison

**Match:**
- Same family XREFs
- Example: FAMC: "@F1@" == "@F1@"

**Modified:**
- Different family XREFs
- Example: FAMC: "@F1@" → "@F2@"

**Added/Removed:**
- Family XREF exists in one but not the other

---

## Configuration Options

### Comparison Mode

**Strict Mode:**
- XREF-based matching only
- Fast, for same-file versions
- No content matching

**Loose Mode:**
- Content-based matching
- Slower, for cross-file comparison
- Uses duplicate detection

**Hybrid Mode (Default):**
- XREF first, content fallback
- Balanced performance and accuracy

### Matching Thresholds

**Similarity Threshold:**
- Minimum similarity for content matching
- Default: 0.85 (high confidence)
- Configurable per use case

**Date Tolerance:**
- Years tolerance for date comparison
- Default: 2 years
- Example: "1800" ≈ "1802" if tolerance = 2

**Place Matching:**
- Level of place hierarchy to match
- Options: exact, city, state, country
- Default: city level

### Output Options

**Detail Level:**
- Summary only
- Field-level details
- Full nested comparison

**Format:**
- Text
- JSON
- HTML
- Unified diff

**Filtering:**
- Show only added
- Show only removed
- Show only modified
- Show all changes

---

## Performance Considerations

### Optimization Strategies

1. **Indexing:**
   - Build indexes by XREF for O(1) lookup
   - Index by content for content matching

2. **Early Termination:**
   - Skip comparison if records are identical
   - Use hash comparison for quick equality check

3. **Parallel Processing:**
   - Compare records in parallel
   - Use worker pools for large files

4. **Caching:**
   - Cache parsed dates/places
   - Cache similarity scores

5. **Incremental Comparison:**
   - Compare only changed sections
   - Skip unchanged records

### Expected Performance

| File Size | Comparison Time (XREF) | Comparison Time (Content) |
|-----------|------------------------|---------------------------|
| 100 records | < 1 second | ~5 seconds |
| 1,000 records | ~1 second | ~1 minute |
| 10,000 records | ~5 seconds | ~10 minutes |
| 100,000 records | ~30 seconds | ~2 hours |

**Note:** Content-based comparison is much slower due to similarity calculations.

---

## API Design

### Core Interface

```go
type GedcomDiffer struct {
    config *DiffConfig
}

type DiffConfig struct {
    // Matching strategy
    MatchingStrategy string // "xref", "content", "hybrid"
    
    // Thresholds
    SimilarityThreshold float64 // For content matching (default: 0.85)
    DateTolerance       int     // Years tolerance (default: 2)
    
    // Options
    IncludeUnchanged    bool    // Include unchanged records in output
    DetailLevel        string  // "summary", "field", "full"
    OutputFormat        string  // "text", "json", "html", "unified"
}

type DiffResult struct {
    Summary    DiffSummary
    Changes    DiffChanges
    Statistics DiffStatistics
}

type DiffSummary struct {
    File1Stats RecordStats
    File2Stats RecordStats
    Changes    ChangeCounts
}

type DiffChanges struct {
    Added    []RecordDiff
    Removed  []RecordDiff
    Modified []RecordModification
}

type RecordModification struct {
    Xref     string
    Type     string
    Changes  []FieldChange
}

type FieldChange struct {
    Field    string
    OldValue string
    NewValue string
    Type     string // "modified", "added", "removed", "semantically_equivalent"
}
```

### Methods

```go
// Compare two GEDCOM files
func (gd *GedcomDiffer) Compare(tree1, tree2 *GedcomTree) (*DiffResult, error)

// Compare two files from disk
func (gd *GedcomDiffer) CompareFiles(file1, file2 string) (*DiffResult, error)

// Generate diff report
func (gd *GedcomDiffer) GenerateReport(result *DiffResult) (string, error)

// Export diff to file
func (gd *GedcomDiffer) ExportDiff(result *DiffResult, format string, outputPath string) error
```

### Query API Integration

```go
// Add diff to Query API
q.Diff(file1, file2).Strategy("hybrid").Execute()

// Compare specific records
q.Compare("@I1@", "@I101@").Execute()
```

---

## Edge Cases and Challenges

### 1. XREF Mismatches

**Challenge:** Same person, different XREF IDs.

**Solution:**
- Use content-based matching as fallback
- Report as "moved" or "renamed" record

### 2. Semantic Equivalence

**Challenge:** Different values but same meaning.

**Example:**
- "1800" vs "ABT 1800"
- "New York" vs "New York, NY"

**Solution:**
- Use date/place parsing and comparison
- Mark as "semantically_equivalent" not "modified"

### 3. Relationship Changes

**Challenge:** Family structure changes.

**Example:**
- Child moved from one family to another
- Spouse relationship added/removed

**Solution:**
- Track relationship changes separately
- Report family-level modifications

### 4. Nested Structure Changes

**Challenge:** Events, sources, notes within records.

**Example:**
- Event date changed within same event
- New source citation added

**Solution:**
- Recursive comparison of nested structures
- Track changes at all hierarchy levels

### 5. Large File Performance

**Challenge:** Comparing very large files (100,000+ records).

**Solution:**
- Use parallel processing
- Implement incremental comparison
- Provide progress reporting
- Allow comparison limits

### 6. False Positives (Content Matching)

**Challenge:** Different people with similar data.

**Example:**
- Two "John Smith" born in 1800

**Solution:**
- Use high similarity threshold
- Require multiple matching fields
- Use relationship data to distinguish

---

## Implementation Phases

### Phase 1: Basic Diff (MVP)

**Scope:**
- XREF-based matching only
- Record-level comparison (added/removed/modified)
- Text output format
- Basic field comparison

**Deliverables:**
- `GedcomDiffer` struct
- XREF-based comparison algorithm
- Text report generation
- Basic tests

### Phase 2: Field-Level Diff

**Scope:**
- Field-level comparison
- Nested structure comparison (events, sources, notes)
- Semantic equivalence detection
- JSON output format

**Deliverables:**
- Field-level diff algorithm
- Semantic comparison (dates, places)
- JSON export
- Enhanced tests

### Phase 3: Content-Based Matching

**Scope:**
- Content-based matching using duplicate detection
- Hybrid matching strategy
- Relationship comparison
- HTML output format

**Deliverables:**
- Content matching integration
- Relationship diff
- HTML report generation
- Performance optimizations

### Phase 4: Advanced Features

**Scope:**
- Unified diff format
- Parallel processing
- Incremental comparison
- CLI integration
- Merge conflict detection

**Deliverables:**
- Unified diff generation
- Performance optimizations
- CLI command
- Merge conflict reporting

---

## Integration Points

### 1. Duplicate Detection Integration

**Use:** Leverage existing duplicate detection for content matching.

**Benefits:**
- Reuse similarity algorithms
- Consistent matching logic
- Proven accuracy

**Integration:**
```go
// Use duplicate detector for content matching
detector := duplicate.NewDuplicateDetector(config)
matches := detector.FindDuplicatesBetween(tree1, tree2)
```

### 2. Query API Integration

**Use:** Add diff capabilities to query API.

**Example:**
```go
q.Diff(tree1, tree2).Strategy("hybrid").Execute()
```

### 3. CLI Integration

**Use:** Add `diff` command to CLI.

**Example:**
```bash
gedcom diff file1.ged file2.ged --format json -o diff.json
```

### 4. Validator Integration

**Use:** Compare validated file with original.

**Example:**
```go
diff := validator.CompareWithOriginal(validatedTree, originalTree)
```

---

## Output Examples

### Text Summary Example

```
GEDCOM Diff Report: family_v1.ged vs family_v2.ged
====================================================

Summary:
  File 1: 100 individuals, 50 families, 20 notes, 15 sources
  File 2: 105 individuals, 52 families, 22 notes, 16 sources
  
  Changes:
    Added:     5 individuals, 2 families, 2 notes, 1 source
    Removed:   2 individuals, 0 families, 0 notes, 0 sources
    Modified:  3 individuals, 0 families, 0 notes, 0 sources
    Unchanged: 95 individuals, 50 families, 20 notes, 15 sources

Added Individuals:
  @I105@: John /Smith/ (b. 1850, New York)
  @I106@: Jane /Smith/ (b. 1852, New York)
  ...

Removed Individuals:
  @I50@: Jane /Doe/ (b. 1800, Boston)
  @I51@: Robert /Doe/ (b. 1802, Boston)

Modified Individuals:
  @I1@: John /Doe/
    NAME: "John /Doe/" → "John /Doe Jr/"
    BIRT.DATE: "1800" → "ABT 1800" (semantically equivalent)
    BIRT.PLAC: "New York" → "New York, NY, USA" (enhanced)
    DEAT: Added
      DATE: "1870"
      PLAC: "Boston"
    NOTE: Added @N5@
  
  @I2@: Jane /Smith/
    BIRT.DATE: "1805" → "1806" (modified)
    OCCU: "Teacher" → "Professor" (modified)
```

### JSON Example

```json
{
  "summary": {
    "file1": {
      "individuals": 100,
      "families": 50,
      "notes": 20,
      "sources": 15
    },
    "file2": {
      "individuals": 105,
      "families": 52,
      "notes": 22,
      "sources": 16
    },
    "changes": {
      "added": {
        "individuals": 5,
        "families": 2,
        "notes": 2,
        "sources": 1
      },
      "removed": {
        "individuals": 2,
        "families": 0,
        "notes": 0,
        "sources": 0
      },
      "modified": {
        "individuals": 3,
        "families": 0,
        "notes": 0,
        "sources": 0
      }
    }
  },
  "changes": {
    "added": [
      {
        "xref": "@I105@",
        "type": "INDI",
        "name": "John /Smith/",
        "birth_date": "1850",
        "birth_place": "New York"
      }
    ],
    "removed": [
      {
        "xref": "@I50@",
        "type": "INDI",
        "name": "Jane /Doe/",
        "birth_date": "1800",
        "birth_place": "Boston"
      }
    ],
    "modified": [
      {
        "xref": "@I1@",
        "type": "INDI",
        "changes": [
          {
            "field": "NAME",
            "path": "NAME",
            "old": "John /Doe/",
            "new": "John /Doe Jr/",
            "type": "modified"
          },
          {
            "field": "BIRT.DATE",
            "path": "BIRT.DATE",
            "old": "1800",
            "new": "ABT 1800",
            "type": "semantically_equivalent",
            "reason": "Dates within tolerance (2 years)"
          },
          {
            "field": "DEAT",
            "path": "DEAT",
            "old": null,
            "new": {
              "DATE": "1870",
              "PLAC": "Boston"
            },
            "type": "added"
          }
        ]
      }
    ]
  }
}
```

---

## Recommendations

### 1. Start with XREF-Based Comparison

**Rationale:**
- Simplest to implement
- Fastest performance
- Covers most use cases (version comparison)
- Can add content matching later

### 2. Integrate with Duplicate Detection

**Rationale:**
- Reuse existing similarity algorithms
- Consistent matching logic
- Proven accuracy
- Reduces code duplication

### 3. Support Multiple Output Formats

**Rationale:**
- Text for human reading
- JSON for programmatic processing
- HTML for visual comparison
- Unified diff for version control

### 4. Make It Configurable

**Rationale:**
- Different use cases need different strategies
- Performance vs accuracy trade-offs
- User should control matching behavior

### 5. Focus on Semantic Differences

**Rationale:**
- More useful than text diff
- Understands GEDCOM structure
- Reports meaningful changes
- Handles semantic equivalence

---

## Open Questions

1. **Should we support three-way merge?**
   - Compare base + two versions
   - Identify conflicts
   - Generate merge suggestions

2. **How to handle XREF conflicts?**
   - Same XREF, different people
   - Different XREF, same person
   - Resolution strategy?

3. **Should we track change history?**
   - Who made the change?
   - When was it made?
   - Why was it changed?

4. **How to handle large files?**
   - Streaming comparison?
   - Incremental diff?
   - Sampling for preview?

5. **Should we support partial comparison?**
   - Compare only specific individuals
   - Compare only specific families
   - Compare only specific fields

---

## Summary

**Recommended Approach:**
1. **Start with XREF-based comparison** (Phase 1)
2. **Add field-level comparison** (Phase 2)
3. **Integrate content-based matching** (Phase 3)
4. **Add advanced features** (Phase 4)

**Key Features:**
- Semantic understanding (not just text diff)
- Multiple matching strategies (XREF, content, hybrid)
- Multiple output formats (text, JSON, HTML, unified)
- Configurable comparison behavior
- Performance optimizations

**Priority:**
1. XREF-based comparison (fast, covers most cases)
2. Field-level details (useful for review)
3. Content matching (for cross-file comparison)
4. Advanced features (nice to have)

---

**Next Steps:**
1. Review and refine design
2. Implement Phase 1 (XREF-based comparison)
3. Test with real GEDCOM files
4. Iterate based on results
