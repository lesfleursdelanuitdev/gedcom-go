# Duplicate Detection Optimization - Blocking Implementation

## Summary

Implemented a **two-stage blocking pipeline** to transform duplicate detection from O(n²) to O(n * avg_block_size), making it scalable to 1.5M+ individuals.

## Changes Made

### 1. New Blocking System (`blocking.go`)

**Blocking Strategy:**
- **Primary Block:** `surname_soundex + birthYear` (and ±1 year for fuzzy matching)
- **Fallback Block 1:** `surname_soundex + given_initial`
- **Fallback Block 2:** `surname_prefix(4) + birth_place_token`

**Key Features:**
- Uses `uint32` IDs instead of pointers to reduce memory
- Precomputes normalized fields (Soundex, prefixes, tokens)
- Hash-based block keys for O(1) lookups
- Limits candidates per person (default: 200) to prevent blowup

### 2. Updated Parallel Processing

- Replaced O(n²) comparison generation with blocking-based candidate generation
- Maintains parallel worker pool for scoring
- Each worker processes independent candidate pairs

### 3. Updated Sequential Processing

- Same blocking strategy for consistency
- Sequential scoring of candidates

### 4. Configuration

Added to `DuplicateConfig`:
- `UseBlocking bool` (default: true) - Enable/disable blocking
- `MaxCandidatesPerPerson int` (default: 200) - Limit candidates per person

## Performance Impact

### Before (O(n²))
- 100K individuals: ~5 billion comparisons → **hours/days**
- 1.5M individuals: ~1.125 trillion comparisons → **impossible**

### After (Blocking)
- 100K individuals: ~20-200 candidates per person → **~2-20M comparisons** → **seconds/minutes**
- 1.5M individuals: ~20-200 candidates per person → **~30-300M comparisons** → **minutes**

**Expected speedup:** 1000x - 100000x reduction in comparisons

## Blocking Keys Explained

### Primary Block: `surname_soundex + birthYear`

**Why it works:**
- Soundex handles name variations (Smith/Smyth/Smythe all map to S530)
- Birth year is a strong discriminator (rare for duplicates to have different years)
- ±1 year expansion handles transcription errors

**Example:**
- "John Smith" born 1850 → Block: `S530|1850`
- "John Smyth" born 1851 → Block: `S530|1851` (matches via ±1 expansion)

### Fallback Blocks

**When primary fails:**
- Missing birth dates → Use `surname_soundex + given_initial`
- Missing/messy places → Use `surname_prefix + birth_place_token`

## Memory Optimization

1. **uint32 IDs:** Store indices instead of pointers (4 bytes vs 8 bytes)
2. **Precomputed fields:** Compute Soundex/prefixes once during indexing
3. **Hash-based keys:** uint64 hashes for efficient map lookups
4. **Candidate limits:** Cap at 200 candidates per person to prevent memory blowup

## Testing

Run performance tests:
```bash
go test -v -run TestPerformance_DuplicateDetection_100K ./pkg/gedcom/duplicate/...
go test -v -run TestPerformance_DuplicateDetection_500K ./pkg/gedcom/duplicate/...
```

## Future Enhancements

1. **Adaptive blocking:** Adjust candidate limits based on block sizes
2. **Multi-pass blocking:** Use multiple blocking strategies and merge results
3. **Relationship-based blocking:** Use parent/spouse surnames for additional blocks
4. **Incremental updates:** Update blocks when individuals are added/modified

## Configuration Example

```go
config := duplicate.DefaultConfig()
config.UseBlocking = true
config.MaxCandidatesPerPerson = 200  // Adjust based on data quality
config.UseParallelProcessing = true
detector := duplicate.NewDuplicateDetector(config)
```

## Notes

- Blocking is enabled by default
- For very clean data, you can increase `MaxCandidatesPerPerson`
- For messy/noisy data, blocking may need tuning (add more fallback blocks)
- The system gracefully handles missing fields (uses fallback blocks)

