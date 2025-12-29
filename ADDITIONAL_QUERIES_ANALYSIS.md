# Additional Genealogical Queries Analysis

**Date:** 2025-01-27  
**Purpose:** Analysis of additional common genealogical queries to test across libraries

---

## 1. Relationship Detection Between Two Persons

### Query: "Are persons P1 and P2 related? If so, what is their relationship?"

#### gedcom-go Support: ✅ **YES - Full Support**

**API:**
```go
// Direct relationship calculation
result, err := q.Individual("@I1@").RelationshipTo("@I2@").Execute()

// Or via graph directly
result, err := graph.CalculateRelationship("@I1@", "@I2@")
```

**Capabilities:**
- ✅ Determines if two individuals are related
- ✅ Calculates relationship type (parent, child, sibling, spouse, ancestor, descendant, cousin, uncle, etc.)
- ✅ Calculates degree (for cousins: 1st, 2nd, 3rd, etc.)
- ✅ Calculates removal (for removed cousins: once removed, twice removed, etc.)
- ✅ Finds shortest path between individuals
- ✅ Finds all paths (up to a limit)
- ✅ Distinguishes between:
  - Direct relationships (parent, child, sibling, spouse)
  - Ancestral relationships (ancestor/descendant)
  - Collateral relationships (cousins, uncles, aunts, etc.)
- ✅ Identifies blood vs. marital relationships
- ✅ Uses Lowest Common Ancestor (LCA) for accurate cousin calculations

**Relationship Types Supported:**
- Direct: parent, child, sibling, spouse
- Ancestral: ancestor, descendant (with generation count)
- Collateral: uncle, aunt, nephew, niece, cousin (with degree and removal)
- Special: grandparent, grandchild, great-grandparent, etc.

#### cacack/gedcom-go Support: ❌ **NO - Manual Implementation Required**

**What's Missing:**
- No built-in relationship calculation API
- Would require:
  1. Finding all ancestors of both individuals
  2. Finding common ancestors
  3. Finding the lowest common ancestor
  4. Calculating path lengths
  5. Determining relationship type based on path structure
  6. Manual implementation of cousin degree/removal logic

**Complexity:** High - requires significant custom code

#### elliotchance/gedcom Support: ❌ **NO - Manual Implementation Required**

**What's Missing:**
- No built-in relationship calculation API
- Has `Parents()`, `Spouses()`, `Families()` methods but no relationship calculation
- Would require similar manual implementation as cacack/gedcom-go

**Complexity:** High - requires significant custom code

---

## 2. Finding Oldest/Most Distant Ancestor

### Query: "What is the oldest ancestor (most generations back) for person P?"

#### gedcom-go Support: ⚠️ **PARTIAL - Can Be Implemented**

**Current Capabilities:**
- ✅ Can get all ancestors: `q.Individual("@I1@").Ancestors().Execute()`
- ✅ Can get ancestors with depth: `AncestorQuery.ExecuteWithPaths()` returns `AncestorPath` with depth
- ❌ No direct "oldest ancestor" method

**Implementation Options:**

**Option 1: By Generation Depth**
```go
// Get all ancestors with paths
ancestors, _ := q.Individual("@I1@").Ancestors().ExecuteWithPaths()

// Find ancestor with maximum depth
var oldest *AncestorPath
maxDepth := -1
for _, ancestor := range ancestors {
    if ancestor.Depth > maxDepth {
        maxDepth = ancestor.Depth
        oldest = ancestor
    }
}
```

**Option 2: By Birth Date**
```go
// Get all ancestors
ancestors, _ := q.Individual("@I1@").Ancestors().Execute()

// Find ancestor with earliest birth date
var oldest *types.IndividualRecord
var earliestDate *time.Time
for _, ancestor := range ancestors {
    birthDate := ancestor.GetBirthDate()
    if birthDate != nil && (earliestDate == nil || birthDate.Before(*earliestDate)) {
        earliestDate = birthDate
        oldest = ancestor
    }
}
```

**Recommendation:** Add a convenience method:
```go
// Proposed API
oldest, _ := q.Individual("@I1@").Ancestors().Oldest() // By depth
// or
oldest, _ := q.Individual("@I1@").Ancestors().OldestByBirthDate()
```

#### cacack/gedcom-go Support: ❌ **NO - Manual Implementation Required**

**What's Required:**
- Manual ancestor traversal
- Track depth or birth dates
- Find maximum

**Complexity:** Medium

#### elliotchance/gedcom Support: ❌ **NO - Manual Implementation Required**

**What's Required:**
- Manual ancestor traversal using `Parents()` recursively
- Track depth or birth dates
- Find maximum

**Complexity:** Medium

---

## 3. Other Common Genealogical Queries to Investigate

### 3.1 Common Ancestors

**Query:** "Find all common ancestors of persons P1 and P2"

#### gedcom-go: ✅ **YES**
```go
common, _ := q.Individual("@I1@").CommonAncestors("@I2@")
// or
common, _ := graph.CommonAncestors("@I1@", "@I2@")
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ❌ **NO** - Manual implementation required

---

### 3.2 Lowest Common Ancestor (LCA/MRCA)

**Query:** "Find the most recent common ancestor of persons P1 and P2"

#### gedcom-go: ✅ **YES**
```go
lca, _ := graph.LowestCommonAncestor("@I1@", "@I2@")
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ❌ **NO** - Manual implementation required

---

### 3.3 Path Finding

**Query:** "Find all paths between persons P1 and P2"

#### gedcom-go: ✅ **YES**
```go
// Shortest path
path, _ := q.Individual("@I1@").PathTo("@I2@").Shortest()

// All paths (up to limit)
paths, _ := q.Individual("@I1@").PathTo("@I2@").MaxLength(10).All()
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ❌ **NO** - Manual implementation required

---

### 3.4 Specific Relationship Queries

**Queries:** "Find all cousins", "Find all uncles", "Find all nephews", etc.

#### gedcom-go: ✅ **YES - Multiple Built-in Methods**
```go
cousins, _ := q.Individual("@I1@").Cousins(1)      // 1st cousins
uncles, _ := q.Individual("@I1@").Uncles()
nephews, _ := q.Individual("@I1@").Nephews()
grandparents, _ := q.Individual("@I1@").Grandparents()
grandchildren, _ := q.Individual("@I1@").Grandchildren()
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ❌ **NO** - Manual implementation required

---

### 3.5 Brick Walls (No Known Parents)

**Query:** "Find all individuals with no known parents"

#### gedcom-go: ⚠️ **PARTIAL - Can Be Implemented**
```go
// Filter individuals with no parents
results, _ := q.Filter().
    Where(func(indi *types.IndividualRecord) bool {
        iq := q.Individual(indi.GetXref())
        parents, _ := iq.Parents()
        return len(parents) == 0
    }).
    Execute()
```

**Note:** This is inefficient - would need to check each individual. Better to add:
```go
// Proposed API
brickWalls, _ := q.BrickWalls() // Individuals with no parents
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ❌ **NO** - Manual implementation required

---

### 3.6 End of Line (No Known Children)

**Query:** "Find all individuals with no known children"

#### gedcom-go: ⚠️ **PARTIAL - Can Be Implemented**
```go
results, _ := q.Filter().
    Where(func(indi *types.IndividualRecord) bool {
        iq := q.Individual(indi.GetXref())
        children, _ := iq.Children()
        return len(children) == 0
    }).
    Execute()
```

**Note:** Same efficiency concern as brick walls.

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ❌ **NO** - Manual implementation required

---

### 3.7 Multiple Spouses

**Query:** "Find all individuals with multiple spouses"

#### gedcom-go: ⚠️ **PARTIAL - Can Be Implemented**
```go
results, _ := q.Filter().
    Where(func(indi *types.IndividualRecord) bool {
        iq := q.Individual(indi.GetXref())
        spouses, _ := iq.Spouses()
        return len(spouses) > 1
    }).
    Execute()
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ⚠️ **PARTIAL** - Has `Spouses()` method, can check length

---

### 3.8 Missing Data Queries

**Queries:**
- "Find individuals with no birth date"
- "Find individuals with no death date (potentially living)"
- "Find individuals with no birth place"

#### gedcom-go: ✅ **YES - Via Filter API**
```go
// No birth date
results, _ := q.Filter().
    Where(func(indi *types.IndividualRecord) bool {
        return indi.GetBirthDate() == nil
    }).
    Execute()

// No death date (potentially living)
results, _ := q.Filter().
    Where(func(indi *types.IndividualRecord) bool {
        return indi.GetDeathDate() == nil
    }).
    Execute()
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ⚠️ **PARTIAL** - Can access date fields, but no filter API

---

### 3.9 Data Quality Queries

**Queries:**
- "Find individuals with unusual age gaps between siblings"
- "Find individuals who married too young (< 16)"
- "Find individuals who had children too young (< 13)"
- "Find individuals who lived too long (> 120 years)"

#### gedcom-go: ⚠️ **PARTIAL - Can Be Implemented via Custom Filters**
```go
// Married too young
results, _ := q.Filter().
    Where(func(indi *types.IndividualRecord) bool {
        // Check marriage dates vs birth date
        // Would need to iterate through families
        // Custom logic required
        return false // Placeholder
    }).
    Execute()
```

**Note:** These require complex custom logic and may not be efficiently implementable.

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ⚠️ **PARTIAL** - Has validation/warning system, but no query API

---

### 3.10 Geographic Queries

**Queries:**
- "Find all individuals born in a specific place"
- "Find all individuals who died in a specific place"
- "Find all individuals who lived in a specific place"

#### gedcom-go: ✅ **YES - Via Filter API**
```go
// Born in specific place
results, _ := q.Filter().
    ByBirthPlace("New York").
    Execute()

// Custom place filter
results, _ := q.Filter().
    Where(func(indi *types.IndividualRecord) bool {
        // Check all events for place
        return false // Placeholder
    }).
    Execute()
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ⚠️ **PARTIAL** - Can access place data, but no filter API

---

### 3.11 Temporal Queries

**Queries:**
- "Find all individuals who lived during a specific time period"
- "Find all individuals born in a specific year/decade"
- "Find all individuals who died in a specific year/decade"

#### gedcom-go: ✅ **YES - Via Filter API**
```go
// Born in date range
start := time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(1900, 12, 31, 23, 59, 59, 0, time.UTC)
results, _ := q.Filter().
    ByBirthDate(start, end).
    Execute()
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ⚠️ **PARTIAL** - Can access dates, but no filter API

---

### 3.12 Name-Based Queries

**Queries:**
- "Find all individuals with a specific name"
- "Find all individuals with names starting with 'John'"
- "Find all unique surnames"

#### gedcom-go: ✅ **YES - Via Filter API**
```go
// Exact name
results, _ := q.Filter().ByName("John Smith").Execute()

// Name starts with
results, _ := q.Filter().ByNameStarts("John").Execute()

// Unique names
uniqueNames, _ := q.UniqueNames()
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ⚠️ **PARTIAL** - Can access names, but no filter API

---

### 3.13 Graph Metrics Queries

**Queries:**
- "Find the most connected individual (most relationships)"
- "Find individuals with highest centrality"
- "Find the diameter of the family tree"
- "Find connected components (separate family trees)"

#### gedcom-go: ✅ **YES - Via Metrics API**
```go
metrics := q.Metrics()

// Centrality
centrality, _ := metrics.Centrality(CentralityDegree)

// Diameter
diameter, _ := metrics.Diameter()

// Connected components
components, _ := metrics.ConnectedComponents()
```

#### cacack/gedcom-go: ❌ **NO** - Manual implementation required
#### elliotchance/gedcom: ❌ **NO** - Manual implementation required

---

## Summary Table

| Query Type | gedcom-go | cacack/gedcom-go | elliotchance/gedcom |
|------------|-----------|------------------|---------------------|
| **Relationship Detection** | ✅ Full | ❌ Manual | ❌ Manual |
| **Oldest Ancestor** | ⚠️ Partial | ❌ Manual | ❌ Manual |
| **Common Ancestors** | ✅ Yes | ❌ Manual | ❌ Manual |
| **Lowest Common Ancestor** | ✅ Yes | ❌ Manual | ❌ Manual |
| **Path Finding** | ✅ Yes | ❌ Manual | ❌ Manual |
| **Specific Relationships** | ✅ Yes | ❌ Manual | ❌ Manual |
| **Brick Walls** | ⚠️ Partial | ❌ Manual | ❌ Manual |
| **End of Line** | ⚠️ Partial | ❌ Manual | ❌ Manual |
| **Multiple Spouses** | ⚠️ Partial | ❌ Manual | ⚠️ Partial |
| **Missing Data** | ✅ Yes | ❌ Manual | ⚠️ Partial |
| **Data Quality** | ⚠️ Partial | ❌ Manual | ⚠️ Partial |
| **Geographic** | ✅ Yes | ❌ Manual | ⚠️ Partial |
| **Temporal** | ✅ Yes | ❌ Manual | ⚠️ Partial |
| **Name-Based** | ✅ Yes | ❌ Manual | ⚠️ Partial |
| **Graph Metrics** | ✅ Yes | ❌ Manual | ❌ Manual |

**Legend:**
- ✅ **Yes**: Built-in API support
- ⚠️ **Partial**: Can be implemented but may require custom code or be inefficient
- ❌ **No**: Requires full manual implementation

---

## Recommended Tests to Add

### High Priority (Core Functionality)

1. **Relationship Detection** ⭐⭐⭐
   - Test: Are P1 and P2 related? What's the relationship?
   - **gedcom-go**: Full support via `CalculateRelationship`
   - **Others**: Manual implementation required
   - **Test Cases:**
     - Direct relationships (parent, child, sibling, spouse)
     - Ancestral relationships (grandparent, great-grandparent, etc.)
     - Collateral relationships (cousin, uncle, nephew, etc.)
     - Unrelated individuals
     - Multiple paths between individuals

2. **Oldest Ancestor** ⭐⭐
   - Test: Find the most distant ancestor (by generation depth)
   - **gedcom-go**: Can be implemented using `Ancestors().ExecuteWithPaths()`
   - **Others**: Manual implementation required
   - **Test Cases:**
     - Find ancestor with maximum depth
     - Find ancestor with earliest birth date
     - Handle individuals with no known ancestors

3. **Common Ancestors** ⭐⭐
   - Test: Find all common ancestors of two individuals
   - **gedcom-go**: Full support via `CommonAncestors`
   - **Others**: Manual implementation required

4. **Lowest Common Ancestor** ⭐⭐
   - Test: Find the most recent common ancestor
   - **gedcom-go**: Full support via `LowestCommonAncestor`
   - **Others**: Manual implementation required

### Medium Priority (Useful Features)

5. **Path Finding** ⭐
   - Test: Find all paths between two individuals
   - **gedcom-go**: Full support via `PathTo().All()`
   - **Others**: Manual implementation required

6. **Specific Relationships** ⭐
   - Test: Find cousins, uncles, nephews, etc.
   - **gedcom-go**: Full support via dedicated methods
   - **Others**: Manual implementation required

7. **Brick Walls / End of Line** ⭐
   - Test: Find individuals with no parents/children
   - **gedcom-go**: Can be implemented via filters
   - **Others**: Manual implementation required

### Lower Priority (Nice to Have)

8. **Graph Metrics** 
   - Test: Centrality, diameter, connected components
   - **gedcom-go**: Full support via `Metrics()`
   - **Others**: Manual implementation required

9. **Data Quality Queries**
   - Test: Unusual age gaps, marriage/childbirth age validation
   - **All**: Require custom logic

---

## Implementation Notes

### For cacack/gedcom-go and elliotchance/gedcom

To implement relationship detection manually:

1. **Find if related:**
   - Get all ancestors of P1
   - Get all ancestors of P2
   - Check for intersection
   - If intersection exists, they're related

2. **Calculate relationship:**
   - Find common ancestors
   - Find lowest common ancestor (LCA)
   - Calculate path from P1 to LCA (depth1)
   - Calculate path from P2 to LCA (depth2)
   - Determine relationship type based on depths:
     - If depth1 == 0: P1 is ancestor of P2
     - If depth2 == 0: P2 is ancestor of P1
     - If depth1 == depth2 == 1: siblings
     - If depth1 == depth2 > 1: cousins (degree = depth - 1)
     - If depth1 != depth2: removed cousins
     - etc.

3. **Performance considerations:**
   - Ancestor traversal can be expensive for deep trees
   - Need to handle cycles in data
   - Need to handle multiple paths

---

## Conclusion

**gedcom-go** has significantly more built-in query capabilities than the other libraries:

- ✅ **Relationship detection**: Full support
- ✅ **Common ancestors**: Full support
- ✅ **Path finding**: Full support
- ✅ **Specific relationships**: Full support (cousins, uncles, etc.)
- ✅ **Filtering**: Comprehensive filter API
- ✅ **Graph metrics**: Full support

**cacack/gedcom-go** and **elliotchance/gedcom** require manual implementation for most advanced queries, which:
- Increases development time
- Increases risk of bugs
- May have performance issues
- Requires deep understanding of genealogical relationship calculations

**Recommendation:** Test relationship detection and oldest ancestor queries as these are core genealogical operations that demonstrate the significant advantage of gedcom-go's built-in capabilities.

