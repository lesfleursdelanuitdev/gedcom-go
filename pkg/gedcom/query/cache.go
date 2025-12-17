package query

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// cacheKey represents a cache key for query results.
type cacheKey struct {
	queryType string
	params    string
}

// queryCache is a simple LRU cache for query results.
type queryCache struct {
	mu      sync.RWMutex
	cache   map[string]interface{}
	maxSize int
}

// newQueryCache creates a new query cache.
func newQueryCache(maxSize int) *queryCache {
	if maxSize <= 0 {
		maxSize = 1000 // Default size
	}
	return &queryCache{
		cache:   make(map[string]interface{}),
		maxSize: maxSize,
	}
}

// get retrieves a value from the cache.
func (qc *queryCache) get(key string) (interface{}, bool) {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	val, ok := qc.cache[key]
	return val, ok
}

// set stores a value in the cache.
func (qc *queryCache) set(key string, value interface{}) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	// If cache is full, remove oldest entry (simple FIFO eviction)
	if len(qc.cache) >= qc.maxSize {
		// Remove first entry (simple eviction strategy)
		for k := range qc.cache {
			delete(qc.cache, k)
			break
		}
	}

	qc.cache[key] = value
}

// clear clears the cache.
func (qc *queryCache) clear() {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	qc.cache = make(map[string]interface{})
}

// makeCacheKey creates a cache key from query type and parameters.
func makeCacheKey(queryType string, params ...interface{}) string {
	h := sha256.New()
	h.Write([]byte(queryType))
	for _, p := range params {
		h.Write([]byte(fmt.Sprintf("%v", p)))
	}
	return hex.EncodeToString(h.Sum(nil))
}
