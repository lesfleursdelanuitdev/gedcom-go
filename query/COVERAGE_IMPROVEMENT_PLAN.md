# Query Package Coverage Improvement Plan

## Current Status
- **Current Coverage: 73.4%**
- **Target Coverage: 80%+**
- **Gap: 6.6%**

## Priority Areas for Improvement

### 1. HIGH PRIORITY - Functions with 0% Coverage

#### Events Query Functions (events_query.go)
These functions are completely untested:
- `GetEventsOnDate` (line 180) - 0%
- `GetEventsOnDateByType` (line 231) - 0%
- `GetRecordsForEvent` (line 155) - 0%
- `matchesDate` (line 250) - 0%

**Test Strategy:**
```go
// Test with real GEDCOM files that have events
// Use royal92.ged or pres2020.ged
func TestEventsQuery_GetEventsOnDate(t *testing.T) {
    // Parse real file
    // Find individuals/families with known event dates
    // Test GetEventsOnDate with specific year/month/day
    // Test GetEventsOnDateByType for specific event types
    // Test GetRecordsForEvent to find all records with a specific event
    // Test matchesDate with various date formats
}
```

#### Birthday Filters (birthday_filters.go)
These have very low coverage (36-41%):
- `ByBirthMonth` - 41.2%
- `ByBirthDay` - 41.2%
- `ByBirthMonthAndDay` - 36.8%

**Test Strategy:**
```go
// Test with real individuals from testdata files
func TestBirthdayFilters_Comprehensive(t *testing.T) {
    // Test ByBirthMonth with individuals born in each month
    // Test ByBirthDay with edge cases (day 1, 31, invalid days)
    // Test ByBirthMonthAndDay with real birth dates
    // Test with range dates, ABOUT dates, etc.
    // Test invalid inputs (month 0, 13, day 0, 32)
}
```

#### Filter Execution (filter_execution.go)
- `filterByBool` (line 380) - 0%
  - Used for boolean filters (HasChildren, HasSpouse, IsLiving)
  - Needs tests with real individuals

**Test Strategy:**
```go
func TestFilterByBool(t *testing.T) {
    // Test with individuals who have children
    // Test with individuals who don't have children
    // Test with individuals who have spouses
    // Test with individuals who don't have spouses
    // Test error handling in checkFunc
}
```

### 2. MEDIUM PRIORITY - Functions with Low Coverage (50-79%)

#### Builder Functions (builder.go)
- `BuildGraph` - 66.7%
- `createEdges` - 57.1%
- `createReferenceEdges` - 64.2%
- `createEventNodesAndEdges` - 67.4%

**Test Strategy:**
- Test BuildGraph with various tree structures
- Test edge creation with missing references
- Test event node creation with different event types
- Test error cases (invalid XREFs, circular references)

#### Cache Functions (cache.go)
- `newQueryCache` - 66.7%
- `set` - 57.1%

**Test Strategy:**
- Test cache creation with different configs
- Test cache set/get operations
- Test cache eviction
- Test concurrent access

#### Config Functions (config.go)
- `LoadConfig` - 72.7%
- `SaveConfig` - 56.2%

**Test Strategy:**
- Test loading config from file
- Test saving config to file
- Test invalid config files
- Test default config values

#### Ancestor/Descendant Queries
- `Count` - 75.0% (both ancestor and descendant)
- `Exists` - 75.0% (both ancestor and descendant)

**Test Strategy:**
- Test Count with non-existent individuals
- Test Count with individuals who have no ancestors/descendants
- Test Exists with various scenarios
- Test edge cases (self, immediate family, deep trees)

### 3. LOW PRIORITY - Functions Near Target (75-79%)

These are close to 80% and just need a few more test cases:
- `Count` methods (75%)
- `Exists` methods (75%)
- `ForEachIndividual` (75%)
- `ForEachFamily` (75%)
- `buildCacheKey` (75%)

## Implementation Plan

### Phase 1: Events Query Testing (Expected +2-3% coverage)
1. Create `events_query_test.go`
2. Test `GetEventsOnDate` with real GEDCOM files
3. Test `GetEventsOnDateByType` for different event types
4. Test `GetRecordsForEvent` to find records with specific events
5. Test `matchesDate` with various date formats and edge cases

### Phase 2: Birthday Filters Testing (Expected +1-2% coverage)
1. Enhance existing birthday filter tests
2. Test with real individuals from testdata files
3. Test all date types (exact, range, ABOUT, BEFORE, AFTER)
4. Test edge cases (invalid months/days, year wraparound)

### Phase 3: Filter Execution Testing (Expected +0.5-1% coverage)
1. Test `filterByBool` function
2. Test with real individuals (with/without children, spouses)
3. Test error handling

### Phase 4: Builder and Cache Testing (Expected +1-2% coverage)
1. Test BuildGraph edge cases
2. Test cache operations
3. Test config loading/saving

### Phase 5: Edge Cases and Error Handling (Expected +0.5-1% coverage)
1. Test Count/Exists with invalid inputs
2. Test ForEach methods with empty graphs
3. Test buildCacheKey edge cases

## Test Files to Create/Enhance

### New Test Files
1. **`events_query_test.go`** - Comprehensive tests for events query functions
2. **`birthday_filters_comprehensive_test.go`** - Enhanced birthday filter tests

### Files to Enhance
1. **`coverage_test.go`** - Add tests for filterByBool
2. **`integration_testdata_test.go`** - Add more edge case tests
3. **`coverage_edge_cases_test.go`** - Add builder and cache edge cases

## Expected Results

After implementing all phases:
- **Target Coverage: 80%+**
- **Estimated Final Coverage: 80-82%**
- **Functions with 0% coverage: < 10** (mostly hybrid storage, which is excluded)

## Quick Wins (Highest Impact, Lowest Effort)

1. **Events Query Functions** - 4 functions, 0% coverage, straightforward to test
2. **Birthday Filters** - 3 functions, 36-41% coverage, need more test cases
3. **filterByBool** - 1 function, 0% coverage, simple boolean logic

These three areas alone should bring coverage from 73.4% to ~78-79%.

## Notes

- **Hybrid Storage**: Excluded from coverage goals (as requested)
- **Lazy Loading**: Lower priority, can be tested later
- **Test Helpers**: Already exist (`testdata_helper.go`, `testhelpers.go`)
- **Real Data**: Use testdata files (royal92.ged, pres2020.ged, etc.) for realistic tests

