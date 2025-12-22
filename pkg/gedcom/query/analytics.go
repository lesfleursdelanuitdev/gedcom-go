package query

import (
	"sort"
	"strings"
)

// NameCount represents a name with its occurrence count.
type NameCount struct {
	Name  string
	Count int
}

// GetMostCommonGivenNames returns the most common first names (given names).
// Returns top N names sorted by count (descending).
func (g *Graph) GetMostCommonGivenNames(limit int) []NameCount {
	if limit <= 0 {
		limit = 10
	}
	
	nameCounts := make(map[string]int)
	
	allIndividuals := g.GetAllIndividuals()
	for _, indiNode := range allIndividuals {
		if indiNode.Individual != nil {
			givenName := indiNode.Individual.GetGivenName()
			if givenName != "" {
				// Normalize: lowercase for counting
				normalized := strings.ToLower(strings.TrimSpace(givenName))
				nameCounts[normalized]++
			}
		}
	}
	
	// Convert to slice and sort
	result := make([]NameCount, 0, len(nameCounts))
	for name, count := range nameCounts {
		result = append(result, NameCount{
			Name:  name,
			Count: count,
		})
	}
	
	// Sort by count (descending), then by name (ascending)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Name < result[j].Name
	})
	
	// Return top N
	if len(result) > limit {
		return result[:limit]
	}
	return result
}

// GetMostCommonSurnames returns the most common last names (surnames).
// Returns top N names sorted by count (descending).
func (g *Graph) GetMostCommonSurnames(limit int) []NameCount {
	if limit <= 0 {
		limit = 10
	}
	
	nameCounts := make(map[string]int)
	
	allIndividuals := g.GetAllIndividuals()
	for _, indiNode := range allIndividuals {
		if indiNode.Individual != nil {
			surname := indiNode.Individual.GetSurname()
			if surname != "" {
				// Normalize: lowercase for counting
				normalized := strings.ToLower(strings.TrimSpace(surname))
				nameCounts[normalized]++
			}
		}
	}
	
	// Convert to slice and sort
	result := make([]NameCount, 0, len(nameCounts))
	for name, count := range nameCounts {
		result = append(result, NameCount{
			Name:  name,
			Count: count,
		})
	}
	
	// Sort by count (descending), then by name (ascending)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Name < result[j].Name
	})
	
	// Return top N
	if len(result) > limit {
		return result[:limit]
	}
	return result
}

