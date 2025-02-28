package session

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	types "github.com/magooney-loon/webserver/types/middleware"
)

const (
	sessionCookie = "session_id"
	maxAge        = 86400 // 24 hours
)

type Session struct {
	ID        string
	Data      map[string]interface{}
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
	}
}

type middleware struct {
	store *Store
}

// New creates a new session middleware
func New(store *Store) types.Middleware {
	return &middleware{store: store}
}

func (m *middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var session *Session

		// Try to get existing session
		if cookie, err := r.Cookie(sessionCookie); err == nil {
			m.store.mu.RLock()
			session = m.store.sessions[cookie.Value]
			m.store.mu.RUnlock()

			// Check if session is expired
			if session != nil && time.Now().After(session.ExpiresAt) {
				m.store.mu.Lock()
				delete(m.store.sessions, cookie.Value)
				m.store.mu.Unlock()
				session = nil
			}
		}

		// Create new session if none exists
		if session == nil {
			session = &Session{
				ID:        uuid.New().String(),
				Data:      make(map[string]interface{}),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(maxAge * time.Second),
			}

			m.store.mu.Lock()
			m.store.sessions[session.ID] = session
			m.store.mu.Unlock()

			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookie,
				Value:    session.ID,
				Path:     "/",
				HttpOnly: true,
				Secure:   r.TLS != nil,
				MaxAge:   maxAge,
			})
		}

		// Store session in request context
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Cleanup removes expired sessions
func (s *Store) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}
