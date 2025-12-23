# Test Timeout Root Cause Analysis

## Problem

Tests are timing out after 2 minutes and then hanging because BadgerDB background goroutines don't shut down properly.

## Root Cause

1. **Test Timeout**: `TestHybridStorage_1M` (1 million individuals) takes longer than 2 minutes to build the graph
2. **BadgerDB Background Goroutines**: When BadgerDB is opened, it creates multiple background goroutines:
   - Compactors (for LSM tree maintenance)
   - Flushers (for memtable flushing)
   - Value log GC (garbage collection)
   - Publishers (for value threshold updates)
   - Connection openers (for SQLite)
3. **Incomplete Cleanup**: When the test times out:
   - `testWithTimeout` calls `t.Fatalf` which stops the test
   - But the goroutine running the test continues
   - `defer graph.Close()` may not execute, or executes too late
   - BadgerDB's `Close()` method needs time to shut down all background goroutines
   - The test framework waits for all goroutines to complete

## Evidence

From the stack trace:
```
goroutine 49 [chan receive, 1 minutes]:
github.com/dgraph-io/badger/v4.(*valueLog).waitOnGC(...)
created by github.com/dgraph-io/badger/v4.Open

goroutine 50 [select, 1 minutes]:
github.com/dgraph-io/badger/v4.(*publisher).listenForUpdates(...)
created by github.com/dgraph-io/badger/v4.Open
```

These goroutines are still running 1 minute after the test started, indicating they're not being cleaned up.

## Solutions

### Solution 1: Ensure Cleanup on Timeout (Immediate Fix)

Modify `testWithTimeout` to ensure cleanup happens even on timeout:

```go
func testWithTimeout(t *testing.T, testName string, fn func(*testing.T)) {
    ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
    defer cancel()

    done := make(chan bool, 1)
    var panicErr interface{}
    var cleanup func() // Store cleanup function

    go func() {
        defer func() {
            if r := recover(); r != nil {
                panicErr = r
            }
            // Ensure cleanup happens
            if cleanup != nil {
                cleanup()
            }
            select {
            case done <- true:
            default:
            }
        }()
        fn(t)
    }()

    select {
    case <-done:
        if panicErr != nil {
            t.Fatalf("Test %s panicked: %v", testName, panicErr)
        }
    case <-ctx.Done():
        // Timeout - ensure cleanup
        if cleanup != nil {
            cleanup()
        }
        t.Fatalf("Test %s timed out after %v", testName, TestTimeout)
    }
}
```

### Solution 2: Skip Large Tests by Default (Recommended)

Large stress tests should only run when explicitly requested:

```go
func TestHybridStorage_1M(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping hybrid storage stress test in short mode")
    }
    // Only run if explicitly requested
    if os.Getenv("RUN_STRESS_TESTS") == "" {
        t.Skip("Skipping stress test. Set RUN_STRESS_TESTS=1 to run")
    }
    // ... rest of test
}
```

### Solution 3: Use Context for Cancellation

Pass context through to long-running operations so they can be cancelled:

```go
func BuildGraphHybrid(ctx context.Context, tree *gedcom.GedcomTree, ...) (*Graph, error) {
    // Check context periodically
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    // ... continue building
}
```

### Solution 4: Increase Timeout for Large Tests

Large tests need more time:

```go
func TestHybridStorage_1M(t *testing.T) {
    // Use longer timeout for this specific test
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    // ... test code
}
```

## Recommended Approach

1. **Immediate**: Skip large stress tests by default (Solution 2)
2. **Short-term**: Ensure cleanup on timeout (Solution 1)
3. **Long-term**: Add context support for cancellation (Solution 3)

## For Users

To run stress tests:
```bash
# Run all tests (skips stress tests)
go test -timeout 2m ./pkg/gedcom/query

# Run stress tests explicitly
RUN_STRESS_TESTS=1 go test -timeout 10m ./pkg/gedcom/query -run TestHybridStorage_1M
```

