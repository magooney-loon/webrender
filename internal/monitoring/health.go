package monitoring

import (
	"runtime"
	"sync/atomic"
	"time"
)

// GetHealthStatus returns the current health status of the system
func (m *Metrics) GetHealthStatus() *HealthStatus {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &HealthStatus{
		Status:        m.determineHealthStatus(),
		Timestamp:     time.Now(),
		Uptime:        time.Since(m.startTime).String(),
		StartTime:     m.startTime.Format(time.RFC3339),
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		NumCGO:        runtime.NumCgoCall(),
	}

	// Memory stats
	status.MemoryStats.Alloc = memStats.Alloc
	status.MemoryStats.TotalAlloc = memStats.TotalAlloc
	status.MemoryStats.Sys = memStats.Sys
	status.MemoryStats.NumGC = memStats.NumGC
	status.MemoryStats.LastGC = time.Unix(0, int64(memStats.LastGC)).Format(time.RFC3339Nano)
	status.MemoryStats.GCCPUFraction = memStats.GCCPUFraction
	status.MemoryStats.HeapInUse = memStats.HeapInuse
	status.MemoryStats.HeapIdle = memStats.HeapIdle
	status.MemoryStats.HeapReleased = memStats.HeapReleased
	status.MemoryStats.HeapObjects = memStats.HeapObjects
	status.MemoryStats.StackInUse = memStats.StackInuse
	status.MemoryStats.MemoryUsage = float64(memStats.Alloc) / float64(memStats.Sys) * 100
	status.MemoryStats.LastGCDuration = m.lastGCDuration.String()

	// Request stats
	status.RequestStats.TotalRequests = atomic.LoadUint64(&m.totalRequests)
	status.RequestStats.ActiveRequests = atomic.LoadInt64(&m.activeRequests)
	status.RequestStats.TotalErrors = atomic.LoadUint64(&m.totalErrors)
	status.RequestStats.AverageLatency = m.avgResponseTime
	status.RequestStats.StatusCodeCounts = m.statusCodeCounts
	status.RequestStats.LastRequestTime = m.lastRequestTime

	// Calculate requests per minute
	if !m.lastRequestTime.IsZero() {
		duration := time.Since(m.startTime).Minutes()
		if duration > 0 {
			status.RequestStats.RequestsPerMinute = float64(m.totalRequests) / duration
		}
	}

	// Calculate error rate
	if m.totalRequests > 0 {
		status.RequestStats.ErrorRate = float64(m.totalErrors) / float64(m.totalRequests) * 100
	}

	// Route stats
	status.RequestStats.RouteStats = make(map[string]RouteStats)
	for path := range m.routeHits {
		routeStats := RouteStats{
			Hits:               m.routeHits[path],
			Errors:             m.routeErrors[path],
			StatusCodes:        m.routeStatusCodes[path],
			LatencyPercentiles: calculatePercentiles(m.routeLatencies[path]),
		}

		if len(m.routeLatencies[path]) > 0 {
			var sum float64
			for _, lat := range m.routeLatencies[path] {
				sum += lat
			}
			routeStats.AverageLatency = sum / float64(len(m.routeLatencies[path]))
		}

		status.RequestStats.RouteStats[path] = routeStats
	}

	// Method stats
	status.RequestStats.MethodStats = make(map[string]MethodStats)
	for method := range m.methodCounts {
		count := m.methodCounts[method]
		errors := m.methodErrors[method]
		errorRate := float64(0)
		if count > 0 {
			errorRate = float64(errors) / float64(count) * 100
		}

		status.RequestStats.MethodStats[method] = MethodStats{
			Count:     count,
			Errors:    errors,
			ErrorRate: errorRate,
		}
	}

	// Calculate latency percentiles
	if len(m.requestLatencies) > 0 {
		sorted := make([]float64, len(m.requestLatencies))
		copy(sorted, m.requestLatencies)
		status.RequestStats.LatencyPercentiles = calculatePercentiles(sorted)
	}

	// Connection stats
	status.ConnectionStats.OpenConnections = atomic.LoadInt64(&m.openConnections)
	status.ConnectionStats.TotalConnections = atomic.LoadUint64(&m.totalConnections)
	status.ConnectionStats.ConnectionErrors = atomic.LoadUint64(&m.connectionErrors)
	status.ConnectionStats.AvgConnDuration = m.avgConnDuration
	if len(m.connDurations) > 0 {
		sorted := make([]float64, len(m.connDurations))
		copy(sorted, m.connDurations)
		status.ConnectionStats.ConnPercentiles = calculatePercentiles(sorted)
	}

	// Payload stats
	status.PayloadStats.TotalBytesIn = atomic.LoadUint64(&m.totalBytesIn)
	status.PayloadStats.TotalBytesOut = atomic.LoadUint64(&m.totalBytesOut)
	status.PayloadStats.MaxPayloadSize = m.maxPayloadSize
	status.PayloadStats.MinPayloadSize = m.minPayloadSize
	if m.totalRequests > 0 {
		status.PayloadStats.AvgPayloadSize = float64(m.totalBytesIn) / float64(m.totalRequests)
	}

	// Cache stats
	totalOps := atomic.LoadUint64(&m.cacheHits) + atomic.LoadUint64(&m.cacheMisses)
	if totalOps > 0 {
		status.CacheStats = struct {
			HitRate     float64 `json:"hit_rate"`
			MissRate    float64 `json:"miss_rate"`
			Evictions   uint64  `json:"evictions"`
			TotalHits   uint64  `json:"total_hits"`
			TotalMisses uint64  `json:"total_misses"`
		}{
			HitRate:     float64(m.cacheHits) / float64(totalOps) * 100,
			MissRate:    float64(m.cacheMisses) / float64(totalOps) * 100,
			Evictions:   atomic.LoadUint64(&m.cacheEvictions),
			TotalHits:   atomic.LoadUint64(&m.cacheHits),
			TotalMisses: atomic.LoadUint64(&m.cacheMisses),
		}
	}

	// Analytics stats
	if m.visitors > 0 {
		// Calculate analytics metrics
		status.AnalyticsStats.TotalVisitors = m.visitors
		status.AnalyticsStats.UniqueVisitors = uint64(len(m.uniqueVisitors))
		status.AnalyticsStats.VisitorsToday = uint64(len(m.visitorsToday))

		// Top pages
		status.AnalyticsStats.TopPages = make(map[string]uint64)
		for path, count := range m.pageViews {
			status.AnalyticsStats.TopPages[path] = count
		}

		// Top referrers
		status.AnalyticsStats.TopReferrers = make(map[string]uint64)
		for ref, count := range m.referrers {
			status.AnalyticsStats.TopReferrers[ref] = count
		}

		// Top user agents
		status.AnalyticsStats.TopUserAgents = make(map[string]uint64)
		for ua, count := range m.userAgents {
			status.AnalyticsStats.TopUserAgents[ua] = count
		}

		// Average session duration
		if len(m.sessionDurations) > 0 {
			var sum float64
			for _, duration := range m.sessionDurations {
				sum += duration
			}
			status.AnalyticsStats.AvgSessionDuration = sum / float64(len(m.sessionDurations))
		}

		// Bounce rate
		if m.visitors > 0 {
			status.AnalyticsStats.BounceRate = float64(m.singlePageVisits) / float64(m.visitors) * 100
		}

		// Visitors per hour
		status.AnalyticsStats.VisitorsPerHour = make([]uint64, 24)
		for i := 0; i < 24; i++ {
			status.AnalyticsStats.VisitorsPerHour[i] = m.visitorsPerHour[i]
		}

		// Visitors by day
		status.AnalyticsStats.VisitorsByDay = make(map[string]uint64)
		for day, count := range m.visitorsByDay {
			status.AnalyticsStats.VisitorsByDay[day] = count
		}

		// New vs returning visitors
		status.AnalyticsStats.NewVsReturningVisitors.New = m.visitors - uint64(len(m.newVsReturning))
		status.AnalyticsStats.NewVsReturningVisitors.Returning = uint64(len(m.newVsReturning))
	}

	// Recent errors (last 10)
	if len(m.lastErrors) > 0 {
		status.RecentErrors = make([]ErrorEntry, 0, 10)
		count := 0
		idx := m.errorIndex - 1
		for count < 10 {
			if idx < 0 {
				idx = len(m.lastErrors) - 1
			}
			if m.lastErrors[idx].Time.IsZero() {
				break
			}
			status.RecentErrors = append(status.RecentErrors, m.lastErrors[idx])
			idx--
			count++
		}
	}

	return status
}

// determineHealthStatus evaluates current metrics to determine system health
func (m *Metrics) determineHealthStatus() string {
	// Check various health indicators
	if atomic.LoadInt64(&m.activeRequests) > 1000 || // Too many active requests
		atomic.LoadInt64(&m.openConnections) > 1000 || // Too many open connections
		m.avgResponseTime > 1000 || // Average response time > 1s
		(m.totalRequests > 100 && float64(m.totalErrors)/float64(m.totalRequests) > 0.5) { // Error rate > 50%
		return "unhealthy"
	}

	if atomic.LoadInt64(&m.activeRequests) > 500 || // Many active requests
		atomic.LoadInt64(&m.openConnections) > 500 || // Many open connections
		m.avgResponseTime > 500 || // Average response time > 500ms
		(m.totalRequests > 100 && float64(m.totalErrors)/float64(m.totalRequests) > 0.1) { // Error rate > 10%
		return "degraded"
	}

	return "healthy"
}
