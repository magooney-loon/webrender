package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/magooney-loon/webserver/internal/config"
	"github.com/magooney-loon/webserver/internal/core/middleware"
	"github.com/magooney-loon/webserver/internal/monitoring"
	"github.com/magooney-loon/webserver/pkg/logger"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	cfg      *config.Config
	log      *logger.Logger
	srv      *http.Server
	wg       sync.WaitGroup
	shutdown chan struct{}
	metrics  *monitoring.Metrics
	router   *http.ServeMux
	mwChain  *middleware.Chain
	options  *ServerOptions
}

// New creates a new server instance with the provided configuration and logger
func New(cfg *config.Config, log *logger.Logger, opts ...Option) *Server {
	options := &ServerOptions{
		Router: &RouterConfig{},
	}

	// Apply options
	for _, opt := range opts {
		opt(options)
	}

	metrics := monitoring.NewMetrics()
	mwChain := setupMiddleware(cfg, log)

	s := &Server{
		cfg:      cfg,
		log:      log,
		shutdown: make(chan struct{}),
		metrics:  metrics,
		mwChain:  mwChain,
		options:  options,
		router:   http.NewServeMux(),
	}

	return s
}

// Start initializes and starts the HTTP server
func (s *Server) Start() error {
	s.setupRoutes()
	handler := s.applyMiddleware(s.router)

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.srv = &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    s.cfg.Server.ReadTimeout,
		WriteTimeout:   s.cfg.Server.WriteTimeout,
		MaxHeaderBytes: int(s.cfg.Server.MaxHeaderSize),
	}

	go s.startHTTPServer(addr)
	return s.waitForShutdown()
}

// startHTTPServer starts the HTTP server and logs any errors
func (s *Server) startHTTPServer(addr string) {
	s.log.Info("starting server", map[string]interface{}{
		"addr": addr,
		"env":  s.cfg.Environment,
	})

	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		s.log.Error("server error", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
}

// waitForShutdown waits for termination signals and performs a graceful shutdown
func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	s.log.Info("shutting down server", nil)

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %v", err)
	}

	s.wg.Wait()
	close(s.shutdown)

	return nil
}
