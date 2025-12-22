# Codebase Refactoring Analysis

## Executive Summary

This analysis identifies large files, code organization issues, and refactoring opportunities in the `gedcom-go` codebase. The codebase is generally well-structured, but several files have grown large and would benefit from refactoring to improve maintainability.

## Large Files Requiring Attention

### 1. `pkg/gedcom/query/graph.go` - **1,225 lines** ⚠️ **HIGH PRIORITY**

**Current Responsibilities:**
- Graph structure definition (`Graph` struct with 20+ fields)
- Node management (GetIndividual, GetFamily, GetNote, GetSource, GetRepository, GetEvent)
- Edge management (AddEdge, RemoveEdge, AddEdgeIncremental, RemoveEdgeIncremental)
- Hybrid storage integration (get*FromHybrid, load*FromHybrid methods)
- Lazy loading logic (ensureEdgesLoaded, load*EdgesUnlocked)
- Graph metrics and analytics
- ID mapping (xrefToID, idToXref)
- Component detection
- Relationship calculations
- Cache management integration

**Issues:**
- **Too many responsibilities**: This file handles graph structure, storage, lazy loading, hybrid mode, caching, and queries
- **High complexity**: ~50+ public and private methods
- **Mixed concerns**: Storage logic, query logic, and graph management are all intertwined
- **Hard to test**: Large surface area makes unit testing difficult
- **Hard to maintain**: Changes to one concern can affect others

**Recommendations:**
1. **Extract Node Access Layer**: Create `graph_nodes.go` for all `Get*` methods
2. **Extract Edge Management**: Create `graph_edges.go` for edge operations
3. **Extract Hybrid Storage Logic**: Move all `*FromHybrid` methods to `graph_hybrid.go`
4. **Extract Lazy Loading**: Move lazy loading logic to `graph_lazy.go`
5. **Extract Graph Metrics**: Move metrics to `graph_metrics.go` (already exists but may need expansion)
6. **Keep Core Structure**: `graph.go` should only contain the `Graph` struct definition and core initialization

**Estimated Impact**: High - Would significantly improve maintainability and testability

---

### 2. `pkg/gedcom/query/hybrid_builder.go` - **1,060 lines** ⚠️ **MEDIUM PRIORITY**

**Current Responsibilities:**
- SQLite schema creation and initialization
- BadgerDB initialization
- Building graph in SQLite (indexes, metadata)
- Building graph in BadgerDB (nodes, edges, serialization)
- Edge building for all node types
- Date parsing and normalization
- Component detection and storage

**Issues:**
- **Mixed storage backends**: SQLite and BadgerDB logic in same file
- **Large functions**: `buildGraphInSQLite` and `buildGraphInBadgerDB` are very long
- **Repetitive patterns**: Similar code for different node types

**Recommendations:**
1. **Split by Storage Backend**:
   - `hybrid_sqlite_builder.go` - All SQLite operations
   - `hybrid_badger_builder.go` - All BadgerDB operations
2. **Extract Common Patterns**: Create helper functions for node type processing
3. **Separate Schema Management**: Move schema creation to `hybrid_schema.go`

**Estimated Impact**: Medium - Would improve clarity and make storage backends easier to maintain independently

---

### 3. `pkg/gedcom/query/filter_query.go` - **572 lines** ⚠️ **LOW PRIORITY**

**Current Responsibilities:**
- Filter query structure and execution
- All filter types (ByName, BySurname, ByGivenName, ByBirthDate, BySex, etc.)
- Index integration
- Hybrid storage query integration
- Result caching

**Issues:**
- **Many filter methods**: 20+ filter methods in one file
- **Mixed execution modes**: Eager, lazy, and hybrid execution logic

**Recommendations:**
1. **Group Related Filters**: 
   - `filter_name.go` - Name-based filters (ByName, BySurname, ByGivenName, etc.)
   - `filter_date.go` - Date-based filters (ByBirthDate, ByBirthYear, ByBirthMonth, etc.)
   - `filter_attributes.go` - Attribute filters (BySex, HasChildren, Living, etc.)
2. **Extract Execution Logic**: Move execution logic to `filter_execution.go`
3. **Keep Core**: `filter_query.go` should only contain the `FilterQuery` struct and core methods

**Estimated Impact**: Low-Medium - Would improve organization but current structure is acceptable

---

### 4. `stress_test.go` - **1,380 lines** ⚠️ **LOW PRIORITY**

**Current Responsibilities:**
- Multiple stress test functions (1M, 1.5M, 5M, 10M, lazy loading variants)
- Test data generation
- Performance measurement utilities
- Test phase execution
- Metrics collection and reporting

**Issues:**
- **Very large test file**: Contains many test functions and helpers
- **Mixed concerns**: Test functions, helpers, and data generation all in one file
- **Hard to navigate**: Finding specific tests is difficult

**Recommendations:**
1. **Split by Test Type**:
   - `stress_test_eager.go` - Eager loading stress tests
   - `stress_test_lazy.go` - Lazy loading stress tests
   - `stress_test_hybrid.go` - Hybrid storage stress tests
2. **Extract Test Helpers**: Create `stress_test_helpers.go` for shared utilities
3. **Extract Data Generation**: Create `stress_test_data.go` for data generation functions

**Estimated Impact**: Low - Test files can be large, but splitting would improve navigation

---

## Code Organization Analysis

### Package Structure: `pkg/gedcom/query`

**Current State:**
- 58 Go files total
- Good separation of concerns in most areas
- Some files are getting large but structure is logical

**Strengths:**
- ✅ Clear separation between query types (ancestor, descendant, relationship, path)
- ✅ Separate files for different node types
- ✅ Hybrid storage is well-isolated
- ✅ Caching is separated
- ✅ Serialization is separated

**Areas for Improvement:**
- ⚠️ `graph.go` is doing too much
- ⚠️ `hybrid_builder.go` mixes SQLite and BadgerDB concerns
- ⚠️ `filter_query.go` has many filter methods (but this is acceptable)

---

## Logical Consistency Check

### ✅ **Well-Organized Areas:**

1. **Query Types**: Each query type has its own file
   - `ancestor_query.go`, `descendant_query.go`, `relationship_query.go`, `path_query.go`
   - Clear separation of concerns

2. **Node Types**: Each node type is well-defined
   - `node.go` contains all node interfaces and implementations
   - Clear inheritance hierarchy

3. **Storage**: Hybrid storage is well-separated
   - `hybrid_storage.go` - Storage initialization
   - `hybrid_serialization.go` - Serialization logic
   - `hybrid_queries.go` - Query helpers
   - `hybrid_cache.go` - Caching layer

4. **New Query Files**: Recently added query files are well-organized
   - `notes_query.go` - Note queries
   - `events_query.go` - Event queries
   - `name_filters.go` - Name filtering
   - `analytics.go` - Analytics
   - `birthday_filters.go` - Birthday filtering

### ⚠️ **Areas Needing Attention:**

1. **`graph.go`**: Too many responsibilities
   - Node access, edge management, storage, lazy loading all mixed
   - **Recommendation**: Split into focused files

2. **`hybrid_builder.go`**: Mixed storage backends
   - SQLite and BadgerDB logic intertwined
   - **Recommendation**: Split by storage backend

3. **`filter_query.go`**: Many filter methods (acceptable but could be better organized)
   - All filters in one file
   - **Recommendation**: Group related filters into separate files

---

## Refactoring Priority

### **High Priority:**
1. **Split `graph.go`** - This is the most critical refactoring
   - Extract node access methods
   - Extract edge management
   - Extract hybrid storage methods
   - Extract lazy loading logic

### **Medium Priority:**
2. **Split `hybrid_builder.go`**
   - Separate SQLite and BadgerDB builders
   - Extract common patterns

### **Low Priority:**
3. **Organize `filter_query.go`**
   - Group related filters (optional, current structure is acceptable)

4. **Split `stress_test.go`**
   - Split by test type (optional, test files can be large)

---

## Code Quality Observations

### ✅ **Strengths:**
- Good use of interfaces (`GraphNode`, `Record`)
- Clear separation of query types
- Well-documented code
- Good test coverage
- Consistent naming conventions
- Proper error handling

### ⚠️ **Areas for Improvement:**
- Some files are getting large (but not unmanageable)
- `graph.go` has too many responsibilities
- Some repetitive patterns in `hybrid_builder.go`
- Test file is very large (but this is acceptable for stress tests)

---

## Recommendations Summary

### Immediate Actions (High Priority):
1. **Refactor `graph.go`**:
   - Create `graph_nodes.go` for all `Get*` methods
   - Create `graph_edges.go` for edge operations
   - Create `graph_hybrid.go` for hybrid storage methods
   - Create `graph_lazy.go` for lazy loading logic
   - Keep only core structure in `graph.go`

### Short-term Actions (Medium Priority):
2. **Refactor `hybrid_builder.go`**:
   - Split into `hybrid_sqlite_builder.go` and `hybrid_badger_builder.go`
   - Extract common patterns

### Long-term Actions (Low Priority):
3. **Organize filter queries** (optional)
4. **Split stress tests** (optional)

---

## Metrics

| File | Lines | Functions | Types | Priority |
|------|-------|-----------|-------|----------|
| `graph.go` | 1,225 | ~50+ | ~10 | **HIGH** |
| `hybrid_builder.go` | 1,060 | ~15 | ~5 | **MEDIUM** |
| `filter_query.go` | 572 | ~30 | ~3 | **LOW** |
| `stress_test.go` | 1,380 | ~20 | ~5 | **LOW** |

---

## Conclusion

The codebase is **generally well-organized and logical**. The main issue is that `graph.go` has grown too large and handles too many responsibilities. Refactoring it would significantly improve maintainability.

The other large files (`hybrid_builder.go`, `filter_query.go`, `stress_test.go`) are acceptable in size, though they could benefit from some organization improvements.

**Overall Assessment**: The codebase is in good shape, with one high-priority refactoring target (`graph.go`) and a few medium/low-priority improvements.

