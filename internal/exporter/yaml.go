package exporter

import (
	"fmt"
	"os"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"gopkg.in/yaml.v3"
)

// YAMLExporter exports a GEDCOM tree to YAML format.
type YAMLExporter struct {
	*BaseExporter
}

// NewYAMLExporter creates a new YAMLExporter.
func NewYAMLExporter(errorManager *gedcom.ErrorManager) *YAMLExporter {
	return &YAMLExporter{
		BaseExporter: NewBaseExporter(errorManager),
	}
}

// ExportToFile exports the tree to a YAML file.
func (ye *YAMLExporter) ExportToFile(tree *gedcom.GedcomTree, filePath string) error {
	yamlData, err := ye.createYAMLStructure(tree)
	if err != nil {
		ye.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Error creating YAML structure: %s", err.Error()),
			0,
			"YAML Export")
		return fmt.Errorf("failed to create YAML structure: %w", err)
	}

	data, err := yaml.Marshal(yamlData)
	if err != nil {
		ye.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Error marshaling YAML: %s", err.Error()),
			0,
			"YAML Export")
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := ye.writeToFile(filePath, string(data)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportToString exports the tree to a YAML string.
func (ye *YAMLExporter) ExportToString(tree *gedcom.GedcomTree) (string, error) {
	yamlData, err := ye.createYAMLStructure(tree)
	if err != nil {
		return "", fmt.Errorf("failed to create YAML structure: %w", err)
	}

	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(data), nil
}

// YAMLGedcom represents the root YAML structure.
type YAMLGedcom struct {
	Version     string                    `yaml:"version"`
	Header      map[string]interface{}    `yaml:"header,omitempty"`
	Submitters  []map[string]interface{}  `yaml:"submitters,omitempty"`
	Individuals map[string]interface{}   `yaml:"individuals,omitempty"`
	Families    map[string]interface{}    `yaml:"families,omitempty"`
	Sources     map[string]interface{}    `yaml:"sources,omitempty"`
	Repositories map[string]interface{}   `yaml:"repositories,omitempty"`
	Notes       map[string]interface{}    `yaml:"notes,omitempty"`
	Multimedia  map[string]interface{}     `yaml:"multimedia,omitempty"`
	Metadata    map[string]interface{}    `yaml:"metadata,omitempty"`
}

// createYAMLStructure creates the YAML structure from the tree.
func (ye *YAMLExporter) createYAMLStructure(tree *gedcom.GedcomTree) (*YAMLGedcom, error) {
	// Reuse JSON exporter to get the structure, then convert to YAML format
	jsonExporter := NewJsonExporter(ye.errorManager)
	jsonData, err := jsonExporter.createJSONStructure(tree)
	if err != nil {
		return nil, err
	}

	yamlGedcom := &YAMLGedcom{
		Version: "5.5.5",
	}

	// Convert JSON structure to YAML-compatible structure
	if header, ok := jsonData["header"].(map[string]interface{}); ok {
		yamlGedcom.Header = header
	}

	if submitters, ok := jsonData["submitter"].(map[string]interface{}); ok && submitters != nil {
		yamlGedcom.Submitters = []map[string]interface{}{submitters}
	}

	if individuals, ok := jsonData["individuals"].(map[string]interface{}); ok {
		yamlGedcom.Individuals = individuals
	}

	if families, ok := jsonData["families"].(map[string]interface{}); ok {
		yamlGedcom.Families = families
	}

	if sources, ok := jsonData["sources"].(map[string]interface{}); ok {
		yamlGedcom.Sources = sources
	}

	if repositories, ok := jsonData["repositories"].(map[string]interface{}); ok {
		yamlGedcom.Repositories = repositories
	}

	if notes, ok := jsonData["notes"].(map[string]interface{}); ok {
		yamlGedcom.Notes = notes
	}

	if multimedia, ok := jsonData["multimedia"].(map[string]interface{}); ok {
		yamlGedcom.Multimedia = multimedia
	}

	if metadata, ok := jsonData["metadata"].(map[string]interface{}); ok {
		yamlGedcom.Metadata = metadata
	}

	return yamlGedcom, nil
}

// writeToFile writes content to a file with error handling.
func (ye *YAMLExporter) writeToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		ye.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Failed to create file: %s", err.Error()),
			0,
			"YAML Export")
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		ye.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Failed to write file: %s", err.Error()),
			0,
			"YAML Export")
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

