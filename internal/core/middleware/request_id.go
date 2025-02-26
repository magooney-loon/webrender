package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	defaultRequestIDHeader = "X-Request-ID"
	defaultRequestIDKey    = contextKey("request_id")
)

// defaultGenerateID generates a new UUID for request tracking
func defaultGenerateID() string {
	return uuid.New().String()
}

// NewRequestID creates a new request ID middleware
func NewRequestID(cfg Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.RequestTracking.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			headerName := cfg.RequestTracking.HeaderName
			if headerName == "" {
				headerName = defaultRequestIDHeader
			}

			contextKey := cfg.RequestTracking.ContextKey
			if contextKey == "" {
				contextKey = string(defaultRequestIDKey)
			}

			generateID := cfg.RequestTracking.GenerateCustomID
			if generateID == nil {
				generateID = defaultGenerateID
			}

			// Try to get request ID from header first
			requestID := r.Header.Get(headerName)
			if requestID == "" {
				requestID = generateID()
			}

			// Set request ID in response header
			w.Header().Set(headerName, requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), contextKey, requestID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(defaultRequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
