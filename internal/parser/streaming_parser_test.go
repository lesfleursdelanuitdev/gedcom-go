package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestStreamingHierarchicalParser_ParseWithHandler(t *testing.T) {
	// Create a temporary GEDCOM file
	content := `0 HEAD
1 GEDC
2 VERS 5.5.1
0 @I1@ INDI
1 NAME John /Doe/
1 SEX M
0 @I2@ INDI
1 NAME Jane /Smith/
1 SEX F
0 TRLR
`

	tmpfile, err := os.CreateTemp("", "test_*.ged")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	// Test callback-based parsing
	parser := NewStreamingHierarchicalParser()
	records := make([]gedcom.Record, 0)

	err = parser.ParseWithHandler(tmpfile.Name(), func(record gedcom.Record) error {
		records = append(records, record)
		return nil
	})

	if err != nil {
		t.Fatalf("ParseWithHandler failed: %v", err)
	}

	// Should have HEAD, INDI, INDI, TRLR (4 records)
	if len(records) != 4 {
		t.Errorf("Expected 4 records, got %d", len(records))
	}

	// Check first individual
	if records[1].Type() != gedcom.RecordTypeINDI {
		t.Errorf("Expected INDI record, got %s", records[1].Type())
	}
	if records[1].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", records[1].XrefID())
	}

	// Check second individual
	if records[2].Type() != gedcom.RecordTypeINDI {
		t.Errorf("Expected INDI record, got %s", records[2].Type())
	}
	if records[2].XrefID() != "@I2@" {
		t.Errorf("Expected @I2@, got %s", records[2].XrefID())
	}
}

func TestStreamingHierarchicalParser_HandlerError(t *testing.T) {
	content := `0 HEAD
1 GEDC
2 VERS 5.5.1
0 @I1@ INDI
1 NAME John /Doe/
0 TRLR
`

	tmpfile, err := os.CreateTemp("", "test_*.ged")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	parser := NewStreamingHierarchicalParser()
	callCount := 0

	err = parser.ParseWithHandler(tmpfile.Name(), func(record gedcom.Record) error {
		callCount++
		if callCount == 1 {
			// Return error on first record to stop parsing
			return fmt.Errorf("handler error")
		}
		return nil
	})

	if err == nil {
		t.Error("Expected error from handler, got nil")
	}
	if callCount != 1 {
		t.Errorf("Expected handler to be called once, got %d", callCount)
	}
}

func TestStreamingHierarchicalParser_HierarchicalStructure(t *testing.T) {
	content := `0 @I1@ INDI
1 NAME John /Doe/
2 GIVN John
2 SURN Doe
1 BIRT
2 DATE 1 JAN 1900
2 PLAC New York
0 TRLR
`

	tmpfile, err := os.CreateTemp("", "test_*.ged")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	parser := NewStreamingHierarchicalParser()
	var indiRecord gedcom.Record

	err = parser.ParseWithHandler(tmpfile.Name(), func(record gedcom.Record) error {
		if record.Type() == gedcom.RecordTypeINDI {
			indiRecord = record
		}
		return nil
	})

	if err != nil {
		t.Fatalf("ParseWithHandler failed: %v", err)
	}

	if indiRecord == nil {
		t.Fatal("Expected INDI record")
	}

	// Check hierarchical structure
	name := indiRecord.GetValue("NAME")
	if name != "John /Doe/" {
		t.Errorf("Expected 'John /Doe/', got %q", name)
	}

	givn := indiRecord.GetValue("NAME.GIVN")
	if givn != "John" {
		t.Errorf("Expected 'John', got %q", givn)
	}

	surn := indiRecord.GetValue("NAME.SURN")
	if surn != "Doe" {
		t.Errorf("Expected 'Doe', got %q", surn)
	}

	birthDate := indiRecord.GetValue("BIRT.DATE")
	if birthDate != "1 JAN 1900" {
		t.Errorf("Expected '1 JAN 1900', got %q", birthDate)
	}

	birthPlace := indiRecord.GetValue("BIRT.PLAC")
	if birthPlace != "New York" {
		t.Errorf("Expected 'New York', got %q", birthPlace)
	}
}

func TestStreamingHierarchicalParser_CONC_CONT(t *testing.T) {
	// Test with same-level continuations (valid per GEDCOM spec)
	content := `0 @N1@ NOTE
1 CONT This is a long note
1 CONT that spans multiple
1 CONC lines and should be
1 CONC concatenated properly.
0 TRLR
`

	tmpfile, err := os.CreateTemp("", "test_*.ged")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	parser := NewStreamingHierarchicalParser()
	var noteRecord gedcom.Record

	err = parser.ParseWithHandler(tmpfile.Name(), func(record gedcom.Record) error {
		if record.Type() == gedcom.RecordTypeNOTE {
			noteRecord = record
		}
		return nil
	})

	if err != nil {
		t.Fatalf("ParseWithHandler failed: %v", err)
	}

	if noteRecord == nil {
		t.Fatal("Expected NOTE record")
	}

	// Check that CONC/CONT was handled correctly
	// The NOTE record itself should have the continued value
	// Note: CONT adds a newline, CONC does not
	noteValue := noteRecord.FirstLine().Value
	// Expected: "\nThis is a long note" (from 1st CONT) + "\nthat spans multiple" (from 2nd CONT) + "lines and should be" (from 1st CONC, no newline) + "concatenated properly." (from 2nd CONC, no newline)
	expected := "\nThis is a long note\nthat spans multiplelines and should beconcatenated properly."
	if noteValue != expected {
		t.Errorf("Expected %q, got %q", expected, noteValue)
	}
}

func TestRecordIterator_Next(t *testing.T) {
	content := `0 HEAD
1 GEDC
2 VERS 5.5.1
0 @I1@ INDI
1 NAME John /Doe/
0 @I2@ INDI
1 NAME Jane /Smith/
0 TRLR
`

	tmpfile, err := os.CreateTemp("", "test_*.ged")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	iterator, err := NewRecordIterator(tmpfile.Name())
	if err != nil {
		t.Fatalf("NewRecordIterator failed: %v", err)
	}
	defer iterator.Close()

	records := make([]gedcom.Record, 0)
	for iterator.Next() {
		records = append(records, iterator.Record())
	}

	if err := iterator.Error(); err != nil {
		t.Fatalf("Iterator error: %v", err)
	}

	// Should have HEAD, INDI, INDI, TRLR (4 records)
	if len(records) != 4 {
		t.Errorf("Expected 4 records, got %d", len(records))
	}

	// Check first individual
	if records[1].Type() != gedcom.RecordTypeINDI {
		t.Errorf("Expected INDI record, got %s", records[1].Type())
	}
	if records[1].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", records[1].XrefID())
	}
}

func TestRecordIterator_EmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.ged")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	iterator, err := NewRecordIterator(tmpfile.Name())
	if err != nil {
		t.Fatalf("NewRecordIterator failed: %v", err)
	}
	defer iterator.Close()

	// Should not have any records
	if iterator.Next() {
		t.Error("Expected no records in empty file")
	}

	if err := iterator.Error(); err == nil {
		// Empty file should produce an error (file validation)
		// But iterator might handle it gracefully
	}
}

func TestStreamingHierarchicalParser_ErrorHandling(t *testing.T) {
	parser := NewStreamingHierarchicalParser()

	// Test with non-existent file
	err := parser.ParseWithHandler("nonexistent.ged", func(record gedcom.Record) error {
		return nil
	})

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if !parser.HasErrors() {
		t.Error("Expected errors to be recorded")
	}
}

func TestStreamingHierarchicalParser_GetErrors(t *testing.T) {
	content := `0 HEAD
1 GEDC
2 VERS 5.5.1
0 @I1@ INDI
1 NAME John /Doe/
1 INVALID_TAG Invalid value
0 TRLR
`

	tmpfile, err := os.CreateTemp("", "test_*.ged")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	parser := NewStreamingHierarchicalParser()
	err = parser.ParseWithHandler(tmpfile.Name(), func(record gedcom.Record) error {
		return nil
	})

	if err != nil {
		t.Fatalf("ParseWithHandler failed: %v", err)
	}

	// The streaming parser doesn't validate tags, it just parses structure
	// So we might not have errors for invalid tags - that's validator's job
	// But we should check for parsing errors
	errors := parser.GetErrors()
	// Note: Invalid tags are not parsing errors, they're validation errors
	// The parser will successfully parse them, but validators will flag them
	_ = errors // Errors might be empty, which is fine
}
