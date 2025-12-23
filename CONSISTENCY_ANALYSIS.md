# Codebase Consistency Analysis

## Executive Summary

The codebase is **largely consistent** with some areas that need attention. Overall architecture and patterns are well-established, but there are a few inconsistencies in test utilities, error messages, and hybrid mode checks.

## ✅ Consistent Areas

### 1. Error Handling
- **Pattern**: Consistent use of `fmt.Errorf` with `%w` for error wrapping
- **Example**: `fmt.Errorf("failed to initialize SQLite: %w", err)`
- **Status**: ✅ Consistent across all files

### 2. Naming Conventions
- **Public functions**: PascalCase (e.g., `GetIndividual`, `BuildGraphHybrid`)
- **Private functions**: camelCase (e.g., `getIndividualFromHybrid`, `buildGraphInSQLite`)
- **Types**: PascalCase (e.g., `Graph`, `HybridStorage`, `FilterQuery`)
- **Status**: ✅ Consistent

### 3. Constructor Functions
- **Pattern**: `New*` prefix for constructors
- **Examples**: `NewGraph`, `NewHybridStorage`, `NewFilterQuery`, `NewHybridCache`
- **Status**: ✅ Consistent

### 4. Mutex Usage
- **Pattern**: `g.mu.RLock()` for reads, `g.mu.Lock()` for writes
- **Defer pattern**: Consistent use of `defer g.mu.Unlock()` / `defer g.mu.RUnlock()`
- **Status**: ✅ Consistent

### 5. Hybrid Mode Checks
- **Pattern**: `if g.hybridMode && g.hybridStorage != nil`
- **Status**: ✅ Mostly consistent (see issues below)

### 6. Code Formatting
- **Status**: ✅ All files properly formatted (gofmt passes)

## ⚠️ Inconsistencies Found

### 1. Test Timeout Utilities (HIGH PRIORITY)

**Issue**: Two different timeout helper functions exist:
- `testWithTimeout` in `hybrid_stress_test.go` (used in stress tests)
- `RunWithTimeout` in `test_timeout.go` (not used anywhere)

**Impact**: Confusion about which to use, potential duplication

**Recommendation**: 
- Remove `RunWithTimeout` from `test_timeout.go` (unused)
- Keep `testWithTimeout` in `hybrid_stress_test.go` for stress tests
- Or consolidate into a single utility

**Files**:
- `pkg/gedcom/query/test_timeout.go` - Contains unused `RunWithTimeout`
- `pkg/gedcom/query/hybrid_stress_test.go` - Contains `testWithTimeout` (used)

### 2. Error Message Capitalization (MEDIUM PRIORITY)

**Issue**: Inconsistent capitalization in error messages:
- Some use: `"failed to initialize SQLite"`
- Some use: `"Failed to build hybrid graph"` (capital F)

**Impact**: Minor - affects user experience and log consistency

**Recommendation**: Standardize on lowercase for error messages:
- ✅ `"failed to initialize SQLite"`
- ❌ `"Failed to build hybrid graph"` → `"failed to build hybrid graph"`

**Files to check**:
- `hybrid_builder.go` - Mixed capitalization
- `hybrid_storage.go` - Consistent lowercase
- `graph.go` - Check for inconsistencies

### 3. Hybrid Mode Check Pattern (LOW PRIORITY)

**Issue**: Most methods check `if g.hybridMode && g.hybridStorage != nil`, but some variations exist:
- Some check `g.hybridMode` first
- Some check `g.hybridStorage != nil` first
- Order doesn't matter functionally, but consistency helps readability

**Impact**: Low - functionally equivalent, but inconsistent style

**Recommendation**: Standardize on: `if g.hybridMode && g.hybridStorage != nil`

**Files**: All `Get*` methods in `graph.go`

### 4. Cleanup Function in testWithTimeout (LOW PRIORITY)

**Issue**: `testWithTimeout` declares `cleanupFunc` but never uses it:
```go
var cleanupFunc func() // Store cleanup function if test provides one
```

**Impact**: Dead code, misleading comment

**Recommendation**: Remove unused `cleanupFunc` variable or implement it properly

**File**: `pkg/gedcom/query/hybrid_stress_test.go:28`

### 5. Missing Context Support (MEDIUM PRIORITY)

**Issue**: Long-running operations don't accept `context.Context` for cancellation:
- `BuildGraphHybrid` doesn't accept context
- Database operations don't support cancellation
- This makes it harder to implement proper timeouts

**Impact**: Medium - affects ability to cancel long-running operations

**Recommendation**: Add context support to:
- `BuildGraphHybrid(ctx context.Context, ...)`
- Database query operations
- Long-running graph construction

### 6. Getter Method Patterns (LOW PRIORITY)

**Issue**: Some getters return `nil` on not found, others might return errors:
- `GetIndividual(xrefID string) *IndividualNode` - returns nil
- `GetFamily(xrefID string) *FamilyNode` - returns nil
- All consistent in returning nil, but no error variant exists

**Impact**: Low - current pattern is consistent, but might want error variants for hybrid mode failures

**Recommendation**: Consider adding error-returning variants:
- `GetIndividualWithError(xrefID string) (*IndividualNode, error)`

## 📊 Statistics

- **Total Go files in query package**: 53
- **Files with formatting issues**: 0 (all pass gofmt)
- **Files with vet issues**: 0 (all pass go vet)
- **Inconsistencies found**: 6 (mostly minor)

## 🔧 Recommended Fixes

### ✅ Priority 1: Remove Unused Code (COMPLETED)
1. ✅ Removed `RunWithTimeout` from `test_timeout.go` (unused)
2. ✅ Removed unused `cleanupFunc` from `testWithTimeout`
3. ✅ Added documentation comment to `GetAllIndividuals` explaining hybrid mode check pattern

### Priority 2: Standardize Error Messages
1. Audit all error messages for capitalization
2. Standardize on lowercase: `"failed to ..."`

### Priority 3: Add Context Support
1. Add `context.Context` parameter to `BuildGraphHybrid`
2. Add context checks in long-running loops
3. Propagate context through database operations

### Priority 4: Documentation
1. Document hybrid mode vs regular mode behavior
2. Add examples showing when to use each mode
3. Document test timeout patterns

## ✅ What's Working Well

1. **Consistent error wrapping**: All errors use `%w` for proper error chain
2. **Consistent mutex patterns**: Proper RLock/Lock usage throughout
3. **Consistent naming**: Clear, predictable naming conventions
4. **Consistent hybrid mode checks**: Pattern is clear and repeated correctly
5. **Consistent cleanup**: `defer Close()` patterns are used correctly
6. **Consistent test structure**: Tests follow similar patterns

## 📝 Summary

The codebase is **well-structured and mostly consistent**. The main issues are:
1. Unused test utility function
2. Minor error message capitalization
3. Missing context support for cancellation

These are all **low-to-medium priority** and don't affect functionality. The codebase demonstrates good engineering practices overall.

