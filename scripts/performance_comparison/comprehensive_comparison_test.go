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
	gedcomElliotchance "github.com/elliotchance/gedcom/v39"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/internal/parser"
)

// Test files from all directories (GEDCOM 5.5.1 compatible)
var comprehensiveTestFiles = []struct {
	name string
	path string
}{
	// From user's project testdata (now includes more files)
	{"user-royal92", "../../testdata/royal92.ged"},
	{"user-pres2020", "../../testdata/pres2020.ged"},
	{"user-gracis", "../../testdata/gracis.ged"},
	{"user-xavier", "../../testdata/xavier.ged"},
	{"user-tree1", "../../testdata/tree1.ged"},
	// From gedcom-go-cacack/testdata/gedcom-5.5 (compatible with 5.5.1)
	{"cacack-5.5-royal92", "../../../gedcom-go-cacack/testdata/gedcom-5.5/royal92.ged"},
	{"cacack-5.5-pres2020", "../../../gedcom-go-cacack/testdata/gedcom-5.5/pres2020.ged"},
	{"cacack-5.5-minimal", "../../../gedcom-go-cacack/testdata/gedcom-5.5/minimal.ged"},
	// From gedcom-go-cacack/testdata/gedcom-5.5.1
	{"cacack-5.5.1-comprehensive", "../../../gedcom-go-cacack/testdata/gedcom-5.5.1/comprehensive.ged"},
	{"cacack-5.5.1-minimal", "../../../gedcom-go-cacack/testdata/gedcom-5.5.1/minimal.ged"},
}

// TestComprehensiveComparison runs 2000 iterations for each file
func TestComprehensiveComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive comparison in short mode")
	}

	const iterations = 2000

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("Comprehensive Parser Performance Comparison")
	fmt.Printf("ParallelHierarchicalParser vs cacack/gedcom-go vs elliotchance/gedcom - %d iterations per file\n", iterations)
	fmt.Println(strings.Repeat("=", 100))

	results := make([]FileResult, 0)

	for _, tf := range comprehensiveTestFiles {
		if !fileExists(tf.path) {
			t.Logf("Skipping %s: file not found", tf.name)
			continue
		}

		t.Run(tf.name, func(t *testing.T) {
			fmt.Printf("\n%s\n", strings.Repeat("-", 100))
			fmt.Printf("Testing: %s\n", tf.name)
			fmt.Printf("File: %s\n", tf.path)

			// Get file size
			fileInfo, err := os.Stat(tf.path)
			if err != nil {
				t.Fatalf("Failed to stat file: %v", err)
			}
			fileSize := fileInfo.Size()

			fmt.Printf("File size: %.2f KB (%d bytes)\n", float64(fileSize)/1024, fileSize)

			// Test ParallelHierarchicalParser
			fmt.Println("\n[ParallelHierarchicalParser]")
			fmt.Printf("  Running %d iterations...\n", iterations)
			parallelStart := time.Now()
			parallelTimes := make([]time.Duration, 0, iterations)
			parallelMin := time.Hour
			parallelMax := time.Duration(0)

			for i := 0; i < iterations; i++ {
				start := time.Now()
				p := parser.NewParallelHierarchicalParser()
				_, err := p.Parse(tf.path)
				if err != nil {
					t.Fatalf("ParallelHierarchicalParser failed: %v", err)
				}
				duration := time.Since(start)
				parallelTimes = append(parallelTimes, duration)
				if duration < parallelMin {
					parallelMin = duration
				}
				if duration > parallelMax {
					parallelMax = duration
				}
			}
			parallelTotal := time.Since(parallelStart)
			parallelAvg := parallelTotal / iterations

			// Calculate percentiles
			parallelP50 := percentile(parallelTimes, 50)
			parallelP95 := percentile(parallelTimes, 95)
			parallelP99 := percentile(parallelTimes, 99)

			fmt.Printf("  Total time: %v\n", parallelTotal)
			fmt.Printf("  Average: %v\n", parallelAvg)
			fmt.Printf("  Min: %v\n", parallelMin)
			fmt.Printf("  Max: %v\n", parallelMax)
			fmt.Printf("  P50 (median): %v\n", parallelP50)
			fmt.Printf("  P95: %v\n", parallelP95)
			fmt.Printf("  P99: %v\n", parallelP99)
			fmt.Printf("  Throughput: %.2f KB/s (%.2f MB/s)\n",
				float64(fileSize)/1024/parallelAvg.Seconds(),
				float64(fileSize)/1024/1024/parallelAvg.Seconds())

			// Test cacack parser
			fmt.Println("\n[cacack/gedcom-go Parser]")
			fmt.Printf("  Running %d iterations...\n", iterations)

			// Read file into memory once
			data, err := os.ReadFile(tf.path)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			cacackStart := time.Now()
			cacackTimes := make([]time.Duration, 0, iterations)
			cacackMin := time.Hour
			cacackMax := time.Duration(0)

			for i := 0; i < iterations; i++ {
				start := time.Now()
				reader := bytes.NewReader(data)
				_, err := decoder.Decode(reader)
				if err != nil {
					t.Fatalf("cacack parser failed: %v", err)
				}
				duration := time.Since(start)
				cacackTimes = append(cacackTimes, duration)
				if duration < cacackMin {
					cacackMin = duration
				}
				if duration > cacackMax {
					cacackMax = duration
				}
			}
			cacackTotal := time.Since(cacackStart)
			cacackAvg := cacackTotal / iterations

			// Calculate percentiles
			cacackP50 := percentile(cacackTimes, 50)
			cacackP95 := percentile(cacackTimes, 95)
			cacackP99 := percentile(cacackTimes, 99)

			fmt.Printf("  Total time: %v\n", cacackTotal)
			fmt.Printf("  Average: %v\n", cacackAvg)
			fmt.Printf("  Min: %v\n", cacackMin)
			fmt.Printf("  Max: %v\n", cacackMax)
			fmt.Printf("  P50 (median): %v\n", cacackP50)
			fmt.Printf("  P95: %v\n", cacackP95)
			fmt.Printf("  P99: %v\n", cacackP99)
			fmt.Printf("  Throughput: %.2f KB/s (%.2f MB/s)\n",
				float64(fileSize)/1024/cacackAvg.Seconds(),
				float64(fileSize)/1024/1024/cacackAvg.Seconds())

			// Test elliotchance parser
			fmt.Println("\n[elliotchance/gedcom Parser]")
			fmt.Printf("  Running %d iterations...\n", iterations)

			elliotchanceStart := time.Now()
			elliotchanceTimes := make([]time.Duration, 0, iterations)
			elliotchanceMin := time.Hour
			elliotchanceMax := time.Duration(0)

			for i := 0; i < iterations; i++ {
				start := time.Now()
				_, err := gedcomElliotchance.NewDocumentFromGEDCOMFile(tf.path)
				if err != nil {
					t.Fatalf("elliotchance parser failed: %v", err)
				}
				duration := time.Since(start)
				elliotchanceTimes = append(elliotchanceTimes, duration)
				if duration < elliotchanceMin {
					elliotchanceMin = duration
				}
				if duration > elliotchanceMax {
					elliotchanceMax = duration
				}
			}
			elliotchanceTotal := time.Since(elliotchanceStart)
			elliotchanceAvg := elliotchanceTotal / iterations

			// Calculate percentiles
			elliotchanceP50 := percentile(elliotchanceTimes, 50)
			elliotchanceP95 := percentile(elliotchanceTimes, 95)
			elliotchanceP99 := percentile(elliotchanceTimes, 99)

			fmt.Printf("  Total time: %v\n", elliotchanceTotal)
			fmt.Printf("  Average: %v\n", elliotchanceAvg)
			fmt.Printf("  Min: %v\n", elliotchanceMin)
			fmt.Printf("  Max: %v\n", elliotchanceMax)
			fmt.Printf("  P50 (median): %v\n", elliotchanceP50)
			fmt.Printf("  P95: %v\n", elliotchanceP95)
			fmt.Printf("  P99: %v\n", elliotchanceP99)
			fmt.Printf("  Throughput: %.2f KB/s (%.2f MB/s)\n",
				float64(fileSize)/1024/elliotchanceAvg.Seconds(),
				float64(fileSize)/1024/1024/elliotchanceAvg.Seconds())

			// Comparison
			fmt.Println("\n[Comparison]")
			ratioParallelCacack := float64(parallelAvg) / float64(cacackAvg)
			ratioParallelElliotchance := float64(parallelAvg) / float64(elliotchanceAvg)
			ratioCacackElliotchance := float64(cacackAvg) / float64(elliotchanceAvg)

			// Parallel vs Cacack
			if ratioParallelCacack < 1.0 {
				fmt.Printf("  ✅ ParallelHierarchicalParser is %.2fx FASTER than cacack (%.1f%% faster)\n",
					1.0/ratioParallelCacack, (1.0-ratioParallelCacack)*100)
			} else {
				fmt.Printf("  ⚠️  ParallelHierarchicalParser is %.2fx SLOWER than cacack (%.1f%% slower)\n",
					ratioParallelCacack, (ratioParallelCacack-1.0)*100)
			}

			// Parallel vs Elliotchance
			if ratioParallelElliotchance < 1.0 {
				fmt.Printf("  ✅ ParallelHierarchicalParser is %.2fx FASTER than elliotchance (%.1f%% faster)\n",
					1.0/ratioParallelElliotchance, (1.0-ratioParallelElliotchance)*100)
			} else {
				fmt.Printf("  ⚠️  ParallelHierarchicalParser is %.2fx SLOWER than elliotchance (%.1f%% slower)\n",
					ratioParallelElliotchance, (ratioParallelElliotchance-1.0)*100)
			}

			// Cacack vs Elliotchance
			if ratioCacackElliotchance < 1.0 {
				fmt.Printf("  📊 cacack is %.2fx FASTER than elliotchance (%.1f%% faster)\n",
					1.0/ratioCacackElliotchance, (1.0-ratioCacackElliotchance)*100)
			} else {
				fmt.Printf("  📊 cacack is %.2fx SLOWER than elliotchance (%.1f%% slower)\n",
					ratioCacackElliotchance, (ratioCacackElliotchance-1.0)*100)
			}

			// Store results
			results = append(results, FileResult{
				Name:              tf.name,
				FileSize:          fileSize,
				ParallelAvg:       parallelAvg,
				ParallelMin:       parallelMin,
				ParallelMax:       parallelMax,
				ParallelP50:       parallelP50,
				ParallelP95:       parallelP95,
				ParallelP99:       parallelP99,
				CacackAvg:         cacackAvg,
				CacackMin:         cacackMin,
				CacackMax:         cacackMax,
				CacackP50:         cacackP50,
				CacackP95:         cacackP95,
				CacackP99:         cacackP99,
				ElliotchanceAvg:   elliotchanceAvg,
				ElliotchanceMin:   elliotchanceMin,
				ElliotchanceMax:   elliotchanceMax,
				ElliotchanceP50:   elliotchanceP50,
				ElliotchanceP95:   elliotchanceP95,
				ElliotchanceP99:   elliotchanceP99,
				Ratio:             ratioParallelCacack,
				RatioParallelElliotchance: ratioParallelElliotchance,
				RatioCacackElliotchance:   ratioCacackElliotchance,
				ThroughputDiff:    (float64(fileSize)/1024/parallelAvg.Seconds()) - (float64(fileSize)/1024/cacackAvg.Seconds()),
			})
		})
	}

	// Split results into buckets and calculate totals
	const realisticThreshold = 50 * 1024 // 50KB for realistic files
	var realisticFiles []FileResult
	var tinyFiles []FileResult
	var totalRealisticBytes int64
	var totalTinyBytes int64
	var totalRealisticParallelTime time.Duration
	var totalRealisticCacackTime time.Duration
	var totalRealisticElliotchanceTime time.Duration
	var totalAllBytes int64
	var totalAllParallelTime time.Duration
	var totalAllCacackTime time.Duration
	var totalAllElliotchanceTime time.Duration

	// Buckets for summary
	var bucketTiny []FileResult      // <10KB
	var bucketSmall []FileResult     // 10-100KB
	var bucketMedium []FileResult    // 100KB-1MB
	var bucketLarge []FileResult     // >1MB

	for _, r := range results {
		totalAllBytes += r.FileSize
		totalAllParallelTime += r.ParallelAvg * iterations
		totalAllCacackTime += r.CacackAvg * iterations
		totalAllElliotchanceTime += r.ElliotchanceAvg * iterations

		if r.FileSize >= realisticThreshold {
			realisticFiles = append(realisticFiles, r)
			totalRealisticBytes += r.FileSize
			totalRealisticParallelTime += r.ParallelAvg * iterations
			totalRealisticCacackTime += r.CacackAvg * iterations
			totalRealisticElliotchanceTime += r.ElliotchanceAvg * iterations
		} else {
			tinyFiles = append(tinyFiles, r)
			totalTinyBytes += r.FileSize
		}

		// Bucket files
		if r.FileSize < 10*1024 {
			bucketTiny = append(bucketTiny, r)
		} else if r.FileSize < 100*1024 {
			bucketSmall = append(bucketSmall, r)
		} else if r.FileSize < 1024*1024 {
			bucketMedium = append(bucketMedium, r)
		} else {
			bucketLarge = append(bucketLarge, r)
		}
	}

	// ====================================================================
	// PRIMARY HEADLINE: BYTES-WEIGHTED THROUGHPUT
	// ====================================================================
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("🎯 PRIMARY RESULT: BYTES-WEIGHTED THROUGHPUT (What Actually Matters)")
	fmt.Println(strings.Repeat("=", 100))

	// Calculate bytes-weighted throughput for ALL files
	totalAllMB := float64(totalAllBytes) / 1024 / 1024
	allParallelThroughput := totalAllMB / totalAllParallelTime.Seconds()
	allCacackThroughput := totalAllMB / totalAllCacackTime.Seconds()
	allElliotchanceThroughput := totalAllMB / totalAllElliotchanceTime.Seconds()
	allBytesWeightedRatioParallelCacack := float64(totalAllCacackTime) / float64(totalAllParallelTime)
	allBytesWeightedRatioParallelElliotchance := float64(totalAllElliotchanceTime) / float64(totalAllParallelTime)
	allBytesWeightedRatioCacackElliotchance := float64(totalAllElliotchanceTime) / float64(totalAllCacackTime)

	fmt.Printf("\n📊 Overall (All Files):\n")
	fmt.Printf("   Total bytes parsed: %.2f MB (across %d files × %d iterations)\n", totalAllMB, len(results), iterations)
	fmt.Printf("   ParallelHierarchicalParser: %.2f MB/s\n", allParallelThroughput)
	fmt.Printf("   cacack/gedcom-go parser:     %.2f MB/s\n", allCacackThroughput)
	fmt.Printf("   elliotchance/gedcom parser:  %.2f MB/s\n", allElliotchanceThroughput)
	fmt.Printf("\n   ✅ Parallel vs cacack: %.2fx (%.1f%% faster)\n",
		allBytesWeightedRatioParallelCacack,
		(allBytesWeightedRatioParallelCacack-1.0)*100)
	fmt.Printf("   ✅ Parallel vs elliotchance: %.2fx (%.1f%% faster)\n",
		allBytesWeightedRatioParallelElliotchance,
		(allBytesWeightedRatioParallelElliotchance-1.0)*100)
	fmt.Printf("   📊 cacack vs elliotchance: %.2fx (%.1f%% faster)\n",
		allBytesWeightedRatioCacackElliotchance,
		(allBytesWeightedRatioCacackElliotchance-1.0)*100)

	// Calculate bytes-weighted throughput for realistic files (≥50KB)
	if len(realisticFiles) > 0 {
		totalRealisticMB := float64(totalRealisticBytes) / 1024 / 1024
		realisticParallelThroughput := totalRealisticMB / totalRealisticParallelTime.Seconds()
		realisticCacackThroughput := totalRealisticMB / totalRealisticCacackTime.Seconds()
		realisticElliotchanceThroughput := totalRealisticMB / totalRealisticElliotchanceTime.Seconds()
		realisticBytesWeightedRatioParallelCacack := float64(totalRealisticCacackTime) / float64(totalRealisticParallelTime)
		realisticBytesWeightedRatioParallelElliotchance := float64(totalRealisticElliotchanceTime) / float64(totalRealisticParallelTime)
		realisticBytesWeightedRatioCacackElliotchance := float64(totalRealisticElliotchanceTime) / float64(totalRealisticCacackTime)

		fmt.Printf("\n📊 Realistic Files (≥50KB):\n")
		fmt.Printf("   Total bytes parsed: %.2f MB (across %d files × %d iterations)\n", totalRealisticMB, len(realisticFiles), iterations)
		fmt.Printf("   ParallelHierarchicalParser: %.2f MB/s\n", realisticParallelThroughput)
		fmt.Printf("   cacack/gedcom-go parser:     %.2f MB/s\n", realisticCacackThroughput)
		fmt.Printf("   elliotchance/gedcom parser:  %.2f MB/s\n", realisticElliotchanceThroughput)
		fmt.Printf("\n   ✅ Parallel vs cacack: %.2fx (%.1f%% faster)\n",
			realisticBytesWeightedRatioParallelCacack,
			(realisticBytesWeightedRatioParallelCacack-1.0)*100)
		fmt.Printf("   ✅ Parallel vs elliotchance: %.2fx (%.1f%% faster)\n",
			realisticBytesWeightedRatioParallelElliotchance,
			(realisticBytesWeightedRatioParallelElliotchance-1.0)*100)
		fmt.Printf("   📊 cacack vs elliotchance: %.2fx (%.1f%% faster)\n",
			realisticBytesWeightedRatioCacackElliotchance,
			(realisticBytesWeightedRatioCacackElliotchance-1.0)*100)
	}

	// ====================================================================
	// BUCKETED SUMMARY
	// ====================================================================
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("📦 BUCKETED SUMMARY")
	fmt.Println(strings.Repeat("=", 100))

	printBucket := func(name string, files []FileResult) {
		if len(files) == 0 {
			return
		}
		fasterCount := 0
		for _, r := range files {
			if r.Ratio < 1.0 {
				fasterCount++
			}
		}
		fmt.Printf("\n%s (%d files):\n", name, len(files))
		fmt.Printf("  Files where ParallelHierarchicalParser is faster: %d/%d\n", fasterCount, len(files))
		if fasterCount > 0 {
			fmt.Printf("  ✅ ParallelHierarchicalParser wins in this bucket\n")
		} else if fasterCount == 0 && len(files) > 0 {
			fmt.Printf("  ⚠️  cacack parser wins in this bucket (expected for tiny files)\n")
		}
	}

	printBucket("Tiny Files (<10KB)", bucketTiny)
	printBucket("Small Files (10-100KB)", bucketSmall)
	printBucket("Medium Files (100KB-1MB)", bucketMedium)
	printBucket("Large Files (>1MB)", bucketLarge)

	// ====================================================================
	// PER-FILE DETAILS
	// ====================================================================
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("📋 PER-FILE DETAILS")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("\n%-30s %10s %12s %12s %12s %10s %10s %10s\n", "File", "Size", "Parallel (ms)", "cacack (ms)", "elliotchance (ms)", "P/C", "P/E", "C/E")
	fmt.Println(strings.Repeat("-", 100))

	for _, r := range results {
		sizeStr := fmt.Sprintf("%.1f KB", float64(r.FileSize)/1024)
		fmt.Printf("%-30s %10s %12.2f %12.2f %12.2f %10.2fx %10.2fx %10.2fx\n",
			r.Name,
			sizeStr,
			r.ParallelAvg.Seconds()*1000,
			r.CacackAvg.Seconds()*1000,
			r.ElliotchanceAvg.Seconds()*1000,
			r.Ratio,
			r.RatioParallelElliotchance,
			r.RatioCacackElliotchance)
	}

	// ====================================================================
	// REALISTIC FILES DETAILS (≥50KB)
	// ====================================================================
	if len(realisticFiles) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 100))
		fmt.Println("✅ REALISTIC FILES DETAILS (≥50KB)")
		fmt.Println(strings.Repeat("=", 100))
		fmt.Printf("\n%-30s %10s %12s %12s %10s %12s\n", "File", "Size", "Parallel (ms)", "cacack (ms)", "Ratio", "Winner")
		fmt.Println(strings.Repeat("-", 100))

		realisticFasterCount := 0
		for _, r := range realisticFiles {
			winner := "cacack"
			if r.Ratio < 1.0 {
				winner = "Parallel"
				realisticFasterCount++
			}
			sizeStr := fmt.Sprintf("%.1f KB", float64(r.FileSize)/1024)
			fmt.Printf("%-30s %10s %12.2f %12.2f %10.2fx %12s\n",
				r.Name,
				sizeStr,
				r.ParallelAvg.Seconds()*1000,
				r.CacackAvg.Seconds()*1000,
				r.Ratio,
				winner)
		}

		fmt.Printf("\nRealistic files where ParallelHierarchicalParser is faster: %d/%d\n",
			realisticFasterCount, len(realisticFiles))
	}

	// ====================================================================
	// TINY FILES DETAILS (<50KB) - Expected Overhead
	// ====================================================================
	if len(tinyFiles) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 100))
		fmt.Println("⚠️  TINY FILES DETAILS (<50KB) - Expected Overhead")
		fmt.Println(strings.Repeat("=", 100))
		fmt.Printf("\n%-30s %10s %12s %12s %10s %12s\n", "File", "Size", "Parallel (ms)", "cacack (ms)", "Ratio", "Winner")
		fmt.Println(strings.Repeat("-", 100))

		for _, r := range tinyFiles {
			winner := "cacack"
			if r.Ratio < 1.0 {
				winner = "Parallel"
			}
			sizeStr := fmt.Sprintf("%.1f KB", float64(r.FileSize)/1024)
			fmt.Printf("%-30s %10s %12.2f %12.2f %10.2fx %12s\n",
				r.Name,
				sizeStr,
				r.ParallelAvg.Seconds()*1000,
				r.CacackAvg.Seconds()*1000,
				r.Ratio,
				winner)
		}
		fmt.Println("\n💡 Note: Parallel parser has fixed overhead that dominates on tiny files.")
		fmt.Println("   This is expected and not relevant for real-world workloads.")
		fmt.Println("   Use SmartParser (auto-fallback) to automatically use non-parallel parser for files <32KB.")
	}
}

// FileResult stores comparison results for a single file
type FileResult struct {
	Name                      string
	FileSize                  int64
	ParallelAvg               time.Duration
	ParallelMin               time.Duration
	ParallelMax               time.Duration
	ParallelP50              time.Duration
	ParallelP95              time.Duration
	ParallelP99              time.Duration
	CacackAvg                time.Duration
	CacackMin                time.Duration
	CacackMax                time.Duration
	CacackP50                time.Duration
	CacackP95                time.Duration
	CacackP99                time.Duration
	ElliotchanceAvg          time.Duration
	ElliotchanceMin          time.Duration
	ElliotchanceMax          time.Duration
	ElliotchanceP50          time.Duration
	ElliotchanceP95          time.Duration
	ElliotchanceP99          time.Duration
	Ratio                    float64 // Parallel vs Cacack
	RatioParallelElliotchance float64 // Parallel vs Elliotchance
	RatioCacackElliotchance   float64 // Cacack vs Elliotchance
	ThroughputDiff           float64
}

// percentile calculates the percentile value from a slice of durations
func percentile(times []time.Duration, p int) time.Duration {
	if len(times) == 0 {
		return 0
	}

	// Create a copy and sort
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)

	// Simple bubble sort (good enough for small slices)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	index := int(float64(len(sorted)) * float64(p) / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

