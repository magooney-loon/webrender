package database

import (
	"sync"
	"time"
)

// Metrics stores and manages database performance metrics
type Metrics struct {
	mu sync.RWMutex

	// Query metrics
	queryCount       uint64
	queryErrors      uint64
	queryLatencies   []time.Duration
	slowQueriesCount uint64
	slowThreshold    time.Duration

	// Transaction metrics
	txCount       uint64
	txCommitted   uint64
	txRolledBack  uint64
	txErrors      uint64
	activeTxCount int64
	avgTxDuration time.Duration
	txLatencies   []time.Duration

	// Cache metrics
	cacheHits      uint64
	cacheMisses    uint64
	cacheEvictions uint64
	cacheSizeBytes uint64
	cacheItemCount int64

	// Connection metrics
	connectionCount uint64
	connectionMax   int64
	connectionIdle  int64
	connectionOpen  int64

	// Storage metrics
	dbSizeBytes    uint64
	dbPageCount    uint64
	dbPageSize     uint64
	indexSizeBytes uint64

	// Timing
	lastQueryTime time.Time
	lastTxTime    time.Time
	lastError     time.Time
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		queryLatencies: make([]time.Duration, 0, 100),
		txLatencies:    make([]time.Duration, 0, 100),
		slowThreshold:  500 * time.Millisecond,
		lastQueryTime:  time.Now(),
		lastTxTime:     time.Now(),
		lastError:      time.Now(),
	}
}

// RecordQueryDuration records the duration of a query
func (m *Metrics) RecordQueryDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queryCount++
	m.lastQueryTime = time.Now()

	// Record for latency calculations
	m.queryLatencies = append(m.queryLatencies, duration)
	if len(m.queryLatencies) > 100 {
		// Keep last 100 queries for statistics
		m.queryLatencies = m.queryLatencies[1:]
	}

	// Record slow queries
	if duration >= m.slowThreshold {
		m.slowQueriesCount++
	}
}

// RecordQueryError records a query error
func (m *Metrics) RecordQueryError() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queryErrors++
	m.lastError = time.Now()
}

// RecordTransactionStart records the start of a transaction
func (m *Metrics) RecordTransactionStart() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.txCount++
	m.activeTxCount++
	m.lastTxTime = time.Now()
}

// RecordTransactionEnd records the end of a transaction
func (m *Metrics) RecordTransactionEnd(duration time.Duration, committed bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record metrics
	m.activeTxCount--
	m.txLatencies = append(m.txLatencies, duration)
	if len(m.txLatencies) > 100 {
		m.txLatencies = m.txLatencies[1:]
	}

	// Calculate average duration
	var totalDuration time.Duration
	for _, d := range m.txLatencies {
		totalDuration += d
	}
	if len(m.txLatencies) > 0 {
		m.avgTxDuration = totalDuration / time.Duration(len(m.txLatencies))
	}

	// Record outcome
	if err != nil {
		m.txErrors++
		m.lastError = time.Now()
	} else if committed {
		m.txCommitted++
	} else {
		m.txRolledBack++
	}
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheHits++
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheMisses++
}

// RecordCacheEviction records a cache eviction
func (m *Metrics) RecordCacheEviction() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheEvictions++
}

// UpdateCacheSize updates cache size metrics
func (m *Metrics) UpdateCacheSize(itemCount int64, sizeBytes uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheItemCount = itemCount
	m.cacheSizeBytes = sizeBytes
}

// UpdateDBStats updates database statistics
func (m *Metrics) UpdateDBStats(sizeBytes, pageCount, pageSize, indexSize uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dbSizeBytes = sizeBytes
	m.dbPageCount = pageCount
	m.dbPageSize = pageSize
	m.indexSizeBytes = indexSize
}

// GetMetrics returns a snapshot of the current metrics
func (m *Metrics) GetMetrics() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate cache hit ratio
	cacheHitRatio := float64(0)
	totalCacheAccess := m.cacheHits + m.cacheMisses
	if totalCacheAccess > 0 {
		cacheHitRatio = float64(m.cacheHits) / float64(totalCacheAccess)
	}

	// Calculate average query latency
	avgQueryLatency := time.Duration(0)
	if len(m.queryLatencies) > 0 {
		var total time.Duration
		for _, latency := range m.queryLatencies {
			total += latency
		}
		avgQueryLatency = total / time.Duration(len(m.queryLatencies))
	}

	return MetricsSnapshot{
		QueryCount:           m.queryCount,
		QueryErrors:          m.queryErrors,
		AvgQueryLatency:      avgQueryLatency,
		SlowQueriesCount:     m.slowQueriesCount,
		TransactionCount:     m.txCount,
		TransactionCommits:   m.txCommitted,
		TransactionRollbacks: m.txRolledBack,
		TransactionErrors:    m.txErrors,
		ActiveTransactions:   m.activeTxCount,
		AvgTxDuration:        m.avgTxDuration,
		CacheHits:            m.cacheHits,
		CacheMisses:          m.cacheMisses,
		CacheEvictions:       m.cacheEvictions,
		CacheHitRatio:        cacheHitRatio,
		CacheItemCount:       m.cacheItemCount,
		CacheSizeBytes:       m.cacheSizeBytes,
		DatabaseSizeBytes:    m.dbSizeBytes,
		LastQueryTime:        m.lastQueryTime,
		LastTransactionTime:  m.lastTxTime,
		LastErrorTime:        m.lastError,
	}
}

// MetricsSnapshot represents a point-in-time snapshot of database metrics
type MetricsSnapshot struct {
	// Query metrics
	QueryCount       uint64
	QueryErrors      uint64
	AvgQueryLatency  time.Duration
	SlowQueriesCount uint64

	// Transaction metrics
	TransactionCount     uint64
	TransactionCommits   uint64
	TransactionRollbacks uint64
	TransactionErrors    uint64
	ActiveTransactions   int64
	AvgTxDuration        time.Duration

	// Cache metrics
	CacheHits      uint64
	CacheMisses    uint64
	CacheEvictions uint64
	CacheHitRatio  float64
	CacheItemCount int64
	CacheSizeBytes uint64

	// Storage metrics
	DatabaseSizeBytes uint64

	// Timing
	LastQueryTime       time.Time
	LastTransactionTime time.Time
	LastErrorTime       time.Time
}
