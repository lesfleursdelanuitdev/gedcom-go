// Package validator provides validation functionality for GEDCOM trees.
//
// The validator package implements comprehensive validation rules based on
// the GEDCOM 5.5.1 specification, checking record structure, cross-references,
// required fields, and data format compliance.
//
// # Validators
//
// The package provides several specialized validators:
//
//   - GedcomValidator: Orchestrates all validators
//   - IndividualValidator: Validates individual (INDI) records
//   - FamilyValidator: Validates family (FAM) records
//   - CrossReferenceValidator: Validates cross-references between records
//   - HeaderValidator: Validates header (HEAD) record
//   - ParallelGedcomValidator: Parallel version of GedcomValidator
//   - ParallelIndividualValidator: Parallel version of IndividualValidator
//   - AdvancedValidator: Pluggable rule system for advanced validation
//   - DateConsistencyValidator: Validates date consistency (Phase 1)
//
// # Validation Rules
//
// Validators check:
//
//   - Required tags: Ensures mandatory fields are present
//   - Tag validity: Verifies tags are valid GEDCOM tags or user-defined
//   - Cross-references: Validates xref format and resolution
//   - Event structure: Validates event tags and subtags
//   - Name structure: Validates name components
//   - Value formats: Validates SEX, DATE, and other typed values
//
// # Usage Example
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/yourorg/gedcom/pkg/gedcom"
//		"github.com/yourorg/gedcom/internal/parser"
//		"github.com/yourorg/gedcom/internal/validator"
//	)
//
//	func main() {
//		// Parse file
//		p := parser.NewHierarchicalParser()
//		tree, err := p.Parse("family.ged")
//		if err != nil {
//			panic(err)
//		}
//
//		// Validate
//		errorManager := gedcom.NewErrorManager()
//		v := validator.NewGedcomValidator(errorManager)
//		err = v.Validate(tree)
//		if err != nil {
//			panic(err)
//		}
//
//		// Check errors
//		errors := errorManager.Errors()
//		if len(errors) > 0 {
//			fmt.Printf("Found %d validation errors:\n", len(errors))
//			for _, err := range errors {
//				fmt.Printf("  %s\n", err)
//			}
//		}
//	}
//
// # Parallel Validation
//
// For large files, use parallel validators for better performance:
//
//	parallelValidator := validator.NewParallelGedcomValidator(errorManager)
//	err = parallelValidator.Validate(tree)
//
// # Error Severity
//
// Errors are categorized by severity:
//
//   - Hint: Best practices, optimizations (optional improvements)
//   - Info: Data quality issues, suggestions (consider improving)
//   - Warning: Unlikely but possible situations (should review)
//   - Severe: Impossible situations, must fix (blocks processing)
//
// # Advanced Validation
//
// Advanced validation provides comprehensive data quality checks beyond
// basic GEDCOM compliance:
//
//   - Date Consistency: Birth before death, reasonable ages, parent-child gaps
//   - Relationship Logic: Bidirectional links, circular references (future)
//   - Duplicate Detection: Exact and fuzzy duplicate detection (future)
//   - Data Quality: Missing data, consistency checks (future)
//
// Usage:
//
//	errorManager := gedcom.NewErrorManager()
//	validator := NewGedcomValidator(errorManager)
//	validator.EnableAdvancedValidation() // Enable with defaults
//	err := validator.Validate(tree)
//
// Or with custom configuration:
//
//	config := NewValidationConfig()
//	config.MinParentAge = 12  // Custom threshold
//	config.MinSeverity = gedcom.SeverityWarning  // Filter severity
//	validator.EnableAdvancedValidationWithConfig(config)
package validator
