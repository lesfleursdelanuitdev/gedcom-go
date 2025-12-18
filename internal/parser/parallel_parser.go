package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// ParallelHierarchicalParser is an experimental parser that attempts to parallelize
// some aspects of parsing. Note: The core parsing must remain sequential due to
// hierarchical structure, but we can parallelize record creation and processing.
type ParallelHierarchicalParser struct {
	tree                *gedcom.GedcomTree
	parentsStack        *LineStack
	continuationHandler *ContinuationHandler
	errorManager        *gedcom.ErrorManager
	
	// Channel for parallel record processing
	recordChan chan *gedcom.GedcomLine
	wg         sync.WaitGroup
}

// NewParallelHierarchicalParser creates a new ParallelHierarchicalParser.
// Note: This is experimental. The sequential parser is recommended for most use cases.
func NewParallelHierarchicalParser() *ParallelHierarchicalParser {
	php := &ParallelHierarchicalParser{
		tree:                gedcom.NewGedcomTree(),
		parentsStack:        NewLineStack(),
		continuationHandler: NewContinuationHandler(),
		errorManager:        gedcom.NewErrorManager(),
		recordChan:          make(chan *gedcom.GedcomLine, 100), // Buffered channel
	}
	
	// Start record processor goroutine
	php.wg.Add(1)
	go php.processRecords()
	
	return php
}

// processRecords processes level 0 records in a separate goroutine.
// This allows the main parsing loop to continue while records are being created.
func (php *ParallelHierarchicalParser) processRecords() {
	defer php.wg.Done()
	
	factory := gedcom.NewRecordFactory()
	
	for line := range php.recordChan {
		record := factory.CreateRecord(line)
		php.tree.AddRecord(record)
	}
}

// Parse parses a GEDCOM file. The parsing itself is sequential, but record
// creation happens in a parallel goroutine.
func (php *ParallelHierarchicalParser) Parse(filePath string) (*gedcom.GedcomTree, error) {
	// Step 1: Validate file
	if err := ValidateFile(filePath); err != nil {
		php.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("File validation failed: %v", err), 0, "File Validation")
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Step 2: Detect encoding
	encoding, err := DetectEncoding(filePath)
	if err != nil {
		php.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Encoding detection failed: %v", err), 0, "Encoding Detection")
		return nil, fmt.Errorf("encoding detection failed: %w", err)
	}
	php.tree.SetEncoding(string(encoding))

	// Step 3: Open file
	file, err := os.Open(filePath)
	if err != nil {
		php.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Failed to open file: %v", err), 0, "File I/O")
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Step 4: Get reader with proper encoding
	reader, err := GetReader(file, encoding)
	if err != nil {
		php.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Failed to create reader: %v", err), 0, "Encoding")
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	// Step 5: Parse file line by line (sequential - required for hierarchy)
	scanner := bufio.NewScanner(reader)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Skip empty lines
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parse the line
		level, tag, value, xrefID, err := ParseLine(line)
		if err != nil {
			php.errorManager.AddError(gedcom.SeverityWarning, fmt.Sprintf("Malformed line: %v", err), lineNumber, "Line Parsing")
			continue
		}

		// Handle CONC/CONT continuation lines
		if tag == "CONC" || tag == "CONT" {
			if err := php.continuationHandler.HandleContinuation(tag, level, value); err != nil {
				php.errorManager.AddError(gedcom.SeverityWarning, fmt.Sprintf("Invalid continuation: %v", err), lineNumber, "CONC/CONT Handling")
				continue
			}
			continue
		}

		// Apply accumulated CONC/CONT value
		if php.continuationHandler.HasAccumulatedValue() {
			if !php.parentsStack.IsEmpty() {
				topLine := php.parentsStack.Peek()
				accumulatedValue := php.continuationHandler.GetAccumulatedValue()
				if topLine.Value != "" {
					topLine.Value += accumulatedValue
				} else {
					topLine.Value = accumulatedValue
				}
			}
		}

		// Handle level 0 (top-level record)
		if level == 0 {
			// Create GedcomLine
			gedcomLine := gedcom.NewGedcomLine(level, tag, value, xrefID)
			gedcomLine.LineNumber = lineNumber

			// Send to parallel processor (non-blocking if channel has space)
			select {
			case php.recordChan <- gedcomLine:
				// Record sent for parallel processing
			default:
				// Channel full, process synchronously (fallback)
				factory := gedcom.NewRecordFactory()
				record := factory.CreateRecord(gedcomLine)
				php.tree.AddRecord(record)
			}

			// Reset stack (new top-level record)
			php.parentsStack.Clear()
			php.parentsStack.Push(gedcomLine)

			// Update last tag
			php.continuationHandler.SetLastTag(tag, level)

			continue
		}

		// Handle level > 0 (child lines) - must be sequential
		parent, err := php.parentsStack.FindParent(level)
		if err != nil {
			php.errorManager.AddError(gedcom.SeverityWarning, fmt.Sprintf("Orphaned line: %s", err.Error()), lineNumber, "Hierarchy")
			continue
		}

		// Create child line
		childLine := gedcom.NewGedcomLine(level, tag, value, "")
		childLine.LineNumber = lineNumber

		// Add as child to parent
		parent.AddChild(childLine)

		// Push to stack
		php.parentsStack.Push(childLine)

		// Update last tag
		php.continuationHandler.SetLastTag(tag, level)
	}

	// Handle remaining CONC/CONT value
	if php.continuationHandler.HasAccumulatedValue() {
		if !php.parentsStack.IsEmpty() {
			topLine := php.parentsStack.Peek()
			accumulatedValue := php.continuationHandler.GetAccumulatedValue()
			if topLine.Value != "" {
				topLine.Value += accumulatedValue
			} else {
				topLine.Value = accumulatedValue
			}
		}
	}

	if err := scanner.Err(); err != nil {
		php.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Error reading file: %v", err), lineNumber, "File I/O")
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Close channel and wait for record processor to finish
	close(php.recordChan)
	php.wg.Wait()

	// Return tree
	return php.tree, nil
}

// GetErrors returns all errors collected during parsing
func (php *ParallelHierarchicalParser) GetErrors() []*gedcom.GedcomError {
	return php.errorManager.Errors()
}

// HasErrors returns true if any errors were encountered
func (php *ParallelHierarchicalParser) HasErrors() bool {
	return php.errorManager.HasErrors()
}

// GetErrorManager returns the error manager
func (php *ParallelHierarchicalParser) GetErrorManager() *gedcom.ErrorManager {
	return php.errorManager
}

// GetTree returns the parsed tree
func (php *ParallelHierarchicalParser) GetTree() *gedcom.GedcomTree {
	return php.tree
}


