package middleware

import (
	"net/http"
	"sync"
	"time"
)

// bucket represents a token bucket for rate limiting
type bucket struct {
	tokens    float64
	capacity  float64
	rate      float64
	lastCheck time.Time
	mu        sync.Mutex
}

// newBucket creates a new token bucket
func newBucket(capacity float64, window time.Duration) *bucket {
	return &bucket{
		tokens:    capacity,
		capacity:  capacity,
		rate:      float64(capacity) / float64(window.Seconds()),
		lastCheck: time.Now(),
	}
}

// allow checks if a request should be allowed and updates the bucket
func (b *bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens = min(b.capacity, b.tokens+elapsed*b.rate)
	b.lastCheck = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// rateLimitStore manages rate limit buckets
type rateLimitStore struct {
	buckets sync.Map
}

// getBucket gets or creates a bucket for the given key
func (s *rateLimitStore) getBucket(key string, rule RateLimitRule) *bucket {
	if b, ok := s.buckets.Load(key); ok {
		return b.(*bucket)
	}

	b := newBucket(float64(rule.Requests), rule.Window)
	s.buckets.Store(key, b)
	return b
}

// NewRateLimit creates a new rate limiting middleware
func NewRateLimit(cfg Config) Middleware {
	store := &rateLimitStore{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Security.RateLimit.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Get rate limit rule for the route
			rule := cfg.Security.RateLimit.Routes[r.URL.Path]
			if rule.Requests == 0 {
				// Use default rule if no route-specific rule exists
				rule = RateLimitRule{
					Requests: cfg.Security.RateLimit.Requests,
					Window:   cfg.Security.RateLimit.Window,
					Message:  "Rate limit exceeded",
				}
			}

			// Generate bucket key based on configuration
			var key string
			if cfg.Security.RateLimit.ByIP {
				key = r.RemoteAddr + ":" + r.URL.Path
			} else if cfg.Security.RateLimit.ByRoute {
				key = r.URL.Path
			} else {
				key = r.RemoteAddr
			}

			// Check rate limit
			bucket := store.getBucket(key, rule)
			if !bucket.allow() {
				message := rule.Message
				if message == "" {
					message = "Rate limit exceeded"
				}
				http.Error(w, message, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
