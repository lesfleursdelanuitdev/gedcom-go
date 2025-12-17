package query

import (
	"fmt"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

// QueryBuilder is the entry point for building queries on a GEDCOM graph.
type QueryBuilder struct {
	graph *Graph
}

// NewQuery creates a new query builder from a GEDCOM tree.
// It builds the graph automatically.
func NewQuery(tree *gedcom.GedcomTree) (*QueryBuilder, error) {
	graph, err := BuildGraph(tree)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	return &QueryBuilder{
		graph: graph,
	}, nil
}

// NewQueryFromGraph creates a new query builder from an existing graph.
func NewQueryFromGraph(graph *Graph) *QueryBuilder {
	return &QueryBuilder{
		graph: graph,
	}
}

// Individual starts a query from a specific individual.
func (qb *QueryBuilder) Individual(xrefID string) *IndividualQuery {
	return &IndividualQuery{
		xrefID: xrefID,
		graph:  qb.graph,
	}
}

// Individuals starts a query from multiple individuals.
func (qb *QueryBuilder) Individuals(xrefIDs ...string) *MultiIndividualQuery {
	return &MultiIndividualQuery{
		xrefIDs: xrefIDs,
		graph:   qb.graph,
	}
}

// AllIndividuals starts a query from all individuals.
func (qb *QueryBuilder) AllIndividuals() *MultiIndividualQuery {
	individuals := qb.graph.GetAllIndividuals()
	xrefIDs := make([]string, 0, len(individuals))
	for id := range individuals {
		xrefIDs = append(xrefIDs, id)
	}
	return &MultiIndividualQuery{
		xrefIDs: xrefIDs,
		graph:   qb.graph,
	}
}

// Filter starts a filter query on all individuals.
func (qb *QueryBuilder) Filter() *FilterQuery {
	return NewFilterQuery(qb.graph)
}

// Family starts a query from a family.
func (qb *QueryBuilder) Family(xrefID string) *FamilyQuery {
	return &FamilyQuery{
		xrefID: xrefID,
		graph:  qb.graph,
	}
}

// Graph returns the internal graph representation for advanced operations.
func (qb *QueryBuilder) Graph() *Graph {
	return qb.graph
}

// Metrics returns a GraphMetricsQuery for graph analytics.
func (qb *QueryBuilder) Metrics() *GraphMetricsQuery {
	return NewGraphMetricsQuery(qb.graph)
}

// IndividualQuery represents a query starting from a specific individual.
type IndividualQuery struct {
	xrefID string
	graph  *Graph
}

// Ancestors finds all ancestors of this individual.
func (iq *IndividualQuery) Ancestors() *AncestorQuery {
	return &AncestorQuery{
		startXrefID: iq.xrefID,
		graph:       iq.graph,
		options:     NewAncestorOptions(),
	}
}

// Descendants finds all descendants of this individual.
func (iq *IndividualQuery) Descendants() *DescendantQuery {
	return &DescendantQuery{
		startXrefID: iq.xrefID,
		graph:       iq.graph,
		options:     NewDescendantOptions(),
	}
}

// Parents returns direct parents.
// Results are cached for repeated queries.
func (iq *IndividualQuery) Parents() ([]*gedcom.IndividualRecord, error) {
	// Check cache
	cacheKey := makeCacheKey("parents", iq.xrefID)
	if cached, ok := iq.graph.cache.get(cacheKey); ok {
		if result, ok := cached.([]*gedcom.IndividualRecord); ok {
			return result, nil
		}
	}

	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	parents := make([]*gedcom.IndividualRecord, 0, len(node.Parents))
	for _, parentNode := range node.Parents {
		if parentNode.Individual != nil {
			parents = append(parents, parentNode.Individual)
		}
	}

	// Cache result
	iq.graph.cache.set(cacheKey, parents)

	return parents, nil
}

// Children returns direct children.
// Results are cached for repeated queries.
func (iq *IndividualQuery) Children() ([]*gedcom.IndividualRecord, error) {
	// Check cache
	cacheKey := makeCacheKey("children", iq.xrefID)
	if cached, ok := iq.graph.cache.get(cacheKey); ok {
		if result, ok := cached.([]*gedcom.IndividualRecord); ok {
			return result, nil
		}
	}

	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	children := make([]*gedcom.IndividualRecord, 0, len(node.Children))
	for _, childNode := range node.Children {
		if childNode.Individual != nil {
			children = append(children, childNode.Individual)
		}
	}

	// Cache result
	iq.graph.cache.set(cacheKey, children)

	return children, nil
}

// Siblings returns siblings (full and half).
func (iq *IndividualQuery) Siblings() ([]*gedcom.IndividualRecord, error) {
	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	siblings := make([]*gedcom.IndividualRecord, 0, len(node.Siblings))
	for _, siblingNode := range node.Siblings {
		if siblingNode.Individual != nil {
			siblings = append(siblings, siblingNode.Individual)
		}
	}

	return siblings, nil
}

// Spouses returns all spouses.
func (iq *IndividualQuery) Spouses() ([]*gedcom.IndividualRecord, error) {
	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	spouses := make([]*gedcom.IndividualRecord, 0, len(node.Spouses))
	for _, spouseNode := range node.Spouses {
		if spouseNode.Individual != nil {
			spouses = append(spouses, spouseNode.Individual)
		}
	}

	return spouses, nil
}

// RelationshipTo finds relationship to another individual.
func (iq *IndividualQuery) RelationshipTo(otherXrefID string) *RelationshipQuery {
	return &RelationshipQuery{
		fromXrefID: iq.xrefID,
		toXrefID:   otherXrefID,
		graph:      iq.graph,
	}
}

// RelationshipTo returns the relationship result directly (convenience method).
func (iq *IndividualQuery) RelationshipToResult(otherXrefID string) (*RelationshipResult, error) {
	return iq.graph.CalculateRelationship(iq.xrefID, otherXrefID)
}

// PathTo finds path(s) to another individual.
func (iq *IndividualQuery) PathTo(otherXrefID string) *PathQuery {
	return &PathQuery{
		fromXrefID: iq.xrefID,
		toXrefID:   otherXrefID,
		graph:      iq.graph,
		options:    NewPathOptions(),
	}
}

// CommonAncestors finds common ancestors with another individual.
func (iq *IndividualQuery) CommonAncestors(otherXrefID string) ([]*gedcom.IndividualRecord, error) {
	nodes, err := iq.graph.CommonAncestors(iq.xrefID, otherXrefID)
	if err != nil {
		return nil, err
	}

	records := make([]*gedcom.IndividualRecord, 0, len(nodes))
	for _, node := range nodes {
		if node.Individual != nil {
			records = append(records, node.Individual)
		}
	}

	return records, nil
}

// Cousins finds cousins (configurable degree).
func (iq *IndividualQuery) Cousins(degree int) ([]*gedcom.IndividualRecord, error) {
	// Get all individuals
	allIndividuals := iq.graph.GetAllIndividuals()
	cousins := make([]*gedcom.IndividualRecord, 0)

	for _, otherNode := range allIndividuals {
		if otherNode.ID() == iq.xrefID {
			continue
		}

		result, err := iq.graph.CalculateRelationship(iq.xrefID, otherNode.ID())
		if err != nil {
			continue
		}

		if result.IsCollateral && result.Degree == degree && result.Removal == 0 {
			if otherNode.Individual != nil {
				cousins = append(cousins, otherNode.Individual)
			}
		}
	}

	return cousins, nil
}

// Uncles finds uncles/aunts.
func (iq *IndividualQuery) Uncles() ([]*gedcom.IndividualRecord, error) {
	// Uncles are siblings of parents
	parents, err := iq.Parents()
	if err != nil {
		return nil, err
	}

	uncles := make([]*gedcom.IndividualRecord, 0)
	for _, parent := range parents {
		parentQuery := iq.graph.GetIndividual(parent.XrefID())
		if parentQuery == nil {
			continue
		}

		for _, siblingNode := range parentQuery.Siblings {
			// Exclude the parent itself
			if siblingNode.ID() != parent.XrefID() && siblingNode.Individual != nil {
				uncles = append(uncles, siblingNode.Individual)
			}
		}
	}

	return uncles, nil
}

// Nephews finds nephews/nieces.
func (iq *IndividualQuery) Nephews() ([]*gedcom.IndividualRecord, error) {
	// Nephews are children of siblings
	siblings, err := iq.Siblings()
	if err != nil {
		return nil, err
	}

	nephews := make([]*gedcom.IndividualRecord, 0)
	for _, sibling := range siblings {
		siblingNode := iq.graph.GetIndividual(sibling.XrefID())
		if siblingNode == nil {
			continue
		}

		for _, childNode := range siblingNode.Children {
			if childNode.Individual != nil {
				nephews = append(nephews, childNode.Individual)
			}
		}
	}

	return nephews, nil
}

// Grandparents returns grandparents.
func (iq *IndividualQuery) Grandparents() ([]*gedcom.IndividualRecord, error) {
	parents, err := iq.Parents()
	if err != nil {
		return nil, err
	}

	grandparents := make([]*gedcom.IndividualRecord, 0)
	for _, parent := range parents {
		parentNode := iq.graph.GetIndividual(parent.XrefID())
		if parentNode == nil {
			continue
		}

		for _, grandparentNode := range parentNode.Parents {
			if grandparentNode.Individual != nil {
				grandparents = append(grandparents, grandparentNode.Individual)
			}
		}
	}

	return grandparents, nil
}

// Grandchildren returns grandchildren.
func (iq *IndividualQuery) Grandchildren() ([]*gedcom.IndividualRecord, error) {
	children, err := iq.Children()
	if err != nil {
		return nil, err
	}

	grandchildren := make([]*gedcom.IndividualRecord, 0)
	for _, child := range children {
		childNode := iq.graph.GetIndividual(child.XrefID())
		if childNode == nil {
			continue
		}

		for _, grandchildNode := range childNode.Children {
			if grandchildNode.Individual != nil {
				grandchildren = append(grandchildren, grandchildNode.Individual)
			}
		}
	}

	return grandchildren, nil
}
