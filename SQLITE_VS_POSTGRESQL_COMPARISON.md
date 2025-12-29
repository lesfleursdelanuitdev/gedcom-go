# SQLite vs PostgreSQL Hybrid Storage Comparison

**Date:** 2025-01-27  
**Purpose:** Comprehensive comparison of SQLite and PostgreSQL hybrid storage modes in `gedcom-go`

---

## Executive Summary

Both SQLite and PostgreSQL hybrid storage modes provide the same functionality, but with different performance characteristics and use cases:

- **SQLite**: Faster for single-file, single-user scenarios. Better for development and small deployments.
- **PostgreSQL**: Better for multi-file, multi-user, and production deployments. Supports concurrent access and shared databases.

---

## Architecture Comparison

### SQLite Hybrid Mode
- **Storage**: SQLite database file + BadgerDB directory (per GEDCOM file)
- **Isolation**: Each GEDCOM file has its own SQLite database
- **Concurrency**: Single writer, multiple readers per file
- **Deployment**: File-based, easy to backup/copy

### PostgreSQL Hybrid Mode
- **Storage**: Shared PostgreSQL database + BadgerDB directory (per GEDCOM file)
- **Isolation**: All files share one database, isolated by `file_id` column
- **Concurrency**: Full PostgreSQL concurrency (multiple writers/readers)
- **Deployment**: Server-based, centralized management

---

## Performance Benchmarks

### Test Environment
- **Test File**: `xavier.ged` (312 individuals, 107 families)
- **Hardware**: Standard development machine
- **Database**: PostgreSQL 14+ on localhost

### Build Time Performance

| Metric | SQLite | PostgreSQL | Ratio |
|--------|--------|------------|-------|
| **Graph Build Time** | 57.3 ms | 386.3 ms | 6.74x slower |
| **Per Individual** | 0.18 ms | 1.24 ms | 6.89x slower |
| **Per Family** | 0.54 ms | 3.61 ms | 6.69x slower |

**Analysis:**
- SQLite is significantly faster for initial graph building
- PostgreSQL overhead comes from:
  - Network latency (even on localhost)
  - Connection pool setup
  - Transaction management
  - Schema validation

### Query Performance

#### FindByXref (100 iterations)

| Metric | SQLite | PostgreSQL | Ratio |
|--------|--------|------------|-------|
| **Total Time** | ~1-2 ms | ~2-3 ms | ~1.5x slower |
| **Per Operation** | ~10-20 μs | ~20-30 μs | ~1.5x slower |

**Analysis:**
- Both are very fast for simple lookups
- PostgreSQL has slight overhead from network round-trip
- Difference is negligible for most use cases

#### FilterQuery Performance (10 iterations)

| Metric | SQLite | PostgreSQL | Ratio |
|--------|--------|------------|-------|
| **Total Time** | 82.0 ms | 1,259.0 ms | 15.3x slower |
| **Per Operation** | 8.20 ms | 125.90 ms | 15.3x slower |

**Analysis:**
- PostgreSQL is significantly slower for complex queries
- Likely due to:
  - Network latency for multiple queries
  - Prepared statement overhead
  - Connection pool management
  - Query planning overhead

**Note:** This is a worst-case scenario. Performance improves with:
- Connection pooling optimization
- Query result caching
- Prepared statement reuse
- Network optimization (local vs remote)

---

## Feature Comparison

| Feature | SQLite | PostgreSQL | Notes |
|---------|--------|------------|-------|
| **Full-Text Search** | FTS5 | GIN indexes + pg_trgm | Both support fast text search |
| **Indexes** | B-tree | B-tree, GIN, GiST | PostgreSQL has more index types |
| **Concurrent Writes** | Limited | Full support | PostgreSQL handles concurrent access better |
| **Multi-File Support** | Separate DBs | Shared DB with `file_id` | PostgreSQL allows shared database |
| **Transactions** | File-level | Row-level | PostgreSQL has better transaction isolation |
| **Backup** | File copy | pg_dump | Both have good backup options |
| **Replication** | None | Built-in | PostgreSQL supports replication |
| **Connection Pooling** | Limited | Full support | PostgreSQL has better pooling |
| **Prepared Statements** | Supported | Supported | Both use prepared statements |
| **Schema Migration** | Manual | Migrations | PostgreSQL has better migration tools |

---

## Use Case Recommendations

### Choose SQLite When:
1. **Single-file deployments**: One GEDCOM file per application instance
2. **Development/Testing**: Fast iteration, easy setup
3. **Small to medium datasets**: < 100K individuals per file
4. **Single-user applications**: Desktop apps, CLI tools
5. **Embedded deployments**: Applications that need to be self-contained
6. **Performance-critical builds**: When build time matters more than query time

### Choose PostgreSQL When:
1. **Multi-file deployments**: Multiple GEDCOM files in one system
2. **Production environments**: High availability, reliability requirements
3. **Concurrent access**: Multiple users/applications accessing same data
4. **Large datasets**: > 100K individuals, need for scalability
5. **Shared infrastructure**: Already have PostgreSQL server
6. **Advanced features needed**: Full-text search, complex queries, replication
7. **REST API backends**: Multiple API instances sharing database

---

## Data Consistency

✅ **Both modes return identical results**

Tested with `xavier.ged`:
- Both returned **5 results** for `ByName("xavier")` query
- Data integrity verified across all query types
- No discrepancies found in test suite

---

## Storage Requirements

### SQLite
- **Per File**: 1 SQLite database file + 1 BadgerDB directory
- **Size**: ~1-5 MB per 1K individuals (depends on data complexity)
- **Backup**: Copy files directly

### PostgreSQL
- **Shared**: 1 PostgreSQL database for all files + 1 BadgerDB directory per file
- **Size**: Similar per-individual, but shared overhead
- **Backup**: Use `pg_dump` or continuous archiving

---

## Migration Path

### SQLite → PostgreSQL
1. Export data from SQLite (if needed)
2. Create PostgreSQL database
3. Build graph with PostgreSQL mode using same `file_id`
4. Verify data consistency
5. Switch application to use PostgreSQL mode

### PostgreSQL → SQLite
1. Export data from PostgreSQL (if needed)
2. Create SQLite database file
3. Build graph with SQLite mode
4. Verify data consistency
5. Switch application to use SQLite mode

**Note:** Both modes use the same BadgerDB format, so graph structure is compatible.

---

## Code Compatibility

✅ **100% API Compatible**

Both modes use the same:
- `Graph` interface
- `FilterQuery` API
- `Query` builders
- Relationship methods
- All query helpers

No code changes needed when switching between modes.

---

## Performance Optimization Tips

### SQLite
- Use WAL mode for better concurrency
- Tune `PRAGMA` settings for your workload
- Consider connection pooling for multiple queries
- Use prepared statements (already implemented)

### PostgreSQL
- Tune connection pool settings (`MaxOpenConns`, `MaxIdleConns`)
- Use connection pooling (already implemented)
- Optimize indexes for your query patterns
- Consider read replicas for read-heavy workloads
- Use prepared statements (already implemented)
- Consider local vs remote database placement

---

## Test Results Summary

### Build Performance
```
SQLite:    57.3 ms  (312 individuals, 107 families)
PostgreSQL: 386.3 ms (312 individuals, 107 families)
Ratio:     6.74x slower for PostgreSQL
```

### Query Performance
```
FindByXref (100 ops):
  SQLite:    ~1-2 ms total (~10-20 μs/op)
  PostgreSQL: ~2-3 ms total (~20-30 μs/op)
  Ratio:     ~1.5x slower for PostgreSQL

FilterQuery (10 ops):
  SQLite:    82.0 ms total (8.20 ms/op)
  PostgreSQL: 1,259.0 ms total (125.90 ms/op)
  Ratio:     15.3x slower for PostgreSQL
```

### Data Consistency
```
✅ Both modes return identical results
✅ All test cases pass
✅ No data discrepancies found
```

---

## Recommendations

### For MVP / Development
**Use SQLite** - Faster, simpler, easier to set up

### For Production / Multi-User
**Use PostgreSQL** - Better concurrency, scalability, reliability

### For Hybrid Approach
Consider using:
- **SQLite** for development/testing
- **PostgreSQL** for production
- Same codebase, just different configuration

---

## Future Optimizations

### PostgreSQL Performance Improvements
1. **Connection Pooling**: Already implemented, can be tuned further
2. **Query Batching**: Batch multiple queries into single transactions
3. **Local Database**: Use local PostgreSQL to reduce network latency
4. **Index Optimization**: Add composite indexes for common query patterns
5. **Materialized Views**: Pre-compute common query results
6. **Read Replicas**: Use read replicas for query-heavy workloads

### SQLite Performance Improvements
1. **WAL Mode**: Already supported, can be enabled
2. **Query Optimization**: Further optimize prepared statements
3. **Cache Tuning**: Adjust SQLite cache size for workload

---

## Conclusion

Both SQLite and PostgreSQL hybrid storage modes are production-ready and provide identical functionality. The choice depends on your specific requirements:

- **Performance**: SQLite is faster for single-file scenarios
- **Scalability**: PostgreSQL is better for multi-file, multi-user scenarios
- **Simplicity**: SQLite is easier to set up and deploy
- **Features**: PostgreSQL offers more advanced database features

**Recommendation**: Start with SQLite for development and MVP, migrate to PostgreSQL when you need multi-file support, concurrent access, or production-grade reliability.

---

## Running Comparison Tests

```bash
# Set up PostgreSQL connection
export DATABASE_URL="postgresql://user:password@localhost:5432/database?sslmode=disable"

# Run comparison test
go test ./query -run TestHybridStorageComparison -v

# Run benchmarks
go test ./query -bench BenchmarkHybridStorageComparison -benchmem
go test ./query -bench BenchmarkHybridQueryComparison -benchmem
```

---

## References

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [BadgerDB Documentation](https://dgraph.io/docs/badger/)
- [Hybrid Storage Architecture](./HYBRID_POSTGRESQL_ANALYSIS.md)
- [PostgreSQL Testing Summary](./POSTGRESQL_TESTING_SUMMARY.md)

