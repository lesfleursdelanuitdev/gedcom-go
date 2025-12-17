package query

import (
	"fmt"
)

// RelationshipResult represents the relationship between two individuals.
type RelationshipResult struct {
	RelationshipType string
	Degree           int
	Removal          int
	Path             *Path
	AllPaths         []*Path
	IsDirect         bool
	IsAncestral      bool
	IsDescendant     bool
	IsCollateral     bool
}

// CalculateRelationship calculates the relationship between two individuals.
func (g *Graph) CalculateRelationship(fromID, toID string) (*RelationshipResult, error) {
	fromNode := g.individuals[fromID]
	toNode := g.individuals[toID]

	if fromNode == nil {
		return nil, fmt.Errorf("individual %s not found", fromID)
	}
	if toNode == nil {
		return nil, fmt.Errorf("individual %s not found", toID)
	}

	result := &RelationshipResult{}

	// Find shortest path
	path, err := g.ShortestPath(fromID, toID)
	if err != nil {
		return nil, err
	}
	result.Path = path

	// Find all paths (limited to reasonable number)
	allPaths, _ := g.AllPaths(fromID, toID, 10)
	result.AllPaths = allPaths

	// Check if direct relationship (parent, child, sibling, spouse)
	result.IsDirect = g.isDirectRelationship(fromNode, toNode)

	// Check if ancestral relationship
	result.IsAncestral = g.isAncestralRelationship(fromID, toID)

	// Check if descendant relationship
	result.IsDescendant = g.isAncestralRelationship(toID, fromID)

	// Check if collateral (cousins, uncles, etc.)
	result.IsCollateral = !result.IsDirect && !result.IsAncestral && !result.IsDescendant

	// Calculate relationship degree and type
	if result.IsDirect {
		result.RelationshipType = g.getDirectRelationshipType(fromNode, toNode)
		result.Degree = 0
		result.Removal = 0
	} else if result.IsAncestral || result.IsDescendant {
		result.RelationshipType = g.getAncestralRelationshipType(result.IsAncestral)
		result.Degree = g.calculateGenerations(path)
		result.Removal = 0
	} else if result.IsCollateral {
		// Find common ancestor
		commonAncestors, _ := g.CommonAncestors(fromID, toID)
		if len(commonAncestors) > 0 {
			lca, _ := g.LowestCommonAncestor(fromID, toID)
			if lca != nil {
				// Calculate degree: generations from LCA to both individuals
				fromDepth := g.getAncestorDepth(fromID, lca.ID())
				toDepth := g.getAncestorDepth(toID, lca.ID())
				result.Degree = min(fromDepth, toDepth) - 1
				result.Removal = abs(fromDepth - toDepth)
				result.RelationshipType = g.getCollateralRelationshipType(result.Degree, result.Removal)
			}
		}
	}

	return result, nil
}

// isDirectRelationship checks if two individuals have a direct relationship.
func (g *Graph) isDirectRelationship(from, to *IndividualNode) bool {
	// Check if parent
	for _, parent := range to.Parents {
		if parent.ID() == from.ID() {
			return true
		}
	}

	// Check if child
	for _, child := range from.Children {
		if child.ID() == to.ID() {
			return true
		}
	}

	// Check if sibling
	for _, sibling := range from.Siblings {
		if sibling.ID() == to.ID() {
			return true
		}
	}

	// Check if spouse
	for _, spouse := range from.Spouses {
		if spouse.ID() == to.ID() {
			return true
		}
	}

	return false
}

// isAncestralRelationship checks if toID is an ancestor of fromID.
func (g *Graph) isAncestralRelationship(fromID, toID string) bool {
	fromNode := g.individuals[fromID]
	if fromNode == nil {
		return false
	}

	// Find all ancestors of fromID
	ancestors := g.findAllAncestors(fromNode, make(map[string]bool))

	// Check if toID is in ancestors
	return ancestors[toID]
}

// getDirectRelationshipType returns the type of direct relationship.
func (g *Graph) getDirectRelationshipType(from, to *IndividualNode) string {
	// Check if parent
	for _, parent := range to.Parents {
		if parent.ID() == from.ID() {
			return "parent"
		}
	}

	// Check if child
	for _, child := range from.Children {
		if child.ID() == to.ID() {
			return "child"
		}
	}

	// Check if sibling
	for _, sibling := range from.Siblings {
		if sibling.ID() == to.ID() {
			return "sibling"
		}
	}

	// Check if spouse
	for _, spouse := range from.Spouses {
		if spouse.ID() == to.ID() {
			return "spouse"
		}
	}

	return "unknown"
}

// getAncestralRelationshipType returns the type of ancestral relationship.
func (g *Graph) getAncestralRelationshipType(isAncestral bool) string {
	if isAncestral {
		return "ancestor"
	}
	return "descendant"
}

// getCollateralRelationshipType returns the type of collateral relationship.
func (g *Graph) getCollateralRelationshipType(degree, removal int) string {
	if degree == 0 && removal == 0 {
		return "sibling"
	}
	if degree == 1 && removal == 0 {
		return "cousin"
	}
	if degree == 1 && removal == 1 {
		return "cousin once removed"
	}
	if degree == 1 && removal > 1 {
		return fmt.Sprintf("cousin %d times removed", removal)
	}
	if degree > 1 && removal == 0 {
		return fmt.Sprintf("%dth cousin", degree)
	}
	if degree > 1 && removal > 0 {
		return fmt.Sprintf("%dth cousin %d times removed", degree, removal)
	}
	return "distant relative"
}

// calculateGenerations counts the number of generations in a path.
func (g *Graph) calculateGenerations(path *Path) int {
	count := 0
	for _, edge := range path.Edges {
		if edge.EdgeType == EdgeTypeFAMC || edge.EdgeType == EdgeTypeCHIL {
			count++
		}
	}
	return count
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
