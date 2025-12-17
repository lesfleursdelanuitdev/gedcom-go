package query

import (
	"testing"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

func TestGraphMetricsQuery_Degree(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	degree, err := query.Metrics().Degree("@I1@")
	if err != nil {
		t.Fatalf("Failed to get degree: %v", err)
	}

	if degree == 0 {
		t.Error("Expected degree > 0")
	}
}

func TestGraphMetricsQuery_InDegree(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	inDegree, err := query.Metrics().InDegree("@I2@")
	if err != nil {
		t.Fatalf("Failed to get in-degree: %v", err)
	}

	// I2 has FAMC edge (outgoing), so in-degree might be 0 or include reverse edges
	// This depends on implementation, but should not error
	_ = inDegree
}

func TestGraphMetricsQuery_IsConnected(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	connected, err := query.Metrics().IsConnected("@I1@", "@I2@")
	if err != nil {
		t.Fatalf("Failed to check connectivity: %v", err)
	}

	if !connected {
		t.Error("Expected @I1@ and @I2@ to be connected")
	}
}

func TestGraphMetricsQuery_NodeCount(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	count := query.Metrics().NodeCount()
	if count == 0 {
		t.Error("Expected node count > 0")
	}
}

func TestGraphMetricsQuery_EdgeCount(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	count := query.Metrics().EdgeCount()
	if count == 0 {
		t.Error("Expected edge count > 0")
	}
}

func TestGraphMetricsQuery_Centrality_Degree(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	centrality, err := query.Metrics().Centrality(CentralityDegree)
	if err != nil {
		t.Fatalf("Failed to calculate centrality: %v", err)
	}

	if len(centrality) == 0 {
		t.Error("Expected centrality values")
	}
}

func TestGraphMetricsQuery_ConnectedComponents(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(gedcom.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(gedcom.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := gedcom.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	components, err := query.Metrics().ConnectedComponents()
	if err != nil {
		t.Fatalf("Failed to get connected components: %v", err)
	}

	if len(components) == 0 {
		t.Error("Expected at least one connected component")
	}

	// I1 and I2 should be in the same component
	foundComponent := false
	for _, component := range components {
		hasI1 := false
		hasI2 := false
		for _, indi := range component {
			if indi.XrefID() == "@I1@" {
				hasI1 = true
			}
			if indi.XrefID() == "@I2@" {
				hasI2 = true
			}
		}
		if hasI1 && hasI2 {
			foundComponent = true
			break
		}
	}

	if !foundComponent {
		t.Error("Expected I1 and I2 to be in the same component")
	}
}

func TestGraphMetricsQuery_AverageDegree(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	avgDegree, err := query.Metrics().AverageDegree()
	if err != nil {
		t.Fatalf("Failed to calculate average degree: %v", err)
	}

	_ = avgDegree // Should not error
}

func TestGraphMetricsQuery_Density(t *testing.T) {
	tree := gedcom.NewGedcomTree()

	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	density, err := query.Metrics().Density()
	if err != nil {
		t.Fatalf("Failed to calculate density: %v", err)
	}

	// Density should be between 0 and 1
	if density < 0 || density > 1 {
		t.Errorf("Expected density between 0 and 1, got %f", density)
	}
}
