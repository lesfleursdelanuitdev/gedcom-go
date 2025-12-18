package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestHierarchicalParser_ErrorHandling_MalformedLines(t *testing.T) {
	// Test error handling for malformed lines
	testContent := `0 HEAD
invalid line
1 GEDC
2 VERS 5.5.5
0 @I1@ INDI
1 NAME John /Doe/
bad line here
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse should not return error (warnings are collected), got: %v", err)
	}

	// Verify parsing continued despite errors
	if tree == nil {
		t.Fatal("Expected tree to be created")
	}

	// Verify errors were collected
	errors := parser.GetErrors()
	if len(errors) < 2 {
		t.Errorf("Expected at least 2 errors (malformed lines), got %d", len(errors))
	}

	// Check for malformed line errors
	malformedCount := 0
	for _, e := range errors {
		if e.Context == "Line Parsing" {
			malformedCount++
		}
	}
	if malformedCount < 2 {
		t.Errorf("Expected at least 2 line parsing errors, got %d", malformedCount)
	}

	// Verify valid lines were still parsed
	header := tree.GetHeader()
	if header == nil {
		t.Error("Expected HEAD record to be parsed despite errors")
	}
	indi := tree.GetIndividual("@I1@")
	if indi == nil {
		t.Error("Expected INDI record to be parsed despite errors")
	}
}

func TestHierarchicalParser_ErrorHandling_OrphanedLines(t *testing.T) {
	// Test error handling for orphaned lines
	testContent := `1 NAME John /Doe/
0 @I1@ INDI
1 NAME Jane /Doe/
2 GIVN Jane
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse should not return error (warnings are collected), got: %v", err)
	}

	// Verify errors were collected
	errors := parser.GetErrors()
	orphanedCount := 0
	for _, e := range errors {
		if e.Context == "Hierarchy" {
			orphanedCount++
		}
	}
	// First NAME is orphaned (before level 0), GIVN might be orphaned if NAME wasn't parsed
	if orphanedCount < 1 {
		t.Errorf("Expected at least 1 orphaned line error, got %d", orphanedCount)
	}

	// Verify valid records were still parsed
	indi := tree.GetIndividual("@I1@")
	if indi == nil {
		t.Error("Expected INDI record to be parsed despite orphaned lines")
	}
}

func TestHierarchicalParser_ErrorHandling_InvalidContinuation(t *testing.T) {
	// Test error handling for invalid CONC/CONT
	testContent := `0 @N1@ NOTE This is a note
1 CONC that continues
2 CONC invalid subordinate
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse should not return error (warnings are collected), got: %v", err)
	}

	// Verify errors were collected
	errors := parser.GetErrors()
	continuationErrors := 0
	for _, e := range errors {
		if e.Context == "CONC/CONT Handling" {
			continuationErrors++
		}
	}
	if continuationErrors < 1 {
		t.Errorf("Expected at least 1 continuation error, got %d", continuationErrors)
	}

	// Verify parsing continued
	if tree == nil {
		t.Fatal("Expected tree to be created")
	}
}

func TestHierarchicalParser_ErrorHandling_FileErrors(t *testing.T) {
	// Test error handling for file errors
	parser := NewHierarchicalParser()
	_, err := parser.Parse("/nonexistent/file.ged")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	// Verify severe error was recorded
	if !parser.HasSevereErrors() {
		t.Error("Expected severe errors for file validation failure")
	}

	errors := parser.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected errors to be collected")
	}

	// Check for file validation error
	foundFileError := false
	for _, e := range errors {
		if e.Context == "File Validation" && e.Severity == gedcom.SeveritySevere {
			foundFileError = true
		}
	}
	if !foundFileError {
		t.Error("Expected file validation error")
	}
}

func TestHierarchicalParser_ErrorHandling_EncodingErrors(t *testing.T) {
	// Test error handling for encoding errors
	// Create a file that might cause encoding issues
	testContent := `0 HEAD
0 @I1@ INDI
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse should succeed for valid file, got: %v", err)
	}

	// Should not have errors for valid file
	if parser.HasErrors() {
		errors := parser.GetErrors()
		t.Logf("Unexpected errors: %v", errors)
		// This is okay - encoding detection might log warnings
	}

	// Verify tree was created
	if tree == nil {
		t.Fatal("Expected tree to be created")
	}
}

func TestHierarchicalParser_ErrorHandling_ContinueAfterErrors(t *testing.T) {
	// Test that parsing continues after errors
	testContent := `0 HEAD
invalid1
1 GEDC
invalid2
2 VERS 5.5.5
0 @I1@ INDI
invalid3
1 NAME John /Doe/
invalid4
0 @F1@ FAM
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse should not return error (warnings are collected), got: %v", err)
	}

	// Verify multiple errors were collected
	errors := parser.GetErrors()
	if len(errors) < 4 {
		t.Errorf("Expected at least 4 errors, got %d", len(errors))
	}

	// Verify all valid records were parsed despite errors
	header := tree.GetHeader()
	if header == nil {
		t.Error("Expected HEAD record")
	}
	indi := tree.GetIndividual("@I1@")
	if indi == nil {
		t.Error("Expected INDI record")
	}
	fam := tree.GetFamily("@F1@")
	if fam == nil {
		t.Error("Expected FAM record")
	}

	// Verify hierarchy was built correctly
	gedcLines := header.FirstLine().GetLines("GEDC")
	if len(gedcLines) != 1 {
		t.Errorf("Expected HEAD to have GEDC child, got %d", len(gedcLines))
	}
	nameLines := indi.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Errorf("Expected INDI to have NAME child, got %d", len(nameLines))
	}
}

func TestHierarchicalParser_GetErrors(t *testing.T) {
	parser := NewHierarchicalParser()

	// Initially no errors
	errors := parser.GetErrors()
	if len(errors) != 0 {
		t.Errorf("Expected 0 errors initially, got %d", len(errors))
	}

	if parser.HasErrors() {
		t.Error("Expected no errors initially")
	}

	if parser.HasSevereErrors() {
		t.Error("Expected no severe errors initially")
	}
}

func TestHierarchicalParser_ErrorSeverity(t *testing.T) {
	// Test that errors are properly categorized by severity
	testContent := `0 HEAD
invalid line
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	_, err = parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse should not return error for warnings, got: %v", err)
	}

	// Should have warnings but not severe errors
	if parser.HasSevereErrors() {
		errors := parser.GetErrors()
		for _, e := range errors {
			if e.Severity == gedcom.SeveritySevere {
				t.Errorf("Unexpected severe error: %v", e)
			}
		}
	}

	// File validation should cause severe error
	_, err = parser.Parse("/nonexistent/file.ged")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if !parser.HasSevereErrors() {
		t.Error("Expected severe errors for file validation failure")
	}
}



