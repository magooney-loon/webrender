package handlers

import (
	"context"
	"net/http"

	"github.com/magooney-loon/webserver/internal/core/middleware"
	"github.com/magooney-loon/webserver/internal/core/template"
	"github.com/magooney-loon/webserver/pkg/logger"
)

// AuthHandlers contains handlers for authentication
type AuthHandlers struct {
	log          *logger.Logger
	engine       *template.Engine
	username     string
	password     string
	systemPrefix string
}

// NewAuthHandlers creates a new instance of AuthHandlers
func NewAuthHandlers(log *logger.Logger, engine *template.Engine, username, password, systemPrefix string) *AuthHandlers {
	return &AuthHandlers{
		log:          log,
		engine:       engine,
		username:     username,
		password:     password,
		systemPrefix: systemPrefix,
	}
}

// HandleLogin handles the login form and authentication
func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// For GET requests, just display the login form
	if r.Method == http.MethodGet {
		h.engine.RenderLogin(w, map[string]interface{}{
			"Title": "Login",
		})
		return
	}

	// For POST requests, process the login form
	if r.Method == http.MethodPost {
		// Parse form
		err := r.ParseForm()
		if err != nil {
			h.log.Error("error parsing login form", map[string]interface{}{
				"error": err.Error(),
			})
			h.engine.RenderLogin(w, map[string]interface{}{
				"Title": "Login",
				"Error": "Error processing the form. Please try again.",
			})
			return
		}

		// Get form values
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Validate credentials
		if middleware.ValidateCredentials(username, password, h.username, h.password) {
			// Create session and set cookie
			middleware.CreateSession(w, username, 86400) // 24 hours

			// Add user to context
			ctx := context.WithValue(r.Context(), middleware.UserContextKey, &middleware.User{Username: username})
			r = r.WithContext(ctx)

			// Redirect to admin dashboard or other protected area
			http.Redirect(w, r, "/system/admin", http.StatusSeeOther)
			return
		}

		// Invalid credentials
		h.engine.RenderLogin(w, map[string]interface{}{
			"Title": "Login",
			"Error": "Invalid username or password.",
		})
		return
	}

	// Method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleLogout handles the logout request
func (h *AuthHandlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Delete the session
	middleware.DeleteSession(w, r)

	// Redirect to login page
	http.Redirect(w, r, h.systemPrefix+"/login", http.StatusSeeOther)
}
