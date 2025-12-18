package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// JsonExporter exports a GEDCOM tree to JSON format.
type JsonExporter struct {
	*BaseExporter
}

// NewJsonExporter creates a new JsonExporter.
func NewJsonExporter(errorManager *gedcom.ErrorManager) *JsonExporter {
	return &JsonExporter{
		BaseExporter: NewBaseExporter(errorManager),
	}
}

// ExportToFile exports the tree to a JSON file.
func (je *JsonExporter) ExportToFile(tree *gedcom.GedcomTree, filePath string) error {
	jsonData, err := je.createJSONStructure(tree)
	if err != nil {
		je.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Error creating JSON structure: %s", err.Error()),
			0,
			"JSON Export")
		return fmt.Errorf("failed to create JSON structure: %w", err)
	}

	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		je.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Error marshaling JSON: %s", err.Error()),
			0,
			"JSON Export")
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := je.writeToFile(filePath, string(data)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportToString exports the tree to a JSON string.
func (je *JsonExporter) ExportToString(tree *gedcom.GedcomTree) (string, error) {
	jsonData, err := je.createJSONStructure(tree)
	if err != nil {
		return "", fmt.Errorf("failed to create JSON structure: %w", err)
	}

	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// createJSONStructure creates the JSON structure from the tree.
func (je *JsonExporter) createJSONStructure(tree *gedcom.GedcomTree) (map[string]interface{}, error) {
	return map[string]interface{}{
		"header":      je.headerToJSON(tree),
		"submitter":   je.submitterToJSON(tree),
		"individuals": je.individualsToJSON(tree),
		"families":    je.familiesToJSON(tree),
		"sources":     je.sourcesToJSON(tree),
		"repositories": je.repositoriesToJSON(tree),
		"multimedia":  je.multimediaToJSON(tree),
		"notes":       je.notesToJSON(tree),
		"metadata":    je.metadataToJSON(tree),
	}, nil
}

// metadataToJSON converts metadata to JSON.
func (je *JsonExporter) metadataToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	header := tree.GetHeader()
	form := ""
	if header != nil {
		form = header.GetValue("GEDC.FORM")
	}

	return map[string]interface{}{
		"encoding":   tree.GetEncoding(),
		"version":    tree.GetVersion(),
		"form":       form,
		"export_date": time.Now().Format(time.RFC3339),
	}
}

// headerToJSON converts header to JSON.
func (je *JsonExporter) headerToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	header := tree.GetHeader()
	if header == nil {
		return nil
	}

	return map[string]interface{}{
		"file": header.GetValue("FILE"),
		"source": map[string]interface{}{
			"system":      header.GetValue("SOUR"),
			"version":     header.GetValue("SOUR.VERS"),
			"name":        header.GetValue("SOUR.NAME"),
			"corporation": header.GetValue("SOUR.CORP"),
		},
		"destination": header.GetValue("DEST"),
		"date":        header.GetValue("DATE"),
		"submitter":   header.GetValue("SUBM"),
		"language":    header.GetValue("LANG"),
	}
}

// submitterToJSON converts submitter to JSON.
func (je *JsonExporter) submitterToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	submitters := tree.GetAllSubmitters()
	if len(submitters) == 0 {
		return nil
	}

	// Get first submitter
	var subm gedcom.Record
	for _, s := range submitters {
		subm = s
		break
	}

	return map[string]interface{}{
		"id":      subm.XrefID(),
		"name":    subm.GetValue("NAME"),
		"address": je.formatAddress(subm),
		"phone":   subm.GetValue("PHON"),
		"email":   subm.GetValue("EMAIL"),
	}
}

// individualsToJSON converts all individuals to JSON.
func (je *JsonExporter) individualsToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	individuals := tree.GetAllIndividuals()
	result := make(map[string]interface{})
	for xref, indi := range individuals {
		result[xref] = je.individualToJSON(indi)
	}
	return result
}

// individualToJSON converts an individual to JSON.
func (je *JsonExporter) individualToJSON(individual gedcom.Record) map[string]interface{} {
	return map[string]interface{}{
		"id":         individual.XrefID(),
		"names":      je.getNames(individual),
		"sex":        individual.GetValue("SEX"),
		"birth":      je.getEvent(individual, "BIRT"),
		"death":      je.getEvent(individual, "DEAT"),
		"events":     je.getEvents(individual),
		"attributes": je.getAttributes(individual),
		"families": map[string]interface{}{
			"spouse": individual.GetValues("FAMS"),
			"child":  individual.GetValues("FAMC"),
		},
		"notes": individual.GetValues("NOTE"),
	}
}

// familiesToJSON converts all families to JSON.
func (je *JsonExporter) familiesToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	families := tree.GetAllFamilies()
	result := make(map[string]interface{})
	for xref, fam := range families {
		result[xref] = je.familyToJSON(fam)
	}
	return result
}

// familyToJSON converts a family to JSON.
func (je *JsonExporter) familyToJSON(family gedcom.Record) map[string]interface{} {
	return map[string]interface{}{
		"id":       family.XrefID(),
		"husband":  family.GetValue("HUSB"),
		"wife":     family.GetValue("WIFE"),
		"children": family.GetValues("CHIL"),
		"events":   je.getEvents(family),
		"notes":    family.GetValues("NOTE"),
	}
}

// sourcesToJSON converts all sources to JSON.
func (je *JsonExporter) sourcesToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	sources := tree.GetAllSources()
	result := make(map[string]interface{})
	for xref, src := range sources {
		result[xref] = je.sourceToJSON(src)
	}
	return result
}

// sourceToJSON converts a source to JSON.
func (je *JsonExporter) sourceToJSON(source gedcom.Record) map[string]interface{} {
	return map[string]interface{}{
		"id":          source.XrefID(),
		"title":       source.GetValue("TITL"),
		"author":      source.GetValue("AUTH"),
		"publication": source.GetValue("PUBL"),
		"repository":  source.GetValue("REPO"),
		"notes":       source.GetValues("NOTE"),
	}
}

// repositoriesToJSON converts all repositories to JSON.
func (je *JsonExporter) repositoriesToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	repositories := tree.GetAllRepositories()
	result := make(map[string]interface{})
	for xref, repo := range repositories {
		result[xref] = je.repositoryToJSON(repo)
	}
	return result
}

// repositoryToJSON converts a repository to JSON.
func (je *JsonExporter) repositoryToJSON(repository gedcom.Record) map[string]interface{} {
	return map[string]interface{}{
		"id":      repository.XrefID(),
		"name":    repository.GetValue("NAME"),
		"address": je.formatAddress(repository),
		"notes":   repository.GetValues("NOTE"),
	}
}

// multimediaToJSON converts all multimedia to JSON.
func (je *JsonExporter) multimediaToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	multimedia := tree.GetAllMultimedia()
	result := make(map[string]interface{})
	for xref, obje := range multimedia {
		result[xref] = je.multimediaItemToJSON(obje)
	}
	return result
}

// multimediaItemToJSON converts a multimedia item to JSON.
func (je *JsonExporter) multimediaItemToJSON(multimedia gedcom.Record) map[string]interface{} {
	return map[string]interface{}{
		"id":     multimedia.XrefID(),
		"file":   multimedia.GetValue("FILE"),
		"format": multimedia.GetValue("FILE.FORM"),
		"title":  multimedia.GetValue("TITL"),
		"notes":  multimedia.GetValues("NOTE"),
	}
}

// notesToJSON converts all notes to JSON.
func (je *JsonExporter) notesToJSON(tree *gedcom.GedcomTree) map[string]interface{} {
	notes := tree.GetAllNotes()
	result := make(map[string]interface{})
	for xref, note := range notes {
		result[xref] = je.noteToJSON(note)
	}
	return result
}

// noteToJSON converts a note to JSON.
func (je *JsonExporter) noteToJSON(note gedcom.Record) map[string]interface{} {
	return map[string]interface{}{
		"id":   note.XrefID(),
		"text": note.GetValue(""),
	}
}

// getNames extracts names from an individual.
func (je *JsonExporter) getNames(individual gedcom.Record) []map[string]interface{} {
	nameLines := individual.GetLines("NAME")
	names := make([]map[string]interface{}, 0, len(nameLines))
	
	for _, nameLine := range nameLines {
		names = append(names, map[string]interface{}{
			"full":    nameLine.Value,
			"given":   nameLine.GetValue("GIVN"),
			"surname": nameLine.GetValue("SURN"),
			"prefix":  nameLine.GetValue("NPFX"),
			"suffix":  nameLine.GetValue("NSFX"),
		})
	}
	
	return names
}

// getEvent extracts an event from a record.
func (je *JsonExporter) getEvent(record gedcom.Record, eventTag string) map[string]interface{} {
	eventLines := record.GetLines(eventTag)
	if len(eventLines) == 0 {
		return nil
	}
	
	eventLine := eventLines[0]
	return map[string]interface{}{
		"date":  eventLine.GetValue("DATE"),
		"place": eventLine.GetValue("PLAC"),
		"notes": je.getNoteValues(eventLine),
	}
}

// getEvents extracts all events from a record.
func (je *JsonExporter) getEvents(record gedcom.Record) []map[string]interface{} {
	eventTags := []string{"BIRT", "DEAT", "MARR", "DIV", "EVEN"}
	events := make([]map[string]interface{}, 0)
	
	for _, tag := range eventTags {
		eventLines := record.GetLines(tag)
		for _, eventLine := range eventLines {
			events = append(events, map[string]interface{}{
				"type":  tag,
				"date":  eventLine.GetValue("DATE"),
				"place": eventLine.GetValue("PLAC"),
				"notes": je.getNoteValues(eventLine),
			})
		}
	}
	
	return events
}

// getAttributes extracts attributes from an individual.
func (je *JsonExporter) getAttributes(individual gedcom.Record) []map[string]interface{} {
	attrTags := []string{"OCCU", "EDUC", "RESI", "TITL", "FACT"}
	attributes := make([]map[string]interface{}, 0)
	
	for _, tag := range attrTags {
		attrLines := individual.GetLines(tag)
		for _, attrLine := range attrLines {
			attributes = append(attributes, map[string]interface{}{
				"type":  tag,
				"value": attrLine.Value,
				"date":  attrLine.GetValue("DATE"),
				"place": attrLine.GetValue("PLAC"),
				"notes": je.getNoteValues(attrLine),
			})
		}
	}
	
	return attributes
}

// formatAddress formats an address from a record.
func (je *JsonExporter) formatAddress(record gedcom.Record) map[string]interface{} {
	addrLines := record.GetLines("ADDR")
	if len(addrLines) == 0 {
		return nil
	}
	
	addrLine := addrLines[0]
	return map[string]interface{}{
		"lines": []interface{}{
			addrLine.GetValue("ADR1"),
			addrLine.GetValue("ADR2"),
			addrLine.GetValue("ADR3"),
		},
		"city":       addrLine.GetValue("CITY"),
		"state":      addrLine.GetValue("STAE"),
		"postal_code": addrLine.GetValue("POST"),
		"country":    addrLine.GetValue("CTRY"),
	}
}

// getNoteValues extracts NOTE values from a line.
func (je *JsonExporter) getNoteValues(line *gedcom.GedcomLine) []string {
	noteLines := line.GetLines("NOTE")
	notes := make([]string, 0, len(noteLines))
	for _, noteLine := range noteLines {
		if noteLine.Value != "" {
			notes = append(notes, noteLine.Value)
		}
	}
	return notes
}

// writeToFile writes content to a file with error handling.
func (je *JsonExporter) writeToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		je.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Permission denied writing to file: %s", filePath),
			0,
			"File I/O")
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		je.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("I/O error writing file: %s", err.Error()),
			0,
			"File I/O")
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}


