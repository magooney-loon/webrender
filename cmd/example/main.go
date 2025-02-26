package main

import (
	"encoding/json"
	"net/http"

	"github.com/magooney-loon/webserver/internal/config"
	"github.com/magooney-loon/webserver/internal/core/middleware"
	"github.com/magooney-loon/webserver/internal/core/server"
)

// Simple handler that returns a JSON message
func helloHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
}

// Custom middleware example that adds a header
func addCustomHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Use the new LoadWithOptions function with sensible defaults
	// Only specify what you need to override
	cfg, log := config.LoadWithOptions(
		// Override specific configuration options
		config.WithEnvironment("development"),
		config.WithServerHost("localhost"),
		config.WithServerPort(8080),
		config.WithSystemAPI(true, "/system"),
		config.WithAuthEnabled(true),
		config.WithAuthCredentials("admin", "secretpass"),
	)

	// Define API routes with group-level middleware
	apiGroup := server.RouteGroup{
		Prefix: "/api/v1",
		// Add middleware that applies to all routes in this group
		Middleware: []middleware.Middleware{
			addCustomHeader,
		},
		Routes: []server.Route{
			{
				Path:        "/example",
				Method:      http.MethodGet,
				Handler:     helloHandler,
				Description: "Simple hello world endpoint",
			},
			{
				Path:    "/secure",
				Method:  http.MethodGet,
				Handler: helloHandler,
				// Route-specific middleware for authentication
				Middleware: []middleware.Middleware{
					middleware.SessionAuth(middleware.AuthConfig{
						Enabled:      cfg.Security.Auth.Enabled,
						Username:     cfg.Security.Auth.Username,
						Password:     cfg.Security.Auth.Password,
						ExcludePaths: cfg.Security.Auth.ExcludePaths,
						CookieName:   "session_token",
						CookieMaxAge: 86400, // 24 hours
					}),
				},
				Description: "Secure endpoint requiring auth",
			},
		},
	}

	// Create server with routes and start it
	srv := server.New(cfg, log,
		server.WithRouteGroup(apiGroup),
		server.WithGlobalMiddleware(addCustomHeader),
	)

	// Start the server (this blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatal("server error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
