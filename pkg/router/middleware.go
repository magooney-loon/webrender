// Package router provides middleware functionality using gorilla/handlers
// To use this package, first install gorilla/handlers:
// go get -u github.com/gorilla/handlers
package router

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
)

// StandardMiddleware adds a set of common middleware to a router
func StandardMiddleware(r *Router) *Router {
	return r.
		UseMiddleware(LoggingMiddleware).
		UseMiddleware(RecoveryMiddleware).
		UseMiddleware(CompressionMiddleware)
}

// LoggingMiddleware logs information about incoming requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return handlers.RecoveryHandler()(next)
}

// CompressionMiddleware compresses responses using gzip
func CompressionMiddleware(next http.Handler) http.Handler {
	return handlers.CompressHandler(next)
}

// CORSMiddleware adds Cross-Origin Resource Sharing headers
func CORSMiddleware(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return handlers.CORS(
			handlers.AllowedOrigins(origins),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
			handlers.AllowCredentials(),
		)(next)
	}
}

// CacheControlMiddleware adds cache control headers
func CacheControlMiddleware(maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age="+time.Duration(maxAge.Seconds()).String())
			next.ServeHTTP(w, r)
		})
	}
}
