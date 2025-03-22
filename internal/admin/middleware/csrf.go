package middleware

import (
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/magooney-loon/webrender/internal/admin/session"
)

var (
	// CSRFMiddleware is the CSRF protection middleware
	CSRFMiddleware func(http.Handler) http.Handler
)

// InitCSRF initializes the CSRF protection middleware
func InitCSRF() error {
	// Generate a random key for CSRF protection
	csrfKey, err := session.GenerateRandomToken()
	if err != nil {
		return err
	}

	// Convert the token to a 32-byte key
	key := []byte(csrfKey)
	if len(key) > 32 {
		key = key[:32]
	}

	// Create the CSRF middleware
	CSRFMiddleware = csrf.Protect(
		key,
		csrf.Secure(true),                      // Require HTTPS
		csrf.Path("/"),                         // Apply to all paths
		csrf.SameSite(csrf.SameSiteStrictMode), // Strict SameSite policy
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "CSRF token validation failed", http.StatusForbidden)
		})),
	)

	return nil
}

// CSRFToken returns the CSRF token for the given request
func CSRFToken(r *http.Request) string {
	return csrf.Token(r)
}

// CSRFField returns an HTML field containing the CSRF token
func CSRFField(r *http.Request) template.HTML {
	return csrf.TemplateField(r)
}
