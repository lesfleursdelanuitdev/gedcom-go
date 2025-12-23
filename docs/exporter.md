# GEDCOM Exporter Documentation

Complete reference guide for exporting GEDCOM data to various formats.

## Table of Contents

- [Overview](#overview)
- [Supported Formats](#supported-formats)
- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Export Formats](#export-formats)
  - [JSON Export](#json-export)
  - [XML Export](#xml-export)
  - [YAML Export](#yaml-export)
  - [GEDCOM Export](#gedcom-export)
- [API Reference](#api-reference)
- [Data Structure](#data-structure)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The exporter package provides functionality for converting GEDCOM trees to various formats, enabling integration with other systems and data transformation workflows. All exporters implement a common interface and support both file and string output.

### Features

- **Multiple Formats**: JSON, XML, YAML, and GEDCOM
- **Round-trip Conversion**: Convert between formats without data loss
- **Header Management**: Automatic header metadata updates
- **Continuation Handling**: Properly handles long lines with CONC/CONT
- **Pretty Printing**: Human-readable output for JSON, XML, YAML
- **Error Handling**: Comprehensive error reporting with severity levels

---

## Supported Formats

| Format | Exporter | File Extension | Pretty Print | Notes |
|--------|----------|---------------|--------------|-------|
| **JSON** | `JsonExporter` | `.json` | ✅ Yes | Most common for APIs |
| **XML** | `XMLExporter` | `.xml` | ✅ Yes | Standard XML format |
| **YAML** | `YAMLExporter` | `.yaml`, `.yml` | ✅ Yes | Human-readable |
| **CSV** | `CSVExporter` | `.csv` | N/A | Tabular format for spreadsheet import |
| **GEDCOM** | `GedcomExporter` | `.ged` | N/A | Native format |

---

## Installation

The exporter package is part of the GEDCOM Go library:

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/internal/exporter"
```

---

## Basic Usage

### Export to File

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/exporter"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func main() {
    // Parse GEDCOM file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Create error manager
    errorManager := gedcom.NewErrorManager()

    // Export to JSON
    jsonExporter := exporter.NewJsonExporter(errorManager)
    err = jsonExporter.ExportToFile(tree, "family.json")
    if err != nil {
        panic(err)
    }

    fmt.Println("Exported successfully!")
}
```

### Export to String

```go
// Export to JSON string
jsonExporter := exporter.NewJsonExporter(errorManager)
jsonString, err := jsonExporter.ExportToString(tree)
if err != nil {
    panic(err)
}

fmt.Println(jsonString)
```

---

## Export Formats

### JSON Export

JSON export provides a structured representation of GEDCOM data, ideal for APIs and programmatic access.

#### Features

- Pretty-printed by default (2-space indentation)
- Hierarchical structure matching GEDCOM organization
- All record types supported
- Metadata included (encoding, version, export date)

#### Example Output Structure

```json
{
  "header": {
    "source": "MyApp",
    "sourceVersion": "1.0",
    "characterSet": "UTF-8",
    "date": "2025-01-15",
    "language": "English"
  },
  "submitter": {
    "id": "@SUBM1@",
    "name": "John Doe",
    "address": {
      "street": "123 Main St",
      "city": "Anytown",
      "state": "ST",
      "postalCode": "12345",
      "country": "USA"
    }
  },
  "individuals": {
    "@I1@": {
      "id": "@I1@",
      "names": [
        {
          "full": "John /Doe/",
          "given": "John",
          "surname": "Doe",
          "prefix": "",
          "suffix": ""
        }
      ],
      "sex": "M",
      "birth": {
        "date": "1900-01-15",
        "place": "Anytown, ST, USA"
      },
      "death": {
        "date": "1970-03-20",
        "place": "Anytown, ST, USA"
      },
      "familiesAsChild": ["@F1@"],
      "familiesAsSpouse": ["@F2@"]
    }
  },
  "families": {
    "@F1@": {
      "id": "@F1@",
      "husband": "@I1@",
      "wife": "@I2@",
      "children": ["@I3@", "@I4@"],
      "marriage": {
        "date": "1920-06-10",
        "place": "Anytown, ST, USA"
      }
    }
  },
  "sources": {},
  "repositories": {},
  "multimedia": {},
  "notes": {},
  "metadata": {
    "encoding": "UTF-8",
    "version": "5.5.1",
    "form": "LINEAGE-LINKED",
    "export_date": "2025-01-15T10:30:00Z"
  }
}
```

#### Usage

```go
errorManager := gedcom.NewErrorManager()
jsonExporter := exporter.NewJsonExporter(errorManager)

// Export to file
err := jsonExporter.ExportToFile(tree, "output.json")

// Export to string
jsonString, err := jsonExporter.ExportToString(tree)
```

---

### XML Export

XML export provides a standard XML representation of GEDCOM data, suitable for integration with XML-based systems.

#### Features

- Standard XML format with proper namespaces
- Pretty-printed with 2-space indentation
- XML header included
- All record types supported

#### Example Output Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<gedcom version="5.5.5">
  <header>
    <source>MyApp</source>
    <sourceVersion>1.0</sourceVersion>
    <characterSet>UTF-8</characterSet>
    <date>2025-01-15</date>
    <language>English</language>
  </header>
  <submitters>
    <submitter id="@SUBM1@">
      <name>John Doe</name>
      <address>
        <street>123 Main St</street>
        <city>Anytown</city>
        <state>ST</state>
        <postalCode>12345</postalCode>
        <country>USA</country>
      </address>
    </submitter>
  </submitters>
  <individuals>
    <individual id="@I1@">
      <names>
        <name>
          <full>John /Doe/</full>
          <given>John</given>
          <surname>Doe</surname>
        </name>
      </names>
      <sex>M</sex>
      <birth>
        <date>1900-01-15</date>
        <place>Anytown, ST, USA</place>
      </birth>
      <death>
        <date>1970-03-20</date>
        <place>Anytown, ST, USA</place>
      </death>
      <familiesAsChild>@F1@</familiesAsChild>
      <familiesAsSpouse>@F2@</familiesAsSpouse>
    </individual>
  </individuals>
  <families>
    <family id="@F1@">
      <husband>@I1@</husband>
      <wife>@I2@</wife>
      <children>@I3@</children>
      <children>@I4@</children>
      <marriage>
        <date>1920-06-10</date>
        <place>Anytown, ST, USA</place>
      </marriage>
    </family>
  </families>
  <metadata>
    <encoding>UTF-8</encoding>
    <version>5.5.1</version>
    <form>LINEAGE-LINKED</form>
    <exportDate>2025-01-15T10:30:00Z</exportDate>
  </metadata>
</gedcom>
```

#### Usage

```go
errorManager := gedcom.NewErrorManager()
xmlExporter := exporter.NewXMLExporter(errorManager)

// Export to file
err := xmlExporter.ExportToFile(tree, "output.xml")

// Export to string
xmlString, err := xmlExporter.ExportToString(tree)
```

---

### YAML Export

YAML export provides a human-readable representation of GEDCOM data, ideal for configuration files and manual editing.

#### Features

- Human-readable format
- Preserves hierarchical structure
- All record types supported
- Metadata included

#### Example Output Structure

```yaml
version: "5.5.5"
header:
  source: MyApp
  sourceVersion: "1.0"
  characterSet: UTF-8
  date: "2025-01-15"
  language: English
submitters:
  - id: "@SUBM1@"
    name: John Doe
    address:
      street: "123 Main St"
      city: Anytown
      state: ST
      postalCode: "12345"
      country: USA
individuals:
  "@I1@":
    id: "@I1@"
    names:
      - full: "John /Doe/"
        given: John
        surname: Doe
    sex: M
    birth:
      date: "1900-01-15"
      place: "Anytown, ST, USA"
    death:
      date: "1970-03-20"
      place: "Anytown, ST, USA"
    familiesAsChild:
      - "@F1@"
    familiesAsSpouse:
      - "@F2@"
families:
  "@F1@":
    id: "@F1@"
    husband: "@I1@"
    wife: "@I2@"
    children:
      - "@I3@"
      - "@I4@"
    marriage:
      date: "1920-06-10"
      place: "Anytown, ST, USA"
metadata:
  encoding: UTF-8
  version: "5.5.1"
  form: LINEAGE-LINKED
  export_date: "2025-01-15T10:30:00Z"
```

#### Usage

```go
errorManager := gedcom.NewErrorManager()
yamlExporter := exporter.NewYAMLExporter(errorManager)

// Export to file
err := yamlExporter.ExportToFile(tree, "output.yaml")

// Export to string
yamlString, err := yamlExporter.ExportToString(tree)
```

---

### CSV Export

CSV export provides a tabular representation of GEDCOM data, ideal for spreadsheet applications and data analysis.

#### Features

- Tabular format with one row per individual
- All key fields included (name, dates, places, relationships)
- Easy import into Excel, Google Sheets, or database systems
- Standard CSV format with proper escaping

#### CSV Columns

The CSV export includes the following columns:

- **XREF**: Record identifier (e.g., "@I1@")
- **Type**: Record type (always "INDI" for individuals)
- **Name**: Full name
- **Sex**: Gender (M, F, U)
- **Birth Date**: Birth date
- **Birth Place**: Birth place
- **Death Date**: Death date
- **Death Place**: Death place
- **Father XREF**: Father's record identifier
- **Mother XREF**: Mother's record identifier
- **Spouse XREFs**: Semicolon-separated list of spouse family XREFs
- **Children XREFs**: Semicolon-separated list of children XREFs
- **Notes**: Pipe-separated list of note XREFs

#### Example Output

```csv
XREF,Type,Name,Sex,Birth Date,Birth Place,Death Date,Death Place,Father XREF,Mother XREF,Spouse XREFs,Children XREFs,Notes
@I1@,INDI,John /Doe/,M,1900-01-15,Anytown, ST,1970-03-20,Anytown, ST,@I10@,@I11@,@F1@;@F2@,@I3@;@I4@,@N1@
@I2@,INDI,Mary /Smith/,F,1902-05-20,Anytown, ST,1975-08-10,Anytown, ST,@I12@,@I13@,@F1@,@I3@;@I4@,
```

#### Usage

```go
errorManager := gedcom.NewErrorManager()
csvExporter := exporter.NewCSVExporter(errorManager)

// Export to file
err := csvExporter.ExportToFile(tree, "output.csv")

// Export to string
csvString, err := csvExporter.ExportToString(tree)
```

#### Use Cases

- **Data Analysis**: Import into Excel or Google Sheets for analysis
- **Database Import**: Import into relational databases
- **Reporting**: Generate reports in spreadsheet format
- **Data Migration**: Convert GEDCOM data for other systems

---

### GEDCOM Export

GEDCOM export re-exports data back to GEDCOM format, useful for normalization, cleaning, and format conversion.

#### Features

- GEDCOM 5.5.1 compliant output
- Automatic header updates (source, date, time)
- Line continuation handling (CONC/CONT for long lines)
- Proper record ordering
- Submitter creation if missing

#### Header Updates

When exporting to GEDCOM, the exporter automatically updates:

- `GEDC.VERS`: GEDCOM version (5.5.1)
- `CHAR`: Character encoding (UTF-8)
- `SOUR`: Source system (from constructor)
- `SOUR.VERS`: Source version (from constructor)
- `DATE`: Export date (current date)
- `TIME`: Export time (current time)

#### Line Continuation

The exporter automatically handles long lines (>255 characters) by splitting them using CONC (concatenation) and CONT (continuation) tags per GEDCOM specification.

#### Usage

```go
errorManager := gedcom.NewErrorManager()
gedcomExporter := exporter.NewGedcomExporter(
    errorManager,
    "MyApp",      // Application name
    "1.0.0",      // Application version
)

// Export to file
err := gedcomExporter.ExportToFile(tree, "output.ged")

// Export to string
gedcomString, err := gedcomExporter.ExportToString(tree)
```

#### Example Output

```
0 HEAD
1 GEDC
2 VERS 5.5.1
2 FORM LINEAGE-LINKED
1 CHAR UTF-8
1 SOUR MyApp
2 VERS 1.0.0
2 NAME MyApp
1 DATE 15 JAN 2025
2 TIME 10:30:00
1 SUBM @SUBM1@
0 @SUBM1@ SUBM
1 NAME John Doe
0 @I1@ INDI
1 NAME John /Doe/
2 GIVN John
2 SURN Doe
1 SEX M
1 BIRT
2 DATE 15 JAN 1900
2 PLAC Anytown, ST, USA
1 DEAT
2 DATE 20 MAR 1970
2 PLAC Anytown, ST, USA
1 FAMC @F1@
1 FAMS @F2@
0 @F1@ FAM
1 HUSB @I1@
1 WIFE @I2@
1 CHIL @I3@
1 CHIL @I4@
1 MARR
2 DATE 10 JUN 1920
2 PLAC Anytown, ST, USA
0 TRLR
```

---

## API Reference

### Exporter Interface

All exporters implement the `Exporter` interface:

```go
type Exporter interface {
    // ExportToFile exports the tree to a file.
    ExportToFile(tree *gedcom.GedcomTree, filePath string) error

    // ExportToString exports the tree to a string.
    ExportToString(tree *gedcom.GedcomTree) (string, error)
}
```

### Constructors

#### JsonExporter

```go
func NewJsonExporter(errorManager *gedcom.ErrorManager) *JsonExporter
```

#### XMLExporter

```go
func NewXMLExporter(errorManager *gedcom.ErrorManager) *XMLExporter
```

#### YAMLExporter

```go
func NewYAMLExporter(errorManager *gedcom.ErrorManager) *YAMLExporter
```

#### GedcomExporter

```go
func NewGedcomExporter(
    errorManager *gedcom.ErrorManager,
    appName string,
    appVersion string,
) *GedcomExporter
```

**Parameters:**
- `errorManager`: Error manager for collecting export errors
- `appName`: Application name (appears in GEDCOM header)
- `appVersion`: Application version (appears in GEDCOM header)

---

## Data Structure

### Exported Data Types

All exporters support the following GEDCOM record types:

| Record Type | JSON Key | XML Element | YAML Key | Description |
|------------|----------|-------------|----------|-------------|
| **Header** | `header` | `<header>` | `header` | File header information |
| **Submitter** | `submitter` | `<submitters><submitter>` | `submitters` | Submitter information |
| **Individual** | `individuals` | `<individuals><individual>` | `individuals` | Individual records |
| **Family** | `families` | `<families><family>` | `families` | Family records |
| **Source** | `sources` | `<sources><source>` | `sources` | Source citations |
| **Repository** | `repositories` | `<repositories><repository>` | `repositories` | Repository records |
| **Note** | `notes` | `<notes><note>` | `notes` | Note records |
| **Multimedia** | `multimedia` | `<multimedia><item>` | `multimedia` | Multimedia objects |
| **Metadata** | `metadata` | `<metadata>` | `metadata` | Export metadata |

### Individual Record Structure

```json
{
  "id": "@I1@",
  "names": [
    {
      "full": "John /Doe/",
      "given": "John",
      "surname": "Doe",
      "prefix": "",
      "suffix": ""
    }
  ],
  "sex": "M",
  "birth": {
    "date": "1900-01-15",
    "place": "Anytown, ST, USA"
  },
  "death": {
    "date": "1970-03-20",
    "place": "Anytown, ST, USA"
  },
  "familiesAsChild": ["@F1@"],
  "familiesAsSpouse": ["@F2@"]
}
```

### Family Record Structure

```json
{
  "id": "@F1@",
  "husband": "@I1@",
  "wife": "@I2@",
  "children": ["@I3@", "@I4@"],
  "marriage": {
    "date": "1920-06-10",
    "place": "Anytown, ST, USA"
  }
}
```

---

## Examples

### Complete Export Workflow

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/exporter"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func main() {
    // Parse GEDCOM file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    errorManager := gedcom.NewErrorManager()

    // Export to multiple formats
    formats := []struct {
        name     string
        exporter exporter.Exporter
        file     string
    }{
        {"JSON", exporter.NewJsonExporter(errorManager), "family.json"},
        {"XML", exporter.NewXMLExporter(errorManager), "family.xml"},
        {"YAML", exporter.NewYAMLExporter(errorManager), "family.yaml"},
        {"GEDCOM", exporter.NewGedcomExporter(errorManager, "MyApp", "1.0.0"), "family_exported.ged"},
    }

    for _, format := range formats {
        fmt.Printf("Exporting to %s...\n", format.name)
        if err := format.exporter.ExportToFile(tree, format.file); err != nil {
            fmt.Printf("Error exporting to %s: %v\n", format.name, err)
        } else {
            fmt.Printf("✓ Exported to %s\n", format.file)
        }
    }

    // Check for errors
    if errors := errorManager.Errors(); len(errors) > 0 {
        fmt.Printf("\nExport completed with %d errors:\n", len(errors))
        for _, err := range errors {
            fmt.Printf("  - %s\n", err.Message)
        }
    }
}
```

### Round-trip Conversion

```go
// GEDCOM → JSON → XML → YAML → GEDCOM
func roundTripConversion(inputFile string) error {
    // Parse
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse(inputFile)
    if err != nil {
        return err
    }

    errorManager := gedcom.NewErrorManager()

    // GEDCOM → JSON
    jsonExporter := exporter.NewJsonExporter(errorManager)
    jsonString, err := jsonExporter.ExportToString(tree)
    if err != nil {
        return err
    }
    fmt.Println("Converted to JSON")

    // JSON → Parse back (would need JSON parser)
    // For now, continue with same tree

    // Export to XML
    xmlExporter := exporter.NewXMLExporter(errorManager)
    err = xmlExporter.ExportToFile(tree, "output.xml")
    if err != nil {
        return err
    }
    fmt.Println("Converted to XML")

    // Export to YAML
    yamlExporter := exporter.NewYAMLExporter(errorManager)
    err = yamlExporter.ExportToFile(tree, "output.yaml")
    if err != nil {
        return err
    }
    fmt.Println("Converted to YAML")

    // Export back to GEDCOM
    gedcomExporter := exporter.NewGedcomExporter(errorManager, "MyApp", "1.0.0")
    err = gedcomExporter.ExportToFile(tree, "output.ged")
    if err != nil {
        return err
    }
    fmt.Println("Converted back to GEDCOM")

    return nil
}
```

### Batch Export

```go
func batchExport(inputFiles []string, outputFormat string) error {
    errorManager := gedcom.NewErrorManager()
    p := parser.NewHierarchicalParser()

    var exporter exporter.Exporter
    var extension string

    switch outputFormat {
    case "json":
        exporter = exporter.NewJsonExporter(errorManager)
        extension = ".json"
    case "xml":
        exporter = exporter.NewXMLExporter(errorManager)
        extension = ".xml"
    case "yaml":
        exporter = exporter.NewYAMLExporter(errorManager)
        extension = ".yaml"
    default:
        return fmt.Errorf("unsupported format: %s", outputFormat)
    }

    for _, inputFile := range inputFiles {
        // Parse
        tree, err := p.Parse(inputFile)
        if err != nil {
            fmt.Printf("Error parsing %s: %v\n", inputFile, err)
            continue
        }

        // Export
        outputFile := strings.TrimSuffix(inputFile, ".ged") + extension
        if err := exporter.ExportToFile(tree, outputFile); err != nil {
            fmt.Printf("Error exporting %s: %v\n", inputFile, err)
            continue
        }

        fmt.Printf("✓ Exported %s → %s\n", inputFile, outputFile)
    }

    return nil
}
```

---

## Best Practices

### Error Handling

Always check for errors and use the error manager:

```go
errorManager := gedcom.NewErrorManager()
exporter := exporter.NewJsonExporter(errorManager)

err := exporter.ExportToFile(tree, "output.json")
if err != nil {
    log.Fatalf("Export failed: %v", err)
}

// Check for warnings/info
errors := errorManager.Errors()
for _, err := range errors {
    switch err.Severity {
    case gedcom.SeveritySevere:
        log.Printf("SEVERE: %s", err.Message)
    case gedcom.SeverityWarning:
        log.Printf("WARNING: %s", err.Message)
    }
}
```

### Memory Management

For large files, consider exporting directly to file rather than string:

```go
// Good: Direct file export (memory efficient)
err := exporter.ExportToFile(tree, "output.json")

// Less efficient: String export (loads entire output in memory)
jsonString, err := exporter.ExportToString(tree)
```

### File Paths

Always use absolute paths or ensure the directory exists:

```go
import "path/filepath"

outputPath := filepath.Join("/path/to/output", "family.json")
err := exporter.ExportToFile(tree, outputPath)
```

### GEDCOM Export Configuration

When exporting to GEDCOM, provide meaningful application information:

```go
gedcomExporter := exporter.NewGedcomExporter(
    errorManager,
    "MyGenealogyApp",  // Descriptive name
    "2.1.0",           // Version number
)
```

---

## Troubleshooting

### Common Issues

#### 1. "Failed to write file" Error

**Problem:** Cannot write to output file.

**Solutions:**
- Check file permissions
- Ensure directory exists
- Verify disk space
- Use absolute paths

```go
// Ensure directory exists
import "os"
import "path/filepath"

outputDir := filepath.Dir(outputFile)
os.MkdirAll(outputDir, 0755)
```

#### 2. "Failed to marshal JSON/XML/YAML" Error

**Problem:** Data structure cannot be serialized.

**Solutions:**
- Check for circular references (shouldn't happen with GEDCOM)
- Verify tree structure is valid
- Check error manager for validation errors

#### 3. GEDCOM Export Line Length Issues

**Problem:** Lines exceed 255 characters.

**Solution:** The exporter automatically handles this with CONC/CONT. If you see issues, check the GEDCOM specification compliance.

#### 4. Missing Submitter in GEDCOM Export

**Problem:** GEDCOM export fails because no submitter exists.

**Solution:** The exporter automatically creates a submitter if missing. Check error messages for details.

### Performance Tips

1. **Use file export for large datasets** (avoids loading entire output in memory)
2. **Reuse error managers** across multiple exports
3. **Parse once, export multiple times** (don't re-parse for each format)

### Debugging

Enable verbose error reporting:

```go
errorManager := gedcom.NewErrorManager()
exporter := exporter.NewJsonExporter(errorManager)

err := exporter.ExportToFile(tree, "output.json")
if err != nil {
    // Check all errors
    for _, exportErr := range errorManager.Errors() {
        fmt.Printf("[%s] %s (Line %d)\n", 
            exportErr.Severity, 
            exportErr.Message, 
            exportErr.LineNumber)
    }
}
```

---

## See Also

- [CLI Documentation](cli.md) - Command-line interface for exports
- [Parser Documentation](../internal/parser/doc.go) - Parsing GEDCOM files
- [GEDCOM Specification](https://www.gedcom.org/) - Official GEDCOM 5.5.1 specification

---

**Last Updated:** 2025-01-27
