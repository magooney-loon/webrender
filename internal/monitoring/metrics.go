package monitoring

import (
	"sync"
	"time"
)

// Metrics stores and manages all system metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	totalRequests    uint64
	activeRequests   int64
	totalErrors      uint64
	requestLatencies []float64 // in milliseconds
	statusCodeCounts map[int]uint64
	avgResponseTime  float64
	lastRequestTime  time.Time

	// Route metrics
	routeLatencies   map[string][]float64
	routeHits        map[string]uint64
	routeErrors      map[string]uint64
	routeStatusCodes map[string]map[int]uint64

	// HTTP method metrics
	methodCounts map[string]uint64
	methodErrors map[string]uint64

	// Error tracking
	errorCounts map[string]uint64 // Error type counts
	lastErrors  []ErrorEntry      // Recent errors circular buffer
	errorIndex  int               // Current position in error buffer

	// Resource metrics
	startTime      time.Time
	lastGCTime     time.Time
	lastGCDuration time.Duration
	numGC          uint32

	// Connection metrics
	openConnections  int64
	totalConnections uint64
	connectionErrors uint64
	avgConnDuration  float64
	connDurations    []float64

	// Payload metrics
	totalBytesIn   uint64
	totalBytesOut  uint64
	maxPayloadSize int64
	minPayloadSize int64

	// Cache metrics (if using response caching)
	cacheHits      uint64
	cacheMisses    uint64
	cacheEvictions uint64

	// Analytics metrics
	visitors          uint64
	uniqueVisitors    map[string]bool        // Map of IP addresses
	visitorsToday     map[string]bool        // Map of today's IP addresses
	visitorSessions   map[string]time.Time   // Last activity time per IP
	sessionDurations  []float64              // Session durations in seconds
	bounceIPs         map[string]bool        // IPs that visited only one page
	singlePageVisits  uint64                 // Number of visits with only one page view
	pageViews         map[string]uint64      // Views per page
	referrers         map[string]uint64      // Count per referrer
	userAgents        map[string]uint64      // Count per user agent
	visitorsPerHour   [24]uint64             // Visits per hour of day
	visitorsByDay     map[string]uint64      // Visits per calendar day
	newVsReturning    map[string]bool        // Map of returning visitor IPs
	visitTimes        map[string][]time.Time // Visit timestamps per IP
	lastVisitorEntry  []VisitorEntry         // Recent visitor entries
	visitorEntryIndex int                    // Current position in visitor buffer
}

// NewMetrics creates and initializes a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		statusCodeCounts: make(map[int]uint64),
		routeLatencies:   make(map[string][]float64),
		routeHits:        make(map[string]uint64),
		routeErrors:      make(map[string]uint64),
		routeStatusCodes: make(map[string]map[int]uint64),
		methodCounts:     make(map[string]uint64),
		methodErrors:     make(map[string]uint64),
		errorCounts:      make(map[string]uint64),
		lastErrors:       make([]ErrorEntry, 100), // Keep last 100 errors
		startTime:        time.Now(),
		requestLatencies: make([]float64, 0, 1000),
		connDurations:    make([]float64, 0, 1000),
		minPayloadSize:   -1, // Initialize to -1 to indicate no requests yet

		// Analytics initialization
		uniqueVisitors:   make(map[string]bool),
		visitorsToday:    make(map[string]bool),
		visitorSessions:  make(map[string]time.Time),
		sessionDurations: make([]float64, 0, 1000),
		bounceIPs:        make(map[string]bool),
		pageViews:        make(map[string]uint64),
		referrers:        make(map[string]uint64),
		userAgents:       make(map[string]uint64),
		visitorsByDay:    make(map[string]uint64),
		newVsReturning:   make(map[string]bool),
		visitTimes:       make(map[string][]time.Time),
		lastVisitorEntry: make([]VisitorEntry, 100), // Keep last
	}
}
