# Codebase Refactoring Analysis

## Executive Summary

This analysis identifies large files, code organization issues, and refactoring opportunities in the `gedcom-go` codebase. The codebase is generally well-structured, and **major refactoring has been completed**:

✅ **Completed Refactoring:**
- `graph.go`: Refactored from 1,225 lines → 114 lines (split into 6 files)
- `hybrid_builder.go`: Refactored from 1,060 lines → 64 lines (split into 4 files)
- `filter_query.go`: Refactored from 572 lines → 52 lines (split into 5 files)

The codebase now has better separation of concerns and improved maintainability.

## Refactored Files

### 1. `pkg/gedcom/query/graph.go` - **114 lines** ✅ **REFACTORED**

**Status**: ✅ **COMPLETED** - Successfully refactored into multiple focused files

**Original State:**
- Was 1,225 lines with too many responsibilities
- Mixed concerns: structure, storage, lazy loading, hybrid mode, caching, queries

**Refactored Structure:**
1. ✅ **`graph.go`** (114 lines) - Core `Graph` struct definition and initialization only
2. ✅ **`graph_nodes.go`** (298 lines) - All `Get*` methods (GetIndividual, GetFamily, GetNote, etc.)
3. ✅ **`graph_edges.go`** (101 lines) - Edge operations (AddEdge, RemoveEdge, etc.)
4. ✅ **`graph_hybrid.go`** (387 lines) - Hybrid storage integration (get*FromHybrid, load*FromHybrid)
5. ✅ **`graph_hybrid_helpers.go`** (178 lines) - Hybrid helper functions
6. ✅ **`graph_metrics.go`** (421 lines) - Graph metrics and analytics

**Result:**
- **Total**: ~1,499 lines across 6 well-organized files
- **Improved maintainability**: Each file has a single, clear responsibility
- **Better testability**: Smaller, focused files are easier to test
- **Clear separation of concerns**: Structure, nodes, edges, hybrid, and metrics are separated

---

### 2. `pkg/gedcom/query/hybrid_builder.go` - **64 lines** ✅ **REFACTORED**

**Status**: ✅ **COMPLETED** - Successfully refactored into multiple focused files

**Original State:**
- Was 1,060 lines with mixed SQLite and BadgerDB logic
- Large functions and repetitive patterns

**Refactored Structure:**
1. ✅ **`hybrid_builder.go`** (64 lines) - Coordinator function only
2. ✅ **`hybrid_sqlite_builder.go`** - All SQLite operations (indexes, metadata)
3. ✅ **`hybrid_badger_builder.go`** - All BadgerDB operations (nodes, edges, serialization)
4. ✅ **`hybrid_builder_helpers.go`** - Common helper functions (date parsing, normalization)

**Result:**
- **Improved clarity**: Clear separation between SQLite and BadgerDB operations
- **Better maintainability**: Each storage backend can be maintained independently
- **Reduced complexity**: Smaller, focused files with single responsibilities

---

### 3. `pkg/gedcom/query/filter_query.go` - **52 lines** ✅ **REFACTORED**

**Status**: ✅ **COMPLETED** - Successfully refactored into multiple focused files

**Original State:**
- Was 572 lines with all filter types and execution logic mixed together

**Refactored Structure:**
1. ✅ **`filter_query.go`** (52 lines) - Core FilterQuery struct and basic methods
2. ✅ **`filter_name.go`** - Name-based filters (ByName, BySurname, ByGivenName, etc.)
3. ✅ **`filter_date.go`** - Date-based filters (ByBirthDate, ByBirthYear, etc.)
4. ✅ **`filter_attributes.go`** - Attribute filters (BySex, HasChildren, HasSpouse, etc.)
5. ✅ **`filter_execution.go`** - Execution logic (Execute, executeEager, executeHybrid)

**Result:**
- **Better organization**: Filters grouped by type
- **Improved maintainability**: Each filter category in its own file
- **Clear separation**: Execution logic separated from filter definitions
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
| `graph.go` | 114 | ~5 | ~2 | **COMPLETED** ✅ |
| `graph_nodes.go` | 298 | ~15 | ~3 | - |
| `graph_edges.go` | 101 | ~8 | ~2 | - |
| `graph_hybrid.go` | 387 | ~20 | ~5 | - |
| `graph_hybrid_helpers.go` | 178 | ~10 | ~2 | - |
| `graph_metrics.go` | 421 | ~25 | ~5 | - |
| `hybrid_builder.go` | 64 | ~2 | ~1 | **COMPLETED** ✅ |
| `hybrid_sqlite_builder.go` | 490 | ~15 | ~3 | - |
| `hybrid_badger_builder.go` | 806 | ~20 | ~4 | - |
| `hybrid_builder_helpers.go` | 37 | ~3 | ~1 | - |
| `filter_query.go` | 52 | ~3 | ~1 | **COMPLETED** ✅ |
| `filter_name.go` | 44 | ~5 | ~1 | - |
| `filter_date.go` | 45 | ~5 | ~1 | - |
| `filter_attributes.go` | 84 | ~6 | ~1 | - |
| `filter_execution.go` | 379 | ~15 | ~2 | - |
| `stress_test.go` | 1,380 | ~20 | ~5 | **LOW** |

---

## Conclusion

The codebase is **generally well-organized and logical**. The main issue is that `graph.go` has grown too large and handles too many responsibilities. Refactoring it would significantly improve maintainability.

The other large files (`hybrid_builder.go`, `filter_query.go`, `stress_test.go`) are acceptable in size, though they could benefit from some organization improvements.

**Overall Assessment**: The codebase is in good shape, with one high-priority refactoring target (`graph.go`) and a few medium/low-priority improvements.

