# GEDCOM Go API Examples

Comprehensive examples demonstrating how to use the GEDCOM Go library.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Parsing GEDCOM Files](#parsing-gedcom-files)
- [Building Graphs](#building-graphs)
- [Query Examples](#query-examples)
- [Error Handling](#error-handling)
- [Metrics Collection](#metrics-collection)
- [Configuration](#configuration)
- [Advanced Examples](#advanced-examples)

---

## Basic Usage

### Parse and Query a GEDCOM File

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/query"
)

func main() {
    // Parse GEDCOM file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        log.Fatal(err)
    }

    // Create query builder
    q, err := query.NewQuery(tree)
    if err != nil {
        log.Fatal(err)
    }

    // Find all individuals named "John"
    results, err := q.Filter().ByName("John").Execute()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d individuals named John\n", len(results))
    for _, indi := range results {
        fmt.Printf("  - %s\n", indi.GetName())
    }
}
```

---

## Parsing GEDCOM Files

### Basic Parsing

```go
p := parser.NewHierarchicalParser()
tree, err := p.Parse("family.ged")
if err != nil {
    log.Fatal(err)
}

// Access parsed data
individuals := tree.GetAllIndividuals()
families := tree.GetAllFamilies()

fmt.Printf("Parsed %d individuals and %d families\n", 
    len(individuals), len(families))
```

### Parsing with Error Collection

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
        fmt.Printf("  [%s] %s (line %d)\n", 
            err.Severity, err.Message, err.LineNumber)
    }
}
```

---

## Building Graphs

### Basic Graph Building

```go
// Build graph from parsed tree
graph, err := query.BuildGraph(tree, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Graph built: %d nodes, %d edges\n", 
    graph.NodeCount(), graph.EdgeCount())
```

### Graph Building with Configuration

```go
// Create custom configuration
config := query.DefaultConfig()
config.Cache.QueryCacheSize = 2000
config.Timeout.QueryTimeout = 2 * time.Minute

// Build graph with configuration
graph, err := query.BuildGraph(tree, config)
if err != nil {
    log.Fatal(err)
}
```

### Hybrid Storage (Large Datasets)

```go
// Build graph with hybrid storage for large datasets
sqlitePath := "indexes.db"
badgerPath := "graph_data"
config := query.DefaultConfig()

graph, err := query.BuildGraphHybrid(tree, sqlitePath, badgerPath, config)
if err != nil {
    log.Fatal(err)
}
defer graph.Close()

// Graph is now stored in databases, loaded on-demand
```

---

## Query Examples

### Filter Queries

#### Find by Name

```go
// Exact name match
results, err := q.Filter().ByNameExact("John Smith").Execute()

// Partial name match
results, err := q.Filter().ByName("John").Execute()

// Name starts with
results, err := q.Filter().ByNameStarts("John").Execute()
```

#### Find by Date Range

```go
start := time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(1900, 12, 31, 23, 59, 59, 0, time.UTC)

results, err := q.Filter().
    ByBirthDate(start, end).
    Execute()
```

#### Find by Year

```go
results, err := q.Filter().ByBirthYear(1850).Execute()
```

#### Combined Filters

```go
results, err := q.Filter().
    ByName("John").
    BySex("M").
    ByBirthYear(1850).
    HasChildren().
    Execute()
```

### Ancestor Queries

```go
// Find all ancestors
ancestors, err := q.Individual("@I1@").Ancestors().Execute()

// Limit generations
ancestors, err := q.Individual("@I1@").
    Ancestors().
    MaxGenerations(5).
    Execute()

// Include self
ancestors, err := q.Individual("@I1@").
    Ancestors().
    IncludeSelf().
    Execute()

// With filter
ancestors, err := q.Individual("@I1@").
    Ancestors().
    Filter(func(indi *gedcom.IndividualRecord) bool {
        return indi.GetSex() == "M"
    }).
    Execute()
```

### Descendant Queries

```go
// Find all descendants
descendants, err := q.Individual("@I1@").Descendants().Execute()

// Limit generations
descendants, err := q.Individual("@I1@").
    Descendants().
    MaxGenerations(3).
    Execute()
```

### Relationship Queries

```go
// Calculate relationship
result, err := q.Individual("@I1@").RelationshipTo("@I2@").Execute()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Relationship: %s\n", result.RelationshipType)
fmt.Printf("Degree: %d\n", result.Degree)
fmt.Printf("Is Direct: %v\n", result.IsDirect)
```

### Path Queries

```go
// Shortest path
path, err := q.Individual("@I1@").PathTo("@I2@").Shortest()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Path length: %d\n", path.Length)

// All paths
paths, err := q.Individual("@I1@").
    PathTo("@I2@").
    MaxLength(10).
    All()
```

### Family Queries

```go
// Get family members
husband, err := q.Family("@F1@").Husband()
wife, err := q.Family("@F1@").Wife()
children, err := q.Family("@F1@").Children()

// Get marriage date
marriageDate, err := q.Family("@F1@").MarriageDate()
```

---

## Error Handling

### Standard Error Handling

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"

// Create standardized error
err := gedcom.NewStandardError(
    gedcom.ErrorTypeQuery,
    gedcom.SeverityWarning,
    "Individual not found",
)

// With context
err := gedcom.NewStandardErrorWithContext(
    gedcom.ErrorTypeQuery,
    gedcom.SeverityWarning,
    "Individual not found",
    "Query.Individual",
)

// Wrap existing error
err := gedcom.WrapError(
    gedcom.ErrorTypeStorage,
    gedcom.SeveritySevere,
    originalErr,
    "HybridStorage.LoadNode",
)

// Check error type
if gedcom.IsQueryError(err) {
    // Handle query error
}

if gedcom.IsStorageError(err) {
    // Handle storage error
}

// Get error severity
severity := gedcom.GetErrorSeverity(err)
```

### Error Manager Usage

```go
errorManager := gedcom.NewErrorManager()

// Add errors
errorManager.AddError(
    gedcom.SeverityWarning,
    "Missing birth date",
    123,
    "Individual Validation",
)

// Check for errors
if errorManager.HasErrors() {
    errors := errorManager.Errors()
    for _, err := range errors {
        fmt.Printf("[%s] %s\n", err.Severity, err.Message)
    }
}

// Get error summary
summary := errorManager.GetErrorSummary()
fmt.Printf("Severe: %d, Warning: %d\n", 
    summary[gedcom.SeveritySevere],
    summary[gedcom.SeverityWarning])
```

---

## Metrics Collection

### Using Metrics

```go
// Get metrics from graph
metrics := graph.GetMetrics()

// Record query execution
start := time.Now()
results, err := q.Filter().ByName("John").Execute()
duration := time.Since(start)
metrics.RecordQuery(duration)

// Record cache operations
metrics.RecordCacheHit()
// or
metrics.RecordCacheMiss()

// Get snapshot
snapshot := metrics.GetSnapshot()
fmt.Printf("Query count: %d\n", snapshot.QueryCount)
fmt.Printf("Average query time: %v\n", snapshot.QueryAvgTime)
fmt.Printf("Cache hit rate: %.2f%%\n", snapshot.CacheHitRate)
```

### Metrics in Queries

```go
// Metrics are automatically collected during query execution
q, err := query.NewQuery(tree)
if err != nil {
    log.Fatal(err)
}

// Execute queries (metrics collected automatically)
results, err := q.Filter().ByName("John").Execute()

// Get metrics snapshot
graph := q.Graph()
metrics := graph.GetMetrics()
snapshot := metrics.GetSnapshot()

fmt.Printf("Total queries: %d\n", snapshot.QueryCount)
fmt.Printf("Cache hits: %d, misses: %d\n", 
    snapshot.CacheHits, snapshot.CacheMisses)
```

---

## Configuration

### Basic Configuration

```go
config := query.DefaultConfig()

// Adjust cache sizes
config.Cache.QueryCacheSize = 2000
config.Cache.HybridNodeCacheSize = 100000

// Adjust timeouts
config.Timeout.QueryTimeout = 2 * time.Minute
config.Timeout.BuildTimeout = 10 * time.Minute

// Adjust database settings
config.Database.SQLiteMaxOpenConns = 20
```

### Loading Configuration from File

```go
// Load from file
config, err := query.LoadConfig("config.json")
if err != nil {
    log.Fatal(err)
}

// Use configuration
graph, err := query.BuildGraph(tree, config)
```

### Saving Configuration

```go
config := query.DefaultConfig()
config.Cache.QueryCacheSize = 2000

// Save to file
err := query.SaveConfig(config, "my-config.json")
if err != nil {
    log.Fatal(err)
}
```

---

## Advanced Examples

### Custom Filter Function

```go
// Custom filter for individuals born in specific century
results, err := q.Filter().Where(func(indi *gedcom.IndividualRecord) bool {
    birthDate := indi.GetBirthDate()
    // Parse and check century
    // ... custom logic ...
    return true
}).Execute()
```

### Batch Processing

```go
// Process large dataset in batches
allIndividuals := tree.GetAllIndividuals()
batchSize := 1000

for i := 0; i < len(allIndividuals); i += batchSize {
    // Process batch
    // ...
}
```

### Concurrent Queries

```go
var wg sync.WaitGroup
resultsChan := make(chan []*gedcom.IndividualRecord, 3)

// Run multiple queries concurrently
wg.Add(3)

go func() {
    defer wg.Done()
    results, _ := q.Filter().ByName("John").Execute()
    resultsChan <- results
}()

go func() {
    defer wg.Done()
    results, _ := q.Filter().ByName("Mary").Execute()
    resultsChan <- results
}()

go func() {
    defer wg.Done()
    results, _ := q.Filter().BySex("F").Execute()
    resultsChan <- results
}()

wg.Wait()
close(resultsChan)

// Collect results
for results := range resultsChan {
    // Process results
}
```

### Graph Metrics

```go
graph := q.Graph()
metrics := graph.GetMetrics()

// Get graph statistics
snapshot := metrics.GetSnapshot()
fmt.Printf("Nodes loaded: %d\n", snapshot.NodesLoaded)
fmt.Printf("Edges loaded: %d\n", snapshot.EdgesLoaded)
fmt.Printf("Graph build time: %v\n", snapshot.GraphBuildTime)
```

---

## Best Practices

1. **Always check errors**: Don't ignore error returns
2. **Use context**: Provide context in errors for debugging
3. **Close resources**: Call `Close()` on graphs with hybrid storage
4. **Monitor metrics**: Use metrics to identify performance issues
5. **Configure appropriately**: Adjust cache sizes and timeouts for your use case
6. **Use appropriate storage mode**: Choose eager, lazy, or hybrid based on dataset size

---

## Troubleshooting

### Common Issues

**Issue**: "Out of memory" errors
- **Solution**: Use lazy loading or hybrid storage for large datasets

**Issue**: Slow queries
- **Solution**: Enable caching, use indexed filters, check metrics

**Issue**: Missing individuals in results
- **Solution**: Check for parsing errors, verify data quality

**Issue**: High memory usage
- **Solution**: Reduce cache sizes, use hybrid storage, enable lazy loading

