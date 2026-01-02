package duplicate

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
)

// generateLargeTreeForDuplicate creates a tree with n individuals for duplicate testing
func generateLargeTreeForDuplicate(n int) *types.GedcomTree {
	tree := types.NewGedcomTree()

	// Create individuals with some intentional duplicates
	for i := 1; i <= n; i++ {
		indiLine := types.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i))

		// Create some duplicates (every 100th person is similar to another)
		name := fmt.Sprintf("Person %d /Test/", i)
		if i%100 == 0 && i > 100 {
			// Make similar to previous person
			name = fmt.Sprintf("Person %d /Test/", i-1)
		}

		indiLine.AddChild(types.NewGedcomLine(1, "NAME", name, ""))

		birthYear := 1800 + (i % 200)
		birtLine := types.NewGedcomLine(1, "BIRT", "", "")
		birtLine.AddChild(types.NewGedcomLine(2, "DATE", fmt.Sprintf("%d", birthYear), ""))
		indiLine.AddChild(birtLine)

		sex := "M"
		if i%2 == 0 {
			sex = "F"
		}
		indiLine.AddChild(types.NewGedcomLine(1, "SEX", sex, ""))

		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	return tree
}

// measureMemory returns current memory usage
func measureMemory() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.Alloc
}

