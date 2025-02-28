package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	database "github.com/magooney-loon/webserver/pkg/database/impl"
	dbtypes "github.com/magooney-loon/webserver/types/database"
	types "github.com/magooney-loon/webserver/types/server"
	logger "github.com/magooney-loon/webserver/utils/logger"
)

var (
	instance *server
	once     sync.Once
	dbConfig = database.DefaultConfig()
)

type server struct {
	srv       *http.Server
	router    *router
	config    *types.ServerConfig
	db        dbtypes.Store
	logger    types.Logger
	configMgr types.ConfigManager
	mu        sync.RWMutex
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

		// Initialize config manager
		configMgr, err := NewConfigManager(db)
		if err != nil {
			log.Fatal("Failed to initialize config manager", types.Fields{"error": err})
		}

		// Load server config from database
		cfg := &types.ServerConfig{}
		if err := loadServerConfig(context.Background(), configMgr, cfg); err != nil {
			log.Fatal("Failed to load server config", types.Fields{"error": err})
		}

		// Apply options after loading from database
		for _, opt := range options {
			opt(cfg)
		}

		r := newRouter()

		instance = &server{
			router:    r,
			config:    cfg,
			db:        db,
			logger:    log,
			configMgr: configMgr,
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
		system.Handle(http.MethodGet, "/config", instance.getConfigHandler)
		system.Handle(http.MethodPost, "/config", instance.setConfigHandler)

		// Watch for config changes
		go instance.watchConfigChanges(context.Background())

		log.Info("Server initialized", types.Fields{
			"port":           cfg.Port,
			"readTimeout":    cfg.ReadTimeout,
			"writeTimeout":   cfg.WriteTimeout,
			"maxHeaderBytes": cfg.MaxHeaderBytes,
		})
	})

	return instance
}

func loadServerConfig(ctx context.Context, configMgr types.ConfigManager, cfg *types.ServerConfig) error {
	// Load port
	if port, err := configMgr.GetInt(ctx, "server.port"); err == nil {
		cfg.Port = port
	} else {
		cfg.Port = 8080 // default
	}

	// Load timeouts
	if readTimeout, err := configMgr.GetInt(ctx, "server.read_timeout"); err == nil {
		cfg.ReadTimeout = readTimeout
	} else {
		cfg.ReadTimeout = 60 // default
	}

	if writeTimeout, err := configMgr.GetInt(ctx, "server.write_timeout"); err == nil {
		cfg.WriteTimeout = writeTimeout
	} else {
		cfg.WriteTimeout = 60 // default
	}

	// Load max header bytes
	if maxHeaderBytes, err := configMgr.GetInt(ctx, "server.max_header_bytes"); err == nil {
		cfg.MaxHeaderBytes = maxHeaderBytes
	} else {
		cfg.MaxHeaderBytes = 1 << 20 // default 1MB
	}

	// Load CORS settings
	if err := configMgr.GetJSON(ctx, "server.allowed_origins", &cfg.AllowedOrigins); err != nil {
		cfg.AllowedOrigins = []string{"*"} // default
	}

	if err := configMgr.GetJSON(ctx, "server.allowed_methods", &cfg.AllowedMethods); err != nil {
		cfg.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"} // default
	}

	if err := configMgr.GetJSON(ctx, "server.allowed_headers", &cfg.AllowedHeaders); err != nil {
		cfg.AllowedHeaders = []string{"Content-Type", "Authorization"} // default
	}

	return nil
}

func (s *server) watchConfigChanges(ctx context.Context) {
	keys := []string{
		"server.port",
		"server.read_timeout",
		"server.write_timeout",
		"server.max_header_bytes",
		"server.allowed_origins",
		"server.allowed_methods",
		"server.allowed_headers",
	}

	for _, key := range keys {
		ch, err := s.configMgr.Watch(ctx, key)
		if err != nil {
			s.logger.Error("Failed to watch config", types.Fields{"key": key, "error": err})
			continue
		}

		go func(key string, ch <-chan types.ConfigUpdate) {
			for update := range ch {
				if err := s.handleConfigUpdate(ctx, update); err != nil {
					s.logger.Error("Failed to handle config update", types.Fields{
						"key":   key,
						"error": err,
					})
				}
			}
		}(key, ch)
	}
}

func (s *server) handleConfigUpdate(ctx context.Context, update types.ConfigUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch update.Key {
	case "server.port":
		if port, err := strconv.Atoi(update.Value); err == nil {
			s.config.Port = port
			s.srv.Addr = fmt.Sprintf(":%d", port)
		}
	case "server.read_timeout":
		if timeout, err := strconv.Atoi(update.Value); err == nil {
			s.config.ReadTimeout = timeout
			s.srv.ReadTimeout = time.Duration(timeout) * time.Second
		}
	case "server.write_timeout":
		if timeout, err := strconv.Atoi(update.Value); err == nil {
			s.config.WriteTimeout = timeout
			s.srv.WriteTimeout = time.Duration(timeout) * time.Second
		}
	case "server.max_header_bytes":
		if bytes, err := strconv.Atoi(update.Value); err == nil {
			s.config.MaxHeaderBytes = bytes
			s.srv.MaxHeaderBytes = bytes
		}
	case "server.allowed_origins":
		var origins []string
		if err := json.Unmarshal([]byte(update.Value), &origins); err == nil {
			s.config.AllowedOrigins = origins
		}
	case "server.allowed_methods":
		var methods []string
		if err := json.Unmarshal([]byte(update.Value), &methods); err == nil {
			s.config.AllowedMethods = methods
		}
	case "server.allowed_headers":
		var headers []string
		if err := json.Unmarshal([]byte(update.Value), &headers); err == nil {
			s.config.AllowedHeaders = headers
		}
	}

	return nil
}

func (s *server) getConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

func (s *server) setConfigHandler(w http.ResponseWriter, r *http.Request) {
	var update struct {
		Key   string           `json:"key"`
		Value string           `json:"value"`
		Type  types.ConfigType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.configMgr.Set(r.Context(), update.Key, update.Value, update.Type); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
