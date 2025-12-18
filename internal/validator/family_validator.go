package validator

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// FamilyValidator validates Family (FAM) records.
type FamilyValidator struct {
	*BaseValidator
	validTags      map[string]bool
	requiredTags   map[string]bool
	eventTags      map[string]bool
	validEventTags map[string]bool
}

// NewFamilyValidator creates a new FamilyValidator.
func NewFamilyValidator(errorManager *gedcom.ErrorManager) *FamilyValidator {
	validTags := map[string]bool{
		"RESN": true, "FAMS": true, "FAMC": true, "HUSB": true, "WIFE": true, "CHIL": true, "NCHI": true,
		"SUBM": true, "SLGS": true, "REFN": true, "RIN": true, "CHAN": true, "NOTE": true, "SOUR": true,
		"OBJE": true, "ANUL": true, "CENS": true, "DIV": true, "DIVF": true, "ENGA": true, "MARB": true,
		"MARC": true, "MARR": true, "MARL": true, "MARS": true, "EVEN": true,
	}

	// Note: Family records require at least one of HUSB or WIFE, not both
	// This is handled in validateStructure with custom logic
	requiredTags := map[string]bool{
		// Empty - we handle required tags specially in validateStructure
	}

	eventTags := map[string]bool{
		"ANUL": true, "CENS": true, "DIV": true, "DIVF": true, "ENGA": true, "MARB": true, "MARC": true,
		"MARR": true, "MARL": true, "MARS": true, "EVEN": true,
	}

	validEventTags := map[string]bool{
		"TYPE": true, "DATE": true, "PLAC": true, "ADDR": true, "AGE": true, "AGNC": true, "CAUS": true,
		"SOUR": true, "NOTE": true, "OBJE": true,
	}

	return &FamilyValidator{
		BaseValidator:  NewBaseValidator(errorManager),
		validTags:      validTags,
		requiredTags:   requiredTags,
		eventTags:      eventTags,
		validEventTags: validEventTags,
	}
}

// Validate validates all family records in the tree.
func (fv *FamilyValidator) Validate(tree *gedcom.GedcomTree) error {
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		fv.validateFamily(xrefID, record)
	}
	return nil
}

// validateFamily validates a single family record.
func (fv *FamilyValidator) validateFamily(xrefID string, record gedcom.Record) {
	fv.validateStructure(xrefID, record)
	fv.validateReferences(xrefID, record)
	fv.validateEvents(xrefID, record)
}

// validateStructure validates the structure and tags of a family record.
func (fv *FamilyValidator) validateStructure(xrefID string, record gedcom.Record) {
	firstLine := record.FirstLine()
	tagsPresent := make(map[string]bool)

	// Collect all tags present and validate them
	for tag, lines := range firstLine.Children {
		for _, line := range lines {
			if !fv.validTags[line.Tag] && !isUserDefinedTag(line.Tag) {
				fv.AddError(gedcom.SeveritySevere,
					"FAM "+xrefID+": Invalid tag "+line.Tag,
					line.LineNumber,
					"Family Validation")
			}
			tagsPresent[tag] = true
		}
	}

	// Family-specific validation: at least one of HUSB or WIFE should be present
	// (This is different from the generic required tags check)
	hasHusband := tagsPresent["HUSB"]
	hasWife := tagsPresent["WIFE"]

	if !hasHusband && !hasWife {
		fv.AddError(gedcom.SeveritySevere,
			"FAM "+xrefID+": Missing required tags (must have at least HUSB or WIFE)",
			firstLine.LineNumber,
			"Family Validation")
	}
}

// validateReferences validates cross-references in a family record.
func (fv *FamilyValidator) validateReferences(xrefID string, record gedcom.Record) {
	// Validate HUSB reference
	husbValue := record.GetValue("HUSB")
	validateXrefReferenceGeneric(
		xrefID, "HUSB", husbValue, record,
		"FAM", "Family Validation",
		fv.GetErrorManager(),
	)

	// Validate WIFE reference
	wifeValue := record.GetValue("WIFE")
	validateXrefReferenceGeneric(
		xrefID, "WIFE", wifeValue, record,
		"FAM", "Family Validation",
		fv.GetErrorManager(),
	)

	// Validate CHIL references
	chilRefs := record.GetValues("CHIL")
	validateXrefReferencesGeneric(
		xrefID, "CHIL", chilRefs, record,
		"FAM", "Family Validation",
		fv.GetErrorManager(),
	)
}

// validateEvents validates event structures.
func (fv *FamilyValidator) validateEvents(xrefID string, record gedcom.Record) {
	marriageEvents := record.GetLines("MARR")
	divorceEvents := record.GetLines("DIV")

	// Check for multiple marriage events
	validateMultipleEventsGeneric(
		xrefID, "MARR", marriageEvents,
		"FAM", "Family Validation",
		fv.GetErrorManager(),
	)

	// Check for multiple divorce events
	validateMultipleEventsGeneric(
		xrefID, "DIV", divorceEvents,
		"FAM", "Family Validation",
		fv.GetErrorManager(),
	)

	// Validate each event structure
	validateEventsGeneric(
		xrefID, record,
		fv.eventTags, fv.validEventTags,
		"FAM", "Family Validation",
		fv.GetErrorManager(),
	)
}

// validateEventStructure validates the structure of an event.
func (fv *FamilyValidator) validateEventStructure(xrefID, eventTag string, eventLine *gedcom.GedcomLine) {
	validateEventStructureGeneric(
		xrefID, eventTag, eventLine,
		fv.validEventTags,
		"FAM", "Family Validation",
		fv.GetErrorManager(),
	)
}
