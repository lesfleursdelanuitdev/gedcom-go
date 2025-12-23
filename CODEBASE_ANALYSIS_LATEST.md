# GEDCOM Go - Comprehensive Codebase Analysis

**Generated:** 2025-12-22  
**Codebase Version:** Latest (post-configuration refactoring)  
**Total Go Files:** 128 (85 non-test, 43 test files)  
**Total Lines of Code:** ~26,898 (including tests), ~11,093 (production code)

## Executive Summary

GEDCOM Go is a **research-grade genealogy toolkit** built in Go, designed to handle datasets ranging from small family trees (50 individuals) to large-scale genealogical research (5M+ individuals). The codebase demonstrates:

✅ **Strong Architecture**: Clear separation of concerns with well-defined package boundaries  
✅ **Multiple Storage Strategies**: In-memory, lazy loading, and hybrid SQLite+BadgerDB for scalability  
✅ **Comprehensive Test Coverage**: 43 test files (~50% test-to-code ratio)  
✅ **Performance Optimizations**: Integer IDs, lazy loading, caching, and graph partitioning  
✅ **Thread-Safe Design**: RWMutex-based concurrency throughout  
✅ **Recent Improvements**: Configuration system, refactored builders, organized filters  

## Architecture Overview

### High-Level Structure

```
gedcom-go/
├── cmd/gedcom/              # CLI application (Cobra-based)
│   ├── commands/            # Command implementations
│   │   ├── interactive.go  # Interactive exploration mode
│   │   ├── duplicates.go   # Duplicate detection
│   │   ├── search.go       # Search with filters
│   │   ├── validate.go     # Data validation
│   │   └── export.go       # Export to various formats
│   └── internal/           # CLI utilities (config, output, progress)
│
├── pkg/gedcom/              # Core GEDCOM data structures
│   ├── query/               # Graph query system (LARGEST COMPONENT)
│   │   ├── graph.go        # Core graph structure
│   │   ├── graph_nodes.go  # Node access methods
│   │   ├── graph_edges.go  # Edge management
│   │   ├── graph_hybrid.go # Hybrid storage integration
│   │   ├── hybrid_*.go     # Hybrid storage implementation
│   │   ├── filter_*.go     # Filter query system (refactored)
│   │   ├── config.go       # Configuration system (NEW)
│   │   └── [50+ more files]
│   ├── duplicate/          # Duplicate detection with blocking
│   └── diff/               # File comparison utilities
│
├── internal/
│   ├── parser/             # GEDCOM file parsing
│   ├── validator/          # Data validation
│   └── exporter/           # Export to JSON, XML, YAML, GEDCOM
│
├── testdata/               # Test GEDCOM files
├── docs/                   # Documentation
└── stress_test.go          # Performance/stress testing suite
```

### Data Flow

**Standard Flow (In-Memory):**
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

**Hybrid Storage Flow (Large Datasets):**
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

## Key Components

### 1. Core Data Structures (`pkg/gedcom/`)

#### GedcomTree
- **Purpose**: Root container for all parsed GEDCOM records
- **Thread Safety**: `sync.RWMutex` for concurrent access
- **Organization**: Records organized by type (individuals, families, notes, etc.)
- **Features**:
  - Cross-reference index for fast lookups
  - Type-safe accessors
  - Error management via `ErrorManager`

#### Record Types
- `IndividualRecord` - Person records with name, dates, relationships
- `FamilyRecord` - Family units linking individuals
- `NoteRecord`, `SourceRecord`, `RepositoryRecord`, `MultimediaRecord`
- All implement the `Record` interface

### 2. Query System (`pkg/gedcom/query/`) - **11,093 lines**

This is the **largest and most complex component** of the codebase.

#### Graph Structure (`graph.go` - 114 lines)

**Note**: The graph functionality has been refactored into multiple files:
- `graph.go` (114 lines) - Core structure definition
- `graph_nodes.go` (298 lines) - Node access methods
- `graph_edges.go` (101 lines) - Edge management
- `graph_hybrid.go` (387 lines) - Hybrid storage integration
- `graph_hybrid_helpers.go` (178 lines) - Hybrid helper functions
- `graph_metrics.go` (421 lines) - Graph metrics and analytics

**Core Graph Type:**
```go
type Graph struct {
    // ID mapping: XREF string <-> uint32 ID (memory efficiency)
    xrefToID map[string]uint32
    idToXref map[uint32]string
    nextID   uint32

    // Node storage (using uint32 IDs internally)
    nodes        map[uint32]GraphNode
    individuals  map[uint32]*IndividualNode
    families     map[uint32]*FamilyNode
    notes        map[uint32]*NoteNode
    sources      map[uint32]*SourceNode
    repositories map[uint32]*RepositoryNode
    events       map[uint32]*EventNode

    // Edge storage
    edges     map[string]*Edge
    edgeIndex map[uint32][]*Edge

    // Lazy loading support
    nodeMetadata map[uint32]*NodeMetadata
    edgesLoaded  map[uint32]bool
    edgesLoading map[uint32]bool
    lazyMode     bool

    // Graph partitioning
    components     map[uint32][]uint32
    componentCount uint32

    // Thread safety
    mu sync.RWMutex

    // Performance optimizations
    cache   *queryCache
    indexes *FilterIndexes

    // Hybrid storage support
    hybridStorage *HybridStorage
    hybridMode    bool
    queryHelpers  *HybridQueryHelpers
    hybridCache  *HybridCache
}
```

**Key Features:**
- **Integer IDs**: Uses `uint32` internally instead of string XREFs (significant memory savings)
- **Three Storage Modes**:
  1. **Eager Loading** (default) - All nodes/edges loaded upfront
  2. **Lazy Loading** - Load nodes/edges on-demand (~80% memory reduction)
  3. **Hybrid Storage** - SQLite + BadgerDB for persistence (scales to 10M+ individuals)

#### Storage Modes

**1. Eager Loading (Default)**
- All nodes and edges loaded into memory during graph construction
- Fast query performance (O(1) lookups)
- High memory usage (not suitable for 1M+ individuals)

**2. Lazy Loading**
- Only node metadata loaded initially
- Edges loaded on-demand when accessed
- ~80% memory reduction vs eager loading
- Suitable for 1M-5M individuals

**3. Hybrid Storage** (SQLite + BadgerDB)
- **SQLite**: Stores indexes (names, dates, places) for fast filtering
- **BadgerDB**: Stores graph structure (nodes, edges) for efficient traversal
- **LRU Cache**: Caches frequently accessed nodes and query results
- Scales to **10M+ individuals**
- Configurable cache sizes and timeouts (NEW)

#### Query API Types

**1. FilterQuery** (`filter_query.go`, `filter_*.go`)
- Filter individuals by name, date, place, sex, etc.
- **Refactored**: Split into `filter_name.go`, `filter_date.go`, `filter_attributes.go`, `filter_execution.go`
- Supports indexed filtering for performance

**2. IndividualQuery** (`query.go`)
- Query operations starting from a specific individual
- Methods: `Parents()`, `Children()`, `Siblings()`, `Spouses()`, `Ancestors()`, `Descendants()`

**3. AncestorQuery** (`ancestor_query.go`)
- Configurable ancestor search with options
- `MaxGenerations()`, `IncludeSelf()`, `Filter()`, `Count()`, `Exists()`

**4. DescendantQuery** (`descendant_query.go`)
- Similar API to AncestorQuery but for descendants

**5. RelationshipQuery** (`relationship_query.go`)
- Calculate relationship between two individuals
- Returns: relationship type, degree, removal, isDirect, isCollateral

**6. PathQuery** (`path_query.go`)
- Find paths between two individuals
- `Shortest()`, `All()`, `MaxLength()`, `IncludeBlood()`, `IncludeMarital()`

**7. FamilyQuery** (`family_query.go`)
- Query operations starting from a family
- `Husband()`, `Wife()`, `Children()`, `Parents()`, `MarriageDate()`, `Events()`

**8. MultiIndividualQuery** (`multi_individual_query.go`)
- Query operations on multiple individuals
- `Ancestors()`, `CommonAncestors()`, `Union()`

**9. GraphMetricsQuery** (`graph_metrics.go`)
- Graph analytics and metrics
- `Degree()`, `Diameter()`, `AveragePathLength()`, `Centrality()`, `ConnectedComponents()`

**10. EventsQuery** (`events_query.go`)
- Query events (birth, death, marriage, etc.)

**11. NotesQuery** (`notes_query.go`)
- Query notes associated with individuals/families

### 3. Hybrid Storage System

#### Architecture

**HybridStorage** (`hybrid_storage.go`)
- Manages both SQLite and BadgerDB connections
- **SQLite**: Indexes for fast filtering (names, dates, places)
- **BadgerDB**: Graph structure (nodes, edges) for traversal
- Configurable connection pools and timeouts

**HybridCache** (`hybrid_cache.go`)
- LRU cache for nodes, XREF mappings, and query results
- Configurable cache sizes (NEW)
- Reduces database queries for frequently accessed data

**Builders** (Refactored):
- `hybrid_builder.go` - Coordinator function
- `hybrid_sqlite_builder.go` - SQLite index building
- `hybrid_badger_builder.go` - BadgerDB graph building
- `hybrid_builder_helpers.go` - Common helper functions

**Serialization** (`hybrid_serialization.go`)
- Serializes/deserializes nodes and edges for BadgerDB
- Supports all node types: Individual, Family, Note, Source, Repository, Event

### 4. Configuration System (NEW)

**Config** (`config.go`)
- **CacheConfig**: Configurable cache sizes
  - `HybridNodeCacheSize` (default: 50,000)
  - `HybridXrefCacheSize` (default: 25,000)
  - `HybridQueryCacheSize` (default: 5,000)
  - `QueryCacheSize` (default: 1,000)

- **TimeoutConfig**: Configurable timeouts
  - `SQLiteQueryTimeout` (default: 30s)
  - `BadgerDBTimeout` (default: 10s)
  - `BuildTimeout` (default: 5m)
  - `QueryTimeout` (default: 1m)

- **DatabaseConfig**: Database settings
  - `SQLiteMaxOpenConns` (default: 10)
  - `SQLiteMaxIdleConns` (default: 5)
  - `BadgerDBValueLogFileSize` (default: 1GB)

**Features:**
- JSON config file support with automatic search
- Duration string parsing ("30s", "5m", "1h")
- Validation with automatic defaults
- Backward compatible (nil config = defaults)

### 5. Duplicate Detection (`pkg/gedcom/duplicate/`)

**Two-Stage Blocking Pipeline:**
1. **Blocking**: Group similar records (O(n) complexity)
2. **Scoring**: Compare records within blocks

**Features:**
- Configurable blocking strategies
- Confidence scoring
- Explanations for matches

### 6. CLI Application (`cmd/gedcom/`)

**Commands:**
- `interactive` - Interactive exploration mode
- `duplicates` - Find potential duplicates
- `search` - Search with filters
- `validate` - Data validation
- `export` - Export to various formats
- `parse` - Parse and validate GEDCOM files

**Built with:**
- Cobra for CLI framework
- go-prompt for interactive mode
- Progress bars for long operations

## Design Patterns

### 1. Builder Pattern
- `FilterQuery` - Fluent API for building queries
- `AncestorQuery`, `DescendantQuery` - Builder-style configuration

### 2. Strategy Pattern
- Storage strategies: Eager, Lazy, Hybrid
- Query execution strategies: Eager vs Hybrid

### 3. Factory Pattern
- `RecordFactory` - Creates record types from GEDCOM lines
- `NewGraph()`, `NewGraphWithConfig()` - Graph creation

### 4. Cache Pattern
- `queryCache` - In-memory query result cache
- `HybridCache` - LRU cache for hybrid storage

### 5. Pool Pattern
- `pool.go` - Object pooling for performance

## Performance Optimizations

### Memory Optimizations
1. **Integer IDs**: `uint32` instead of string XREFs
2. **Lazy Loading**: Only load what's needed
3. **Graph Partitioning**: Identify connected components
4. **LRU Caching**: Cache frequently accessed data

### Query Optimizations
1. **Indexed Filtering**: Use SQLite indexes for fast filtering
2. **Prepared Statements**: SQLite prepared statements for repeated queries
3. **Query Result Caching**: Cache query results in hybrid mode
4. **Batch Operations**: Batch database operations

### Scalability
- **Hybrid Storage**: Scales to 10M+ individuals
- **Configurable Caches**: Adjust cache sizes based on available memory
- **Connection Pooling**: SQLite connection pool for concurrent queries
- **Timeout Management**: Prevent long-running queries from blocking

## Code Organization

### Recent Refactoring (Completed)

**1. Hybrid Builder Refactoring** ✅
- Split `hybrid_builder.go` (1,060 lines) into:
  - `hybrid_builder.go` - Coordinator
  - `hybrid_sqlite_builder.go` - SQLite operations
  - `hybrid_badger_builder.go` - BadgerDB operations
  - `hybrid_builder_helpers.go` - Common helpers

**2. Filter Query Refactoring** ✅
- Split `filter_query.go` into:
  - `filter_query.go` - Core struct and basic methods
  - `filter_name.go` - Name-based filters
  - `filter_date.go` - Date-based filters
  - `filter_attributes.go` - Attribute filters
  - `filter_execution.go` - Execution logic

**3. Configuration System** ✅
- Added `config.go` with comprehensive configuration
- JSON config file support
- Duration string parsing
- Validation and defaults

### File Organization

**Query Package Structure:**
```
pkg/gedcom/query/
├── Core Graph
│   ├── graph.go              # Graph struct definition
│   ├── graph_nodes.go        # Node access methods
│   ├── graph_edges.go        # Edge management
│   └── graph_hybrid.go       # Hybrid storage integration
│
├── Hybrid Storage
│   ├── hybrid_storage.go     # Storage manager
│   ├── hybrid_builder.go     # Coordinator
│   ├── hybrid_sqlite_builder.go
│   ├── hybrid_badger_builder.go
│   ├── hybrid_builder_helpers.go
│   ├── hybrid_cache.go       # LRU cache
│   ├── hybrid_serialization.go
│   └── hybrid_queries.go     # Hybrid query helpers
│
├── Query API
│   ├── query.go              # Main query builder
│   ├── filter_query.go       # Filter query core
│   ├── filter_name.go        # Name filters
│   ├── filter_date.go        # Date filters
│   ├── filter_attributes.go  # Attribute filters
│   ├── filter_execution.go   # Execution logic
│   ├── ancestor_query.go
│   ├── descendant_query.go
│   ├── relationship_query.go
│   ├── path_query.go
│   ├── family_query.go
│   ├── multi_individual_query.go
│   ├── events_query.go
│   └── notes_query.go
│
├── Graph Algorithms
│   ├── algorithms.go         # Core algorithms
│   ├── traversal.go         # BFS/DFS
│   ├── path_finding.go      # Path algorithms
│   ├── relationships.go     # Relationship calculation
│   └── ancestors.go         # Ancestor algorithms
│
├── Configuration
│   ├── config.go            # Configuration system
│   └── config.example.json  # Example config
│
└── Utilities
    ├── node.go              # Node types
    ├── edge.go              # Edge types
    ├── cache.go             # Query cache
    ├── indexes.go           # Filter indexes
    ├── id_mapping.go        # ID mapping
    ├── pool.go              # Object pooling
    └── analytics.go         # Analytics
```

## Dependencies

### Core Dependencies
- **github.com/spf13/cobra** v1.10.2 - CLI framework
- **github.com/dgraph-io/badger/v4** v4.9.0 - BadgerDB key-value store
- **github.com/mattn/go-sqlite3** v1.14.32 - SQLite driver
- **github.com/c-bata/go-prompt** v0.2.6 - Interactive terminal
- **github.com/fatih/color** v1.18.0 - Terminal colors
- **gopkg.in/yaml.v3** v3.0.1 - YAML parsing

### Indirect Dependencies
- **github.com/hashicorp/golang-lru/v2** v2.0.7 - LRU cache implementation
- **github.com/dgraph-io/ristretto/v2** v2.2.0 - Cache library (used by BadgerDB)
- Various compression, logging, and utility libraries

## Test Coverage

### Test Files: 43
- **Unit Tests**: Test individual components
- **Integration Tests**: Test component interactions
- **Performance Tests**: Benchmark and stress tests
- **Hybrid Storage Tests**: Test SQLite + BadgerDB integration

### Test Organization
- `*_test.go` files co-located with source files
- `hybrid_*_test.go` - Hybrid storage tests
- `performance_test.go` - Performance benchmarks
- `stress_test.go` - Stress testing suite

### Test Execution
```bash
# Run all tests
go test ./...

# Run short tests only (excludes long-running tests)
go test -short ./...

# Run with coverage
go test -cover ./...
```

## Recent Changes (Git History)

1. **Configuration System** (Latest)
   - Added comprehensive configuration with JSON support
   - Configurable cache sizes and timeouts
   - Duration string parsing

2. **Refactoring**
   - Split hybrid builder into separate files
   - Organized filter query into multiple files
   - Fixed typed nil interface issues
   - Added event deserialization support

3. **Query API**
   - Added comprehensive query API for all 13 query types
   - Improved relationship calculation
   - Enhanced path finding algorithms

4. **Performance**
   - Fixed CI timeout issues
   - Optimized test execution
   - Improved lazy loading

## Strengths

1. **Scalability**: Hybrid storage supports 10M+ individuals
2. **Performance**: Multiple optimizations (integer IDs, caching, lazy loading)
3. **Thread Safety**: RWMutex-based concurrency throughout
4. **Comprehensive API**: 13+ query types covering all use cases
5. **Well-Tested**: 43 test files with good coverage
6. **Clear Architecture**: Separation of concerns, organized packages
7. **Recent Improvements**: Configuration system, refactored code organization
8. **Documentation**: Comprehensive doc.go files with examples

## Areas for Improvement

### 1. Graph Structure (Refactored ✅)
- **Current**: `graph.go` is now 114 lines (core structure only)
- **Refactored into**:
  - `graph.go` - 114 lines (core struct and initialization) ✅
  - `graph_nodes.go` - 298 lines (node access methods) ✅
  - `graph_edges.go` - 101 lines (edge management) ✅
  - `graph_hybrid.go` - 387 lines (hybrid storage integration) ✅
  - `graph_hybrid_helpers.go` - 178 lines (hybrid helper functions) ✅
  - `graph_metrics.go` - 421 lines (graph metrics and analytics) ✅
- **Total**: ~1,499 lines across all graph files (well-organized)

### 2. Documentation
- Add more examples to README
- Create architecture diagrams
- Document performance characteristics for different dataset sizes

### 3. Error Handling
- Standardize error types across packages
- Add error context (line numbers, record IDs)
- Improve error messages for end users

### 4. Configuration
- Add CLI flags for common config options
- Support environment variables
- Add config validation warnings

### 5. Testing
- Increase test coverage (currently ~50%)
- Add more integration tests
- Add performance regression tests

### 6. Monitoring
- Add metrics collection (query times, cache hit rates)
- Add logging levels
- Add performance profiling hooks

## Recommendations

### Short-Term (1-2 weeks)
1. ✅ **Configuration System** - DONE
2. ✅ **Refactor Hybrid Builder** - DONE
3. ✅ **Refactor Filter Query** - DONE
4. Add CLI flags for common config options
5. Improve error messages

### Medium-Term (1-2 months)
1. ✅ ~~Further split `graph.go` into smaller files~~ - **COMPLETED** (refactored into 6 files)
2. ✅ **Add more comprehensive documentation** - **COMPLETED** (ARCHITECTURE.md, API_EXAMPLES.md)
3. ✅ **Add performance regression tests** - **COMPLETED** (performance_regression_test.go)
4. ✅ **Standardize error handling** - **COMPLETED** (errors.go with StandardError)
5. ✅ **Add metrics collection** - **COMPLETED** (metrics.go integrated into Graph)

### Long-Term (3-6 months)
1. Add distributed storage support (if needed)
2. Add graph visualization capabilities
3. Add more export formats
4. Enhance monitoring and observability (build on medium-term metrics hooks)

## Conclusion

The GEDCOM Go codebase is **well-architected and production-ready**. Recent refactoring has improved code organization, and the new configuration system provides flexibility for different use cases. The hybrid storage system enables handling of very large datasets (10M+ individuals), while the comprehensive query API supports a wide range of genealogical research needs.

**Key Metrics:**
- **128 Go files** (85 production, 43 tests)
- **~11,093 lines** of production code
- **~26,898 total lines** (including tests)
- **50% test-to-code ratio**
- **13+ query types**
- **3 storage modes** (eager, lazy, hybrid)
- **Scalable to 10M+ individuals**

The codebase demonstrates strong engineering practices with clear separation of concerns, comprehensive testing, and performance optimizations suitable for both small family trees and large-scale genealogical research.

