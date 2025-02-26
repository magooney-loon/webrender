package template

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/magooney-loon/webserver/internal/core/middleware"
)

// RenderPage renders a page with the default layout
func (e *Engine) RenderPage(w http.ResponseWriter, r *http.Request, pageTemplate string, data map[string]interface{}) error {
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// If data is nil, initialize it
	if data == nil {
		data = make(map[string]interface{})
	}

	// Add user data from context if available
	if user, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User); ok {
		data["User"] = user
	}

	// Add current page template name to the data for conditional rendering
	data["PageTemplate"] = pageTemplate

	// Also set Active property for navigation highlighting
	if _, ok := data["CurrentPage"]; !ok {
		data["CurrentPage"] = pageTemplate
	}

	// Log what we're about to render
	e.log.Info("rendering page", map[string]interface{}{
		"template":      pageTemplate,
		"defaultLayout": e.config.DefaultLayout,
		"development":   e.config.Development,
	})

	// If in development mode and reload on request is enabled, reload templates
	if e.config.Development && e.config.ReloadOnRequest {
		if err := e.ReloadTemplates(); err != nil {
			e.log.Error("failed to reload templates", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// First check if this is a direct template (no layout)
	if pageTemplate == "login" {
		// Login page doesn't use base layout
		if tmpl := e.templates.Lookup("login-content"); tmpl != nil {
			if err := e.templates.ExecuteTemplate(w, "login-content", data); err != nil {
				e.log.Error("login template render error", map[string]interface{}{
					"error": err.Error(),
				})
				return e.renderError(w, err)
			}
			return nil
		}
	}

	// For normal pages, use the base layout
	// Execute the base template - this will use the PageTemplate value to decide which content to show
	if err := e.templates.ExecuteTemplate(w, e.config.DefaultLayout, data); err != nil {
		e.log.Error("template render error", map[string]interface{}{
			"template": pageTemplate,
			"layout":   e.config.DefaultLayout,
			"error":    err.Error(),
		})
		return e.renderError(w, err)
	}

	return nil
}

// Helper function to render errors
func (e *Engine) renderError(w http.ResponseWriter, err error) error {
	// In development mode, show the error
	if e.config.Development {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// In production, show a generic error
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	return err
}

// RenderLogin renders the login page with the login layout
func (e *Engine) RenderLogin(w http.ResponseWriter, data map[string]interface{}) error {
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// If data is nil, initialize it
	if data == nil {
		data = make(map[string]interface{})
	}

	// Always set the title for the login page
	if _, exists := data["Title"]; !exists {
		data["Title"] = "Login"
	}

	// Use the template to render the login page
	err := e.templates.ExecuteTemplate(w, "login", data)
	if err != nil {
		e.log.Error("login template render error", map[string]interface{}{
			"error": err.Error(),
		})

		// In development mode, show the error
		if e.config.Development {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		// In production, show a generic error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	return nil
}

// RenderJSON renders JSON response
func (e *Engine) RenderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		e.log.Error("json encode error", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RenderError renders an error page
func (e *Engine) RenderError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	data := map[string]interface{}{
		"Title":   fmt.Sprintf("%d Error", status),
		"Status":  status,
		"Message": message,
	}

	// Try to render error template
	err := e.templates.ExecuteTemplate(w, "error", data)
	if err != nil {
		// If error template doesn't exist, render a simple error
		e.log.Error("error template render error", map[string]interface{}{
			"error": err.Error(),
		})

		fmt.Fprintf(w, "<html><body><h1>%d Error</h1><p>%s</p></body></html>", status, message)
	}
}
