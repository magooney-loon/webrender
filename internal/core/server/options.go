package server

import (
	"net/http"

	"github.com/magooney-loon/webserver/internal/core/middleware"
)

// Route represents a single API route
type Route struct {
	Path        string
	Method      string
	Handler     http.HandlerFunc
	Middleware  []middleware.Middleware
	Description string
}

// RouteGroup represents a group of routes with shared middleware and prefix
type RouteGroup struct {
	Prefix     string
	Middleware []middleware.Middleware
	Routes     []Route
}

// RouterConfig holds all route configurations
type RouterConfig struct {
	Groups []RouteGroup
	Routes []Route // Global routes
}

// ServerOptions represents server configuration options
type ServerOptions struct {
	Middleware []middleware.Middleware
	Router     *RouterConfig
}

// Option is a function that configures server options
type Option func(*ServerOptions)

// WithRouteGroup adds a route group to the server
func WithRouteGroup(group RouteGroup) Option {
	return func(opts *ServerOptions) {
		if opts.Router == nil {
			opts.Router = &RouterConfig{}
		}
		opts.Router.Groups = append(opts.Router.Groups, group)
	}
}

// WithRoute adds a global route to the server
func WithRoute(route Route) Option {
	return func(opts *ServerOptions) {
		if opts.Router == nil {
			opts.Router = &RouterConfig{}
		}
		opts.Router.Routes = append(opts.Router.Routes, route)
	}
}

// WithGlobalMiddleware adds middleware to the server
func WithGlobalMiddleware(mw middleware.Middleware) Option {
	return func(opts *ServerOptions) {
		opts.Middleware = append(opts.Middleware, mw)
	}
}
