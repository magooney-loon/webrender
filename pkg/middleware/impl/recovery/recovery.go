package recovery

import (
	"net/http"
	"runtime/debug"

	types "github.com/magooney-loon/webserver/types/middleware"
	server "github.com/magooney-loon/webserver/types/server"
	logger "github.com/magooney-loon/webserver/utils/logger"
)

type middleware struct {
	// Optional error handler
	errorHandler func(w http.ResponseWriter, r *http.Request, err interface{})
	logger       server.Logger
}

// Option configures the recovery middleware
type Option func(*middleware)

// WithErrorHandler sets a custom error handler
func WithErrorHandler(handler func(w http.ResponseWriter, r *http.Request, err interface{})) Option {
	return func(m *middleware) {
		m.errorHandler = handler
	}
}

// New creates a new recovery middleware
func New(opts ...Option) types.Middleware {
	m := &middleware{
		errorHandler: defaultErrorHandler,
		logger:       logger.NewLogger(),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				stack := debug.Stack()
				m.logger.Error("Panic recovered in request handler", server.Fields{
					"error":  err,
					"stack":  string(stack),
					"path":   r.URL.Path,
					"method": r.Method,
				})

				// Call error handler
				m.errorHandler(w, r, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// defaultErrorHandler is the default handler for recovered panics
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error": "Internal Server Error"}`))
}
