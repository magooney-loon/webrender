package types

import "net/http"

// Middleware represents a middleware component
type Middleware interface {
	// Wrap wraps an http.Handler with middleware functionality
	Wrap(http.Handler) http.Handler
}

// Chain represents a chain of middleware
type Chain interface {
	// Add adds middleware to the chain
	Add(middleware ...Middleware) Chain

	// Then wraps the final handler with all middleware in the chain
	Then(http.Handler) http.Handler
}

// Options for various middleware types
type (
	// LoggerOptions configures the logger middleware
	LoggerOptions struct {
		Format     string
		TimeFormat string
		UTC        bool
	}

	// AuthOptions configures the auth middleware
	AuthOptions struct {
		Realm          string
		ExcludePaths   []string
		TokenValidator func(string) bool
	}

	// CORSOptions configures the CORS middleware
	CORSOptions struct {
		AllowedOrigins   []string
		AllowedMethods   []string
		AllowedHeaders   []string
		ExposedHeaders   []string
		AllowCredentials bool
		MaxAge           int
	}

	// SecurityOptions configures security headers middleware
	SecurityOptions struct {
		HSTS            bool
		HSTSMaxAge      int
		FrameOptions    string
		ContentSecurity string
		ReferrerPolicy  string
	}
)
