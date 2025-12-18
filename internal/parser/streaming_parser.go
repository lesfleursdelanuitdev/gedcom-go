package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// RecordHandler is a callback function that processes records as they're parsed.
// It receives the record and can return an error to stop parsing.
type RecordHandler func(record gedcom.Record) error

// StreamingHierarchicalParser is a parser that processes records incrementally
// without loading the entire tree into memory. This is useful for very large files (>100MB).
//
// The parser processes records one at a time using a callback function, allowing
// the caller to handle records immediately without storing them all in memory.
type StreamingHierarchicalParser struct {
	continuationHandler *ContinuationHandler
	errorManager        *gedcom.ErrorManager
}

// NewStreamingHierarchicalParser creates a new StreamingHierarchicalParser.
func NewStreamingHierarchicalParser() *StreamingHierarchicalParser {
	return &StreamingHierarchicalParser{
		continuationHandler: NewContinuationHandler(),
		errorManager:        gedcom.NewErrorManager(),
	}
}

// ParseWithHandler parses a GEDCOM file and calls the handler for each level-0 record
// as it's encountered. This allows processing very large files without loading
// the entire tree into memory.
//
// The handler is called for each complete record (INDI, FAM, NOTE, etc.) as soon
// as it's fully parsed. If the handler returns an error, parsing stops.
//
// Example:
//
//	parser := NewStreamingHierarchicalParser()
//	err := parser.ParseWithHandler("large.ged", func(record gedcom.Record) error {
//		// Process record immediately
//		fmt.Printf("Found record: %s\n", record.Type())
//		return nil // Continue parsing
//	})
func (shp *StreamingHierarchicalParser) ParseWithHandler(filePath string, handler RecordHandler) error {
	// Step 1: Validate file
	if err := ValidateFile(filePath); err != nil {
		shp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("File validation failed: %v", err), 0, "File Validation")
		return fmt.Errorf("file validation failed: %w", err)
	}

	// Step 2: Detect encoding
	encoding, err := DetectEncoding(filePath)
	if err != nil {
		shp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Encoding detection failed: %v", err), 0, "Encoding Detection")
		return fmt.Errorf("encoding detection failed: %w", err)
	}

	// Step 3: Open file
	file, err := os.Open(filePath)
	if err != nil {
		shp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Failed to open file: %v", err), 0, "File I/O")
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Step 4: Get reader with proper encoding
	reader, err := GetReader(file, encoding)
	if err != nil {
		shp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Failed to create reader: %v", err), 0, "Encoding")
		return fmt.Errorf("failed to create reader: %w", err)
	}

	// Step 5: Parse file line by line
	return shp.parseStream(reader, handler)
}

// parseStream performs the actual streaming parsing.
func (shp *StreamingHierarchicalParser) parseStream(reader io.Reader, handler RecordHandler) error {
	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	parentsStack := NewLineStack()
	factory := gedcom.NewRecordFactory()

	// Track current record being built
	var currentRecordLine *gedcom.GedcomLine

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
			shp.errorManager.AddError(gedcom.SeverityWarning, fmt.Sprintf("Malformed line: %v", err), lineNumber, "Line Parsing")
			continue
		}

		// Handle CONC/CONT continuation lines
		if tag == "CONC" || tag == "CONT" {
			if err := shp.continuationHandler.HandleContinuation(tag, level, value); err != nil {
				shp.errorManager.AddError(gedcom.SeverityWarning, fmt.Sprintf("Invalid continuation: %v", err), lineNumber, "CONC/CONT Handling")
				continue
			}
			// Continue accumulating - don't apply yet
			continue
		}

		// Apply accumulated CONC/CONT value to previous line (for non-CONC/CONT lines)
		if shp.continuationHandler.HasAccumulatedValue() {
			if !parentsStack.IsEmpty() {
				topLine := parentsStack.Peek()
				accumulatedValue := shp.continuationHandler.GetAccumulatedValue()
				if topLine.Value != "" {
					topLine.Value += accumulatedValue
				} else {
					topLine.Value = accumulatedValue
				}
			}
			// Reset after applying
			shp.continuationHandler.Reset()
		}

		// Handle level 0 (top-level record)
		if level == 0 {
			// If we have a previous record, apply any accumulated CONC/CONT and yield it
			if currentRecordLine != nil {
				// Apply any remaining accumulated value
				if shp.continuationHandler.HasAccumulatedValue() {
					accumulatedValue := shp.continuationHandler.GetAccumulatedValue()
					if currentRecordLine.Value != "" {
						currentRecordLine.Value += accumulatedValue
					} else {
						currentRecordLine.Value = accumulatedValue
					}
				}
				record := factory.CreateRecord(currentRecordLine)
				if err := handler(record); err != nil {
					return fmt.Errorf("handler error: %w", err)
				}
			}

			// Reset continuation handler for new record
			shp.continuationHandler.Reset()

			// Create new level 0 record
			gedcomLine := gedcom.NewGedcomLine(level, tag, value, xrefID)
			gedcomLine.LineNumber = lineNumber

			// Reset stack for new record
			parentsStack.Clear()
			parentsStack.Push(gedcomLine)

			// Set as current record
			currentRecordLine = gedcomLine

			// Update continuation handler
			shp.continuationHandler.SetLastTag(tag, level)

			continue
		}

		// Handle level > 0 (child lines)
		// Find parent using stack
		parent, err := parentsStack.FindParent(level)
		if err != nil {
			// Orphaned line - no parent found
			shp.errorManager.AddError(gedcom.SeverityWarning, fmt.Sprintf("Orphaned line: %s", err.Error()), lineNumber, "Hierarchy")
			continue
		}

		// Create child line
		childLine := gedcom.NewGedcomLine(level, tag, value, "")
		childLine.LineNumber = lineNumber

		// Add as child to parent
		parent.AddChild(childLine)

		// Push to stack
		parentsStack.Push(childLine)

		// Update continuation handler
		shp.continuationHandler.SetLastTag(tag, level)
	}

	// Handle remaining CONC/CONT value (if file ends with continuation)
	// Apply to the appropriate line in the current record
	if shp.continuationHandler.HasAccumulatedValue() {
		if !parentsStack.IsEmpty() {
			topLine := parentsStack.Peek()
			accumulatedValue := shp.continuationHandler.GetAccumulatedValue()
			if topLine.Value != "" {
				topLine.Value += accumulatedValue
			} else {
				topLine.Value = accumulatedValue
			}
		} else if currentRecordLine != nil {
			// Apply to the record itself if stack is empty
			accumulatedValue := shp.continuationHandler.GetAccumulatedValue()
			if currentRecordLine.Value != "" {
				currentRecordLine.Value += accumulatedValue
			} else {
				currentRecordLine.Value = accumulatedValue
			}
		}
	}

	// Yield the last record if any
	if currentRecordLine != nil {
		record := factory.CreateRecord(currentRecordLine)
		if err := handler(record); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		shp.errorManager.AddError(gedcom.SeveritySevere, fmt.Sprintf("Error reading file: %v", err), lineNumber, "File I/O")
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// GetErrors returns all errors collected during parsing.
func (shp *StreamingHierarchicalParser) GetErrors() []*gedcom.GedcomError {
	return shp.errorManager.Errors()
}

// HasErrors returns true if any errors were encountered.
func (shp *StreamingHierarchicalParser) HasErrors() bool {
	return shp.errorManager.HasErrors()
}

// HasSevereErrors returns true if any severe errors were encountered.
func (shp *StreamingHierarchicalParser) HasSevereErrors() bool {
	return shp.errorManager.HasSevereErrors()
}

// GetErrorManager returns the error manager (for advanced usage).
func (shp *StreamingHierarchicalParser) GetErrorManager() *gedcom.ErrorManager {
	return shp.errorManager
}

// RecordIterator provides an iterator-based API for streaming parsing.
// This allows processing records one at a time using a Next() method.
type RecordIterator struct {
	parser     *StreamingHierarchicalParser
	recordChan chan gedcom.Record
	errorChan  chan error
	done       chan bool
	current    gedcom.Record
	err        error
}

// NewRecordIterator creates a new RecordIterator for the given file.
// The iterator starts parsing in a background goroutine.
func NewRecordIterator(filePath string) (*RecordIterator, error) {
	parser := NewStreamingHierarchicalParser()
	iterator := &RecordIterator{
		parser:     parser,
		recordChan: make(chan gedcom.Record, 10), // Buffered channel
		errorChan:  make(chan error, 1),
		done:       make(chan bool),
	}

	// Start parsing in background
	go func() {
		defer close(iterator.recordChan)
		defer close(iterator.errorChan)

		err := parser.ParseWithHandler(filePath, func(record gedcom.Record) error {
			select {
			case iterator.recordChan <- record:
				return nil
			case <-iterator.done:
				return fmt.Errorf("iterator closed")
			}
		})

		if err != nil {
			select {
			case iterator.errorChan <- err:
			case <-iterator.done:
			}
		}
	}()

	return iterator, nil
}

// Next advances the iterator to the next record and returns true if a record is available.
// Returns false when there are no more records or an error occurred.
func (ri *RecordIterator) Next() bool {
	// Check if we already have an error
	if ri.err != nil {
		return false
	}

	// Try to get a record from the channel
	select {
	case record, ok := <-ri.recordChan:
		if !ok {
			// Channel closed, check for errors
			select {
			case err := <-ri.errorChan:
				if err != nil {
					ri.err = err
				}
			default:
				// No error, just end of records
			}
			return false
		}
		ri.current = record
		return true
	case err := <-ri.errorChan:
		if err != nil {
			ri.err = err
			return false
		}
		// If error was nil, try reading from recordChan again
		select {
		case record, ok := <-ri.recordChan:
			if !ok {
				return false
			}
			ri.current = record
			return true
		default:
			return false
		}
	}
}

// Record returns the current record. Must be called after Next() returns true.
func (ri *RecordIterator) Record() gedcom.Record {
	return ri.current
}

// Error returns any error that occurred during parsing.
func (ri *RecordIterator) Error() error {
	return ri.err
}

// Close stops the iterator and releases resources.
func (ri *RecordIterator) Close() {
	close(ri.done)
	// Drain channels to allow goroutine to exit
	for range ri.recordChan {
	}
}

// GetErrors returns all errors collected during parsing.
func (ri *RecordIterator) GetErrors() []*gedcom.GedcomError {
	return ri.parser.GetErrors()
}
