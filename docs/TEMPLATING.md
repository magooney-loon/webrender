# Templating & API Guide

This document covers both the templating engine for HTML views and the REST API structure.

## Template Engine

Our template engine uses Go's `html/template` with custom functions and layouts.

### Core Types

```go
// Engine represents the template engine
type Engine struct {
    templates   *template.Template
    cache       map[string]*template.Template
    cacheMutex  sync.RWMutex
    log         *logger.Logger
    config      Config
}

// Config represents template engine configuration
type Config struct {
    TemplatesDir     string           // Root directory for templates
    LayoutsDir       string           // Directory for layout templates
    PartialsDir      string           // Directory for partial templates
    CacheTemplates   bool             // Enable template caching
    ReloadOnRequest  bool             // Reload templates on each request
    CustomFuncs      template.FuncMap  // Custom template functions
    DelimLeft        string           // Left delimiter (default: {{)
    DelimRight       string           // Right delimiter (default: }})
}

// CommonFuncs returns the default template functions
func CommonFuncs() template.FuncMap {
    return template.FuncMap{
        "formatDate": func(t time.Time) string {
            return t.Format("2006-01-02")
        },
        "markdown": func(s string) template.HTML {
            return template.HTML(blackfriday.Run([]byte(s)))
        },
        "json": func(v interface{}) string {
            b, _ := json.Marshal(v)
            return string(b)
        },
    }
}
```

### Directory Structure
```
web/
├── templates/
│   ├── layouts/
│   │   ├── base.html
│   │   └── dashboard.html
│   ├── partials/
│   │   ├── header.html
│   │   ├── footer.html
│   │   └── nav.html
│   └── pages/
│       ├── home.html
│       ├── login.html
│       └── dashboard/
│           ├── index.html
│           └── settings.html
└── static/
    ├── css/
    ├── js/
    └── images/
```

### Base Layout
```html
<!-- layouts/base.html -->
<!DOCTYPE html>
<html lang="en" data-theme="{{ .Theme }}">
<head>
    <meta charset="UTF-8">
    <title>{{ .Title }} - WebServer</title>
    <link rel="stylesheet" href="/static/css/main.css">
    {{ block "head" . }}{{ end }}
</head>
<body data-layout="flex col" data-gap="4">
    {{ template "partials/header" . }}
    <main data-layout="flex col" data-gap="8" data-p="4">
        {{ block "content" . }}{{ end }}
    </main>
    {{ template "partials/footer" . }}
    <script src="/static/js/main.js"></script>
    {{ block "scripts" . }}{{ end }}
</body>
</html>
```

## REST API Structure

Our API follows REST principles with versioning and consistent response formats.

### Core Types

```go
// Response represents the standard API response format
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *Error      `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

// Error represents API error details
type Error struct {
    Code    string `json:"code"`    // Machine-readable error code
    Message string `json:"message"` // Human-readable error message
}

// Meta contains pagination and other metadata
type Meta struct {
    Total   int `json:"total"`    // Total number of items
    Page    int `json:"page"`     // Current page number
    PerPage int `json:"per_page"` // Items per page
}

// QueryOptions represents common query parameters
type QueryOptions struct {
    Page     int
    PerPage  int
    SortBy   string
    SortDesc bool
    Search   string
    Filters  map[string]interface{}
}

// Handler represents a route handler with context and error handling
type Handler func(context.Context, *http.Request) (*Response, error)

// Middleware represents a function that wraps an HTTP handler
type Middleware func(http.Handler) http.Handler
```

### API Routes
```go
// RouteGroup defines a group of routes with shared prefix and middleware
type RouteGroup struct {
    Prefix     string                 // URL prefix for all routes
    Middleware []Middleware           // Middleware for all routes
    Routes     []Route               // Routes in this group
}

// Route defines a single HTTP route
type Route struct {
    Path        string               // URL path
    Method      string               // HTTP method
    Handler     Handler              // Request handler
    Middleware  []Middleware         // Route-specific middleware
    Description string               // Route description
}

// Example route group
api := server.RouteGroup{
    Prefix: "/api/v1",
    Middleware: []middleware.Middleware{
        middleware.CORS(),
        middleware.RateLimit(),
        middleware.Auth(),
    },
    Routes: []server.Route{
        {
            Path:    "/resources",
            Method:  http.MethodGet,
            Handler: handlers.ListResources,
        },
    },
}
```

### Error Handling
```go
// Standard error codes
const (
    ErrNotFound     = "NOT_FOUND"
    ErrBadRequest   = "BAD_REQUEST"
    ErrUnauthorized = "UNAUTHORIZED"
    ErrInternal     = "INTERNAL_ERROR"
)

// Error handling middleware
func ErrorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Error("panic recovered", map[string]interface{}{
                    "error": err,
                    "stack": debug.Stack(),
                })
                RespondError(w, http.StatusInternalServerError, &Error{
                    Code:    ErrInternal,
                    Message: "Internal server error",
                })
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

## WebSocket Support

For real-time features, we use WebSocket connections:

```go
// Client represents a WebSocket client connection
type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

// Hub maintains the set of active clients
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

// WebSocket handler
func HandleWebSocket(hub *Hub) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Error("websocket upgrade failed", err)
            return
        }
        
        client := &Client{
            hub:  hub,
            conn: conn,
            send: make(chan []byte, 256),
        }
        
        client.hub.register <- client
        
        go client.writePump()
        go client.readPump()
    }
}
``` 