# Authentication System

This document describes the authentication system implemented in the webserver.

## Core Types

```go
// AuthConfig represents authentication configuration
type AuthConfig struct {
    Enabled      bool       // Enable/disable authentication
    Username     string     // Admin username
    Password     string     // Admin password
    ExcludePaths []string  // Paths not requiring auth
    CookieName   string    // Session cookie name
    CookieMaxAge int       // Session cookie max age
}

// User represents an authenticated user
type User struct {
    Username string
}

// Context keys for auth data
const (
    UserContextKey contextKey = "user"
    SessionCookie  string     = "session_token"
)
```

## Components

### Middleware

The authentication system is implemented as middleware in `internal/core/middleware/auth.go`:

```go
// SessionAuth creates a middleware that performs session-based authentication
func SessionAuth(config AuthConfig) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !config.Enabled {
                next.ServeHTTP(w, r)
                return
            }

            // Check if path is excluded
            for _, path := range config.ExcludePaths {
                if strings.HasPrefix(r.URL.Path, path) {
                    next.ServeHTTP(w, r)
                    return
                }
            }

            // Check for session cookie
            cookie, err := r.Cookie(config.CookieName)
            if err != nil || cookie.Value == "" {
                http.Redirect(w, r, "/system/login", http.StatusSeeOther)
                return
            }

            // Validate session
            user, valid := sessions[cookie.Value]
            if !valid {
                http.Redirect(w, r, "/system/login", http.StatusSeeOther)
                return
            }

            // Add user to context
            ctx := context.WithValue(r.Context(), UserContextKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// CreateSession creates a new session and sets the session cookie
func CreateSession(w http.ResponseWriter, username string, maxAge int) string {
    sessionToken := generateSessionToken()
    sessions[sessionToken] = &User{Username: username}

    http.SetCookie(w, &http.Cookie{
        Name:     SessionCookie,
        Value:    sessionToken,
        Path:     "/",
        MaxAge:   maxAge,
        HttpOnly: true,
        Secure:   false, // Set to true in production
    })

    return sessionToken
}

// DeleteSession removes a session and clears the cookie
func DeleteSession(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie(SessionCookie)
    if err == nil && cookie.Value != "" {
        delete(sessions, cookie.Value)
    }

    http.SetCookie(w, &http.Cookie{
        Name:     SessionCookie,
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        HttpOnly: true,
        Secure:   false, // Set to true in production
    })
}

// ValidateCredentials validates user credentials
func ValidateCredentials(username, password, expectedUsername, expectedPassword string) bool {
    return username == expectedUsername && password == expectedPassword
}
```

### Handlers

Authentication handlers in `internal/api/handlers/auth.go`:

```go
// LoginRequest represents login form data
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

// HandleLogin processes login requests
func HandleLogin(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // Render login form
        engine.Render(w, "login", nil)
        
    case http.MethodPost:
        var req LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            RespondError(w, http.StatusBadRequest, "Invalid request")
            return
        }

        if !ValidateCredentials(req.Username, req.Password, cfg.Auth.Username, cfg.Auth.Password) {
            RespondError(w, http.StatusUnauthorized, "Invalid credentials")
            return
        }

        // Create session
        CreateSession(w, req.Username, cfg.Auth.CookieMaxAge)
        
        RespondJSON(w, Response{
            Success: true,
            Data: map[string]string{
                "redirect": "/dashboard",
            },
        })
    }
}

// HandleLogout processes logout requests
func HandleLogout(w http.ResponseWriter, r *http.Request) {
    DeleteSession(w, r)
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}
```

## Configuration

Authentication settings can be configured via environment variables:

```env
# Enable/disable authentication
AUTH_ENABLED=true

# Set admin credentials
AUTH_USERNAME=admin
AUTH_PASSWORD=secure_password

# Exclude paths from authentication
AUTH_EXCLUDE_PATHS=/api/public,/docs,/system/login

# Session cookie settings
SESSION_COOKIE_NAME=session_token
SESSION_COOKIE_MAX_AGE=86400
SESSION_COOKIE_SECURE=true
SESSION_COOKIE_HTTP_ONLY=true
```

## Usage

### Protecting Routes

Apply the authentication middleware to routes or route groups:

```go
// Protect a route group
systemGroup := server.RouteGroup{
    Prefix: "/system",
    Middleware: []middleware.Middleware{
        middleware.SessionAuth(middleware.AuthConfig{
            Enabled:      true,
            Username:     "admin",
            Password:     "password",
            ExcludePaths: []string{"/system/login"},
        }),
    },
    Routes: []server.Route{
        // Protected routes
    },
}

// Apply to individual routes
route := server.Route{
    Path:    "/protected",
    Method:  http.MethodGet,
    Handler: protectedHandler,
    Middleware: []middleware.Middleware{
        middleware.SessionAuth(middleware.AuthConfig{...}),
    },
}
```

### Accessing User Information

Access user information from the request context:

```go
func protectedHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Use user.Username for personalized content
}
```

### Displaying User Information in Templates

When passing user data to templates, use the "Username" parameter:

```go
// In your handler:
engine.Render(w, "dashboard", map[string]interface{}{
    "Title":    "Dashboard",
    "Username": user.Username,
    "Status":   systemStatus,
})

// In template files (header.html, etc.):
{{if .Username}}
    <div class="user-info">Welcome {{.Username}}</div>
{{end}}
```

## Security Considerations

1. **Session Management**
   - Currently uses in-memory storage (replace with Redis/DB in production)
   - Implement secure session token generation
   - Add session expiration and rotation
   - Enable secure cookie flags in production

2. **Password Security**
   - Implement password hashing (bcrypt/argon2)
   - Add password complexity requirements
   - Implement rate limiting for login attempts
   - Add MFA support

3. **CSRF Protection**
   - Add CSRF tokens to forms
   - Validate tokens on POST/PUT/DELETE requests
   - Set SameSite cookie attribute

4. **Headers**
   - Set security headers (HSTS, CSP, etc.)
   - Configure CORS properly
   - Use TLS in production
``` 