# Usage Examples

This document provides comprehensive examples of using the gedcom-go library.

## Table of Contents

1. [Basic Parsing](#basic-parsing)
2. [Accessing Records](#accessing-records)
3. [Validation](#validation)
4. [Exporting](#exporting)
5. [Error Handling](#error-handling)
6. [Advanced Usage](#advanced-usage)

## Basic Parsing

### Parse a GEDCOM File

```go
package main

import (
	"fmt"
	"log"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
)

func main() {
	// Create a hierarchical parser
	p := parser.NewHierarchicalParser()
	
	// Parse the file
	tree, err := p.Parse("family.ged")
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}
	
	// Check for parsing errors
	if p.HasErrors() {
		errors := p.GetErrors()
		fmt.Printf("Found %d parsing errors:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  %s\n", err)
		}
	}
	
	fmt.Printf("Successfully parsed GEDCOM file\n")
}
```

### Using Basic Parser (Backward Compatible)

```go
p := parser.NewBasicParser()
tree, err := p.Parse("family.ged")
if err != nil {
	log.Fatal(err)
}
```

## Accessing Records

### Get All Individuals

```go
individuals := tree.GetAllIndividuals()
for xrefID, record := range individuals {
	indi := record.(*gedcom.IndividualRecord)
	name := indi.GetName()
	birthDate := indi.GetBirthDate()
	fmt.Printf("%s: %s (born %s)\n", xrefID, name, birthDate)
}
```

### Get Individual by Xref

```go
indi := tree.GetIndividual("@I1@")
if indi != nil {
	fmt.Printf("Name: %s\n", indi.GetName())
	fmt.Printf("Sex: %s\n", indi.GetSex())
}
```

### Get Families

```go
families := tree.GetAllFamilies()
for xrefID, record := range families {
	fam := record.(*gedcom.FamilyRecord)
	husband := fam.GetHusband()
	wife := fam.GetWife()
	children := fam.GetChildren()
	
	fmt.Printf("Family %s:\n", xrefID)
	fmt.Printf("  Husband: %s\n", husband)
	fmt.Printf("  Wife: %s\n", wife)
	fmt.Printf("  Children: %d\n", len(children))
}
```

### Access Individual Data

```go
indi := tree.GetIndividual("@I1@")
if indi != nil {
	// Basic information
	fmt.Printf("Name: %s\n", indi.GetName())
	fmt.Printf("Sex: %s\n", indi.GetSex())
	
	// Birth information
	birthDate := indi.GetBirthDate()
	birthPlace := indi.GetBirthPlace()
	if birthDate != "" || birthPlace != "" {
		fmt.Printf("Born: %s in %s\n", birthDate, birthPlace)
	}
	
	// Death information
	deathDate := indi.GetDeathDate()
	deathPlace := indi.GetDeathPlace()
	if deathDate != "" || deathPlace != "" {
		fmt.Printf("Died: %s in %s\n", deathDate, deathPlace)
	}
	
	// Family relationships
	familiesAsChild := indi.GetFamiliesAsChild()
	familiesAsSpouse := indi.GetFamiliesAsSpouse()
	fmt.Printf("Families as child: %d\n", len(familiesAsChild))
	fmt.Printf("Families as spouse: %d\n", len(familiesAsSpouse))
}
```

### Access Family Data

```go
fam := tree.GetFamily("@F1@")
if fam != nil {
	// Marriage information
	marriageDate := fam.GetMarriageDate()
	marriagePlace := fam.GetMarriagePlace()
	if marriageDate != "" || marriagePlace != "" {
		fmt.Printf("Married: %s in %s\n", marriageDate, marriagePlace)
	}
	
	// Divorce information
	divorceDate := fam.GetDivorceDate()
	divorcePlace := fam.GetDivorcePlace()
	if divorceDate != "" || divorcePlace != "" {
		fmt.Printf("Divorced: %s in %s\n", divorceDate, divorcePlace)
	}
	
	// Children
	children := fam.GetChildren()
	for _, childXref := range children {
		child := tree.GetIndividual(childXref)
		if child != nil {
			fmt.Printf("  Child: %s\n", child.GetName())
		}
	}
}
```

## Validation

### Basic Validation

```go
package main

import (
	"fmt"
	"log"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/validator"
)

func main() {
	// Parse file
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse("family.ged")
	if err != nil {
		log.Fatal(err)
	}
	
	// Create validator
	errorManager := gedcom.NewErrorManager()
	v := validator.NewGedcomValidator(errorManager)
	
	// Validate
	err = v.Validate(tree)
	if err != nil {
		log.Fatal(err)
	}
	
	// Check errors
	errors := errorManager.Errors()
	if len(errors) > 0 {
		fmt.Printf("Found %d validation errors:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  [%s] %s (line %d)\n", 
				err.Severity, err.Message, err.LineNumber)
		}
	} else {
		fmt.Println("No validation errors found")
	}
}
```

### Parallel Validation (for Large Files)

```go
errorManager := gedcom.NewErrorManager()
parallelValidator := validator.NewParallelGedcomValidator(errorManager)
err := parallelValidator.Validate(tree)
if err != nil {
	log.Fatal(err)
}

errors := errorManager.Errors()
fmt.Printf("Found %d validation errors\n", len(errors))
```

### Individual Validator Only

```go
errorManager := gedcom.NewErrorManager()
indiValidator := validator.NewIndividualValidator(errorManager)
err := indiValidator.Validate(tree)
if err != nil {
	log.Fatal(err)
}
```

## Exporting

### Export to JSON

```go
package main

import (
	"fmt"
	"log"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/exporter"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
)

func main() {
	// Parse file
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse("family.ged")
	if err != nil {
		log.Fatal(err)
	}
	
	// Export to JSON
	jsonExporter := exporter.NewJsonExporter()
	
	// Export to string
	json, err := jsonExporter.ExportToString(tree)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(json)
	
	// Export to file
	err = jsonExporter.ExportToFile(tree, "family.json")
	if err != nil {
		log.Fatal(err)
	}
}
```

### Export to XML

```go
xmlExporter := exporter.NewXMLExporter()
xml, err := xmlExporter.ExportToString(tree)
if err != nil {
	log.Fatal(err)
}
fmt.Println(xml)

// Or export to file
err = xmlExporter.ExportToFile(tree, "family.xml")
```

### Export to YAML

```go
yamlExporter := exporter.NewYAMLExporter()
yaml, err := yamlExporter.ExportToString(tree)
if err != nil {
	log.Fatal(err)
}
fmt.Println(yaml)

// Or export to file
err = yamlExporter.ExportToFile(tree, "family.yaml")
```

### Export Back to GEDCOM

```go
gedcomExporter := exporter.NewGedcomExporter()
err := gedcomExporter.ExportToFile(tree, "output.ged")
if err != nil {
	log.Fatal(err)
}
```

## Error Handling

### Comprehensive Error Handling

```go
package main

import (
	"fmt"
	"log"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/validator"
)

func main() {
	// Parse with error collection
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse("family.ged")
	if err != nil {
		log.Fatalf("Parse failed: %v", err)
	}
	
	// Check parsing errors
	if p.HasErrors() {
		errors := p.GetErrors()
		severeErrors := []*gedcom.GedcomError{}
		warnings := []*gedcom.GedcomError{}
		
		for _, err := range errors {
			if err.Severity == gedcom.SeveritySevere {
				severeErrors = append(severeErrors, err)
			} else {
				warnings = append(warnings, err)
			}
		}
		
		fmt.Printf("Parsing: %d severe errors, %d warnings\n", 
			len(severeErrors), len(warnings))
	}
	
	// Validate
	errorManager := gedcom.NewErrorManager()
	v := validator.NewGedcomValidator(errorManager)
	err = v.Validate(tree)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
	
	// Check validation errors
	errors := errorManager.Errors()
	if len(errors) > 0 {
		fmt.Printf("Validation: %d errors found\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  [%s] %s (line %d, context: %s)\n",
				err.Severity, err.Message, err.LineNumber, err.Context)
		}
	}
}
```

## Advanced Usage

### Working with GedcomLine Directly

```go
// Get the first line of a record
indi := tree.GetIndividual("@I1@")
if indi != nil {
	firstLine := indi.FirstLine()
	
	// Access children by tag
	nameLines := firstLine.GetLines("NAME")
	for _, nameLine := range nameLines {
		fmt.Printf("Name: %s\n", nameLine.Value)
		
		// Access sub-components
		givenName := nameLine.GetValue("GIVN")
		surname := nameLine.GetValue("SURN")
		fmt.Printf("  Given: %s, Surname: %s\n", givenName, surname)
	}
	
	// Get all birth events
	birthLines := firstLine.GetLines("BIRT")
	for _, birthLine := range birthLines {
		date := birthLine.GetValue("DATE")
		place := birthLine.GetValue("PLAC")
		fmt.Printf("Birth: %s in %s\n", date, place)
	}
}
```

### Modifying Records

```go
// Set a value using dot notation
indi := tree.GetIndividual("@I1@")
if indi != nil {
	firstLine := indi.FirstLine()
	
	// Update birth date
	err := firstLine.SetValue("BIRT.DATE", "1 JAN 1900")
	if err != nil {
		log.Printf("Failed to set value: %v", err)
	}
	
	// Add a note
	err = firstLine.SetValue("NOTE", "This is a note")
	if err != nil {
		log.Printf("Failed to add note: %v", err)
	}
}
```

### Converting Between Formats

```go
// GEDCOM → JSON → XML → YAML → GEDCOM

// Step 1: Parse GEDCOM
p := parser.NewHierarchicalParser()
tree, err := p.Parse("input.ged")
if err != nil {
	log.Fatal(err)
}

// Step 2: Export to JSON
jsonExporter := exporter.NewJsonExporter()
err = jsonExporter.ExportToFile(tree, "output.json")
if err != nil {
	log.Fatal(err)
}

// Step 3: Export to XML
xmlExporter := exporter.NewXMLExporter()
err = xmlExporter.ExportToFile(tree, "output.xml")
if err != nil {
	log.Fatal(err)
}

// Step 4: Export to YAML
yamlExporter := exporter.NewYAMLExporter()
err = yamlExporter.ExportToFile(tree, "output.yaml")
if err != nil {
	log.Fatal(err)
}

// Step 5: Export back to GEDCOM
gedcomExporter := exporter.NewGedcomExporter()
err = gedcomExporter.ExportToFile(tree, "output.ged")
if err != nil {
	log.Fatal(err)
}
```

### Using Two-Phase Parser (for Large Files)

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"

twoPhaseParser := parser.NewTwoPhaseParser()
tree, err := twoPhaseParser.Parse("large_family.ged")
if err != nil {
	log.Fatal(err)
}

// Two-phase parser is optimized for large files
fmt.Printf("Parsed %d individuals\n", len(tree.GetAllIndividuals()))
```

### Thread-Safe Access

```go
// GedcomTree is thread-safe
tree := parser.NewHierarchicalParser().Parse("family.ged")

// Multiple goroutines can access the tree concurrently
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		individuals := tree.GetAllIndividuals()
		fmt.Printf("Goroutine found %d individuals\n", len(individuals))
	}()
}
wg.Wait()
```

## Complete Example: Family Tree Analysis

```go
package main

import (
	"fmt"
	"log"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/validator"
)

func main() {
	// Parse
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse("family.ged")
	if err != nil {
		log.Fatal(err)
	}
	
	// Validate
	errorManager := gedcom.NewErrorManager()
	v := validator.NewGedcomValidator(errorManager)
	err = v.Validate(tree)
	if err != nil {
		log.Fatal(err)
	}
	
	// Analyze
	fmt.Println("=== Family Tree Analysis ===\n")
	
	// Individuals
	individuals := tree.GetAllIndividuals()
	fmt.Printf("Individuals: %d\n", len(individuals))
	
	// Families
	families := tree.GetAllFamilies()
	fmt.Printf("Families: %d\n\n", len(families))
	
	// Print individuals
	fmt.Println("=== Individuals ===")
	for xrefID, record := range individuals {
		indi := record.(*gedcom.IndividualRecord)
		fmt.Printf("%s: %s", xrefID, indi.GetName())
		if sex := indi.GetSex(); sex != "" {
			fmt.Printf(" (%s)", sex)
		}
		if birthDate := indi.GetBirthDate(); birthDate != "" {
			fmt.Printf(" - Born: %s", birthDate)
		}
		if deathDate := indi.GetDeathDate(); deathDate != "" {
			fmt.Printf(" - Died: %s", deathDate)
		}
		fmt.Println()
	}
	
	// Print families
	fmt.Println("\n=== Families ===")
	for xrefID, record := range families {
		fam := record.(*gedcom.FamilyRecord)
		fmt.Printf("%s: ", xrefID)
		
		if husband := fam.GetHusband(); husband != "" {
			husbandIndi := tree.GetIndividual(husband)
			if husbandIndi != nil {
				fmt.Printf("%s", husbandIndi.GetName())
			}
		}
		
		fmt.Printf(" & ")
		
		if wife := fam.GetWife(); wife != "" {
			wifeIndi := tree.GetIndividual(wife)
			if wifeIndi != nil {
				fmt.Printf("%s", wifeIndi.GetName())
			}
		}
		
		if marriageDate := fam.GetMarriageDate(); marriageDate != "" {
			fmt.Printf(" (married %s)", marriageDate)
		}
		
		children := fam.GetChildren()
		if len(children) > 0 {
			fmt.Printf(" - %d children", len(children))
		}
		
		fmt.Println()
	}
	
	// Validation errors
	errors := errorManager.Errors()
	if len(errors) > 0 {
		fmt.Printf("\n=== Validation Errors: %d ===\n", len(errors))
		for _, err := range errors {
			fmt.Printf("[%s] %s (line %d)\n", 
				err.Severity, err.Message, err.LineNumber)
		}
	}
}
```


