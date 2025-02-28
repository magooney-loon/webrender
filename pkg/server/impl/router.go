package impl

import (
	"net/http"
	"strings"
	"sync"

	types "github.com/magooney-loon/webserver/types/server"
)

type router struct {
	routes     map[string]http.HandlerFunc
	groups     map[string]*routeGroup
	middleware []func(http.Handler) http.Handler
	mu         sync.RWMutex
}

func newRouter() *router {
	return &router{
		routes: make(map[string]http.HandlerFunc),
		groups: make(map[string]*routeGroup),
	}
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find handler in routes or groups
	handler := r.findHandler(req.Method, req.URL.Path)
	if handler == nil {
		http.NotFound(w, req)
		return
	}

	// Apply middleware in reverse order
	var h http.Handler = http.HandlerFunc(handler)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}

	h.ServeHTTP(w, req)
}

func (r *router) Handle(method, path string, handler http.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[r.buildKey(method, path)] = handler
}

func (r *router) Group(prefix string) types.RouteGroup {
	r.mu.Lock()
	defer r.mu.Unlock()

	g := &routeGroup{
		router: r,
		prefix: prefix,
	}
	r.groups[prefix] = g
	return g
}

func (r *router) Use(middleware ...func(http.Handler) http.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware = append(r.middleware, middleware...)
}

func (r *router) findHandler(method, path string) http.HandlerFunc {
	// Check direct routes first
	if handler, ok := r.routes[r.buildKey(method, path)]; ok {
		return handler
	}

	// Check groups
	for prefix, group := range r.groups {
		if strings.HasPrefix(path, prefix) {
			if handler := group.findHandler(method, strings.TrimPrefix(path, prefix)); handler != nil {
				return handler
			}
		}
	}

	return nil
}

func (r *router) buildKey(method, path string) string {
	return strings.ToUpper(method) + ":" + path
}

type routeGroup struct {
	router     *router
	prefix     string
	middleware []func(http.Handler) http.Handler
}

func (g *routeGroup) Handle(method, path string, handler http.HandlerFunc) {
	fullPath := g.prefix + "/" + strings.TrimPrefix(path, "/")

	// Apply group middleware
	var h http.Handler = http.HandlerFunc(handler)
	for i := len(g.middleware) - 1; i >= 0; i-- {
		h = g.middleware[i](h)
	}

	g.router.Handle(method, fullPath, h.ServeHTTP)
}

func (g *routeGroup) Group(prefix string) types.RouteGroup {
	return g.router.Group(g.prefix + "/" + strings.TrimPrefix(prefix, "/"))
}

func (g *routeGroup) Use(middleware ...func(http.Handler) http.Handler) {
	g.middleware = append(g.middleware, middleware...)
}

func (g *routeGroup) findHandler(method, path string) http.HandlerFunc {
	return g.router.routes[g.router.buildKey(method, path)]
}
