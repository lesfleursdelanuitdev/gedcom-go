# Storage Strategy Analysis: Hybrid Mode vs PostgreSQL + Redis

**Date:** 2025-01-27  
**Purpose:** Determine if existing hybrid mode (BadgerDB + SQLite) is sufficient for REST API, or if PostgreSQL + Redis is needed

---

## Executive Summary

**Recommendation:** **Hybrid mode is sufficient for graph storage, but you need additional storage for API metadata.**

**Best Approach:** **Hybrid Mode (BadgerDB + SQLite) for graphs + Lightweight SQLite for API metadata**

**Alternative:** Hybrid Mode + PostgreSQL (only if you need advanced features like full-text search, complex queries, or multi-server setup)

---

## 1. Current Hybrid Mode Capabilities

### 1.1 What Hybrid Mode Does

**SQLite Component:**
- ✅ Stores indexed metadata (name, birth_date, birth_place, sex, has_children, has_spouse, living)
- ✅ Fast filtering and searching with indexes
- ✅ Full-text search (FTS5) for names and places
- ✅ Prepared statements for performance
- ✅ WAL mode for concurrent reads
- ✅ Composite indexes for common query patterns

**BadgerDB Component:**
- ✅ Stores complete graph structure (nodes, edges)
- ✅ Key-value storage optimized for graph traversal
- ✅ Serialized node data (IndividualNode, FamilyNode, etc.)
- ✅ Edge storage for relationships
- ✅ Memory-mapped I/O for performance

**In-Memory Cache:**
- ✅ LRU cache for nodes (configurable size)
- ✅ LRU cache for XREF ↔ nodeID mappings
- ✅ LRU cache for query results
- ✅ Thread-safe operations

### 1.2 Hybrid Mode Strengths

1. **Embedded Databases:**
   - No separate server process required
   - File-based storage (easy backup, migration)
   - Low overhead
   - Perfect for single-server deployment

2. **Performance:**
   - SQLite indexes for fast filtering
   - BadgerDB optimized for graph operations
   - In-memory caching reduces disk I/O
   - Tested up to 1M+ individuals

3. **Simplicity:**
   - No external dependencies
   - Self-contained per file
   - Easy to deploy

---

## 2. REST API Requirements

### 2.1 Graph Storage (Per File)

**Requirement:** Store and query graph for each uploaded GEDCOM file

| Need | Hybrid Mode | PostgreSQL + Redis |
|------|-------------|-------------------|
| Store graph structure | ✅ BadgerDB | ❌ Would need custom schema |
| Fast filtering/search | ✅ SQLite indexes | ✅ PostgreSQL indexes |
| Graph traversal | ✅ BadgerDB | ⚠️ Would need graph extension |
| Per-file isolation | ✅ Separate files | ⚠️ Need file_id in all queries |
| Performance | ✅ Optimized | ✅ Good, but more overhead |

**Verdict:** **Hybrid mode is BETTER for graph storage** - it's purpose-built for this use case.

---

### 2.2 API Metadata Storage

**Requirement:** Store file metadata, users, API keys, exports, rate limits

| Need | Hybrid Mode | PostgreSQL | SQLite (separate) |
|------|-------------|------------|------------------|
| File metadata | ❌ Not designed for this | ✅ Perfect | ✅ Sufficient |
| User management | ❌ Not designed for this | ✅ Perfect | ✅ Sufficient |
| API keys | ❌ Not designed for this | ✅ Perfect | ✅ Sufficient |
| Export tracking | ❌ Not designed for this | ✅ Perfect | ✅ Sufficient |
| Rate limiting | ❌ Not designed for this | ✅ Perfect | ✅ Sufficient |
| Multi-file queries | ❌ Per-file only | ✅ Cross-file queries | ⚠️ Limited |

**Verdict:** **Need separate storage for API metadata** - hybrid mode is per-file only.

---

### 2.3 Caching Strategy

**Requirement:** Cache parsed graphs, query results across requests

| Need | Hybrid Mode Cache | Redis |
|------|-------------------|-------|
| Per-request caching | ✅ In-memory LRU | ✅ In-memory |
| Shared across requests | ❌ Per-graph instance | ✅ Shared |
| Persistence | ❌ Lost on restart | ✅ Optional persistence |
| Distributed caching | ❌ Single server only | ✅ Multi-server |
| TTL support | ⚠️ LRU eviction only | ✅ Built-in TTL |
| Cache invalidation | ⚠️ Manual | ✅ Built-in |

**Verdict:** **Hybrid mode cache is sufficient for single-server**, Redis needed for multi-server or advanced features.

---

## 3. Comparison Matrix

### 3.1 Graph Storage

| Feature | Hybrid Mode | PostgreSQL + Redis |
|---------|-------------|-------------------|
| **Purpose-built for graphs** | ✅ Yes | ❌ No |
| **Performance** | ✅ Excellent | ⚠️ Good (more overhead) |
| **Storage efficiency** | ✅ Optimized | ⚠️ Less efficient |
| **Query speed** | ✅ Fast | ✅ Fast |
| **Concurrent access** | ✅ SQLite WAL | ✅ PostgreSQL |
| **Backup/restore** | ✅ File copy | ⚠️ pg_dump required |
| **Deployment complexity** | ✅ Simple (embedded) | ⚠️ Requires server setup |
| **Memory usage** | ✅ Lower | ⚠️ Higher |
| **Scalability** | ⚠️ Single server | ✅ Multi-server |

**Winner:** **Hybrid Mode** for graph storage

---

### 3.2 API Metadata

| Feature | Hybrid Mode | PostgreSQL | SQLite (separate) |
|---------|-------------|------------|-------------------|
| **File metadata** | ❌ No | ✅ Yes | ✅ Yes |
| **User management** | ❌ No | ✅ Yes | ✅ Yes |
| **API keys** | ❌ No | ✅ Yes | ✅ Yes |
| **Complex queries** | ❌ No | ✅ Yes | ⚠️ Limited |
| **Full-text search** | ❌ No | ✅ Yes | ⚠️ FTS5 only |
| **Transactions** | ⚠️ Per-file | ✅ Yes | ✅ Yes |
| **Concurrent writes** | ⚠️ Limited | ✅ Excellent | ⚠️ Limited |
| **Deployment** | ✅ Embedded | ⚠️ Server required | ✅ Embedded |

**Winner:** **PostgreSQL** for advanced features, **SQLite** for simplicity

---

### 3.3 Caching

| Feature | Hybrid Mode Cache | Redis |
|---------|-------------------|-------|
| **In-memory speed** | ✅ Fast | ✅ Fast |
| **Shared across requests** | ❌ No | ✅ Yes |
| **TTL support** | ❌ No | ✅ Yes |
| **Distributed** | ❌ No | ✅ Yes |
| **Persistence** | ❌ No | ✅ Optional |
| **Memory usage** | ✅ Lower | ⚠️ Higher |
| **Setup complexity** | ✅ None | ⚠️ Server required |

**Winner:** **Hybrid cache sufficient for single-server**, **Redis for multi-server**

---

## 4. Recommended Architecture

### 4.1 Option 1: Hybrid Mode + SQLite (Recommended for MVP)

**Architecture:**
```
REST API Server
    ↓
┌─────────────────────────────────────┐
│  Hybrid Mode (Per File)             │
│  - SQLite: Indexed metadata          │
│  - BadgerDB: Graph structure         │
│  - In-memory LRU cache                │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│  API Metadata SQLite                 │
│  - files table (file_id, name, etc.) │
│  - users table (if needed)           │
│  - api_keys table                    │
│  - exports table                      │
│  - rate_limits table                  │
└─────────────────────────────────────┘
```

**Pros:**
- ✅ Simple deployment (no external servers)
- ✅ Hybrid mode already implemented and tested
- ✅ Fast for graph operations
- ✅ Easy backup (file copy)
- ✅ Low resource usage
- ✅ Perfect for single-server deployment

**Cons:**
- ⚠️ Limited concurrent writes (SQLite)
- ⚠️ No distributed caching
- ⚠️ No cross-file queries (but API doesn't need this)
- ⚠️ SQLite limitations for complex queries

**Best For:**
- Single-server deployment
- MVP / initial release
- Small to medium scale (< 1000 concurrent users)
- Simple requirements

---

### 4.2 Option 2: Hybrid Mode + PostgreSQL (Recommended for Scale)

**Architecture:**
```
REST API Server
    ↓
┌─────────────────────────────────────┐
│  Hybrid Mode (Per File)             │
│  - SQLite: Indexed metadata          │
│  - BadgerDB: Graph structure         │
│  - In-memory LRU cache                │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│  PostgreSQL                          │
│  - files table                       │
│  - users table                       │
│  - api_keys table                    │
│  - exports table                     │
│  - rate_limits table                 │
│  - Complex queries, full-text search │
└─────────────────────────────────────┘
```

**Pros:**
- ✅ Hybrid mode for graphs (best performance)
- ✅ PostgreSQL for metadata (advanced features)
- ✅ Excellent concurrent writes
- ✅ Full-text search capabilities
- ✅ Complex queries across tables
- ✅ Better for multi-user scenarios

**Cons:**
- ⚠️ Requires PostgreSQL server setup
- ⚠️ More complex deployment
- ⚠️ Higher resource usage
- ⚠️ More moving parts

**Best For:**
- Production at scale
- Multi-user scenarios
- Need for complex queries
- Future growth

---

### 4.3 Option 3: Hybrid Mode + PostgreSQL + Redis (Overkill for MVP)

**Architecture:**
```
REST API Server
    ↓
┌─────────────────────────────────────┐
│  Hybrid Mode (Per File)             │
│  - SQLite: Indexed metadata          │
│  - BadgerDB: Graph structure         │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│  PostgreSQL (Metadata)              │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│  Redis (Shared Cache)               │
│  - Parsed graph cache               │
│  - Query result cache                │
│  - Rate limiting                     │
└─────────────────────────────────────┘
```

**Pros:**
- ✅ All benefits of Option 2
- ✅ Shared cache across requests
- ✅ Distributed caching (multi-server)
- ✅ Advanced rate limiting
- ✅ Cache persistence

**Cons:**
- ⚠️ Most complex setup
- ⚠️ Highest resource usage
- ⚠️ Overkill for single-server
- ⚠️ More failure points

**Best For:**
- Multi-server deployment
- High-traffic scenarios
- Need for distributed caching
- Enterprise scale

---

## 5. Detailed Analysis: What Each Component Needs

### 5.1 Graph Storage (Per File)

**What's needed:**
- Store complete graph structure (nodes, edges)
- Fast filtering by name, date, place, sex
- Fast graph traversal (ancestors, descendants, paths)
- Per-file isolation

**Hybrid Mode Assessment:**
- ✅ **Perfect fit** - purpose-built for this
- ✅ SQLite handles all filtering needs
- ✅ BadgerDB handles graph structure
- ✅ In-memory cache reduces disk I/O
- ✅ Tested up to 1M+ individuals

**PostgreSQL Assessment:**
- ⚠️ Would need custom schema for graph storage
- ⚠️ Not optimized for graph operations
- ⚠️ More overhead for simple queries
- ✅ Better for complex cross-file queries (but API doesn't need this)

**Verdict:** **Keep Hybrid Mode for graph storage** ✅

---

### 5.2 API Metadata Storage

**What's needed:**
- File metadata (file_id, name, size, upload date, parse status)
- User accounts (if implementing authentication)
- API keys (if using API key auth)
- Export tracking (export_id, format, status, download_url)
- Rate limiting (per API key, per endpoint)

**Hybrid Mode Assessment:**
- ❌ **Not designed for this** - hybrid mode is per-file only
- ❌ No way to query across files
- ❌ No user/API key management

**SQLite (Separate) Assessment:**
- ✅ Sufficient for MVP
- ✅ Simple schema
- ✅ Embedded (no server)
- ✅ Fast for simple queries
- ⚠️ Limited concurrent writes
- ⚠️ No full-text search (unless FTS5)

**PostgreSQL Assessment:**
- ✅ Excellent for metadata
- ✅ Handles concurrent writes well
- ✅ Full-text search
- ✅ Complex queries
- ✅ Better for scale
- ⚠️ Requires server setup

**Verdict:** **SQLite sufficient for MVP, PostgreSQL for scale** ✅

---

### 5.3 Caching Strategy

**What's needed:**
- Cache parsed graphs (avoid re-parsing)
- Cache query results (ancestors, descendants, etc.)
- Cache file metadata
- Shared across HTTP requests

**Hybrid Mode Cache Assessment:**
- ✅ Fast in-memory LRU cache
- ✅ Already implemented
- ❌ Per-graph instance (not shared across requests)
- ❌ Lost on server restart
- ❌ No TTL support

**Redis Assessment:**
- ✅ Shared across all requests
- ✅ TTL support
- ✅ Distributed (multi-server)
- ✅ Optional persistence
- ⚠️ Requires server setup
- ⚠️ Network overhead (minimal)

**Verdict:** **Hybrid cache sufficient for single-server, Redis for multi-server** ✅

---

## 6. Storage Layout

### 6.1 Option 1: Hybrid Mode + SQLite (Recommended)

**File Structure:**
```
/var/lib/ligneous/
├── api_metadata.db          # SQLite for API metadata
├── files/
│   ├── {file_id_1}/
│   │   ├── original.ged
│   │   ├── indexes.db       # SQLite for this file
│   │   └── graph/            # BadgerDB for this file
│   ├── {file_id_2}/
│   │   ├── original.ged
│   │   ├── indexes.db
│   │   └── graph/
│   └── ...
└── exports/
    └── {export_id}.{ext}
```

**API Metadata Schema (SQLite):**
```sql
-- Files table
CREATE TABLE files (
    id TEXT PRIMARY KEY,  -- UUID
    user_id TEXT,
    name TEXT NOT NULL,
    original_filename TEXT,
    size INTEGER NOT NULL,
    individuals_count INTEGER,
    families_count INTEGER,
    parse_status TEXT,  -- pending, parsed, error
    parse_errors INTEGER DEFAULT 0,
    parse_warnings INTEGER DEFAULT 0,
    created_at INTEGER,
    updated_at INTEGER
);

-- API keys (if using API key auth)
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY,
    key_hash TEXT UNIQUE NOT NULL,
    user_id TEXT,
    name TEXT,
    created_at INTEGER,
    expires_at INTEGER,
    last_used_at INTEGER
);

-- Exports
CREATE TABLE exports (
    id TEXT PRIMARY KEY,
    file_id TEXT REFERENCES files(id),
    format TEXT NOT NULL,
    size INTEGER,
    status TEXT,  -- pending, completed, failed
    download_url TEXT,
    expires_at INTEGER,
    created_at INTEGER
);

-- Rate limiting (optional, can use in-memory)
CREATE TABLE rate_limits (
    api_key_id TEXT,
    endpoint TEXT,
    count INTEGER DEFAULT 0,
    window_start INTEGER,
    PRIMARY KEY (api_key_id, endpoint, window_start)
);
```

**Pros:**
- ✅ Simple file-based structure
- ✅ Easy backup (copy directory)
- ✅ No external dependencies
- ✅ Fast for graph operations

---

### 6.2 Option 2: Hybrid Mode + PostgreSQL

**File Structure:**
```
/var/lib/ligneous/
├── files/
│   ├── {file_id_1}/
│   │   ├── original.ged
│   │   ├── indexes.db       # SQLite for this file
│   │   └── graph/            # BadgerDB for this file
│   └── ...
└── exports/
    └── {export_id}.{ext}

PostgreSQL Database:
- files table
- users table
- api_keys table
- exports table
- rate_limits table
```

**Pros:**
- ✅ Better for concurrent writes
- ✅ Advanced query capabilities
- ✅ Full-text search
- ✅ Better for scale

---

## 7. Performance Considerations

### 7.1 Graph Query Performance

**Hybrid Mode:**
- SQLite indexes: **Very fast** (microseconds for indexed queries)
- BadgerDB graph traversal: **Fast** (optimized for graph operations)
- In-memory cache: **Eliminates disk I/O** for hot data
- **Tested up to 1M+ individuals** ✅

**PostgreSQL:**
- Indexes: **Fast** (similar to SQLite)
- Graph traversal: **Slower** (not optimized for graphs)
- Network overhead: **Minimal but present**
- **Would need custom schema** ⚠️

**Verdict:** **Hybrid mode is faster for graph operations** ✅

---

### 7.2 Concurrent Access

**Hybrid Mode (SQLite):**
- WAL mode: **Good for concurrent reads**
- Writes: **Limited** (single writer, multiple readers)
- **Sufficient for API** (mostly reads, occasional writes)

**PostgreSQL:**
- **Excellent concurrent access**
- **Better for high write load**
- **Overkill for mostly-read API**

**Verdict:** **Hybrid mode sufficient for API** (mostly reads) ✅

---

### 7.3 Memory Usage

**Hybrid Mode:**
- SQLite: **Low** (file-based, memory-mapped)
- BadgerDB: **Low** (memory-mapped I/O)
- Cache: **Configurable** (LRU eviction)
- **Total: ~50-100MB per 100K individuals**

**PostgreSQL + Redis:**
- PostgreSQL: **Higher** (connection pools, buffers)
- Redis: **Higher** (in-memory cache)
- **Total: ~200-300MB per 100K individuals**

**Verdict:** **Hybrid mode uses less memory** ✅

---

## 8. Scalability Analysis

### 8.1 Single Server

**Hybrid Mode + SQLite:**
- ✅ **Perfect fit**
- ✅ Handles 1000+ concurrent requests
- ✅ Low resource usage
- ✅ Simple deployment

**PostgreSQL + Redis:**
- ⚠️ **Overkill**
- ⚠️ More resource usage
- ⚠️ More complex setup
- ✅ Better for future growth

**Verdict:** **Hybrid Mode + SQLite for single server** ✅

---

### 8.2 Multi-Server (Future)

**Hybrid Mode + SQLite:**
- ❌ **Not suitable** (file-based, not shared)
- ❌ No distributed caching
- ❌ File synchronization issues

**Hybrid Mode + PostgreSQL + Redis:**
- ✅ **Suitable**
- ✅ Shared metadata (PostgreSQL)
- ✅ Distributed cache (Redis)
- ✅ File storage can be shared (NFS/S3)

**Verdict:** **PostgreSQL + Redis needed for multi-server** ✅

---

## 9. Migration Path

### 9.1 Start Simple, Scale Later

**Phase 1: MVP (Hybrid Mode + SQLite)**
```
✅ Use existing hybrid mode for graphs
✅ Add SQLite for API metadata
✅ In-memory cache per request
✅ Single server deployment
```

**Phase 2: Scale (Add PostgreSQL)**
```
✅ Keep hybrid mode for graphs
✅ Migrate API metadata to PostgreSQL
✅ Add Redis for shared caching (optional)
✅ Multi-server support
```

**Migration Strategy:**
- Graph storage: **No migration needed** (hybrid mode stays)
- API metadata: **Simple migration** (SQLite → PostgreSQL)
- Cache: **Add Redis layer** (hybrid cache + Redis)

---

## 10. Final Recommendation

### 10.1 For MVP / Initial Release

**Use: Hybrid Mode (BadgerDB + SQLite) + Separate SQLite for API Metadata**

**Why:**
- ✅ Hybrid mode already implemented and tested
- ✅ Perfect for graph storage (purpose-built)
- ✅ Simple deployment (no external servers)
- ✅ Fast performance
- ✅ Low resource usage
- ✅ Easy backup (file copy)
- ✅ Sufficient for single-server

**What to add:**
- Separate SQLite database for API metadata
- File management layer (create/delete files)
- Graph loading from hybrid storage per request

---

### 10.2 For Production at Scale

**Use: Hybrid Mode (BadgerDB + SQLite) + PostgreSQL + Optional Redis**

**Why:**
- ✅ Keep hybrid mode (best for graphs)
- ✅ PostgreSQL for metadata (better concurrent access)
- ✅ Redis for shared caching (optional, if needed)
- ✅ Better for multi-user scenarios
- ✅ Future-proof for growth

**When to migrate:**
- When you need multi-server deployment
- When you have high concurrent write load
- When you need advanced query features
- When you need distributed caching

---

## 11. Implementation Strategy

### 11.1 Phase 1: MVP (Hybrid Mode + SQLite)

**Storage Components:**
1. **Hybrid Mode** (existing):
   - One SQLite + BadgerDB per file
   - Located at: `/var/lib/ligneous/files/{file_id}/`

2. **API Metadata SQLite** (new):
   - Single database: `/var/lib/ligneous/api_metadata.db`
   - Tables: files, api_keys, exports, rate_limits

3. **File Storage**:
   - Original GEDCOM files: `/var/lib/ligneous/files/{file_id}/original.ged`
   - Exports: `/var/lib/ligneous/exports/{export_id}.{ext}`

**API Flow:**
```
1. Upload file → Store in filesystem
2. Parse → Build hybrid graph (SQLite + BadgerDB)
3. Store metadata → API metadata SQLite
4. Query → Load graph from hybrid storage
5. Cache → In-memory LRU (per graph instance)
```

---

### 11.2 Phase 2: Scale (Add PostgreSQL)

**Storage Components:**
1. **Hybrid Mode** (keep):
   - Still one SQLite + BadgerDB per file
   - Still best for graph operations

2. **PostgreSQL** (new):
   - Migrate API metadata from SQLite
   - Add advanced features (full-text search, etc.)

3. **Redis** (optional):
   - Shared cache for parsed graphs
   - Query result cache
   - Rate limiting

**Migration:**
- Graph storage: **No change** (hybrid mode stays)
- API metadata: **Migrate SQLite → PostgreSQL**
- Cache: **Add Redis layer** (hybrid cache + Redis)

---

## 12. Cost-Benefit Analysis

### 12.1 Development Effort

| Approach | Development Time | Complexity |
|----------|----------------|------------|
| Hybrid + SQLite | **Low** (reuse existing) | **Low** |
| Hybrid + PostgreSQL | **Medium** (add PostgreSQL) | **Medium** |
| Hybrid + PostgreSQL + Redis | **High** (add both) | **High** |

**Recommendation:** Start with Hybrid + SQLite, migrate later if needed.

---

### 12.2 Operational Effort

| Approach | Setup | Maintenance | Monitoring |
|----------|-------|-------------|------------|
| Hybrid + SQLite | **Very Easy** | **Easy** | **Simple** |
| Hybrid + PostgreSQL | **Medium** | **Medium** | **More complex** |
| Hybrid + PostgreSQL + Redis | **Complex** | **Complex** | **Complex** |

**Recommendation:** Start simple, add complexity only when needed.

---

## 13. Conclusion

### 13.1 Answer to "Is Hybrid Mode Sufficient?"

**For Graph Storage:** ✅ **YES - Hybrid mode is perfect and better than PostgreSQL**

**For API Metadata:** ❌ **NO - Need separate storage** (SQLite or PostgreSQL)

**For Caching:** ⚠️ **MOSTLY - Sufficient for single-server, Redis for multi-server**

---

### 13.2 Recommended Approach

**MVP / Initial Release:**
```
Hybrid Mode (BadgerDB + SQLite) for graphs
+ 
Separate SQLite for API metadata
+
In-memory LRU cache (existing)
```

**Production at Scale:**
```
Hybrid Mode (BadgerDB + SQLite) for graphs
+
PostgreSQL for API metadata
+
Optional Redis for shared caching
```

---

### 13.3 Key Insights

1. **Don't replace hybrid mode** - it's purpose-built and performs better than PostgreSQL for graph operations

2. **Add lightweight metadata storage** - SQLite is sufficient for MVP, PostgreSQL for scale

3. **Cache is optional** - Hybrid mode's in-memory cache is sufficient for single-server; add Redis only if you need:
   - Shared cache across requests
   - Distributed caching
   - TTL-based expiration

4. **Start simple, scale later** - Easy migration path from SQLite → PostgreSQL

5. **Hybrid mode is the secret weapon** - It's already optimized for your exact use case (genealogy graphs)

---

**Final Verdict:** **Hybrid mode is sufficient and optimal for graph storage. Add SQLite for API metadata (MVP) or PostgreSQL (scale). Redis is optional for advanced caching needs.**

