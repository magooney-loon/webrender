package impl

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	database "github.com/magooney-loon/webserver/pkg/database/impl"
	types "github.com/magooney-loon/webserver/types/server"
	logger "github.com/magooney-loon/webserver/utils/logger"
)

var (
	instance *server
	once     sync.Once
	dbConfig = database.DefaultConfig()
)

type server struct {
	srv    *http.Server
	router *router
	config *types.ServerConfig
	db     *database.Database
	logger types.Logger
	mu     sync.RWMutex
}

// NewServer initializes the server singleton
func NewServer(options ...types.ServerOption) types.Server {
	once.Do(func() {
		// Initialize logger
		log := logger.NewLogger()

		// Initialize the database
		if err := database.Initialize(dbConfig); err != nil {
			log.Fatal("Failed to initialize database", types.Fields{"error": err})
		}

		db, err := database.GetInstance()
		if err != nil {
			log.Fatal("Failed to get database instance", types.Fields{"error": err})
		}

		cfg := &types.ServerConfig{
			Port:           8080,
			ReadTimeout:    60,
			WriteTimeout:   60,
			MaxHeaderBytes: 1 << 20, // 1MB
		}

		for _, opt := range options {
			opt(cfg)
		}

		r := newRouter()

		instance = &server{
			router: r,
			config: cfg,
			db:     db,
			logger: log,
		}

		instance.srv = &http.Server{
			Addr:           fmt.Sprintf(":%d", cfg.Port),
			Handler:        r,
			ReadTimeout:    time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(cfg.WriteTimeout) * time.Second,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
		}

		// Register system reserved endpoints
		system := r.Group("/system")
		system.Handle(http.MethodGet, "/dashboard", dashboardHandler)
		system.Handle(http.MethodGet, "/login", loginHandler)

		log.Info("Server initialized", types.Fields{
			"port":           cfg.Port,
			"readTimeout":    cfg.ReadTimeout,
			"writeTimeout":   cfg.WriteTimeout,
			"maxHeaderBytes": cfg.MaxHeaderBytes,
		})
	})

	return instance
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// GetInstance returns the server singleton
func GetInstance() types.Server {
	return instance
}

func (s *server) Start(ctx context.Context) error {
	s.logger.Info("Starting server", types.Fields{"addr": s.srv.Addr})

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server error", types.Fields{"error": err})
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("Context cancelled, stopping server")
		return s.Stop(ctx)
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
}

func (s *server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Stopping server")

	if err := s.db.Close(); err != nil {
		s.logger.Error("Failed to close database", types.Fields{"error": err})
	}

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown server", types.Fields{"error": err})
		return err
	}

	s.logger.Info("Server stopped")
	return nil
}

func (s *server) Router() types.Router {
	return s.router
}
