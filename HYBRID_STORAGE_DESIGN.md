# Hybrid Storage Design: SQLite + BadgerDB

## Architecture Overview

### Two-Database Approach

**SQLite Database (`indexes.db`):**
- Purpose: Indexes and metadata
- Strengths: Automatic index maintenance, FTS5 full-text search, complex queries
- Use cases: Filtering, searching, metadata lookups

**BadgerDB Database (`graph.badger`):**
- Purpose: Graph structure (nodes, edges, components)
- Strengths: Pure Go, fast graph traversal, memory-mapped
- Use cases: Graph operations, relationship queries, path finding

### Data Flow

```
┌─────────────────┐
│  GEDCOM File   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Parser         │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Hybrid Graph Builder               │
│  ┌──────────────┐  ┌──────────────┐ │
│  │  SQLite      │  │  BadgerDB    │ │
│  │  (Indexes)   │  │  (Graph)     │ │
│  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│  Query API      │
│  (Unified)      │
└─────────────────┘
```

---

## SQLite Database Schema

### Tables

#### 1. `nodes` - Node Metadata and Indexed Fields

```sql
CREATE TABLE nodes (
    id INTEGER PRIMARY KEY,           -- Internal uint32 ID
    xref TEXT UNIQUE NOT NULL,       -- GEDCOM XREF (e.g., "@I123@")
    type TEXT NOT NULL,               -- 'individual', 'family', 'note', etc.
    
    -- Indexed fields (for filtering)
    name TEXT,                        -- Full name
    name_lower TEXT,                  -- Lowercase name (for case-insensitive)
    birth_date INTEGER,                -- Unix timestamp (NULL if unknown)
    birth_place TEXT,                 -- Birth place
    sex TEXT,                         -- 'M', 'F', 'U'
    
    -- Boolean flags (for fast filtering)
    has_children INTEGER DEFAULT 0,    -- 0 or 1
    has_spouse INTEGER DEFAULT 0,     -- 0 or 1
    living INTEGER DEFAULT 0,         -- 0 or 1
    
    -- Metadata
    created_at INTEGER,                -- Unix timestamp
    updated_at INTEGER                 -- Unix timestamp
);

-- Indexes for fast lookups
CREATE INDEX idx_nodes_xref ON nodes(xref);
CREATE INDEX idx_nodes_type ON nodes(type);
CREATE INDEX idx_nodes_name_lower ON nodes(name_lower);
CREATE INDEX idx_nodes_birth_date ON nodes(birth_date);
CREATE INDEX idx_nodes_birth_place ON nodes(birth_place);
CREATE INDEX idx_nodes_sex ON nodes(sex);
CREATE INDEX idx_nodes_has_children ON nodes(has_children);
CREATE INDEX idx_nodes_has_spouse ON nodes(has_spouse);
CREATE INDEX idx_nodes_living ON nodes(living);

-- Composite indexes for common queries
CREATE INDEX idx_nodes_name_date ON nodes(name_lower, birth_date);
CREATE INDEX idx_nodes_place_date ON nodes(birth_place, birth_date);
```

#### 2. `nodes_fts` - Full-Text Search

```sql
CREATE VIRTUAL TABLE nodes_fts USING fts5(
    name,           -- Full name
    birth_place,    -- Birth place
    content='nodes', -- Content table
    content_rowid='id' -- Row ID mapping
);

-- Triggers to keep FTS in sync
CREATE TRIGGER nodes_fts_insert AFTER INSERT ON nodes BEGIN
    INSERT INTO nodes_fts(rowid, name, birth_place)
    VALUES (new.id, new.name, new.birth_place);
END;

CREATE TRIGGER nodes_fts_update AFTER UPDATE ON nodes BEGIN
    UPDATE nodes_fts SET name = new.name, birth_place = new.birth_place
    WHERE rowid = new.id;
END;

CREATE TRIGGER nodes_fts_delete AFTER DELETE ON nodes BEGIN
    DELETE FROM nodes_fts WHERE rowid = old.id;
END;
```

#### 3. `xref_mapping` - XREF to ID Mapping

```sql
CREATE TABLE xref_mapping (
    xref TEXT PRIMARY KEY,            -- GEDCOM XREF
    node_id INTEGER NOT NULL,         -- Internal ID
    FOREIGN KEY (node_id) REFERENCES nodes(id)
);

CREATE INDEX idx_xref_mapping_node_id ON xref_mapping(node_id);
```

#### 4. `components` - Connected Components

```sql
CREATE TABLE components (
    component_id INTEGER PRIMARY KEY,
    node_id INTEGER NOT NULL,
    FOREIGN KEY (node_id) REFERENCES nodes(id)
);

CREATE INDEX idx_components_node_id ON components(node_id);
CREATE INDEX idx_components_component_id ON components(component_id);
```

---

## BadgerDB Database Schema

### Key-Value Layout

#### 1. Node Data

```
Key: "node:{nodeID}" (uint32)
Value: Serialized NodeData
  - IndividualNode: Full IndividualRecord data
  - FamilyNode: Full FamilyRecord data
  - NoteNode: Full NoteRecord data
  - etc.
```

**Serialization Format:**
- Use `encoding/gob` or `encoding/json` (gob is faster, smaller)
- Store full node data (not just metadata)

#### 2. Edge Data

```
Key: "edge:{nodeID}:{direction}" (direction: "in" or "out")
Value: Serialized []EdgeData
  - EdgeData: {ToID, EdgeType, FamilyID, Properties}
```

**Alternative (More Efficient):**
```
Key: "edge:{fromID}:{toID}:{edgeType}"
Value: Serialized EdgeData
```

**For fast traversal:**
```
Key: "edges:{nodeID}:in"  → []EdgeData (incoming edges)
Key: "edges:{nodeID}:out" → []EdgeData (outgoing edges)
```

#### 3. Component Data

```
Key: "component:{componentID}"
Value: Serialized []uint32 (node IDs in component)
```

#### 4. Metadata

```
Key: "meta:next_id" → uint32 (next available node ID)
Key: "meta:component_count" → uint32 (number of components)
Key: "meta:node_count" → uint32 (total number of nodes)
```

---

## Query Flow Examples

### Example 1: Filter by Name and Date

```go
// 1. Query SQLite for matching node IDs
query := `
    SELECT id FROM nodes
    WHERE name_lower LIKE ? || '%'
      AND birth_date BETWEEN ? AND ?
    LIMIT 1000
`
rows, _ := db.Query(query, strings.ToLower("John"), startTime, endTime)

var nodeIDs []uint32
for rows.Next() {
    var id uint32
    rows.Scan(&id)
    nodeIDs = append(nodeIDs, id)
}

// 2. Load node data from BadgerDB
var results []*IndividualRecord
for _, nodeID := range nodeIDs {
    key := fmt.Sprintf("node:%d", nodeID)
    data, _ := badgerDB.Get(key)
    node := deserializeNode(data)
    results = append(results, node.Individual)
}
```

### Example 2: Full-Text Search

```go
// 1. Query SQLite FTS5
query := `
    SELECT id FROM nodes_fts
    WHERE name MATCH ?
    LIMIT 1000
`
rows, _ := db.Query(query, "John Smith")

var nodeIDs []uint32
for rows.Next() {
    var id uint32
    rows.Scan(&id)
    nodeIDs = append(nodeIDs, id)
}

// 2. Load from BadgerDB (same as above)
```

### Example 3: Graph Traversal (Parents)

```go
// 1. Get node ID from SQLite (if needed)
var nodeID uint32
db.QueryRow("SELECT id FROM nodes WHERE xref = ?", xrefID).Scan(&nodeID)

// 2. Get edges from BadgerDB
key := fmt.Sprintf("edges:%d:out", nodeID)
data, _ := badgerDB.Get(key)
edges := deserializeEdges(data)

// 3. Filter for parent edges (FAMC)
var parentIDs []uint32
for _, edge := range edges {
    if edge.EdgeType == "FAMC" {
        // Get family node
        famKey := fmt.Sprintf("node:%d", edge.ToID)
        famData, _ := badgerDB.Get(famKey)
        family := deserializeNode(famData)
        
        // Get family's parent edges (HUSB, WIFE)
        famEdgesKey := fmt.Sprintf("edges:%d:out", edge.ToID)
        famEdgesData, _ := badgerDB.Get(famEdgesKey)
        famEdges := deserializeEdges(famEdgesData)
        
        for _, famEdge := range famEdges {
            if famEdge.EdgeType == "HUSB" || famEdge.EdgeType == "WIFE" {
                parentIDs = append(parentIDs, famEdge.ToID)
            }
        }
    }
}

// 4. Load parent nodes from BadgerDB
var parents []*IndividualRecord
for _, parentID := range parentIDs {
    key := fmt.Sprintf("node:%d", parentID)
    data, _ := badgerDB.Get(key)
    node := deserializeNode(data)
    parents = append(parents, node.Individual)
}
```

---

## Implementation Strategy

### Phase 1: Database Setup

1. **Create SQLite Database:**
   - Initialize schema
   - Create indexes
   - Set up FTS5 table

2. **Create BadgerDB Database:**
   - Open BadgerDB instance
   - Configure memory-mapped I/O
   - Set up key prefixes

### Phase 2: Graph Construction

1. **Parse GEDCOM:**
   - Parse file (streaming if possible)
   - Generate node IDs (uint32)

2. **Build SQLite Indexes:**
   - Insert into `nodes` table
   - Insert into `nodes_fts` (via triggers)
   - Insert into `xref_mapping`
   - Build component mapping

3. **Store Graph in BadgerDB:**
   - Serialize and store nodes
   - Serialize and store edges
   - Store components

### Phase 3: Query API Integration

1. **Update FilterQuery:**
   - Query SQLite for node IDs
   - Load nodes from BadgerDB
   - Return results

2. **Update Graph Queries:**
   - Get node ID from SQLite (xref → ID)
   - Load node/edges from BadgerDB
   - Return results

3. **Update Relationship Queries:**
   - Use BadgerDB for graph traversal
   - Use SQLite for metadata lookups

### Phase 4: Caching Layer

1. **Hot Data Cache:**
   - LRU cache for frequently accessed nodes
   - Cache size: 10K-100K nodes
   - Cache in Go map (fast access)

2. **Index Cache:**
   - Cache recent query results
   - Cache node ID lookups (xref → ID)

---

## Data Structures

### NodeData (Serialized in BadgerDB)

```go
type NodeData struct {
    ID        uint32
    Xref      string
    NodeType  string
    Data      []byte  // Serialized IndividualRecord, FamilyRecord, etc.
}
```

### EdgeData (Serialized in BadgerDB)

```go
type EdgeData struct {
    FromID    uint32
    ToID      uint32
    EdgeType  string
    FamilyID  uint32  // For FAMC/FAMS edges
    Direction string  // "forward", "backward", "bidirectional"
    Properties map[string]interface{}
}
```

### HybridGraph Structure

```go
type HybridGraph struct {
    // Databases
    sqliteDB  *sql.DB
    badgerDB  *badger.DB
    
    // Caches
    nodeCache *lru.Cache  // nodeID → *IndividualNode
    xrefCache *lru.Cache  // xref → nodeID
    
    // Metadata
    nextID    uint32
    componentCount uint32
    
    // Thread safety
    mu sync.RWMutex
}
```

---

## Performance Optimizations

### 1. Batch Operations

**SQLite:**
```go
// Batch insert for better performance
tx, _ := db.Begin()
stmt, _ := tx.Prepare("INSERT INTO nodes (...) VALUES (...)")
for _, node := range nodes {
    stmt.Exec(...)
}
tx.Commit()
```

**BadgerDB:**
```go
// Batch write for better performance
txn := db.NewTransaction(true)
for _, node := range nodes {
    key := fmt.Sprintf("node:%d", node.ID)
    data := serializeNode(node)
    txn.Set([]byte(key), data)
}
txn.Commit()
```

### 2. Prepared Statements

```go
// Reuse prepared statements
var (
    getNodeByXref *sql.Stmt
    getNodeByID   *sql.Stmt
    searchByName  *sql.Stmt
)
```

### 3. Connection Pooling

```go
// SQLite connection pool
db.SetMaxOpenConns(10)
db.SetMaxIdleConns(5)
```

### 4. Memory-Mapped I/O

**BadgerDB:**
```go
opts := badger.DefaultOptions(dbPath)
opts.ValueLogLoadingMode = options.MemoryMap  // Memory-mapped
opts.TableLoadingMode = options.MemoryMap
```

**SQLite:**
```sql
PRAGMA mmap_size = 268435456;  -- 256 MB memory-mapped
```

### 5. Index Optimization

**SQLite:**
- Use covering indexes (include all needed columns)
- Analyze tables for query planner
- Use EXPLAIN QUERY PLAN to optimize

**BadgerDB:**
- Use key prefixes for efficient iteration
- Batch reads when possible
- Use iterators for range queries

---

## Memory Management

### Memory Budget

For 10M individuals:
- **SQLite indexes:** ~1-2 GB (memory-mapped, OS handles paging)
- **BadgerDB data:** ~10-15 GB (memory-mapped, OS handles paging)
- **Hot cache:** ~100-200 MB (10K-100K nodes in RAM)
- **Total RAM:** 2-5 GB (only hot data)

### Cache Strategy

1. **Node Cache (LRU):**
   - Size: 10K-100K nodes
   - Evict: Least recently used
   - Store: Full node data

2. **XREF Cache (LRU):**
   - Size: 10K-50K entries
   - Evict: Least recently used
   - Store: xref → nodeID mapping

3. **Query Result Cache:**
   - Size: 1K-10K queries
   - Evict: Least recently used
   - Store: Query → []nodeID

---

## Migration Path

### Step 1: Add Hybrid Mode (Optional)

```go
type Graph struct {
    // ... existing fields ...
    
    // Hybrid storage (optional)
    hybridMode bool
    sqliteDB   *sql.DB
    badgerDB   *badger.DB
}
```

### Step 2: Implement Hybrid Builder

```go
func BuildGraphHybrid(tree *gedcom.GedcomTree, sqlitePath, badgerPath string) (*Graph, error) {
    // 1. Initialize databases
    sqliteDB, _ := sql.Open("sqlite3", sqlitePath)
    badgerDB, _ := badger.Open(badgerOpts)
    
    // 2. Build SQLite indexes
    buildSQLiteIndexes(sqliteDB, tree)
    
    // 3. Build BadgerDB graph
    buildBadgerGraph(badgerDB, tree)
    
    // 4. Create hybrid graph
    return &Graph{
        hybridMode: true,
        sqliteDB:   sqliteDB,
        badgerDB:   badgerDB,
    }, nil
}
```

### Step 3: Update Query Methods

```go
func (g *Graph) GetIndividual(xrefID string) *IndividualNode {
    if g.hybridMode {
        return g.getIndividualHybrid(xrefID)
    }
    // Existing in-memory implementation
    return g.getIndividualMemory(xrefID)
}
```

### Step 4: Gradual Migration

- Keep existing in-memory mode as default
- Add hybrid mode as opt-in
- Test with large datasets
- Switch to hybrid mode for 5M+ individuals

---

## Error Handling

### Database Errors

1. **SQLite Errors:**
   - Connection failures → Retry with backoff
   - Locked database → Wait and retry
   - Corrupted database → Rebuild from BadgerDB

2. **BadgerDB Errors:**
   - Disk full → Return error
   - Corrupted data → Rebuild from SQLite
   - I/O errors → Retry with backoff

### Consistency

1. **Transaction Management:**
   - Use transactions for multi-step operations
   - Rollback on errors
   - Ensure atomicity

2. **Data Validation:**
   - Validate data before storing
   - Check referential integrity
   - Handle missing data gracefully

---

## Testing Strategy

### Unit Tests

1. **Database Operations:**
   - Test SQLite schema creation
   - Test BadgerDB key-value operations
   - Test serialization/deserialization

2. **Query Operations:**
   - Test filter queries
   - Test graph traversal
   - Test relationship queries

### Integration Tests

1. **End-to-End:**
   - Build graph from GEDCOM
   - Run queries
   - Verify results

2. **Performance Tests:**
   - Test with 1M, 5M, 10M individuals
   - Measure memory usage
   - Measure query performance

### Stress Tests

1. **Large Datasets:**
   - Test with 10M+ individuals
   - Monitor memory usage
   - Monitor query performance

2. **Concurrent Access:**
   - Test with multiple goroutines
   - Test read/write contention
   - Test cache eviction

---

## Expected Results

### Memory Usage

- **10M individuals:** 2-5 GB RAM (only hot data)
- **20M individuals:** 3-6 GB RAM
- **50M individuals:** 5-10 GB RAM

### Performance

- **Indexed queries:** 10-100µs (SQLite lookup + BadgerDB load)
- **Graph traversal:** 10-100µs per hop (BadgerDB)
- **Full-text search:** 50-500µs (FTS5)

### Scalability

- **10M+ individuals:** Feasible on 16-32 GB systems
- **50M+ individuals:** Feasible on 32-64 GB systems
- **100M+ individuals:** Feasible on 64-128 GB systems

---

## Next Steps

1. **Design Review:** Review this design document
2. **Prototype:** Build minimal prototype
3. **Benchmark:** Test with 1M individuals
4. **Iterate:** Refine based on results
5. **Implement:** Full implementation
6. **Test:** Comprehensive testing
7. **Deploy:** Gradual rollout

