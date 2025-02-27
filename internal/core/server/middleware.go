package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/magooney-loon/webserver/internal/config"
	"github.com/magooney-loon/webserver/internal/core/middleware"
	"github.com/magooney-loon/webserver/pkg/logger"
)

// setupMiddleware configures the middleware chain based on application configuration
func setupMiddleware(cfg *config.Config, log *logger.Logger) *middleware.Chain {
	mwConfig := &middleware.Config{}

	// Configure logging middleware
	mwConfig.Logging.Enabled = true
	mwConfig.Logging.IncludeHeaders = true

	// Configure security settings
	mwConfig.Security.CORS = middleware.CORSConfig{
		Enabled:          cfg.Security.CORS.Enabled,
		AllowedOrigins:   cfg.Security.CORS.AllowedOrigins,
		AllowedMethods:   cfg.Security.CORS.AllowedMethods,
		AllowedHeaders:   cfg.Security.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.Security.CORS.ExposedHeaders,
		AllowCredentials: cfg.Security.CORS.AllowCredentials,
		MaxAge:           cfg.Security.CORS.MaxAge,
	}

	mwConfig.Security.Headers = middleware.SecurityHeadersConfig{
		XSSProtection:           cfg.Security.Headers.XSSProtection,
		ContentTypeOptions:      cfg.Security.Headers.ContentTypeOptions,
		XFrameOptions:           cfg.Security.Headers.XFrameOptions,
		ContentSecurityPolicy:   cfg.Security.Headers.ContentSecurityPolicy,
		ReferrerPolicy:          cfg.Security.Headers.ReferrerPolicy,
		StrictTransportSecurity: cfg.Security.Headers.StrictTransportSecurity,
		PermissionsPolicy:       cfg.Security.Headers.PermissionsPolicy,
	}

	mwConfig.Security.RateLimit = middleware.RateLimitConfig{
		Enabled:  cfg.Security.RateLimit.Enabled,
		Requests: cfg.Security.RateLimit.Requests,
		Window:   cfg.Security.RateLimit.Window,
		ByIP:     cfg.Security.RateLimit.ByIP,
		ByRoute:  cfg.Security.RateLimit.ByRoute,
		Routes:   make(map[string]middleware.RateLimitRule),
	}

	// Copy route rules
	for path, rule := range cfg.Security.RateLimit.Routes {
		mwConfig.Security.RateLimit.Routes[path] = middleware.RateLimitRule{
			Requests: rule.Requests,
			Window:   rule.Window,
			Message:  rule.Message,
		}
	}

	// Configure request tracking
	mwConfig.RequestTracking.Enabled = true
	mwConfig.RequestTracking.HeaderName = "X-Request-ID"

	return middleware.New(mwConfig)
}

// metricsMiddleware tracks request metrics
func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.metrics.IncrementActiveRequests()

		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		s.metrics.DecrementActiveRequests()
		s.metrics.RecordRequest(
			r.URL.Path,
			r.Method,
			wrapped.status,
			time.Since(start),
			r.ContentLength,
			wrapped.size,
		)

		// Track visitor analytics - only for successful page views and non-excluded paths
		if wrapped.status < 400 && r.Method == "GET" && !shouldExcludeFromAnalytics(r.URL.Path) {
			// Get client IP address, respecting X-Forwarded-For header if present
			clientIP := r.Header.Get("X-Forwarded-For")
			if clientIP == "" {
				clientIP = r.RemoteAddr
			}

			// Record the visit
			s.metrics.RecordVisit(
				r.URL.Path,
				clientIP,
				r.Header.Get("User-Agent"),
				r.Header.Get("Referer"),
			)
		}
	})
}

// shouldExcludeFromAnalytics checks if a path should be excluded from analytics tracking
func shouldExcludeFromAnalytics(path string) bool {
	// Exclude static files, API endpoints, and monitoring endpoints
	prefixes := []string{
		"/static/",
		"/api/",
		"/system/health", // Only exclude health check endpoint, not admin dashboard
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// applyMiddleware applies global middleware to the router
func (s *Server) applyMiddleware(handler http.Handler) http.Handler {
	s.mwChain.Use(
		middleware.NewLogging(s.log, *s.mwChain.Config),
		middleware.NewRequestID(*s.mwChain.Config),
		middleware.NewCORS(*s.mwChain.Config),
		middleware.NewRateLimit(*s.mwChain.Config),
		middleware.NewSecurityHeaders(*s.mwChain.Config),
		s.metricsMiddleware,
	)

	if len(s.options.Middleware) > 0 {
		s.mwChain.Use(s.options.Middleware...)
	}

	s.logMiddlewareStatus()
	return s.mwChain.Then(handler)
}

// logMiddlewareStatus logs the current middleware configuration
func (s *Server) logMiddlewareStatus() {
	cfg := s.mwChain.Config
	middlewareStatus := map[string]interface{}{
		"logging": map[string]interface{}{
			"enabled":         cfg.Logging.Enabled,
			"include_headers": cfg.Logging.IncludeHeaders,
			"skip_paths":      cfg.Logging.SkipPaths,
		},
		"request_tracking": map[string]interface{}{
			"enabled":     cfg.RequestTracking.Enabled,
			"header_name": cfg.RequestTracking.HeaderName,
		},
		"cors": map[string]interface{}{
			"enabled":           cfg.Security.CORS.Enabled,
			"allowed_origins":   cfg.Security.CORS.AllowedOrigins,
			"allowed_methods":   cfg.Security.CORS.AllowedMethods,
			"allow_credentials": cfg.Security.CORS.AllowCredentials,
		},
		"rate_limit": map[string]interface{}{
			"enabled":  cfg.Security.RateLimit.Enabled,
			"requests": cfg.Security.RateLimit.Requests,
			"window":   cfg.Security.RateLimit.Window.String(),
			"by_ip":    cfg.Security.RateLimit.ByIP,
			"by_route": cfg.Security.RateLimit.ByRoute,
		},
		"security_headers": map[string]interface{}{
			"xss_protection":  cfg.Security.Headers.XSSProtection != "",
			"content_type":    cfg.Security.Headers.ContentTypeOptions != "",
			"frame_options":   cfg.Security.Headers.XFrameOptions != "",
			"csp":             cfg.Security.Headers.ContentSecurityPolicy != "",
			"referrer_policy": cfg.Security.Headers.ReferrerPolicy != "",
			"hsts":            cfg.Security.Headers.StrictTransportSecurity != "",
			"permissions":     cfg.Security.Headers.PermissionsPolicy != "",
		},
	}

	s.log.Info("middleware configuration", middlewareStatus)
}
