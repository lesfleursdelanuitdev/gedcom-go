//go:build external_comparison
// +build external_comparison

package performance_comparison

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cacack/gedcom-go/decoder"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/internal/parser"
)

// Test files to compare (GEDCOM 5.5.1 compatible)
var testFiles = []struct {
	name     string
	path     string
	expected string // "cacack", "user", or "both"
}{
	// From cacack project (GEDCOM 5.5.1)
	{
		name:     "cacack-5.5.1-minimal",
		path:     "../../../gedcom-go-cacack/testdata/gedcom-5.5.1/minimal.ged",
		expected: "both",
	},
	{
		name:     "cacack-5.5.1-comprehensive",
		path:     "../../../gedcom-go-cacack/testdata/gedcom-5.5.1/comprehensive.ged",
		expected: "both",
	},
	// From cacack project (GEDCOM 5.5 - compatible with 5.5.1)
	{
		name:     "cacack-5.5-royal92",
		path:     "../../../gedcom-go-cacack/testdata/gedcom-5.5/royal92.ged",
		expected: "both",
	},
	{
		name:     "cacack-5.5-minimal",
		path:     "../../../gedcom-go-cacack/testdata/gedcom-5.5/minimal.ged",
		expected: "both",
	},
	// From user's project
	{
		name:     "user-royal92",
		path:     "../../testdata/royal92.ged",
		expected: "both",
	},
	// From family-tree folder
	{
		name:     "family-tree-gracis",
		path:     "../../../../family-tree/gedcom/gracis.ged",
		expected: "both",
	},
	{
		name:     "family-tree-xavier",
		path:     "../../../../family-tree/gedcom/xavier.ged",
		expected: "both",
	},
	{
		name:     "family-tree-tree1",
		path:     "../../../../family-tree/gedcom/tree1.ged",
		expected: "both",
	},
}

// BenchmarkCacackParser benchmarks the cacack/gedcom-go parser
func BenchmarkCacackParser(b *testing.B) {
	for _, tf := range testFiles {
		if !fileExists(tf.path) {
			b.Logf("Skipping %s: file not found", tf.name)
			continue
		}

		b.Run(tf.name, func(b *testing.B) {
			file, err := os.Open(tf.path)
			if err != nil {
				b.Skipf("Failed to open file: %v", err)
			}
			defer file.Close()

			// Read file into memory for fair comparison
			data, err := io.ReadAll(file)
			if err != nil {
				b.Fatalf("Failed to read file: %v", err)
			}

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
}

// BenchmarkUserParser benchmarks the user's gedcom-go parser
func BenchmarkUserParser(b *testing.B) {
	for _, tf := range testFiles {
		if !fileExists(tf.path) {
			b.Logf("Skipping %s: file not found", tf.name)
			continue
		}

		b.Run(tf.name, func(b *testing.B) {
			parser := parser.NewHierarchicalParser()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := parser.Parse(tf.path)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
			}
		})
	}
}

// TestParserComparison runs a detailed comparison test
func TestParserComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping parser comparison in short mode")
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Parser Performance Comparison (GEDCOM 5.5.1 compatible files)")
	fmt.Println(strings.Repeat("=", 80))

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

			// Test cacack parser
			fmt.Println("\n[Cacack Parser]")
			cacackTime, cacackAllocs, cacackBytes := benchmarkCacackParser(tf.path, t)
			if cacackTime > 0 {
				fmt.Printf("  Time: %v\n", cacackTime)
				fmt.Printf("  Throughput: %.2f KB/s\n", float64(fileSize)/1024/cacackTime.Seconds())
				fmt.Printf("  Allocations: %d\n", cacackAllocs)
				fmt.Printf("  Bytes allocated: %d\n", cacackBytes)
			}

			// Test user parser
			fmt.Println("\n[User Parser]")
			userTime, userAllocs, userBytes := benchmarkUserParser(tf.path, t)
			if userTime > 0 {
				fmt.Printf("  Time: %v\n", userTime)
				fmt.Printf("  Throughput: %.2f KB/s\n", float64(fileSize)/1024/userTime.Seconds())
				fmt.Printf("  Allocations: %d\n", userAllocs)
				fmt.Printf("  Bytes allocated: %d\n", userBytes)
			}

			// Compare results
			if cacackTime > 0 && userTime > 0 {
				fmt.Println("\n[Comparison]")
				ratio := float64(userTime) / float64(cacackTime)
				if ratio > 1.0 {
					fmt.Printf("  Cacack parser is %.2fx faster\n", ratio)
				} else {
					fmt.Printf("  User parser is %.2fx faster\n", 1.0/ratio)
				}

				allocRatio := float64(userAllocs) / float64(cacackAllocs)
				if allocRatio > 1.0 {
					fmt.Printf("  Cacack parser uses %.2fx fewer allocations\n", allocRatio)
				} else {
					fmt.Printf("  User parser uses %.2fx fewer allocations\n", 1.0/allocRatio)
				}

				bytesRatio := float64(userBytes) / float64(cacackBytes)
				if bytesRatio > 1.0 {
					fmt.Printf("  Cacack parser uses %.2fx less memory\n", bytesRatio)
				} else {
					fmt.Printf("  User parser uses %.2fx less memory\n", 1.0/bytesRatio)
				}
			}
		})
	}
}

// benchmarkCacackParser runs a single benchmark for cacack parser
func benchmarkCacackParser(filePath string, t *testing.T) (time.Duration, uint64, uint64) {
	file, err := os.Open(filePath)
	if err != nil {
		t.Logf("Failed to open file: %v", err)
		return 0, 0, 0
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		t.Logf("Failed to read file: %v", err)
		return 0, 0, 0
	}

	const iterations = 10
	var totalTime time.Duration
	var totalAllocs uint64
	var totalBytes uint64

	for i := 0; i < iterations; i++ {
		start := time.Now()
		reader := bytes.NewReader(data)
		_, err := decoder.Decode(reader)
		if err != nil {
			t.Logf("Parse failed: %v", err)
			return 0, 0, 0
		}
		duration := time.Since(start)
		totalTime += duration
		// Note: We can't easily get allocs/bytes without using testing.B
		// This is a simplified version
	}

	avgTime := totalTime / iterations
	return avgTime, totalAllocs, totalBytes
}

// benchmarkUserParser runs a single benchmark for user parser
func benchmarkUserParser(filePath string, t *testing.T) (time.Duration, uint64, uint64) {
	parser := parser.NewHierarchicalParser()

	const iterations = 10
	var totalTime time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := parser.Parse(filePath)
		if err != nil {
			t.Logf("Parse failed: %v", err)
			return 0, 0, 0
		}
		duration := time.Since(start)
		totalTime += duration
	}

	avgTime := totalTime / iterations
	return avgTime, 0, 0 // Can't easily get allocs without testing.B
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
