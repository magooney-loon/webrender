package middleware

import "net/http"

// Default security header values
var defaultSecurityHeaders = map[string]string{
	"X-XSS-Protection":          "1; mode=block",
	"X-Content-Type-Options":    "nosniff",
	"X-Frame-Options":           "DENY",
	"Content-Security-Policy":   "default-src 'self'",
	"Referrer-Policy":           "strict-origin-when-cross-origin",
	"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
}

// NewSecurityHeaders creates a new security headers middleware
func NewSecurityHeaders(cfg Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set default headers if not configured
			if cfg.Security.Headers.XSSProtection == "" {
				w.Header().Set("X-XSS-Protection", defaultSecurityHeaders["X-XSS-Protection"])
			} else {
				w.Header().Set("X-XSS-Protection", cfg.Security.Headers.XSSProtection)
			}

			if cfg.Security.Headers.ContentTypeOptions == "" {
				w.Header().Set("X-Content-Type-Options", defaultSecurityHeaders["X-Content-Type-Options"])
			} else {
				w.Header().Set("X-Content-Type-Options", cfg.Security.Headers.ContentTypeOptions)
			}

			if cfg.Security.Headers.XFrameOptions == "" {
				w.Header().Set("X-Frame-Options", defaultSecurityHeaders["X-Frame-Options"])
			} else {
				w.Header().Set("X-Frame-Options", cfg.Security.Headers.XFrameOptions)
			}

			if cfg.Security.Headers.ContentSecurityPolicy == "" {
				w.Header().Set("Content-Security-Policy", defaultSecurityHeaders["Content-Security-Policy"])
			} else {
				w.Header().Set("Content-Security-Policy", cfg.Security.Headers.ContentSecurityPolicy)
			}

			if cfg.Security.Headers.ReferrerPolicy == "" {
				w.Header().Set("Referrer-Policy", defaultSecurityHeaders["Referrer-Policy"])
			} else {
				w.Header().Set("Referrer-Policy", cfg.Security.Headers.ReferrerPolicy)
			}

			if cfg.Security.Headers.StrictTransportSecurity == "" {
				w.Header().Set("Strict-Transport-Security", defaultSecurityHeaders["Strict-Transport-Security"])
			} else {
				w.Header().Set("Strict-Transport-Security", cfg.Security.Headers.StrictTransportSecurity)
			}

			if cfg.Security.Headers.PermissionsPolicy == "" {
				w.Header().Set("Permissions-Policy", defaultSecurityHeaders["Permissions-Policy"])
			} else {
				w.Header().Set("Permissions-Policy", cfg.Security.Headers.PermissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}
