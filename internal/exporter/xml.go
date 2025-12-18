package exporter

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// XMLExporter exports a GEDCOM tree to XML format.
type XMLExporter struct {
	*BaseExporter
}

// NewXMLExporter creates a new XMLExporter.
func NewXMLExporter(errorManager *gedcom.ErrorManager) *XMLExporter {
	return &XMLExporter{
		BaseExporter: NewBaseExporter(errorManager),
	}
}

// ExportToFile exports the tree to an XML file.
func (xe *XMLExporter) ExportToFile(tree *gedcom.GedcomTree, filePath string) error {
	xmlData, err := xe.createXMLStructure(tree)
	if err != nil {
		xe.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Error creating XML structure: %s", err.Error()),
			0,
			"XML Export")
		return fmt.Errorf("failed to create XML structure: %w", err)
	}

	data, err := xml.MarshalIndent(xmlData, "", "  ")
	if err != nil {
		xe.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Error marshaling XML: %s", err.Error()),
			0,
			"XML Export")
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML header
	xmlContent := xml.Header + string(data)

	if err := xe.writeToFile(filePath, xmlContent); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportToString exports the tree to an XML string.
func (xe *XMLExporter) ExportToString(tree *gedcom.GedcomTree) (string, error) {
	xmlData, err := xe.createXMLStructure(tree)
	if err != nil {
		return "", fmt.Errorf("failed to create XML structure: %w", err)
	}

	data, err := xml.MarshalIndent(xmlData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}

	return xml.Header + string(data), nil
}

// XMLGedcom represents the root XML element.
type XMLGedcom struct {
	XMLName      xml.Name         `xml:"gedcom"`
	Version      string           `xml:"version,attr"`
	Header       *XMLHeader       `xml:"header,omitempty"`
	Submitters   []*XMLSubmitter  `xml:"submitters>submitter,omitempty"`
	Individuals  []*XMLIndividual `xml:"individuals>individual,omitempty"`
	Families     []*XMLFamily     `xml:"families>family,omitempty"`
	Sources      []*XMLSource     `xml:"sources>source,omitempty"`
	Repositories []*XMLRepository `xml:"repositories>repository,omitempty"`
	Notes        []*XMLNote       `xml:"notes>note,omitempty"`
	Multimedia   []*XMLMultimedia `xml:"multimedia>item,omitempty"`
	Metadata     *XMLMetadata     `xml:"metadata,omitempty"`
}

// XMLHeader represents header information.
type XMLHeader struct {
	File          string `xml:"file,omitempty"`
	Source        string `xml:"source,omitempty"`
	SourceVersion string `xml:"sourceVersion,omitempty"`
	CharacterSet  string `xml:"characterSet,omitempty"`
	Date          string `xml:"date,omitempty"`
	Language      string `xml:"language,omitempty"`
}

// XMLSubmitter represents a submitter.
type XMLSubmitter struct {
	ID      string      `xml:"id,attr"`
	Name    string      `xml:"name,omitempty"`
	Address *XMLAddress `xml:"address,omitempty"`
	Phone   string      `xml:"phone,omitempty"`
	Email   string      `xml:"email,omitempty"`
}

// XMLIndividual represents an individual.
type XMLIndividual struct {
	ID         string          `xml:"id,attr"`
	Names      []*XMLName      `xml:"names>name,omitempty"`
	Sex        string          `xml:"sex,omitempty"`
	Birth      *XMLEvent       `xml:"birth,omitempty"`
	Death      *XMLEvent       `xml:"death,omitempty"`
	Events     []*XMLEvent     `xml:"events>event,omitempty"`
	Attributes []*XMLAttribute `xml:"attributes>attribute,omitempty"`
	Families   *XMLFamilyRefs  `xml:"families,omitempty"`
	Notes      []string        `xml:"notes>note,omitempty"`
}

// XMLFamily represents a family.
type XMLFamily struct {
	ID       string      `xml:"id,attr"`
	Husband  string      `xml:"husband,omitempty"`
	Wife     string      `xml:"wife,omitempty"`
	Children []string    `xml:"children>child,omitempty"`
	Events   []*XMLEvent `xml:"events>event,omitempty"`
	Notes    []string    `xml:"notes>note,omitempty"`
}

// XMLSource represents a source.
type XMLSource struct {
	ID          string   `xml:"id,attr"`
	Title       string   `xml:"title,omitempty"`
	Author      string   `xml:"author,omitempty"`
	Publication string   `xml:"publication,omitempty"`
	Repository  string   `xml:"repository,omitempty"`
	Notes       []string `xml:"notes>note,omitempty"`
}

// XMLRepository represents a repository.
type XMLRepository struct {
	ID      string      `xml:"id,attr"`
	Name    string      `xml:"name,omitempty"`
	Address *XMLAddress `xml:"address,omitempty"`
	Notes   []string    `xml:"notes>note,omitempty"`
}

// XMLNote represents a note.
type XMLNote struct {
	ID   string `xml:"id,attr"`
	Text string `xml:",chardata"`
}

// XMLMultimedia represents a multimedia item.
type XMLMultimedia struct {
	ID     string   `xml:"id,attr"`
	File   string   `xml:"file,omitempty"`
	Format string   `xml:"format,omitempty"`
	Title  string   `xml:"title,omitempty"`
	Notes  []string `xml:"notes>note,omitempty"`
}

// XMLName represents a name.
type XMLName struct {
	Full    string `xml:"full,omitempty"`
	Given   string `xml:"given,omitempty"`
	Surname string `xml:"surname,omitempty"`
	Prefix  string `xml:"prefix,omitempty"`
	Suffix  string `xml:"suffix,omitempty"`
}

// XMLEvent represents an event.
type XMLEvent struct {
	Type  string   `xml:"type,attr"`
	Date  string   `xml:"date,omitempty"`
	Place string   `xml:"place,omitempty"`
	Notes []string `xml:"notes>note,omitempty"`
}

// XMLAttribute represents an attribute.
type XMLAttribute struct {
	Type  string   `xml:"type,attr"`
	Value string   `xml:"value,omitempty"`
	Date  string   `xml:"date,omitempty"`
	Place string   `xml:"place,omitempty"`
	Notes []string `xml:"notes>note,omitempty"`
}

// XMLAddress represents an address.
type XMLAddress struct {
	Lines      []string `xml:"lines>line,omitempty"`
	City       string   `xml:"city,omitempty"`
	State      string   `xml:"state,omitempty"`
	PostalCode string   `xml:"postalCode,omitempty"`
	Country    string   `xml:"country,omitempty"`
}

// XMLFamilyRefs represents family references.
type XMLFamilyRefs struct {
	Spouse []string `xml:"spouse,omitempty"`
	Child  []string `xml:"child,omitempty"`
}

// XMLMetadata represents metadata.
type XMLMetadata struct {
	Encoding   string `xml:"encoding,omitempty"`
	Version    string `xml:"version,omitempty"`
	Form       string `xml:"form,omitempty"`
	ExportDate string `xml:"exportDate,omitempty"`
}

// createXMLStructure creates the XML structure from the tree.
func (xe *XMLExporter) createXMLStructure(tree *gedcom.GedcomTree) (*XMLGedcom, error) {
	xmlGedcom := &XMLGedcom{
		Version: "5.5.5",
	}

	// Header
	header := tree.GetHeader()
	if header != nil {
		xmlGedcom.Header = &XMLHeader{
			File:          header.GetValue("FILE"),
			Source:        header.GetValue("SOUR"),
			SourceVersion: header.GetValue("SOUR.VERS"),
			CharacterSet:  header.GetValue("CHAR"),
			Date:          header.GetValue("DATE"),
			Language:      header.GetValue("LANG"),
		}
	}

	// Submitters
	submitters := tree.GetAllSubmitters()
	for _, subm := range submitters {
		xmlGedcom.Submitters = append(xmlGedcom.Submitters, &XMLSubmitter{
			ID:      subm.XrefID(),
			Name:    subm.GetValue("NAME"),
			Address: xe.formatAddressXML(subm),
			Phone:   subm.GetValue("PHON"),
			Email:   subm.GetValue("EMAIL"),
		})
	}

	// Individuals
	individuals := tree.GetAllIndividuals()
	for _, indi := range individuals {
		xmlGedcom.Individuals = append(xmlGedcom.Individuals, xe.individualToXML(indi))
	}

	// Families
	families := tree.GetAllFamilies()
	for _, fam := range families {
		xmlGedcom.Families = append(xmlGedcom.Families, xe.familyToXML(fam))
	}

	// Sources
	sources := tree.GetAllSources()
	for _, src := range sources {
		xmlGedcom.Sources = append(xmlGedcom.Sources, xe.sourceToXML(src))
	}

	// Repositories
	repositories := tree.GetAllRepositories()
	for _, repo := range repositories {
		xmlGedcom.Repositories = append(xmlGedcom.Repositories, xe.repositoryToXML(repo))
	}

	// Notes
	notes := tree.GetAllNotes()
	for _, note := range notes {
		xmlGedcom.Notes = append(xmlGedcom.Notes, &XMLNote{
			ID:   note.XrefID(),
			Text: note.GetValue(""),
		})
	}

	// Multimedia
	multimedia := tree.GetAllMultimedia()
	for _, obje := range multimedia {
		xmlGedcom.Multimedia = append(xmlGedcom.Multimedia, xe.multimediaToXML(obje))
	}

	// Metadata
	xmlGedcom.Metadata = &XMLMetadata{
		Encoding:   tree.GetEncoding(),
		Version:    tree.GetVersion(),
		Form:       header.GetValue("GEDC.FORM"),
		ExportDate: time.Now().Format(time.RFC3339),
	}

	return xmlGedcom, nil
}

// Helper methods (reuse JSON exporter logic)
func (xe *XMLExporter) individualToXML(individual gedcom.Record) *XMLIndividual {
	jsonExporter := NewJsonExporter(xe.errorManager)
	jsonData := jsonExporter.individualToJSON(individual)

	xmlIndi := &XMLIndividual{
		ID: individual.XrefID(),
	}

	if names, ok := jsonData["names"].([]map[string]interface{}); ok {
		for _, name := range names {
			xmlIndi.Names = append(xmlIndi.Names, &XMLName{
				Full:    getString(name, "full"),
				Given:   getString(name, "given"),
				Surname: getString(name, "surname"),
				Prefix:  getString(name, "prefix"),
				Suffix:  getString(name, "suffix"),
			})
		}
	}

	xmlIndi.Sex = getString(jsonData, "sex")

	if birth, ok := jsonData["birth"].(map[string]interface{}); ok {
		xmlIndi.Birth = &XMLEvent{
			Type:  "BIRT",
			Date:  getString(birth, "date"),
			Place: getString(birth, "place"),
		}
	}

	if death, ok := jsonData["death"].(map[string]interface{}); ok {
		xmlIndi.Death = &XMLEvent{
			Type:  "DEAT",
			Date:  getString(death, "date"),
			Place: getString(death, "place"),
		}
	}

	if events, ok := jsonData["events"].([]map[string]interface{}); ok {
		for _, event := range events {
			xmlIndi.Events = append(xmlIndi.Events, &XMLEvent{
				Type:  getString(event, "type"),
				Date:  getString(event, "date"),
				Place: getString(event, "place"),
			})
		}
	}

	if families, ok := jsonData["families"].(map[string]interface{}); ok {
		xmlIndi.Families = &XMLFamilyRefs{}
		if spouse, ok := families["spouse"].([]string); ok {
			xmlIndi.Families.Spouse = spouse
		}
		if child, ok := families["child"].([]string); ok {
			xmlIndi.Families.Child = child
		}
	}

	return xmlIndi
}

func (xe *XMLExporter) familyToXML(family gedcom.Record) *XMLFamily {
	jsonExporter := NewJsonExporter(xe.errorManager)
	jsonData := jsonExporter.familyToJSON(family)

	return &XMLFamily{
		ID:       family.XrefID(),
		Husband:  getString(jsonData, "husband"),
		Wife:     getString(jsonData, "wife"),
		Children: getStringSlice(jsonData, "children"),
	}
}

func (xe *XMLExporter) sourceToXML(source gedcom.Record) *XMLSource {
	return &XMLSource{
		ID:          source.XrefID(),
		Title:       source.GetValue("TITL"),
		Author:      source.GetValue("AUTH"),
		Publication: source.GetValue("PUBL"),
		Repository:  source.GetValue("REPO"),
		Notes:       source.GetValues("NOTE"),
	}
}

func (xe *XMLExporter) repositoryToXML(repository gedcom.Record) *XMLRepository {
	return &XMLRepository{
		ID:      repository.XrefID(),
		Name:    repository.GetValue("NAME"),
		Address: xe.formatAddressXML(repository),
		Notes:   repository.GetValues("NOTE"),
	}
}

func (xe *XMLExporter) multimediaToXML(multimedia gedcom.Record) *XMLMultimedia {
	return &XMLMultimedia{
		ID:     multimedia.XrefID(),
		File:   multimedia.GetValue("FILE"),
		Format: multimedia.GetValue("FILE.FORM"),
		Title:  multimedia.GetValue("TITL"),
		Notes:  multimedia.GetValues("NOTE"),
	}
}

func (xe *XMLExporter) formatAddressXML(record gedcom.Record) *XMLAddress {
	addrLines := record.GetLines("ADDR")
	if len(addrLines) == 0 {
		return nil
	}

	addrLine := addrLines[0]
	lines := []string{}
	for _, line := range []string{"ADR1", "ADR2", "ADR3"} {
		if val := addrLine.GetValue(line); val != "" {
			lines = append(lines, val)
		}
	}

	return &XMLAddress{
		Lines:      lines,
		City:       addrLine.GetValue("CITY"),
		State:      addrLine.GetValue("STAE"),
		PostalCode: addrLine.GetValue("POST"),
		Country:    addrLine.GetValue("CTRY"),
	}
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return []string{}
}

// writeToFile writes content to a file with error handling.
func (xe *XMLExporter) writeToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		xe.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Failed to create file: %s", err.Error()),
			0,
			"XML Export")
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		xe.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Failed to write file: %s", err.Error()),
			0,
			"XML Export")
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
