# Lazy Graph Construction + Memory-Mapped Storage + Partitioning Design

## Overview

This document outlines the design for implementing:
1. **Lazy/On-Demand Graph Construction** - Build nodes and edges on-demand
2. **Memory-Mapped Files** - Store graph data on disk, let OS handle paging
3. **Graph Partitioning** - Split into connected components, load on-demand

## Architecture

### Graph States

The graph can exist in three states:
1. **Skeleton** - Only node metadata (XREF, type) stored, no nodes/edges loaded
2. **Partial** - Some nodes/edges loaded (on-demand or cached)
3. **Full** - All nodes/edges loaded (current behavior, for compatibility)

### Node Metadata (Skeleton)

For each node, store minimal metadata:
```go
type NodeMetadata struct {
    XrefID   string
    NodeType NodeType
    // Location in memory-mapped file (if using mmap)
    FileOffset uint64
    // Component ID (for partitioning)
    ComponentID uint32
}
```

### Lazy Loading Strategy

**Node Loading:**
- When `GetIndividual(xrefID)` is called:
  1. Check if node is already loaded in memory
  2. If not, load from GEDCOM tree (or mmap file)
  3. Create node and add to graph
  4. Cache in memory

**Edge Loading:**
- When traversing edges (e.g., `OutEdges()`):
  1. Check if edges are already loaded
  2. If not, build edges on-demand from GEDCOM tree
  3. Cache edges for this node
  4. Add to edgeIndex

**Relationship Caching:**
- Keep query cache for frequently accessed relationships
- Cache Parents, Children, Spouses, Siblings when computed
- This is already implemented via query cache

### Memory-Mapped Files

**Storage Format:**
- Use a structured binary format for nodes/edges
- Memory-map the file for random access
- OS handles paging automatically

**File Structure:**
```
[Header]
[Node Metadata Array] - Fixed size entries
[Edge Data Array] - Variable size entries
[String Pool] - Shared strings (names, places)
```

**Implementation:**
- Use `golang.org/x/sys/unix` for mmap on Linux
- Or use `github.com/edsrzf/mmap-go` for cross-platform
- Keep hot nodes in RAM cache
- Cold nodes stay on disk

### Graph Partitioning

**Component Detection:**
- During skeleton creation, identify connected components
- Use BFS/DFS to find all connected nodes
- Assign ComponentID to each node

**Component Storage:**
- Store each component separately (or mark in metadata)
- Load components on-demand
- Cross-component queries require loading multiple components

**Component Loading:**
- When accessing a node, load its entire component
- Cache loaded components
- Unload components when memory pressure

## Implementation Plan

### Phase 1: Lazy Node Loading
1. Add `NodeMetadata` struct
2. Build skeleton during `BuildGraph()`
3. Modify `GetIndividual()`, `GetFamily()`, etc. to load on-demand
4. Cache loaded nodes

### Phase 2: Lazy Edge Loading
1. Modify edge access to build on-demand
2. Cache edges per node
3. Update traversal algorithms to trigger edge loading

### Phase 3: Memory-Mapped Storage
1. Design binary format
2. Implement mmap file operations
3. Store/load nodes from mmap file
4. Keep hot cache in RAM

### Phase 4: Graph Partitioning
1. Implement component detection
2. Store component metadata
3. Load components on-demand
4. Handle cross-component queries

## Benefits

1. **Memory Savings:**
   - Only load what's queried (80-90% savings for typical usage)
   - Memory-mapped files let OS handle paging
   - Partitioning reduces peak memory (70-90% if components are small)

2. **Scalability:**
   - Can handle datasets larger than RAM
   - 5M individuals feasible on 16-32 GB systems

3. **Performance:**
   - Hot data stays in RAM (fast)
   - Cold data on disk (acceptable for genealogy)
   - Relationship caching maintains fast queries

## Trade-offs

1. **Complexity:**
   - More complex codebase
   - Need to handle lazy loading edge cases
   - Memory-mapped files require careful design

2. **Performance:**
   - First access to a node/component is slower
   - Disk I/O for cold data
   - Still fast for hot data (cached)

3. **Compatibility:**
   - Need to maintain backward compatibility
   - Some operations may require full graph

