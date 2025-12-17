package query

import (
	"github.com/yourorg/gedcom/pkg/gedcom"
)

// MultiIndividualQuery represents a query starting from multiple individuals.
type MultiIndividualQuery struct {
	xrefIDs []string
	graph   *Graph
}

// Ancestors finds all ancestors of all individuals in the query.
func (miq *MultiIndividualQuery) Ancestors() ([]*gedcom.IndividualRecord, error) {
	allAncestors := make(map[string]*IndividualNode)

	for _, xrefID := range miq.xrefIDs {
		ancestorQuery := &AncestorQuery{
			startXrefID: xrefID,
			graph:       miq.graph,
			options:     NewAncestorOptions(),
		}

		ancestors, err := ancestorQuery.Execute()
		if err != nil {
			return nil, err
		}

		// Add to map (automatically deduplicates)
		for _, ancestor := range ancestors {
			node := miq.graph.GetIndividual(ancestor.XrefID())
			if node != nil {
				allAncestors[node.ID()] = node
			}
		}
	}

	// Convert to records
	records := make([]*gedcom.IndividualRecord, 0, len(allAncestors))
	for _, node := range allAncestors {
		if node.Individual != nil {
			records = append(records, node.Individual)
		}
	}

	return records, nil
}

// Descendants finds all descendants of all individuals in the query.
func (miq *MultiIndividualQuery) Descendants() ([]*gedcom.IndividualRecord, error) {
	allDescendants := make(map[string]*IndividualNode)

	for _, xrefID := range miq.xrefIDs {
		descendantQuery := &DescendantQuery{
			startXrefID: xrefID,
			graph:       miq.graph,
			options:     NewDescendantOptions(),
		}

		descendants, err := descendantQuery.Execute()
		if err != nil {
			return nil, err
		}

		// Add to map (automatically deduplicates)
		for _, descendant := range descendants {
			node := miq.graph.GetIndividual(descendant.XrefID())
			if node != nil {
				allDescendants[node.ID()] = node
			}
		}
	}

	// Convert to records
	records := make([]*gedcom.IndividualRecord, 0, len(allDescendants))
	for _, node := range allDescendants {
		if node.Individual != nil {
			records = append(records, node.Individual)
		}
	}

	return records, nil
}

// CommonAncestors finds common ancestors across all individuals in the query.
func (miq *MultiIndividualQuery) CommonAncestors() ([]*gedcom.IndividualRecord, error) {
	if len(miq.xrefIDs) == 0 {
		return []*gedcom.IndividualRecord{}, nil
	}

	if len(miq.xrefIDs) == 1 {
		// Single individual - return all their ancestors
		ancestors, err := miq.Ancestors()
		return ancestors, err
	}

	// Find common ancestors of first two
	common, err := miq.graph.CommonAncestors(miq.xrefIDs[0], miq.xrefIDs[1])
	if err != nil {
		return nil, err
	}

	// Intersect with remaining individuals
	for i := 2; i < len(miq.xrefIDs); i++ {
		nextCommon, err := miq.graph.CommonAncestors(miq.xrefIDs[0], miq.xrefIDs[i])
		if err != nil {
			return nil, err
		}

		// Find intersection
		commonMap := make(map[string]*IndividualNode)
		for _, node := range common {
			commonMap[node.ID()] = node
		}

		newCommon := make([]*IndividualNode, 0)
		for _, node := range nextCommon {
			if _, exists := commonMap[node.ID()]; exists {
				newCommon = append(newCommon, node)
			}
		}

		common = newCommon
	}

	// Convert to records
	records := make([]*gedcom.IndividualRecord, 0, len(common))
	for _, node := range common {
		if node.Individual != nil {
			records = append(records, node.Individual)
		}
	}

	return records, nil
}

// Union returns the union of results from multiple individual queries.
// This is useful when you want to combine results from different queries.
func (miq *MultiIndividualQuery) Union(queries ...func(*IndividualQuery) ([]*gedcom.IndividualRecord, error)) ([]*gedcom.IndividualRecord, error) {
	allResults := make(map[string]*gedcom.IndividualRecord)

	for _, xrefID := range miq.xrefIDs {
		indiQuery := &IndividualQuery{
			xrefID: xrefID,
			graph:  miq.graph,
		}

		for _, queryFunc := range queries {
			results, err := queryFunc(indiQuery)
			if err != nil {
				return nil, err
			}

			for _, result := range results {
				allResults[result.XrefID()] = result
			}
		}
	}

	// Convert to slice
	records := make([]*gedcom.IndividualRecord, 0, len(allResults))
	for _, record := range allResults {
		records = append(records, record)
	}

	return records, nil
}

// Intersection returns the intersection of results from multiple individual queries.
func (miq *MultiIndividualQuery) Intersection(queries ...func(*IndividualQuery) ([]*gedcom.IndividualRecord, error)) ([]*gedcom.IndividualRecord, error) {
	if len(queries) == 0 {
		return []*gedcom.IndividualRecord{}, nil
	}

	// Get results from first query for all individuals
	firstResults := make(map[string]bool)
	for _, xrefID := range miq.xrefIDs {
		indiQuery := &IndividualQuery{
			xrefID: xrefID,
			graph:  miq.graph,
		}

		results, err := queries[0](indiQuery)
		if err != nil {
			return nil, err
		}

		for _, result := range results {
			firstResults[result.XrefID()] = true
		}
	}

	// Intersect with remaining queries
	for i := 1; i < len(queries); i++ {
		currentResults := make(map[string]bool)
		for _, xrefID := range miq.xrefIDs {
			indiQuery := &IndividualQuery{
				xrefID: xrefID,
				graph:  miq.graph,
			}

			results, err := queries[i](indiQuery)
			if err != nil {
				return nil, err
			}

			for _, result := range results {
				if firstResults[result.XrefID()] {
					currentResults[result.XrefID()] = true
				}
			}
		}

		firstResults = currentResults
	}

	// Convert to records
	records := make([]*gedcom.IndividualRecord, 0, len(firstResults))
	for xrefID := range firstResults {
		node := miq.graph.GetIndividual(xrefID)
		if node != nil && node.Individual != nil {
			records = append(records, node.Individual)
		}
	}

	return records, nil
}

// Execute returns all individual records in the query.
func (miq *MultiIndividualQuery) Execute() ([]*gedcom.IndividualRecord, error) {
	records := make([]*gedcom.IndividualRecord, 0, len(miq.xrefIDs))

	for _, xrefID := range miq.xrefIDs {
		node := miq.graph.GetIndividual(xrefID)
		if node != nil && node.Individual != nil {
			records = append(records, node.Individual)
		}
	}

	return records, nil
}

// Count returns the number of individuals in the query.
func (miq *MultiIndividualQuery) Count() int {
	return len(miq.xrefIDs)
}
