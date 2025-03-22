package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/magooney-loon/webrender/internal/admin/session"
)

// UserContextKey is the key used to store user information in the request context
type UserContextKey string

const (
	// UserKey is the context key for the user
	UserKey UserContextKey = "user"

	// RoleKey is the context key for the user role
	RoleKey UserContextKey = "role"
)

// RequireAdminAuth is middleware that ensures the user is authenticated as admin
func RequireAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is authenticated via session
		if !session.IsAuthenticated(r) {
			// If this is an API request, return 401
			if strings.HasPrefix(r.URL.Path, "/_/api/") {
				http.Error(w, "Unauthorized access", http.StatusUnauthorized)
				return
			}

			// Otherwise redirect to login page
			http.Redirect(w, r, "/_/login", http.StatusFound)
			return
		}

		// Get user role from session
		role := session.GetUserRole(r)

		// Check if user has admin role
		if role != "admin" && role != "super-admin" {
			// If API request, return 403
			if strings.HasPrefix(r.URL.Path, "/_/api/") {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			// Otherwise redirect to login
			http.Redirect(w, r, "/_/login?error=insufficient_permissions", http.StatusFound)
			return
		}

		// Set user context
		username := session.GetUsername(r)
		ctx := context.WithValue(r.Context(), UserKey, username)
		ctx = context.WithValue(ctx, RoleKey, role)

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the username from the request context
func GetUserFromContext(r *http.Request) string {
	if user, ok := r.Context().Value(UserKey).(string); ok {
		return user
	}
	return ""
}

// GetRoleFromContext retrieves the user role from the request context
func GetRoleFromContext(r *http.Request) string {
	if role, ok := r.Context().Value(RoleKey).(string); ok {
		return role
	}
	return ""
}
