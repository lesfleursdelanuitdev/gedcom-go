package validator

import (
	"github.com/yourorg/gedcom/pkg/gedcom"
)

// BasicValidationConfig holds configuration for basic generic validation functions.
type BasicValidationConfig struct {
	RecordTypePrefix string // "INDI", "FAM", etc.
	ValidTags        map[string]bool
	RequiredTags     map[string]bool
	Context          string // Error context
}

// validateStructureGeneric validates the structure and tags of a record using generics.
// This reduces duplication between IndividualValidator and FamilyValidator.
func validateStructureGeneric(
	xrefID string,
	record gedcom.Record,
	config BasicValidationConfig,
	errorManager *gedcom.ErrorManager,
) {
	tagsPresent := make(map[string]bool)
	firstLine := record.FirstLine()

	// Collect all tags present
	for tag, lines := range firstLine.Children {
		for _, line := range lines {
			if !config.ValidTags[line.Tag] && !isUserDefinedTag(line.Tag) {
				errorManager.AddError(gedcom.SeveritySevere,
					config.RecordTypePrefix+" "+xrefID+": Invalid tag "+line.Tag,
					line.LineNumber,
					config.Context)
			}
			tagsPresent[tag] = true
		}
	}

	// Check required tags
	for requiredTag := range config.RequiredTags {
		if !tagsPresent[requiredTag] {
			errorManager.AddError(gedcom.SeveritySevere,
				config.RecordTypePrefix+" "+xrefID+": Missing required tag "+requiredTag,
				firstLine.LineNumber,
				config.Context)
		}
	}
}

// validateEventStructureGeneric validates the structure of an event using generics.
// This reduces duplication between IndividualValidator and FamilyValidator.
func validateEventStructureGeneric(
	xrefID, eventTag string,
	eventLine *gedcom.GedcomLine,
	validEventTags map[string]bool,
	recordTypePrefix, context string,
	errorManager *gedcom.ErrorManager,
) {
	for _, lines := range eventLine.Children {
		for _, line := range lines {
			if !validEventTags[line.Tag] && !isUserDefinedTag(line.Tag) {
				errorManager.AddError(gedcom.SeverityWarning,
					recordTypePrefix+" "+xrefID+": Invalid subtag "+line.Tag+" in "+eventTag+" event",
					line.LineNumber,
					context)
			}
		}
	}
}

// validateXrefReferenceGeneric validates a single xref reference using generics.
// This reduces duplication in reference validation.
func validateXrefReferenceGeneric(
	xrefID, tagName, value string,
	record gedcom.Record,
	recordTypePrefix, context string,
	errorManager *gedcom.ErrorManager,
) {
	if value != "" && !isValidXref(value) {
		lines := record.GetLines(tagName)
		if len(lines) > 0 {
			errorManager.AddError(gedcom.SeveritySevere,
				recordTypePrefix+" "+xrefID+": Invalid "+tagName+" reference format "+value,
				lines[0].LineNumber,
				context)
		}
	}
}

// validateXrefReferencesGeneric validates multiple xref references using generics.
// This reduces duplication in reference validation.
func validateXrefReferencesGeneric(
	xrefID, tagName string,
	values []string,
	record gedcom.Record,
	recordTypePrefix, context string,
	errorManager *gedcom.ErrorManager,
) {
	for _, value := range values {
		if !isValidXref(value) {
			lines := record.GetLines(tagName)
			if len(lines) > 0 {
				errorManager.AddError(gedcom.SeveritySevere,
					recordTypePrefix+" "+xrefID+": Invalid "+tagName+" reference format "+value,
					lines[0].LineNumber,
					context)
			}
		}
	}
}

// validateMultipleEventsGeneric checks for multiple occurrences of an event tag.
// This reduces duplication in event validation.
func validateMultipleEventsGeneric(
	xrefID, eventTag string,
	events []*gedcom.GedcomLine,
	recordTypePrefix, context string,
	errorManager *gedcom.ErrorManager,
) {
	if len(events) > 1 {
		errorManager.AddError(gedcom.SeverityWarning,
			recordTypePrefix+" "+xrefID+": Multiple "+eventTag+" events",
			events[1].LineNumber,
			context)
	}
}

// validateEventsGeneric validates event structures using generics.
// This reduces duplication between IndividualValidator and FamilyValidator.
func validateEventsGeneric(
	xrefID string,
	record gedcom.Record,
	eventTags map[string]bool,
	validEventTags map[string]bool,
	recordTypePrefix, context string,
	errorManager *gedcom.ErrorManager,
) {
	// Validate each event structure
	for eventTag := range eventTags {
		eventLines := record.GetLines(eventTag)
		for _, eventLine := range eventLines {
			validateEventStructureGeneric(
				xrefID, eventTag, eventLine,
				validEventTags,
				recordTypePrefix, context,
				errorManager,
			)
		}
	}
}

// validateTagValueGeneric validates a tag value against a set of valid values.
// This reduces duplication in value validation (e.g., SEX values).
func validateTagValueGeneric(
	xrefID, tagName, value string,
	validValues map[string]bool,
	record gedcom.Record,
	recordTypePrefix, context string,
	errorManager *gedcom.ErrorManager,
) {
	if value != "" && !validValues[value] {
		lines := record.GetLines(tagName)
		if len(lines) > 0 {
			errorManager.AddError(gedcom.SeveritySevere,
				recordTypePrefix+" "+xrefID+": Invalid "+tagName+" value "+value,
				lines[0].LineNumber,
				context)
		}
	}
}

// validateSubtagStructureGeneric validates the structure of subtags under a parent tag.
// This reduces duplication in name structure validation and similar patterns.
func validateSubtagStructureGeneric(
	xrefID, parentTag string,
	parentLine *gedcom.GedcomLine,
	validSubtags map[string]bool,
	recordTypePrefix, context string,
	errorManager *gedcom.ErrorManager,
) {
	for _, lines := range parentLine.Children {
		for _, line := range lines {
			if !validSubtags[line.Tag] && !isUserDefinedTag(line.Tag) {
				errorManager.AddError(gedcom.SeverityWarning,
					recordTypePrefix+" "+xrefID+": Invalid subtag "+line.Tag+" in "+parentTag+" structure",
					line.LineNumber,
					context)
			}
		}
	}
}
