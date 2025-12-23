# Blocking System Improvements

## Summary

Implemented comprehensive improvements to the duplicate detection blocking system based on expert feedback. The system now includes:

1. **Comprehensive Metrics & Instrumentation**
2. **Adaptive Blocking** (handles giant blocks)
3. **Better Edge Case Handling** (missing data, multi-part surnames)
4. **Prioritized Candidate Selection** (better matches first)
5. **Multiple Blocking Strategies** (primary + fallbacks + rescue)

## Key Improvements

### 1. Metrics & Instrumentation

Added `BlockingMetrics` struct that tracks:
- % people with ≥1 block key
- Average candidates per person (before/after cap)
- Number of blocks created
- Top 20 largest block sizes
- Block type usage (primary vs fallback)
- #pairs generated vs #pairs actually scored

### 2. Adaptive Blocking

- **Max Block Size**: Blocks larger than 5000 are skipped (configurable)
- **Prioritized Selection**: When candidates exceed cap, best matches are selected first based on:
  - Exact year match (highest priority)
  - Year difference (±1, ±2)
  - Surname exact match (not just Soundex)
  - Place match
  - Given name prefix match

### 3. Edge Case Handling

#### Missing Data
- **No "unknown" blocks**: Don't create blocks with empty/zero values (prevents giant junk blocks)
- **5-year buckets**: For missing/uncertain dates, use 5-year buckets
- **Fallback chains**: Multiple fallback strategies ensure people without primary blocks still get candidates

#### Multi-Part Surnames
- **Smart extraction**: Handles "van der Berg", "de la Cruz", etc.
- Uses last significant word for Soundex (skips common prefixes: van, von, de, del, de la, der, den, du, le, la, les)

#### Place Tokenization
- Skips common place words: "the", "of", "in", "on", "at", "to", "for", "and", "or", "county", "city", "town", "state", "province"
- Uses first significant token for blocking

### 4. Blocking Strategies

#### Primary Block
- `surname_soundex + birthYear` (exact year)
- Expanded: `surname_soundex + birthYear ±1, ±2` (better recall)
- 5-year bucket: `surname_soundex + birthYearBucket` (for uncertain dates)

#### Fallback Blocks
1. `surname_soundex + given_initial` (when birth year missing)
2. `surname_soundex + given_prefix(2)` (looser than initial)
3. `surname_prefix(4) + birth_place_token` (when year missing)

#### Rescue Block
- `given_prefix(3) + surname_prefix(3) + place_token` (only for people with no other blocks)

### 5. Candidate Generation

- **Deduplication**: Only compare `(i, j)` where `j > i` (prevents double scoring)
- **No self-pairs**: Person never compares to themselves
- **Smart capping**: When `MaxCandidatesPerPerson` is reached, prioritize best matches

## Current Issue: 0% Block Coverage

The metrics show 0% of people have blocks, which indicates a potential issue with:
1. Surname extraction from GEDCOM format
2. Birth year extraction
3. Block key generation

**Next Steps:**
1. Add debug logging to see what values are being extracted
2. Verify GEDCOM parsing extracts surnames correctly from "Person %d /Test/" format
3. Check if birth year extraction works for "DATE 1800" format
4. Test with a small dataset to verify blocking works

## Files Modified

- `pkg/gedcom/duplicate/blocking.go`: Core blocking logic with improvements
- `pkg/gedcom/duplicate/blocking_metrics.go`: New metrics system
- `pkg/gedcom/duplicate/detector.go`: Integrated blocking metrics into results
- `pkg/gedcom/duplicate/parallel.go`: Updated to use new blocking
- `pkg/gedcom/duplicate/sequential.go`: Updated to use new blocking
- `stress_test.go`: Added blocking metrics output

## Testing

Run with:
```bash
go test -v -run TestStress_1_5M_Comprehensive -timeout 30m
```

Check for "Blocking Metrics" section in output to see:
- People with blocks
- Average candidates per person
- Top block sizes
- Block type usage

