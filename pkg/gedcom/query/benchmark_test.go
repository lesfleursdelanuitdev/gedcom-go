package query

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// generateLargeTree creates a tree with n individuals in a balanced binary tree structure
func generateLargeTree(n int) *gedcom.GedcomTree {
	tree := gedcom.NewGedcomTree()

	// Track which family each child belongs to
	childToFamily := make(map[int]int)

	// Create families and relationships in a binary tree structure
	// Each family has 2 children, creating a balanced tree
	familyID := 1
	for i := 0; i < n-1; i += 2 {
		if i+1 >= n {
			break
		}

		// Create family
		famLine := gedcom.NewGedcomLine(0, "FAM", "", fmt.Sprintf("@F%d@", familyID))
		famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", fmt.Sprintf("@I%d@", i+1), ""))
		if i+2 < n {
			famLine.AddChild(gedcom.NewGedcomLine(1, "WIFE", fmt.Sprintf("@I%d@", i+2), ""))
		}
		if i+3 < n {
			famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", fmt.Sprintf("@I%d@", i+3), ""))
			childToFamily[i+3] = familyID
		}
		if i+4 < n {
			famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", fmt.Sprintf("@I%d@", i+4), ""))
			childToFamily[i+4] = familyID
		}

		fam := gedcom.NewFamilyRecord(famLine)
		tree.AddRecord(fam)

		familyID++
	}

	// Create individuals with FAMC already set
	for i := 0; i < n; i++ {
		indiLine := gedcom.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i+1))
		if famID, ok := childToFamily[i]; ok {
			indiLine.AddChild(gedcom.NewGedcomLine(1, "FAMC", fmt.Sprintf("@F%d@", famID), ""))
		}
		indi := gedcom.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	return tree
}

// generateLinearTree creates a linear tree (chain) with n individuals
func generateLinearTree(n int) *gedcom.GedcomTree {
	tree := gedcom.NewGedcomTree()

	// Create families first
	for i := 0; i < n-1; i++ {
		famLine := gedcom.NewGedcomLine(0, "FAM", "", fmt.Sprintf("@F%d@", i+1))
		famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", fmt.Sprintf("@I%d@", i+1), ""))
		famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", fmt.Sprintf("@I%d@", i+2), ""))
		fam := gedcom.NewFamilyRecord(famLine)
		tree.AddRecord(fam)
	}

	// Create individuals with FAMC already set
	for i := 0; i < n; i++ {
		indiLine := gedcom.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i+1))
		if i > 0 {
			// Add FAMC to all children (except first)
			indiLine.AddChild(gedcom.NewGedcomLine(1, "FAMC", fmt.Sprintf("@F%d@", i), ""))
		}
		indi := gedcom.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	return tree
}

// BenchmarkGraphConstruction benchmarks graph construction for various sizes
func BenchmarkGraphConstruction_100(b *testing.B) {
	tree := generateLargeTree(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildGraph(tree)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGraphConstruction_1000(b *testing.B) {
	tree := generateLargeTree(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildGraph(tree)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGraphConstruction_10000(b *testing.B) {
	tree := generateLargeTree(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildGraph(tree)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIndividualQuery benchmarks individual queries
func BenchmarkIndividualQuery_Parents(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individual("@I500@").Parents()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndividualQuery_Ancestors(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individual("@I500@").Ancestors().MaxGenerations(10).Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndividualQuery_Descendants(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individual("@I1@").Descendants().MaxGenerations(10).Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPathFinding benchmarks path finding algorithms
func BenchmarkShortestPath_Linear_100(b *testing.B) {
	tree := generateLinearTree(100)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individual("@I1@").PathTo("@I100@").Shortest()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkShortestPath_Linear_1000(b *testing.B) {
	tree := generateLinearTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individual("@I1@").PathTo("@I1000@").Shortest()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkShortestPath_Tree_1000(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individual("@I1@").PathTo("@I500@").Shortest()
		if err != nil {
			// Path might not exist, that's okay for benchmark
		}
	}
}

// BenchmarkRelationshipCalculation benchmarks relationship calculations
func BenchmarkRelationshipCalculation(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individual("@I1@").RelationshipTo("@I500@").Execute()
		if err != nil {
			// Relationship might not exist, that's okay for benchmark
		}
	}
}

// BenchmarkMetrics benchmarks graph metrics calculations
func BenchmarkMetrics_Degree(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Metrics().Degree("@I500@")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMetrics_CentralityDegree(b *testing.B) {
	tree := generateLargeTree(500) // Smaller for centrality (O(V²))
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Metrics().Centrality(CentralityDegree)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMetrics_ConnectedComponents(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Metrics().ConnectedComponents()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMetrics_Diameter(b *testing.B) {
	tree := generateLargeTree(200) // Smaller for diameter (O(V²))
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Metrics().Diameter()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFilter benchmarks filtering operations
func BenchmarkFilter_ByName(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Filter().ByName("John").Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFilter_Complex(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Filter().
			BySex("M").
			HasChildren().
			Execute()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMultiIndividualQuery benchmarks multi-individual operations
func BenchmarkMultiIndividualQuery_Ancestors(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individuals("@I100@", "@I200@", "@I300@").Ancestors()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMultiIndividualQuery_CommonAncestors(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := query.Individuals("@I100@", "@I200@").CommonAncestors()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentAccess benchmarks concurrent query access
func BenchmarkConcurrentAccess(b *testing.B) {
	tree := generateLargeTree(1000)
	query, err := NewQuery(tree)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := query.Individual("@I500@").Parents()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
