# Performance Testing Plan for Parser and Query Operations

## Overview

This plan outlines comprehensive performance testing for:
1. **Parser Performance**: Measuring parse and validation time for test data files
2. **Query Performance**: Measuring execution time for typical genealogical queries

## Test Data Files

Available test data files in `/testdata`:
- `xavier.ged` - 101KB, 5,822 lines (small dataset)
- `gracis.ged` - 163KB, 10,324 lines (medium dataset)
- `tree1.ged` - 211KB, 12,714 lines (medium dataset)
- `royal92.ged` - 488KB, 30,683 lines (large dataset)
- `pres2020.ged` - 1.1MB, ~large (very large dataset)

## Part 1: Parser Performance Testing

### 1.1 Objectives
- Measure parsing time for each test data file
- Measure validation time (basic and advanced)
- Track memory usage during parsing
- Calculate throughput (individuals/second, families/second)
- Compare sequential vs parallel parsing (if applicable)

### 1.2 Test Structure

#### Test: `TestParserPerformance_AllTestDataFiles`
For each test data file:
1. **Parse Phase**:
   - Measure time to parse the file
   - Record: file size, number of individuals, number of families
   - Calculate: parsing throughput (individuals/sec, families/sec)
   - Track: memory usage before/after parsing

2. **Basic Validation Phase**:
   - Measure time to run basic validation (GedcomValidator)
   - Record: number of errors/warnings found
   - Track: memory usage

3. **Advanced Validation Phase** (optional):
   - Measure time to run advanced validation (with date consistency checks)
   - Record: number of advanced validation errors/warnings

4. **Graph Construction Phase** (for query testing):
   - Measure time to build graph from parsed tree
   - Record: number of nodes, number of edges
   - Track: memory usage

### 1.3 Metrics to Collect

For each test data file:
- **File Metrics**:
  - File size (bytes)
  - Number of lines
  - Number of individuals
  - Number of families
  - Number of other records (notes, sources, etc.)

- **Performance Metrics**:
  - Parse duration (ms)
  - Basic validation duration (ms)
  - Advanced validation duration (ms) [optional]
  - Graph construction duration (ms)
  - Total duration (ms)

- **Throughput Metrics**:
  - Individuals parsed per second
  - Families parsed per second
  - Records validated per second

- **Memory Metrics**:
  - Memory before parsing (MB)
  - Memory after parsing (MB)
  - Memory after validation (MB)
  - Memory after graph construction (MB)
  - Peak memory usage (MB)

- **Quality Metrics**:
  - Number of parsing errors
  - Number of validation errors (by severity)
  - Number of validation warnings

### 1.4 Expected Output Format

```
Parser Performance Results
==========================

xavier.ged (101KB, 5,822 lines)
  Parse:            XX.XXXms (XXX individuals/sec, XXX families/sec)
  Basic Validation: XX.XXXms (XXX records/sec)
  Graph Build:      XX.XXXms
  Total:            XX.XXXms
  Memory:           XXX MB (peak: XXX MB)
  Errors:           X severe, X warnings

[... similar for other files ...]
```

## Part 2: Query Performance Testing

### 2.1 Objectives
- Measure query execution time for typical genealogical queries
- Test queries on different dataset sizes
- Track query performance with and without caching
- Measure memory impact of queries

### 2.2 Typical Queries to Test

#### 2.2.1 Relationship Queries
1. **Parents Query**
   - Query: Get parents of a specific individual
   - Test cases:
     - Individual with both parents
     - Individual with one parent
     - Individual with no parents (root ancestor)
   - Expected: < 1ms for most cases

2. **Children Query**
   - Query: Get children of a specific individual
   - Test cases:
     - Individual with many children (5+)
     - Individual with few children (1-2)
     - Individual with no children
   - Expected: < 1ms for most cases

3. **Siblings Query**
   - Query: Get siblings of a specific individual
   - Test cases:
     - Individual with many siblings
     - Individual with no siblings (only child)
   - Expected: < 1ms for most cases

4. **Spouses Query**
   - Query: Get spouses of a specific individual
   - Test cases:
     - Individual with multiple marriages
     - Individual with one spouse
     - Individual with no spouse
   - Expected: < 1ms for most cases

#### 2.2.2 Ancestral Queries
5. **Ancestors Query**
   - Query: Get all ancestors of a specific individual
   - Test cases:
     - Deep ancestry (10+ generations)
     - Shallow ancestry (2-3 generations)
     - Individual with no ancestors
   - Expected: < 10ms for deep ancestry, < 1ms for shallow

6. **Descendants Query**
   - Query: Get all descendants of a specific individual
   - Test cases:
     - Large descendant tree (100+ descendants)
     - Small descendant tree (10-20 descendants)
     - Individual with no descendants
   - Expected: < 50ms for large trees, < 5ms for small

#### 2.2.3 Family Queries
7. **Families for Individual**
   - Query: Get all families a person belongs to (as child or spouse)
   - Test cases:
     - Person in multiple families (multiple marriages)
     - Person in one family
     - Person in no families
   - Expected: < 1ms for most cases

8. **Family Members Query**
   - Query: Get all members of a family (husband, wife, children)
   - Test cases:
     - Large family (10+ children)
     - Small family (1-2 children)
     - Family with one parent
   - Expected: < 1ms for most cases

9. **Family with Given Parent**
   - Query: Find families where a specific individual is a parent
   - Test cases:
     - Parent in multiple families
     - Parent in one family
     - Individual who is not a parent
   - Expected: < 1ms for most cases

#### 2.2.4 Path Finding Queries
10. **Shortest Path Query**
    - Query: Find shortest path between two individuals
    - Test cases:
      - Close relatives (siblings, parent-child)
      - Distant relatives (cousins, etc.)
      - Unrelated individuals
    - Expected: < 10ms for most cases

11. **All Paths Query**
    - Query: Find all paths between two individuals
    - Test cases:
      - Multiple paths (through different ancestors)
      - Single path
      - No path (unrelated)
    - Expected: < 50ms for most cases

#### 2.2.5 Filter Queries
12. **Filter by Name**
    - Query: Find individuals with a specific name pattern
    - Test cases:
      - Common name (many matches)
      - Rare name (few matches)
      - No matches
    - Expected: < 10ms for most cases

13. **Filter by Birth Date**
    - Query: Find individuals born in a specific year/range
    - Test cases:
      - Narrow range (1 year)
      - Wide range (50 years)
      - No matches
    - Expected: < 10ms for most cases

### 2.3 Test Structure

#### Test: `TestQueryPerformance_AllTestDataFiles`
For each test data file:
1. **Setup Phase**:
   - Parse the file
   - Build the graph
   - Select representative individuals for testing:
     - Individual with deep ancestry
     - Individual with many descendants
     - Individual with multiple families
     - Individual with many siblings
     - Individual with multiple spouses
     - Root ancestor (no parents)
     - Leaf individual (no descendants)

2. **Query Execution Phase**:
   For each query type:
   - Execute query multiple times (e.g., 10 iterations)
   - Measure: min, max, average, median execution time
   - Track: memory usage
   - Record: result count

3. **Cache Performance Phase**:
   - Execute same queries again (should use cache)
   - Compare: cached vs uncached performance
   - Calculate: cache speedup factor

### 2.4 Metrics to Collect

For each query type:
- **Performance Metrics**:
  - First execution time (cold cache) (ms)
  - Average execution time (ms)
  - Min execution time (ms)
  - Max execution time (ms)
  - Median execution time (ms)
  - Cached execution time (ms)
  - Cache speedup factor (X times faster)

- **Result Metrics**:
  - Number of results returned
  - Result size (if applicable)

- **Memory Metrics**:
  - Memory before query (MB)
  - Memory after query (MB)
  - Memory delta (MB)

### 2.5 Expected Output Format

```
Query Performance Results - royal92.ged
========================================

Relationship Queries:
  Parents:           X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
  Children:          X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
  Siblings:          X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
  Spouses:           X.XXXms (avg), X.XXXms (cached) - X.Xx speedup

Ancestral Queries:
  Ancestors:         X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
  Descendants:       X.XXXms (avg), X.XXXms (cached) - X.Xx speedup

Family Queries:
  Families for Individual: X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
  Family Members:          X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
  Family with Given Parent: X.XXXms (avg), X.XXXms (cached) - X.Xx speedup

Path Finding:
  Shortest Path:     X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
  All Paths:         X.XXXms (avg), X.XXXms (cached) - X.Xx speedup

Filter Queries:
  Filter by Name:    X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
  Filter by Birth Date: X.XXXms (avg), X.XXXms (cached) - X.Xx speedup
```

## Part 3: Implementation Structure

### 3.1 Test File Organization

Create a new test file: `parser/performance_testdata_test.go`
- Contains parser performance tests using real testdata files
- Uses existing `parser/performance_test.go` as reference for structure

Create a new test file: `query/performance_testdata_test.go`
- Contains query performance tests using real testdata files
- Uses existing `query/performance_test.go` as reference for structure

### 3.2 Helper Functions Needed

1. **Parser Performance Helpers**:
   - `measureParsePerformance(filename string) ParseMetrics`
   - `measureValidationPerformance(tree *GedcomTree) ValidationMetrics`
   - `measureGraphConstructionPerformance(tree *GedcomTree) GraphMetrics`

2. **Query Performance Helpers**:
   - `selectTestIndividuals(graph *Graph) TestIndividuals`
   - `measureQueryPerformance(queryFunc func(), iterations int) QueryMetrics`
   - `measureCachedQueryPerformance(queryFunc func(), iterations int) QueryMetrics`

3. **Reporting Helpers**:
   - `printParserPerformanceReport(results []ParseMetrics)`
   - `printQueryPerformanceReport(results []QueryMetrics)`
   - `exportPerformanceReportToJSON(results interface{})` [optional]

### 3.3 Test Execution Strategy

1. **Individual Test Execution**:
   - Each test data file can be tested independently
   - Use Go's `-run` flag to test specific files: `go test -run TestParserPerformance_xavier`

2. **Batch Execution**:
   - Run all tests: `go test -run TestParserPerformance -v`
   - Run all query tests: `go test -run TestQueryPerformance -v`

3. **Benchmark Mode**:
   - Use Go's benchmark framework for statistical accuracy
   - Run: `go test -bench=BenchmarkParserPerformance -benchmem`

### 3.4 Test Data Selection Strategy

For query performance testing, select individuals that represent:
- **Edge Cases**:
  - Root ancestors (no parents)
  - Leaf individuals (no descendants)
  - Individuals with no families
  - Individuals with many relationships

- **Typical Cases**:
  - Individuals with 2-3 children
  - Individuals with 1-2 siblings
  - Individuals with one spouse
  - Individuals with moderate ancestry (3-5 generations)

- **Stress Cases**:
  - Individuals with deep ancestry (10+ generations)
  - Individuals with many descendants (50+)
  - Individuals with multiple marriages (3+)
  - Large families (10+ children)

## Part 4: Success Criteria

### 4.1 Parser Performance Targets

Based on existing benchmarks:
- **Small files** (< 200KB): < 100ms parse time
- **Medium files** (200KB - 500KB): < 500ms parse time
- **Large files** (> 500KB): < 2s parse time
- **Throughput**: > 50,000 individuals/second
- **Memory**: < 20MB per 1,000 individuals

### 4.2 Query Performance Targets

Based on existing benchmarks:
- **Simple queries** (parents, children, siblings, spouses): < 1ms
- **Ancestral queries** (ancestors, descendants): < 10ms for typical cases
- **Path finding**: < 10ms for shortest path, < 50ms for all paths
- **Filter queries**: < 10ms for indexed filters
- **Cache speedup**: > 10x for repeated queries

## Part 5: Reporting and Analysis

### 5.1 Output Formats

1. **Console Output**: Human-readable summary during test execution
2. **JSON Output**: Machine-readable detailed results (optional)
3. **Markdown Report**: Formatted report for documentation (optional)

### 5.2 Analysis Points

1. **Scaling Analysis**:
   - How does performance scale with file size?
   - Are there performance bottlenecks at certain sizes?
   - Memory usage trends

2. **Query Pattern Analysis**:
   - Which queries are fastest?
   - Which queries benefit most from caching?
   - Are there query patterns that need optimization?

3. **Comparison Analysis**:
   - Compare performance across different test data files
   - Identify files with unusual performance characteristics
   - Document any anomalies

## Part 6: Future Enhancements

1. **Continuous Performance Monitoring**:
   - Integrate with CI/CD to track performance over time
   - Alert on performance regressions

2. **Performance Profiling**:
   - Use Go's `pprof` for detailed profiling
   - Identify hot paths and optimization opportunities

3. **Comparative Analysis**:
   - Compare with other GEDCOM libraries (if available)
   - Document performance advantages

4. **Real-World Dataset Testing**:
   - Test with actual user datasets (with permission)
   - Validate performance claims with real data

## Implementation Notes

- Use Go's `testing` package for test structure
- Use `time` package for precise timing
- Use `runtime` package for memory measurements
- Consider using `testing.B` for benchmark-style tests
- Ensure tests are deterministic and repeatable
- Document any assumptions or limitations
- Include error handling for file I/O operations
- Make tests skip gracefully if testdata files are missing

