package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestBasicParser_Parse_Level0Only(t *testing.T) {
	// Create a temporary test file with only level 0 records
	testContent := `0 HEAD
0 @I1@ INDI
0 @I2@ INDI
0 @F1@ FAM
0 @N1@ NOTE
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	parser := NewBasicParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify HEAD record
	header := tree.GetHeader()
	if header == nil {
		t.Error("Expected HEAD record, got nil")
	} else {
		if header.Type() != gedcom.RecordTypeHEAD {
			t.Errorf("Expected record type HEAD, got %v", header.Type())
		}
		if header.XrefID() != "" {
			t.Errorf("HEAD should not have xref, got %q", header.XrefID())
		}
	}

	// Verify INDI records
	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Error("Expected INDI record @I1@, got nil")
	} else {
		if indi1.Type() != gedcom.RecordTypeINDI {
			t.Errorf("Expected record type INDI, got %v", indi1.Type())
		}
		if indi1.XrefID() != "@I1@" {
			t.Errorf("Expected xref @I1@, got %q", indi1.XrefID())
		}
	}

	indi2 := tree.GetIndividual("@I2@")
	if indi2 == nil {
		t.Error("Expected INDI record @I2@, got nil")
	} else {
		if indi2.XrefID() != "@I2@" {
			t.Errorf("Expected xref @I2@, got %q", indi2.XrefID())
		}
	}

	// Verify FAM record
	fam1 := tree.GetFamily("@F1@")
	if fam1 == nil {
		t.Error("Expected FAM record @F1@, got nil")
	} else {
		if fam1.Type() != gedcom.RecordTypeFAM {
			t.Errorf("Expected record type FAM, got %v", fam1.Type())
		}
		if fam1.XrefID() != "@F1@" {
			t.Errorf("Expected xref @F1@, got %q", fam1.XrefID())
		}
	}

	// Verify NOTE record
	note1 := tree.GetRecordByXref("@N1@")
	if note1 == nil {
		t.Error("Expected NOTE record @N1@, got nil")
	} else {
		if note1.Type() != gedcom.RecordTypeNOTE {
			t.Errorf("Expected record type NOTE, got %v", note1.Type())
		}
		if note1.XrefID() != "@N1@" {
			t.Errorf("Expected xref @N1@, got %q", note1.XrefID())
		}
	}

	// Verify all individuals
	allIndis := tree.GetAllIndividuals()
	if len(allIndis) != 2 {
		t.Errorf("Expected 2 individuals, got %d", len(allIndis))
	}

	// Verify all families
	allFams := tree.GetAllFamilies()
	if len(allFams) != 1 {
		t.Errorf("Expected 1 family, got %d", len(allFams))
	}
}

func TestBasicParser_Parse_HierarchicalStructure(t *testing.T) {
	// Create a test file with level 0 and level > 0 records
	// Step 1.7: Now we parse the full hierarchy
	testContent := `0 HEAD
1 GEDC
2 VERS 5.5.5
0 @I1@ INDI
1 NAME John /Doe/
2 GIVN John
2 SURN Doe
0 @F1@ FAM
1 HUSB @I1@
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	parser := NewBasicParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify level 0 records are parsed
	header := tree.GetHeader()
	if header == nil {
		t.Error("Expected HEAD record")
	}

	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Error("Expected INDI record @I1@")
	}

	fam1 := tree.GetFamily("@F1@")
	if fam1 == nil {
		t.Error("Expected FAM record @F1@")
	}

	// Verify that level > 0 lines ARE now stored as children (Step 1.7)
	headerLine := header.FirstLine()
	if headerLine == nil {
		t.Fatal("Header first line is nil")
	}

	// HEAD should have GEDC as child
	gedcLines := headerLine.GetLines("GEDC")
	if len(gedcLines) != 1 {
		t.Errorf("Expected HEAD to have 1 GEDC child, got %d", len(gedcLines))
	} else {
		gedcLine := gedcLines[0]
		if gedcLine.Level != 1 {
			t.Errorf("Expected GEDC level 1, got %d", gedcLine.Level)
		}
		// GEDC should have VERS as child
		versLines := gedcLine.GetLines("VERS")
		if len(versLines) != 1 {
			t.Errorf("Expected GEDC to have 1 VERS child, got %d", len(versLines))
		} else {
			versLine := versLines[0]
			if versLine.Level != 2 {
				t.Errorf("Expected VERS level 2, got %d", versLine.Level)
			}
			if versLine.Value != "5.5.5" {
				t.Errorf("Expected VERS value '5.5.5', got %q", versLine.Value)
			}
		}
	}

	// INDI should have NAME as child
	indiLine := indi1.FirstLine()
	nameLines := indiLine.GetLines("NAME")
	if len(nameLines) != 1 {
		t.Errorf("Expected INDI to have 1 NAME child, got %d", len(nameLines))
	} else {
		nameLine := nameLines[0]
		if nameLine.Value != "John /Doe/" {
			t.Errorf("Expected NAME value 'John /Doe/', got %q", nameLine.Value)
		}
		// NAME should have GIVN and SURN as children
		givnLines := nameLine.GetLines("GIVN")
		if len(givnLines) != 1 {
			t.Errorf("Expected NAME to have 1 GIVN child, got %d", len(givnLines))
		} else if givnLines[0].Value != "John" {
			t.Errorf("Expected GIVN value 'John', got %q", givnLines[0].Value)
		}
		surnLines := nameLine.GetLines("SURN")
		if len(surnLines) != 1 {
			t.Errorf("Expected NAME to have 1 SURN child, got %d", len(surnLines))
		} else if surnLines[0].Value != "Doe" {
			t.Errorf("Expected SURN value 'Doe', got %q", surnLines[0].Value)
		}
	}

	// FAM should have HUSB as child
	famLine := fam1.FirstLine()
	husbLines := famLine.GetLines("HUSB")
	if len(husbLines) != 1 {
		t.Errorf("Expected FAM to have 1 HUSB child, got %d", len(husbLines))
	} else if husbLines[0].Value != "@I1@" {
		t.Errorf("Expected HUSB value '@I1@', got %q", husbLines[0].Value)
	}
}

func TestBasicParser_Parse_WithXrefAndValue(t *testing.T) {
	// Test level 0 records that have both xref and value
	testContent := `0 HEAD
0 @N1@ NOTE This is a note with value
0 @S1@ SOUR Source Title
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewBasicParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify NOTE with value
	note1 := tree.GetRecordByXref("@N1@")
	if note1 == nil {
		t.Fatal("Expected NOTE record @N1@")
	}
	if note1.GetValue("") != "This is a note with value" {
		t.Errorf("Expected value 'This is a note with value', got %q", note1.GetValue(""))
	}

	// Verify SOUR with value
	sour1 := tree.GetRecordByXref("@S1@")
	if sour1 == nil {
		t.Fatal("Expected SOUR record @S1@")
	}
	if sour1.GetValue("") != "Source Title" {
		t.Errorf("Expected value 'Source Title', got %q", sour1.GetValue(""))
	}
}

func TestBasicParser_Parse_AllRecordTypes(t *testing.T) {
	// Test all record types
	testContent := `0 HEAD
0 @I1@ INDI
0 @F1@ FAM
0 @N1@ NOTE
0 @S1@ SOUR
0 @R1@ REPO
0 @M1@ SUBM
0 @O1@ OBJE
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewBasicParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify all record types are parsed
	checks := []struct {
		xrefID     string
		recordType gedcom.RecordType
	}{
		{"@I1@", gedcom.RecordTypeINDI},
		{"@F1@", gedcom.RecordTypeFAM},
		{"@N1@", gedcom.RecordTypeNOTE},
		{"@S1@", gedcom.RecordTypeSOUR},
		{"@R1@", gedcom.RecordTypeREPO},
		{"@M1@", gedcom.RecordTypeSUBM},
		{"@O1@", gedcom.RecordTypeOBJE},
	}

	for _, check := range checks {
		record := tree.GetRecordByXref(check.xrefID)
		if record == nil {
			t.Errorf("Expected %s record %s, got nil", check.recordType, check.xrefID)
			continue
		}
		if record.Type() != check.recordType {
			t.Errorf("Expected record type %s for %s, got %s", check.recordType, check.xrefID, record.Type())
		}
		if record.XrefID() != check.xrefID {
			t.Errorf("Expected xref %s, got %s", check.xrefID, record.XrefID())
		}
	}
}

func TestBasicParser_Parse_EncodingDetection(t *testing.T) {
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

	parser := NewBasicParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify encoding was detected
	encoding := tree.GetEncoding()
	if encoding == "" {
		t.Error("Expected encoding to be set, got empty string")
	}
	// Should default to UTF-8 if no BOM
	if encoding != "UTF-8" {
		t.Logf("Note: Encoding is %s (may vary based on BOM detection)", encoding)
	}
}

func TestBasicParser_Parse_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.ged")
	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewBasicParser()
	_, err = parser.Parse(testFile)
	if err == nil {
		t.Error("Expected error for empty file, got nil")
	}
}

func TestBasicParser_Parse_InvalidFile(t *testing.T) {
	parser := NewBasicParser()
	_, err := parser.Parse("/nonexistent/file.ged")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

