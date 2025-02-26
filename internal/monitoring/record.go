package monitoring

import (
	"sync/atomic"
	"time"
)

// RecordRequest records metrics for a HTTP request
func (m *Metrics) RecordRequest(path, method string, statusCode int, duration time.Duration, bytesIn, bytesOut int64) {
	atomic.AddUint64(&m.totalRequests, 1)
	atomic.AddUint64(&m.totalBytesIn, uint64(bytesIn))
	atomic.AddUint64(&m.totalBytesOut, uint64(bytesOut))

	if statusCode >= 400 {
		atomic.AddUint64(&m.totalErrors, 1)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update status code counts
	m.statusCodeCounts[statusCode]++

	// Update request latencies
	latencyMs := float64(duration.Milliseconds())
	m.requestLatencies = append(m.requestLatencies, latencyMs)
	if len(m.requestLatencies) > 1000 {
		m.requestLatencies = m.requestLatencies[1:]
	}

	// Update route metrics
	if m.routeStatusCodes[path] == nil {
		m.routeStatusCodes[path] = make(map[int]uint64)
	}
	m.routeHits[path]++
	m.routeStatusCodes[path][statusCode]++
	if statusCode >= 400 {
		m.routeErrors[path]++
	}

	// Update route latencies
	if m.routeLatencies[path] == nil {
		m.routeLatencies[path] = make([]float64, 0, 100)
	}
	m.routeLatencies[path] = append(m.routeLatencies[path], latencyMs)
	if len(m.routeLatencies[path]) > 100 {
		m.routeLatencies[path] = m.routeLatencies[path][1:]
	}

	// Update method metrics
	m.methodCounts[method]++
	if statusCode >= 400 {
		m.methodErrors[method]++
	}

	// Update payload size metrics
	if m.minPayloadSize == -1 || bytesIn < m.minPayloadSize {
		m.minPayloadSize = bytesIn
	}
	if bytesIn > m.maxPayloadSize {
		m.maxPayloadSize = bytesIn
	}

	// Calculate new average response time
	var sum float64
	for _, lat := range m.requestLatencies {
		sum += lat
	}
	m.avgResponseTime = sum / float64(len(m.requestLatencies))
	m.lastRequestTime = time.Now()
}

// RecordError records an error occurrence
func (m *Metrics) RecordError(err error, code, path, method string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errorCounts[code]++
	m.lastErrors[m.errorIndex] = ErrorEntry{
		Time:    time.Now(),
		Message: err.Error(),
		Code:    code,
		Path:    path,
		Method:  method,
	}
	m.errorIndex = (m.errorIndex + 1) % len(m.lastErrors)
}

// RecordConnection records metrics for an HTTP connection
func (m *Metrics) RecordConnection(duration time.Duration) {
	atomic.AddUint64(&m.totalConnections, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	durationMs := float64(duration.Milliseconds())
	m.connDurations = append(m.connDurations, durationMs)
	if len(m.connDurations) > 1000 {
		m.connDurations = m.connDurations[1:]
	}

	var sum float64
	for _, d := range m.connDurations {
		sum += d
	}
	m.avgConnDuration = sum / float64(len(m.connDurations))
}

// RecordCacheOperation records a cache hit or miss
func (m *Metrics) RecordCacheOperation(hit bool) {
	if hit {
		atomic.AddUint64(&m.cacheHits, 1)
	} else {
		atomic.AddUint64(&m.cacheMisses, 1)
	}
}

// RecordCacheEviction records a cache eviction
func (m *Metrics) RecordCacheEviction() {
	atomic.AddUint64(&m.cacheEvictions, 1)
}

// IncrementActiveRequests increments the count of active requests
func (m *Metrics) IncrementActiveRequests() {
	atomic.AddInt64(&m.activeRequests, 1)
}

// DecrementActiveRequests decrements the count of active requests
func (m *Metrics) DecrementActiveRequests() {
	atomic.AddInt64(&m.activeRequests, -1)
}

// IncrementOpenConnections increments the count of open connections
func (m *Metrics) IncrementOpenConnections() {
	atomic.AddInt64(&m.openConnections, 1)
}

// DecrementOpenConnections decrements the count of open connections
func (m *Metrics) DecrementOpenConnections() {
	atomic.AddInt64(&m.openConnections, -1)
}

// RecordVisit records a visitor's page view for analytics
func (m *Metrics) RecordVisit(path, ip, userAgent, referrer string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Increment total visitors
	m.visitors++

	// Track unique visitors
	isNewVisitor := !m.uniqueVisitors[ip]
	if isNewVisitor {
		m.uniqueVisitors[ip] = true
	}

	// Track visitors by hour of day
	currentHour := time.Now().Hour()
	m.visitorsPerHour[currentHour]++

	// Track visitors by day
	today := time.Now().Format("2006-01-02")
	m.visitorsByDay[today]++

	// Track visitors today
	m.visitorsToday[ip] = true

	// Track page views
	m.pageViews[path]++

	// Track referrers if provided
	if referrer != "" {
		m.referrers[referrer]++
	}

	// Track user agents
	m.userAgents[userAgent]++

	// Store visit time
	now := time.Now()
	if m.visitTimes[ip] == nil {
		m.visitTimes[ip] = make([]time.Time, 0, 10)
	}
	m.visitTimes[ip] = append(m.visitTimes[ip], now)

	// Handle session tracking
	lastVisit, hasSession := m.visitorSessions[ip]
	m.visitorSessions[ip] = now

	// If returning in the same session (30 min threshold)
	if hasSession && now.Sub(lastVisit) < 30*time.Minute {
		// Update session duration
		sessionDuration := now.Sub(lastVisit).Seconds()
		m.sessionDurations = append(m.sessionDurations, sessionDuration)
		if len(m.sessionDurations) > 1000 {
			m.sessionDurations = m.sessionDurations[1:]
		}

		// No longer a bounce visit if they view multiple pages
		if len(m.visitTimes[ip]) > 1 {
			delete(m.bounceIPs, ip)
		}
	} else {
		// Start of a new session
		if len(m.visitTimes[ip]) == 1 {
			// Potential bounce - we'll mark it as a bounce for now
			m.bounceIPs[ip] = true
			m.singlePageVisits++
		}

		// Track new vs returning
		if len(m.visitTimes[ip]) > 1 {
			m.newVsReturning[ip] = true // Mark as returning visitor
		}
	}

	// Record this visit entry
	m.lastVisitorEntry[m.visitorEntryIndex] = VisitorEntry{
		Path:      path,
		Timestamp: now,
		UserAgent: userAgent,
		IP:        ip,
		Referrer:  referrer,
	}
	m.visitorEntryIndex = (m.visitorEntryIndex + 1) % len(m.lastVisitorEntry)
}
