// Package router provides enhanced HTTP routing using gorilla/mux
// To use this package, first install gorilla/mux:
// go get -u github.com/gorilla/mux
package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Router extends gorilla/mux to provide additional functionality
type Router struct {
	*mux.Router
	middlewares []func(http.Handler) http.Handler
}

// New creates a new Router instance
func New() *Router {
	return &Router{
		Router:      mux.NewRouter(),
		middlewares: []func(http.Handler) http.Handler{},
	}
}

// WithStrictSlash sets the router's StrictSlash option
func (r *Router) WithStrictSlash(value bool) *Router {
	r.Router.StrictSlash(value)
	return r
}

// UseMiddleware adds middleware to the router
func (r *Router) UseMiddleware(middleware func(http.Handler) http.Handler) *Router {
	r.middlewares = append(r.middlewares, middleware)
	return r
}

// Group creates a new subrouter with the given path prefix
func (r *Router) Group(pathPrefix string) *Router {
	return &Router{
		Router:      r.Router.PathPrefix(pathPrefix).Subrouter(),
		middlewares: r.middlewares,
	}
}

// API creates a subrouter specifically for API endpoints
func (r *Router) API() *Router {
	return r.Group("/api")
}

// GetHandler returns an http.Handler with all middleware applied
func (r *Router) GetHandler() http.Handler {
	var handler http.Handler = r.Router

	// Apply middlewares in reverse order
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}

	return handler
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.GetHandler().ServeHTTP(w, req)
}
