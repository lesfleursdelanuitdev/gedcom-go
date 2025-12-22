package query

import (
	"fmt"
	"strings"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// NoteInfo represents information about a note (either referenced or inline).
type NoteInfo struct {
	XrefID     string // Empty for inline notes
	Text       string // Note text content
	IsInline   bool   // True if inline, false if referenced
	NoteRecord *gedcom.NoteRecord // nil for inline notes
}

// GetAllNotesForIndividual returns all notes associated with an individual,
// including both referenced notes (via NOTE xrefs) and inline notes (embedded in the record).
func (iq *IndividualQuery) GetAllNotes() ([]NoteInfo, error) {
	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	notes := make([]NoteInfo, 0)

	// 1. Get referenced notes (via NOTE edges)
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeNOTE {
			if noteNode, ok := edge.To.(*NoteNode); ok && noteNode.Note != nil {
				// Get full note text (including CONT/CONC lines)
				noteText := getFullNoteText(noteNode.Note)
				notes = append(notes, NoteInfo{
					XrefID:     noteNode.ID(),
					Text:       noteText,
					IsInline:   false,
					NoteRecord: noteNode.Note,
				})
			}
		}
	}

	// 2. Get inline notes (embedded in the record structure)
	if node.Individual != nil {
		inlineNotes := getInlineNotes(node.Individual)
		notes = append(notes, inlineNotes...)
	}

	return notes, nil
}

// GetAllNotesForFamily returns all notes associated with a family,
// including both referenced notes (via NOTE xrefs) and inline notes (embedded in the record).
func (fq *FamilyQuery) GetAllNotes() ([]NoteInfo, error) {
	node := fq.graph.GetFamily(fq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	notes := make([]NoteInfo, 0)

	// 1. Get referenced notes (via NOTE edges)
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeNOTE {
			if noteNode, ok := edge.To.(*NoteNode); ok && noteNode.Note != nil {
				// Get full note text (including CONT/CONC lines)
				noteText := getFullNoteText(noteNode.Note)
				notes = append(notes, NoteInfo{
					XrefID:     noteNode.ID(),
					Text:       noteText,
					IsInline:   false,
					NoteRecord: noteNode.Note,
				})
			}
		}
	}

	// 2. Get inline notes (embedded in the record structure)
	if node.Family != nil {
		inlineNotes := getInlineNotesForFamily(node.Family)
		notes = append(notes, inlineNotes...)
	}

	return notes, nil
}

// GetRecordsForNote returns all records (individuals, families, events) that reference this note.
func (g *Graph) GetRecordsForNote(noteXrefID string) ([]GraphNode, error) {
	noteNode := g.GetNote(noteXrefID)
	if noteNode == nil {
		return nil, fmt.Errorf("note %s not found", noteXrefID)
	}

	records := make([]GraphNode, 0)
	seen := make(map[string]bool)

	// Traverse in-edges to find all records that reference this note
	for _, edge := range noteNode.InEdges() {
		if edge.From != nil {
			fromID := edge.From.ID()
			if !seen[fromID] {
				seen[fromID] = true
				records = append(records, edge.From)
			}
		}
	}

	return records, nil
}

// getFullNoteText extracts the complete note text including CONT/CONC continuation lines.
func getFullNoteText(noteRecord *gedcom.NoteRecord) string {
	if noteRecord == nil {
		return ""
	}

	// Get the base note value
	text := noteRecord.GetText()
	
	// Get all CONT (continue) and CONC (concatenate) lines
	contLines := noteRecord.GetLines("CONT")
	concLines := noteRecord.GetLines("CONC")

	// Build full text
	var parts []string
	if text != "" {
		parts = append(parts, text)
	}

	// Process CONC lines (concatenate without newline)
	for _, concLine := range concLines {
		if concLine.Value != "" {
			if len(parts) > 0 {
				parts[len(parts)-1] += concLine.Value
			} else {
				parts = append(parts, concLine.Value)
			}
		}
	}

	// Process CONT lines (continue with newline)
	for _, contLine := range contLines {
		if contLine.Value != "" {
			parts = append(parts, contLine.Value)
		}
	}

	return strings.Join(parts, "\n")
}

// getInlineNotes extracts inline notes from an individual record.
// Inline notes are NOTE lines that have text directly (not an xref).
func getInlineNotes(indi *gedcom.IndividualRecord) []NoteInfo {
	notes := make([]NoteInfo, 0)
	
	// Get all NOTE lines
	noteLines := indi.GetLines("NOTE")
	for _, noteLine := range noteLines {
		// If the NOTE line has a value (not an xref), it's an inline note
		if noteLine.Value != "" && !strings.HasPrefix(noteLine.Value, "@") {
			// Build full text including CONT/CONC
			text := noteLine.Value
			for _, contLine := range noteLine.GetLines("CONT") {
				if contLine.Value != "" {
					text += "\n" + contLine.Value
				}
			}
			for _, concLine := range noteLine.GetLines("CONC") {
				if concLine.Value != "" {
					text += concLine.Value
				}
			}
			
			notes = append(notes, NoteInfo{
				XrefID:     "",
				Text:       text,
				IsInline:   true,
				NoteRecord: nil,
			})
		}
	}

	return notes
}

// getInlineNotesForFamily extracts inline notes from a family record.
func getInlineNotesForFamily(fam *gedcom.FamilyRecord) []NoteInfo {
	notes := make([]NoteInfo, 0)
	
	// Get all NOTE lines
	noteLines := fam.GetLines("NOTE")
	for _, noteLine := range noteLines {
		// If the NOTE line has a value (not an xref), it's an inline note
		if noteLine.Value != "" && !strings.HasPrefix(noteLine.Value, "@") {
			// Build full text including CONT/CONC
			text := noteLine.Value
			for _, contLine := range noteLine.GetLines("CONT") {
				if contLine.Value != "" {
					text += "\n" + contLine.Value
				}
			}
			for _, concLine := range noteLine.GetLines("CONC") {
				if concLine.Value != "" {
					text += concLine.Value
				}
			}
			
			notes = append(notes, NoteInfo{
				XrefID:     "",
				Text:       text,
				IsInline:   true,
				NoteRecord: nil,
			})
		}
	}

	return notes
}

