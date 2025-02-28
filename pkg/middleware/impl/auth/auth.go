package auth

import (
	"context"
	"net/http"
	"strings"

	session "github.com/magooney-loon/webserver/pkg/middleware/impl/session"
	types "github.com/magooney-loon/webserver/types/middleware"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

type middleware struct {
	opts  types.AuthOptions
	roles map[string][]Role // map[userID][]Role
}

// New creates a new auth middleware
func New(opts types.AuthOptions) types.Middleware {
	return &middleware{
		opts:  opts,
		roles: make(map[string][]Role),
	}
}

// AddRole adds a role to a user
func (m *middleware) AddRole(userID string, role Role) {
	if roles, exists := m.roles[userID]; exists {
		for _, r := range roles {
			if r == role {
				return
			}
		}
		m.roles[userID] = append(roles, role)
	} else {
		m.roles[userID] = []Role{role}
	}
}

// HasRole checks if a user has a specific role
func (m *middleware) HasRole(userID string, role Role) bool {
	if roles, exists := m.roles[userID]; exists {
		for _, r := range roles {
			if r == role {
				return true
			}
		}
	}
	return false
}

func (m *middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is excluded
		for _, path := range m.opts.ExcludePaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Get auth token from header
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix if present
		token = strings.TrimPrefix(token, "Bearer ")

		// Validate token
		if !m.opts.TokenValidator(token) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Get user session
		session := r.Context().Value("session").(*session.Session)
		if session == nil {
			http.Error(w, "No session found", http.StatusUnauthorized)
			return
		}

		// Check user roles (if set in session)
		if userID, ok := session.Data["user_id"].(string); ok {
			if roles, exists := m.roles[userID]; exists {
				// Store roles in request context
				ctx := context.WithValue(r.Context(), "roles", roles)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Default to guest role if no roles found
		ctx := context.WithValue(r.Context(), "roles", []Role{RoleGuest})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
