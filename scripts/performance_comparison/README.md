# Performance Comparison Tests

This directory contains performance comparison tests for the ligneous-gedcom parser.

## Test Categories

### Internal Parser Comparison (Always Available)

**Test:** `TestInternalParserComparison`

Compares all internal parsers (HierarchicalParser, ParallelHierarchicalParser, TwoPhaseParser, StreamingParser, SmartParser).

**Run:**
```bash
go test ./scripts/performance_comparison -run TestInternalParserComparison
```

### External Parser Comparison (Requires Build Tag)

**Tests:** `TestComprehensiveComparison`, `TestParserComparison`, `TestRoyal92Comparison`

These tests compare ligneous-gedcom parsers with external parsers:
- `cacack/gedcom-go`
- `elliotchance/gedcom`

**These tests are excluded from normal test runs** to avoid integration test failures when external dependencies are not available.

**To run external comparison tests:**
```bash
# Run with build tag
go test -tags=external_comparison ./scripts/performance_comparison -run TestComprehensiveComparison -timeout 30m

# Or run all external comparison tests
go test -tags=external_comparison ./scripts/performance_comparison -timeout 30m
```

**Note:** External comparison tests require:
- `gedcom-go-cacack` repository cloned at `/apps/gedcom-go-cacack`
- `gedcom-elliotchance` repository cloned at `/apps/gedcom-elliotchance`
- Dependencies: `github.com/cacack/gedcom-go` and `github.com/elliotchance/gedcom/v39`

## Why Build Tags?

The external comparison tests use build tags (`//go:build external_comparison`) to exclude them from normal test runs because:

1. **External dependencies** - They require other repositories to be cloned locally
2. **Integration test failures** - They cause failures in CI/CD when external repos aren't available
3. **Optional feature** - Performance comparison with external parsers is optional, not required for core functionality

## Running All Tests

**Normal test run (excludes external comparisons):**
```bash
go test ./scripts/performance_comparison
```

**With external comparisons:**
```bash
go test -tags=external_comparison ./scripts/performance_comparison
```

## Documentation

See the various `.md` files in this directory for detailed analysis and results:
- `CODE_COMPARISON.md` - Code comparison between parsers
- `ELLIOTCHANCE_PERFORMANCE_ANALYSIS.md` - Why elliotchance/gedcom is slower
- `FINAL_RESULTS_ANALYSIS.md` - Performance results analysis
- `OPTIMIZATION_PLAN.md` - Optimization strategies
- And more...
