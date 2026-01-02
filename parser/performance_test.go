package parser

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
)

// generateLargeGEDCOMFile generates a GEDCOM file with n individuals
func generateLargeGEDCOMFile(filename string, n int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	file.WriteString("0 HEAD\n")
	file.WriteString("1 SOUR Test Generator\n")
	file.WriteString("1 GEDC\n")
	file.WriteString("2 VERS 5.5.1\n")
	file.WriteString("0 @SUBM@ SUBM\n")
	file.WriteString("1 NAME Test\n")

	// Track relationships
	childToFamily := make(map[int]int)
	familyID := 1
	indiID := 1

	// Create families
	for indiID < n {
		numChildren := 2
		if indiID%10 == 0 {
			numChildren = 1
		} else if indiID%5 == 0 {
			numChildren = 3
		}

		if indiID+numChildren+1 >= n {
			numChildren = n - indiID - 1
			if numChildren <= 0 {
				break
			}
		}

		// Family record
		file.WriteString(fmt.Sprintf("0 @F%d@ FAM\n", familyID))

		// Husband
		if indiID < n {
			file.WriteString(fmt.Sprintf("1 HUSB @I%d@\n", indiID))
			indiID++
		}

		// Wife
		if indiID < n {
			file.WriteString(fmt.Sprintf("1 WIFE @I%d@\n", indiID))
			indiID++
		}

		// Children
		for i := 0; i < numChildren && indiID < n; i++ {
			file.WriteString(fmt.Sprintf("1 CHIL @I%d@\n", indiID))
			childToFamily[indiID] = familyID
			indiID++
		}

		familyID++
	}

	// Create individuals
	for i := 1; i <= n; i++ {
		file.WriteString(fmt.Sprintf("0 @I%d@ INDI\n", i))
		file.WriteString(fmt.Sprintf("1 NAME Person %d /Test/\n", i))

		birthYear := 1800 + (i % 200)
		file.WriteString("1 BIRT\n")
		file.WriteString(fmt.Sprintf("2 DATE %d\n", birthYear))

		sex := "M"
		if i%2 == 0 {
			sex = "F"
		}
		file.WriteString(fmt.Sprintf("1 SEX %s\n", sex))

		if famID, ok := childToFamily[i]; ok {
			file.WriteString(fmt.Sprintf("1 FAMC @F%d@\n", famID))
		}
	}

	file.WriteString("0 TRLR\n")
	return nil
}

// measureMemory returns current memory usage
func measureMemory() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.Alloc
}

