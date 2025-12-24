package performance_comparison

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/internal/parser"
)

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestInternalParserComparison compares all internal parsers
func TestInternalParserComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping parser comparison in short mode")
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Internal Parser Performance Comparison")
	fmt.Println(strings.Repeat("=", 80))

	testFiles := []struct {
		name string
		path string
	}{
		{"royal92", "../../testdata/royal92.ged"},
		{"pres2020", "../../testdata/pres2020.ged"},
		{"gracis", "../../testdata/gracis.ged"},
		{"xavier", "../../testdata/xavier.ged"},
		{"tree1", "../../testdata/tree1.ged"},
	}

	for _, tf := range testFiles {
		if !fileExists(tf.path) {
			t.Logf("Skipping %s: file not found", tf.name)
			continue
		}

		t.Run(tf.name, func(t *testing.T) {
			fmt.Printf("\n--- Testing: %s ---\n", tf.name)
			fmt.Printf("File: %s\n", tf.path)

			// Get file size
			fileInfo, err := os.Stat(tf.path)
			if err != nil {
				t.Fatalf("Failed to stat file: %v", err)
			}
			fileSize := fileInfo.Size()
			fmt.Printf("File size: %.2f KB\n", float64(fileSize)/1024)

			results := make(map[string]time.Duration)

			// Test HierarchicalParser
			fmt.Println("\n[HierarchicalParser]")
			time, err := benchmarkParser("HierarchicalParser", func() (interface{}, error) {
				p := parser.NewHierarchicalParser()
				_, err := p.Parse(tf.path)
				return p, err
			}, t)
			if err == nil {
				results["HierarchicalParser"] = time
				fmt.Printf("  Time: %v\n", time)
				fmt.Printf("  Throughput: %.2f KB/s\n", float64(fileSize)/1024/time.Seconds())
			}

			// Test ParallelHierarchicalParser
			fmt.Println("\n[ParallelHierarchicalParser]")
			time, err = benchmarkParser("ParallelHierarchicalParser", func() (interface{}, error) {
				p := parser.NewParallelHierarchicalParser()
				_, err := p.Parse(tf.path)
				return p, err
			}, t)
			if err == nil {
				results["ParallelHierarchicalParser"] = time
				fmt.Printf("  Time: %v\n", time)
				fmt.Printf("  Throughput: %.2f KB/s\n", float64(fileSize)/1024/time.Seconds())
			}

			// Test TwoPhaseParser
			fmt.Println("\n[TwoPhaseParser]")
			time, err = benchmarkParser("TwoPhaseParser", func() (interface{}, error) {
				p := parser.NewTwoPhaseParser()
				_, err := p.Parse(tf.path)
				return p, err
			}, t)
			if err == nil {
				results["TwoPhaseParser"] = time
				fmt.Printf("  Time: %v\n", time)
				fmt.Printf("  Throughput: %.2f KB/s\n", float64(fileSize)/1024/time.Seconds())
			}

			// Compare results
			if len(results) > 1 {
				fmt.Println("\n[Comparison]")
				baseline := results["HierarchicalParser"]
				for name, duration := range results {
					if name == "HierarchicalParser" {
						continue
					}
					ratio := float64(duration) / float64(baseline)
					if ratio < 1.0 {
						fmt.Printf("  %s is %.2fx faster than HierarchicalParser\n", name, 1.0/ratio)
					} else {
						fmt.Printf("  %s is %.2fx slower than HierarchicalParser\n", name, ratio)
					}
				}
			}
		})
	}
}

// BenchmarkInternalParsers benchmarks all internal parsers
func BenchmarkInternalParsers(b *testing.B) {
	testFiles := []struct {
		name string
		path string
	}{
		{"royal92", "../../testdata/royal92.ged"},
		{"pres2020", "../../testdata/pres2020.ged"},
		{"gracis", "../../testdata/gracis.ged"},
		{"xavier", "../../testdata/xavier.ged"},
		{"tree1", "../../testdata/tree1.ged"},
	}

	for _, tf := range testFiles {
		if !fileExists(tf.path) {
			b.Logf("Skipping %s: file not found", tf.name)
			continue
		}

		// HierarchicalParser
		b.Run(tf.name+"/HierarchicalParser", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				p := parser.NewHierarchicalParser()
				_, err := p.Parse(tf.path)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
			}
		})

		// ParallelHierarchicalParser
		b.Run(tf.name+"/ParallelHierarchicalParser", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				p := parser.NewParallelHierarchicalParser()
				_, err := p.Parse(tf.path)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
			}
		})

		// TwoPhaseParser
		b.Run(tf.name+"/TwoPhaseParser", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				p := parser.NewTwoPhaseParser()
				_, err := p.Parse(tf.path)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
			}
		})
	}
}

// benchmarkParser runs a parser function multiple times and returns average time
func benchmarkParser(name string, parseFunc func() (interface{}, error), t *testing.T) (time.Duration, error) {
	const iterations = 5
	var totalTime time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := parseFunc()
		if err != nil {
			return 0, fmt.Errorf("%s failed: %v", name, err)
		}
		duration := time.Since(start)
		totalTime += duration
	}

	avgTime := totalTime / iterations
	return avgTime, nil
}

