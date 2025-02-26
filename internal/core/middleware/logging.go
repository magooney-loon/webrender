package middleware

import (
	"net/http"
	"time"

	"github.com/magooney-loon/webserver/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}

// NewLogging creates a new logging middleware
func NewLogging(log *logger.Logger, cfg Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Logging.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Skip logging for specified paths
			for _, path := range cfg.Logging.SkipPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Log request
			fields := map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"query":  r.URL.RawQuery,
				"ip":     r.RemoteAddr,
			}

			// Include headers if configured
			if cfg.Logging.IncludeHeaders {
				headers := make(map[string]string)
				if len(cfg.Logging.HeadersToLog) > 0 {
					for _, header := range cfg.Logging.HeadersToLog {
						if value := r.Header.Get(header); value != "" {
							headers[header] = value
						}
					}
				} else {
					for header := range r.Header {
						headers[header] = r.Header.Get(header)
					}
				}
				fields["headers"] = headers
			}

			log.Info("incoming request", fields)

			// Call next handler
			next.ServeHTTP(rw, r)

			// Log response
			duration := time.Since(start)
			fields["status"] = rw.status
			fields["size"] = rw.size
			fields["duration_ms"] = float64(duration.Nanoseconds()) / 1e6

			if rw.status >= 400 {
				log.Error("request error", fields)
			} else {
				log.Info("request completed", fields)
			}
		})
	}
}
