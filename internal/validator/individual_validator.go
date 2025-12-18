package validator

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// IndividualValidator validates Individual (INDI) records.
type IndividualValidator struct {
	*BaseValidator
	validTags      map[string]bool
	requiredTags   map[string]bool
	eventTags      map[string]bool
	validEventTags map[string]bool
	validNameTags  map[string]bool
	validSexValues map[string]bool
}

// NewIndividualValidator creates a new IndividualValidator.
func NewIndividualValidator(errorManager *gedcom.ErrorManager) *IndividualValidator {
	validTags := map[string]bool{
		"RESN": true, "NAME": true, "SEX": true, "ALIA": true, "ASSO": true, "ANCI": true, "DESI": true,
		"RFN": true, "AFN": true, "REFN": true, "RIN": true, "CHAN": true, "NOTE": true, "SOUR": true,
		"OBJE": true, "FAMS": true, "FAMC": true, "CAST": true, "DSCR": true, "EDUC": true, "IDNO": true,
		"NATI": true, "NCHI": true, "NMR": true, "OCCU": true, "PROP": true, "RELI": true, "RESI": true,
		"TITL": true, "FACT": true, "BIRT": true, "CHR": true, "DEAT": true, "BURI": true, "CREM": true,
		"ADOP": true, "BAPM": true, "BARM": true, "BASM": true, "BLES": true, "CHRA": true, "CONF": true,
		"FCOM": true, "ORDN": true, "NATU": true, "EMIG": true, "IMMI": true, "CENS": true, "PROB": true,
		"WILL": true, "GRAD": true, "RETI": true, "EVEN": true, "BAPL": true, "CONL": true, "ENDL": true,
		"SLGC": true,
	}

	requiredTags := map[string]bool{
		"NAME": true,
	}

	eventTags := map[string]bool{
		"BIRT": true, "CHR": true, "DEAT": true, "BURI": true, "CREM": true, "ADOP": true, "BAPM": true,
		"BARM": true, "BASM": true, "BLES": true, "CHRA": true, "CONF": true, "FCOM": true, "ORDN": true,
		"NATU": true, "EMIG": true, "IMMI": true, "CENS": true, "PROB": true, "WILL": true, "GRAD": true,
		"RETI": true, "EVEN": true,
	}

	validEventTags := map[string]bool{
		"TYPE": true, "DATE": true, "PLAC": true, "ADDR": true, "AGE": true, "AGNC": true, "CAUS": true,
		"SOUR": true, "NOTE": true, "OBJE": true,
	}

	validNameTags := map[string]bool{
		"NPFX": true, "GIVN": true, "NICK": true, "SPFX": true, "SURN": true, "NSFX": true, "SOUR": true,
		"NOTE": true, "FONE": true, "ROMN": true,
	}

	validSexValues := map[string]bool{
		"M": true, "F": true, "U": true, "X": true, "N": true,
	}

	return &IndividualValidator{
		BaseValidator:  NewBaseValidator(errorManager),
		validTags:      validTags,
		requiredTags:   requiredTags,
		eventTags:      eventTags,
		validEventTags: validEventTags,
		validNameTags:  validNameTags,
		validSexValues: validSexValues,
	}
}

// Validate validates all individual records in the tree.
func (iv *IndividualValidator) Validate(tree *gedcom.GedcomTree) error {
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		iv.validateIndividual(xrefID, record)
	}
	return nil
}

// validateIndividual validates a single individual record.
func (iv *IndividualValidator) validateIndividual(xrefID string, record gedcom.Record) {
	iv.validateStructure(xrefID, record)
	iv.validateReferences(xrefID, record)
	iv.validateSex(xrefID, record)
	iv.validateEvents(xrefID, record)
	iv.validateNames(xrefID, record)
}

// validateStructure validates the structure and tags of an individual record.
func (iv *IndividualValidator) validateStructure(xrefID string, record gedcom.Record) {
	config := BasicValidationConfig{
		RecordTypePrefix: "INDI",
		ValidTags:        iv.validTags,
		RequiredTags:     iv.requiredTags,
		Context:          "Individual Validation",
	}
	validateStructureGeneric(xrefID, record, config, iv.GetErrorManager())
}

// validateReferences validates cross-references in an individual record.
func (iv *IndividualValidator) validateReferences(xrefID string, record gedcom.Record) {
	// Validate FAMS references
	famsRefs := record.GetValues("FAMS")
	validateXrefReferencesGeneric(
		xrefID, "FAMS", famsRefs, record,
		"INDI", "Individual Validation",
		iv.GetErrorManager(),
	)

	// Validate FAMC references
	famcRefs := record.GetValues("FAMC")
	validateXrefReferencesGeneric(
		xrefID, "FAMC", famcRefs, record,
		"INDI", "Individual Validation",
		iv.GetErrorManager(),
	)
}

// validateSex validates the SEX value.
func (iv *IndividualValidator) validateSex(xrefID string, record gedcom.Record) {
	sexValue := record.GetValue("SEX")
	validateTagValueGeneric(
		xrefID, "SEX", sexValue, iv.validSexValues,
		record, "INDI", "Individual Validation",
		iv.GetErrorManager(),
	)
}

// validateEvents validates event structures.
func (iv *IndividualValidator) validateEvents(xrefID string, record gedcom.Record) {
	birthEvents := record.GetLines("BIRT")
	deathEvents := record.GetLines("DEAT")

	// Check for multiple birth events
	validateMultipleEventsGeneric(
		xrefID, "BIRT", birthEvents,
		"INDI", "Individual Validation",
		iv.GetErrorManager(),
	)

	// Check for multiple death events
	validateMultipleEventsGeneric(
		xrefID, "DEAT", deathEvents,
		"INDI", "Individual Validation",
		iv.GetErrorManager(),
	)

	// Validate each event structure
	validateEventsGeneric(
		xrefID, record,
		iv.eventTags, iv.validEventTags,
		"INDI", "Individual Validation",
		iv.GetErrorManager(),
	)
}

// validateEventStructure validates the structure of an event.
func (iv *IndividualValidator) validateEventStructure(xrefID, eventTag string, eventLine *gedcom.GedcomLine) {
	validateEventStructureGeneric(
		xrefID, eventTag, eventLine,
		iv.validEventTags,
		"INDI", "Individual Validation",
		iv.GetErrorManager(),
	)
}

// validateNames validates name structures.
func (iv *IndividualValidator) validateNames(xrefID string, record gedcom.Record) {
	nameLines := record.GetLines("NAME")

	if len(nameLines) == 0 {
		iv.AddError(gedcom.SeveritySevere,
			"INDI "+xrefID+": Missing NAME tag",
			record.FirstLine().LineNumber,
			"Individual Validation")
		return
	}

	for _, nameLine := range nameLines {
		iv.validateNameStructure(xrefID, nameLine)
	}
}

// validateNameStructure validates the structure of a name.
func (iv *IndividualValidator) validateNameStructure(xrefID string, nameLine *gedcom.GedcomLine) {
	validateSubtagStructureGeneric(
		xrefID, "NAME", nameLine,
		iv.validNameTags,
		"INDI", "Individual Validation",
		iv.GetErrorManager(),
	)
}
