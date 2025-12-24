//go:build external_comparison
// +build external_comparison

package performance_comparison

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cacack/gedcom-go/decoder"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/internal/parser"
)

// TestRoyal92Comparison compares the fastest internal parser with cacack parser
func TestRoyal92Comparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping royal92 comparison in short mode")
	}

	filePath := "../../testdata/royal92.ged"
	if !fileExists(filePath) {
		t.Skipf("Test file not found: %s", filePath)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Royal92.ged Performance Comparison")
	fmt.Println("Fastest Internal Parser (ParallelHierarchicalParser) vs cacack/gedcom-go")
	fmt.Println(strings.Repeat("=", 80))

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	fileSize := fileInfo.Size()
	fmt.Printf("\nFile: %s\n", filePath)
	fmt.Printf("File size: %.2f KB (%d bytes, %d lines)\n", float64(fileSize)/1024, fileSize, 30683)

	const iterations = 10
	results := make(map[string]time.Duration)

	// Test ParallelHierarchicalParser (fastest internal parser)
	fmt.Println("\n[ParallelHierarchicalParser - Fastest Internal Parser]")
	var parallelTime time.Duration
	for i := 0; i < iterations; i++ {
		start := time.Now()
		p := parser.NewParallelHierarchicalParser()
		_, err := p.Parse(filePath)
		if err != nil {
			t.Fatalf("ParallelHierarchicalParser failed: %v", err)
		}
		duration := time.Since(start)
		parallelTime += duration
	}
	parallelAvg := parallelTime / iterations
	results["ParallelHierarchicalParser"] = parallelAvg
	fmt.Printf("  Average time (%d iterations): %v\n", iterations, parallelAvg)
	fmt.Printf("  Throughput: %.2f KB/s\n", float64(fileSize)/1024/parallelAvg.Seconds())
	fmt.Printf("  Throughput: %.2f MB/s\n", float64(fileSize)/1024/1024/parallelAvg.Seconds())

	// Test cacack parser
	fmt.Println("\n[cacack/gedcom-go Parser]")
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Read file into memory for fair comparison
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var cacackTime time.Duration
	for i := 0; i < iterations; i++ {
		start := time.Now()
		reader := bytes.NewReader(data)
		_, err := decoder.Decode(reader)
		if err != nil {
			t.Fatalf("cacack parser failed: %v", err)
		}
		duration := time.Since(start)
		cacackTime += duration
	}
	cacackAvg := cacackTime / iterations
	results["cacack"] = cacackAvg
	fmt.Printf("  Average time (%d iterations): %v\n", iterations, cacackAvg)
	fmt.Printf("  Throughput: %.2f KB/s\n", float64(fileSize)/1024/cacackAvg.Seconds())
	fmt.Printf("  Throughput: %.2f MB/s\n", float64(fileSize)/1024/1024/cacackAvg.Seconds())

	// Detailed comparison
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Detailed Comparison")
	fmt.Println(strings.Repeat("=", 80))

	parallelTimeMs := parallelAvg.Seconds() * 1000
	cacackTimeMs := cacackAvg.Seconds() * 1000

	fmt.Printf("\nTime Difference:\n")
	fmt.Printf("  ParallelHierarchicalParser: %.2f ms\n", parallelTimeMs)
	fmt.Printf("  cacack parser:              %.2f ms\n", cacackTimeMs)
	fmt.Printf("  Difference:                 %.2f ms (%.1f%%)\n",
		parallelTimeMs-cacackTimeMs,
		((parallelTimeMs-cacackTimeMs)/cacackTimeMs)*100)

	ratio := float64(parallelAvg) / float64(cacackAvg)
	if ratio < 1.0 {
		fmt.Printf("\n✅ ParallelHierarchicalParser is %.2fx FASTER than cacack parser\n", 1.0/ratio)
	} else {
		fmt.Printf("\n⚠️  ParallelHierarchicalParser is %.2fx SLOWER than cacack parser\n", ratio)
	}

	// Throughput comparison
	parallelThroughput := float64(fileSize) / 1024 / parallelAvg.Seconds()
	cacackThroughput := float64(fileSize) / 1024 / cacackAvg.Seconds()
	fmt.Printf("\nThroughput Comparison:\n")
	fmt.Printf("  ParallelHierarchicalParser: %.2f KB/s\n", parallelThroughput)
	fmt.Printf("  cacack parser:              %.2f KB/s\n", cacackThroughput)
	fmt.Printf("  Difference:                 %.2f KB/s\n", parallelThroughput-cacackThroughput)

	// Performance summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Performance Summary")
	fmt.Println(strings.Repeat("=", 80))
	if ratio < 1.0 {
		fmt.Printf("✅ Your ParallelHierarchicalParser outperforms cacack parser by %.1f%%\n", (1.0-ratio)*100)
	} else if ratio < 1.1 {
		fmt.Printf("✅ Your ParallelHierarchicalParser is competitive (within 10%%)\n")
	} else {
		fmt.Printf("⚠️  Your ParallelHierarchicalParser is %.1f%% slower than cacack parser\n", (ratio-1.0)*100)
		fmt.Printf("   Consider implementing ParseLineFast optimization\n")
	}
}

// BenchmarkRoyal92Comparison benchmarks both parsers for royal92.ged
func BenchmarkRoyal92Comparison(b *testing.B) {
	filePath := "../../testdata/royal92.ged"
	if !fileExists(filePath) {
		b.Skipf("Test file not found: %s", filePath)
	}

	// Read file into memory for cacack parser
	data, err := os.ReadFile(filePath)
	if err != nil {
		b.Fatalf("Failed to read file: %v", err)
	}

	// Benchmark ParallelHierarchicalParser
	b.Run("ParallelHierarchicalParser", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := parser.NewParallelHierarchicalParser()
			_, err := p.Parse(filePath)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}
		}
	})

	// Benchmark cacack parser
	b.Run("cacack", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			reader := bytes.NewReader(data)
			_, err := decoder.Decode(reader)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}
		}
	})
}

