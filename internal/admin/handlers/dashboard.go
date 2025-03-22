package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/magooney-loon/webrender/internal/admin/components"
	"github.com/magooney-loon/webrender/internal/admin/middleware"
	"github.com/magooney-loon/webrender/internal/admin/session"
	"github.com/magooney-loon/webrender/pkg/state"
	tmpl "github.com/magooney-loon/webrender/pkg/template"
)

// RegisterAdminRoutes registers all admin dashboard routes
func RegisterAdminRoutes(r *mux.Router, sm *state.StateManager) {
	// Initialize session management
	session.Initialize()

	// Initialize CSRF protection
	if err := middleware.InitCSRF(); err != nil {
		log.Fatalf("Failed to initialize CSRF protection: %v", err)
	}

	// Create a subrouter for admin routes without authentication (login page)
	publicAdminRouter := r.PathPrefix("/_").Subrouter()

	// Apply CSRF middleware to all admin routes
	publicAdminRouter.Use(middleware.CSRFMiddleware)

	// Login page route
	publicAdminRouter.HandleFunc("/login", AdminLoginPageHandler).Methods("GET")
	publicAdminRouter.HandleFunc("/login", AdminLoginHandler).Methods("POST")

	// Logout route (doesn't need auth)
	publicAdminRouter.HandleFunc("/logout", AdminLogoutHandler).Methods("GET")

	// Create a subrouter for protected admin routes
	adminRouter := r.PathPrefix("/_").Subrouter()
	adminRouter.Use(middleware.RequireAdminAuth)

	// Register components
	dashboard := components.NewAdminDashboard("admin-dashboard")
	if err := sm.RegisterComponent(dashboard); err != nil {
		panic("Failed to register admin dashboard component: " + err.Error())
	}

	// Admin dashboard route
	adminRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Render the dashboard component
		dashboardHTML, err := sm.RenderComponent("admin-dashboard", map[string]interface{}{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Read client JS content
		clientJSPath := filepath.Join("pkg", "websocket", "client.js")
		clientJSContent, err := os.ReadFile(clientJSPath)
		if err != nil {
			http.Error(w, "Failed to load WebSocket client: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create page data
		data := tmpl.PageData{
			Title:    "Admin Dashboard",
			Content:  template.HTML(dashboardHTML),
			Styles:   template.CSS(components.GetDashboardStyles()),
			Scripts:  template.JS(components.GetDashboardScripts()),
			ClientJS: template.JS(clientJSContent),
		}

		// Render the page using base template
		if err := tmpl.GetBaseTemplate().Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Users management
	adminRouter.HandleFunc("/users", AdminUsersHandler).Methods("GET")

	// Settings page
	adminRouter.HandleFunc("/settings", AdminSettingsHandler).Methods("GET")

	// Analytics page
	adminRouter.HandleFunc("/analytics", AdminAnalyticsHandler).Methods("GET")
}

// AdminLoginPageHandler serves the login page
func AdminLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// If user is already authenticated, redirect to dashboard
	if session.IsAuthenticated(r) {
		http.Redirect(w, r, "/_/", http.StatusFound)
		return
	}

	// Get error message if any
	errorMsg := r.URL.Query().Get("error")
	errorHTML := ""
	if errorMsg == "invalid_credentials" {
		errorHTML = `<div class="bg-vercel-error bg-opacity-10 border border-vercel-error border-opacity-30 text-red-300 px-4 py-3 rounded mb-4" role="alert">
			<span class="font-bold">Error:</span> Invalid username or password
		</div>`
	} else if errorMsg == "insufficient_permissions" {
		errorHTML = `<div class="bg-vercel-error bg-opacity-10 border border-vercel-error border-opacity-30 text-red-300 px-4 py-3 rounded mb-4" role="alert">
			<span class="font-bold">Error:</span> You don't have permission to access the admin area
		</div>`
	}

	// Get CSRF token field
	csrfField := middleware.CSRFField(r)

	loginHTML := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<link rel="icon" href="/static/logo.svg" type="image/svg+xml">
		<title>Admin Login</title>
		<link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
		<style>
			/* Vercel dark theme styles */
			.vercel-card {
				background: rgba(32, 32, 36, 0.5);
				backdrop-filter: blur(10px);
				border: 1px solid rgba(63, 63, 70, 0.5);
				border-radius: 0.5rem;
				overflow: hidden;
				transition: all 0.2s ease-in-out;
			}
			.vercel-input {
				background: rgba(32, 32, 36, 0.5);
				border: 1px solid rgba(63, 63, 70, 0.5);
				border-radius: 0.375rem;
				color: #fff;
				font-size: 0.875rem;
				outline: none;
				padding: 0.625rem 0.75rem;
				transition: all 0.2s ease;
			}
			.vercel-input:focus {
				border-color: #0070f3;
				box-shadow: 0 0 0 2px rgba(0, 112, 243, 0.2);
			}
			.vercel-btn-primary {
				align-items: center;
				border-radius: 0.375rem;
				display: inline-flex;
				font-weight: 500;
				justify-content: center;
				outline: none;
				padding: 0.5rem 1rem;
				position: relative;
				transition: all 0.2s ease;
				white-space: nowrap;
				background: #0070f3;
				color: #fff;
			}
			.vercel-btn-primary:hover {
				background: #0761d1;
			}
		</style>
	</head>
	<body class="bg-black h-screen flex items-center justify-center">
		<div class="vercel-card p-8 w-96 text-white">
			<h1 class="text-2xl font-bold mb-6 text-center">Admin Login</h1>
			` + errorHTML + `
			<form method="POST" action="/_/login">
				` + string(csrfField) + `
				<div class="mb-4">
					<label class="block text-vercel-gray-300 text-sm font-bold mb-2" for="username">
						Username
					</label>
					<input class="vercel-input w-full" 
						id="username" name="username" type="text" placeholder="Username" required>
				</div>
				<div class="mb-6">
					<label class="block text-vercel-gray-300 text-sm font-bold mb-2" for="password">
						Password
					</label>
					<input class="vercel-input w-full" 
						id="password" name="password" type="password" placeholder="******************" required>
				</div>
				<div class="flex items-center justify-between">
					<button class="vercel-btn-primary w-full" 
						type="submit">
						Sign In
					</button>
				</div>
			</form>
		</div>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(loginHTML))
}

// AdminLoginHandler processes login form submissions
func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Check credentials (in a real app, this would use a database and proper password hashing)
	var role string
	if username == "admin" && password == "passpass" {
		role = "admin"
	} else if username == "superadmin" && password == "superpass" {
		role = "super-admin"
	} else {
		// Invalid credentials
		http.Redirect(w, r, "/_/login?error=invalid_credentials", http.StatusFound)
		return
	}

	// Create a session for the authenticated user
	err = session.CreateUserSession(w, r, username, role)
	if err != nil {
		http.Error(w, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard
	http.Redirect(w, r, "/_/", http.StatusFound)
}

// AdminLogoutHandler handles user logout
func AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session
	err := session.ClearSession(w, r)
	if err != nil {
		http.Error(w, "Failed to logout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to login page
	http.Redirect(w, r, "/_/login", http.StatusFound)
}

// AdminUsersHandler handles the admin users page
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Admin Users</h1><p>User management page (placeholder)</p>"))
}

// AdminSettingsHandler handles the admin settings page
func AdminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Admin Settings</h1><p>Settings page (placeholder)</p>"))
}

// AdminAnalyticsHandler handles the admin analytics page
func AdminAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Admin Analytics</h1><p>Analytics page (placeholder)</p>"))
}
