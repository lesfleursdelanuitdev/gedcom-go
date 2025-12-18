package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// FamilyQuery represents a query starting from a family.
type FamilyQuery struct {
	xrefID string
	graph  *Graph
}

// Husband returns the husband's individual record.
func (fq *FamilyQuery) Husband() (*gedcom.IndividualRecord, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	if famNode.Husband == nil {
		return nil, nil // No husband
	}

	return famNode.Husband.Individual, nil
}

// Wife returns the wife's individual record.
func (fq *FamilyQuery) Wife() (*gedcom.IndividualRecord, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	if famNode.Wife == nil {
		return nil, nil // No wife
	}

	return famNode.Wife.Individual, nil
}

// Children returns all children's individual records.
func (fq *FamilyQuery) Children() ([]*gedcom.IndividualRecord, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	children := make([]*gedcom.IndividualRecord, 0, len(famNode.Children))
	for _, childNode := range famNode.Children {
		if childNode.Individual != nil {
			children = append(children, childNode.Individual)
		}
	}

	return children, nil
}

// Parents returns the parents (husband and wife) of this family.
func (fq *FamilyQuery) Parents() ([]*gedcom.IndividualRecord, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	parents := make([]*gedcom.IndividualRecord, 0, 2)
	if famNode.Husband != nil && famNode.Husband.Individual != nil {
		parents = append(parents, famNode.Husband.Individual)
	}
	if famNode.Wife != nil && famNode.Wife.Individual != nil {
		parents = append(parents, famNode.Wife.Individual)
	}

	return parents, nil
}

// Events returns all family events (marriage, divorce, etc.).
func (fq *FamilyQuery) Events() ([]*EventNode, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	events := make([]*EventNode, 0)
	for _, edge := range famNode.OutEdges() {
		if edge.EdgeType == EdgeTypeHasEvent {
			if eventNode, ok := edge.To.(*EventNode); ok {
				events = append(events, eventNode)
			}
		}
	}

	return events, nil
}

// MarriageDate returns the marriage date.
func (fq *FamilyQuery) MarriageDate() (string, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return "", fmt.Errorf("family %s not found", fq.xrefID)
	}

	if famNode.Family != nil {
		return famNode.Family.GetMarriageDate(), nil
	}

	return "", nil
}

// MarriageDateParsed returns the marriage date as a parsed GedcomDate.
func (fq *FamilyQuery) MarriageDateParsed() (*gedcom.GedcomDate, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	if famNode.Family != nil {
		return famNode.Family.GetMarriageDateParsed()
	}

	return nil, nil
}

// DivorceDate returns the divorce date.
func (fq *FamilyQuery) DivorceDate() (string, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return "", fmt.Errorf("family %s not found", fq.xrefID)
	}

	if famNode.Family != nil {
		return famNode.Family.GetDivorceDate(), nil
	}

	return "", nil
}

// DivorceDateParsed returns the divorce date as a parsed GedcomDate.
func (fq *FamilyQuery) DivorceDateParsed() (*gedcom.GedcomDate, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	if famNode.Family != nil {
		return famNode.Family.GetDivorceDateParsed()
	}

	return nil, nil
}

// MarriagePlace returns the marriage place.
func (fq *FamilyQuery) MarriagePlace() (string, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return "", fmt.Errorf("family %s not found", fq.xrefID)
	}

	if famNode.Family != nil {
		return famNode.Family.GetMarriagePlace(), nil
	}

	return "", nil
}

// DivorcePlace returns the divorce place.
func (fq *FamilyQuery) DivorcePlace() (string, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return "", fmt.Errorf("family %s not found", fq.xrefID)
	}

	if famNode.Family != nil {
		return famNode.Family.GetDivorcePlace(), nil
	}

	return "", nil
}

// GetRecord returns the underlying FamilyRecord.
func (fq *FamilyQuery) GetRecord() (*gedcom.FamilyRecord, error) {
	famNode := fq.graph.GetFamily(fq.xrefID)
	if famNode == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	return famNode.Family, nil
}

// Exists checks if the family exists.
func (fq *FamilyQuery) Exists() bool {
	return fq.graph.GetFamily(fq.xrefID) != nil
}
