# Indexing Analysis: Storage Options for Complex Queries

## Current Indexing Requirements

### Index Types Currently Used

1. **Name Index** (Multiple Query Types)
   - Exact match: `findByNameExact(name)` → O(1) lookup
   - Prefix match: `findByNameStarts(prefix)` → Prefix scan
   - Substring match: `findByName(pattern)` → Full scan with contains check
   - Word-based: Index individual words in names

2. **Date Index** (Range Queries)
   - Sorted array: `birthDateIndex []*dateIndexEntry`
   - Binary search: `findByBirthDate(start, end)` → O(log n) + linear scan
   - Range queries: Start date to end date

3. **Place Index** (Substring Queries)
   - Word-based: Index individual words in places
   - Substring match: `findByBirthPlace(place)` → Full scan with contains

4. **Sex Index** (Exact Lookup)
   - Direct lookup: `findBySex(sex)` → O(1) map lookup

5. **Boolean Indexes** (Direct Lookup)
   - Has children: `hasChildrenIndex map[string]bool`
   - Has spouse: `hasSpouseIndex map[string]bool`
   - Living: `livingIndex map[string]bool`

### Query Patterns

**Indexed (Fast):**
- `ByName(pattern)` - Substring match
- `ByNameExact(name)` - Exact match
- `ByNameStarts(prefix)` - Prefix match
- `ByBirthDate(start, end)` - Date range
- `ByBirthPlace(place)` - Substring match
- `BySex(sex)` - Exact match
- `HasChildren()`, `HasSpouse()`, `Living()` - Boolean

**Non-Indexed (Slower):**
- `ByNameEnds(suffix)` - Suffix match (no suffix index)
- `Deceased()` - Checks death date field
- Custom `Where()` filters

### Index Storage Requirements

For 10M individuals:
- **Name index:** ~500 MB (assuming average 3 words per name, ~50 bytes per entry)
- **Date index:** ~240 MB (sorted array of 10M entries, ~24 bytes per entry)
- **Place index:** ~300 MB (similar to name index)
- **Sex index:** ~80 MB (3 values: M, F, U)
- **Boolean indexes:** ~30 MB each (3 × 10M × 1 byte)
- **Total indexes:** ~1.2 GB (in-memory)

**Problem:** These indexes are currently stored in RAM, which contributes to memory pressure at 10M scale.

---

## Storage Options with Indexing

### Option 1: BadgerDB with Custom Indexes

**How it works:**
- Store indexes as key-value pairs in BadgerDB
- Use prefix-based keys for range queries
- Keep hot index entries in Go map (LRU cache)

**Index Storage Strategy:**

```
# Name Index (Exact)
Key: "idx:name:exact:{lowercase_name}" → Value: []nodeID (serialized)

# Name Index (Prefix)
Key: "idx:name:prefix:{lowercase_prefix}" → Value: []nodeID (serialized)
# Problem: Prefix matching requires iterating all keys with prefix

# Name Index (Word-based)
Key: "idx:name:word:{word}" → Value: []nodeID (serialized)

# Date Index (Range)
Key: "idx:date:{timestamp}" → Value: nodeID (uint32)
# Sorted by key (BadgerDB maintains key order)
# Range query: Iterator from start timestamp to end timestamp

# Place Index
Key: "idx:place:{lowercase_place}" → Value: []nodeID (serialized)

# Sex Index
Key: "idx:sex:{sex}" → Value: []nodeID (serialized)

# Boolean Indexes
Key: "idx:has_children:{nodeID}" → Value: true/false (1 byte)
Key: "idx:has_spouse:{nodeID}" → Value: true/false (1 byte)
Key: "idx:living:{nodeID}" → Value: true/false (1 byte)
```

**Pros:**
- ✅ **Range queries:** BadgerDB iterators support efficient range scans
- ✅ **Prefix queries:** Iterator with prefix works well
- ✅ **Memory efficient:** Indexes stored on disk, only hot entries in RAM
- ✅ **Pure Go:** No CGO dependency

**Cons:**
- ⚠️ **Substring queries:** Still requires full scan (same as current)
- ⚠️ **Complex index management:** Must manually maintain indexes
- ⚠️ **Index updates:** Must update indexes when data changes
- ⚠️ **No built-in indexes:** Must implement index logic yourself

**Index Performance:**
- **Exact lookup:** O(1) expected (from cache) or O(log n) from BadgerDB
- **Prefix query:** O(k) where k = number of matching keys (efficient iterator)
- **Range query:** O(k) where k = number of results (efficient iterator)
- **Substring query:** O(n) - still requires full scan (same as current)

**Memory Usage:**
- **Hot index entries:** 10K-100K entries in RAM (~10-100 MB)
- **Cold index entries:** On disk, memory-mapped
- **Total:** 2-5 GB RAM for 10M individuals (including indexes)

---

### Option 2: SQLite with Built-in Indexes

**How it works:**
- Use SQLite's built-in indexing
- Create indexes on columns
- Use SQL queries for filtering

**Index Storage Strategy:**

```sql
-- Nodes table
CREATE TABLE nodes (
    id INTEGER PRIMARY KEY,
    xref TEXT UNIQUE,
    type TEXT,
    name TEXT,
    birth_date INTEGER,  -- Unix timestamp
    birth_place TEXT,
    sex TEXT,
    data BLOB  -- Serialized node data
);

-- Indexes
CREATE INDEX idx_name ON nodes(name);
CREATE INDEX idx_name_lower ON nodes(LOWER(name));
CREATE INDEX idx_birth_date ON nodes(birth_date);
CREATE INDEX idx_birth_place ON nodes(birth_place);
CREATE INDEX idx_sex ON nodes(sex);

-- Full-text search (for substring matching)
CREATE VIRTUAL TABLE nodes_fts USING fts5(name, birth_place);

-- Edges table
CREATE TABLE edges (
    from_id INTEGER,
    to_id INTEGER,
    type TEXT,
    data BLOB
);
CREATE INDEX idx_edges_from ON edges(from_id);
CREATE INDEX idx_edges_to ON edges(to_id);
```

**Query Examples:**

```sql
-- Exact name match
SELECT * FROM nodes WHERE LOWER(name) = LOWER(?)

-- Prefix match
SELECT * FROM nodes WHERE LOWER(name) LIKE LOWER(?) || '%'

-- Substring match (full-text search)
SELECT * FROM nodes_fts WHERE name MATCH ?

-- Date range
SELECT * FROM nodes WHERE birth_date BETWEEN ? AND ?

-- Combined query
SELECT * FROM nodes 
WHERE LOWER(name) LIKE LOWER(?) || '%'
  AND birth_date BETWEEN ? AND ?
  AND sex = ?
```

**Pros:**
- ✅ **Built-in indexes:** SQLite handles index creation/maintenance
- ✅ **Full-text search:** FTS5 for efficient substring matching
- ✅ **Complex queries:** SQL makes complex queries easy
- ✅ **Automatic optimization:** Query planner optimizes queries
- ✅ **Multiple indexes:** Can have multiple indexes per table
- ✅ **Index maintenance:** SQLite handles index updates automatically

**Cons:**
- ⚠️ **CGO dependency:** Requires CGO (SQLite is C library)
- ⚠️ **SQL overhead:** SQL parsing, query planning overhead
- ⚠️ **Less control:** Can't fine-tune index structure as much
- ⚠️ **Memory:** Indexes still consume memory (though can be tuned)

**Index Performance:**
- **Exact lookup:** O(log n) with B-tree index
- **Prefix query:** O(log n) + O(k) with B-tree index
- **Substring query:** O(k) with FTS5 (full-text search)
- **Range query:** O(log n) + O(k) with B-tree index
- **Combined queries:** Query planner optimizes automatically

**Memory Usage:**
- **Indexes:** SQLite indexes stored in memory-mapped files
- **Query cache:** SQLite query cache for hot data
- **Total:** 2-5 GB RAM for 10M individuals (similar to BadgerDB)

---

### Option 3: Hybrid Approach (BadgerDB + In-Memory Indexes)

**How it works:**
- Store nodes/edges in BadgerDB
- Keep **hot indexes** in RAM (LRU cache)
- Store **cold indexes** in BadgerDB

**Index Storage Strategy:**

```
# Hot Indexes (In RAM)
- Name index (exact): map[string][]uint32 (10K-100K entries)
- Date index (recent): sorted slice (last 1M entries)
- Sex index: map[string][]uint32 (always small, keep in RAM)

# Cold Indexes (In BadgerDB)
- Name index (all): "idx:name:{name}" → []nodeID
- Date index (all): "idx:date:{timestamp}" → nodeID
- Place index: "idx:place:{place}" → []nodeID
```

**Pros:**
- ✅ **Best of both:** Fast hot data, efficient cold data
- ✅ **Memory efficient:** Only hot indexes in RAM
- ✅ **Flexible:** Can tune what stays in RAM

**Cons:**
- ⚠️ **Complexity:** More complex to implement
- ⚠️ **Cache management:** Must manage LRU cache for indexes

**Memory Usage:**
- **Hot indexes:** ~100-200 MB (10K-100K entries)
- **Cold indexes:** On disk, memory-mapped
- **Total:** 2-5 GB RAM for 10M individuals

---

### Option 4: Pebble with Custom Indexes

**Similar to BadgerDB:**
- Same indexing strategy as BadgerDB
- Slightly better performance (newer design)
- Less battle-tested

**Verdict:** Very similar to BadgerDB, but newer and less proven.

---

## Comparison: Indexing Capabilities

| Feature | BadgerDB | SQLite | Hybrid (BadgerDB + RAM) |
|---------|----------|--------|-------------------------|
| **Exact Lookup** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Prefix Query** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Substring Query** | ⭐⭐ | ⭐⭐⭐⭐⭐ (FTS5) | ⭐⭐ |
| **Range Query** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Combined Queries** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Index Maintenance** | ⭐⭐ (manual) | ⭐⭐⭐⭐⭐ (automatic) | ⭐⭐⭐ (semi-automatic) |
| **Memory Efficiency** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Complexity** | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ |

---

## Recommendation: SQLite for Indexing

### Why SQLite for Indexing?

1. **Built-in Indexes:**
   - SQLite handles index creation/maintenance automatically
   - No need to manually maintain indexes
   - Query planner optimizes queries automatically

2. **Full-Text Search:**
   - FTS5 provides efficient substring matching
   - Better than manual substring scanning
   - Handles complex text queries

3. **Complex Queries:**
   - SQL makes complex queries easy
   - Can combine multiple filters efficiently
   - Query planner optimizes automatically

4. **Index Maintenance:**
   - SQLite updates indexes automatically when data changes
   - No manual index update logic needed

5. **Memory Efficiency:**
   - Indexes stored in memory-mapped files
   - Similar memory usage to BadgerDB
   - OS handles paging

### When to Use BadgerDB Instead?

**Use BadgerDB if:**
- You want pure Go (no CGO)
- You have simple indexing needs
- You want maximum control over index structure
- You're okay with manual index maintenance

**Use SQLite if:**
- You need complex queries
- You want full-text search
- You want automatic index maintenance
- You're okay with CGO dependency

---

## Hybrid Recommendation: SQLite for Indexing, BadgerDB for Graph

**Best of Both Worlds:**

1. **SQLite for Indexing:**
   - Store indexes in SQLite
   - Use SQL queries for filtering
   - Full-text search for substring matching
   - Automatic index maintenance

2. **BadgerDB for Graph:**
   - Store nodes/edges in BadgerDB
   - Fast graph traversal
   - Memory-mapped for efficiency

**Architecture:**
```
SQLite Database:
  - Indexes (name, date, place, sex, etc.)
  - Metadata (xref → nodeID mapping)

BadgerDB Database:
  - Nodes (nodeID → NodeData)
  - Edges (nodeID → []EdgeData)
  - Components (componentID → []nodeID)
```

**Query Flow:**
1. Use SQLite to find matching nodeIDs (fast index lookup)
2. Use BadgerDB to load node/edge data for those nodeIDs
3. Return results

**Pros:**
- ✅ Best indexing (SQLite)
- ✅ Best graph storage (BadgerDB)
- ✅ Memory efficient (both memory-mapped)
- ✅ Fast queries (indexed lookups + fast graph access)

**Cons:**
- ⚠️ Two databases to manage
- ⚠️ More complex architecture
- ⚠️ CGO dependency (SQLite)

---

## Final Recommendation

### For Maximum Indexing Capabilities: SQLite

**Why:**
- Built-in indexes with automatic maintenance
- Full-text search (FTS5) for substring matching
- Complex queries with SQL
- Query planner optimizes automatically
- Memory-mapped for efficiency

**Trade-off:**
- CGO dependency (but manageable)
- SQL overhead (but query planner optimizes)

### For Pure Go with Good Indexing: BadgerDB + Custom Indexes

**Why:**
- Pure Go (no CGO)
- Good range/prefix query support
- Memory-mapped for efficiency
- Full control over index structure

**Trade-off:**
- Manual index maintenance
- No built-in full-text search (substring queries slower)
- More complex to implement

### For Best of Both: Hybrid (SQLite + BadgerDB)

**Why:**
- SQLite for indexing (best indexing capabilities)
- BadgerDB for graph (pure Go, fast graph operations)
- Memory efficient (both memory-mapped)

**Trade-off:**
- Two databases to manage
- More complex architecture
- CGO dependency (SQLite)

---

## Decision Matrix

| Requirement | BadgerDB | SQLite | Hybrid |
|-------------|----------|--------|--------|
| **Indexing Complexity** | Manual | Automatic | Automatic |
| **Full-Text Search** | ❌ | ✅ | ✅ |
| **Complex Queries** | ⚠️ | ✅ | ✅ |
| **Pure Go** | ✅ | ❌ | ❌ |
| **Memory Efficiency** | ✅ | ✅ | ✅ |
| **Implementation Complexity** | Medium | Low | High |
| **Maintenance** | High | Low | Medium |

**Recommendation:** 
- **If you need advanced indexing:** SQLite
- **If you want pure Go:** BadgerDB with custom indexes
- **If you want best of both:** Hybrid (SQLite + BadgerDB)

