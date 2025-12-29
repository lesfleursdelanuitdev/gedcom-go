package query

import (
	"fmt"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
)

// QueryMetrics holds query performance metrics
type QueryMetrics struct {
	QueryType        string
	FirstExecution   time.Duration // Cold cache
	AvgExecution     time.Duration
	MinExecution     time.Duration
	MaxExecution     time.Duration
	MedianExecution  time.Duration
	CachedExecution  time.Duration // Warm cache
	CacheSpeedup     float64        // X times faster
	ResultCount      int
	MemoryDelta      uint64
	Iterations       int
}

// TestIndividuals holds representative individuals for testing
type TestIndividuals struct {
	WithParents      string // Individual with both parents
	WithChildren     string // Individual with children
	WithSiblings     string // Individual with siblings
	WithSpouses      string // Individual with spouses
	DeepAncestry     string // Individual with deep ancestry (10+ generations)
	ManyDescendants  string // Individual with many descendants (50+)
	MultipleFamilies string // Individual in multiple families
	RootAncestor     string // Individual with no parents
	LeafIndividual   string // Individual with no descendants
	NoFamilies       string // Individual with no families
}

// measureMemoryTestData returns current memory usage in bytes
func measureMemoryTestData() uint64 {
	var m runtime.MemStats
	runtime.GC() // Force GC before measurement
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// measureQueryPerformance measures query performance over multiple iterations
func measureQueryPerformance(queryType string, iterations int, queryFunc func() (int, error)) (*QueryMetrics, error) {
	if iterations < 1 {
		iterations = 10 // Default
	}

	metrics := &QueryMetrics{
		QueryType:  queryType,
		Iterations: iterations,
	}

	durations := make([]time.Duration, 0, iterations)
	memoryBefore := measureMemoryTestData()

	// First execution (cold cache)
	start := time.Now()
	resultCount, err := queryFunc()
	metrics.FirstExecution = time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	metrics.ResultCount = resultCount
	durations = append(durations, metrics.FirstExecution)

	// Subsequent executions
	for i := 1; i < iterations; i++ {
		start := time.Now()
		_, err := queryFunc()
		duration := time.Since(start)
		if err != nil {
			return nil, fmt.Errorf("query failed on iteration %d: %w", i, err)
		}
		durations = append(durations, duration)
	}

	// Calculate statistics
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	metrics.MinExecution = durations[0]
	metrics.MaxExecution = durations[len(durations)-1]
	metrics.MedianExecution = durations[len(durations)/2]

	// Calculate average (excluding first execution for cached average)
	if len(durations) > 1 {
		var sum time.Duration
		for i := 1; i < len(durations); i++ {
			sum += durations[i]
		}
		metrics.CachedExecution = sum / time.Duration(len(durations)-1)
	} else {
		metrics.CachedExecution = metrics.FirstExecution
	}

	// Calculate overall average
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	metrics.AvgExecution = sum / time.Duration(len(durations))

	// Calculate cache speedup
	if metrics.CachedExecution > 0 {
		metrics.CacheSpeedup = float64(metrics.FirstExecution) / float64(metrics.CachedExecution)
	} else {
		metrics.CacheSpeedup = 1.0
	}

	memoryAfter := measureMemoryTestData()
	metrics.MemoryDelta = memoryAfter - memoryBefore

	return metrics, nil
}

// selectTestIndividuals selects representative individuals from the graph for testing
func selectTestIndividuals(graph *Graph, qb *QueryBuilder) TestIndividuals {
	ti := TestIndividuals{}

	allIndividuals := graph.GetAllIndividuals()
	if len(allIndividuals) == 0 {
		return ti
	}

	// Convert to slice for iteration
	individuals := make([]string, 0, len(allIndividuals))
	for xref := range allIndividuals {
		individuals = append(individuals, xref)
	}

	// Find individuals with various characteristics
	for _, xref := range individuals {
		iq := qb.Individual(xref)

		// Find individual with parents
		if ti.WithParents == "" {
			parents, err := iq.Parents()
			if err == nil && len(parents) >= 2 {
				ti.WithParents = xref
			}
		}

		// Find individual with children
		if ti.WithChildren == "" {
			children, err := iq.Children()
			if err == nil && len(children) > 0 {
				ti.WithChildren = xref
			}
		}

		// Find individual with siblings
		if ti.WithSiblings == "" {
			siblings, err := iq.Siblings()
			if err == nil && len(siblings) > 0 {
				ti.WithSiblings = xref
			}
		}

		// Find individual with spouses
		if ti.WithSpouses == "" {
			spouses, err := iq.Spouses()
			if err == nil && len(spouses) > 0 {
				ti.WithSpouses = xref
			}
		}

		// Find individual with deep ancestry
		if ti.DeepAncestry == "" {
			ancestors, err := iq.Ancestors().Execute()
			if err == nil && len(ancestors) >= 10 {
				ti.DeepAncestry = xref
			}
		}

		// Find individual with many descendants
		if ti.ManyDescendants == "" {
			descendants, err := iq.Descendants().Execute()
			if err == nil && len(descendants) >= 50 {
				ti.ManyDescendants = xref
			}
		}

		// Find root ancestor (no parents)
		if ti.RootAncestor == "" {
			parents, err := iq.Parents()
			if err == nil && len(parents) == 0 {
				ti.RootAncestor = xref
			}
		}

		// Find leaf individual (no descendants)
		if ti.LeafIndividual == "" {
			descendants, err := iq.Descendants().Execute()
			if err == nil && len(descendants) == 0 {
				ti.LeafIndividual = xref
			}
		}

		// Find individual with multiple families
		if ti.MultipleFamilies == "" {
			// Count families where individual is a spouse
			node := graph.GetIndividual(xref)
			if node != nil && node.Individual != nil {
				fams := node.Individual.GetFamiliesAsSpouse()
				if len(fams) >= 2 {
					ti.MultipleFamilies = xref
				}
			}
		}

		// If we found most of them, we can break early
		if ti.WithParents != "" && ti.WithChildren != "" && ti.WithSiblings != "" &&
			ti.WithSpouses != "" && ti.RootAncestor != "" && ti.LeafIndividual != "" {
			break
		}
	}

	// Use first individual as fallback for any missing
	if ti.WithParents == "" && len(individuals) > 0 {
		ti.WithParents = individuals[0]
	}
	if ti.WithChildren == "" && len(individuals) > 0 {
		ti.WithChildren = individuals[0]
	}
	if ti.RootAncestor == "" && len(individuals) > 0 {
		ti.RootAncestor = individuals[0]
	}
	if ti.LeafIndividual == "" && len(individuals) > 0 {
		ti.LeafIndividual = individuals[0]
	}

	return ti
}

// printQueryPerformanceReport prints a formatted query performance report
func printQueryPerformanceReport(metrics *QueryMetrics) {
	fmt.Printf("  %-30s: %6.2fms (avg), %6.2fms (cached) - %5.1fx speedup, %d results\n",
		metrics.QueryType,
		metrics.AvgExecution.Seconds()*1000,
		metrics.CachedExecution.Seconds()*1000,
		metrics.CacheSpeedup,
		metrics.ResultCount)
}

// TestQueryPerformance_AllTestDataFiles tests query performance on all test data files
func TestQueryPerformance_AllTestDataFiles(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}

	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged", "pres2020.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
				return
			}

			// Parse and build graph
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			graph, err := BuildGraph(tree)
			if err != nil {
				t.Fatalf("Failed to build graph for %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder: %v", err)
			}

			// Select test individuals
			testIndis := selectTestIndividuals(graph, qb)

			fmt.Printf("\n=== Query Performance Results - %s ===\n", filename)
			fmt.Printf("==========================================\n\n")

			// Test relationship queries
			fmt.Printf("Relationship Queries:\n")
			
			if testIndis.WithParents != "" {
				metrics, err := measureQueryPerformance("Parents", 10, func() (int, error) {
					results, err := qb.Individual(testIndis.WithParents).Parents()
					if err != nil {
						return 0, err
					}
					return len(results), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			if testIndis.WithChildren != "" {
				metrics, err := measureQueryPerformance("Children", 10, func() (int, error) {
					results, err := qb.Individual(testIndis.WithChildren).Children()
					if err != nil {
						return 0, err
					}
					return len(results), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			if testIndis.WithSiblings != "" {
				metrics, err := measureQueryPerformance("Siblings", 10, func() (int, error) {
					results, err := qb.Individual(testIndis.WithSiblings).Siblings()
					if err != nil {
						return 0, err
					}
					return len(results), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			if testIndis.WithSpouses != "" {
				metrics, err := measureQueryPerformance("Spouses", 10, func() (int, error) {
					results, err := qb.Individual(testIndis.WithSpouses).Spouses()
					if err != nil {
						return 0, err
					}
					return len(results), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			// Test ancestral queries
			fmt.Printf("\nAncestral Queries:\n")

			if testIndis.DeepAncestry != "" {
				metrics, err := measureQueryPerformance("Ancestors (deep)", 5, func() (int, error) {
					results, err := qb.Individual(testIndis.DeepAncestry).Ancestors().Execute()
					if err != nil {
						return 0, err
					}
					return len(results), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			} else if testIndis.WithParents != "" {
				metrics, err := measureQueryPerformance("Ancestors", 5, func() (int, error) {
					results, err := qb.Individual(testIndis.WithParents).Ancestors().Execute()
					if err != nil {
						return 0, err
					}
					return len(results), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			if testIndis.ManyDescendants != "" {
				metrics, err := measureQueryPerformance("Descendants (many)", 5, func() (int, error) {
					results, err := qb.Individual(testIndis.ManyDescendants).Descendants().Execute()
					if err != nil {
						return 0, err
					}
					return len(results), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			} else if testIndis.WithChildren != "" {
				metrics, err := measureQueryPerformance("Descendants", 5, func() (int, error) {
					results, err := qb.Individual(testIndis.WithChildren).Descendants().Execute()
					if err != nil {
						return 0, err
					}
					return len(results), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			// Test family queries
			fmt.Printf("\nFamily Queries:\n")

			if testIndis.WithParents != "" {
				metrics, err := measureQueryPerformance("Families for Individual", 10, func() (int, error) {
					node := graph.GetIndividual(testIndis.WithParents)
					if node == nil || node.Individual == nil {
						return 0, nil
					}
					// Get families where individual is a child
					famc := node.Individual.GetFamiliesAsChild()
					// Get families where individual is a spouse
					fams := node.Individual.GetFamiliesAsSpouse()
					return len(famc) + len(fams), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			// Get a family to test with
			allFamilies := graph.GetAllFamilies()
			var testFamilyXref string
			if len(allFamilies) > 0 {
				for xref := range allFamilies {
					testFamilyXref = xref
					break
				}
			}

			if testFamilyXref != "" {
				metrics, err := measureQueryPerformance("Family Members", 10, func() (int, error) {
					husband, _ := qb.Family(testFamilyXref).Husband()
					wife, _ := qb.Family(testFamilyXref).Wife()
					children, err := qb.Family(testFamilyXref).Children()
					if err != nil {
						return 0, err
					}
					count := 0
					if husband != nil {
						count++
					}
					if wife != nil {
						count++
					}
					return count + len(children), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			if testIndis.WithChildren != "" {
				metrics, err := measureQueryPerformance("Family with Given Parent", 10, func() (int, error) {
					// Get families where individual is a spouse
					node := graph.GetIndividual(testIndis.WithChildren)
					if node == nil || node.Individual == nil {
						return 0, nil
					}
					fams := node.Individual.GetFamiliesAsSpouse()
					return len(fams), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			// Test path finding
			fmt.Printf("\nPath Finding:\n")

			if testIndis.WithParents != "" && testIndis.WithChildren != "" {
				metrics, err := measureQueryPerformance("Shortest Path", 5, func() (int, error) {
					path, err := qb.Individual(testIndis.WithParents).PathTo(testIndis.WithChildren).Shortest()
					if err != nil {
						return 0, err
					}
					if path == nil {
						return 0, nil
					}
					return len(path.Nodes), nil
				})
				if err == nil {
					printQueryPerformanceReport(metrics)
				}
			}

			// Test filter queries
			fmt.Printf("\nFilter Queries:\n")

			metrics, err := measureQueryPerformance("Filter by Name", 10, func() (int, error) {
				results, err := qb.Filter().ByName("John").Execute()
				if err != nil {
					return 0, err
				}
				return len(results), nil
			})
			if err == nil {
				printQueryPerformanceReport(metrics)
			}

			metrics, err = measureQueryPerformance("Filter by Birth Date", 10, func() (int, error) {
				start := time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(1900, 12, 31, 23, 59, 59, 999999999, time.UTC)
				results, err := qb.Filter().ByBirthDate(start, end).Execute()
				if err != nil {
					return 0, err
				}
				return len(results), nil
			})
			if err == nil {
				printQueryPerformanceReport(metrics)
			}
		})
	}
}

