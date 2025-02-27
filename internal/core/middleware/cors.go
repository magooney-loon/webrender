package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// Default CORS settings
var (
	defaultAllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"}
	defaultAllowedHeaders = []string{
		"Accept",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"Authorization",
		"X-CSRF-Token",
		"X-Requested-With",
	}
)

// NewCORS creates a new CORS middleware
func NewCORS(cfg Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Security.CORS.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Set default methods if none specified
			allowedMethods := cfg.Security.CORS.AllowedMethods
			if len(allowedMethods) == 0 {
				allowedMethods = defaultAllowedMethods
			}

			// Set default headers if none specified
			allowedHeaders := cfg.Security.CORS.AllowedHeaders
			if len(allowedHeaders) == 0 {
				allowedHeaders = defaultAllowedHeaders
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				// Set CORS headers
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

				if cfg.Security.CORS.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.Security.CORS.MaxAge))
				}
			}

			// Set common CORS headers
			origin := r.Header.Get("Origin")
			if origin != "" {
				// Check if origin is allowed
				allowed := false
				if len(cfg.Security.CORS.AllowedOrigins) == 0 || contains(cfg.Security.CORS.AllowedOrigins, "*") {
					allowed = true
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					for _, allowedOrigin := range cfg.Security.CORS.AllowedOrigins {
						if allowedOrigin == origin {
							allowed = true
							w.Header().Set("Access-Control-Allow-Origin", origin)
							break
						}
					}
				}

				if !allowed {
					http.Error(w, "Origin not allowed", http.StatusForbidden)
					return
				}

				// Set other CORS headers
				if cfg.Security.CORS.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if len(cfg.Security.CORS.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.Security.CORS.ExposedHeaders, ", "))
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// contains checks if a string is present in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
