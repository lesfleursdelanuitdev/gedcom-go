# GEDCOM Go Architecture

## Overview

GEDCOM Go is a research-grade genealogy toolkit built in Go, designed to handle datasets ranging from small family trees (50 individuals) to large-scale genealogical research (5M+ individuals).

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Application                        │
│                  (cmd/gedcom/commands/)                    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Core Packages                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Parser     │  │  Validator   │  │   Exporter   │     │
│  │  (internal)  │  │  (internal)  │  │  (internal)  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  GEDCOM Data Structures                     │
│                    (pkg/gedcom/)                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  GedcomTree  │  │   Records    │  │    Error     │     │
│  │              │  │  (Individual,│  │  Management  │     │
│  │              │  │   Family,    │  │              │     │
│  │              │  │   Note, etc)│  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Query System                             │
│                 (pkg/gedcom/query/)                         │
│  ┌────────────────────────────────────────────────────┐    │
│  │                    Graph                            │    │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐   │    │
│  │  │   Nodes   │  │   Edges    │  │  Metrics   │   │    │
│  │  └────────────┘  └────────────┘  └────────────┘   │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │              Query API                             │    │
│  │  (Filter, Ancestor, Descendant, Relationship, etc.) │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │            Storage Strategies                       │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐         │    │
│  │  │  Eager   │  │   Lazy   │  │  Hybrid  │         │    │
│  │  │ (Memory) │  │ (On-Demand)│ │(SQLite+ │         │    │
│  │  │          │  │          │  │ BadgerDB)│         │    │
│  │  └──────────┘  └──────────┘  └──────────┘         │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow

### Standard Flow (In-Memory)

```
GEDCOM File
    ↓
[Parser] → GedcomTree (in-memory)
    ↓
[Graph Builder] → Graph (nodes & edges)
    ↓
[Query API] → Results
    ↓
[CLI/Export] → Output
```

### Hybrid Storage Flow (Large Datasets)

```
GEDCOM File
    ↓
[Parser] → GedcomTree
    ↓
[Hybrid Builder] → SQLite (indexes) + BadgerDB (graph data)
    ↓
[Query API] → Loads on-demand from databases
    ↓
[LRU Cache] → Cached results
    ↓
[CLI/Export] → Output
```

## Core Components

### 1. Parser (`internal/parser/`)

**Responsibility**: Parse GEDCOM files into structured data

- **HierarchicalParser**: Main parser that builds tree structure
- **Line-by-line parsing**: Handles GEDCOM 5.5.1 specification
- **Error collection**: Collects parsing errors without stopping

### 2. Validator (`internal/validator/`)

**Responsibility**: Validate GEDCOM data structure and content

- **Basic validation**: Required fields, tag validity, cross-references
- **Advanced validation**: Date consistency, relationship logic
- **Parallel validation**: Concurrent validation for performance

### 3. Graph System (`pkg/gedcom/query/`)

**Responsibility**: Convert GEDCOM tree into queryable graph

#### Graph Structure

- **`graph.go`** (114 lines): Core `Graph` struct definition
- **`graph_nodes.go`** (298 lines): Node access methods
- **`graph_edges.go`** (101 lines): Edge management
- **`graph_hybrid.go`** (387 lines): Hybrid storage integration
- **`graph_hybrid_helpers.go`** (178 lines): Hybrid helper functions
- **`graph_metrics.go`** (421 lines): Graph metrics and analytics

#### Storage Modes

1. **Eager Loading** (default)
   - All nodes/edges loaded into memory
   - Fast queries (O(1) lookups)
   - High memory usage
   - Suitable for < 1M individuals

2. **Lazy Loading**
   - Only metadata loaded initially
   - Edges loaded on-demand
   - ~80% memory reduction
   - Suitable for 1M-5M individuals

3. **Hybrid Storage** (SQLite + BadgerDB)
   - SQLite: Indexes for fast filtering
   - BadgerDB: Graph structure for traversal
   - LRU cache for frequently accessed data
   - Scales to 10M+ individuals

### 4. Query API (`pkg/gedcom/query/`)

**Responsibility**: Provide fluent API for querying graph

#### Query Types

1. **FilterQuery**: Filter individuals by criteria
2. **IndividualQuery**: Query from specific individual
3. **AncestorQuery**: Find ancestors with options
4. **DescendantQuery**: Find descendants with options
5. **RelationshipQuery**: Calculate relationships
6. **PathQuery**: Find paths between individuals
7. **FamilyQuery**: Query family records
8. **MultiIndividualQuery**: Query multiple individuals
9. **GraphMetricsQuery**: Graph analytics
10. **EventsQuery**: Query events
11. **NotesQuery**: Query notes

### 5. Duplicate Detection (`pkg/gedcom/duplicate/`)

**Responsibility**: Identify potential duplicate individuals

- **Two-stage blocking pipeline**: O(n) complexity
- **Configurable strategies**: Name, date, place matching
- **Confidence scoring**: Ranked results with explanations

### 6. Diff System (`pkg/gedcom/diff/`)

**Responsibility**: Compare two GEDCOM files semantically

- **Matching strategies**: XREF, content, or hybrid
- **Change tracking**: Track who, when, what changed
- **Semantic understanding**: Date tolerance, equivalence

## Design Patterns

### 1. Builder Pattern
- `FilterQuery`, `AncestorQuery`, `DescendantQuery`: Fluent API construction

### 2. Strategy Pattern
- Storage strategies: Eager, Lazy, Hybrid
- Query execution strategies: Eager vs Hybrid

### 3. Factory Pattern
- `RecordFactory`: Creates record types from GEDCOM lines
- `NewGraph()`, `NewGraphWithConfig()`: Graph creation

### 4. Cache Pattern
- `queryCache`: In-memory query result cache
- `HybridCache`: LRU cache for hybrid storage

### 5. Pool Pattern
- `pool.go`: Object pooling for performance

## Performance Optimizations

### Memory Optimizations
1. **Integer IDs**: `uint32` instead of string XREFs
2. **Lazy Loading**: Only load what's needed
3. **Graph Partitioning**: Identify connected components
4. **LRU Caching**: Cache frequently accessed data

### Query Optimizations
1. **Indexed Filtering**: SQLite indexes for fast filtering
2. **Prepared Statements**: SQLite prepared statements
3. **Query Result Caching**: Cache query results
4. **Batch Operations**: Batch database operations

## Thread Safety

- **RWMutex**: Used throughout for concurrent access
- **Thread-safe operations**: All query operations are safe
- **Concurrent reads**: Multiple readers supported
- **Write protection**: Mutex for write operations

## Configuration System

### Config Structure

```go
type Config struct {
    Cache    CacheConfig    // Cache sizes
    Timeout  TimeoutConfig  // Operation timeouts
    Database DatabaseConfig // Database settings
}
```

### Configuration Sources

1. JSON config file (multiple search paths)
2. Environment variables (future)
3. CLI flags (future)
4. Default values

## Error Handling

### Error Types

- **GedcomError**: Structured error with severity, message, line number, context
- **ErrorSeverity**: hint, info, warning, severe
- **ErrorManager**: Thread-safe error collection

### Error Flow

```
Operation → Error → ErrorManager → Collection → Reporting
```

## Scalability

### Dataset Sizes

- **Small (50-50K)**: Eager loading, in-memory
- **Medium (50K-1M)**: Lazy loading
- **Large (1M-5M)**: Lazy loading with optimizations
- **Very Large (5M-10M+)**: Hybrid storage (SQLite + BadgerDB)

### Performance Characteristics

- **Graph Building**: O(n) where n = number of records
- **Query Operations**: O(1) to O(log n) depending on operation
- **Memory Usage**: Configurable based on storage mode
- **Disk Usage**: Hybrid storage uses persistent databases

## Extension Points

### Custom Validators
- Implement `Validator` interface
- Add to `AdvancedValidator`

### Custom Filters
- Extend `FilterQuery` with custom filter methods
- Use `Where()` for custom filter functions

### Custom Storage
- Implement storage interface (future)
- Add new storage backend

## Testing Strategy

### Unit Tests
- Individual component testing
- Mock dependencies
- Fast execution

### Integration Tests
- Component interaction testing
- Real data scenarios
- End-to-end workflows

### Performance Tests
- Benchmark tests
- Stress tests (1M, 5M, 10M individuals)
- Regression tests

## Future Enhancements

1. **Metrics Collection**: Query times, cache hit rates
2. **Distributed Storage**: Multi-node support
3. **Graph Visualization**: Visual representation
4. **More Export Formats**: Additional output formats
5. **API Server**: REST API for remote access

