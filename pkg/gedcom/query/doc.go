// Package query provides a fluent, builder-style API for querying GEDCOM genealogical data.
//
// The query package builds on top of a graph representation of GEDCOM data, enabling
// efficient traversal, path finding, relationship calculation, and complex filtering.
//
// # Quick Start
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/yourorg/gedcom/pkg/gedcom"
//		"github.com/yourorg/gedcom/internal/parser"
//		"github.com/yourorg/gedcom/pkg/gedcom/query"
//	)
//
//	func main() {
//		// Parse GEDCOM file
//		p := parser.NewHierarchicalParser()
//		tree, err := p.Parse("family.ged")
//		if err != nil {
//			panic(err)
//		}
//
//		// Create query builder
//		q, err := query.NewQuery(tree)
//		if err != nil {
//			panic(err)
//		}
//
//		// Find all ancestors
//		ancestors, _ := q.Individual("@I1@").Ancestors().Execute()
//		for _, ancestor := range ancestors {
//			fmt.Printf("Ancestor: %s\n", ancestor.GetName())
//		}
//	}
//
// # Query Types
//
// ## IndividualQuery
//
// Query operations starting from a specific individual:
//
//	// Direct relationships
//	parents, _ := q.Individual("@I1@").Parents()
//	children, _ := q.Individual("@I1@").Children()
//	siblings, _ := q.Individual("@I1@").Siblings()
//	spouses, _ := q.Individual("@I1@").Spouses()
//
//	// Extended relationships
//	grandparents, _ := q.Individual("@I1@").Grandparents()
//	uncles, _ := q.Individual("@I1@").Uncles()
//	cousins, _ := q.Individual("@I1@").Cousins(1) // 1st cousins
//
//	// Complex queries
//	ancestors, _ := q.Individual("@I1@").Ancestors().MaxGenerations(5).Execute()
//	descendants, _ := q.Individual("@I1@").Descendants().IncludeSelf().Execute()
//
// ## AncestorQuery
//
// Configurable ancestor search with options:
//
//	ancestors, _ := q.Individual("@I1@").
//		Ancestors().
//		MaxGenerations(3).        // Limit to 3 generations
//		IncludeSelf().            // Include starting individual
//		Filter(func(indi *gedcom.IndividualRecord) bool {
//			return indi.GetSex() == "M"  // Only males
//		}).
//		Execute()
//
//	count, _ := q.Individual("@I1@").Ancestors().Count()
//	exists, _ := q.Individual("@I1@").Ancestors().Exists()
//
// ## DescendantQuery
//
// Configurable descendant search (same API as AncestorQuery):
//
//	descendants, _ := q.Individual("@I1@").
//		Descendants().
//		MaxGenerations(2).
//		Execute()
//
// ## RelationshipQuery
//
// Calculate relationship between two individuals:
//
//	result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()
//
//	fmt.Printf("Relationship: %s\n", result.RelationshipType)
//	fmt.Printf("Degree: %d\n", result.Degree)        // For cousins: 1st, 2nd, etc.
//	fmt.Printf("Removal: %d\n", result.Removal)      // For removed cousins
//	fmt.Printf("Is Direct: %v\n", result.IsDirect)
//	fmt.Printf("Is Collateral: %v\n", result.IsCollateral)
//
// ## PathQuery
//
// Find paths between two individuals:
//
//	// Shortest path
//	path, _ := q.Individual("@I1@").PathTo("@I2@").Shortest()
//	fmt.Printf("Path length: %d\n", path.Length)
//
//	// All paths
//	paths, _ := q.Individual("@I1@").
//		PathTo("@I2@").
//		MaxLength(10).
//		IncludeBlood(true).
//		IncludeMarital(false).
//		All()
//
// ## FilterQuery
//
// Filter individuals by various criteria:
//
//	// Filter by name
//	results, _ := q.Filter().ByName("John").Execute()
//
//	// Filter by multiple criteria (AND logic)
//	results, _ := q.Filter().
//		ByName("John").
//		BySex("M").
//		HasChildren().
//		Execute()
//
//	// Filter by date range
//	start := time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)
//	end := time.Date(1900, 12, 31, 23, 59, 59, 0, time.UTC)
//	results, _ := q.Filter().
//		ByBirthDate(start, end).
//		ByBirthPlace("New York").
//		Execute()
//
//	// Count matching individuals
//	count, _ := q.Filter().Living().HasSpouse().Count()
//
// ## FamilyQuery
//
// Query operations starting from a family:
//
//	husband, _ := q.Family("@F1@").Husband()
//	wife, _ := q.Family("@F1@").Wife()
//	children, _ := q.Family("@F1@").Children()
//	parents, _ := q.Family("@F1@").Parents()  // Husband and wife
//
//	marriageDate, _ := q.Family("@F1@").MarriageDate()
//	events, _ := q.Family("@F1@").Events()
//
// ## MultiIndividualQuery
//
// Query operations on multiple individuals:
//
//	// Find ancestors of all individuals (union)
//	ancestors, _ := q.Individuals("@I1@", "@I2@", "@I3@").Ancestors()
//
//	// Find common ancestors
//	common, _ := q.Individuals("@I1@", "@I2@").CommonAncestors()
//
//	// Union of results
//	results, _ := q.Individuals("@I1@", "@I2@").
//		Union(
//			func(iq *query.IndividualQuery) ([]*gedcom.IndividualRecord, error) {
//				return iq.Parents()
//			},
//			func(iq *query.IndividualQuery) ([]*gedcom.IndividualRecord, error) {
//				return iq.Siblings()
//			},
//		)
//
// ## GraphMetricsQuery
//
// Graph analytics and metrics:
//
//	metrics := q.Metrics()
//
//	// Node metrics
//	degree, _ := metrics.Degree("@I1@")
//	inDegree, _ := metrics.InDegree("@I1@")
//	outDegree, _ := metrics.OutDegree("@I1@")
//
//	// Graph metrics
//	diameter, _ := metrics.Diameter()
//	avgPathLength, _ := metrics.AveragePathLength()
//	avgDegree, _ := metrics.AverageDegree()
//	density, _ := metrics.Density()
//
//	// Centrality measures
//	centrality, _ := metrics.Centrality(query.CentralityDegree)
//	betweenness, _ := metrics.Centrality(query.CentralityBetweenness)
//	closeness, _ := metrics.Centrality(query.CentralityCloseness)
//
//	// Connectivity
//	connected, _ := metrics.IsConnected("@I1@", "@I2@")
//	components, _ := metrics.ConnectedComponents()
//
//	// Longest path
//	longestPath, _ := metrics.LongestPath()
//
// # Graph Algorithms
//
// The package also provides direct access to graph algorithms:
//
//	graph := q.Graph()
//
//	// Traversal
//	graph.BFS("@I1@", func(node GraphNode) bool {
//		fmt.Printf("Visited: %s\n", node.ID())
//		return true  // Continue
//	})
//
//	// Path finding
//	path, _ := graph.ShortestPath("@I1@", "@I2@")
//	allPaths, _ := graph.AllPaths("@I1@", "@I2@", 10)
//
//	// Ancestors
//	common, _ := graph.CommonAncestors("@I1@", "@I2@")
//	lca, _ := graph.LowestCommonAncestor("@I1@", "@I2@")
//
//	// Relationships
//	result, _ := graph.CalculateRelationship("@I1@", "@I2@")
//
// # Performance
//
// The query package is designed for performance:
//
//   - Graph is built once and reused for multiple queries
//   - Cached relationships (parents, children, spouses) for O(1) access
//   - Efficient algorithms (BFS for shortest path, DFS for all paths)
//   - Thread-safe operations
//
// # Thread Safety
//
// All query operations are thread-safe and can be called concurrently.
// The underlying graph uses RWMutex for concurrent read access.
//
// # Examples
//
// ## Example 1: Find All Ancestors
//
//	ancestors, _ := q.Individual("@I1@").
//		Ancestors().
//		MaxGenerations(5).
//		Execute()
//
//	for _, ancestor := range ancestors {
//		fmt.Printf("Ancestor: %s\n", ancestor.GetName())
//	}
//
// ## Example 2: Find Relationship
//
//	result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()
//	fmt.Printf("Relationship: %s\n", result.RelationshipType)
//	fmt.Printf("Degree: %d\n", result.Degree)
//
// ## Example 3: Find All Cousins
//
//	cousins, _ := q.Individual("@I1@").Cousins(1)  // 1st cousins
//	for _, cousin := range cousins {
//		fmt.Printf("Cousin: %s\n", cousin.GetName())
//	}
//
// ## Example 4: Find Path Between Two Individuals
//
//	paths, _ := q.Individual("@I1@").PathTo("@I2@").All()
//	for _, path := range paths {
//		fmt.Printf("Path length: %d\n", path.Length)
//		for _, node := range path.Nodes {
//			if indi, ok := node.(*query.IndividualNode); ok {
//				fmt.Printf("  -> %s\n", indi.Individual.GetName())
//			}
//		}
//	}
//
// ## Example 5: Complex Filtering
//
//	results, _ := q.Filter().
//		ByName("John").
//		BySex("M").
//		ByBirthDate(
//			time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC),
//			time.Date(1900, 12, 31, 23, 59, 59, 0, time.UTC),
//		).
//		HasChildren().
//		Execute()
//
// ## Example 6: Graph Metrics
//
//	metrics := q.Metrics()
//
//	// Find most connected individual
//	centrality, _ := metrics.Centrality(query.CentralityDegree)
//	maxDegree := 0.0
//	mostConnected := ""
//	for id, degree := range centrality {
//		if degree > maxDegree {
//			maxDegree = degree
//			mostConnected = id
//		}
//	}
//	fmt.Printf("Most connected: %s (degree: %.0f)\n", mostConnected, maxDegree)
//
//	// Check graph connectivity
//	components, _ := metrics.ConnectedComponents()
//	fmt.Printf("Number of connected components: %d\n", len(components))
package query
