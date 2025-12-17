package query

import (
	"strings"
	"time"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

// Filter represents a filter function for individuals.
type Filter func(*gedcom.IndividualRecord) bool

// FilterQuery represents a query with filtering capabilities.
type FilterQuery struct {
	graph   *Graph
	filters []Filter

	// Indexed filter state
	nameFilter        string
	birthDateStart    *time.Time
	birthDateEnd      *time.Time
	birthPlaceFilter  string
	sexFilter         string
	hasChildrenFilter *bool
	hasSpouseFilter   *bool
	livingFilter      *bool
}

// NewFilterQuery creates a new FilterQuery.
func NewFilterQuery(graph *Graph) *FilterQuery {
	return &FilterQuery{
		graph:   graph,
		filters: make([]Filter, 0),
	}
}

// Where adds a filter condition.
func (fq *FilterQuery) Where(filter Filter) *FilterQuery {
	fq.filters = append(fq.filters, filter)
	return fq
}

// ByName filters by name (case-insensitive substring match).
// Uses index for fast lookup.
func (fq *FilterQuery) ByName(pattern string) *FilterQuery {
	fq.nameFilter = pattern
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		name := strings.ToLower(indi.GetName())
		return strings.Contains(name, strings.ToLower(pattern))
	})
}

// ByBirthDate filters by birth date range.
// Uses index for fast lookup.
func (fq *FilterQuery) ByBirthDate(start, end time.Time) *FilterQuery {
	fq.birthDateStart = &start
	fq.birthDateEnd = &end
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		birthDate, err := indi.GetBirthDateParsed()
		if err != nil || birthDate == nil || !birthDate.IsValid() {
			return false
		}

		birthTime := birthDate.Earliest()
		return !birthTime.Before(start) && !birthTime.After(end)
	})
}

// ByBirthPlace filters by birth place (case-insensitive substring match).
// Uses index for fast lookup.
func (fq *FilterQuery) ByBirthPlace(place string) *FilterQuery {
	fq.birthPlaceFilter = place
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		birthPlace := strings.ToLower(indi.GetBirthPlace())
		return strings.Contains(birthPlace, strings.ToLower(place))
	})
}

// BySex filters by sex.
// Uses index for fast lookup.
func (fq *FilterQuery) BySex(sex string) *FilterQuery {
	fq.sexFilter = sex
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return strings.ToUpper(indi.GetSex()) == strings.ToUpper(sex)
	})
}

// HasChildren filters individuals with children.
// Uses index for fast lookup.
func (fq *FilterQuery) HasChildren() *FilterQuery {
	hasChildren := true
	fq.hasChildrenFilter = &hasChildren
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return fq.graph.indexes.hasChildren(indi.XrefID())
	})
}

// HasSpouse filters individuals with spouses.
// Uses index for fast lookup.
func (fq *FilterQuery) HasSpouse() *FilterQuery {
	hasSpouse := true
	fq.hasSpouseFilter = &hasSpouse
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return fq.graph.indexes.hasSpouse(indi.XrefID())
	})
}

// Living filters living individuals (no death date).
// Uses index for fast lookup.
func (fq *FilterQuery) Living() *FilterQuery {
	living := true
	fq.livingFilter = &living
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return fq.graph.indexes.isLiving(indi.XrefID())
	})
}

// Deceased filters deceased individuals (has death date).
func (fq *FilterQuery) Deceased() *FilterQuery {
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		deathDate := indi.GetDeathDate()
		return deathDate != ""
	})
}

// Execute runs the filter and returns matching individuals.
// Uses indexes for fast filtering when possible.
func (fq *FilterQuery) Execute() ([]*gedcom.IndividualRecord, error) {
	// Build candidate set using indexes
	candidateSet := make(map[string]bool)
	indexes := fq.graph.indexes

	// Start with all individuals
	allIndividuals := fq.graph.GetAllIndividuals()
	initialSet := make(map[string]bool)
	for xrefID := range allIndividuals {
		initialSet[xrefID] = true
	}

	// Apply indexed filters to narrow down candidates
	if fq.nameFilter != "" {
		indexed := indexes.findByName(fq.nameFilter)
		if len(indexed) == 0 {
			return []*gedcom.IndividualRecord{}, nil // No matches
		}
		for _, xrefID := range indexed {
			if initialSet[xrefID] {
				candidateSet[xrefID] = true
			}
		}
		// Update initial set for next filter
		initialSet = candidateSet
		candidateSet = make(map[string]bool)
	}

	if fq.birthDateStart != nil && fq.birthDateEnd != nil {
		indexed := indexes.findByBirthDate(*fq.birthDateStart, *fq.birthDateEnd)
		if len(indexed) == 0 {
			return []*gedcom.IndividualRecord{}, nil // No matches
		}
		indexedSet := make(map[string]bool)
		for _, xrefID := range indexed {
			indexedSet[xrefID] = true
		}
		// Intersect with current set
		for xrefID := range initialSet {
			if indexedSet[xrefID] {
				candidateSet[xrefID] = true
			}
		}
		initialSet = candidateSet
		candidateSet = make(map[string]bool)
	}

	if fq.birthPlaceFilter != "" {
		indexed := indexes.findByBirthPlace(fq.birthPlaceFilter)
		if len(indexed) == 0 {
			return []*gedcom.IndividualRecord{}, nil // No matches
		}
		indexedSet := make(map[string]bool)
		for _, xrefID := range indexed {
			indexedSet[xrefID] = true
		}
		for xrefID := range initialSet {
			if indexedSet[xrefID] {
				candidateSet[xrefID] = true
			}
		}
		initialSet = candidateSet
		candidateSet = make(map[string]bool)
	}

	if fq.sexFilter != "" {
		indexed := indexes.findBySex(fq.sexFilter)
		if len(indexed) == 0 {
			return []*gedcom.IndividualRecord{}, nil // No matches
		}
		indexedSet := make(map[string]bool)
		for _, xrefID := range indexed {
			indexedSet[xrefID] = true
		}
		for xrefID := range initialSet {
			if indexedSet[xrefID] {
				candidateSet[xrefID] = true
			}
		}
		initialSet = candidateSet
		candidateSet = make(map[string]bool)
	}

	if fq.hasChildrenFilter != nil {
		for xrefID := range initialSet {
			if indexes.hasChildren(xrefID) == *fq.hasChildrenFilter {
				candidateSet[xrefID] = true
			}
		}
		initialSet = candidateSet
		candidateSet = make(map[string]bool)
	}

	if fq.hasSpouseFilter != nil {
		for xrefID := range initialSet {
			if indexes.hasSpouse(xrefID) == *fq.hasSpouseFilter {
				candidateSet[xrefID] = true
			}
		}
		initialSet = candidateSet
		candidateSet = make(map[string]bool)
	}

	if fq.livingFilter != nil {
		for xrefID := range initialSet {
			if indexes.isLiving(xrefID) == *fq.livingFilter {
				candidateSet[xrefID] = true
			}
		}
		initialSet = candidateSet
	}

	// If no indexed filters were used, use all individuals
	if len(initialSet) == 0 && fq.nameFilter == "" && fq.birthDateStart == nil &&
		fq.birthPlaceFilter == "" && fq.sexFilter == "" &&
		fq.hasChildrenFilter == nil && fq.hasSpouseFilter == nil && fq.livingFilter == nil {
		for xrefID := range allIndividuals {
			initialSet[xrefID] = true
		}
	}

	// Apply remaining custom filters
	results := make([]*gedcom.IndividualRecord, 0)
	for xrefID := range initialSet {
		node := fq.graph.GetIndividual(xrefID)
		if node == nil || node.Individual == nil {
			continue
		}

		// Apply all filters (indexed filters are already applied via candidate set)
		matches := true
		for _, filter := range fq.filters {
			if !filter(node.Individual) {
				matches = false
				break
			}
		}

		if matches {
			results = append(results, node.Individual)
		}
	}

	return results, nil
}

// Count returns the number of matching individuals.
func (fq *FilterQuery) Count() (int, error) {
	results, err := fq.Execute()
	if err != nil {
		return 0, err
	}
	return len(results), nil
}
