package query

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// generateVeryLargeTree creates a tree with n individuals efficiently
// Uses a more realistic family structure with varying family sizes
func generateVeryLargeTree(n int) *types.GedcomTree {
	tree := types.NewGedcomTree()

	// Pre-allocate slices for better performance
	individuals := make([]*types.IndividualRecord, 0, n)
	families := make([]*types.FamilyRecord, 0, n/2)

	// Track relationships
	childToFamily := make(map[int]int, n)
	familyID := 1
	indiID := 1

	// Create families with 1-3 children each (realistic distribution)
	for indiID < n {
		// Determine family size (1-3 children, weighted toward 2)
		numChildren := 2
		if indiID%10 == 0 {
			numChildren = 1 // 10% have 1 child
		} else if indiID%5 == 0 {
			numChildren = 3 // 20% have 3 children
		}

		// Check if we have enough individuals left
		if indiID+numChildren+1 >= n {
			numChildren = n - indiID - 1
			if numChildren <= 0 {
				break
			}
		}

		// Create family
		famLine := types.NewGedcomLine(0, "FAM", "", fmt.Sprintf("@F%d@", familyID))

		// Husband
		if indiID < n {
			famLine.AddChild(types.NewGedcomLine(1, "HUSB", fmt.Sprintf("@I%d@", indiID), ""))
			indiID++
		}

		// Wife
		if indiID < n {
			famLine.AddChild(types.NewGedcomLine(1, "WIFE", fmt.Sprintf("@I%d@", indiID), ""))
			indiID++
		}

		// Children
		for i := 0; i < numChildren && indiID < n; i++ {
			famLine.AddChild(types.NewGedcomLine(1, "CHIL", fmt.Sprintf("@I%d@", indiID), ""))
			childToFamily[indiID] = familyID
			indiID++
		}

		fam := types.NewFamilyRecord(famLine)
		families = append(families, fam)
		familyID++
	}

	// Create all individuals
	for i := 1; i <= n; i++ {
		indiLine := types.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i))

		// Add name
		indiLine.AddChild(types.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", i), ""))

		// Add birth date (distributed across years 1800-2000)
		birthYear := 1800 + (i % 200)
		birtLine := types.NewGedcomLine(1, "BIRT", "", "")
		birtLine.AddChild(types.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", birthYear), ""))
		indiLine.AddChild(birtLine)

		// Add sex (alternating)
		sex := "M"
		if i%2 == 0 {
			sex = "F"
		}
		indiLine.AddChild(types.NewGedcomLine(1, "SEX", sex, ""))

		// Add FAMC if child
		if famID, ok := childToFamily[i]; ok {
			indiLine.AddChild(types.NewGedcomLine(1, "FAMC", fmt.Sprintf("@F%d@", famID), ""))
		}

		indi := types.NewIndividualRecord(indiLine)
		individuals = append(individuals, indi)
	}

	// Add all records to tree
	for _, fam := range families {
		tree.AddRecord(fam)
	}
	for _, indi := range individuals {
		tree.AddRecord(indi)
	}

	return tree
}

// PerformanceMetrics holds performance test results
type PerformanceMetrics struct {
	Operation    string
	DatasetSize  int
	Duration     time.Duration
	MemoryBefore uint64
	MemoryAfter  uint64
	MemoryPeak   uint64
	Throughput   float64 // operations per second
}

// measureMemory returns current memory usage in bytes
func measureMemory() uint64 {
	var m runtime.MemStats
	runtime.GC() // Force GC before measurement
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// runPerformanceTest runs a test function and collects metrics
func runPerformanceTest(operation string, datasetSize int, fn func()) PerformanceMetrics {
	before := measureMemory()
	start := time.Now()

	fn()

	duration := time.Since(start)
	after := measureMemory()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	peak := m.TotalAlloc

	throughput := float64(datasetSize) / duration.Seconds()

	return PerformanceMetrics{
		Operation:    operation,
		DatasetSize:  datasetSize,
		Duration:     duration,
		MemoryBefore: before,
		MemoryAfter:  after,
		MemoryPeak:   peak,
		Throughput:   throughput,
	}
}

