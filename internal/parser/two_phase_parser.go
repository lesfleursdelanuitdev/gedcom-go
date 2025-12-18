package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// RawRecord represents a record with unparsed child lines.
type RawRecord struct {
	Level      int
	Tag        string
	Value      string
	XrefID     string
	LineNumber int
	RawLines   []string // Unparsed child lines (level > 0)
}

// TwoPhaseParser implements a two-phase parsing approach:
// Phase 1: Collect records sequentially (identify boundaries)
// Phase 2: Parse record children in parallel
type TwoPhaseParser struct {
	tree         *gedcom.GedcomTree
	errorManager *gedcom.ErrorManager
	records      []*RawRecord
}

// NewTwoPhaseParser creates a new TwoPhaseParser.
func NewTwoPhaseParser() *TwoPhaseParser {
	return &TwoPhaseParser{
		tree:         gedcom.NewGedcomTree(),
		errorManager: gedcom.NewErrorManager(),
		records:      make([]*RawRecord, 0),
	}
}

// Parse parses a GEDCOM file using two-phase approach.
func (tpp *TwoPhaseParser) Parse(filePath string) (*gedcom.GedcomTree, error) {
	// Step 1: Validate file
	if err := ValidateFile(filePath); err != nil {
		tpp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("File validation failed: %v", err), 0, "File Validation")
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Step 2: Detect encoding
	encoding, err := DetectEncoding(filePath)
	if err != nil {
		tpp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Encoding detection failed: %v", err), 0, "Encoding Detection")
		return nil, fmt.Errorf("encoding detection failed: %w", err)
	}
	tpp.tree.SetEncoding(string(encoding))

	// Step 3: Open file
	file, err := os.Open(filePath)
	if err != nil {
		tpp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Failed to open file: %v", err), 0, "File I/O")
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Step 4: Get reader with proper encoding
	reader, err := GetReader(file, encoding)
	if err != nil {
		tpp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Failed to create reader: %v", err), 0, "Encoding")
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	// PHASE 1: Collect records sequentially
	scanner := bufio.NewScanner(reader)
	if err := tpp.collectRecords(scanner); err != nil {
		return nil, err
	}

	// PHASE 2: Parse records in parallel
	tpp.parseRecordsParallel()

	return tpp.tree, nil
}

// collectRecords (Phase 1) collects all level 0 records with their raw child lines.
// This phase must be sequential to identify record boundaries.
func (tpp *TwoPhaseParser) collectRecords(scanner *bufio.Scanner) error {
	lineNumber := 0
	var currentRecord *RawRecord

	for scanner.Scan() {
		lineNumber++
		rawLine := scanner.Text()

		// Skip empty lines
		rawLine = strings.TrimSpace(rawLine)
		if len(rawLine) == 0 {
			continue
		}

		// Parse the line
		level, tag, value, xrefID, err := ParseLine(rawLine)
		if err != nil {
			tpp.errorManager.AddError(gedcom.SeverityWarning, fmt.Sprintf("Malformed line: %v", err), lineNumber, "Line Parsing")
			continue
		}

		// Handle CONC/CONT continuation lines
		if tag == "CONC" || tag == "CONT" {
			// Apply continuation to previous line
			if currentRecord != nil && len(currentRecord.RawLines) > 0 {
				// Append to last raw line's value
				lastRawLine := currentRecord.RawLines[len(currentRecord.RawLines)-1]
				if tag == "CONC" {
					currentRecord.RawLines[len(currentRecord.RawLines)-1] = lastRawLine + value
				} else {
					currentRecord.RawLines[len(currentRecord.RawLines)-1] = lastRawLine + "\n" + value
				}
			} else if currentRecord != nil {
				// Append to record value
				if tag == "CONC" {
					currentRecord.Value += value
				} else {
					currentRecord.Value += "\n" + value
				}
			}
			continue
		}

		// Handle level 0 (new record)
		if level == 0 {
			// Save previous record if exists
			if currentRecord != nil {
				tpp.records = append(tpp.records, currentRecord)
			}

			// Start new record
			currentRecord = &RawRecord{
				Level:      level,
				Tag:        tag,
				Value:      value,
				XrefID:     xrefID,
				LineNumber: lineNumber,
				RawLines:   make([]string, 0),
			}
			continue
		}

		// Handle level > 0 (child line of current record)
		if currentRecord == nil {
			// Orphaned line - no parent record
			tpp.errorManager.AddError(gedcom.SeverityWarning,
				fmt.Sprintf("Orphaned line at level %d with no parent record: %s", level, tag),
				lineNumber,
				"Hierarchy")
			continue
		}

		// Add raw line to current record
		currentRecord.RawLines = append(currentRecord.RawLines, rawLine)
	}

	// Save last record
	if currentRecord != nil {
		tpp.records = append(tpp.records, currentRecord)
	}

	if err := scanner.Err(); err != nil {
		tpp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Error reading file: %v", err), lineNumber, "File I/O")
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// parseRecordsParallel (Phase 2) parses each record's children in parallel.
func (tpp *TwoPhaseParser) parseRecordsParallel() {
	const numWorkers = 4 // Adjust based on CPU cores
	workChan := make(chan *RawRecord, len(tpp.records))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rawRecord := range workChan {
				tpp.parseRecord(rawRecord)
			}
		}()
	}

	// Send records to workers
	for _, rawRecord := range tpp.records {
		workChan <- rawRecord
	}
	close(workChan)

	// Wait for all workers to complete
	wg.Wait()
}

// parseRecord parses a single record and its children.
func (tpp *TwoPhaseParser) parseRecord(rawRecord *RawRecord) {
	// Create the main GedcomLine for this record
	mainLine := gedcom.NewGedcomLine(rawRecord.Level, rawRecord.Tag, rawRecord.Value, rawRecord.XrefID)
	mainLine.LineNumber = rawRecord.LineNumber

	// Parse child lines using stack-based approach
	stack := NewLineStack()
	stack.Push(mainLine)

	for _, rawLine := range rawRecord.RawLines {
		level, tag, value, _, err := ParseLine(rawLine)
		if err != nil {
			// Skip malformed lines (already logged in phase 1)
			continue
		}

		// Find parent using stack
		parent, err := stack.FindParent(level)
		if err != nil {
			// Orphaned line within record - skip
			continue
		}

		// Create child line
		childLine := gedcom.NewGedcomLine(level, tag, value, "")
		// Extract line number from raw line if possible (approximate)
		childLine.LineNumber = rawRecord.LineNumber // Approximate

		// Add as child to parent
		parent.AddChild(childLine)

		// Push to stack
		stack.Push(childLine)
	}

	// Create record from main line
	factory := gedcom.NewRecordFactory()
	record := factory.CreateRecord(mainLine)

	// Add to tree (thread-safe)
	tpp.tree.AddRecord(record)
}

// GetErrors returns all errors collected during parsing
func (tpp *TwoPhaseParser) GetErrors() []*gedcom.GedcomError {
	return tpp.errorManager.Errors()
}

// HasErrors returns true if any errors were encountered
func (tpp *TwoPhaseParser) HasErrors() bool {
	return tpp.errorManager.HasErrors()
}

// GetErrorManager returns the error manager
func (tpp *TwoPhaseParser) GetErrorManager() *gedcom.ErrorManager {
	return tpp.errorManager
}

// GetTree returns the parsed tree
func (tpp *TwoPhaseParser) GetTree() *gedcom.GedcomTree {
	return tpp.tree
}

