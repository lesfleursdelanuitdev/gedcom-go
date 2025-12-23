# GEDCOM Go Codebase Analysis

**Generated:** 2025-01-27  
**Total Go Files:** 185 (106 non-test, 79 test files)  
**Total Lines of Code:** ~24,220 (excluding tests)

## Executive Summary

GEDCOM Go is a well-architected, production-ready genealogy toolkit written in Go. The codebase demonstrates:

- **Strong separation of concerns** with clear package boundaries
- **Multiple storage strategies** (in-memory, lazy loading, hybrid SQLite+BadgerDB)
- **Comprehensive test coverage** (79 test files, ~43% test-to-code ratio)
- **Performance optimizations** for large datasets (5M+ individuals)
- **Thread-safe design** throughout

## Architecture Overview

### High-Level Structure

```
gedcom-go/
├── cmd/gedcom/          # CLI application (Cobra-based)
│   ├── commands/        # Command implementations
│   └── internal/        # CLI utilities (config, output, progress)
├── pkg/gedcom/          # Core GEDCOM data structures
│   ├── query/           # Graph query system (largest component)
│   ├── duplicate/       # Duplicate detection with blocking
│   └── diff/            # File comparison utilities
├── internal/
│   ├── parser/          # GEDCOM file parsing
│   ├── validator/       # Data validation
│   └── exporter/        # Export to various formats
└── stress_test.go       # Performance/stress testing suite
```

### Data Flow

```
GEDCOM File
    ↓
[Parser] → GedcomTree (in-memory)
    ↓
[Graph Builder] → Graph (with nodes & edges)
    ↓
[Query API] → Results
    ↓
[CLI/Export] → Output
```

**Alternative Flow (Hybrid Storage):**
```
GEDCOM File
    ↓
[Parser] → GedcomTree
    ↓
[Hybrid Builder] → SQLite (indexes) + BadgerDB (graph data)
    ↓
[Query API] → Loads on-demand from databases
```

## Key Components

### 1. Core Data Structures (`pkg/gedcom/`)

**GedcomTree** - Root container for all records
- Thread-safe with `sync.RWMutex`
- Organized by record type (individuals, families, notes, etc.)
- Cross-reference index for fast lookups
- ~236 lines

**Record Types:**
- `IndividualRecord` - Person records with name, dates, relationships
- `FamilyRecord` - Family units linking individuals
- `NoteRecord`, `SourceRecord`, `RepositoryRecord`, etc.
- All implement the `Record` interface

**Key Features:**
- Type-safe accessors (`GetName()`, `GetBirthDate()`, etc.)
- Hierarchical structure via `GedcomLine` children
- Error management via `ErrorManager`

### 2. Query System (`pkg/gedcom/query/`)

**Graph** - Central graph structure (1,124 lines)
- Converts GEDCOM tree into graph with nodes and edges
- Supports three modes:
  1. **Eager loading** (default) - All nodes/edges loaded upfront
  2. **Lazy loading** - Load nodes/edges on-demand
  3. **Hybrid storage** - SQLite + BadgerDB for persistence

**Key Optimizations:**
- **Integer IDs**: Uses `uint32` internally instead of string XREFs (memory savings)
- **Lazy Loading**: Only loads nodes/edges when accessed (~80% memory reduction)
- **Graph Partitioning**: Identifies connected components for selective loading
- **LRU Cache**: Query result caching for hybrid storage

**Query API:**
- `FilterQuery` - Filter individuals by name, date, place, etc.
- `IndividualQuery` - Query specific individual (ancestors, descendants, relationships)
- `FamilyQuery` - Query family relationships
- `PathQuery` - Find paths between individuals
- `RelationshipQuery` - Calculate relationship degrees

**Storage Strategies:**

1. **In-Memory (Default)**
   - Fast, but limited by RAM
   - Good for datasets up to ~1.5M individuals

2. **Lazy Loading**
   - Builds skeleton (`NodeMetadata`) first
   - Loads full node data and edges on-demand
   - Reduces peak memory by ~80%
   - Good for 1.5M-5M individuals

3. **Hybrid Storage** (SQLite + BadgerDB)
   - SQLite: Indexes (name, date, place, sex, boolean flags, FTS5)
   - BadgerDB: Graph structure (nodes, edges, components)
   - Memory-mapped files for OS-managed paging
   - Supports 10M+ individuals

### 3. Duplicate Detection (`pkg/gedcom/duplicate/`)

**Two-Stage Blocking Pipeline:**

**Stage A: Blocking (O(n) complexity)**
- Generates candidate pairs using inverted indexes
- Multiple blocking strategies:
  - `surname_soundex + birth_year`
  - `surname_prefix + given_prefix + birth_year_bucket`
  - `birth_place_token + surname_soundex`
  - `parents' surnames + birth_year`
- Reduces comparisons from O(n²) to O(n × avg_block_size)

**Stage B: Scoring (expensive, high precision)**
- Detailed similarity scoring:
  - Name similarity (with phonetic matching)
  - Date similarity (with tolerance)
  - Place similarity
  - Sex matching
  - Relationship matching
- Only scores candidates from Stage A

**Performance:**
- 1.5M individuals: ~8.96 seconds (with blocking)
- Parallel processing with worker pools
- Adaptive blocking (skips overly large blocks)

### 4. Parser (`internal/parser/`)

**HierarchicalParser** - Main parser
- Stack-based algorithm for hierarchical structure
- Handles continuation lines (CONT/CONC)
- Encoding detection (UTF-8, ANSEL, ASCII)
- Error recovery and reporting

**Features:**
- Line-by-line parsing with parent tracking
- Record factory pattern for type creation
- Thread-safe error collection

### 5. CLI (`cmd/gedcom/`)

**Commands:**
- `interactive` - Interactive exploration mode
- `duplicates` - Find potential duplicates
- `search` - Search with filters
- `validate` - Data quality validation
- `export` - Export to JSON/XML/YAML/GEDCOM
- `parse` - Parse and validate files

**Features:**
- Cobra-based CLI framework
- Progress bars for long operations
- Colored output (configurable)
- Config file support

## Design Patterns

### 1. **Factory Pattern**
- `RecordFactory` creates appropriate record types from `GedcomLine`
- `NewIndividualRecord()`, `NewFamilyRecord()`, etc.

### 2. **Builder Pattern**
- `BuildGraph()` - Builds graph from tree
- `BuildGraphLazy()` - Builds lazy-loading graph
- `BuildGraphHybrid()` - Builds hybrid storage graph

### 3. **Strategy Pattern**
- Multiple storage strategies (eager, lazy, hybrid)
- Multiple blocking strategies for duplicate detection
- Multiple export formats

### 4. **Cache Pattern**
- `queryCache` - LRU cache for query results
- `HybridCache` - Multi-level cache (nodes, XREF mappings, queries)

### 5. **Pool Pattern**
- Memory pools for temporary data structures (`pool.go`)
- Reduces allocations in hot paths

## Performance Characteristics

### Memory Usage (Peak)

| Dataset Size | Eager Loading | Lazy Loading | Hybrid Storage |
|-------------|---------------|--------------|----------------|
| 10K         | ~150 MB       | ~150 MB      | ~150 MB        |
| 200K        | ~3 GB         | ~600 MB      | ~600 MB        |
| 1.5M        | ~21 GB        | ~4 GB        | ~4 GB          |
| 5M          | OOM           | ~14 GB       | ~14 GB         |
| 10M         | OOM           | OOM          | ~28 GB         |

### Query Performance

- **Small trees (50-50K)**: Instant (< 1 second)
- **Medium trees (50K-200K)**: < 1 second
- **Large datasets (500K-5M)**: 1-5 seconds (with scoping)

### Duplicate Detection

- **1.5M individuals**: ~8.96 seconds (with blocking)
- **Without blocking**: Would be O(n²) and timeout

## Testing

### Test Coverage

- **79 test files** covering all major components
- **Unit tests** for individual functions
- **Integration tests** for end-to-end workflows
- **Performance tests** for stress testing
- **Timeout tests** (2-minute limit per test)

### Test Organization

- `*_test.go` files alongside source files
- `stress_test.go` - Comprehensive stress testing suite
- Test data in `testdata/` directory

## Strengths

1. **Scalability**
   - Handles datasets from 50 to 10M+ individuals
   - Multiple storage strategies for different scales
   - Memory-efficient with lazy loading and hybrid storage

2. **Performance**
   - O(n²) → O(n) duplicate detection with blocking
   - Parallel processing where beneficial
   - Efficient indexing and caching

3. **Code Quality**
   - Clear package boundaries
   - Comprehensive documentation
   - Thread-safe design
   - Error handling throughout

4. **Usability**
   - User-friendly CLI
   - Interactive mode for exploration
   - Clear error messages and warnings
   - Comprehensive README and workflows guide

5. **Extensibility**
   - Plugin-like architecture for validators
   - Multiple export formats
   - Configurable duplicate detection

## Areas for Potential Improvement

### 1. **Documentation**
- Some complex algorithms could use more inline comments
- API documentation could be more comprehensive
- Architecture diagrams would be helpful

### 2. **Error Handling**
- Some functions return `nil` on error (could use explicit errors)
- Error context could be richer in some cases

### 3. **Code Organization**
- `query/` package is large (could be split into sub-packages)
- Graph functionality has been refactored into multiple files:
  - `graph.go` (114 lines) - Core structure
  - `graph_nodes.go` (298 lines) - Node access
  - `graph_edges.go` (101 lines) - Edge management
  - `graph_hybrid.go` (387 lines) - Hybrid storage
  - `graph_hybrid_helpers.go` (178 lines) - Hybrid helpers
  - `graph_metrics.go` (421 lines) - Metrics

### 4. **Testing**
- Some edge cases might not be covered
- Integration tests could be more comprehensive
- Performance regression tests could be automated

### 5. **Dependencies**
- BadgerDB and SQLite are heavy dependencies
- Could consider lighter alternatives for smaller use cases

### 6. **Configuration**
- Some hardcoded values (cache sizes, timeouts)
- Could be more configurable via config file

## Dependencies

### Core Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/dgraph-io/badger/v4` - Key-value store for hybrid storage
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/hashicorp/golang-lru/v2` - LRU cache
- `github.com/c-bata/go-prompt` - Interactive prompt
- `github.com/schollz/progressbar/v3` - Progress bars

### Development Dependencies
- Standard Go testing package
- No external test frameworks (keeps it simple)

## Recommendations

### Short Term
1. Add more inline documentation for complex algorithms
2. Consider splitting large files (`graph.go`, `builder.go`)
3. Add performance regression tests to CI

### Medium Term
1. Consider splitting `query/` package into sub-packages
2. Add more configuration options
3. Improve error messages with more context

### Long Term
1. Consider plugin system for custom validators/exporters
2. Add support for incremental updates (currently rebuilds graph)
3. Consider distributed storage for very large datasets

## Conclusion

GEDCOM Go is a well-engineered, production-ready genealogy toolkit. The codebase demonstrates:

- **Strong architecture** with clear separation of concerns
- **Performance optimizations** for large-scale datasets
- **Comprehensive testing** with good coverage
- **User-focused design** with clear CLI and documentation

The hybrid storage approach (SQLite + BadgerDB) is particularly innovative and allows the tool to scale to 10M+ individuals while maintaining good query performance.

The codebase is maintainable, testable, and ready for production use.

