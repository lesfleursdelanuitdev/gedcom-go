# Lazy Loading Test Timeout Configuration

## Summary

All lazy loading stress tests now have a **90-second (1.5 minute) timeout** to prevent tests from running indefinitely.

## Implementation

The timeout is implemented using Go's `time.AfterFunc` which will panic if the test exceeds 90 seconds:

```go
// Set timeout to 90 seconds (1.5 minutes)
timer := time.AfterFunc(90*time.Second, func() {
    panic("test timed out after 90 seconds")
})
defer timer.Stop()
```

## Test Results

### TestStress_LazyLoading_1M
- **Timeout:** 90 seconds
- **Actual Duration:** ~51 seconds
- **Status:** ✅ PASS (completes well within timeout)

### TestStress_LazyLoading_1_5M
- **Timeout:** 90 seconds
- **Status:** Configured (test may need longer for 1.5M individuals)

## Notes

- The timeout is set per test function
- Tests that exceed 90 seconds will panic with "test timed out after 90 seconds"
- For very large datasets (5M+), the timeout may need to be adjusted or the test may need to be split into smaller phases

