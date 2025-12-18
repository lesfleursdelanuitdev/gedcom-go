package diff

import (
	"fmt"
	"strings"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// GenerateReport generates a text report from diff results.
func (gd *GedcomDiffer) GenerateReport(result *DiffResult) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("GEDCOM Diff Report\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	// Summary
	gd.writeSummary(&sb, result.Summary)

	// Added records
	if len(result.Changes.Added) > 0 {
		gd.writeAddedRecords(&sb, result.Changes.Added)
	}

	// Removed records
	if len(result.Changes.Removed) > 0 {
		gd.writeRemovedRecords(&sb, result.Changes.Removed)
	}

	// Modified records
	if len(result.Changes.Modified) > 0 {
		gd.writeModifiedRecords(&sb, result.Changes.Modified)
	}

	// Statistics
	gd.writeStatistics(&sb, result.Statistics)

	// Change history (if enabled)
	if gd.config.TrackHistory && len(result.History) > 0 {
		gd.writeChangeHistory(&sb, result.History)
	}

	return sb.String(), nil
}

// writeSummary writes the summary section.
func (gd *GedcomDiffer) writeSummary(sb *strings.Builder, summary DiffSummary) {
	sb.WriteString("Summary:\n")
	sb.WriteString(fmt.Sprintf("  File 1: %d individuals, %d families, %d notes, %d sources\n",
		summary.File1Stats.Individuals,
		summary.File1Stats.Families,
		summary.File1Stats.Notes,
		summary.File1Stats.Sources))
	sb.WriteString(fmt.Sprintf("  File 2: %d individuals, %d families, %d notes, %d sources\n\n",
		summary.File2Stats.Individuals,
		summary.File2Stats.Families,
		summary.File2Stats.Notes,
		summary.File2Stats.Sources))

	sb.WriteString("  Changes:\n")
	sb.WriteString(fmt.Sprintf("    Added:     %d individuals, %d families\n",
		summary.Changes.Added.Individuals,
		summary.Changes.Added.Families))
	sb.WriteString(fmt.Sprintf("    Removed:   %d individuals, %d families\n",
		summary.Changes.Removed.Individuals,
		summary.Changes.Removed.Families))
	sb.WriteString(fmt.Sprintf("    Modified:  %d individuals, %d families\n",
		summary.Changes.Modified.Individuals,
		summary.Changes.Modified.Families))
	sb.WriteString(fmt.Sprintf("    Unchanged: %d individuals, %d families\n\n",
		summary.Changes.Unchanged.Individuals,
		summary.Changes.Unchanged.Families))
}

// writeAddedRecords writes the added records section.
func (gd *GedcomDiffer) writeAddedRecords(sb *strings.Builder, added []RecordDiff) {
	sb.WriteString("Added Records:\n")
	sb.WriteString(strings.Repeat("-", 50) + "\n")
	for _, diff := range added {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", diff.Xref, diff.Type))
		if indi, ok := diff.Record.(*gedcom.IndividualRecord); ok {
			name := indi.GetName()
			birthDate := indi.GetBirthDate()
			birthPlace := indi.GetBirthPlace()
			sb.WriteString(fmt.Sprintf("    Name: %s\n", name))
			if birthDate != "" {
				sb.WriteString(fmt.Sprintf("    Birth: %s", birthDate))
				if birthPlace != "" {
					sb.WriteString(fmt.Sprintf(", %s", birthPlace))
				}
				sb.WriteString("\n")
			}
		}
		if gd.config.TrackHistory && len(diff.History) > 0 {
			gd.writeRecordHistory(sb, diff.History)
		}
		sb.WriteString("\n")
	}
}

// writeRemovedRecords writes the removed records section.
func (gd *GedcomDiffer) writeRemovedRecords(sb *strings.Builder, removed []RecordDiff) {
	sb.WriteString("Removed Records:\n")
	sb.WriteString(strings.Repeat("-", 50) + "\n")
	for _, diff := range removed {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", diff.Xref, diff.Type))
		if indi, ok := diff.Record.(*gedcom.IndividualRecord); ok {
			name := indi.GetName()
			birthDate := indi.GetBirthDate()
			birthPlace := indi.GetBirthPlace()
			sb.WriteString(fmt.Sprintf("    Name: %s\n", name))
			if birthDate != "" {
				sb.WriteString(fmt.Sprintf("    Birth: %s", birthDate))
				if birthPlace != "" {
					sb.WriteString(fmt.Sprintf(", %s", birthPlace))
				}
				sb.WriteString("\n")
			}
		}
		if gd.config.TrackHistory && len(diff.History) > 0 {
			gd.writeRecordHistory(sb, diff.History)
		}
		sb.WriteString("\n")
	}
}

// writeModifiedRecords writes the modified records section.
func (gd *GedcomDiffer) writeModifiedRecords(sb *strings.Builder, modified []RecordModification) {
	sb.WriteString("Modified Records:\n")
	sb.WriteString(strings.Repeat("-", 50) + "\n")
	for _, mod := range modified {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", mod.Xref, mod.Type))
		for _, change := range mod.Changes {
			switch change.Type {
			case ChangeTypeModified:
				sb.WriteString(fmt.Sprintf("    %s: %v → %v\n",
					change.Path, change.OldValue, change.NewValue))
			case ChangeTypeAdded:
				sb.WriteString(fmt.Sprintf("    %s: Added (%v)\n",
					change.Path, change.NewValue))
			case ChangeTypeRemoved:
				sb.WriteString(fmt.Sprintf("    %s: Removed (%v)\n",
					change.Path, change.OldValue))
			case ChangeTypeSemanticallyEquivalent:
				sb.WriteString(fmt.Sprintf("    %s: %v → %v (semantically equivalent)\n",
					change.Path, change.OldValue, change.NewValue))
			}
		}
		if gd.config.TrackHistory && len(mod.History) > 0 {
			gd.writeRecordHistory(sb, mod.History)
		}
		sb.WriteString("\n")
	}
}

// writeStatistics writes the statistics section.
func (gd *GedcomDiffer) writeStatistics(sb *strings.Builder, stats DiffStatistics) {
	sb.WriteString("Statistics:\n")
	sb.WriteString(strings.Repeat("-", 50) + "\n")
	sb.WriteString(fmt.Sprintf("  Processing Time: %v\n", stats.ProcessingTime))
	sb.WriteString(fmt.Sprintf("  Total Records (File 1): %d\n", stats.TotalRecords1))
	sb.WriteString(fmt.Sprintf("  Total Records (File 2): %d\n", stats.TotalRecords2))
	sb.WriteString(fmt.Sprintf("  Records Compared: %d\n", stats.RecordsCompared))
	sb.WriteString("\n")
}

// writeChangeHistory writes the change history section.
func (gd *GedcomDiffer) writeChangeHistory(sb *strings.Builder, history []ChangeHistory) {
	sb.WriteString("Change History:\n")
	sb.WriteString(strings.Repeat("-", 50) + "\n")
	for _, entry := range history {
		sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.ChangeType,
			entry.Field))
		if entry.OldValue != "" || entry.NewValue != "" {
			sb.WriteString(fmt.Sprintf("    %s → %s\n", entry.OldValue, entry.NewValue))
		}
		if entry.Author != "" {
			sb.WriteString(fmt.Sprintf("    Author: %s\n", entry.Author))
		}
		if entry.Reason != "" {
			sb.WriteString(fmt.Sprintf("    Reason: %s\n", entry.Reason))
		}
		sb.WriteString("\n")
	}
}

// writeRecordHistory writes history for a specific record.
func (gd *GedcomDiffer) writeRecordHistory(sb *strings.Builder, history []ChangeHistory) {
	sb.WriteString("    History:\n")
	for _, entry := range history {
		sb.WriteString(fmt.Sprintf("      [%s] %s: %s\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.ChangeType,
			entry.Field))
	}
}
