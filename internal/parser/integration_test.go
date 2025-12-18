package parser

import (
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// TestIntegration_SampleGed tests parsing with the sample.ged file
func TestIntegration_SampleGed(t *testing.T) {
	sampleFile := "../../../family-tree/flask-backend/gedcom/sample.ged"
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", sampleFile)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse sample.ged: %v", err)
	}

	// Verify HEAD record
	header := tree.GetHeader()
	if header == nil {
		t.Fatal("Expected HEAD record")
	}

	// Verify GEDC structure
	gedcLines := header.FirstLine().GetLines("GEDC")
	if len(gedcLines) != 1 {
		t.Errorf("Expected 1 GEDC child, got %d", len(gedcLines))
	}
	versLines := gedcLines[0].GetLines("VERS")
	if len(versLines) != 1 {
		t.Errorf("Expected 1 VERS child, got %d", len(versLines))
	}
	if versLines[0].Value != "5.5.5" {
		t.Errorf("Expected VERS value '5.5.5', got %q", versLines[0].Value)
	}

	// Verify individuals
	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Fatal("Expected INDI @I1@")
	}
	nameValue := indi1.GetValue("NAME")
	if nameValue != "Robert Eugene /Williams/" {
		t.Errorf("Expected NAME 'Robert Eugene /Williams/', got %q", nameValue)
	}

	indi2 := tree.GetIndividual("@I2@")
	if indi2 == nil {
		t.Fatal("Expected INDI @I2@")
	}

	indi3 := tree.GetIndividual("@I3@")
	if indi3 == nil {
		t.Fatal("Expected INDI @I3@")
	}

	// Verify families
	fam1 := tree.GetFamily("@F1@")
	if fam1 == nil {
		t.Fatal("Expected FAM @F1@")
	}
	husbValue := fam1.GetValue("HUSB")
	if husbValue != "@I1@" {
		t.Errorf("Expected HUSB '@I1@', got %q", husbValue)
	}

	fam2 := tree.GetFamily("@F2@")
	if fam2 == nil {
		t.Fatal("Expected FAM @F2@")
	}

	// Verify sources
	sour1 := tree.GetRecordByXref("@S1@")
	if sour1 == nil {
		t.Fatal("Expected SOUR @S1@")
	}
	if sour1.Type() != gedcom.RecordTypeSOUR {
		t.Errorf("Expected SOUR type, got %v", sour1.Type())
	}

	// Verify repositories
	repo1 := tree.GetRecordByXref("@R1@")
	if repo1 == nil {
		t.Fatal("Expected REPO @R1@")
	}
	if repo1.Type() != gedcom.RecordTypeREPO {
		t.Errorf("Expected REPO type, got %v", repo1.Type())
	}

	// Verify submitter
	subm1 := tree.GetRecordByXref("@U1@")
	if subm1 == nil {
		t.Fatal("Expected SUBM @U1@")
	}
	if subm1.Type() != gedcom.RecordTypeSUBM {
		t.Errorf("Expected SUBM type, got %v", subm1.Type())
	}

	// Verify hierarchy depth
	birtLines := indi1.FirstLine().GetLines("BIRT")
	if len(birtLines) != 1 {
		t.Fatalf("Expected 1 BIRT child, got %d", len(birtLines))
	}
	dateLines := birtLines[0].GetLines("DATE")
	if len(dateLines) != 1 {
		t.Fatalf("Expected 1 DATE child under BIRT, got %d", len(dateLines))
	}
	if dateLines[0].Value != "2 Oct 1822" {
		t.Errorf("Expected BIRT DATE '2 Oct 1822', got %q", dateLines[0].Value)
	}

	// Verify SOUR citation hierarchy
	sourLines := birtLines[0].GetLines("SOUR")
	if len(sourLines) != 1 {
		t.Fatalf("Expected 1 SOUR child under BIRT, got %d", len(sourLines))
	}
	if sourLines[0].Value != "@S1@" {
		t.Errorf("Expected SOUR value '@S1@', got %q", sourLines[0].Value)
	}
	pageLines := sourLines[0].GetLines("PAGE")
	if len(pageLines) != 1 {
		t.Fatalf("Expected 1 PAGE child under SOUR, got %d", len(pageLines))
	}
	if pageLines[0].Value != "Sec. 2, p. 45" {
		t.Errorf("Expected PAGE value 'Sec. 2, p. 45', got %q", pageLines[0].Value)
	}

	t.Logf("Successfully parsed sample.ged with %d errors", len(parser.GetErrors()))
}

// TestIntegration_LargeFiles tests parsing with large real-world GEDCOM files
func TestIntegration_LargeFiles(t *testing.T) {
	testFiles := []struct {
		name     string
		filePath string
		minLines int
	}{
		{"gracis.ged", "../../../family-tree/gedcom/gracis.ged", 10000},
		{"xavier.ged", "../../../family-tree/gedcom/xavier.ged", 5000},
		{"tree1.ged", "../../../family-tree/gedcom/tree1.ged", 12000},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.filePath); os.IsNotExist(err) {
				t.Skipf("File not found: %s", tt.filePath)
			}

			parser := NewHierarchicalParser()
			tree, err := parser.Parse(tt.filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tt.name, err)
			}

			if tree == nil {
				t.Fatal("Expected tree to be created")
			}

			// Verify HEAD exists
			header := tree.GetHeader()
			if header == nil {
				t.Error("Expected HEAD record")
			}

			// Verify we have records
			allIndis := tree.GetAllIndividuals()
			allFams := tree.GetAllFamilies()

			if len(allIndis) == 0 && len(allFams) == 0 {
				t.Error("Expected at least some records to be parsed")
			}

			// Log statistics
			errors := parser.GetErrors()
			errorSummary := parser.GetErrorManager().GetErrorSummary()
			t.Logf("%s: %d individuals, %d families, %d total errors (%d warnings, %d severe)",
				tt.name, len(allIndis), len(allFams), len(errors),
				errorSummary[gedcom.SeverityWarning],
				errorSummary[gedcom.SeveritySevere])

			// Verify parsing didn't fail completely
			if parser.HasSevereErrors() {
				severeErrors := parser.GetErrorManager().GetErrorsBySeverity(gedcom.SeveritySevere)
				for _, e := range severeErrors {
					t.Logf("Severe error: %v", e)
				}
				// Severe errors are acceptable if they don't prevent parsing
			}
		})
	}
}

// TestIntegration_TreeStructure verifies the tree structure is correct
func TestIntegration_TreeStructure(t *testing.T) {
	sampleFile := "../../../family-tree/flask-backend/gedcom/sample.ged"
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", sampleFile)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify parent-child relationships
	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Fatal("Expected INDI @I1@")
	}

	// Verify NAME structure
	nameLines := indi1.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Fatalf("Expected 1 NAME child, got %d", len(nameLines))
	}
	nameLine := nameLines[0]
	if nameLine.Parent != indi1.FirstLine() {
		t.Error("NAME parent should be INDI")
	}
	if nameLine.Level != 1 {
		t.Errorf("Expected NAME level 1, got %d", nameLine.Level)
	}

	// Verify GIVN/SURN structure
	givnLines := nameLine.GetLines("GIVN")
	if len(givnLines) != 1 {
		t.Fatalf("Expected 1 GIVN child, got %d", len(givnLines))
	}
	givnLine := givnLines[0]
	if givnLine.Parent != nameLine {
		t.Error("GIVN parent should be NAME")
	}
	if givnLine.Level != 2 {
		t.Errorf("Expected GIVN level 2, got %d", givnLine.Level)
	}

	// Verify BIRT structure
	birtLines := indi1.FirstLine().GetLines("BIRT")
	if len(birtLines) != 1 {
		t.Fatalf("Expected 1 BIRT child, got %d", len(birtLines))
	}
	birtLine := birtLines[0]
	if birtLine.Parent != indi1.FirstLine() {
		t.Error("BIRT parent should be INDI")
	}

	// Verify DATE under BIRT
	dateLines := birtLine.GetLines("DATE")
	if len(dateLines) != 1 {
		t.Fatalf("Expected 1 DATE child under BIRT, got %d", len(dateLines))
	}
	dateLine := dateLines[0]
	if dateLine.Parent != birtLine {
		t.Error("DATE parent should be BIRT")
	}
}

// TestIntegration_ErrorHandling verifies error handling with real files
func TestIntegration_ErrorHandling(t *testing.T) {
	sampleFile := "../../../family-tree/flask-backend/gedcom/sample.ged"
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", sampleFile)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify tree was created despite any errors
	if tree == nil {
		t.Fatal("Expected tree to be created")
	}

			// Check errors
			parseErrors := parser.GetErrors()
			if len(parseErrors) > 0 {
				t.Logf("Found %d errors during parsing:", len(parseErrors))
				for _, e := range parseErrors {
					t.Logf("  %v", e)
				}
			}

	// Verify we can still access records
	header := tree.GetHeader()
	if header == nil {
		t.Error("Expected HEAD record even with errors")
	}
}

// TestIntegration_XrefIndex verifies cross-reference index is correct
func TestIntegration_XrefIndex(t *testing.T) {
	sampleFile := "../../../family-tree/flask-backend/gedcom/sample.ged"
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", sampleFile)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify xrefs are accessible
	xrefs := []string{"@I1@", "@I2@", "@I3@", "@F1@", "@F2@", "@S1@", "@R1@", "@U1@"}

	for _, xref := range xrefs {
		record := tree.GetRecordByXref(xref)
		if record == nil {
			t.Errorf("Expected record %s to be in xref index", xref)
		}
	}

	// Verify xrefs in FAM records
	fam1 := tree.GetFamily("@F1@")
	if fam1 == nil {
		t.Fatal("Expected FAM @F1@")
	}

	husbXref := fam1.GetValue("HUSB")
	if husbXref != "@I1@" {
		t.Errorf("Expected HUSB '@I1@', got %q", husbXref)
	}

	// Verify the xref points to actual record
	husbRecord := tree.GetRecordByXref(husbXref)
	if husbRecord == nil {
		t.Errorf("Expected HUSB xref %s to resolve to a record", husbXref)
	}
}

// TestIntegration_DeepHierarchy verifies deep nesting works correctly
func TestIntegration_DeepHierarchy(t *testing.T) {
	sampleFile := "../../../family-tree/flask-backend/gedcom/sample.ged"
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", sampleFile)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify deep hierarchy: HEAD -> SOUR -> CORP -> ADDR -> CITY
	header := tree.GetHeader()
	if header == nil {
		t.Fatal("Expected HEAD record")
	}

	sourLines := header.FirstLine().GetLines("SOUR")
	if len(sourLines) != 1 {
		t.Fatalf("Expected 1 SOUR child, got %d", len(sourLines))
	}

	corpLines := sourLines[0].GetLines("CORP")
	if len(corpLines) != 1 {
		t.Fatalf("Expected 1 CORP child, got %d", len(corpLines))
	}

	addrLines := corpLines[0].GetLines("ADDR")
	if len(addrLines) != 1 {
		t.Fatalf("Expected 1 ADDR child, got %d", len(addrLines))
	}

	cityLines := addrLines[0].GetLines("CITY")
	if len(cityLines) != 1 {
		t.Fatalf("Expected 1 CITY child, got %d", len(cityLines))
	}

	if cityLines[0].Value != "LEIDEN" {
		t.Errorf("Expected CITY value 'LEIDEN', got %q", cityLines[0].Value)
	}
}

// TestIntegration_Performance tests parsing performance
func TestIntegration_Performance(t *testing.T) {
	testFiles := []struct {
		name     string
		filePath string
	}{
		{"sample.ged", "../../../family-tree/flask-backend/gedcom/sample.ged"},
		{"gracis.ged", "../../../family-tree/gedcom/gracis.ged"},
		{"xavier.ged", "../../../family-tree/gedcom/xavier.ged"},
		{"tree1.ged", "../../../family-tree/gedcom/tree1.ged"},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.filePath); os.IsNotExist(err) {
				t.Skipf("File not found: %s", tt.filePath)
			}

			parser := NewHierarchicalParser()
			tree, err := parser.Parse(tt.filePath)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if tree == nil {
				t.Fatal("Expected tree to be created")
			}

			// Get file info for statistics
			info, err := os.Stat(tt.filePath)
			if err != nil {
				t.Fatalf("Failed to get file info: %v", err)
			}

			allIndis := tree.GetAllIndividuals()
			allFams := tree.GetAllFamilies()

			t.Logf("%s: %d bytes, %d individuals, %d families, %d errors",
				tt.name, info.Size(), len(allIndis), len(allFams), len(parser.GetErrors()))
		})
	}
}

// TestIntegration_EdgeCases tests edge cases with real files
func TestIntegration_EdgeCases(t *testing.T) {
	sampleFile := "../../../family-tree/flask-backend/gedcom/sample.ged"
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", sampleFile)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Test empty selectors
	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Fatal("Expected INDI @I1@")
	}

	// Empty selector should return the record's value (empty for INDI)
	value := indi1.GetValue("")
	if value != "" {
		t.Errorf("Expected empty value for INDI, got %q", value)
	}

	// Test non-existent selectors
	nonExistent := indi1.GetValue("NONEXISTENT")
	if nonExistent != "" {
		t.Errorf("Expected empty value for non-existent selector, got %q", nonExistent)
	}

	// Test nested selectors
	birtDate := indi1.GetValue("BIRT.DATE")
	if birtDate != "2 Oct 1822" {
		t.Errorf("Expected BIRT.DATE '2 Oct 1822', got %q", birtDate)
	}

	// Test multiple values
	nameLines := indi1.FirstLine().GetLines("FAMS")
	if len(nameLines) != 2 {
		t.Errorf("Expected 2 FAMS children, got %d", len(nameLines))
	}
}

