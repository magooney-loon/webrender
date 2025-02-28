package cors

import (
	"net/http"
	"strconv"
	"strings"

	types "github.com/magooney-loon/webserver/types/middleware"
)

type middleware struct {
	opts types.CORSOptions
}

// New creates a new CORS middleware
func New(opts types.CORSOptions) types.Middleware {
	return &middleware{opts: opts}
}

func (m *middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight requests
		if r.Method == http.MethodOptions {
			m.handlePreflight(w, r)
			return
		}

		// Handle actual request
		m.setCORSHeaders(w, r)
		next.ServeHTTP(w, r)
	})
}

func (m *middleware) handlePreflight(w http.ResponseWriter, r *http.Request) {
	// Set standard CORS headers
	m.setCORSHeaders(w, r)

	// Handle preflight specific headers
	if reqMethod := r.Header.Get("Access-Control-Request-Method"); reqMethod != "" {
		// Check if method is allowed
		allowed := false
		for _, method := range m.opts.AllowedMethods {
			if method == reqMethod {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
		// Check if headers are allowed
		requestedHeaders := strings.Split(strings.ToLower(reqHeaders), ",")
		for _, header := range requestedHeaders {
			header = strings.TrimSpace(header)
			allowed := false
			for _, allowedHeader := range m.opts.AllowedHeaders {
				if strings.ToLower(allowedHeader) == header {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "Header not allowed", http.StatusForbidden)
				return
			}
		}
	}

	// Set preflight response headers
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.opts.AllowedMethods, ", "))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.opts.AllowedHeaders, ", "))
	if m.opts.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(m.opts.MaxAge))
	}

	w.WriteHeader(http.StatusNoContent)
}

func (m *middleware) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range m.opts.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if !allowed {
		return
	}

	// Set standard CORS headers
	w.Header().Set("Access-Control-Allow-Origin", origin)
	if m.opts.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	if len(m.opts.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(m.opts.ExposedHeaders, ", "))
	}
}
