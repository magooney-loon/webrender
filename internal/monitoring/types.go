package monitoring

import "time"

// ErrorEntry represents a single error event
type ErrorEntry struct {
	Time    time.Time
	Message string
	Code    string
	Path    string
	Method  string
}

// VisitorEntry represents a single visitor analytics entry
type VisitorEntry struct {
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	Referrer  string    `json:"referrer,omitempty"`
}

// HealthStatus represents the complete health status of the system
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"` // Human-readable status message

	// Server metrics
	Uptime        string `json:"uptime"`
	StartTime     string `json:"start_time"`
	GoVersion     string `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
	NumCPU        int    `json:"num_cpu"`
	NumCGO        int64  `json:"num_cgo_calls"`

	// Memory metrics
	MemoryStats struct {
		Alloc          uint64  `json:"alloc_bytes"`
		TotalAlloc     uint64  `json:"total_alloc_bytes"`
		Sys            uint64  `json:"sys_bytes"`
		NumGC          uint32  `json:"num_gc"`
		LastGC         string  `json:"last_gc_time"`
		GCCPUFraction  float64 `json:"gc_cpu_fraction"`
		HeapInUse      uint64  `json:"heap_in_use_bytes"`
		HeapIdle       uint64  `json:"heap_idle_bytes"`
		HeapReleased   uint64  `json:"heap_released_bytes"`
		HeapObjects    uint64  `json:"heap_objects"`
		StackInUse     uint64  `json:"stack_in_use_bytes"`
		MemoryUsage    float64 `json:"memory_usage_percent"`
		LastGCDuration string  `json:"last_gc_duration"`
	} `json:"memory_stats"`

	// Request metrics
	RequestStats struct {
		TotalRequests      uint64                 `json:"total_requests"`
		ActiveRequests     int64                  `json:"active_requests"`
		TotalErrors        uint64                 `json:"total_errors"`
		AverageLatency     float64                `json:"average_latency_ms"`
		StatusCodeCounts   map[int]uint64         `json:"status_code_counts"`
		RequestsPerMinute  float64                `json:"requests_per_minute"`
		ErrorRate          float64                `json:"error_rate"`
		LastRequestTime    time.Time              `json:"last_request_time"`
		LatencyPercentiles map[string]float64     `json:"latency_percentiles"`
		RouteStats         map[string]RouteStats  `json:"route_stats"`
		MethodStats        map[string]MethodStats `json:"method_stats"`
	} `json:"request_stats"`

	// Connection metrics
	ConnectionStats struct {
		OpenConnections  int64              `json:"open_connections"`
		TotalConnections uint64             `json:"total_connections"`
		ConnectionErrors uint64             `json:"connection_errors"`
		AvgConnDuration  float64            `json:"avg_conn_duration_ms"`
		ConnPercentiles  map[string]float64 `json:"conn_duration_percentiles"`
	} `json:"connection_stats"`

	// Payload metrics
	PayloadStats struct {
		TotalBytesIn   uint64  `json:"total_bytes_in"`
		TotalBytesOut  uint64  `json:"total_bytes_out"`
		MaxPayloadSize int64   `json:"max_payload_size"`
		MinPayloadSize int64   `json:"min_payload_size"`
		AvgPayloadSize float64 `json:"avg_payload_size"`
	} `json:"payload_stats"`

	// Cache metrics
	CacheStats struct {
		HitRate     float64 `json:"hit_rate"`
		MissRate    float64 `json:"miss_rate"`
		Evictions   uint64  `json:"evictions"`
		TotalHits   uint64  `json:"total_hits"`
		TotalMisses uint64  `json:"total_misses"`
	} `json:"cache_stats,omitempty"`

	// Analytics metrics
	AnalyticsStats struct {
		TotalVisitors          uint64            `json:"total_visitors"`
		UniqueVisitors         uint64            `json:"unique_visitors"`
		VisitorsToday          uint64            `json:"visitors_today"`
		TopPages               map[string]uint64 `json:"top_pages"`
		TopReferrers           map[string]uint64 `json:"top_referrers"`
		TopUserAgents          map[string]uint64 `json:"top_user_agents"`
		AvgSessionDuration     float64           `json:"avg_session_duration_seconds"`
		BounceRate             float64           `json:"bounce_rate"`
		VisitorsPerHour        []uint64          `json:"visitors_per_hour"`
		VisitorsByDay          map[string]uint64 `json:"visitors_by_day"`
		NewVsReturningVisitors struct {
			New       uint64 `json:"new"`
			Returning uint64 `json:"returning"`
		} `json:"new_vs_returning"`
	} `json:"analytics_stats,omitempty"`

	// Recent errors
	RecentErrors []ErrorEntry `json:"recent_errors,omitempty"`
}

// RouteStats contains metrics for a specific route
type RouteStats struct {
	Hits               uint64             `json:"hits"`
	Errors             uint64             `json:"errors"`
	AverageLatency     float64            `json:"avg_latency_ms"`
	StatusCodes        map[int]uint64     `json:"status_codes"`
	LatencyPercentiles map[string]float64 `json:"latency_percentiles"`
}

// MethodStats contains metrics for a specific HTTP method
type MethodStats struct {
	Count     uint64  `json:"count"`
	Errors    uint64  `json:"errors"`
	ErrorRate float64 `json:"error_rate"`
}
