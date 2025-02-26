package server

import (
	"net/http"
	"os"
	"strings"

	"github.com/magooney-loon/webserver/internal/api/handlers"
	"github.com/magooney-loon/webserver/internal/core/middleware"
	"github.com/magooney-loon/webserver/internal/core/template"
)

// setupRoutes configures all routes including groups and global routes
func (s *Server) setupRoutes() {
	// Initialize template engine with development flag based on environment
	isDev := os.Getenv("GO_ENV") == "development"
	templateConfig := template.DefaultConfig()
	templateConfig.Development = isDev
	templateConfig.ReloadOnRequest = isDev

	// Initialize the template engine
	engine, err := template.New(s.log, templateConfig)
	if err != nil {
		s.log.Error("failed to initialize template engine", map[string]interface{}{
			"error": err.Error(),
		})
		panic(err)
	}

	// Add auth routes
	authHandlers := handlers.NewAuthHandlers(s.log, engine, s.cfg.Security.Auth.Username, s.cfg.Security.Auth.Password, s.cfg.System.Prefix)

	// Register login/logout routes within the system prefix
	s.router.HandleFunc(s.cfg.System.Prefix+"/login", authHandlers.HandleLogin)
	s.router.HandleFunc(s.cfg.System.Prefix+"/logout", authHandlers.HandleLogout)

	// Initialize system handlers with the template engine
	systemHandlers := handlers.NewSystemHandlers(s.metrics, s.log, engine)

	// Setup system routes if enabled
	if s.cfg.System.Enabled {
		systemGroup := RouteGroup{
			Prefix: s.cfg.System.Prefix,
			// Add auth middleware to all system routes
			Middleware: []middleware.Middleware{
				middleware.SessionAuth(middleware.AuthConfig{
					Enabled:      s.cfg.Security.Auth.Enabled,
					Username:     s.cfg.Security.Auth.Username,
					Password:     s.cfg.Security.Auth.Password,
					ExcludePaths: s.cfg.Security.Auth.ExcludePaths,
					CookieName:   "session_token",
					CookieMaxAge: 86400, // 24 hours
				}),
			},
			Routes: []Route{
				{
					Path:        "/health",
					Method:      http.MethodGet,
					Handler:     systemHandlers.HandleHealth,
					Description: "Health check endpoint",
				},
				{
					Path:        "/admin",
					Method:      http.MethodGet,
					Handler:     systemHandlers.HandleAdmin,
					Description: "Admin endpoint",
				},
				{
					Path:        "/settings",
					Method:      http.MethodGet,
					Handler:     systemHandlers.HandleSettings,
					Description: "Settings page",
				},
				{
					Path:        "/settings",
					Method:      http.MethodPost,
					Handler:     systemHandlers.HandleSaveSettings,
					Description: "Save settings",
				},
				{
					Path:        "/settings/reset",
					Method:      http.MethodPost,
					Handler:     systemHandlers.HandleResetSettings,
					Description: "Reset settings to defaults",
				},
			},
		}
		s.registerRouteGroup(systemGroup)
	}

	// Register API route groups
	for _, group := range s.options.Router.Groups {
		s.registerRouteGroup(group)
	}

	// Register global routes
	for _, route := range s.options.Router.Routes {
		s.registerRoute(route)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	s.router.Handle("/static/", http.StripPrefix("/static/", fs))
}

// registerRouteGroup registers a group of routes with shared middleware
func (s *Server) registerRouteGroup(group RouteGroup) {
	// Create a handler map to track paths we've already registered
	handlerMap := make(map[string]map[string]http.HandlerFunc)

	// First, organize all handlers by path and method
	for _, route := range group.Routes {
		fullPath := group.Prefix + route.Path
		handler := route.Handler

		// Apply route-specific middleware
		if len(route.Middleware) > 0 {
			chain := middleware.New(s.mwChain.Config)
			chain.Use(route.Middleware...)
			handler = chain.Then(http.HandlerFunc(handler)).ServeHTTP
		}

		// Apply group middleware
		if len(group.Middleware) > 0 {
			chain := middleware.New(s.mwChain.Config)
			chain.Use(group.Middleware...)
			handler = chain.Then(http.HandlerFunc(handler)).ServeHTTP
		}

		// Initialize method map if needed
		if handlerMap[fullPath] == nil {
			handlerMap[fullPath] = make(map[string]http.HandlerFunc)
		}

		// Add this handler to our map
		handlerMap[fullPath][route.Method] = http.HandlerFunc(handler)

		s.log.Info("mapped route", map[string]interface{}{
			"path":        fullPath,
			"method":      route.Method,
			"description": route.Description,
		})
	}

	// Now register a single handler for each path that routes to the correct method
	for path, methodHandlers := range handlerMap {
		methodRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Look up the handler for this method
			if handler, ok := methodHandlers[r.Method]; ok {
				handler(w, r)
			} else {
				// Method not allowed
				w.Header().Set("Allow", getAllowedMethods(methodHandlers))
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		s.router.Handle(path, methodRouter)
		s.log.Info("registered path with method routing", map[string]interface{}{
			"path":    path,
			"methods": getAllowedMethods(methodHandlers),
		})
	}
}

// getAllowedMethods returns a comma-separated list of methods allowed for a path
func getAllowedMethods(methodHandlers map[string]http.HandlerFunc) string {
	methods := make([]string, 0, len(methodHandlers))
	for method := range methodHandlers {
		methods = append(methods, method)
	}
	return strings.Join(methods, ", ")
}

// registerRoute registers a single route
func (s *Server) registerRoute(route Route) {
	handler := route.Handler

	// Apply route-specific middleware
	if len(route.Middleware) > 0 {
		chain := middleware.New(s.mwChain.Config)
		chain.Use(route.Middleware...)
		handler = chain.Then(http.HandlerFunc(handler)).ServeHTTP
	}

	s.router.HandleFunc(route.Path, handler)
	s.log.Info("registered route", map[string]interface{}{
		"path":        route.Path,
		"method":      route.Method,
		"description": route.Description,
	})
}
