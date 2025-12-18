package diff

import (
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// DiffConfig holds configuration for GEDCOM comparison.
type DiffConfig struct {
	// Matching strategy
	MatchingStrategy string // "xref", "content", "hybrid" (default: "xref")

	// Thresholds
	SimilarityThreshold float64 // For content matching (default: 0.85)
	DateTolerance       int     // Years tolerance for date comparison (default: 2)

	// Options
	IncludeUnchanged bool   // Include unchanged records in output (default: false)
	DetailLevel      string // "summary", "field", "full" (default: "field")
	OutputFormat     string // "text", "json", "html", "unified" (default: "text")

	// Change history
	TrackHistory bool // Track change history (default: true)
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *DiffConfig {
	return &DiffConfig{
		MatchingStrategy:    "xref",
		SimilarityThreshold: 0.85,
		DateTolerance:       2,
		IncludeUnchanged:    false,
		DetailLevel:         "field",
		OutputFormat:        "text",
		TrackHistory:        true,
	}
}

// ChangeType represents the type of change.
type ChangeType string

const (
	ChangeTypeAdded                  ChangeType = "added"
	ChangeTypeRemoved                ChangeType = "removed"
	ChangeTypeModified               ChangeType = "modified"
	ChangeTypeSemanticallyEquivalent ChangeType = "semantically_equivalent"
	ChangeTypeUnchanged              ChangeType = "unchanged"
)

// ChangeHistory tracks who, when, and what changed.
type ChangeHistory struct {
	Timestamp  time.Time
	Author     string // Optional: who made the change
	Reason     string // Optional: why the change was made
	ChangeType ChangeType
	Field      string // Field that changed (e.g., "NAME", "BIRT.DATE")
	OldValue   string
	NewValue   string
}

// FieldChange represents a change to a specific field.
type FieldChange struct {
	Field    string
	Path     string // Full path (e.g., "BIRT.DATE")
	OldValue interface{}
	NewValue interface{}
	Type     ChangeType
	History  []ChangeHistory // Change history for this field
}

// RecordModification represents changes to a record.
type RecordModification struct {
	Xref    string
	Type    string // "INDI", "FAM", "NOTE", etc.
	Changes []FieldChange
	History []ChangeHistory // Overall record change history
}

// RecordDiff represents a record that was added or removed.
type RecordDiff struct {
	Xref    string
	Type    string
	Record  gedcom.Record
	History []ChangeHistory
}

// DiffSummary provides summary statistics.
type DiffSummary struct {
	File1Stats RecordStats
	File2Stats RecordStats
	Changes    ChangeCounts
}

// RecordStats provides statistics for a file.
type RecordStats struct {
	Individuals  int
	Families     int
	Notes        int
	Sources      int
	Repositories int
	Submitters   int
	Multimedia   int
}

// ChangeCounts provides counts of changes.
type ChangeCounts struct {
	Added     RecordTypeCounts
	Removed   RecordTypeCounts
	Modified  RecordTypeCounts
	Unchanged RecordTypeCounts
}

// RecordTypeCounts provides counts by record type.
type RecordTypeCounts struct {
	Individuals  int
	Families     int
	Notes        int
	Sources      int
	Repositories int
	Submitters   int
	Multimedia   int
}

// DiffChanges contains all detected changes.
type DiffChanges struct {
	Added    []RecordDiff
	Removed  []RecordDiff
	Modified []RecordModification
}

// DiffResult holds the complete diff result.
type DiffResult struct {
	Summary    DiffSummary
	Changes    DiffChanges
	Statistics DiffStatistics
	History    []ChangeHistory // Global change history
}

// DiffStatistics provides performance and processing statistics.
type DiffStatistics struct {
	ProcessingTime  time.Duration
	ComparisonTime  time.Duration
	TotalRecords1   int
	TotalRecords2   int
	RecordsCompared int
	FieldsCompared  int
}

// GedcomDiffer performs semantic comparison of GEDCOM files.
type GedcomDiffer struct {
	config *DiffConfig
}

// NewGedcomDiffer creates a new GEDCOM differ with the given configuration.
func NewGedcomDiffer(config *DiffConfig) *GedcomDiffer {
	if config == nil {
		config = DefaultConfig()
	}
	return &GedcomDiffer{
		config: config,
	}
}

// Compare compares two GEDCOM trees and returns the differences.
func (gd *GedcomDiffer) Compare(tree1, tree2 *gedcom.GedcomTree) (*DiffResult, error) {
	startTime := time.Now()

	// Build statistics
	stats1 := gd.buildStatistics(tree1)
	stats2 := gd.buildStatistics(tree2)

	// Perform comparison based on strategy
	var changes DiffChanges
	var err error

	switch gd.config.MatchingStrategy {
	case "xref":
		changes, err = gd.compareByXref(tree1, tree2)
	case "content":
		changes, err = gd.compareByContent(tree1, tree2)
	case "hybrid":
		changes, err = gd.compareHybrid(tree1, tree2)
	default:
		changes, err = gd.compareByXref(tree1, tree2)
	}

	if err != nil {
		return nil, err
	}

	// Build summary
	summary := gd.buildSummary(stats1, stats2, changes)

	// Build statistics
	statistics := DiffStatistics{
		ProcessingTime:  time.Since(startTime),
		TotalRecords1:   gd.countRecords(tree1),
		TotalRecords2:   gd.countRecords(tree2),
		RecordsCompared: len(changes.Added) + len(changes.Removed) + len(changes.Modified),
	}

	// Collect global change history
	history := gd.collectGlobalHistory(changes)

	return &DiffResult{
		Summary:    summary,
		Changes:    changes,
		Statistics: statistics,
		History:    history,
	}, nil
}

// CompareFiles compares two GEDCOM files from disk.
func (gd *GedcomDiffer) CompareFiles(file1, file2 string) (*DiffResult, error) {
	// Parse both files
	// This would use the parser package
	// For now, return error indicating files need to be parsed first
	return nil, nil // TODO: Implement file parsing
}

// buildStatistics builds statistics for a GEDCOM tree.
func (gd *GedcomDiffer) buildStatistics(tree *gedcom.GedcomTree) RecordStats {
	allIndi := tree.GetAllIndividuals()
	allFam := tree.GetAllFamilies()
	allNotes := tree.GetAllNotes()
	allSources := tree.GetAllSources()
	allRepos := tree.GetAllRepositories()
	allSubm := tree.GetAllSubmitters()
	allMult := tree.GetAllMultimedia()

	return RecordStats{
		Individuals:  len(allIndi),
		Families:     len(allFam),
		Notes:        len(allNotes),
		Sources:      len(allSources),
		Repositories: len(allRepos),
		Submitters:   len(allSubm),
		Multimedia:   len(allMult),
	}
}

// countRecords counts total records in a tree.
func (gd *GedcomDiffer) countRecords(tree *gedcom.GedcomTree) int {
	stats := gd.buildStatistics(tree)
	return stats.Individuals + stats.Families + stats.Notes + stats.Sources +
		stats.Repositories + stats.Submitters + stats.Multimedia
}

// buildSummary builds a summary from statistics and changes.
func (gd *GedcomDiffer) buildSummary(stats1, stats2 RecordStats, changes DiffChanges) DiffSummary {
	// Count changes by type
	addedCounts := gd.countChangesByType(changes.Added)
	removedCounts := gd.countChangesByType(changes.Removed)
	modifiedCounts := gd.countChangesByTypeFromModifications(changes.Modified)

	// Calculate unchanged (records that exist in both and weren't modified)
	unchangedCounts := RecordTypeCounts{
		Individuals: stats1.Individuals - removedCounts.Individuals - modifiedCounts.Individuals,
		Families:    stats1.Families - removedCounts.Families - modifiedCounts.Families,
		Notes:       stats1.Notes - removedCounts.Notes - modifiedCounts.Notes,
		Sources:     stats1.Sources - removedCounts.Sources - modifiedCounts.Sources,
	}

	return DiffSummary{
		File1Stats: stats1,
		File2Stats: stats2,
		Changes: ChangeCounts{
			Added:     addedCounts,
			Removed:   removedCounts,
			Modified:  modifiedCounts,
			Unchanged: unchangedCounts,
		},
	}
}

// countChangesByType counts changes by record type.
func (gd *GedcomDiffer) countChangesByType(diffs []RecordDiff) RecordTypeCounts {
	counts := RecordTypeCounts{}
	for _, diff := range diffs {
		switch diff.Type {
		case "INDI":
			counts.Individuals++
		case "FAM":
			counts.Families++
		case "NOTE":
			counts.Notes++
		case "SOUR":
			counts.Sources++
		case "REPO":
			counts.Repositories++
		case "SUBM":
			counts.Submitters++
		case "OBJE":
			counts.Multimedia++
		}
	}
	return counts
}

// countChangesByTypeFromModifications counts modifications by record type.
func (gd *GedcomDiffer) countChangesByTypeFromModifications(mods []RecordModification) RecordTypeCounts {
	counts := RecordTypeCounts{}
	for _, mod := range mods {
		switch mod.Type {
		case "INDI":
			counts.Individuals++
		case "FAM":
			counts.Families++
		case "NOTE":
			counts.Notes++
		case "SOUR":
			counts.Sources++
		case "REPO":
			counts.Repositories++
		case "SUBM":
			counts.Submitters++
		case "OBJE":
			counts.Multimedia++
		}
	}
	return counts
}

// collectGlobalHistory collects all change history entries.
func (gd *GedcomDiffer) collectGlobalHistory(changes DiffChanges) []ChangeHistory {
	history := make([]ChangeHistory, 0)

	// Collect from added records
	for _, added := range changes.Added {
		history = append(history, added.History...)
	}

	// Collect from removed records
	for _, removed := range changes.Removed {
		history = append(history, removed.History...)
	}

	// Collect from modified records
	for _, modified := range changes.Modified {
		history = append(history, modified.History...)
		for _, fieldChange := range modified.Changes {
			history = append(history, fieldChange.History...)
		}
	}

	return history
}

// createChangeHistory creates a change history entry.
func (gd *GedcomDiffer) createChangeHistory(changeType ChangeType, field, oldValue, newValue string) ChangeHistory {
	return ChangeHistory{
		Timestamp:  time.Now(),
		ChangeType: changeType,
		Field:      field,
		OldValue:   oldValue,
		NewValue:   newValue,
	}
}
