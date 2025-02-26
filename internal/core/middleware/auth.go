package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// contextKey type already exists in middleware package
// type contextKey string

type AuthConfig struct {
	Enabled      bool
	Username     string
	Password     string
	ExcludePaths []string
	CookieName   string
	CookieMaxAge int
}

const (
	UserContextKey contextKey = "user"
	SessionCookie  string     = "session_token"
)

type User struct {
	Username string
}

// Active sessions (in a real app, you'd use a proper session store/database)
var sessions = make(map[string]*User)

// SessionAuth creates a middleware that performs session-based authentication
func SessionAuth(config AuthConfig) Middleware {
	cookieName := SessionCookie
	if config.CookieName != "" {
		cookieName = config.CookieName
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Debug info
			fmt.Printf("Request to path: %s\n", r.URL.Path)

			if !config.Enabled {
				fmt.Println("Auth disabled, proceeding to handler")
				next.ServeHTTP(w, r)
				return
			}

			// Check if path is excluded
			for _, path := range config.ExcludePaths {
				if strings.HasPrefix(r.URL.Path, path) {
					fmt.Printf("Path %s is excluded, skipping auth\n", r.URL.Path)
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check for session cookie
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				fmt.Printf("No session cookie found: %v\n", err)
				// Redirect to login page
				http.Redirect(w, r, "/system/login", http.StatusSeeOther)
				return
			}

			if cookie.Value == "" {
				fmt.Println("Session cookie value is empty")
				// Redirect to login page
				http.Redirect(w, r, "/system/login", http.StatusSeeOther)
				return
			}

			fmt.Printf("Found session cookie: %s\n", cookie.Value)

			// Validate session
			user, valid := sessions[cookie.Value]
			if !valid {
				fmt.Printf("Invalid session token: %s\n", cookie.Value)
				// Invalid session, redirect to login
				http.Redirect(w, r, "/system/login", http.StatusSeeOther)
				return
			}

			fmt.Printf("Valid session for user: %s\n", user.Username)

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CreateSession creates a new session and sets the session cookie
func CreateSession(w http.ResponseWriter, username string, maxAge int) string {
	// Generate a simple session token - in production use a more secure method
	sessionToken := generateSessionToken()

	fmt.Printf("Creating session for user %s with token %s\n", username, sessionToken)

	// Store in session map
	sessions[sessionToken] = &User{Username: username}

	// Print current sessions
	fmt.Println("Current active sessions:")
	for token, user := range sessions {
		fmt.Printf("- Token: %s, User: %s\n", token, user.Username)
	}

	// Set cookie with correct settings for development
	cookie := &http.Cookie{
		Name:     SessionCookie,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   false, // Set to false for development
	}

	http.SetCookie(w, cookie)
	fmt.Println("Set session cookie successfully")

	return sessionToken
}

// DeleteSession removes a session and clears the cookie
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookie)
	if err == nil && cookie.Value != "" {
		fmt.Printf("Deleting session with token: %s\n", cookie.Value)
		// Remove from sessions map
		delete(sessions, cookie.Value)
	}

	// Clear cookie with proper settings
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // Set to false for development
	})

	fmt.Println("Session cookie cleared")
}

// ValidateCredentials validates user credentials - replace with database lookup in production
func ValidateCredentials(username, password, expectedUsername, expectedPassword string) bool {
	return username == expectedUsername && password == expectedPassword
}

// generateSessionToken creates a new session token
// In production use a cryptographically secure method, e.g. UUID
func generateSessionToken() string {
	return "session_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
