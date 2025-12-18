package validator

import (
	"sync"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// ParallelIndividualValidator validates individuals in parallel.
type ParallelIndividualValidator struct {
	*BaseValidator
	validTags      map[string]bool
	requiredTags   map[string]bool
	eventTags      map[string]bool
	validEventTags map[string]bool
	validNameTags  map[string]bool
	validSexValues map[string]bool
}

// NewParallelIndividualValidator creates a new ParallelIndividualValidator.
func NewParallelIndividualValidator(errorManager *gedcom.ErrorManager) *ParallelIndividualValidator {
	// Reuse the same tag maps as IndividualValidator
	iv := NewIndividualValidator(errorManager)
	return &ParallelIndividualValidator{
		BaseValidator:  iv.BaseValidator,
		validTags:      iv.validTags,
		requiredTags:   iv.requiredTags,
		eventTags:      iv.eventTags,
		validEventTags: iv.validEventTags,
		validNameTags:  iv.validNameTags,
		validSexValues: iv.validSexValues,
	}
}

// Validate validates all individual records in parallel.
func (piv *ParallelIndividualValidator) Validate(tree *gedcom.GedcomTree) error {
	individuals := tree.GetAllIndividuals()
	
	// Use a worker pool pattern for parallel validation
	const numWorkers = 4 // Adjust based on CPU cores
	workChan := make(chan struct {
		xrefID string
		record gedcom.Record
	}, len(individuals))
	
	var wg sync.WaitGroup
	
	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workChan {
				piv.validateIndividual(work.xrefID, work.record)
			}
		}()
	}
	
	// Send work to channel
	for xrefID, record := range individuals {
		workChan <- struct {
			xrefID string
			record gedcom.Record
		}{xrefID, record}
	}
	close(workChan)
	
	// Wait for all workers to complete
	wg.Wait()
	
	return nil
}

// validateIndividual validates a single individual record (same as IndividualValidator).
func (piv *ParallelIndividualValidator) validateIndividual(xrefID string, record gedcom.Record) {
	piv.validateStructure(xrefID, record)
	piv.validateReferences(xrefID, record)
	piv.validateSex(xrefID, record)
	piv.validateEvents(xrefID, record)
	piv.validateNames(xrefID, record)
}

// validateStructure validates the structure and tags of an individual record.
func (piv *ParallelIndividualValidator) validateStructure(xrefID string, record gedcom.Record) {
	tagsPresent := make(map[string]bool)
	firstLine := record.FirstLine()

	// Collect all tags present
	for tag, lines := range firstLine.Children {
		for _, line := range lines {
			if !piv.validTags[line.Tag] && !isUserDefinedTag(line.Tag) {
				piv.AddError(gedcom.SeveritySevere,
					"INDI "+xrefID+": Invalid tag "+line.Tag,
					line.LineNumber,
					"Individual Validation")
			}
			tagsPresent[tag] = true
		}
	}

	// Check required tags
	for requiredTag := range piv.requiredTags {
		if !tagsPresent[requiredTag] {
			piv.AddError(gedcom.SeveritySevere,
				"INDI "+xrefID+": Missing required tag "+requiredTag,
				firstLine.LineNumber,
				"Individual Validation")
		}
	}
}

// validateReferences validates cross-references in an individual record.
func (piv *ParallelIndividualValidator) validateReferences(xrefID string, record gedcom.Record) {
	// Validate FAMS references
	famsRefs := record.GetValues("FAMS")
	for _, famsRef := range famsRefs {
		if !isValidXref(famsRef) {
			famsLines := record.GetLines("FAMS")
			if len(famsLines) > 0 {
				piv.AddError(gedcom.SeveritySevere,
					"INDI "+xrefID+": Invalid FAMS reference format "+famsRef,
					famsLines[0].LineNumber,
					"Individual Validation")
			}
		}
	}

	// Validate FAMC references
	famcRefs := record.GetValues("FAMC")
	for _, famcRef := range famcRefs {
		if !isValidXref(famcRef) {
			famcLines := record.GetLines("FAMC")
			if len(famcLines) > 0 {
				piv.AddError(gedcom.SeveritySevere,
					"INDI "+xrefID+": Invalid FAMC reference format "+famcRef,
					famcLines[0].LineNumber,
					"Individual Validation")
			}
		}
	}
}

// validateSex validates the SEX value.
func (piv *ParallelIndividualValidator) validateSex(xrefID string, record gedcom.Record) {
	sexValue := record.GetValue("SEX")
	if sexValue != "" && !piv.validSexValues[sexValue] {
		sexLines := record.GetLines("SEX")
		if len(sexLines) > 0 {
			piv.AddError(gedcom.SeveritySevere,
				"INDI "+xrefID+": Invalid SEX value "+sexValue,
				sexLines[0].LineNumber,
				"Individual Validation")
		}
	}
}

// validateEvents validates event structures.
func (piv *ParallelIndividualValidator) validateEvents(xrefID string, record gedcom.Record) {
	birthEvents := record.GetLines("BIRT")
	deathEvents := record.GetLines("DEAT")

	// Check for multiple birth events
	if len(birthEvents) > 1 {
		piv.AddError(gedcom.SeverityWarning,
			"INDI "+xrefID+": Multiple BIRT events",
			birthEvents[1].LineNumber,
			"Individual Validation")
	}

	// Check for multiple death events
	if len(deathEvents) > 1 {
		piv.AddError(gedcom.SeverityWarning,
			"INDI "+xrefID+": Multiple DEAT events",
			deathEvents[1].LineNumber,
			"Individual Validation")
	}

	// Validate each event structure
	for eventTag := range piv.eventTags {
		eventLines := record.GetLines(eventTag)
		for _, eventLine := range eventLines {
			piv.validateEventStructure(xrefID, eventTag, eventLine)
		}
	}
}

// validateEventStructure validates the structure of an event.
func (piv *ParallelIndividualValidator) validateEventStructure(xrefID, eventTag string, eventLine *gedcom.GedcomLine) {
	for _, lines := range eventLine.Children {
		for _, line := range lines {
			if !piv.validEventTags[line.Tag] && !isUserDefinedTag(line.Tag) {
				piv.AddError(gedcom.SeverityWarning,
					"INDI "+xrefID+": Invalid subtag "+line.Tag+" in "+eventTag+" event",
					line.LineNumber,
					"Individual Validation")
			}
		}
	}
}

// validateNames validates name structures.
func (piv *ParallelIndividualValidator) validateNames(xrefID string, record gedcom.Record) {
	nameLines := record.GetLines("NAME")

	if len(nameLines) == 0 {
		piv.AddError(gedcom.SeveritySevere,
			"INDI "+xrefID+": Missing NAME tag",
			record.FirstLine().LineNumber,
			"Individual Validation")
		return
	}

	for _, nameLine := range nameLines {
		piv.validateNameStructure(xrefID, nameLine)
	}
}

// validateNameStructure validates the structure of a name.
func (piv *ParallelIndividualValidator) validateNameStructure(xrefID string, nameLine *gedcom.GedcomLine) {
	for _, lines := range nameLine.Children {
		for _, line := range lines {
			if !piv.validNameTags[line.Tag] && !isUserDefinedTag(line.Tag) {
				piv.AddError(gedcom.SeverityWarning,
					"INDI "+xrefID+": Invalid subtag "+line.Tag+" in NAME structure",
					line.LineNumber,
					"Individual Validation")
			}
		}
	}
}


