// Package diff provides semantic comparison functionality for GEDCOM files.
//
// The diff system identifies differences between two GEDCOM files at a
// meaningful level, not just line-by-line text comparison. It understands
// GEDCOM structure and reports changes in terms of:
//   - Added/removed individuals, families, notes, sources
//   - Modified record data (names, dates, places, events)
//   - Relationship changes (family connections)
//   - Structural differences (hierarchical changes)
//
// The system also tracks change history, recording who, when, and what changed.
//
// Basic Usage:
//
//	// Create a differ with default configuration
//	differ := diff.NewGedcomDiffer(diff.DefaultConfig())
//
//	// Compare two GEDCOM trees
//	result, err := differ.Compare(tree1, tree2)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Generate text report
//	report, err := differ.GenerateReport(result)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(report)
//
// Matching Strategies:
//
// The differ supports three matching strategies:
//
//   - "xref": Match records by XREF ID (fast, for same-file versions)
//   - "content": Match records by content similarity (slower, for cross-file)
//   - "hybrid": XREF first, content fallback (balanced)
//
// Configuration:
//
//	config := &diff.DiffConfig{
//		MatchingStrategy:  "hybrid",
//		SimilarityThreshold: 0.85,
//		DateTolerance:     2,
//		DetailLevel:       "field",
//		OutputFormat:      "text",
//		TrackHistory:      true,
//	}
//
// Change History:
//
// When TrackHistory is enabled, the system records:
//   - Timestamp of each change
//   - Field that changed
//   - Old and new values
//   - Optional author and reason
//
// Output Formats:
//
//   - "text": Human-readable text report
//   - "json": Structured JSON (coming soon)
//   - "html": Visual HTML diff (coming soon)
//   - "unified": Git-style unified diff (coming soon)
//
// Semantic Understanding:
//
// The differ understands semantic equivalence:
//   - Dates: "1800" ≈ "ABT 1800" (within tolerance)
//   - Places: "New York" ≈ "New York, NY" (hierarchy match)
//   - Relationships: Tracks family structure changes
package diff
