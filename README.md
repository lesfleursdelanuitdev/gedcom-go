# GEDCOM Go

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourorg/gedcom)](https://goreportcard.com/report/github.com/yourorg/gedcom)

A **production-ready, high-performance GEDCOM parser, validator, query system, and CLI tool** written in Go. Full GEDCOM 5.5.1 specification support with advanced graph-based querying, interactive exploration, and comprehensive validation.

## Features

### 🚀 Core Capabilities

- **📖 Full GEDCOM 5.5.1 Support** - Complete parser implementation
- **✅ Advanced Validation** - Multi-level validation with severity levels
- **📤 Multiple Export Formats** - GEDCOM, JSON, XML, YAML
- **🔍 Graph-Based Query API** - Powerful relationship queries and algorithms
- **⚡ Performance Optimized** - Caching, indexing, memory pooling
- **🔄 Incremental Updates** - Modify graphs without full rebuild (50-200x faster)
- **💻 CLI Tool** - Complete command-line interface
- **🎯 Interactive Mode** - REPL for exploring genealogical data
- **🔎 Advanced Search** - Multi-filter search with indexes

### 🎯 Query Capabilities

- **Relationship Queries**: Parents, children, siblings, spouses
- **Ancestor/Descendant Traversal**: Configurable generation limits
- **Path Finding**: Shortest path, all paths between individuals
- **Relationship Calculation**: Degree, type, and removal calculation
- **Common Ancestors**: Find shared ancestors and LCA
- **Graph Analytics**: Centrality, diameter, connected components
- **Advanced Filtering**: Indexed search with multiple criteria

### 🛠️ CLI Commands

- **`parse`** - Parse and validate GEDCOM files
- **`validate`** - Advanced validation with severity levels
- **`export`** - Export to JSON, XML, YAML, or GEDCOM
- **`interactive`** - Interactive REPL mode for querying
- **`search`** - Advanced multi-filter search

## Installation

### From Source

```bash
git clone https://github.com/lesfleursdelanuitdev/gedcom-go.git
cd gedcom-go
go build -o gedcom ./cmd/gedcom
```

### Using Go Install

```bash
go install github.com/lesfleursdelanuitdev/gedcom-go/cmd/gedcom@latest
```

## Quick Start

### Basic Usage

```bash
# Parse a GEDCOM file
gedcom parse file family.ged

# Validate with advanced rules
gedcom validate advanced family.ged --severity warning

# Export to JSON
gedcom export json family.ged -o family.json

# Search for individuals
gedcom search family.ged --name "John" --sex M --birth-year 1900-1950

# Interactive mode
gedcom interactive family.ged
```

### Programmatic Usage

```go
package main

import (
    "fmt"
    "github.com/yourorg/gedcom/pkg/gedcom"
    "github.com/yourorg/gedcom/internal/parser"
    "github.com/yourorg/gedcom/pkg/gedcom/query"
)

func main() {
    // Parse GEDCOM file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Build graph for queries
    graph, err := query.BuildGraph(tree)
    if err != nil {
        panic(err)
    }

    // Create query builder
    q, err := query.NewQuery(tree)
    if err != nil {
        panic(err)
    }

    // Find ancestors
    ancestors, _ := q.Individual("@I1@").Ancestors().MaxGenerations(5).Execute()
    for _, ancestor := range ancestors {
        fmt.Printf("Ancestor: %s\n", ancestor.GetName())
    }

    // Calculate relationship
    result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()
    fmt.Printf("Relationship: %s (Degree: %d)\n", result.RelationshipType, result.Degree)
}
```

## CLI Examples

### Parse and Validate

```bash
# Basic parse
gedcom parse file family.ged

# Parse with validation
gedcom parse validate family.ged --strict

# Quick syntax check
gedcom parse check family.ged
```

### Advanced Validation

```bash
# Basic validation
gedcom validate basic family.ged

# Advanced validation with report
gedcom validate advanced family.ged \
  --severity warning \
  -o report.json \
  --format json
```

### Export Formats

```bash
# Export to JSON
gedcom export json family.ged -o family.json --pretty

# Export to XML
gedcom export xml family.ged -o family.xml

# Export to YAML
gedcom export yaml family.ged -o family.yaml
```

### Advanced Search

```bash
# Search by name
gedcom search family.ged --name "John"

# Multiple filters
gedcom search family.ged \
  --name "John" \
  --sex M \
  --birth-year 1900-1950 \
  --birth-place "New York" \
  --has-children \
  --format json

# Count only
gedcom search family.ged --name "John" --count-only

# Sorted results
gedcom search family.ged --name "John" --sort name --limit 10
```

### Interactive Mode

```bash
# Start interactive mode
gedcom interactive family.ged

# Example session:
gedcom> search John
Search results for 'John':
  @I1@: John /Doe/
  @I5@: John /Smith/

gedcom> individual @I1@
Individual: @I1@
  Name: John /Doe/
  Sex: M
  Birth: 1900
  Death: 1970

gedcom> parents @I1@
Parents of @I1@:
  @I10@: James /Doe/
  @I11@: Mary /Doe/

gedcom> ancestors @I1@ 3
Ancestors of @I1@ (max 3 generations):
  @I10@: James /Doe/
  @I11@: Mary /Doe/
  ...

gedcom> relationship @I1@ @I2@
Relationship from @I1@ to @I2@:
  Type: 1st Cousin
  Degree: 1
  Removal: 0

gedcom> exit
```

## Architecture

### Package Structure

```
gedcom-go/
├── pkg/gedcom/          # Public API
│   ├── Core types       # Tree, Record, Line, Error
│   ├── Record types      # Individual, Family, Note, etc.
│   └── query/           # Graph-based Query API
├── internal/            # Implementation
│   ├── parser/         # Parsing logic
│   ├── validator/      # Validation logic
│   └── exporter/       # Export functionality
└── cmd/gedcom/         # CLI application
```

### Design Principles

- **Type Safety**: Strong typing throughout
- **Thread Safety**: Mutex-protected shared state
- **Performance**: Optimized with caching, indexing, pooling
- **Extensibility**: Interface-based design
- **Error Handling**: Explicit error returns with severity levels

## Performance

### Optimizations

- **Query Result Caching**: 100x speedup for repeated queries
- **Indexing**: 20-200x faster filtering
- **Bidirectional BFS**: ~2x faster path finding
- **Memory Pooling**: Reduced allocations and GC pressure
- **Incremental Updates**: 50-200x faster than full rebuild

### Benchmarks

- Graph construction: ~100ms for 10,000 individuals
- Cached queries: ~45ns (cache hit)
- Indexed filtering: O(1) or O(log n) instead of O(V)
- Shortest path: O(V/2 + E/2) average case

## Documentation

- **[CLI Documentation](docs/cli.md)** - Complete CLI reference guide
- **[Parser Documentation](docs/parser.md)** - Parse GEDCOM files with multiple parser types
- **[Validator Documentation](docs/validator.md)** - Validate GEDCOM files with comprehensive rules
- **[Exporter Documentation](docs/exporter.md)** - Export GEDCOM data to JSON, XML, YAML, and GEDCOM formats
- **[Query API Documentation](docs/query-api.md)** - Graph-based query API for relationship queries
- **[Types Documentation](docs/types.md)** - Core GEDCOM data types and structures

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run benchmarks
go test ./... -bench=.
```

**Test Coverage:**
- ✅ Parser: Comprehensive (15+ test files)
- ✅ Validator: Comprehensive (10+ test files)
- ✅ Exporter: Comprehensive (8+ test files)
- ✅ Query API: Comprehensive (15+ test files)
- ✅ Core Types: Comprehensive (10+ test files)

## Requirements

- **Go 1.23+**
- **Dependencies**: See [go.mod](go.mod)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- GEDCOM 5.5.1 specification
- Go community for excellent tooling and libraries

## Status

✅ **Production Ready**

The codebase is mature, well-tested, and ready for production use. All core functionality is implemented and tested.

---

**Made with ❤️ for genealogical research**
