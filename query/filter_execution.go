package query

import (
	"fmt"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// Execute runs the filter and returns matching individuals.
// Uses indexes for fast filtering when possible.
// If hybrid mode is enabled, uses SQLite for lookups.
func (fq *FilterQuery) Execute() ([]*types.IndividualRecord, error) {
	// Record metrics if available
	start := time.Now()
	defer func() {
		if fq.graph.metrics != nil {
			duration := time.Since(start)
			fq.graph.metrics.RecordQuery(duration)
		}
	}()

	// If hybrid mode, use database queries (SQLite or PostgreSQL)
	if fq.graph.hybridMode && (fq.graph.queryHelpers != nil || fq.graph.queryHelpersPostgres != nil) {
		return fq.executeHybrid()
	}

	return fq.executeEager()
}

// executeEager executes the filter query using in-memory indexes
func (fq *FilterQuery) executeEager() ([]*types.IndividualRecord, error) {
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
	// Name filters (only one can be active at a time)
	if fq.nameExactFilter != "" {
		// For exact match, we can use the name index with exact lookup
		indexed := indexes.findByNameExact(fq.nameExactFilter)
		if len(indexed) == 0 {
			return []*types.IndividualRecord{}, nil // No matches
		}
		for _, xrefID := range indexed {
			if initialSet[xrefID] {
				candidateSet[xrefID] = true
			}
		}
		initialSet = candidateSet
		candidateSet = make(map[string]bool)
	} else if fq.nameStartsFilter != "" {
		// For starts with, we can use prefix matching on the index
		indexed := indexes.findByNameStarts(fq.nameStartsFilter)
		if len(indexed) == 0 {
			return []*types.IndividualRecord{}, nil // No matches
		}
		for _, xrefID := range indexed {
			if initialSet[xrefID] {
				candidateSet[xrefID] = true
			}
		}
		initialSet = candidateSet
		candidateSet = make(map[string]bool)
	} else if fq.nameFilter != "" {
		indexed := indexes.findByName(fq.nameFilter)
		if len(indexed) == 0 {
			return []*types.IndividualRecord{}, nil // No matches
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
	// Note: nameEndsFilter is not indexed efficiently, will use Where() filter

	if fq.birthDateStart != nil && fq.birthDateEnd != nil {
		indexed := indexes.findByBirthDate(*fq.birthDateStart, *fq.birthDateEnd)
		if len(indexed) == 0 {
			return []*types.IndividualRecord{}, nil // No matches
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
			return []*types.IndividualRecord{}, nil // No matches
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
			return []*types.IndividualRecord{}, nil // No matches
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
	if len(initialSet) == 0 && fq.nameFilter == "" && fq.nameExactFilter == "" &&
		fq.nameStartsFilter == "" && fq.birthDateStart == nil &&
		fq.birthPlaceFilter == "" && fq.sexFilter == "" &&
		fq.hasChildrenFilter == nil && fq.hasSpouseFilter == nil && fq.livingFilter == nil {
		for xrefID := range allIndividuals {
			initialSet[xrefID] = true
		}
	}

	// Apply remaining custom filters
	results := make([]*types.IndividualRecord, 0)
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

// HybridQueryHelper is an interface for both SQLite and PostgreSQL query helpers
type HybridQueryHelper interface {
	FindByXref(xref string) (uint32, error)
	FindXrefByID(nodeID uint32) (string, error)
	FindByName(pattern string) ([]uint32, error)
	FindByNameExact(name string) ([]uint32, error)
	FindByNameStarts(prefix string) ([]uint32, error)
	FindByBirthDate(start, end time.Time) ([]uint32, error)
	FindByBirthPlace(place string) ([]uint32, error)
	FindBySex(sex string) ([]uint32, error)
	HasChildren(nodeID uint32) (bool, error)
	HasSpouse(nodeID uint32) (bool, error)
	IsLiving(nodeID uint32) (bool, error)
	GetAllIndividualIDs() ([]uint32, error)
}

// executeHybrid executes the filter query using hybrid storage (SQLite + BadgerDB or PostgreSQL + BadgerDB)
func (fq *FilterQuery) executeHybrid() ([]*types.IndividualRecord, error) {
	// Get the appropriate query helpers (SQLite or PostgreSQL)
	var helpers HybridQueryHelper
	if fq.graph.queryHelpersPostgres != nil {
		helpers = fq.graph.queryHelpersPostgres
	} else if fq.graph.queryHelpers != nil {
		helpers = fq.graph.queryHelpers
	} else {
		return nil, fmt.Errorf("no query helpers available")
	}

	var candidateIDs []uint32
	var err error

	// Build cache key for this query
	cacheKey := fq.buildCacheKey()

	// Check cache first
	if fq.graph.hybridCache != nil {
		if cached, found := fq.graph.hybridCache.GetQuery(cacheKey); found {
			candidateIDs = cached
		}
	}

	// If not in cache, query SQLite
	if candidateIDs == nil {
		// Start with all individuals or apply filters
		if fq.nameExactFilter != "" {
			candidateIDs, err = helpers.FindByNameExact(fq.nameExactFilter)
		} else if fq.nameStartsFilter != "" {
			candidateIDs, err = helpers.FindByNameStarts(fq.nameStartsFilter)
		} else if fq.nameFilter != "" {
			candidateIDs, err = helpers.FindByName(fq.nameFilter)
		} else {
			candidateIDs, err = helpers.GetAllIndividualIDs()
		}

		if err != nil {
			return nil, fmt.Errorf("failed to query database: %w", err)
		}

		// Cache the initial result
		if fq.graph.hybridCache != nil {
			fq.graph.hybridCache.SetQuery(cacheKey, candidateIDs)
		}
	}

	// Apply date filter
	if fq.birthDateStart != nil && fq.birthDateEnd != nil {
		dateIDs, err := helpers.FindByBirthDate(*fq.birthDateStart, *fq.birthDateEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to query by birth date: %w", err)
		}
		candidateIDs = intersectIDs(candidateIDs, dateIDs)
	}

	// Apply place filter
	if fq.birthPlaceFilter != "" {
		placeIDs, err := helpers.FindByBirthPlace(fq.birthPlaceFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to query by birth place: %w", err)
		}
		candidateIDs = intersectIDs(candidateIDs, placeIDs)
	}

	// Apply sex filter
	if fq.sexFilter != "" {
		sexIDs, err := helpers.FindBySex(fq.sexFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to query by sex: %w", err)
		}
		candidateIDs = intersectIDs(candidateIDs, sexIDs)
	}

	// Apply boolean filters
	if fq.hasChildrenFilter != nil {
		candidateIDs = filterByBool(candidateIDs, helpers.HasChildren, *fq.hasChildrenFilter)
	}

	if fq.hasSpouseFilter != nil {
		candidateIDs = filterByBool(candidateIDs, helpers.HasSpouse, *fq.hasSpouseFilter)
	}

	if fq.livingFilter != nil {
		candidateIDs = filterByBool(candidateIDs, helpers.IsLiving, *fq.livingFilter)
	}

	// Convert node IDs to XREFs and load nodes
	results := make([]*types.IndividualRecord, 0)
	for _, nodeID := range candidateIDs {
		xref, err := helpers.FindXrefByID(nodeID)
		if err != nil || xref == "" {
			continue
		}

		node := fq.graph.GetIndividual(xref)
		if node == nil || node.Individual == nil {
			continue
		}

		// Apply remaining custom filters
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

	// Cache final result if different from initial
	if fq.graph.hybridCache != nil && len(candidateIDs) != len(results) {
		// Create final cache key with all filters applied
		finalCacheKey := fq.buildCacheKey() + "_final"
		resultIDs := make([]uint32, 0, len(results))
		for _, record := range results {
			xref := record.XrefID()
			if nodeID, found := fq.graph.hybridCache.GetXrefToID(xref); found {
				resultIDs = append(resultIDs, nodeID)
			}
		}
		if len(resultIDs) > 0 {
			fq.graph.hybridCache.SetQuery(finalCacheKey, resultIDs)
		}
	}

	return results, nil
}

// buildCacheKey creates a cache key for the current filter query
func (fq *FilterQuery) buildCacheKey() string {
	key := "filter:"
	if fq.nameExactFilter != "" {
		key += "name_exact:" + fq.nameExactFilter + ":"
	} else if fq.nameStartsFilter != "" {
		key += "name_starts:" + fq.nameStartsFilter + ":"
	} else if fq.nameFilter != "" {
		key += "name:" + fq.nameFilter + ":"
	}
	if fq.birthDateStart != nil && fq.birthDateEnd != nil {
		key += fmt.Sprintf("date:%d-%d:", fq.birthDateStart.Unix(), fq.birthDateEnd.Unix())
	}
	if fq.birthPlaceFilter != "" {
		key += "place:" + fq.birthPlaceFilter + ":"
	}
	if fq.sexFilter != "" {
		key += "sex:" + fq.sexFilter + ":"
	}
	if fq.hasChildrenFilter != nil {
		key += fmt.Sprintf("children:%v:", *fq.hasChildrenFilter)
	}
	if fq.hasSpouseFilter != nil {
		key += fmt.Sprintf("spouse:%v:", *fq.hasSpouseFilter)
	}
	if fq.livingFilter != nil {
		key += fmt.Sprintf("living:%v:", *fq.livingFilter)
	}
	return key
}

// intersectIDs returns the intersection of two ID slices
func intersectIDs(a, b []uint32) []uint32 {
	bSet := make(map[uint32]bool)
	for _, id := range b {
		bSet[id] = true
	}

	var result []uint32
	for _, id := range a {
		if bSet[id] {
			result = append(result, id)
		}
	}
	return result
}

// filterByBool filters IDs by a boolean check function
func filterByBool(ids []uint32, checkFunc func(uint32) (bool, error), want bool) []uint32 {
	var result []uint32
	for _, id := range ids {
		has, err := checkFunc(id)
		if err == nil && has == want {
			result = append(result, id)
		}
	}
	return result
}
