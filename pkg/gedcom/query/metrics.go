package query

import (
	"sync"
	"time"
)

// Metrics collects performance and usage metrics
type Metrics struct {
	mu sync.RWMutex

	// Query metrics
	QueryCount        int64
	QueryTotalTime    time.Duration
	QueryAvgTime      time.Duration
	QueryMaxTime      time.Duration
	QueryMinTime      time.Duration

	// Cache metrics
	CacheHits         int64
	CacheMisses       int64
	CacheHitRate      float64

	// Storage metrics
	StorageReads      int64
	StorageWrites     int64
	StorageReadTime   time.Duration
	StorageWriteTime  time.Duration

	// Graph metrics
	NodesLoaded       int64
	EdgesLoaded       int64
	GraphBuildTime    time.Duration

	// Error metrics
	ErrorCount        int64
	ErrorByType       map[string]int64
}

// NewMetrics creates a new Metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		ErrorByType: make(map[string]int64),
	}
}

// RecordQuery records a query execution
func (m *Metrics) RecordQuery(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.QueryCount++
	m.QueryTotalTime += duration

	if m.QueryCount == 1 {
		m.QueryMaxTime = duration
		m.QueryMinTime = duration
	} else {
		if duration > m.QueryMaxTime {
			m.QueryMaxTime = duration
		}
		if duration < m.QueryMinTime {
			m.QueryMinTime = duration
		}
	}

	m.QueryAvgTime = m.QueryTotalTime / time.Duration(m.QueryCount)
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheHits++
	m.updateCacheHitRate()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheMisses++
	m.updateCacheHitRate()
}

func (m *Metrics) updateCacheHitRate() {
	total := m.CacheHits + m.CacheMisses
	if total > 0 {
		m.CacheHitRate = float64(m.CacheHits) / float64(total) * 100.0
	}
}

// RecordStorageRead records a storage read operation
func (m *Metrics) RecordStorageRead(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StorageReads++
	m.StorageReadTime += duration
}

// RecordStorageWrite records a storage write operation
func (m *Metrics) RecordStorageWrite(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StorageWrites++
	m.StorageWriteTime += duration
}

// RecordNodeLoad records a node load operation
func (m *Metrics) RecordNodeLoad() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.NodesLoaded++
}

// RecordEdgeLoad records an edge load operation
func (m *Metrics) RecordEdgeLoad() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EdgesLoaded++
}

// RecordGraphBuild records graph build time
func (m *Metrics) RecordGraphBuild(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GraphBuildTime = duration
}

// RecordError records an error occurrence
func (m *Metrics) RecordError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorCount++
	m.ErrorByType[errorType]++
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errorByType := make(map[string]int64)
	for k, v := range m.ErrorByType {
		errorByType[k] = v
	}

	return MetricsSnapshot{
		QueryCount:        m.QueryCount,
		QueryAvgTime:      m.QueryAvgTime,
		QueryMaxTime:      m.QueryMaxTime,
		QueryMinTime:      m.QueryMinTime,
		CacheHits:         m.CacheHits,
		CacheMisses:       m.CacheMisses,
		CacheHitRate:      m.CacheHitRate,
		StorageReads:      m.StorageReads,
		StorageWrites:     m.StorageWrites,
		StorageReadTime:   m.StorageReadTime,
		StorageWriteTime:  m.StorageWriteTime,
		NodesLoaded:       m.NodesLoaded,
		EdgesLoaded:       m.EdgesLoaded,
		GraphBuildTime:    m.GraphBuildTime,
		ErrorCount:         m.ErrorCount,
		ErrorByType:       errorByType,
	}
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.QueryCount = 0
	m.QueryTotalTime = 0
	m.QueryAvgTime = 0
	m.QueryMaxTime = 0
	m.QueryMinTime = 0
	m.CacheHits = 0
	m.CacheMisses = 0
	m.CacheHitRate = 0
	m.StorageReads = 0
	m.StorageWrites = 0
	m.StorageReadTime = 0
	m.StorageWriteTime = 0
	m.NodesLoaded = 0
	m.EdgesLoaded = 0
	m.GraphBuildTime = 0
	m.ErrorCount = 0
	m.ErrorByType = make(map[string]int64)
}

// MetricsSnapshot is a read-only snapshot of metrics
type MetricsSnapshot struct {
	QueryCount        int64
	QueryAvgTime      time.Duration
	QueryMaxTime      time.Duration
	QueryMinTime      time.Duration
	CacheHits         int64
	CacheMisses       int64
	CacheHitRate      float64
	StorageReads      int64
	StorageWrites     int64
	StorageReadTime   time.Duration
	StorageWriteTime  time.Duration
	NodesLoaded       int64
	EdgesLoaded       int64
	GraphBuildTime    time.Duration
	ErrorCount        int64
	ErrorByType       map[string]int64
}

