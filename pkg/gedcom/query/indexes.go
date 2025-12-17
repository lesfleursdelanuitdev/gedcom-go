package query

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

// FilterIndexes contains indexes for fast filtering.
type FilterIndexes struct {
	mu sync.RWMutex

	// Name index: lowercase name -> []xrefID
	nameIndex map[string][]string

	// Date index: sorted by birth date
	birthDateIndex []*dateIndexEntry

	// Place index: lowercase place -> []xrefID
	placeIndex map[string][]string

	// Sex index: sex -> []xrefID
	sexIndex map[string][]string

	// Has children index: xrefID -> bool
	hasChildrenIndex map[string]bool

	// Has spouse index: xrefID -> bool
	hasSpouseIndex map[string]bool

	// Living index: xrefID -> bool
	livingIndex map[string]bool
}

// dateIndexEntry represents an entry in the date index.
type dateIndexEntry struct {
	xrefID    string
	birthDate *gedcom.GedcomDate
}

// newFilterIndexes creates a new FilterIndexes.
func newFilterIndexes() *FilterIndexes {
	return &FilterIndexes{
		nameIndex:        make(map[string][]string),
		birthDateIndex:   make([]*dateIndexEntry, 0),
		placeIndex:       make(map[string][]string),
		sexIndex:         make(map[string][]string),
		hasChildrenIndex: make(map[string]bool),
		hasSpouseIndex:   make(map[string]bool),
		livingIndex:      make(map[string]bool),
	}
}

// buildIndexes builds all indexes from the graph.
func (fi *FilterIndexes) buildIndexes(graph *Graph) {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	// Clear existing indexes
	fi.nameIndex = make(map[string][]string)
	fi.birthDateIndex = make([]*dateIndexEntry, 0)
	fi.placeIndex = make(map[string][]string)
	fi.sexIndex = make(map[string][]string)
	fi.hasChildrenIndex = make(map[string]bool)
	fi.hasSpouseIndex = make(map[string]bool)
	fi.livingIndex = make(map[string]bool)

	individuals := graph.GetAllIndividuals()

	// Build indexes
	for xrefID, node := range individuals {
		if node.Individual == nil {
			continue
		}

		indi := node.Individual

		// Name index
		name := strings.ToLower(indi.GetName())
		if name != "" {
			fi.nameIndex[name] = append(fi.nameIndex[name], xrefID)
			// Also index by individual words
			words := strings.Fields(name)
			for _, word := range words {
				if len(word) > 2 { // Only index words longer than 2 chars
					fi.nameIndex[word] = append(fi.nameIndex[word], xrefID)
				}
			}
		}

		// Birth date index
		birthDate, err := indi.GetBirthDateParsed()
		if err == nil && birthDate != nil && birthDate.IsValid() {
			fi.birthDateIndex = append(fi.birthDateIndex, &dateIndexEntry{
				xrefID:    xrefID,
				birthDate: birthDate,
			})
		}

		// Place index
		birthPlace := strings.ToLower(indi.GetBirthPlace())
		if birthPlace != "" {
			fi.placeIndex[birthPlace] = append(fi.placeIndex[birthPlace], xrefID)
			// Also index by individual words
			words := strings.Fields(birthPlace)
			for _, word := range words {
				if len(word) > 2 {
					fi.placeIndex[word] = append(fi.placeIndex[word], xrefID)
				}
			}
		}

		// Sex index
		sex := strings.ToUpper(indi.GetSex())
		if sex != "" {
			fi.sexIndex[sex] = append(fi.sexIndex[sex], xrefID)
		}

		// Has children index
		fi.hasChildrenIndex[xrefID] = len(node.Children) > 0

		// Has spouse index
		fi.hasSpouseIndex[xrefID] = len(node.Spouses) > 0

		// Living index
		fi.livingIndex[xrefID] = indi.GetDeathDate() == ""
	}

	// Sort birth date index
	sort.Slice(fi.birthDateIndex, func(i, j int) bool {
		dateI := fi.birthDateIndex[i].birthDate.Earliest()
		dateJ := fi.birthDateIndex[j].birthDate.Earliest()
		return dateI.Before(dateJ)
	})
}

// findByName finds individuals by name pattern (case-insensitive substring).
func (fi *FilterIndexes) findByName(pattern string) []string {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	patternLower := strings.ToLower(pattern)
	resultSet := make(map[string]bool)

	// Check exact matches and word matches
	for key, xrefIDs := range fi.nameIndex {
		if strings.Contains(key, patternLower) {
			for _, xrefID := range xrefIDs {
				resultSet[xrefID] = true
			}
		}
	}

	result := make([]string, 0, len(resultSet))
	for xrefID := range resultSet {
		result = append(result, xrefID)
	}

	return result
}

// findByBirthDate finds individuals by birth date range.
func (fi *FilterIndexes) findByBirthDate(start, end time.Time) []string {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	result := make([]string, 0)

	// Binary search for start date
	startIdx := sort.Search(len(fi.birthDateIndex), func(i int) bool {
		date := fi.birthDateIndex[i].birthDate.Earliest()
		return !date.Before(start)
	})

	// Linear scan from start to end
	for i := startIdx; i < len(fi.birthDateIndex); i++ {
		entry := fi.birthDateIndex[i]
		date := entry.birthDate.Earliest()
		if date.After(end) {
			break
		}
		result = append(result, entry.xrefID)
	}

	return result
}

// findByBirthPlace finds individuals by birth place pattern (case-insensitive substring).
func (fi *FilterIndexes) findByBirthPlace(place string) []string {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	placeLower := strings.ToLower(place)
	resultSet := make(map[string]bool)

	for key, xrefIDs := range fi.placeIndex {
		if strings.Contains(key, placeLower) {
			for _, xrefID := range xrefIDs {
				resultSet[xrefID] = true
			}
		}
	}

	result := make([]string, 0, len(resultSet))
	for xrefID := range resultSet {
		result = append(result, xrefID)
	}

	return result
}

// findBySex finds individuals by sex.
func (fi *FilterIndexes) findBySex(sex string) []string {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	sexUpper := strings.ToUpper(sex)
	return fi.sexIndex[sexUpper]
}

// hasChildren checks if an individual has children.
func (fi *FilterIndexes) hasChildren(xrefID string) bool {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	return fi.hasChildrenIndex[xrefID]
}

// hasSpouse checks if an individual has a spouse.
func (fi *FilterIndexes) hasSpouse(xrefID string) bool {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	return fi.hasSpouseIndex[xrefID]
}

// isLiving checks if an individual is living.
func (fi *FilterIndexes) isLiving(xrefID string) bool {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	return fi.livingIndex[xrefID]
}
