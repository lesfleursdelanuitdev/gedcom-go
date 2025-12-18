package parser

import (
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// TestIntegration_SpecializedRecords verifies that specialized record types are created
func TestIntegration_SpecializedRecords(t *testing.T) {
	sampleFile := "../../../family-tree/flask-backend/gedcom/sample.ged"
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", sampleFile)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify HEAD is HeaderRecord
	header := tree.GetHeader()
	if header == nil {
		t.Fatal("Expected HEAD record")
	}
	if _, ok := header.(*gedcom.HeaderRecord); !ok {
		t.Error("Expected HEAD to be HeaderRecord type")
	}
	headerRecord := header.(*gedcom.HeaderRecord)
	if headerRecord.GetGedcomVersion() != "5.5.5" {
		t.Errorf("Expected GEDCOM version '5.5.5', got %q", headerRecord.GetGedcomVersion())
	}

	// Verify INDI is IndividualRecord
	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Fatal("Expected INDI @I1@")
	}
	if _, ok := indi1.(*gedcom.IndividualRecord); !ok {
		t.Error("Expected INDI to be IndividualRecord type")
	}
	indiRecord := indi1.(*gedcom.IndividualRecord)
	if indiRecord.GetName() != "Robert Eugene /Williams/" {
		t.Errorf("Expected name 'Robert Eugene /Williams/', got %q", indiRecord.GetName())
	}
	if indiRecord.GetBirthDate() != "2 Oct 1822" {
		t.Errorf("Expected birth date '2 Oct 1822', got %q", indiRecord.GetBirthDate())
	}

	// Verify FAM is FamilyRecord
	fam1 := tree.GetFamily("@F1@")
	if fam1 == nil {
		t.Fatal("Expected FAM @F1@")
	}
	if _, ok := fam1.(*gedcom.FamilyRecord); !ok {
		t.Error("Expected FAM to be FamilyRecord type")
	}
	famRecord := fam1.(*gedcom.FamilyRecord)
	if famRecord.GetHusband() != "@I1@" {
		t.Errorf("Expected husband '@I1@', got %q", famRecord.GetHusband())
	}
	if famRecord.GetMarriageDate() != "Dec 1859" {
		t.Errorf("Expected marriage date 'Dec 1859', got %q", famRecord.GetMarriageDate())
	}

	// Verify SOUR is SourceRecord
	sour1 := tree.GetRecordByXref("@S1@")
	if sour1 == nil {
		t.Fatal("Expected SOUR @S1@")
	}
	if _, ok := sour1.(*gedcom.SourceRecord); !ok {
		t.Error("Expected SOUR to be SourceRecord type")
	}
	sourRecord := sour1.(*gedcom.SourceRecord)
	title := sourRecord.GetTitle()
	if title != "Madison County Birth, Death, and Marriage Records" {
		t.Errorf("Expected source title 'Madison County Birth, Death, and Marriage Records', got %q", title)
	}

	// Verify REPO is RepositoryRecord
	repo1 := tree.GetRecordByXref("@R1@")
	if repo1 == nil {
		t.Fatal("Expected REPO @R1@")
	}
	if _, ok := repo1.(*gedcom.RepositoryRecord); !ok {
		t.Error("Expected REPO to be RepositoryRecord type")
	}
	repoRecord := repo1.(*gedcom.RepositoryRecord)
	if repoRecord.GetName() != "Family History Library" {
		t.Errorf("Expected repository name 'Family History Library', got %q", repoRecord.GetName())
	}

	// Verify SUBM is SubmitterRecord
	subm1 := tree.GetRecordByXref("@U1@")
	if subm1 == nil {
		t.Fatal("Expected SUBM @U1@")
	}
	if _, ok := subm1.(*gedcom.SubmitterRecord); !ok {
		t.Error("Expected SUBM to be SubmitterRecord type")
	}
	submRecord := subm1.(*gedcom.SubmitterRecord)
	if submRecord.GetName() != "Reldon Poulson" {
		t.Errorf("Expected submitter name 'Reldon Poulson', got %q", submRecord.GetName())
	}
}



