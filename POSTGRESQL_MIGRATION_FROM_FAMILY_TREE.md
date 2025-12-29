# PostgreSQL Migration: Using gedcom_family_tree Connection for gedcom-go

**Date:** 2025-01-27  
**Purpose:** Analyze using the PostgreSQL connection setup from `gedcom_family_tree` to replace SQLite in `gedcom-go`

---

## Executive Summary

**Can we use the connection?** ⚠️ **Partially - same database server, different schemas**

**Can we share the database?** ❌ **Not recommended - different purposes and schemas**

**Can we use the same connection pattern?** ✅ **Yes - same connection string format**

**Recommendation:** 
- ✅ **Use same PostgreSQL server** (if available)
- ❌ **Use separate database** (different schema, different purpose)
- ✅ **Use Go PostgreSQL driver** (`lib/pq` or `pgx`)
- ✅ **Follow same connection string pattern** (`DATABASE_URL`)

---

## 1. Current State Analysis

### 1.1 gedcom_family_tree (PostgreSQL)

**Technology Stack:**
- **Language:** TypeScript/JavaScript
- **ORM:** Prisma
- **Database:** PostgreSQL
- **Connection:** `DATABASE_URL` environment variable
- **Driver:** Prisma Client (uses `pg` under the hood)

**Schema:**
- Authentication: `users`, `sessions`, `verification_tokens`
- GEDCOM: `gedcom_trees`, `records`, `individuals`, `families`, `sources`, etc.
- Relationships: `family_spouses`, `family_children`, `record_associations`, etc.
- Events: `gedcom_events`, `record_events`

**Purpose:** Full GEDCOM data storage with user management

---

### 1.2 gedcom-go (SQLite)

**Technology Stack:**
- **Language:** Go
- **ORM:** Custom (using `database/sql`)
- **Database:** SQLite (in hybrid mode)
- **Connection:** File path
- **Driver:** `github.com/mattn/go-sqlite3`

**Schema:**
- Nodes: `nodes` (id, xref, type, name, birth_date, etc.)
- Mapping: `xref_mapping` (xref → node_id)
- Components: `components` (component_id, node_id)
- Full-text search: `nodes_fts` (FTS5)

**Purpose:** Indexed metadata for fast filtering (graph structure in BadgerDB)

---

## 2. Can We Use the Same Connection?

### 2.1 Connection String Format

**gedcom_family_tree:**
```javascript
// Prisma schema
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

// Connection string format
DATABASE_URL=postgresql://user:password@host:port/database
```

**gedcom-go (if migrated):**
```go
// Go code
import (
    "database/sql"
    _ "github.com/lib/pq"  // PostgreSQL driver
)

// Connection string format (same!)
DATABASE_URL=postgresql://user:password@host:port/database

db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

**Verdict:** ✅ **Same connection string format** - can use same `DATABASE_URL` pattern

---

### 2.2 Database Server

**Can we use the same PostgreSQL server?** ✅ **Yes**

**Should we use the same database?** ❌ **No - different schemas**

**Recommended Approach:**
```
PostgreSQL Server (same instance)
├── Database: gedcom_family_tree  (for Next.js app)
│   ├── users table
│   ├── gedcom_trees table
│   ├── records table
│   └── ...
│
└── Database: ligneous_graphs    (for gedcom-go)
    ├── nodes table (with file_id)
    ├── xref_mapping table (with file_id)
    └── components table (with file_id)
```

**Why Separate Databases:**
- ✅ **Different schemas** - incompatible table structures
- ✅ **Different purposes** - full data vs indexed metadata
- ✅ **Isolation** - easier to manage, backup, migrate
- ✅ **No conflicts** - table names won't clash
- ✅ **Independent scaling** - can optimize separately

---

## 3. Schema Comparison

### 3.1 gedcom_family_tree Schema (Prisma)

**Main Tables:**
- `gedcom_trees` - Top-level tree container
- `records` - All GEDCOM records (base table)
- `individuals` - Individual records
- `families` - Family records
- `sources` - Source records
- `multimedia` - Multimedia records
- `gedcom_events` - Event records
- Edge tables: `family_spouses`, `family_children`, `record_associations`, etc.

**Characteristics:**
- **Full data storage** - stores complete GEDCOM data
- **Normalized** - separate tables for each record type
- **Relationships** - explicit edge tables for relationships
- **User management** - includes authentication tables

---

### 3.2 gedcom-go Schema (SQLite → PostgreSQL)

**Main Tables:**
- `nodes` - Indexed metadata (name, birth_date, birth_place, sex, etc.)
- `xref_mapping` - XREF → node_id mapping
- `components` - Component relationships
- `nodes_fts` - Full-text search (if using PostgreSQL FTS)

**Characteristics:**
- **Indexed metadata only** - not full data storage
- **Denormalized** - single `nodes` table for all types
- **Graph structure** - stored in BadgerDB, not PostgreSQL
- **Fast filtering** - optimized for search queries

---

### 3.3 Schema Incompatibility

**Key Differences:**

| Aspect | gedcom_family_tree | gedcom-go |
|--------|-------------------|-----------|
| **Purpose** | Full data storage | Indexed metadata |
| **Structure** | Normalized (many tables) | Denormalized (few tables) |
| **Graph Storage** | In PostgreSQL | In BadgerDB |
| **Record Types** | Separate tables | Single `nodes` table |
| **Relationships** | Explicit edge tables | In BadgerDB |
| **User Management** | Included | Not included |

**Verdict:** ❌ **Schemas are incompatible** - cannot share same database

---

## 4. Migration Strategy

### 4.1 Option 1: Same Server, Separate Database (Recommended)

**Architecture:**
```
PostgreSQL Server (shared)
├── Database: gedcom_family_tree
│   └── (Prisma schema - Next.js app)
│
└── Database: ligneous_graphs
    └── (gedcom-go schema - Go library)
```

**Connection Strings:**
```bash
# For gedcom_family_tree (Next.js)
DATABASE_URL=postgresql://user:password@localhost:5432/gedcom_family_tree

# For gedcom-go (Go)
DATABASE_URL=postgresql://user:password@localhost:5432/ligneous_graphs
```

**Pros:**
- ✅ Same PostgreSQL server (shared infrastructure)
- ✅ Separate schemas (no conflicts)
- ✅ Independent management
- ✅ Can use same connection pool settings
- ✅ Easy to backup/restore separately

**Cons:**
- ⚠️ Two databases to manage
- ⚠️ Two connection strings

**Verdict:** ✅ **Recommended** - best of both worlds

---

### 4.2 Option 2: Same Server, Same Database, Different Schemas

**Architecture:**
```
PostgreSQL Server
└── Database: gedcom_db
    ├── Schema: family_tree
    │   └── (Prisma tables)
    │
    └── Schema: ligneous
        └── (gedcom-go tables)
```

**Connection Strings:**
```bash
# For gedcom_family_tree
DATABASE_URL=postgresql://user:password@localhost:5432/gedcom_db?search_path=family_tree

# For gedcom-go
DATABASE_URL=postgresql://user:password@localhost:5432/gedcom_db?search_path=ligneous
```

**Pros:**
- ✅ Single database
- ✅ Schema isolation
- ✅ Shared connection pool

**Cons:**
- ⚠️ More complex setup
- ⚠️ Schema management overhead
- ⚠️ Potential for confusion

**Verdict:** ⚠️ **Possible but not recommended** - adds complexity

---

### 4.3 Option 3: Separate Servers

**Architecture:**
```
PostgreSQL Server 1
└── Database: gedcom_family_tree

PostgreSQL Server 2
└── Database: ligneous_graphs
```

**Pros:**
- ✅ Complete isolation
- ✅ Independent scaling
- ✅ No resource contention

**Cons:**
- ❌ More infrastructure
- ❌ Higher cost
- ❌ More complex setup

**Verdict:** ❌ **Overkill** - not needed unless at scale

---

## 5. Implementation Details

### 5.1 Go PostgreSQL Driver Options

**Option 1: lib/pq (Standard)**
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

**Pros:**
- ✅ Standard `database/sql` interface
- ✅ Well-maintained
- ✅ Compatible with existing code

**Cons:**
- ⚠️ Slower than pgx
- ⚠️ Less features

---

**Option 2: pgx (Recommended)**
```go
import (
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/stdlib"
)

// Can use with database/sql
db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))

// OR use pgx directly (better performance)
conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
```

**Pros:**
- ✅ Better performance
- ✅ More features
- ✅ Better error handling
- ✅ Can use with `database/sql` or directly

**Cons:**
- ⚠️ Slightly more complex

**Verdict:** ✅ **Use pgx** - better performance, more features

---

### 5.2 Code Changes Required

**Current Code (SQLite):**
```go
// hybrid_storage.go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func (hs *HybridStorage) initSQLite(config *Config) error {
    db, err := sql.Open("sqlite3", hs.sqlitePath+"?_journal_mode=WAL")
    // ...
}
```

**New Code (PostgreSQL):**
```go
// hybrid_storage.go
import (
    "database/sql"
    "os"
    _ "github.com/jackc/pgx/v5/stdlib"  // PostgreSQL driver
)

func (hs *HybridStorage) initPostgreSQL(config *Config) error {
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        return fmt.Errorf("DATABASE_URL environment variable not set")
    }
    
    db, err := sql.Open("pgx", databaseURL)
    if err != nil {
        return fmt.Errorf("failed to open PostgreSQL: %w", err)
    }
    
    // Set connection pool settings
    db.SetMaxOpenConns(config.Database.PostgreSQLMaxOpenConns)
    db.SetMaxIdleConns(config.Database.PostgreSQLMaxIdleConns)
    
    hs.postgresDB = db
    
    // Create schema
    if err := hs.createPostgreSQLSchema(); err != nil {
        db.Close()
        return fmt.Errorf("failed to create PostgreSQL schema: %w", err)
    }
    
    return nil
}
```

---

### 5.3 Schema Migration

**SQLite Schema:**
```sql
CREATE TABLE nodes (
    id INTEGER PRIMARY KEY,
    xref TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    name TEXT,
    -- ...
);
```

**PostgreSQL Schema (with file_id):**
```sql
CREATE TABLE nodes (
    file_id TEXT NOT NULL,
    id INTEGER NOT NULL,
    xref TEXT NOT NULL,
    type TEXT NOT NULL,
    name TEXT,
    -- ...
    PRIMARY KEY (file_id, id),
    UNIQUE (file_id, xref)
);

CREATE INDEX idx_nodes_file_id ON nodes(file_id);
CREATE INDEX idx_nodes_xref ON nodes(file_id, xref);
-- ... (other indexes include file_id)
```

**Key Changes:**
- ✅ Add `file_id` column to all tables
- ✅ Change `INTEGER PRIMARY KEY` to `INTEGER` (with composite primary key)
- ✅ Change `TEXT` to `VARCHAR` or `TEXT` (PostgreSQL supports both)
- ✅ Update all indexes to include `file_id`
- ✅ Update all queries to filter by `file_id`

---

## 6. Connection Sharing Analysis

### 6.1 Can We Share the Connection?

**Prisma Connection (gedcom_family_tree):**
- Managed by Prisma Client
- Singleton pattern
- Connection pooling handled by Prisma

**Go Connection (gedcom-go):**
- Managed by `database/sql`
- Connection pooling handled by Go
- Can use same connection string

**Verdict:** ✅ **Can use same connection string format, but separate connections**

---

### 6.2 Connection Pool Considerations

**Same Database Server:**
- ✅ Shared connection pool limits
- ✅ Need to coordinate max connections
- ✅ Can optimize together

**Different Databases:**
- ✅ Independent connection pools
- ✅ No coordination needed
- ✅ Can optimize separately

**Recommendation:** Use separate databases to avoid connection pool conflicts

---

## 7. Benefits of Using Same PostgreSQL Server

### 7.1 Infrastructure Benefits

**Shared Server:**
- ✅ **Single PostgreSQL instance** - easier to manage
- ✅ **Shared backups** - can backup both databases together
- ✅ **Shared monitoring** - single point of monitoring
- ✅ **Cost efficiency** - one server instead of two
- ✅ **Easier deployment** - one PostgreSQL setup

---

### 7.2 Operational Benefits

**Same Server:**
- ✅ **Consistent configuration** - same PostgreSQL version, settings
- ✅ **Shared maintenance** - one server to update
- ✅ **Easier troubleshooting** - single point of investigation
- ✅ **Unified logging** - all queries in one place

---

## 8. Challenges and Considerations

### 8.1 Schema Differences

**Challenge:** Completely different schemas

**Solution:**
- ✅ Use separate databases
- ✅ No schema conflicts
- ✅ Independent evolution

---

### 8.2 Connection Management

**Challenge:** Two different applications connecting

**Solution:**
- ✅ Use connection pooling
- ✅ Set appropriate limits
- ✅ Monitor connections

**Example Configuration:**
```go
// gedcom-go
db.SetMaxOpenConns(10)  // Limit connections
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(time.Hour)

// Prisma (gedcom_family_tree)
// Configured in Prisma Client
// Default: 10 connections
```

---

### 8.3 Migration Path

**Challenge:** Migrating from SQLite to PostgreSQL

**Solution:**
1. ✅ Add PostgreSQL support (parallel to SQLite)
2. ✅ Support both (configurable)
3. ✅ Migrate existing files gradually
4. ✅ New files use PostgreSQL

---

## 9. Recommended Architecture

### 9.1 Database Layout

```
PostgreSQL Server (localhost:5432)
│
├── Database: gedcom_family_tree
│   ├── Schema: public
│   │   ├── users
│   │   ├── gedcom_trees
│   │   ├── records
│   │   ├── individuals
│   │   └── ...
│   │
│   └── Purpose: Full GEDCOM data storage (Next.js app)
│
└── Database: ligneous_graphs
    ├── Schema: public
    │   ├── nodes (with file_id)
    │   ├── xref_mapping (with file_id)
    │   ├── components (with file_id)
    │   └── ...
    │
    └── Purpose: Indexed metadata for graph queries (Go library)
```

---

### 9.2 Connection Strings

**Environment Variables:**
```bash
# For gedcom_family_tree (Next.js)
DATABASE_URL=postgresql://user:password@localhost:5432/gedcom_family_tree

# For gedcom-go (Go)
DATABASE_URL=postgresql://user:password@localhost:5432/ligneous_graphs
```

**Or use different variable names:**
```bash
# For gedcom_family_tree
DATABASE_URL=postgresql://user:password@localhost:5432/gedcom_family_tree

# For gedcom-go
LIGNEOUS_DATABASE_URL=postgresql://user:password@localhost:5432/ligneous_graphs
```

---

### 9.3 Implementation Steps

**Phase 1: Setup PostgreSQL**
1. ✅ Install PostgreSQL (if not already installed)
2. ✅ Create `ligneous_graphs` database
3. ✅ Create user with appropriate permissions
4. ✅ Test connection

**Phase 2: Add PostgreSQL Support to gedcom-go**
1. ✅ Add `pgx` driver to `go.mod`
2. ✅ Create `HybridStoragePostgres` type
3. ✅ Implement PostgreSQL schema creation
4. ✅ Update queries to include `file_id`
5. ✅ Add configuration option (SQLite vs PostgreSQL)

**Phase 3: Migration**
1. ✅ Support both SQLite and PostgreSQL
2. ✅ Migrate existing files (optional)
3. ✅ New files use PostgreSQL
4. ✅ Eventually deprecate SQLite (optional)

---

## 10. Code Example

### 10.1 PostgreSQL Storage Implementation

```go
// hybrid_storage_postgres.go
package query

import (
    "database/sql"
    "fmt"
    "os"
    
    _ "github.com/jackc/pgx/v5/stdlib"
)

type HybridStoragePostgres struct {
    postgresDB *sql.DB
    badgerDB   *badger.DB
    badgerPath string
    fileID     string  // Current file ID
}

func NewHybridStoragePostgres(fileID, badgerPath string, config *Config) (*HybridStoragePostgres, error) {
    if config == nil {
        config = DefaultConfig()
    }
    
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL environment variable not set")
    }
    
    hs := &HybridStoragePostgres{
        fileID:     fileID,
        badgerPath: badgerPath,
    }
    
    // Initialize PostgreSQL
    if err := hs.initPostgreSQL(config, databaseURL); err != nil {
        return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
    }
    
    // Initialize BadgerDB
    if err := hs.initBadgerDB(config); err != nil {
        hs.Close()
        return nil, fmt.Errorf("failed to initialize BadgerDB: %w", err)
    }
    
    return hs, nil
}

func (hs *HybridStoragePostgres) initPostgreSQL(config *Config, databaseURL string) error {
    db, err := sql.Open("pgx", databaseURL)
    if err != nil {
        return fmt.Errorf("failed to open PostgreSQL: %w", err)
    }
    
    // Set connection pool settings
    db.SetMaxOpenConns(config.Database.PostgreSQLMaxOpenConns)
    db.SetMaxIdleConns(config.Database.PostgreSQLMaxIdleConns)
    db.SetConnMaxLifetime(config.Database.PostgreSQLConnMaxLifetime)
    
    // Test connection
    if err := db.Ping(); err != nil {
        db.Close()
        return fmt.Errorf("failed to ping PostgreSQL: %w", err)
    }
    
    hs.postgresDB = db
    
    // Create schema
    if err := hs.createPostgreSQLSchema(); err != nil {
        db.Close()
        return fmt.Errorf("failed to create PostgreSQL schema: %w", err)
    }
    
    return nil
}

func (hs *HybridStoragePostgres) createPostgreSQLSchema() error {
    schema := `
    -- Nodes table (with file_id)
    CREATE TABLE IF NOT EXISTS nodes (
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
    CREATE INDEX IF NOT EXISTS idx_nodes_file_id ON nodes(file_id);
    CREATE INDEX IF NOT EXISTS idx_nodes_xref ON nodes(file_id, xref);
    CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(file_id, type);
    CREATE INDEX IF NOT EXISTS idx_nodes_name_lower ON nodes(file_id, name_lower);
    CREATE INDEX IF NOT EXISTS idx_nodes_birth_date ON nodes(file_id, birth_date);
    -- ... (other indexes)
    
    -- XREF mapping table
    CREATE TABLE IF NOT EXISTS xref_mapping (
        file_id TEXT NOT NULL,
        xref TEXT NOT NULL,
        node_id INTEGER NOT NULL,
        PRIMARY KEY (file_id, xref),
        FOREIGN KEY (file_id, node_id) REFERENCES nodes(file_id, id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_xref_mapping_file_id ON xref_mapping(file_id);
    CREATE INDEX IF NOT EXISTS idx_xref_mapping_node_id ON xref_mapping(file_id, node_id);
    
    -- Components table
    CREATE TABLE IF NOT EXISTS components (
        file_id TEXT NOT NULL,
        component_id INTEGER NOT NULL,
        node_id INTEGER NOT NULL,
        PRIMARY KEY (file_id, component_id, node_id),
        FOREIGN KEY (file_id, node_id) REFERENCES nodes(file_id, id)
    );
    
    CREATE INDEX IF NOT EXISTS idx_components_file_id ON components(file_id);
    CREATE INDEX IF NOT EXISTS idx_components_node_id ON components(file_id, node_id);
    `
    
    _, err := hs.postgresDB.Exec(schema)
    return err
}

func (hs *HybridStoragePostgres) Close() error {
    var errs []error
    
    if hs.postgresDB != nil {
        if err := hs.postgresDB.Close(); err != nil {
            errs = append(errs, fmt.Errorf("failed to close PostgreSQL: %w", err))
        }
    }
    
    if hs.badgerDB != nil {
        if err := hs.badgerDB.Close(); err != nil {
            errs = append(errs, fmt.Errorf("failed to close BadgerDB: %w", err))
        }
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("errors closing databases: %v", errs)
    }
    
    return nil
}

func (hs *HybridStoragePostgres) PostgreSQL() *sql.DB {
    return hs.postgresDB
}

func (hs *HybridStoragePostgres) BadgerDB() *badger.DB {
    return hs.badgerDB
}

func (hs *HybridStoragePostgres) FileID() string {
    return hs.fileID
}
```

---

### 10.2 Query Updates

**Current (SQLite):**
```go
func (h *HybridQueryHelpers) FindByXref(xref string) (uint32, error) {
    var nodeID uint32
    err := h.stmtFindByXref.QueryRow(xref).Scan(&nodeID)
    // ...
}
```

**New (PostgreSQL with file_id):**
```go
func (h *HybridQueryHelpersPostgres) FindByXref(fileID string, xref string) (uint32, error) {
    var nodeID uint32
    err := h.stmtFindByXref.QueryRow(fileID, xref).Scan(&nodeID)
    // ...
}
```

---

## 11. Comparison: Shared vs Separate

### 11.1 Same Server, Separate Databases (Recommended)

| Aspect | Benefit |
|--------|---------|
| **Infrastructure** | ✅ Single PostgreSQL server |
| **Management** | ✅ Shared server, separate databases |
| **Isolation** | ✅ Complete schema isolation |
| **Conflicts** | ✅ No table name conflicts |
| **Backup** | ✅ Can backup together or separately |
| **Scaling** | ✅ Can optimize independently |
| **Complexity** | ✅ Moderate (two databases) |

---

### 11.2 Same Server, Same Database, Different Schemas

| Aspect | Benefit |
|--------|---------|
| **Infrastructure** | ✅ Single database |
| **Management** | ⚠️ Schema management overhead |
| **Isolation** | ⚠️ Schema-level isolation |
| **Conflicts** | ✅ No conflicts (different schemas) |
| **Backup** | ✅ Single database backup |
| **Scaling** | ⚠️ Shared optimization |
| **Complexity** | ❌ Higher (schema management) |

---

## 12. Final Recommendation

### 12.1 Architecture

**Use: Same PostgreSQL Server, Separate Databases**

```
PostgreSQL Server (shared)
├── Database: gedcom_family_tree
│   └── Prisma schema (Next.js app)
│
└── Database: ligneous_graphs
    └── gedcom-go schema (Go library)
```

---

### 12.2 Connection Setup

**gedcom_family_tree (Next.js):**
```bash
DATABASE_URL=postgresql://user:password@localhost:5432/gedcom_family_tree
```

**gedcom-go (Go):**
```bash
DATABASE_URL=postgresql://user:password@localhost:5432/ligneous_graphs
# OR
LIGNEOUS_DATABASE_URL=postgresql://user:password@localhost:5432/ligneous_graphs
```

---

### 12.3 Implementation Steps

1. ✅ **Setup PostgreSQL** (if not already available)
2. ✅ **Create `ligneous_graphs` database**
3. ✅ **Add `pgx` driver to gedcom-go**
4. ✅ **Implement PostgreSQL storage** (parallel to SQLite)
5. ✅ **Update queries** (add `file_id` parameter)
6. ✅ **Support both** (SQLite and PostgreSQL configurable)
7. ✅ **Migrate gradually** (optional)

---

## 13. Key Insights

### 13.1 What We Can Reuse

- ✅ **Connection string format** - same `DATABASE_URL` pattern
- ✅ **PostgreSQL server** - can use same instance
- ✅ **Connection pooling** - similar patterns
- ✅ **Best practices** - same principles

---

### 13.2 What We Cannot Reuse

- ❌ **Prisma** - JavaScript/TypeScript only, can't use in Go
- ❌ **Schema** - completely different structures
- ❌ **Same database** - incompatible schemas
- ❌ **ORM** - need Go-specific solution

---

### 13.3 What We Should Do

- ✅ **Use same PostgreSQL server** - shared infrastructure
- ✅ **Use separate databases** - schema isolation
- ✅ **Use Go PostgreSQL driver** - `pgx` recommended
- ✅ **Follow same patterns** - connection string, pooling, etc.

---

## 14. Conclusion

### 14.1 Answer to "Can we use that connection?"

**Connection Setup:** ✅ **Yes - same connection string format**

**Database Server:** ✅ **Yes - can use same PostgreSQL instance**

**Database:** ❌ **No - use separate database (different schema)**

**ORM/Driver:** ❌ **No - need Go driver (pgx), not Prisma**

---

### 14.2 Recommended Approach

**Infrastructure:**
- ✅ Use same PostgreSQL server
- ✅ Create separate database: `ligneous_graphs`
- ✅ Use same connection string pattern

**Implementation:**
- ✅ Add `pgx` driver to gedcom-go
- ✅ Implement PostgreSQL storage (parallel to SQLite)
- ✅ Support both (configurable)
- ✅ Migrate gradually

**Benefits:**
- ✅ Shared infrastructure
- ✅ Schema isolation
- ✅ Independent evolution
- ✅ Better concurrent access
- ✅ Industry standard

---

**Final Verdict:** **Yes, we can use the same PostgreSQL server and connection pattern, but use a separate database for gedcom-go due to incompatible schemas. Use `pgx` driver for Go, not Prisma.**

