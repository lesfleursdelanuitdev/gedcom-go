package query

// Parents returns all parents of this individual.
// Computes parents from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *IndividualNode) Parents() []*IndividualNode {
	return node.getParentsFromEdges()
}

// getParentsFromEdges computes parents from edges (replaces cached Parents field).
// Phase 1: Optimized to use indexed edges. Phase 2: Uses cached parents if available.
func (node *IndividualNode) getParentsFromEdges() []*IndividualNode {
	// Phase 2: Use cached parents for O(1) access (fastest)
	if len(node.parents) > 0 {
		return node.parents
	}

	// Phase 1: Use indexed FAMC edges (no filtering needed)
	parents := make([]*IndividualNode, 0, 2)
	seen := make(map[string]bool)

	for _, edge := range node.famcEdges {
		if edge.Family != nil {
			famNode := edge.Family
			// Phase 1: Use indexed edges for O(1) access
			if famNode.husbandEdge != nil {
				if indiNode, ok := famNode.husbandEdge.To.(*IndividualNode); ok {
					if !seen[indiNode.ID()] {
						seen[indiNode.ID()] = true
						parents = append(parents, indiNode)
					}
				}
			}
			if famNode.wifeEdge != nil {
				if indiNode, ok := famNode.wifeEdge.To.(*IndividualNode); ok {
					if !seen[indiNode.ID()] {
						seen[indiNode.ID()] = true
						parents = append(parents, indiNode)
					}
				}
			}
		}
	}

	return parents
}

// Children returns all children of this individual.
// Computes children from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *IndividualNode) Children() []*IndividualNode {
	return node.getChildrenFromEdges()
}

// getChildrenFromEdges computes children from edges (replaces cached Children field).
// Phase 1: Optimized to use indexed edges for faster access.
func (node *IndividualNode) getChildrenFromEdges() []*IndividualNode {
	children := make([]*IndividualNode, 0)
	seen := make(map[string]bool)

	// Phase 1: Use indexed FAMS edges (no filtering needed)
	for _, edge := range node.famsEdges {
		if edge.Family != nil {
			famNode := edge.Family
			// Phase 1: Use indexed CHIL edges (no filtering needed)
			for _, famEdge := range famNode.chilEdges {
				if indiNode, ok := famEdge.To.(*IndividualNode); ok {
					if !seen[indiNode.ID()] {
						seen[indiNode.ID()] = true
						children = append(children, indiNode)
					}
				}
			}
		}
	}

	return children
}

// Spouses returns all spouses of this individual.
// Computes spouses from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *IndividualNode) Spouses() []*IndividualNode {
	return node.getSpousesFromEdges()
}

// getSpousesFromEdges computes spouses from edges (replaces cached Spouses field).
// Phase 1: Optimized to use indexed edges for faster access.
func (node *IndividualNode) getSpousesFromEdges() []*IndividualNode {
	spouses := make([]*IndividualNode, 0)
	seen := make(map[string]bool)

	// Phase 1: Use indexed FAMS edges (no filtering needed)
	for _, edge := range node.famsEdges {
		if edge.Family != nil {
			famNode := edge.Family
			// Phase 1: Use indexed edges for O(1) access
			if famNode.husbandEdge != nil {
				if indiNode, ok := famNode.husbandEdge.To.(*IndividualNode); ok {
					if indiNode.ID() != node.ID() && !seen[indiNode.ID()] {
						seen[indiNode.ID()] = true
						spouses = append(spouses, indiNode)
					}
				}
			}
			if famNode.wifeEdge != nil {
				if indiNode, ok := famNode.wifeEdge.To.(*IndividualNode); ok {
					if indiNode.ID() != node.ID() && !seen[indiNode.ID()] {
						seen[indiNode.ID()] = true
						spouses = append(spouses, indiNode)
					}
				}
			}
		}
	}

	return spouses
}

// Siblings returns all siblings of this individual.
// Computes siblings from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *IndividualNode) Siblings() []*IndividualNode {
	return node.getSiblingsFromEdges()
}

// getSiblingsFromEdges computes siblings from edges (replaces cached Siblings field).
// Phase 1: Optimized to use indexed edges for faster access.
func (node *IndividualNode) getSiblingsFromEdges() []*IndividualNode {
	siblings := make([]*IndividualNode, 0)
	seen := make(map[string]bool)

	// Phase 1: Use indexed FAMC edges (no filtering needed)
	parentFamilies := make(map[string]*FamilyNode)
	for _, edge := range node.famcEdges {
		if edge.Family != nil {
			parentFamilies[edge.Family.ID()] = edge.Family
		}
	}

	// Phase 1: Use indexed CHIL edges (no filtering needed)
	for _, famNode := range parentFamilies {
		for _, famEdge := range famNode.chilEdges {
			if indiNode, ok := famEdge.To.(*IndividualNode); ok {
				if indiNode.ID() != node.ID() && !seen[indiNode.ID()] {
					seen[indiNode.ID()] = true
					siblings = append(siblings, indiNode)
				}
			}
		}
	}

	return siblings
}

// Husband returns the husband of this family.
// Computes husband from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *FamilyNode) Husband() *IndividualNode {
	return node.getHusbandFromEdges()
}

// getHusbandFromEdges computes husband from edges (replaces cached Husband field).
// Phase 1: Optimized to use indexed edge for O(1) access.
func (node *FamilyNode) getHusbandFromEdges() *IndividualNode {
	// Phase 1: Use indexed edge for O(1) access
	if node.husbandEdge != nil {
		if indiNode, ok := node.husbandEdge.To.(*IndividualNode); ok {
			return indiNode
		}
	}
	// Fallback: Search through all edges (shouldn't happen if graph is properly built)
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeHUSB {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				return indiNode
			}
		}
	}
	return nil
}

// Wife returns the wife of this family.
// Computes wife from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *FamilyNode) Wife() *IndividualNode {
	return node.getWifeFromEdges()
}

// getWifeFromEdges computes wife from edges (replaces cached Wife field).
// Phase 1: Optimized to use indexed edge for O(1) access.
func (node *FamilyNode) getWifeFromEdges() *IndividualNode {
	// Phase 1: Use indexed edge for O(1) access
	if node.wifeEdge != nil {
		if indiNode, ok := node.wifeEdge.To.(*IndividualNode); ok {
			return indiNode
		}
	}
	// Fallback: Search through all edges (shouldn't happen if graph is properly built)
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeWIFE {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				return indiNode
			}
		}
	}
	return nil
}

// Children returns all children of this family.
// Computes children from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *FamilyNode) Children() []*IndividualNode {
	return node.getChildrenFromEdges()
}

// getChildrenFromEdges computes children from edges (replaces cached Children field).
// Phase 1: Optimized to use indexed edges for faster access.
func (node *FamilyNode) getChildrenFromEdges() []*IndividualNode {
	// Phase 1: Use indexed edges - no filtering needed
	children := make([]*IndividualNode, 0, len(node.chilEdges))
	for _, edge := range node.chilEdges {
		if indiNode, ok := edge.To.(*IndividualNode); ok {
			children = append(children, indiNode)
		}
	}
	return children
}

