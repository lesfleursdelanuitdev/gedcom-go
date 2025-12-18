package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

const (
	// MaxLineLength is the maximum length for a GEDCOM line (255 characters per spec)
	MaxLineLength = 255
	// PreferredLineLength is the preferred length before splitting with CONC/CONT
	PreferredLineLength = 240
)

// GedcomExporter exports a GEDCOM tree to GEDCOM file format.
type GedcomExporter struct {
	*BaseExporter
	appName    string
	appVersion string
}

// NewGedcomExporter creates a new GedcomExporter.
func NewGedcomExporter(errorManager *gedcom.ErrorManager, appName, appVersion string) *GedcomExporter {
	return &GedcomExporter{
		BaseExporter: NewBaseExporter(errorManager),
		appName:      appName,
		appVersion:   appVersion,
	}
}

// ExportToFile exports the tree to a GEDCOM file.
func (ge *GedcomExporter) ExportToFile(tree *gedcom.GedcomTree, filePath string) error {
	// Update header with metadata
	if err := ge.updateHeader(tree, filePath); err != nil {
		return fmt.Errorf("failed to update header: %w", err)
	}

	// Ensure submitter exists
	if err := ge.ensureSubmitter(tree); err != nil {
		return fmt.Errorf("failed to ensure submitter: %w", err)
	}

	// Generate GEDCOM content
	content, err := ge.ExportToString(tree)
	if err != nil {
		return fmt.Errorf("failed to generate GEDCOM content: %w", err)
	}

	// Write to file
	if err := ge.writeToFile(filePath, content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportToString exports the tree to a GEDCOM format string.
func (ge *GedcomExporter) ExportToString(tree *gedcom.GedcomTree) (string, error) {
	var lines []string

	// Add header
	header := tree.GetHeader()
	if header != nil {
		headerLines := ge.lineToGED(header.FirstLine())
		lines = append(lines, headerLines...)
	}

	// Add submitter (should be right after header)
	submitters := tree.GetAllSubmitters()
	if len(submitters) > 0 {
		// Get first submitter
		for _, subm := range submitters {
			submLines := ge.lineToGED(subm.FirstLine())
			lines = append(lines, submLines...)
			break // Only add first submitter
		}
	}

	// Add all other records in order
	// INDI, FAM, SOUR, REPO, NOTE, OBJE
	individuals := tree.GetAllIndividuals()
	for _, indi := range individuals {
		indiLines := ge.lineToGED(indi.FirstLine())
		lines = append(lines, indiLines...)
	}

	families := tree.GetAllFamilies()
	for _, fam := range families {
		famLines := ge.lineToGED(fam.FirstLine())
		lines = append(lines, famLines...)
	}

	sources := tree.GetAllSources()
	for _, src := range sources {
		srcLines := ge.lineToGED(src.FirstLine())
		lines = append(lines, srcLines...)
	}

	repositories := tree.GetAllRepositories()
	for _, repo := range repositories {
		repoLines := ge.lineToGED(repo.FirstLine())
		lines = append(lines, repoLines...)
	}

	notes := tree.GetAllNotes()
	for _, note := range notes {
		noteLines := ge.lineToGED(note.FirstLine())
		lines = append(lines, noteLines...)
	}

	multimedia := tree.GetAllMultimedia()
	for _, obje := range multimedia {
		objeLines := ge.lineToGED(obje.FirstLine())
		lines = append(lines, objeLines...)
	}

	// Add trailer
	lines = append(lines, "0 TRLR")

	return strings.Join(lines, "\n") + "\n", nil
}

// lineToGED converts a GedcomLine to GEDCOM format, handling CONC/CONT for long lines.
func (ge *GedcomExporter) lineToGED(line *gedcom.GedcomLine) []string {
	lines := []string{}
	
	// Convert the line itself
	lineStr := ge.formatGEDLine(line)
	
	// Handle long lines with CONC/CONT
	if len(lineStr) > MaxLineLength {
		lines = append(lines, ge.splitLongLine(line)...)
	} else {
		lines = append(lines, lineStr)
	}
	
	// Process children
	for _, children := range line.Children {
		for _, child := range children {
			lines = append(lines, ge.lineToGED(child)...)
		}
	}
	
	return lines
}

// formatGEDLine formats a single line to GEDCOM format.
func (ge *GedcomExporter) formatGEDLine(line *gedcom.GedcomLine) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("%d", line.Level))
	
	if line.XrefID != "" {
		parts = append(parts, line.XrefID)
	}
	
	parts = append(parts, line.Tag)
	
	if line.Value != "" {
		parts = append(parts, line.Value)
	}
	
	return strings.Join(parts, " ")
}

// splitLongLine splits a long line using CONC/CONT continuation lines.
func (ge *GedcomExporter) splitLongLine(line *gedcom.GedcomLine) []string {
	lines := []string{}
	
	// Format the base line
	var parts []string
	parts = append(parts, fmt.Sprintf("%d", line.Level))
	if line.XrefID != "" {
		parts = append(parts, line.XrefID)
	}
	parts = append(parts, line.Tag)
	
	baseLine := strings.Join(parts, " ")
	value := line.Value
	
	// Calculate how much value we can fit on the first line
	firstLineMaxLen := MaxLineLength - len(baseLine) - 1 // -1 for space
	if firstLineMaxLen < 0 {
		firstLineMaxLen = 0
	}
	
	// Split value into chunks
	if len(value) <= firstLineMaxLen {
		// Fits on one line
		lines = append(lines, baseLine+" "+value)
		return lines
	}
	
	// First line with partial value
	firstValue := value[:firstLineMaxLen]
	lines = append(lines, baseLine+" "+firstValue)
	value = value[firstLineMaxLen:]
	
	// Remaining lines with CONC (no newline) or CONT (with newline)
	// Use CONC for continuation on same line, CONT for new line
	// For simplicity, we'll use CONC for all continuations
	level := line.Level + 1
	concTag := "CONC"
	
	for len(value) > 0 {
		remaining := MaxLineLength - len(fmt.Sprintf("%d %s ", level, concTag))
		if remaining < 0 {
			remaining = 0
		}
		
		if len(value) <= remaining {
			// Last chunk
			lines = append(lines, fmt.Sprintf("%d %s %s", level, concTag, value))
			break
		}
		
		// Split at word boundary if possible
		chunk := value[:remaining]
		if remaining < len(value) {
			// Try to split at space
			lastSpace := strings.LastIndex(chunk, " ")
			if lastSpace > remaining/2 {
				chunk = value[:lastSpace]
				value = value[lastSpace+1:]
			} else {
				value = value[remaining:]
			}
		} else {
			value = ""
		}
		
		lines = append(lines, fmt.Sprintf("%d %s %s", level, concTag, chunk))
	}
	
	return lines
}

// updateHeader updates the header with metadata.
func (ge *GedcomExporter) updateHeader(tree *gedcom.GedcomTree, filePath string) error {
	header := tree.GetHeader()
	if header == nil {
		return fmt.Errorf("GEDCOM structure is missing a header record")
	}

	firstLine := header.FirstLine()
	now := time.Now()

	// Update GEDC.VERS (GEDCOM version)
	if firstLine.GetValue("GEDC.VERS") == "" {
		firstLine.SetValue("GEDC.VERS", "5.5.5")
	}

	// Ensure GEDC structure exists
	if len(firstLine.GetLines("GEDC")) == 0 {
		gedcLine := gedcom.NewGedcomLine(1, "GEDC", "", "")
		versLine := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")
		gedcLine.AddChild(versLine)
		firstLine.AddChild(gedcLine)
	}

	// Update CHAR (character encoding)
	if firstLine.GetValue("CHAR") == "" {
		firstLine.SetValue("CHAR", "UTF-8")
	}

	// Update SOUR (source system)
	if firstLine.GetValue("SOUR") == "" {
		firstLine.SetValue("SOUR", ge.appName)
	}
	firstLine.SetValue("SOUR.VERS", ge.appVersion)

	// Update DATE (export date)
	dateStr := now.Format("02 Jan 2006")
	firstLine.SetValue("DATE", dateStr)

	// Update TIME (export time)
	timeStr := now.Format("15:04:05")
	firstLine.SetValue("TIME", timeStr)

	// Update FILE (file name)
	// Extract just the filename from the path
	fileName := filepath.Base(filePath)
	firstLine.SetValue("FILE", fileName)

	return nil
}

// ensureSubmitter ensures a submitter record exists.
func (ge *GedcomExporter) ensureSubmitter(tree *gedcom.GedcomTree) error {
	submitters := tree.GetAllSubmitters()
	if len(submitters) == 0 {
		return fmt.Errorf("GEDCOM structure is missing a submitter record")
	}
	return nil
}

// writeToFile writes content to a file with error handling.
func (ge *GedcomExporter) writeToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		ge.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Failed to create file: %s", err.Error()),
			0,
			"GEDCOM Export")
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		ge.AddError(gedcom.SeveritySevere,
			fmt.Sprintf("Failed to write file: %s", err.Error()),
			0,
			"GEDCOM Export")
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

