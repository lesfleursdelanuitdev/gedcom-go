package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
)

// TestIntegration_RealData_BasicQueries tests basic query operations on real GEDCOM files
func TestIntegration_RealData_BasicQueries(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
		"tree1.ged",
	}
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "royal92.ged")
		// pres2020.ged is large, test separately
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse the file
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			// Build graph
			graph, err := BuildGraph(tree)
			if err != nil {
				t.Fatalf("Failed to build graph for %s: %v", filename, err)
			}

			// Verify graph was built
			if graph == nil {
				t.Fatal("Graph is nil")
			}

			// Test basic query operations
			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Test getting all individuals
			allIndividuals := qb.AllIndividuals()
			results, err := allIndividuals.Execute()
			if err != nil {
				t.Fatalf("Failed to get all individuals for %s: %v", filename, err)
			}

			if len(results) == 0 {
				t.Errorf("Expected at least 1 individual in %s, got 0", filename)
			}

			// Test getting a specific individual (use first one)
			if len(results) > 0 {
				firstXref := results[0].XrefID()
				individualQuery := qb.Individual(firstXref)
				if individualQuery == nil {
					t.Errorf("Individual query returned nil for %s", firstXref)
				}
			}

			// Test getting all families
			families := qb.Families()
			familyResults, err := families.All()
			if err != nil {
				t.Fatalf("Failed to get families for %s: %v", filename, err)
			}
			_ = familyResults // Just verify it doesn't panic

			// Test filter query
			filterQuery := qb.Filter()
			if filterQuery == nil {
				t.Error("Filter query returned nil")
			}
		})
	}
}

// TestIntegration_RealData_GraphBuilding tests graph building on real files
func TestIntegration_RealData_GraphBuilding(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse the file
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			// Build graph
			graph, err := BuildGraph(tree)
			if err != nil {
				t.Fatalf("Failed to build graph for %s: %v", filename, err)
			}

			// Verify graph properties
			nodeCount := graph.NodeCount()
			if nodeCount == 0 {
				t.Errorf("Expected at least 1 node in %s, got 0", filename)
			}

			edgeCount := graph.EdgeCount()
			_ = edgeCount // Just verify it doesn't panic

			// Test getting all nodes
			allNodes := graph.GetAllNodes()
			if len(allNodes) == 0 {
				t.Errorf("Expected at least 1 node in GetAllNodes for %s", filename)
			}

			// Test getting all individuals
			allIndividuals := graph.GetAllIndividuals()
			if len(allIndividuals) == 0 {
				t.Errorf("Expected at least 1 individual in GetAllIndividuals for %s", filename)
			}

			// Test getting all families
			allFamilies := graph.GetAllFamilies()
			_ = allFamilies // Just verify it doesn't panic
		})
	}
}

// TestIntegration_RealData_RelationshipQueries tests relationship queries on real files
func TestIntegration_RealData_RelationshipQueries(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Get all individuals
			allIndividuals, err := qb.AllIndividuals().Execute()
			if err != nil {
				t.Fatalf("Failed to get all individuals for %s: %v", filename, err)
			}

			if len(allIndividuals) == 0 {
				t.Skipf("No individuals found in %s", filename)
			}

			// Test relationship queries on first few individuals
			testCount := 5
			if len(allIndividuals) < testCount {
				testCount = len(allIndividuals)
			}

			for i := 0; i < testCount; i++ {
				xref := allIndividuals[i].XrefID()
				iq := qb.Individual(xref)

				// Test Parents
				parents, err := iq.Parents()
				if err != nil {
					t.Errorf("Failed to get parents for %s in %s: %v", xref, filename, err)
				}
				_ = parents // Just verify it doesn't panic

				// Test Children
				children, err := iq.Children()
				if err != nil {
					t.Errorf("Failed to get children for %s in %s: %v", xref, filename, err)
				}
				_ = children // Just verify it doesn't panic

				// Test Siblings
				siblings, err := iq.Siblings()
				if err != nil {
					t.Errorf("Failed to get siblings for %s in %s: %v", xref, filename, err)
				}
				_ = siblings // Just verify it doesn't panic

				// Test Spouses
				spouses, err := iq.Spouses()
				if err != nil {
					t.Errorf("Failed to get spouses for %s in %s: %v", xref, filename, err)
				}
				_ = spouses // Just verify it doesn't panic

				// Test Ancestors
				ancestors, err := iq.Ancestors().Execute()
				if err != nil {
					t.Errorf("Failed to get ancestors for %s in %s: %v", xref, filename, err)
				}
				_ = ancestors // Just verify it doesn't panic

				// Test Descendants
				descendants, err := iq.Descendants().Execute()
				if err != nil {
					t.Errorf("Failed to get descendants for %s in %s: %v", xref, filename, err)
				}
				_ = descendants // Just verify it doesn't panic
			}
		})
	}
}

// TestIntegration_RealData_GraphRelationships tests graph relationship helper methods
func TestIntegration_RealData_GraphRelationships(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
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

			// Get all individuals
			allIndividuals := graph.GetAllIndividuals()
			if len(allIndividuals) == 0 {
				t.Skipf("No individuals found in %s", filename)
			}

			// Test relationship helper methods on first few individuals
			testCount := 5
			if len(allIndividuals) < testCount {
				testCount = len(allIndividuals)
			}

			count := 0
			for xref := range allIndividuals {
				if count >= testCount {
					break
				}

				// Test GetChildren
				children, err := graph.GetChildren(xref)
				if err != nil {
					t.Errorf("Failed to get children for %s in %s: %v", xref, filename, err)
				}
				_ = children

				// Test GetParents
				parents, err := graph.GetParents(xref)
				if err != nil {
					t.Errorf("Failed to get parents for %s in %s: %v", xref, filename, err)
				}
				_ = parents

				// Test GetSiblings
				siblings, err := graph.GetSiblings(xref)
				if err != nil {
					t.Errorf("Failed to get siblings for %s in %s: %v", xref, filename, err)
				}
				_ = siblings

				// Test GetSpouses
				spouses, err := graph.GetSpouses(xref)
				if err != nil {
					t.Errorf("Failed to get spouses for %s in %s: %v", xref, filename, err)
				}
				_ = spouses

				count++
			}

			// Test family relationship methods
			allFamilies := graph.GetAllFamilies()
			if len(allFamilies) > 0 {
				count = 0
				for xref, _ := range allFamilies {
					if count >= 3 {
						break
					}

					// Test GetFamilyHusband
					husband, err := graph.GetFamilyHusband(xref)
					if err != nil {
						// Family might not have husband, that's okay
						_ = err
					}
					_ = husband

					// Test GetFamilyWife
					wife, err := graph.GetFamilyWife(xref)
					if err != nil {
						// Family might not have wife, that's okay
						_ = err
					}
					_ = wife

					// Test GetFamilyChildren
					children, err := graph.GetFamilyChildren(xref)
					if err != nil {
						t.Errorf("Failed to get family children for %s in %s: %v", xref, filename, err)
					}
					_ = children

					count++
				}
			}
		})
	}
}

// TestIntegration_RealData_FilterQueries tests filter queries on real data
func TestIntegration_RealData_FilterQueries(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Test filter by name (partial match)
			fq := qb.Filter()
			results, err := fq.ByName("").Execute()
			if err != nil {
				t.Fatalf("Failed to execute filter by name for %s: %v", filename, err)
			}
			_ = results // Just verify it doesn't panic

			// Test filter by sex
			maleResults, err := fq.BySex("M").Execute()
			if err != nil {
				t.Fatalf("Failed to execute filter by sex for %s: %v", filename, err)
			}
			_ = maleResults

			femaleResults, err := qb.Filter().BySex("F").Execute()
			if err != nil {
				t.Fatalf("Failed to execute filter by sex (F) for %s: %v", filename, err)
			}
			_ = femaleResults

			// Test HasChildren filter (tests filterByBool)
			hasChildrenResults, err := qb.Filter().HasChildren().Execute()
			if err != nil {
				t.Fatalf("Failed to execute HasChildren filter for %s: %v", filename, err)
			}
			_ = hasChildrenResults

			// Test HasSpouse filter (tests filterByBool)
			hasSpouseResults, err := qb.Filter().HasSpouse().Execute()
			if err != nil {
				t.Fatalf("Failed to execute HasSpouse filter for %s: %v", filename, err)
			}
			_ = hasSpouseResults
		})
	}
}

// TestIntegration_RealData_PathFinding tests path finding on real data
func TestIntegration_RealData_PathFinding(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Get all individuals
			allIndividuals, err := qb.AllIndividuals().Execute()
			if err != nil {
				t.Fatalf("Failed to get all individuals for %s: %v", filename, err)
			}

			if len(allIndividuals) < 2 {
				t.Skipf("Need at least 2 individuals for path finding in %s", filename)
			}

			// Test path finding between first two individuals
			fromXref := allIndividuals[0].XrefID()
			toXref := allIndividuals[1].XrefID()

			// Test PathTo
			pathQuery := qb.Individual(fromXref).PathTo(toXref)
			if pathQuery == nil {
				t.Error("PathTo returned nil")
			}

			// Test Shortest path
			shortestPath, err := pathQuery.Shortest()
			if err != nil {
				// Path might not exist, that's okay
				_ = err
			}
			_ = shortestPath

			// Test All paths
			allPaths, err := pathQuery.All()
			if err != nil {
				// Path might not exist, that's okay
				_ = err
			}
			_ = allPaths

			// Test Count
			count, err := pathQuery.Count()
			if err != nil {
				// Path might not exist, that's okay
				_ = err
			}
			_ = count
		})
	}
}

// TestIntegration_RealData_RelationshipCalculation tests relationship calculation on real data
func TestIntegration_RealData_RelationshipCalculation(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Get all individuals
			allIndividuals, err := qb.AllIndividuals().Execute()
			if err != nil {
				t.Fatalf("Failed to get all individuals for %s: %v", filename, err)
			}

			if len(allIndividuals) < 2 {
				t.Skipf("Need at least 2 individuals for relationship calculation in %s", filename)
			}

			// Test relationship calculation between first few pairs
			testCount := 5
			if len(allIndividuals) < testCount {
				testCount = len(allIndividuals)
			}

			for i := 0; i < testCount && i+1 < len(allIndividuals); i++ {
				fromXref := allIndividuals[i].XrefID()
				toXref := allIndividuals[i+1].XrefID()

				// Test RelationshipTo
				relationshipQuery := qb.Individual(fromXref).RelationshipTo(toXref)
				if relationshipQuery == nil {
					t.Error("RelationshipTo returned nil")
				}

				// Test Execute
				result, err := relationshipQuery.Execute()
				if err != nil {
					// Relationship might not exist, that's okay
					_ = err
				}
				if result != nil {
					// Test GetRelationshipType
					relType := result.GetRelationshipType()
					_ = relType

					// Test IsBloodRelation
					isBlood := result.IsBloodRelation()
					_ = isBlood

					// Test IsMaritalRelation
					isMarital := result.IsMaritalRelation()
					_ = isMarital
				}
			}
		})
	}
}

// TestIntegration_RealData_CollectionQueries tests collection queries on real data
func TestIntegration_RealData_CollectionQueries(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Test Names collection
			names := qb.Names()
			allNames, err := names.All()
			if err != nil {
				t.Fatalf("Failed to get all names for %s: %v", filename, err)
			}
			_ = allNames

			uniqueNames, err := names.Unique().Execute()
			if err != nil {
				t.Fatalf("Failed to get unique names for %s: %v", filename, err)
			}
			_ = uniqueNames

			nameCount, err := names.Count()
			if err != nil {
				t.Fatalf("Failed to count names for %s: %v", filename, err)
			}
			_ = nameCount

			// Test Places collection
			places := qb.Places()
			allPlaces, err := places.All()
			if err != nil {
				t.Fatalf("Failed to get all places for %s: %v", filename, err)
			}
			_ = allPlaces

			uniquePlaces, err := places.Unique().Execute()
			if err != nil {
				t.Fatalf("Failed to get unique places for %s: %v", filename, err)
			}
			_ = uniquePlaces

			// Test Events collection
			events := qb.Events()
			allEvents, err := events.All()
			if err != nil {
				t.Fatalf("Failed to get all events for %s: %v", filename, err)
			}
			_ = allEvents

			uniqueEvents, err := events.Unique().Execute()
			if err != nil {
				t.Fatalf("Failed to get unique events for %s: %v", filename, err)
			}
			_ = uniqueEvents
		})
	}
}

// TestIntegration_RealData_NotesQuery tests notes query on real data
func TestIntegration_RealData_NotesQuery(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Get all individuals
			allIndividuals, err := qb.AllIndividuals().Execute()
			if err != nil {
				t.Fatalf("Failed to get all individuals for %s: %v", filename, err)
			}

			if len(allIndividuals) == 0 {
				t.Skipf("No individuals found in %s", filename)
			}

			// Test GetAllNotes on first few individuals
			testCount := 10
			if len(allIndividuals) < testCount {
				testCount = len(allIndividuals)
			}

			for i := 0; i < testCount; i++ {
				xref := allIndividuals[i].XrefID()
				iq := qb.Individual(xref)

				// Test GetAllNotes (tests 0% coverage function)
				notes, err := iq.GetAllNotes()
				if err != nil {
					// Individual might not exist, that's okay
					_ = err
				}
				_ = notes // Just verify it doesn't panic
			}

			// Test GetAllNotes for families
			allFamilies, err := qb.Families().All()
			if err != nil {
				t.Fatalf("Failed to get families for %s: %v", filename, err)
			}

			if len(allFamilies) > 0 {
				testCount = 5
				if len(allFamilies) < testCount {
					testCount = len(allFamilies)
				}

				for i := 0; i < testCount; i++ {
					xref := allFamilies[i].XrefID()
					fq := qb.Family(xref)

					// Test GetAllNotes for family
					notes, err := fq.GetAllNotes()
					if err != nil {
						// Family might not exist, that's okay
						_ = err
					}
					_ = notes // Just verify it doesn't panic
				}
			}
		})
	}
}

// TestIntegration_RealData_GraphMetrics tests graph metrics on real data
func TestIntegration_RealData_GraphMetrics(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
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

			// Test Metrics() method (tests 0% coverage function)
			metricsQuery := graph.Metrics()
			if metricsQuery == nil {
				t.Error("Metrics() returned nil")
			}

			// Get all individuals
			allIndividuals := graph.GetAllIndividuals()
			if len(allIndividuals) == 0 {
				t.Skipf("No individuals found in %s", filename)
			}

			// Test metrics on first few individuals
			testCount := 5
			if len(allIndividuals) < testCount {
				testCount = len(allIndividuals)
			}

			count := 0
			for xref := range allIndividuals {
				if count >= testCount {
					break
				}

				// Test Degree
				degree, err := metricsQuery.Degree(xref)
				if err != nil {
					t.Errorf("Failed to get degree for %s in %s: %v", xref, filename, err)
				}
				_ = degree

				// Test InDegree
				inDegree, err := metricsQuery.InDegree(xref)
				if err != nil {
					t.Errorf("Failed to get in-degree for %s in %s: %v", xref, filename, err)
				}
				_ = inDegree

				// Test OutDegree (tests 0% coverage function)
				outDegree, err := metricsQuery.OutDegree(xref)
				if err != nil {
					t.Errorf("Failed to get out-degree for %s in %s: %v", xref, filename, err)
				}
				_ = outDegree

				count++
			}

			// Test Centrality (degree)
			centrality, err := metricsQuery.Centrality(CentralityDegree)
			if err != nil {
				t.Errorf("Failed to calculate degree centrality for %s: %v", filename, err)
			}
			_ = centrality

			// Test ConnectedComponents
			components, err := metricsQuery.ConnectedComponents()
			if err != nil {
				t.Errorf("Failed to get connected components for %s: %v", filename, err)
			}
			_ = components
		})
	}
}

// TestIntegration_RealData_CollateralRelationships tests collateral relationship queries
func TestIntegration_RealData_CollateralRelationships(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Get all individuals
			allIndividuals, err := qb.AllIndividuals().Execute()
			if err != nil {
				t.Fatalf("Failed to get all individuals for %s: %v", filename, err)
			}

			if len(allIndividuals) == 0 {
				t.Skipf("No individuals found in %s", filename)
			}

			// Test collateral relationship queries on first few individuals
			testCount := 5
			if len(allIndividuals) < testCount {
				testCount = len(allIndividuals)
			}

			for i := 0; i < testCount; i++ {
				xref := allIndividuals[i].XrefID()
				iq := qb.Individual(xref)

				// Test Cousins (tests getCollateralRelationshipType)
				cousins, err := iq.Cousins(1)
				if err != nil {
					t.Errorf("Failed to get cousins for %s in %s: %v", xref, filename, err)
				}
				_ = cousins

				// Test Uncles
				uncles, err := iq.Uncles()
				if err != nil {
					t.Errorf("Failed to get uncles for %s in %s: %v", xref, filename, err)
				}
				_ = uncles

				// Test Nephews
				nephews, err := iq.Nephews()
				if err != nil {
					t.Errorf("Failed to get nephews for %s in %s: %v", xref, filename, err)
				}
				_ = nephews

				// Test Grandparents (tests getAncestralRelationshipType)
				grandparents, err := iq.Grandparents()
				if err != nil {
					t.Errorf("Failed to get grandparents for %s in %s: %v", xref, filename, err)
				}
				_ = grandparents

				// Test Grandchildren
				grandchildren, err := iq.Grandchildren()
				if err != nil {
					t.Errorf("Failed to get grandchildren for %s in %s: %v", xref, filename, err)
				}
				_ = grandchildren
			}
		})
	}
}

// TestIntegration_RealData_AncestorDescendantEdgeCases tests edge cases for ancestor/descendant queries
func TestIntegration_RealData_AncestorDescendantEdgeCases(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}
	
	// Only test larger files if not in short mode
	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder for %s: %v", filename, err)
			}

			// Get all individuals
			allIndividuals, err := qb.AllIndividuals().Execute()
			if err != nil {
				t.Fatalf("Failed to get all individuals for %s: %v", filename, err)
			}

			if len(allIndividuals) == 0 {
				t.Skipf("No individuals found in %s", filename)
			}

			// Test on first individual
			xref := allIndividuals[0].XrefID()
			iq := qb.Individual(xref)

			// Test Ancestors with MaxGenerations (tests findAncestorsWithDepth)
			ancestors, err := iq.Ancestors().MaxGenerations(2).Execute()
			if err != nil {
				t.Errorf("Failed to get ancestors with max generations for %s in %s: %v", xref, filename, err)
			}
			_ = ancestors

			// Test Ancestors Count (tests Count method)
			ancestorCount, err := iq.Ancestors().Count()
			if err != nil {
				t.Errorf("Failed to count ancestors for %s in %s: %v", xref, filename, err)
			}
			_ = ancestorCount

			// Test Ancestors Exists (tests Exists method)
			ancestorsExist, err := iq.Ancestors().Exists()
			if err != nil {
				t.Errorf("Failed to check ancestors existence for %s in %s: %v", xref, filename, err)
			}
			_ = ancestorsExist

			// Test Descendants Count
			descendantCount, err := iq.Descendants().Count()
			if err != nil {
				t.Errorf("Failed to count descendants for %s in %s: %v", xref, filename, err)
			}
			_ = descendantCount

			// Test Descendants Exists
			descendantsExist, err := iq.Descendants().Exists()
			if err != nil {
				t.Errorf("Failed to check descendants existence for %s in %s: %v", xref, filename, err)
			}
			_ = descendantsExist

			// Test Ancestors ExecuteWithPaths (tests ExecuteWithPaths)
			ancestorsWithPaths, err := iq.Ancestors().ExecuteWithPaths()
			if err != nil {
				t.Errorf("Failed to get ancestors with paths for %s in %s: %v", xref, filename, err)
			}
			_ = ancestorsWithPaths
		})
	}
}

