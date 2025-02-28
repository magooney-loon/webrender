package types

import (
	"context"
	"net/http"
)

// Server represents the core HTTP server interface
type Server interface {
	// Start starts the server with the given context
	Start(ctx context.Context) error
	// Stop gracefully stops the server
	Stop(ctx context.Context) error
	// Router returns the server's router
	Router() Router
}

// Router defines the interface for routing HTTP requests
type Router interface {
	// Handle registers a handler for a specific path and method
	Handle(method, path string, handler http.HandlerFunc)
	// Group creates a new route group with the given prefix
	Group(prefix string) RouteGroup
	// Use adds middleware to the router
	Use(middleware ...func(http.Handler) http.Handler)
}

// RouteGroup represents a group of routes with shared prefix and middleware
type RouteGroup interface {
	// Handle registers a handler within the group
	Handle(method, path string, handler http.HandlerFunc)
	// Group creates a nested route group
	Group(prefix string) RouteGroup
	// Use adds middleware specific to this group
	Use(middleware ...func(http.Handler) http.Handler)
}

// ServerOption represents a function that modifies server configuration
type ServerOption func(*ServerConfig)

// ServerConfig holds the configuration for the server
type ServerConfig struct {
	Port           int
	ReadTimeout    int
	WriteTimeout   int
	MaxHeaderBytes int
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	ExposedHeaders []string
	TrustedProxies []string
}
