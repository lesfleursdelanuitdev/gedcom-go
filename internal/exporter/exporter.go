package exporter

import (
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// Exporter is the interface that all exporters must implement.
type Exporter interface {
	// ExportToFile exports the tree to a file.
	ExportToFile(tree *gedcom.GedcomTree, filePath string) error

	// ExportToString exports the tree to a string.
	ExportToString(tree *gedcom.GedcomTree) (string, error)
}

// BaseExporter provides common functionality for all exporters.
type BaseExporter struct {
	errorManager *gedcom.ErrorManager
}

// NewBaseExporter creates a new BaseExporter.
func NewBaseExporter(errorManager *gedcom.ErrorManager) *BaseExporter {
	return &BaseExporter{
		errorManager: errorManager,
	}
}

// AddError is a helper method to add errors.
func (be *BaseExporter) AddError(severity gedcom.ErrorSeverity, message string, lineNumber int, context string) {
	be.errorManager.AddError(severity, message, lineNumber, context)
}


