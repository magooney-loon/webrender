package security

import (
	"fmt"
	"net/http"

	types "github.com/magooney-loon/webserver/types/middleware"
)

type middleware struct {
	opts types.SecurityOptions
}

// New creates a new security headers middleware
func New(opts types.SecurityOptions) types.Middleware {
	return &middleware{opts: opts}
}

func (m *middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		if m.opts.HSTS {
			w.Header().Set("Strict-Transport-Security",
				fmt.Sprintf("max-age=%d; includeSubDomains", m.opts.HSTSMaxAge))
		}

		if m.opts.FrameOptions != "" {
			w.Header().Set("X-Frame-Options", m.opts.FrameOptions)
		} else {
			w.Header().Set("X-Frame-Options", "DENY")
		}

		if m.opts.ContentSecurity != "" {
			w.Header().Set("Content-Security-Policy", m.opts.ContentSecurity)
		} else {
			// Set default CSP
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
		}

		if m.opts.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", m.opts.ReferrerPolicy)
		} else {
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		}

		// Additional security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

		next.ServeHTTP(w, r)
	})
}
