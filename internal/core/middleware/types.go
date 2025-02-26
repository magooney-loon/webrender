package middleware

import (
	"net/http"
	"time"
)

// Config represents the unified configuration for all middleware
type Config struct {
	// Logging configuration
	Logging struct {
		Enabled        bool     // Enable/disable logging middleware
		IncludeHeaders bool     // Include request headers in logs
		HeadersToLog   []string // Specific headers to log (empty = all)
		SkipPaths      []string // Paths to skip logging
	}

	// Security configuration
	Security struct {
		// CORS settings
		CORS CORSConfig
		// Rate Limiting settings
		RateLimit RateLimitConfig
		// Security Headers
		Headers SecurityHeadersConfig
		// Auth settings
		Auth AuthConfig
	}

	// Request tracking configuration
	RequestTracking struct {
		Enabled          bool          // Enable/disable request tracking
		HeaderName       string        // Request ID header name
		ContextKey       string        // Context key for request ID
		GenerateCustomID func() string // Custom ID generator
	}
}

// CORSConfig defines CORS settings
type CORSConfig struct {
	Enabled          bool     // Enable/disable CORS
	AllowedOrigins   []string // Allowed origins (["*"] = all)
	AllowedMethods   []string // Allowed HTTP methods
	AllowedHeaders   []string // Allowed request headers
	ExposedHeaders   []string // Headers accessible to browser
	AllowCredentials bool     // Allow credentials (cookies, auth)
	MaxAge           int      // Preflight cache duration
}

// SecurityHeadersConfig defines security header settings
type SecurityHeadersConfig struct {
	XSSProtection           string // XSS protection header
	ContentTypeOptions      string // Content type options
	XFrameOptions           string // Frame options (clickjacking)
	ContentSecurityPolicy   string // CSP directives
	ReferrerPolicy          string // Referrer policy
	StrictTransportSecurity string // HSTS settings
	PermissionsPolicy       string // Permissions policy
}

// RateLimitConfig defines rate limiting settings
type RateLimitConfig struct {
	Enabled  bool                     // Enable/disable rate limiting
	Requests int                      // Max requests per window
	Window   time.Duration            // Time window for rate limiting
	ByIP     bool                     // Enable IP-based limiting
	ByRoute  bool                     // Enable route-based limiting
	Routes   map[string]RateLimitRule // Route-specific rules
}

// RateLimitRule defines rate limiting rules for specific routes
type RateLimitRule struct {
	Requests int           // Max requests allowed
	Window   time.Duration // Time window
	Message  string        // Custom rate limit message
}

// Middleware defines the standard middleware interface
type Middleware func(http.Handler) http.Handler

// Chain represents a chain of middleware
type Chain struct {
	middlewares []Middleware // Ordered list of middleware
	Config      *Config      // Middleware configuration
}

// New creates a new middleware chain with the given configuration
func New(config *Config) *Chain {
	return &Chain{
		middlewares: make([]Middleware, 0),
		Config:      config,
	}
}

// Use adds middleware to the chain
func (c *Chain) Use(middleware ...Middleware) {
	c.middlewares = append(c.middlewares, middleware...)
}

// Then wraps the final handler with all middleware in the chain
func (c *Chain) Then(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}
