# Query Capability Comparison: gedcom-go vs gedcom-elliotchance (gedcomq)

**Date:** 2025-01-27  
**Purpose:** Compare the types of questions each query system can answer

---

## Executive Summary

| Category | gedcom-go Query API | gedcom-elliotchance (gedcomq) | Winner |
|----------|---------------------|-------------------------------|--------|
| **Relationship Queries** | ✅ Comprehensive | ❌ Limited | **gedcom-go** ⭐ |
| **Graph Algorithms** | ✅ Advanced | ❌ None | **gedcom-go** ⭐ |
| **Filtering** | ✅ Rich | ✅ Rich | **Tie** |
| **Data Extraction** | ✅ Good | ✅ Excellent | **gedcomq** ⭐ |
| **Flexibility** | ⚠️ Structured API | ✅ Scriptable | **gedcomq** ⭐ |
| **Performance** | ✅ Optimized | ⚠️ Variable | **gedcom-go** ⭐ |

**Overall Winner:** **gedcom-go** for relationship/graph queries, **gedcomq** for data extraction/flexibility

---

## 1. Relationship Queries

### 1.1 Direct Relationships

**Question:** "Who are the parents/children/siblings/spouses of person X?"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Parents | ✅ `q.Individual("@I1@").Parents()` | ⚠️ Manual traversal | gedcom-go: O(1) cached |
| Children | ✅ `q.Individual("@I1@").Children()` | ⚠️ Manual traversal | gedcom-go: O(1) cached |
| Siblings | ✅ `q.Individual("@I1@").Siblings()` | ⚠️ Manual traversal | gedcom-go: Built-in |
| Spouses | ✅ `q.Individual("@I1@").Spouses()` | ⚠️ Manual traversal | gedcom-go: Built-in |

**Winner:** **gedcom-go** ⭐ (dedicated methods, optimized)

---

### 1.2 Extended Relationships

**Question:** "Who are the grandparents/uncles/cousins of person X?"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Grandparents | ✅ `q.Individual("@I1@").Grandparents()` | ❌ Manual (2 levels) | gedcom-go: Built-in |
| Grandchildren | ✅ `q.Individual("@I1@").Grandchildren()` | ❌ Manual (2 levels) | gedcom-go: Built-in |
| Uncles/Aunts | ✅ `q.Individual("@I1@").Uncles()` | ❌ Manual (complex) | gedcom-go: Built-in |
| Cousins | ✅ `q.Individual("@I1@").Cousins(1)` | ❌ Manual (very complex) | gedcom-go: Degree parameter |
| Nephews/Nieces | ✅ `q.Individual("@I1@").Nephews()` | ❌ Manual (complex) | gedcom-go: Built-in |

**Winner:** **gedcom-go** ⭐ (comprehensive extended relationships)

---

### 1.3 Ancestral/Descendant Queries

**Question:** "Find all ancestors/descendants of person X (with depth limits)"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Ancestors | ✅ `q.Individual("@I1@").Ancestors().MaxGenerations(5).Execute()` | ❌ Manual recursion | gedcom-go: Optimized, cached |
| Descendants | ✅ `q.Individual("@I1@").Descendants().MaxGenerations(3).Execute()` | ❌ Manual recursion | gedcom-go: Optimized, cached |
| Include Self | ✅ `.IncludeSelf()` | ❌ Manual | gedcom-go: Option |
| Filter During Traversal | ✅ `.Filter(func(...) bool)` | ❌ Manual | gedcom-go: Built-in filtering |

**Winner:** **gedcom-go** ⭐ (optimized, configurable)

---

### 1.4 Relationship Calculation

**Question:** "What is the relationship between person X and person Y?"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Calculate Relationship | ✅ `q.Individual("@I1@").RelationshipTo("@I2@").Execute()` | ❌ Not possible | gedcom-go: Full relationship types |
| Relationship Type | ✅ Returns: "father", "cousin", "uncle", etc. | ❌ Not possible | gedcom-go: Human-readable |
| Degree/Removal | ✅ Returns degree (1st, 2nd cousin) and removal | ❌ Not possible | gedcom-go: Precise calculation |
| Is Direct/Collateral | ✅ Returns boolean flags | ❌ Not possible | gedcom-go: Relationship classification |

**Winner:** **gedcom-go** ⭐ (unique capability)

---

### 1.5 Path Finding

**Question:** "Find the path(s) between person X and person Y"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Shortest Path | ✅ `q.Individual("@I1@").PathTo("@I2@").Shortest()` | ❌ Not possible | gedcom-go: BFS algorithm |
| All Paths | ✅ `q.Individual("@I1@").PathTo("@I2@").All()` | ❌ Not possible | gedcom-go: DFS algorithm |
| Max Length | ✅ `.MaxLength(10)` | ❌ Not possible | gedcom-go: Configurable |
| Include Blood/Marital | ✅ `.IncludeBlood(true).IncludeMarital(false)` | ❌ Not possible | gedcom-go: Path type filtering |

**Winner:** **gedcom-go** ⭐ (unique capability)

---

### 1.6 Multi-Individual Queries

**Question:** "Find common ancestors of multiple people"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Common Ancestors | ✅ `q.Individuals("@I1@", "@I2@").CommonAncestors()` | ❌ Not possible | gedcom-go: Set intersection |
| Union of Queries | ✅ `.Union(func1, func2)` | ❌ Not possible | gedcom-go: Flexible composition |
| Ancestors of Multiple | ✅ `q.Individuals("@I1@", "@I2@").Ancestors()` | ❌ Manual | gedcom-go: Union operation |

**Winner:** **gedcom-go** ⭐ (unique capability)

---

## 2. Graph Algorithms & Analytics

### 2.1 Graph Metrics

**Question:** "What are the graph statistics (diameter, density, etc.)?"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Graph Diameter | ✅ `q.Metrics().Diameter()` | ❌ Not possible | gedcom-go: Graph theory |
| Average Path Length | ✅ `q.Metrics().AveragePathLength()` | ❌ Not possible | gedcom-go: Graph theory |
| Graph Density | ✅ `q.Metrics().Density()` | ❌ Not possible | gedcom-go: Graph theory |
| Average Degree | ✅ `q.Metrics().AverageDegree()` | ❌ Not possible | gedcom-go: Graph theory |

**Winner:** **gedcom-go** ⭐ (unique capability)

---

### 2.2 Centrality Measures

**Question:** "Who are the most connected/important people in the tree?"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Degree Centrality | ✅ `q.Metrics().Centrality(CentralityDegree)` | ❌ Not possible | gedcom-go: Graph theory |
| Betweenness Centrality | ✅ `q.Metrics().Centrality(CentralityBetweenness)` | ❌ Not possible | gedcom-go: Graph theory |
| Closeness Centrality | ✅ `q.Metrics().Centrality(CentralityCloseness)` | ❌ Not possible | gedcom-go: Graph theory |
| Node Degree | ✅ `q.Metrics().Degree("@I1@")` | ❌ Not possible | gedcom-go: Individual metrics |

**Winner:** **gedcom-go** ⭐ (unique capability)

---

### 2.3 Connectivity Analysis

**Question:** "Are two people connected? How many connected components?"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Is Connected | ✅ `q.Metrics().IsConnected("@I1@", "@I2@")` | ❌ Not possible | gedcom-go: Graph theory |
| Connected Components | ✅ `q.Metrics().ConnectedComponents()` | ❌ Not possible | gedcom-go: Graph theory |
| Longest Path | ✅ `q.Metrics().LongestPath()` | ❌ Not possible | gedcom-go: Graph theory |
| Lowest Common Ancestor | ✅ `q.Graph().LowestCommonAncestor("@I1@", "@I2@")` | ❌ Not possible | gedcom-go: Graph algorithm |

**Winner:** **gedcom-go** ⭐ (unique capability)

---

### 2.4 Graph Traversal

**Question:** "Traverse the graph in different orders"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| BFS Traversal | ✅ `q.Graph().BFS("@I1@", callback)` | ❌ Not possible | gedcom-go: Breadth-first |
| DFS Traversal | ✅ `q.Graph().DFS("@I1@", callback)` | ❌ Not possible | gedcom-go: Depth-first |
| Custom Traversal | ✅ Callback-based | ❌ Not possible | gedcom-go: Flexible |

**Winner:** **gedcom-go** ⭐ (unique capability)

---

## 3. Filtering & Search

### 3.1 Name Filtering

**Question:** "Find people with name containing 'John'"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| By Name (substring) | ✅ `q.Filter().ByName("John")` | ✅ `.Individuals \| Only(.Name \| .String = "John")` | Both: Similar |
| By Name (exact) | ✅ `q.Filter().ByNameExact("John Smith")` | ✅ `.Individuals \| Only(.Name \| .String = "John Smith")` | Both: Similar |
| By Name (starts) | ✅ `q.Filter().ByNameStarts("John")` | ⚠️ Manual string ops | gedcom-go: Built-in |
| By Name (ends) | ✅ `q.Filter().ByNameEnds("Smith")` | ⚠️ Manual string ops | gedcom-go: Built-in |

**Winner:** **Tie** (both capable, gedcom-go slightly more convenient)

---

### 3.2 Date Filtering

**Question:** "Find people born between 1800-1900"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| By Birth Date Range | ✅ `q.Filter().ByBirthDate(start, end)` | ⚠️ Manual date parsing | gedcom-go: Built-in |
| By Birth Year | ✅ `q.Filter().ByBirthYear(1850)` | ⚠️ Manual | gedcom-go: Built-in |
| By Birth Month/Day | ✅ `q.Filter().ByBirthMonth(12).ByBirthDay(25)` | ⚠️ Manual | gedcom-go: Built-in |
| By Birth Date Before/After | ✅ `q.Filter().ByBirthDateBefore(1900)` | ⚠️ Manual | gedcom-go: Built-in |

**Winner:** **gedcom-go** ⭐ (more date filtering options)

---

### 3.3 Attribute Filtering

**Question:** "Find living males with children"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| By Sex | ✅ `q.Filter().BySex("M")` | ✅ `.Individuals \| Only(.Sex = "M")` | Both: Similar |
| Living | ✅ `q.Filter().Living()` | ✅ `.Individuals \| Only(.IsLiving)` | Both: Similar |
| Deceased | ✅ `q.Filter().Deceased()` | ✅ `.Individuals \| Only(!.IsLiving)` | Both: Similar |
| Has Children | ✅ `q.Filter().HasChildren()` | ⚠️ Manual traversal | gedcom-go: Built-in, indexed |
| Has Spouse | ✅ `q.Filter().HasSpouse()` | ⚠️ Manual traversal | gedcom-go: Built-in, indexed |
| No Children | ✅ `q.Filter().NoChildren()` | ⚠️ Manual | gedcom-go: Built-in |
| No Spouse | ✅ `q.Filter().NoSpouse()` | ⚠️ Manual | gedcom-go: Built-in |

**Winner:** **gedcom-go** ⭐ (more relationship-based filters)

---

### 3.4 Place Filtering

**Question:** "Find people born in New York"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| By Birth Place | ✅ `q.Filter().ByBirthPlace("New York")` | ⚠️ Manual access | gedcom-go: Built-in |
| By Death Place | ❌ Not available | ⚠️ Manual access | gedcomq: Can access via nodes |

**Winner:** **Tie** (gedcom-go has birth place, gedcomq can access any field)

---

### 3.5 Complex Filtering

**Question:** "Find people matching multiple criteria"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Multiple Filters (AND) | ✅ `.ByName("John").BySex("M").HasChildren()` | ✅ `Only(.Name = "John" && .Sex = "M" && ...)` | Both: Similar |
| Custom Filter Function | ✅ `.Where(func(indi) bool { ... })` | ✅ `Only(condition)` | Both: Similar |
| Count Results | ✅ `.Count()` | ✅ `Length` | Both: Similar |
| Exists Check | ✅ `.Exists()` | ⚠️ Manual | gedcom-go: Built-in |

**Winner:** **Tie** (both capable, different syntax)

---

## 4. Data Extraction

### 4.1 Collection Queries

**Question:** "Get all names/places/events in the tree"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| All Names | ✅ `q.Names().All()` | ✅ `.Individuals \| .Name \| .String` | Both: Similar |
| All Places | ✅ `q.Places().All()` | ✅ `.Individuals \| .Birth \| .Place \| .String` | gedcomq: More flexible |
| All Events | ✅ `q.Events().All()` | ✅ `.Individuals \| .AllEvents` | Both: Similar |
| All Families | ✅ `q.Families().All()` | ✅ `.Families` | Both: Similar |
| Unique Values | ✅ `.Unique()` | ⚠️ Manual | gedcom-go: Built-in |
| Group By | ✅ `.By(func)` | ⚠️ Manual | gedcom-go: Built-in |

**Winner:** **gedcomq** ⭐ (more flexible field access)

---

### 4.2 Field Access

**Question:** "Get specific fields from individuals"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Access Any Field | ⚠️ Via record methods | ✅ `.Individuals \| .Property` | gedcomq: More flexible |
| Nested Field Access | ⚠️ Via record methods | ✅ `.Individuals \| .Birth \| .Date \| .String` | gedcomq: Pipe-based |
| Custom Object Creation | ⚠️ Manual | ✅ `{name: .Name, age: .Age}` | gedcomq: Built-in |
| Tag Path Access | ❌ Not available | ✅ `NodesWithTagPath("BIRT", "DATE")` | gedcomq: Unique |

**Winner:** **gedcomq** ⭐ (more flexible data extraction)

---

### 4.3 Aggregation

**Question:** "Count, sum, average values"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Count | ✅ `.Count()` | ✅ `Length` | Both: Similar |
| First N | ⚠️ Manual slice | ✅ `First(n)` | gedcomq: Built-in |
| Last N | ⚠️ Manual slice | ✅ `Last(n)` | gedcomq: Built-in |
| Sum/Average | ❌ Not available | ⚠️ Manual | Neither: Limited |

**Winner:** **gedcomq** ⭐ (more aggregation functions)

---

## 5. Advanced Features

### 5.1 Family Queries

**Question:** "Query starting from a family"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Family Husband | ✅ `q.Family("@F1@").Husband()` | ✅ `.Families \| .Husband` | Both: Similar |
| Family Wife | ✅ `q.Family("@F1@").Wife()` | ✅ `.Families \| .Wife` | Both: Similar |
| Family Children | ✅ `q.Family("@F1@").Children()` | ✅ `.Families \| .Children` | Both: Similar |
| Family Events | ✅ `q.Family("@F1@").Events()` | ✅ `.Families \| .Events` | Both: Similar |
| Marriage Date | ✅ `q.Family("@F1@").MarriageDate()` | ⚠️ Manual access | gedcom-go: Built-in |

**Winner:** **Tie** (both capable)

---

### 5.2 Multi-Document Operations

**Question:** "Merge or compare multiple GEDCOM files"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| Merge Documents | ❌ Not available | ✅ `MergeDocumentsAndIndividuals(doc1, doc2)` | gedcomq: Unique |
| Compare Documents | ✅ `diff` package | ❌ Not available | gedcom-go: Unique (separate package) |

**Winner:** **Tie** (different approaches)

---

### 5.3 Output Formats

**Question:** "Export results in different formats"

| Query | gedcom-go | gedcomq | Notes |
|-------|-----------|---------|-------|
| JSON | ⚠️ Manual | ✅ Built-in JSON formatter | gedcomq: Built-in |
| CSV | ⚠️ Manual | ✅ Built-in CSV formatter | gedcomq: Built-in |
| HTML | ⚠️ Manual | ✅ Built-in HTML formatter | gedcomq: Built-in |
| GEDCOM | ⚠️ Manual | ✅ Built-in GEDCOM formatter | gedcomq: Built-in |
| Pretty JSON | ⚠️ Manual | ✅ Built-in | gedcomq: Built-in |

**Winner:** **gedcomq** ⭐ (multiple built-in formatters)

---

## 6. Query Examples Comparison

### Example 1: "Find all ancestors of person X"

**gedcom-go:**
```go
ancestors, _ := q.Individual("@I1@").Ancestors().MaxGenerations(5).Execute()
```

**gedcomq:**
```bash
# Not possible - requires manual recursion
```

**Winner:** **gedcom-go** ⭐

---

### Example 2: "Find relationship between two people"

**gedcom-go:**
```go
result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()
fmt.Printf("Relationship: %s (degree: %d)\n", result.RelationshipType, result.Degree)
```

**gedcomq:**
```bash
# Not possible
```

**Winner:** **gedcom-go** ⭐

---

### Example 3: "Find all living people named John born in 1850"

**gedcom-go:**
```go
results, _ := q.Filter().
    ByName("John").
    Living().
    ByBirthYear(1850).
    Execute()
```

**gedcomq:**
```bash
.Individuals | Only(.Name | .String = "John" && .IsLiving && .Birth | .Date | .String = "1850")
```

**Winner:** **Tie** (both capable, different syntax)

---

### Example 4: "Get all unique birth places"

**gedcom-go:**
```go
places, _ := q.Places().All()
```

**gedcomq:**
```bash
.Individuals | .Birth | .Place | .String
```

**Winner:** **gedcomq** ⭐ (more flexible, can filter/transform)

---

### Example 5: "Find most connected person (degree centrality)"

**gedcom-go:**
```go
centrality, _ := q.Metrics().Centrality(query.CentralityDegree)
maxDegree := 0.0
mostConnected := ""
for id, degree := range centrality {
    if degree > maxDegree {
        maxDegree = degree
        mostConnected = id
    }
}
```

**gedcomq:**
```bash
# Not possible
```

**Winner:** **gedcom-go** ⭐

---

### Example 6: "Find all cousins of person X"

**gedcom-go:**
```go
cousins, _ := q.Individual("@I1@").Cousins(1) // 1st cousins
```

**gedcomq:**
```bash
# Not possible - requires complex manual traversal
```

**Winner:** **gedcom-go** ⭐

---

### Example 7: "Get JSON of all individuals with name and age"

**gedcom-go:**
```go
individuals, _ := q.AllIndividuals()
results := make([]map[string]interface{}, 0)
for _, indi := range individuals {
    results = append(results, map[string]interface{}{
        "name": indi.GetName(),
        "age": indi.GetAge(),
    })
}
json, _ := json.Marshal(results)
```

**gedcomq:**
```bash
.Individuals | {name: .Name | .String, age: .Age | .String}
```

**Winner:** **gedcomq** ⭐ (much simpler)

---

### Example 8: "Find shortest path between two people"

**gedcom-go:**
```go
path, _ := q.Individual("@I1@").PathTo("@I2@").Shortest()
```

**gedcomq:**
```bash
# Not possible
```

**Winner:** **gedcom-go** ⭐

---

## 7. Summary by Question Type

### Questions gedcom-go CAN answer (but gedcomq CANNOT):

1. ✅ "What is the relationship between person X and person Y?"
2. ✅ "Find all ancestors/descendants of person X (with depth limit)"
3. ✅ "Find all cousins/uncles/grandparents of person X"
4. ✅ "Find the shortest path between person X and person Y"
5. ✅ "Find common ancestors of person X and person Y"
6. ✅ "What is the graph diameter/density?"
7. ✅ "Who are the most connected people (centrality measures)?"
8. ✅ "Are person X and person Y connected in the graph?"
9. ✅ "Find all paths between person X and person Y"
10. ✅ "What is the lowest common ancestor of person X and person Y?"

**Count: 10 unique capabilities**

---

### Questions gedcomq CAN answer (but gedcom-go CANNOT easily):

1. ✅ "Merge two GEDCOM documents"
2. ✅ "Extract data in CSV/HTML format (built-in)"
3. ✅ "Access any nested field via tag paths"
4. ✅ "Create custom objects with any fields"
5. ✅ "Use variables for complex multi-step queries"
6. ✅ "Interactive exploration with `?` function"

**Count: 6 unique capabilities**

---

### Questions BOTH can answer:

1. ✅ "Find people by name/sex/date"
2. ✅ "Get all individuals/families"
3. ✅ "Filter by multiple criteria"
4. ✅ "Count matching results"
5. ✅ "Access family members"
6. ✅ "Get events/dates/places"

**Count: 6 shared capabilities**

---

## 8. Final Verdict

### gedcom-go Query API Strengths:

✅ **Relationship Queries:** Comprehensive (ancestors, descendants, cousins, uncles, etc.)  
✅ **Graph Algorithms:** Advanced (path finding, centrality, connectivity)  
✅ **Performance:** Optimized with caching and indexing  
✅ **Type Safety:** Compile-time checked  
✅ **Relationship Calculation:** Unique capability to calculate relationship types  

### gedcomq Strengths:

✅ **Data Extraction:** More flexible field access  
✅ **Output Formats:** Multiple built-in formatters  
✅ **Scriptability:** Can be used as command-line tool  
✅ **Flexibility:** Can access any nested field  
✅ **Document Merging:** Built-in merge capability  

---

## 9. Recommendation

**Use gedcom-go Query API when:**
- You need relationship queries (ancestors, descendants, cousins, etc.)
- You need graph algorithms (path finding, centrality, connectivity)
- You need relationship calculation between two people
- Performance is critical
- You're building a Go application

**Use gedcomq when:**
- You need flexible data extraction
- You need multiple output formats (CSV, HTML, JSON)
- You're doing ad-hoc queries or data analysis
- You need to merge GEDCOM documents
- You're working from command line

**Answer to "Which can answer more questions?":**

**gedcom-go Query API** can answer **more unique types of questions** (10 vs 6), especially:
- Relationship queries (ancestors, descendants, cousins, uncles, etc.)
- Graph algorithms (path finding, centrality, connectivity)
- Relationship calculation between two people

**gedcomq** is better for:
- Data extraction and transformation
- Output formatting
- Ad-hoc queries
- Document merging

**Overall Winner:** **gedcom-go Query API** for relationship/graph queries, **gedcomq** for data extraction/flexibility.

---

**Comparison Complete** ✅

