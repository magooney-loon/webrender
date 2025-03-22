package pkg

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/magooney-loon/webrender/internal/admin/handlers"
	"github.com/magooney-loon/webrender/pkg/component"
	"github.com/magooney-loon/webrender/pkg/router"
	"github.com/magooney-loon/webrender/pkg/state"
	tmpl "github.com/magooney-loon/webrender/pkg/template"
	"github.com/magooney-loon/webrender/pkg/websocket"
)

// WebRender is the main entry point for the WebRender library
type WebRender struct {
	// Core components
	StateManager      *state.StateManager
	ComponentRegistry *component.Registry
	WebSocketManager  *websocket.Manager

	// HTTP routing and handlers
	Router   *router.Router
	ServeMux *http.ServeMux // Kept for backward compatibility

	// Configuration
	StaticDir string

	// Client JavaScript content
	ClientJSContent string

	// Base template data
	BaseTemplate *template.Template
}

// Config contains configuration options for WebRender
type Config struct {
	// Directory for static files
	StaticDir string

	// HTTP handlers
	Router   *router.Router
	ServeMux *http.ServeMux

	// Admin panel
	EnableAdminPanel bool

	// Auto-register components directories
	AutoRegisterDirs []string

	// Auto-register component namespace
	AutoRegisterNamespace string

	// Base template configuration
	UseBaseTemplate bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		StaticDir:             "./static",
		ServeMux:              http.NewServeMux(),
		Router:                router.New().WithStrictSlash(true),
		EnableAdminPanel:      true,
		AutoRegisterDirs:      []string{"pkg/components"},
		AutoRegisterNamespace: "app",
		UseBaseTemplate:       true,
	}
}

// StandardMiddleware applies standard middleware to the router
func (wr *WebRender) StandardMiddleware() {
	router.StandardMiddleware(wr.Router)
}

// New creates a new WebRender instance
func New(config Config) (*WebRender, error) {
	// Create instance
	wr := &WebRender{
		StaticDir: config.StaticDir,
		ServeMux:  config.ServeMux,
		Router:    config.Router,
	}

	// Initialize state manager
	wr.StateManager = state.NewStateManager()

	// Get reference to component registry and WebSocket manager
	wr.ComponentRegistry = wr.StateManager.GetComponentRegistry()
	wr.WebSocketManager = wr.StateManager.GetWebSocketManager()

	// Store reference to base template
	wr.BaseTemplate = tmpl.GetBaseTemplate()

	// Read client JS content
	clientJSPath := filepath.Join("pkg", "websocket", "client.js")
	clientJSContent, err := os.ReadFile(clientJSPath)
	if err != nil {
		return nil, err
	}

	// Store client JS content
	wr.ClientJSContent = string(clientJSContent)

	// Register static file handler with Gorilla Mux
	wr.Router.RegisterStaticHandler(wr.StaticDir, "/static")

	// Apply standard middleware
	wr.StandardMiddleware()

	// Setup WebSocket handler on both ServeMux and Router
	wr.ServeMux.HandleFunc("/ws", wr.StateManager.HandleWebSocket)
	wr.Router.Router.HandleFunc("/ws", wr.StateManager.HandleWebSocket).Methods("GET")

	// Auto-register components if directories are specified
	if len(config.AutoRegisterDirs) > 0 {
		autoReg := component.NewAutoRegistration(wr.ComponentRegistry, config.AutoRegisterNamespace)
		for _, dir := range config.AutoRegisterDirs {
			err = autoReg.RegisterDirectory(dir)
			if err != nil {
				fmt.Printf("Warning: Auto-registration for directory %s failed: %v\n", dir, err)
			}
		}
	}

	// Register admin routes if enabled
	if config.EnableAdminPanel {
		handlers.RegisterAdminRoutes(wr.Router.Router, wr.StateManager)
	}

	return wr, nil
}

// RegisterComponent registers a component with WebRender
func (wr *WebRender) RegisterComponent(c *component.Component) error {
	return wr.StateManager.RegisterComponent(c)
}

// RenderComponent renders a component with props
func (wr *WebRender) RenderComponent(id string, props map[string]interface{}) (string, error) {
	return wr.StateManager.RenderComponent(id, props)
}

// ParseTemplate parses a template and registers it with the state manager
func (wr *WebRender) ParseTemplate(name, content string) error {
	return wr.StateManager.ParseString(name, content)
}

// RenderTemplate renders a template with data
func (wr *WebRender) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	return wr.StateManager.Render(w, name, data)
}

// HandleFunc registers an HTTP handler function
// Deprecated: Use Router.Router.HandleFunc instead
func (wr *WebRender) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	wr.ServeMux.HandleFunc(pattern, handler)
	// Also register with the router for backward compatibility
	wr.Router.Router.HandleFunc(pattern, handler)
}

// GetClientJS returns the WebSocket client JavaScript as template.JS
func (wr *WebRender) GetClientJS() template.JS {
	return template.JS(wr.ClientJSContent)
}

// Route adds a route with handler that automatically renders with the base template
func (wr *WebRender) Route(path string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	return wr.Router.Router.HandleFunc(path, handler)
}

// RouteWithTemplate adds a route that automatically renders content using the base template
func (wr *WebRender) RouteWithTemplate(path string, title string, getContentFn func() (template.HTML, error), getStylesFn func() template.CSS, getScriptsFn func() template.JS) *mux.Route {
	return wr.Router.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// Get the content HTML
		content, err := getContentFn()
		if err != nil {
			http.Error(w, "Failed to render content: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get styles and scripts
		var styles template.CSS
		var scripts template.JS

		if getStylesFn != nil {
			styles = getStylesFn()
		}

		if getScriptsFn != nil {
			scripts = getScriptsFn()
		}

		// Render the page with the base template
		wr.BaseTemplate.Execute(w, tmpl.PageData{
			Title:    title,
			Content:  content,
			Styles:   styles,
			Scripts:  scripts,
			ClientJS: wr.GetClientJS(),
		})
	})
}

// ComponentRoute adds a route that renders a specific component
func (wr *WebRender) ComponentRoute(path string, title string, componentID string, props map[string]interface{}, getStylesFn func() template.CSS, getScriptsFn func() template.JS) *mux.Route {
	return wr.RouteWithTemplate(path, title, func() (template.HTML, error) {
		html, err := wr.RenderComponent(componentID, props)
		return template.HTML(html), err
	}, getStylesFn, getScriptsFn)
}

// AutoRegisterComponents auto-registers components from a directory
func (wr *WebRender) AutoRegisterComponents(dir string, namespace string) error {
	autoReg := component.NewAutoRegistration(wr.ComponentRegistry, namespace)
	return autoReg.RegisterDirectory(dir)
}

// ServeHTTP implements the http.Handler interface
func (wr *WebRender) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use the router by default, which includes middleware support
	wr.Router.ServeHTTP(w, r)
}

// Start starts the HTTP server on the specified address
func (wr *WebRender) Start(addr string) error {
	fmt.Printf("Server starting at http://localhost%s\n", addr)
	fmt.Printf("Admin dashboard at http://localhost%s/_/\n", addr)
	return http.ListenAndServe(addr, wr)
}
