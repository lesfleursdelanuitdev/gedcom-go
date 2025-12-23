# Hybrid Storage Implementation Plan

## Overview

This document outlines the step-by-step plan for implementing the hybrid storage approach (SQLite + BadgerDB) to enable handling 10M+ individuals.

## Implementation Phases

### Phase 1: Foundation (Week 1)

**Goal:** Set up databases and basic infrastructure

1. **Add Dependencies**
   - Add SQLite driver: `github.com/mattn/go-sqlite3`
   - Add BadgerDB: `github.com/dgraph-io/badger/v4`
   - Add LRU cache: `github.com/hashicorp/golang-lru/v2`

2. **Database Initialization**
   - Create SQLite schema (tables, indexes, FTS5)
   - Create BadgerDB database structure
   - Add database connection management

3. **Serialization**
   - Implement node serialization (gob/json)
   - Implement edge serialization
   - Implement component serialization

**Deliverables:**
- Database schemas created
- Basic serialization working
- Unit tests for database operations

---

### Phase 2: Graph Construction (Week 2)

**Goal:** Build graph and store in both databases

1. **SQLite Index Building**
   - Insert nodes into `nodes` table
   - Build FTS5 index (via triggers)
   - Build xref mapping
   - Build component mapping

2. **BadgerDB Graph Storage**
   - Store nodes in BadgerDB
   - Store edges in BadgerDB
   - Store components in BadgerDB

3. **Hybrid Graph Builder**
   - Create `BuildGraphHybrid()` function
   - Coordinate SQLite and BadgerDB writes
   - Handle errors and rollbacks

**Deliverables:**
- Graph construction working
- Data stored in both databases
- Integration tests passing

---

### Phase 3: Query API Integration (Week 3)

**Goal:** Update query API to use hybrid storage

1. **FilterQuery Updates**
   - Query SQLite for node IDs
   - Load nodes from BadgerDB
   - Return results

2. **Graph Query Updates**
   - Get node ID from SQLite (xref → ID)
   - Load node/edges from BadgerDB
   - Return results

3. **Relationship Query Updates**
   - Use BadgerDB for graph traversal
   - Use SQLite for metadata lookups

**Deliverables:**
- All query types working
- Performance benchmarks
- Integration tests

---

### Phase 4: Caching and Optimization (Week 4)

**Goal:** Add caching and optimize performance

1. **LRU Cache Implementation**
   - Node cache (10K-100K nodes)
   - XREF cache (10K-50K entries)
   - Query result cache (1K-10K queries)

2. **Performance Optimizations**
   - Batch operations
   - Prepared statements
   - Connection pooling
   - Memory-mapped I/O

3. **Memory Management**
   - Cache eviction policies
   - Memory budget management
   - GC optimization

**Deliverables:**
- Caching working
- Performance optimized
- Memory usage within budget

---

### Phase 5: Testing and Validation (Week 5)

**Goal:** Comprehensive testing and validation

1. **Unit Tests**
   - Database operations
   - Serialization
   - Query operations

2. **Integration Tests**
   - End-to-end workflows
   - Large dataset tests (1M, 5M, 10M)
   - Concurrent access tests

3. **Performance Tests**
   - Memory usage benchmarks
   - Query performance benchmarks
   - Scalability tests

**Deliverables:**
- All tests passing
- Performance benchmarks documented
- Ready for production

---

## File Structure

```
pkg/gedcom/query/
├── graph.go                 # Existing graph (keep for backward compatibility)
├── hybrid_graph.go          # New hybrid graph implementation
├── hybrid_builder.go        # Graph construction for hybrid mode
├── hybrid_query.go          # Query operations for hybrid mode
├── hybrid_storage.go        # Database operations
│   ├── sqlite_storage.go    # SQLite operations
│   └── badger_storage.go    # BadgerDB operations
├── hybrid_cache.go          # Caching layer
└── hybrid_serialization.go  # Serialization/deserialization
```

---

## Key Design Decisions

### 1. Backward Compatibility

- Keep existing in-memory graph as default
- Add hybrid mode as opt-in feature
- Gradual migration path

### 2. Data Consistency

- Use transactions for multi-step operations
- Validate data before storing
- Handle errors gracefully

### 3. Performance

- Cache hot data in RAM
- Use memory-mapped I/O
- Batch operations when possible

### 4. Memory Management

- LRU cache with configurable size
- Memory budget management
- GC optimization

---

## Success Criteria

### Functional

- ✅ All existing queries work with hybrid storage
- ✅ Performance matches or exceeds in-memory for hot data
- ✅ Handles 10M+ individuals without OOM

### Performance

- ✅ Memory usage: 2-5 GB for 10M individuals
- ✅ Query performance: <100µs for indexed queries
- ✅ Graph traversal: <100µs per hop

### Quality

- ✅ All tests passing
- ✅ No memory leaks
- ✅ Error handling robust

---

## Risks and Mitigations

### Risk 1: CGO Dependency (SQLite)

**Mitigation:**
- Use `github.com/mattn/go-sqlite3` (well-maintained)
- Consider pure Go alternative if issues arise
- Test on multiple platforms

### Risk 2: Database Corruption

**Mitigation:**
- Regular backups
- Transaction management
- Data validation
- Recovery procedures

### Risk 3: Performance Degradation

**Mitigation:**
- Comprehensive benchmarking
- Performance monitoring
- Cache optimization
- Query optimization

### Risk 4: Complexity

**Mitigation:**
- Clear separation of concerns
- Comprehensive documentation
- Unit tests for each component
- Gradual rollout

---

## Timeline

- **Week 1:** Foundation
- **Week 2:** Graph Construction
- **Week 3:** Query API Integration
- **Week 4:** Caching and Optimization
- **Week 5:** Testing and Validation

**Total:** 5 weeks for full implementation

---

## Next Steps

1. Review and approve design
2. Set up development environment
3. Create feature branch
4. Begin Phase 1 implementation

