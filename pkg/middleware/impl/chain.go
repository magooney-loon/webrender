package impl

import (
	"net/http"

	types "github.com/magooney-loon/webserver/types/middleware"
)

// chain implements the middleware.Chain interface
type chain struct {
	middlewares []types.Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...types.Middleware) types.Chain {
	return &chain{
		middlewares: middlewares,
	}
}

// Add adds middleware to the chain
func (c *chain) Add(middlewares ...types.Middleware) types.Chain {
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}

// Then wraps the final handler with all middleware in the chain
func (c *chain) Then(h http.Handler) http.Handler {
	// Apply middleware in reverse order so the first middleware
	// in the chain is the outermost wrapper
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i].Wrap(h)
	}
	return h
}
