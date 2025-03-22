package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	// SessionName is the name of the session cookie
	SessionName = "admin_session"

	// UserKey is the key used to store user data in the session
	UserKey = "user"

	// RoleKey is the key used to store role data in the session
	RoleKey = "role"

	// MaxAge defines how long the session cookie will last (in seconds)
	MaxAge = 3600 // 1 hour
)

var (
	// Store is the session store used for admin sessions
	Store *sessions.CookieStore
)

// Init initializes the session store with secure keys
func Init() error {
	// Load or generate secure keys
	hashKey, blockKey, err := LoadOrGenerateKeys()
	if err != nil {
		return err
	}

	Store = sessions.NewCookieStore(hashKey, blockKey)

	// Configure the session store
	Store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Requires HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   MaxAge,
	}

	return nil
}

// Get retrieves the session for the given request
func Get(r *http.Request) (*sessions.Session, error) {
	return Store.Get(r, SessionName)
}

// Save saves the session and sets the cookie in the response
func Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return session.Save(r, w)
}

// CreateUserSession creates a new session for the authenticated user
func CreateUserSession(w http.ResponseWriter, r *http.Request, username string, role string) error {
	session, err := Get(r)
	if err != nil {
		return err
	}

	// Set session values
	session.Values[UserKey] = username
	session.Values[RoleKey] = role

	// Generate a unique session ID
	sessionID, err := GenerateRandomToken()
	if err != nil {
		return err
	}
	session.ID = sessionID

	// Set session expiry
	session.Options.MaxAge = MaxAge

	// Save the session
	return Save(r, w, session)
}

// ClearSession removes the session data and expires the cookie
func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := Get(r)
	if err != nil {
		return err
	}

	// Clear session values
	session.Values = make(map[interface{}]interface{})

	// Expire the cookie
	session.Options.MaxAge = -1

	// Save the session (which will delete it)
	return Save(r, w, session)
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(r *http.Request) bool {
	session, err := Get(r)
	if err != nil {
		return false
	}

	// Check if user exists in session
	_, ok := session.Values[UserKey]
	return ok && !session.IsNew
}

// GetUserRole returns the role of the authenticated user
func GetUserRole(r *http.Request) string {
	session, err := Get(r)
	if err != nil {
		return ""
	}

	// Get role from session
	role, ok := session.Values[RoleKey].(string)
	if !ok {
		return ""
	}

	return role
}

// GetUsername returns the username of the authenticated user
func GetUsername(r *http.Request) string {
	session, err := Get(r)
	if err != nil {
		return ""
	}

	// Get username from session
	username, ok := session.Values[UserKey].(string)
	if !ok {
		return ""
	}

	return username
}
