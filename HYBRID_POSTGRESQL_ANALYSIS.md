# Hybrid Mode: SQLite vs PostgreSQL Analysis

**Date:** 2025-01-27  
**Purpose:** Analyze replacing SQLite with PostgreSQL in hybrid mode (BadgerDB + SQLite → BadgerDB + PostgreSQL)

---

## Executive Summary

**Current Architecture:** BadgerDB (graph structure) + SQLite (indexed metadata) - **per-file**

**Proposed Architecture:** BadgerDB (graph structure) + PostgreSQL (indexed metadata) - **two options**

**Key Decision:** **Per-file PostgreSQL databases** vs **Shared PostgreSQL database with file_id**

**Recommendation:** 
- **Keep SQLite for MVP** (simpler, sufficient)
- **Consider PostgreSQL for scale** (better concurrent access, advanced features)
- **If using PostgreSQL, use shared database with file_id** (better resource utilization)

---

## 1. Current Hybrid Mode Architecture

### 1.1 How It Works Now

**Per-File Storage:**
```
/var/lib/ligneous/files/{file_id}/
├── original.ged              # Original GEDCOM file
├── indexes.db                # SQLite database (per-file)
└── graph/                    # BadgerDB directory (per-file)
    ├── MANIFEST
    ├── *.sst
    └── *.vlog
```

**SQLite Role:**
- Stores indexed metadata per file (name, birth_date, birth_place, sex, etc.)
- Fast filtering queries (by name, date, place, sex)
- Full-text search (FTS5)
- Prepared statements for performance
- WAL mode for concurrent reads

**BadgerDB Role:**
- Stores complete graph structure (nodes, edges)
- Key-value storage optimized for graph traversal
- Serialized node data
- Memory-mapped I/O

**Key Characteristic:** **Each file is completely isolated** - separate SQLite DB and BadgerDB directory.

---

## 2. PostgreSQL Options

### 2.1 Option A: Per-File PostgreSQL Databases

**Architecture:**
```
PostgreSQL Server:
├── Database: file_abc123     # One database per file
│   ├── nodes table
│   ├── xref_mapping table
│   └── components table
├── Database: file_def456
│   ├── nodes table
│   ├── xref_mapping table
│   └── components table
└── ...

BadgerDB (unchanged):
/var/lib/ligneous/files/{file_id}/graph/
```

**Pros:**
- ✅ Maintains per-file isolation (like SQLite)
- ✅ Easy to drop entire file (DROP DATABASE)
- ✅ No file_id column needed in queries
- ✅ Similar to current SQLite approach

**Cons:**
- ❌ **PostgreSQL overhead per database** (connection pools, metadata, etc.)
- ❌ **Resource intensive** (each database has overhead)
- ❌ **Connection management complexity** (need to connect to different databases)
- ❌ **Not scalable** (PostgreSQL has limits on number of databases)
- ❌ **Backup complexity** (need to backup each database separately)
- ❌ **Wasteful** (PostgreSQL databases have significant overhead)

**Verdict:** ❌ **Not recommended** - defeats the purpose of using PostgreSQL

---

### 2.2 Option B: Shared PostgreSQL Database with file_id

**Architecture:**
```
PostgreSQL Server:
└── Database: ligneous_graphs
    └── nodes table
        ├── file_id (partition key)
        ├── id (node ID within file)
        ├── xref
        ├── type
        ├── name
        └── ...
    └── xref_mapping table
        ├── file_id
        ├── xref
        └── node_id
    └── components table
        ├── file_id
        ├── component_id
        └── node_id

BadgerDB (unchanged):
/var/lib/ligneous/files/{file_id}/graph/
```

**Pros:**
- ✅ **Efficient resource utilization** (one database, shared connection pool)
- ✅ **Better for concurrent access** (PostgreSQL excels here)
- ✅ **Easier backup** (single database)
- ✅ **Better query optimization** (PostgreSQL can optimize across files)
- ✅ **Cross-file queries possible** (if needed in future)
- ✅ **Better monitoring** (single database to monitor)
- ✅ **Partitioning support** (can partition by file_id for performance)

**Cons:**
- ⚠️ **Must include file_id in all queries** (adds complexity)
- ⚠️ **Schema changes** (add file_id column everywhere)
- ⚠️ **Slightly more complex** (need to filter by file_id)
- ⚠️ **Network overhead** (but minimal with connection pooling)

**Verdict:** ✅ **Recommended if using PostgreSQL** - this is the proper way

---

## 3. Detailed Comparison

### 3.1 Performance

| Aspect | SQLite (Current) | PostgreSQL (Per-DB) | PostgreSQL (Shared) |
|--------|------------------|---------------------|---------------------|
| **Query Speed** | ✅ Very Fast | ✅ Fast | ✅ Fast |
| **Index Performance** | ✅ Excellent | ✅ Excellent | ✅ Excellent |
| **Concurrent Reads** | ✅ Good (WAL) | ✅ Excellent | ✅ Excellent |
| **Concurrent Writes** | ⚠️ Limited | ✅ Excellent | ✅ Excellent |
| **Network Overhead** | ✅ None (embedded) | ⚠️ Minimal | ⚠️ Minimal |
| **Connection Overhead** | ✅ None | ❌ High (per DB) | ✅ Low (pooled) |
| **Memory Usage** | ✅ Low | ❌ High (per DB) | ✅ Medium |

**Winner:** 
- **Single-file queries:** SQLite ≈ PostgreSQL (Shared) > PostgreSQL (Per-DB)
- **Concurrent access:** PostgreSQL (Shared) > PostgreSQL (Per-DB) > SQLite
- **Resource efficiency:** SQLite > PostgreSQL (Shared) > PostgreSQL (Per-DB)

---

### 3.2 Scalability

| Aspect | SQLite (Current) | PostgreSQL (Per-DB) | PostgreSQL (Shared) |
|--------|------------------|---------------------|---------------------|
| **Number of Files** | ✅ Unlimited | ❌ Limited (~1000 DBs) | ✅ Unlimited |
| **Concurrent Users** | ⚠️ Limited | ✅ Good | ✅ Excellent |
| **Write Throughput** | ⚠️ Limited | ✅ Good | ✅ Excellent |
| **Connection Pooling** | ✅ N/A (embedded) | ❌ Complex | ✅ Simple |
| **Resource Growth** | ✅ Linear | ❌ Exponential | ✅ Linear |

**Winner:** **PostgreSQL (Shared) > SQLite > PostgreSQL (Per-DB)**

---

### 3.3 Deployment Complexity

| Aspect | SQLite (Current) | PostgreSQL (Per-DB) | PostgreSQL (Shared) |
|--------|------------------|---------------------|---------------------|
| **Setup** | ✅ None (embedded) | ⚠️ PostgreSQL server | ⚠️ PostgreSQL server |
| **Configuration** | ✅ Minimal | ❌ Complex (per DB) | ✅ Simple |
| **Connection Management** | ✅ Simple | ❌ Complex | ✅ Simple (pool) |
| **Backup** | ✅ File copy | ❌ Per-database | ✅ Single database |
| **Monitoring** | ✅ Simple | ❌ Complex | ✅ Simple |
| **Dependencies** | ✅ None | ⚠️ PostgreSQL server | ⚠️ PostgreSQL server |

**Winner:** **SQLite > PostgreSQL (Shared) > PostgreSQL (Per-DB)**

---

### 3.4 Code Complexity

| Aspect | SQLite (Current) | PostgreSQL (Per-DB) | PostgreSQL (Shared) |
|--------|------------------|---------------------|---------------------|
| **Query Changes** | ✅ None | ✅ Minimal | ⚠️ Add file_id filter |
| **Connection Handling** | ✅ Simple | ❌ Complex | ✅ Simple (pool) |
| **Schema Changes** | ✅ None | ✅ None | ⚠️ Add file_id column |
| **Error Handling** | ✅ Simple | ⚠️ More complex | ⚠️ More complex |
| **Testing** | ✅ Simple | ⚠️ More complex | ⚠️ More complex |

**Winner:** **SQLite > PostgreSQL (Shared) > PostgreSQL (Per-DB)**

---

## 4. Schema Changes Required

### 4.1 Current SQLite Schema (Per-File)

```sql
-- nodes table (one per file)
CREATE TABLE nodes (
    id INTEGER PRIMARY KEY,
    xref TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    name TEXT,
    name_lower TEXT,
    birth_date INTEGER,
    birth_place TEXT,
    sex TEXT,
    has_children INTEGER DEFAULT 0,
    has_spouse INTEGER DEFAULT 0,
    living INTEGER DEFAULT 0,
    created_at INTEGER,
    updated_at INTEGER
);
```

**Queries:**
```sql
SELECT id FROM nodes WHERE xref = ?;
SELECT id FROM nodes WHERE name_lower LIKE ?;
```

---

### 4.2 PostgreSQL Schema (Shared Database)

```sql
-- nodes table (shared, with file_id)
CREATE TABLE nodes (
    file_id TEXT NOT NULL,
    id INTEGER NOT NULL,
    xref TEXT NOT NULL,
    type TEXT NOT NULL,
    name TEXT,
    name_lower TEXT,
    birth_date INTEGER,
    birth_place TEXT,
    sex TEXT,
    has_children INTEGER DEFAULT 0,
    has_spouse INTEGER DEFAULT 0,
    living INTEGER DEFAULT 0,
    created_at INTEGER,
    updated_at INTEGER,
    PRIMARY KEY (file_id, id),
    UNIQUE (file_id, xref)
);

-- Indexes
CREATE INDEX idx_nodes_file_id ON nodes(file_id);
CREATE INDEX idx_nodes_xref ON nodes(file_id, xref);
CREATE INDEX idx_nodes_name_lower ON nodes(file_id, name_lower);
CREATE INDEX idx_nodes_birth_date ON nodes(file_id, birth_date);
-- ... (other indexes include file_id)
```

**Queries:**
```sql
SELECT id FROM nodes WHERE file_id = ? AND xref = ?;
SELECT id FROM nodes WHERE file_id = ? AND name_lower LIKE ?;
```

**Key Change:** **All queries must include `file_id` filter**

---

### 4.3 Code Changes Required

**Current Code:**
```go
// HybridQueryHelpers
func (h *HybridQueryHelpers) FindByXref(xref string) (uint32, error) {
    var nodeID uint32
    err := h.stmtFindByXref.QueryRow(xref).Scan(&nodeID)
    // ...
}
```

**PostgreSQL Code (Shared):**
```go
// HybridQueryHelpers
func (h *HybridQueryHelpers) FindByXref(fileID string, xref string) (uint32, error) {
    var nodeID uint32
    err := h.stmtFindByXref.QueryRow(fileID, xref).Scan(&nodeID)
    // ...
}
```

**Impact:** 
- ✅ **Moderate** - need to pass `fileID` to all query methods
- ✅ **Straightforward** - just add parameter
- ⚠️ **All call sites need updating** - but manageable

---

## 5. Use Case Analysis

### 5.1 Single-Server, Low to Medium Traffic

**Scenario:**
- 1-10 concurrent users
- 10-100 files
- Mostly read operations
- Occasional file uploads

**SQLite (Current):**
- ✅ **Perfect fit**
- ✅ Simple deployment
- ✅ Fast enough
- ✅ Low resource usage

**PostgreSQL (Shared):**
- ⚠️ **Overkill**
- ⚠️ More complex setup
- ✅ Better concurrent access (but not needed)
- ⚠️ Higher resource usage

**Verdict:** **SQLite is better** ✅

---

### 5.2 Single-Server, High Traffic

**Scenario:**
- 50-200 concurrent users
- 100-1000 files
- Many concurrent reads
- Frequent file uploads

**SQLite (Current):**
- ⚠️ **May struggle with concurrent writes**
- ⚠️ WAL mode helps, but limited
- ✅ Still fast for reads

**PostgreSQL (Shared):**
- ✅ **Better for concurrent access**
- ✅ Handles writes better
- ✅ Better connection pooling
- ⚠️ Network overhead (minimal)

**Verdict:** **PostgreSQL (Shared) is better** ✅

---

### 5.3 Multi-Server Deployment

**Scenario:**
- Multiple API servers
- Shared storage (NFS/S3)
- Need shared database

**SQLite (Current):**
- ❌ **Not suitable** (file-based, not shared)
- ❌ Each server would have separate SQLite
- ❌ No way to share indexes

**PostgreSQL (Shared):**
- ✅ **Required**
- ✅ Shared database accessible from all servers
- ✅ Consistent indexes across servers
- ✅ Better for distributed systems

**Verdict:** **PostgreSQL (Shared) is required** ✅

---

## 6. Migration Path

### 6.1 From SQLite to PostgreSQL (Shared)

**Phase 1: Add PostgreSQL Support (Parallel)**
```
1. Add PostgreSQL driver support
2. Create new HybridStoragePostgres type
3. Implement same interface as HybridStorageSQLite
4. Add file_id to all queries
5. Keep SQLite as default
```

**Phase 2: Migrate Existing Files**
```
1. For each file:
   - Read from SQLite
   - Insert into PostgreSQL (with file_id)
   - Verify data
2. Keep SQLite as backup
3. Switch to PostgreSQL
```

**Phase 3: New Files Use PostgreSQL**
```
1. New files automatically use PostgreSQL
2. Old files can stay on SQLite (or migrate)
3. Eventually migrate all to PostgreSQL
```

**Complexity:** **Medium** - requires code changes but straightforward

---

### 6.2 Backward Compatibility

**Option 1: Support Both**
- Keep SQLite support
- Add PostgreSQL support
- Choose at runtime (config)
- **Pros:** Flexible, gradual migration
- **Cons:** More code to maintain

**Option 2: Replace SQLite**
- Remove SQLite support
- Migrate all files to PostgreSQL
- **Pros:** Simpler codebase
- **Cons:** Breaking change, migration required

**Recommendation:** **Support both initially, migrate gradually**

---

## 7. Cost-Benefit Analysis

### 7.1 Development Cost

| Task | SQLite (Current) | PostgreSQL (Shared) |
|------|------------------|---------------------|
| **Initial Setup** | ✅ Done | ⚠️ 2-3 days |
| **Schema Changes** | ✅ None | ⚠️ 1 day |
| **Query Updates** | ✅ None | ⚠️ 1-2 days |
| **Testing** | ✅ Done | ⚠️ 2-3 days |
| **Migration Scripts** | ✅ N/A | ⚠️ 1-2 days |
| **Total** | ✅ **0 days** | ⚠️ **7-11 days** |

**Verdict:** **SQLite has zero cost** (already done)

---

### 7.2 Operational Cost

| Aspect | SQLite (Current) | PostgreSQL (Shared) |
|--------|------------------|---------------------|
| **Server Setup** | ✅ None | ⚠️ PostgreSQL server |
| **Maintenance** | ✅ Minimal | ⚠️ Database maintenance |
| **Monitoring** | ✅ Simple | ⚠️ Database monitoring |
| **Backup** | ✅ File copy | ⚠️ pg_dump |
| **Resource Usage** | ✅ Low | ⚠️ Medium-High |

**Verdict:** **SQLite has lower operational cost**

---

### 7.3 Performance Benefit

| Scenario | SQLite Benefit | PostgreSQL Benefit |
|----------|----------------|-------------------|
| **Single user, single file** | ✅ Same | ✅ Same |
| **Multiple users, single file** | ⚠️ May be slower | ✅ Better |
| **Multiple users, multiple files** | ⚠️ Limited | ✅ Better |
| **High write load** | ⚠️ Limited | ✅ Better |

**Verdict:** **PostgreSQL only beneficial at scale**

---

## 8. Recommendations

### 8.1 For MVP / Initial Release

**Use: SQLite (Current Hybrid Mode)**

**Why:**
- ✅ Already implemented and tested
- ✅ Zero development cost
- ✅ Simple deployment
- ✅ Fast enough for initial users
- ✅ Low resource usage
- ✅ Easy backup (file copy)

**When to Reconsider:**
- When you have 50+ concurrent users
- When you need multi-server deployment
- When concurrent writes become a bottleneck
- When you need advanced PostgreSQL features

---

### 8.2 For Production at Scale

**Use: PostgreSQL (Shared Database with file_id)**

**Why:**
- ✅ Better concurrent access
- ✅ Better for multi-server
- ✅ Advanced features (full-text search, partitioning)
- ✅ Better monitoring and management
- ✅ Industry standard

**Implementation:**
- ✅ Use shared database (not per-file databases)
- ✅ Add file_id to all tables and queries
- ✅ Use connection pooling
- ✅ Consider partitioning by file_id for very large scale

---

### 8.3 Hybrid Approach (Best of Both)

**Use: SQLite for small deployments, PostgreSQL for scale**

**Architecture:**
```
Configuration:
- development/small: SQLite (embedded)
- production/large: PostgreSQL (shared)

Code:
- Abstract storage interface
- SQLite implementation (current)
- PostgreSQL implementation (new)
- Choose at runtime based on config
```

**Benefits:**
- ✅ Simple for development/testing
- ✅ Scalable for production
- ✅ Gradual migration path
- ✅ Best of both worlds

---

## 9. Key Insights

### 9.1 SQLite vs PostgreSQL for This Use Case

**SQLite is Better When:**
- ✅ Single-server deployment
- ✅ Low to medium traffic
- ✅ Mostly read operations
- ✅ Simple deployment requirements
- ✅ Per-file isolation is desired

**PostgreSQL is Better When:**
- ✅ Multi-server deployment
- ✅ High concurrent access
- ✅ Many concurrent writes
- ✅ Need advanced features
- ✅ Shared database is acceptable

---

### 9.2 Per-File vs Shared Database

**Per-File Databases (SQLite):**
- ✅ **Isolation** - each file is independent
- ✅ **Simple queries** - no file_id needed
- ✅ **Easy deletion** - just delete files
- ✅ **Easy backup** - copy directory

**Shared Database (PostgreSQL):**
- ✅ **Efficiency** - better resource utilization
- ✅ **Scalability** - handles more files
- ✅ **Concurrent access** - better performance
- ⚠️ **Complexity** - need file_id everywhere

**Verdict:** **Per-file for simplicity, Shared for scale**

---

### 9.3 BadgerDB Stays the Same

**Important:** BadgerDB (graph structure) doesn't change regardless of SQLite vs PostgreSQL choice.

**Why:**
- BadgerDB is optimized for graph operations
- It's file-based (per-file directories)
- No need to change it
- Works with both SQLite and PostgreSQL

**Architecture:**
```
BadgerDB (unchanged):
/var/lib/ligneous/files/{file_id}/graph/

SQLite (current):
/var/lib/ligneous/files/{file_id}/indexes.db

OR

PostgreSQL (shared):
PostgreSQL server: ligneous_graphs database
  - nodes table (with file_id)
  - xref_mapping table (with file_id)
  - components table (with file_id)
```

---

## 10. Final Recommendation

### 10.1 Short Term (MVP)

**Keep SQLite in Hybrid Mode**

**Reasons:**
1. ✅ Already implemented and working
2. ✅ Zero development cost
3. ✅ Sufficient for initial users
4. ✅ Simple deployment
5. ✅ Fast enough

**Action:** **No changes needed** ✅

---

### 10.2 Medium Term (Scale)

**Add PostgreSQL Support (Shared Database)**

**Reasons:**
1. ✅ Better concurrent access
2. ✅ Better for multi-server
3. ✅ Industry standard
4. ✅ Advanced features

**Action:**
1. Implement PostgreSQL storage (parallel to SQLite)
2. Add file_id to schema and queries
3. Support both (configurable)
4. Migrate gradually

---

### 10.3 Long Term (Enterprise)

**Use PostgreSQL (Shared) + Optional Redis**

**Reasons:**
1. ✅ Best performance at scale
2. ✅ Multi-server support
3. ✅ Advanced features
4. ✅ Better monitoring

**Action:**
1. Migrate all files to PostgreSQL
2. Consider removing SQLite support (or keep for dev)
3. Add Redis for shared caching (optional)

---

## 11. Conclusion

### 11.1 Answer to "What if we switched to PostgreSQL?"

**For Graph Storage (BadgerDB):** ✅ **No change needed** - BadgerDB stays the same

**For Indexed Metadata (SQLite → PostgreSQL):**

**Option 1: Per-File PostgreSQL Databases**
- ❌ **Not recommended** - wasteful, complex, doesn't scale

**Option 2: Shared PostgreSQL Database with file_id**
- ✅ **Recommended for scale** - efficient, scalable, industry standard
- ⚠️ **Requires code changes** - add file_id to schema and queries
- ⚠️ **More complex setup** - need PostgreSQL server

**Current SQLite:**
- ✅ **Sufficient for MVP** - already implemented, simple, fast
- ✅ **Keep for now** - migrate to PostgreSQL when you need scale

---

### 11.2 Decision Matrix

| Scenario | Recommendation |
|----------|---------------|
| **MVP / Initial Release** | ✅ SQLite (current) |
| **Single Server, Low Traffic** | ✅ SQLite (current) |
| **Single Server, High Traffic** | ⚠️ Consider PostgreSQL (shared) |
| **Multi-Server Deployment** | ✅ PostgreSQL (shared) required |
| **Development/Testing** | ✅ SQLite (simpler) |
| **Production at Scale** | ✅ PostgreSQL (shared) |

---

### 11.3 Migration Strategy

**Phase 1: Keep SQLite (Now)**
- ✅ Use current hybrid mode
- ✅ Monitor performance
- ✅ Identify bottlenecks

**Phase 2: Add PostgreSQL Support (When Needed)**
- ✅ Implement PostgreSQL storage (shared database)
- ✅ Support both (configurable)
- ✅ Migrate gradually

**Phase 3: Full Migration (At Scale)**
- ✅ All files on PostgreSQL
- ✅ Optional: Remove SQLite support
- ✅ Optional: Add Redis for caching

---

**Final Verdict:** **SQLite is sufficient for now. Switch to PostgreSQL (shared database) when you need better concurrent access or multi-server deployment. Don't use per-file PostgreSQL databases - use a shared database with file_id.**

