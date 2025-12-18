package validator

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// CrossReferenceValidator validates cross-references between records.
type CrossReferenceValidator struct {
	*BaseValidator
}

// NewCrossReferenceValidator creates a new CrossReferenceValidator.
func NewCrossReferenceValidator(errorManager *gedcom.ErrorManager) *CrossReferenceValidator {
	return &CrossReferenceValidator{
		BaseValidator: NewBaseValidator(errorManager),
	}
}

// Validate validates all cross-references in the tree.
func (crv *CrossReferenceValidator) Validate(tree *gedcom.GedcomTree) error {
	crv.validateXrefIDs(tree)
	crv.validateCrossReferences(tree)
	return nil
}

// validateXrefIDs validates the format of all xref IDs.
func (crv *CrossReferenceValidator) validateXrefIDs(tree *gedcom.GedcomTree) {
	// Validate individual xrefs
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		if !isValidXrefID(xrefID) {
			crv.AddError(gedcom.SeveritySevere,
				"Invalid cross-reference ID: "+xrefID+" in INDI record",
				record.FirstLine().LineNumber,
				"Cross-Reference Validation")
		}
	}

	// Validate family xrefs
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		if !isValidXrefID(xrefID) {
			crv.AddError(gedcom.SeveritySevere,
				"Invalid cross-reference ID: "+xrefID+" in FAM record",
				record.FirstLine().LineNumber,
				"Cross-Reference Validation")
		}
	}

	// Validate other record type xrefs
	allRecords := []struct {
		xrefID string
		record gedcom.Record
		typ    string
	}{}

	// Collect all records with xrefs
	for xrefID, record := range individuals {
		allRecords = append(allRecords, struct {
			xrefID string
			record gedcom.Record
			typ    string
		}{xrefID, record, "INDI"})
	}
	for xrefID, record := range families {
		allRecords = append(allRecords, struct {
			xrefID string
			record gedcom.Record
			typ    string
		}{xrefID, record, "FAM"})
	}

	// Check notes, sources, repositories, submitters, multimedia
	// We'll need to add methods to tree to get all of these
	// For now, we'll validate what we can access
}

// validateCrossReferences validates that all cross-references point to existing records.
func (crv *CrossReferenceValidator) validateCrossReferences(tree *gedcom.GedcomTree) {
	crv.validateFamilyReferences(tree)
	crv.validateIndividualReferences(tree)
}

// validateFamilyReferences validates references in family records.
func (crv *CrossReferenceValidator) validateFamilyReferences(tree *gedcom.GedcomTree) {
	families := tree.GetAllFamilies()
	individuals := tree.GetAllIndividuals()

	for xrefID, record := range families {
		// Validate HUSB reference
		husbValue := record.GetValue("HUSB")
		if husbValue != "" {
			if _, exists := individuals[husbValue]; !exists {
				husbLines := record.GetLines("HUSB")
				if len(husbLines) > 0 {
					crv.AddError(gedcom.SeveritySevere,
						"Invalid cross-reference: "+husbValue+" in FAM record "+xrefID,
						husbLines[0].LineNumber,
						"Cross-Reference Validation")
				}
			}
		}

		// Validate WIFE reference
		wifeValue := record.GetValue("WIFE")
		if wifeValue != "" {
			if _, exists := individuals[wifeValue]; !exists {
				wifeLines := record.GetLines("WIFE")
				if len(wifeLines) > 0 {
					crv.AddError(gedcom.SeveritySevere,
						"Invalid cross-reference: "+wifeValue+" in FAM record "+xrefID,
						wifeLines[0].LineNumber,
						"Cross-Reference Validation")
				}
			}
		}

		// Validate CHIL references
		chilRefs := record.GetValues("CHIL")
		for _, chilRef := range chilRefs {
			if _, exists := individuals[chilRef]; !exists {
				chilLines := record.GetLines("CHIL")
				if len(chilLines) > 0 {
					crv.AddError(gedcom.SeveritySevere,
						"Invalid cross-reference: "+chilRef+" in FAM record "+xrefID,
						chilLines[0].LineNumber,
						"Cross-Reference Validation")
				}
			}
		}
	}
}

// validateIndividualReferences validates references in individual records.
func (crv *CrossReferenceValidator) validateIndividualReferences(tree *gedcom.GedcomTree) {
	individuals := tree.GetAllIndividuals()
	families := tree.GetAllFamilies()

	for xrefID, record := range individuals {
		// Validate FAMS references
		famsRefs := record.GetValues("FAMS")
		for _, famsRef := range famsRefs {
			if _, exists := families[famsRef]; !exists {
				famsLines := record.GetLines("FAMS")
				if len(famsLines) > 0 {
					crv.AddError(gedcom.SeveritySevere,
						"Invalid cross-reference: "+famsRef+" in INDI record "+xrefID,
						famsLines[0].LineNumber,
						"Cross-Reference Validation")
				}
			}
		}

		// Validate FAMC references
		famcRefs := record.GetValues("FAMC")
		for _, famcRef := range famcRefs {
			if _, exists := families[famcRef]; !exists {
				famcLines := record.GetLines("FAMC")
				if len(famcLines) > 0 {
					crv.AddError(gedcom.SeveritySevere,
						"Invalid cross-reference: "+famcRef+" in INDI record "+xrefID,
						famcLines[0].LineNumber,
						"Cross-Reference Validation")
				}
			}
		}
	}
}
