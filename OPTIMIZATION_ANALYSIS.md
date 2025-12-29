# Performance Optimization Analysis: Making gedcom-go Competitive

**Date:** 2025-01-27  
**Purpose:** Analyze performance bottlenecks and propose optimizations to make gedcom-go competitive with gedcom-go-cacack for ancestor queries

---

## Executive Summary

**Current Performance Gap:** gedcom-go-cacack is **0.16x to 0.97x faster** (16-84% faster) for ancestor queries due to:
1. **Multiple edge traversals** per query
2. **Type assertions** and edge filtering overhead
3. **No direct parent access** - must traverse through family nodes
4. **Graph abstraction overhead** for simple operations

**Key Insight:** We can maintain graph capabilities while adding **fast-path optimizations** for common queries.

---

## Current Implementation Analysis

### Performance Bottlenecks Identified

#### 1. **Multiple Edge Traversals (Critical)**

**Current Flow:**
```
IndividualNode.findAncestors()
  → node.OutEdges()                    // Iterate ALL edges
    → Filter for EdgeTypeFAMC         // Check each edge type
      → edge.Family                    // Get family node
        → famNode.getHusbandFromEdges() // Iterate edges AGAIN
          → Filter for EdgeTypeHUSB    // Check each edge type
        → famNode.getWifeFromEdges()    // Iterate edges AGAIN
          → Filter for EdgeTypeWIFE    // Check each edge type
```

**Cost:** 
- 3 edge iterations per ancestor level
- Type assertions for each edge
- Edge type filtering for each edge

**gedcom-go-cacack Flow:**
```
Individual.findAncestors()
  → ind.ChildInFamilies                // Direct array access
    → doc.GetFamily(xref)              // O(1) map lookup
      → family.Husband                  // Direct field access
      → family.Wife                     // Direct field access
    → doc.GetIndividual(xref)          // O(1) map lookup
```

**Cost:**
- 1 array iteration
- 3 O(1) map lookups
- No type assertions
- No edge filtering

#### 2. **Edge Storage Structure**

**Current:** All edges stored in flat arrays per node
```go
type BaseNode struct {
    inEdges  []*Edge  // All incoming edges
    outEdges []*Edge  // All outgoing edges
}
```

**Problem:** Must iterate all edges to find specific types (FAMC, HUSB, WIFE)

**Solution:** Index edges by type during graph construction

#### 3. **No Direct Parent Caching**

**Current:** Parents computed on-demand via edge traversal
```go
func (node *IndividualNode) getParentsFromEdges() []*IndividualNode {
    // Iterate edges, find FAMC, then iterate family edges...
}
```

**Problem:** Recomputes parents every time, even for repeated queries

**Solution:** Cache parents during graph construction (optional, memory trade-off)

#### 4. **String-based ID Lookups**

**Current:** Uses string XREFs for all lookups
```go
visited := make(map[string]bool)  // String keys
ancestors := make(map[string]*IndividualNode)  // String keys
```

**Problem:** String comparisons are slower than integer comparisons

**Solution:** Use uint32 IDs internally (already available in graph!)

---

## Optimization Strategies

### Strategy 1: Edge Type Indexing (High Impact, Low Risk)

**Concept:** Pre-index edges by type during graph construction

**Implementation:**
```go
type IndividualNode struct {
    *BaseNode
    Individual *types.IndividualRecord
    
    // NEW: Indexed edges for fast access
    famcEdges []*Edge  // Only FAMC edges (parent families)
    famsEdges []*Edge  // Only FAMS edges (spouse families)
}

type FamilyNode struct {
    *BaseNode
    Family *types.FamilyRecord
    
    // NEW: Direct parent references
    husbandEdge *Edge  // Only HUSB edge (or nil)
    wifeEdge    *Edge  // Only WIFE edge (or nil)
    chilEdges   []*Edge  // Only CHIL edges
}
```

**Benefits:**
- **Eliminates edge filtering** - direct access to relevant edges
- **Reduces iterations** - only iterate relevant edges
- **Faster parent access** - O(1) for husband/wife instead of O(n) edge scan

**Performance Gain:** ~2-3x faster for ancestor queries

**Memory Cost:** Minimal (just reorganized edge storage)

**Risk:** Low (doesn't change API, just internal structure)

---

### Strategy 2: Direct Parent Caching (High Impact, Medium Risk)

**Concept:** Cache parent nodes directly on IndividualNode during construction

**Implementation:**
```go
type IndividualNode struct {
    *BaseNode
    Individual *types.IndividualRecord
    
    // NEW: Cached parents (computed during graph construction)
    parents []*IndividualNode  // Direct parent references
    // OR: parentFamilies []*FamilyNode  // Parent families (lighter)
}
```

**Benefits:**
- **Eliminates edge traversal** for parent queries
- **O(1) parent access** instead of O(n) edge scan
- **Faster ancestor queries** - direct parent links

**Performance Gain:** ~3-5x faster for ancestor queries

**Memory Cost:** ~8-16 bytes per individual (2 pointers)

**Risk:** Medium (adds memory, but significant speedup)

**Trade-off:** Can make this optional (configurable during graph construction)

---

### Strategy 3: Fast-Path Ancestor Query (High Impact, Low Risk)

**Concept:** Add optimized ancestor query that uses direct record access for simple cases

**Implementation:**
```go
// Fast-path: Use direct record access when graph overhead isn't needed
func (aq *AncestorQuery) ExecuteFast() ([]*types.IndividualRecord, error) {
    // Use tree.GetIndividual() and record.GetFamiliesAsChild()
    // Similar to gedcom-go-cacack approach
    // Fall back to graph-based query for complex cases
}
```

**Benefits:**
- **Matches gedcom-go-cacack performance** for simple queries
- **Maintains graph capabilities** for complex queries
- **Best of both worlds**

**Performance Gain:** ~5-10x faster for simple ancestor queries

**Memory Cost:** None (uses existing tree structure)

**Risk:** Low (additive, doesn't break existing code)

---

### Strategy 4: Use uint32 IDs Internally (Medium Impact, Low Risk)

**Concept:** Use uint32 IDs for visited tracking and lookups (already available!)

**Current:**
```go
visited := make(map[string]bool)  // String keys
ancestors := make(map[string]*IndividualNode)  // String keys
```

**Optimized:**
```go
visited := make(map[uint32]bool)  // uint32 keys
ancestors := make(map[uint32]*IndividualNode)  // uint32 keys
```

**Benefits:**
- **Faster map lookups** (integer comparison vs string comparison)
- **Less memory** (4 bytes vs ~10-20 bytes per key)
- **Better cache locality**

**Performance Gain:** ~10-20% faster

**Memory Cost:** Negative (saves memory)

**Risk:** Very Low (graph already has ID mapping)

---

### Strategy 5: Pre-filter Edges by Type (Medium Impact, Low Risk)

**Concept:** Store edges in separate slices by type during construction

**Implementation:**
```go
type BaseNode struct {
    // Existing
    inEdges  []*Edge
    outEdges []*Edge
    
    // NEW: Type-indexed edges (optional, for performance)
    edgesByType map[EdgeType][]*Edge  // Indexed by edge type
}

func (bn *BaseNode) GetEdgesByType(edgeType EdgeType) []*Edge {
    if bn.edgesByType != nil {
        return bn.edgesByType[edgeType]  // O(1) lookup
    }
    // Fallback: filter from outEdges
    return bn.filterEdgesByType(edgeType)
}
```

**Benefits:**
- **O(1) edge type lookup** instead of O(n) filtering
- **Reduces iterations** in ancestor queries

**Performance Gain:** ~1.5-2x faster

**Memory Cost:** Small (one map per node)

**Risk:** Low (can be optional/conditional)

---

### Strategy 6: Batch Edge Processing (Low Impact, Medium Risk)

**Concept:** Process multiple ancestor queries in batch to amortize overhead

**Benefits:**
- **Amortizes graph overhead** across multiple queries
- **Better cache utilization**

**Performance Gain:** ~10-15% for batch queries

**Risk:** Medium (requires API changes)

**Recommendation:** Lower priority, focus on single-query optimizations first

---

## Recommended Implementation Plan

### Phase 1: Quick Wins (High Impact, Low Risk)

**Priority 1: Edge Type Indexing**
- Add `famcEdges`, `famsEdges` to IndividualNode
- Add `husbandEdge`, `wifeEdge`, `chilEdges` to FamilyNode
- Update graph builder to populate indexed edges
- Update ancestor query to use indexed edges

**Expected Gain:** 2-3x faster

**Priority 2: uint32 ID Usage**
- Change visited/ancestors maps to use uint32 keys
- Use graph's existing ID mapping

**Expected Gain:** 10-20% faster

**Combined Expected Gain:** 2.5-3.5x faster (should match or beat gedcom-go-cacack)

---

### Phase 2: Advanced Optimizations (High Impact, Medium Risk)

**Priority 3: Direct Parent Caching**
- Add `parents []*IndividualNode` to IndividualNode
- Populate during graph construction
- Make it optional (config flag)

**Expected Gain:** Additional 1.5-2x faster (total 4-7x faster)

**Priority 4: Fast-Path Query**
- Add `ExecuteFast()` method that uses direct record access
- Auto-select based on query complexity

**Expected Gain:** 5-10x faster for simple queries

---

### Phase 3: Fine-Tuning (Medium Impact, Low Risk)

**Priority 5: Edge Type Pre-filtering**
- Add `edgesByType` map to BaseNode
- Use for other relationship queries too

**Expected Gain:** Additional 1.5x faster

---

## Detailed Implementation Notes

### Edge Type Indexing Implementation

**In `builder.go` (createFamilyEdges):**
```go
// When creating FAMC edge, also add to node's famcEdges
indiNode.famcEdges = append(indiNode.famcEdges, edge)

// When creating HUSB edge, store directly
famNode.husbandEdge = edge

// When creating WIFE edge, store directly
famNode.wifeEdge = edge
```

**In `ancestor_query.go` (findAncestors):**
```go
// OLD: for _, edge := range node.OutEdges() {
// NEW: for _, edge := range node.famcEdges {
    if edge.Family != nil {
        famNode := edge.Family
        // OLD: husband := famNode.getHusbandFromEdges()
        // NEW: husband := famNode.husbandEdge.To.(*IndividualNode)
        // OLD: wife := famNode.getWifeFromEdges()
        // NEW: wife := famNode.wifeEdge.To.(*IndividualNode)
    }
}
```

**Benefits:**
- Eliminates edge filtering loops
- Direct access to relevant edges
- ~2-3x faster

---

### Direct Parent Caching Implementation

**In `builder.go` (after createFamilyEdges):**
```go
// Populate parent cache for all individuals
for xrefID, indiNode := range graph.individuals {
    parents := make([]*IndividualNode, 0, 2)
    for _, edge := range indiNode.famcEdges {
        if edge.Family != nil {
            famNode := edge.Family
            if famNode.husbandEdge != nil {
                parents = append(parents, famNode.husbandEdge.To.(*IndividualNode))
            }
            if famNode.wifeEdge != nil {
                parents = append(parents, famNode.wifeEdge.To.(*IndividualNode))
            }
        }
    }
    indiNode.parents = parents
}
```

**In `ancestor_query.go` (findAncestors):**
```go
// OLD: Traverse edges to find parents
// NEW: Direct parent access
for _, parent := range node.parents {
    ancestors[parent.ID()] = parent
    aq.findAncestors(parent, ancestors, visited, depth+1)
}
```

**Benefits:**
- Eliminates all edge traversal for parent access
- ~3-5x faster

---

### uint32 ID Usage Implementation

**In `ancestor_query.go`:**
```go
// OLD:
visited := make(map[string]bool)
ancestors := make(map[string]*IndividualNode)

// NEW:
visited := make(map[uint32]bool)
ancestors := make(map[uint32]*IndividualNode)

// Use graph's ID mapping
nodeID := aq.graph.getID(node.ID())
visited[nodeID] = true
ancestors[nodeID] = node
```

**Benefits:**
- Faster map operations
- Less memory
- ~10-20% faster

---

## Performance Projections

### Current Performance (baseline)
- **royal92.ged @I1@ (unlimited):** ~300,000 ns
- **pres2020.ged @I100@ (unlimited):** ~1,000 ns

### After Phase 1 (Edge Indexing + uint32 IDs)
- **royal92.ged @I1@ (unlimited):** ~100,000-120,000 ns (2.5-3x faster)
- **pres2020.ged @I100@ (unlimited):** ~300-400 ns (2.5-3x faster)
- **Status:** Should match or beat gedcom-go-cacack

### After Phase 2 (Parent Caching)
- **royal92.ged @I1@ (unlimited):** ~50,000-70,000 ns (4-6x faster)
- **pres2020.ged @I100@ (unlimited):** ~150-200 ns (5-7x faster)
- **Status:** Should significantly outperform gedcom-go-cacack

---

## Memory Impact Analysis

### Edge Type Indexing
- **Cost:** ~0 bytes (just reorganizing existing edges)
- **Benefit:** Significant performance gain

### Direct Parent Caching
- **Cost:** ~8-16 bytes per individual (2 pointers)
- **For 1M individuals:** ~8-16 MB
- **Benefit:** 3-5x performance gain
- **Trade-off:** Worth it for performance-critical applications

### uint32 ID Usage
- **Cost:** Negative (saves memory)
- **Benefit:** 10-20% performance gain

**Total Memory Cost:** ~8-16 MB for 1M individuals (negligible for performance gain)

---

## Risk Assessment

### Low Risk Optimizations
1. ✅ **Edge Type Indexing** - Internal change, no API impact
2. ✅ **uint32 ID Usage** - Internal change, no API impact
3. ✅ **Edge Type Pre-filtering** - Optional, can be conditional

### Medium Risk Optimizations
1. ⚠️ **Direct Parent Caching** - Adds memory, but configurable
2. ⚠️ **Fast-Path Query** - New API method, but additive

### Mitigation Strategies
- Make parent caching **optional** (config flag)
- Add **benchmarks** to verify improvements
- Maintain **backward compatibility** (existing API unchanged)
- Add **feature flags** for gradual rollout

---

## Implementation Priority

### Must Have (Phase 1)
1. **Edge Type Indexing** - Biggest impact, lowest risk
2. **uint32 ID Usage** - Easy win, significant gain

**Timeline:** 1-2 days  
**Expected Result:** Match or beat gedcom-go-cacack performance

### Should Have (Phase 2)
3. **Direct Parent Caching** - High impact, configurable
4. **Fast-Path Query** - Best of both worlds

**Timeline:** 2-3 days  
**Expected Result:** Significantly outperform gedcom-go-cacack

### Nice to Have (Phase 3)
5. **Edge Type Pre-filtering** - Additional optimization
6. **Batch Processing** - For specific use cases

**Timeline:** 1-2 days  
**Expected Result:** Further improvements

---

## Testing Strategy

### Performance Benchmarks
- Run existing `ancestor_benchmark_comparison_test.go`
- Compare before/after for each optimization
- Verify results match (same ancestor counts)

### Regression Testing
- Ensure all existing tests pass
- Verify graph integrity after optimizations
- Test with various dataset sizes

### Memory Profiling
- Measure memory usage before/after
- Verify memory trade-offs are acceptable
- Profile with large datasets (1M+ individuals)

---

## Conclusion

**Key Insights:**
1. **Edge traversal is the bottleneck** - Multiple iterations per query
2. **Direct access beats graph traversal** for simple queries
3. **We can have both** - Fast simple queries + powerful graph capabilities

**Recommended Approach:**
1. **Start with Phase 1** (Edge Indexing + uint32 IDs)
   - Should match gedcom-go-cacack performance
   - Low risk, high impact
2. **Add Phase 2** if needed (Parent Caching)
   - Should outperform gedcom-go-cacack
   - Configurable memory trade-off
3. **Keep graph capabilities** for complex queries
   - Best of both worlds

**Expected Outcome:**
- **Phase 1:** Match gedcom-go-cacack (2.5-3x faster)
- **Phase 2:** Outperform gedcom-go-cacack (4-7x faster)
- **Maintain:** All graph capabilities for complex queries

---

**Analysis Complete** ✅

