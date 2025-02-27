package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/magooney-loon/webserver/internal/config"
	"github.com/magooney-loon/webserver/internal/core/middleware"
	"github.com/magooney-loon/webserver/internal/core/server"
	"github.com/magooney-loon/webserver/internal/database"
	"github.com/magooney-loon/webserver/pkg/logger"
)

// /api/v1/user/cached?id=1
// /api/v1/user?id=1

// User represents a user in the database
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Simple handler that returns a JSON message
func helloHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
}

// Handler that demonstrates database access
func userHandler(w http.ResponseWriter, r *http.Request) {
	// Get the logger from config since we're not using request context yet
	_, log := config.LoadWithOptions()

	// Get database instance
	db, err := database.GetInstance()
	if err != nil {
		log.Error("failed to get database instance", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get user ID from request (simplified, would normally use path params)
	userID := r.URL.Query().Get("id")
	if userID == "" {
		userID = "1" // Default to first user if not specified
	}

	log.Info("looking up user", map[string]interface{}{
		"user_id": userID,
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Query the database
	var user User
	row := db.QueryRow(ctx, "SELECT id, username, email, created_at FROM users WHERE id = ?", userID)

	// Parse the result
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt); err != nil {
		log.Error("database query failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	log.Info("user found", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})

	// Return the user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Handler that demonstrates database access with caching
func cachedUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the logger from config since we're not using request context yet
	_, log := config.LoadWithOptions()

	// Get database instance
	db, err := database.GetInstance()
	if err != nil {
		log.Error("failed to get database instance", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get user ID from request (simplified, would normally use path params)
	userID := r.URL.Query().Get("id")
	if userID == "" {
		userID = "1" // Default to first user if not specified
	}

	log.Info("looking up user with caching", map[string]interface{}{
		"user_id": userID,
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Record time to measure performance
	startTime := time.Now()

	// First, get cache stats before the query
	cacheStatsBefore := db.GetCacheStats()

	// Query the database with caching
	row := db.QueryRowCached(ctx, "SELECT id, username, email, created_at FROM users WHERE id = ?", userID)

	// Parse the result
	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt); err != nil {
		log.Error("database query failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get cache statistics after the query
	cacheStatsAfter := db.GetCacheStats()

	// Calculate query duration
	queryDuration := time.Since(startTime)

	// Determine if this was a cache hit
	wasHit := cacheStatsAfter.Hits > cacheStatsBefore.Hits

	log.Info("user found (cached query)", map[string]interface{}{
		"user_id":       user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"duration_ms":   queryDuration.Milliseconds(),
		"was_cache_hit": wasHit,
		"cache_hits":    cacheStatsAfter.Hits,
		"cache_misses":  cacheStatsAfter.Misses,
	})

	// Create response with user data and cache stats
	response := struct {
		User           User    `json:"user"`
		QueryTimeMs    int64   `json:"query_time_ms"`
		WasCacheHit    bool    `json:"was_cache_hit"`
		CacheHits      uint64  `json:"cache_hits"`
		CacheMisses    uint64  `json:"cache_misses"`
		CacheHitRate   float64 `json:"cache_hit_rate"`
		CacheItemCount int     `json:"cache_item_count"`
	}{
		User:           user,
		QueryTimeMs:    queryDuration.Milliseconds(),
		WasCacheHit:    wasHit,
		CacheHits:      cacheStatsAfter.Hits,
		CacheMisses:    cacheStatsAfter.Misses,
		CacheHitRate:   cacheStatsAfter.HitRatio,
		CacheItemCount: cacheStatsAfter.Items,
	}

	// Return the enhanced response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler to clear the cache
func clearCacheHandler(w http.ResponseWriter, r *http.Request) {
	_, log := config.LoadWithOptions()

	// Get database instance
	db, err := database.GetInstance()
	if err != nil {
		log.Error("failed to get database instance", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get cache stats before clearing
	statsBefore := db.GetCacheStats()

	// Clear the cache
	db.ClearCache()

	// Get cache stats after clearing
	statsAfter := db.GetCacheStats()

	log.Info("cache cleared", map[string]interface{}{
		"items_before": statsBefore.Items,
		"items_after":  statsAfter.Items,
	})

	// Return success response
	response := map[string]interface{}{
		"success":      true,
		"items_before": statsBefore.Items,
		"items_after":  statsAfter.Items,
		"message":      "Cache cleared successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom middleware example that adds a header
func addCustomHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		next.ServeHTTP(w, r)
	})
}

// Setup database tables and initial data
func setupDatabase(log *logger.Logger) error {
	db, err := database.GetInstance()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("setting up database tables", nil)

	// Create users table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Check if we already have sample data
	var count int
	row := db.QueryRow(ctx, "SELECT COUNT(*) FROM users")
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("failed to check user count: %w", err)
	}

	// Insert sample data if needed
	if count == 0 {
		log.Info("inserting sample user data", nil)
		_, err = db.Exec(ctx, `
			INSERT INTO users (username, email, password_hash) VALUES 
			('alice', 'alice@example.com', 'password_hash_1'),
			('bob', 'bob@example.com', 'password_hash_2'),
			('charlie', 'charlie@example.com', 'password_hash_3')
		`)
		if err != nil {
			return fmt.Errorf("failed to insert sample data: %w", err)
		}
		log.Info("sample data inserted successfully", nil)
	} else {
		log.Info("database already contains users", map[string]interface{}{
			"count": count,
		})
	}

	return nil
}

func main() {
	// Use the new LoadWithOptions function with sensible defaults
	// Only specify what you need to override
	cfg, log := config.LoadWithOptions(
		// Override specific configuration options
		config.WithEnvironment("development"),
		config.WithServerHost("localhost"),
		config.WithServerPort(8080),
		config.WithSystemAPI(true, "/system"),
		config.WithAuthEnabled(true),
		config.WithAuthCredentials("admin", "secretpass"),
	)

	// Initialize the database with caching enabled
	log.Info("initializing database", map[string]interface{}{
		"path":          "./data/app.db",
		"pool_size":     5,
		"wal_mode":      true,
		"cache_enabled": true,
		"cache_ttl":     "5m",
	})

	dbConfig := database.DefaultConfig()
	dbConfig.Path = "./data/app.db" // Store in data directory
	dbConfig.PoolSize = 5
	dbConfig.WALMode = true
	dbConfig.AutoMigrate = false // Disable auto-migrations to prevent transaction conflicts

	// Enable query caching
	dbConfig.CacheEnabled = true
	dbConfig.CacheTTL = 5 * time.Minute
	dbConfig.CacheMaxItems = 1000

	if err := database.Initialize(dbConfig); err != nil {
		log.Fatal("database initialization failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Setup database schema and sample data
	if err := setupDatabase(log); err != nil {
		log.Fatal("database setup failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	log.Info("database initialized successfully", nil)

	// Ensure database is closed when application exits
	db, _ := database.GetInstance()
	defer func() {
		log.Info("closing database connection", nil)
		db.Close()
	}()

	// Define API routes with group-level middleware
	apiGroup := server.RouteGroup{
		Prefix: "/api/v1",
		// Add middleware that applies to all routes in this group
		Middleware: []middleware.Middleware{
			addCustomHeader,
		},
		Routes: []server.Route{
			{
				Path:        "/example",
				Method:      http.MethodGet,
				Handler:     helloHandler,
				Description: "Simple hello world endpoint",
			},
			{
				Path:        "/user",
				Method:      http.MethodGet,
				Handler:     userHandler,
				Description: "Get user by ID",
			},
			{
				Path:        "/user/cached",
				Method:      http.MethodGet,
				Handler:     cachedUserHandler,
				Description: "Get user by ID (with caching)",
			},
			{
				Path:        "/cache/clear",
				Method:      http.MethodPost,
				Handler:     clearCacheHandler,
				Description: "Clear the query cache",
			},
			{
				Path:    "/secure",
				Method:  http.MethodGet,
				Handler: helloHandler,
				// Route-specific middleware for authentication
				Middleware: []middleware.Middleware{
					middleware.SessionAuth(middleware.AuthConfig{
						Enabled:      cfg.Security.Auth.Enabled,
						Username:     cfg.Security.Auth.Username,
						Password:     cfg.Security.Auth.Password,
						ExcludePaths: cfg.Security.Auth.ExcludePaths,
						CookieName:   "session_token",
						CookieMaxAge: 86400, // 24 hours
					}),
				},
				Description: "Secure endpoint requiring auth",
			},
		},
	}

	// Create server with routes and start it
	srv := server.New(cfg, log,
		server.WithRouteGroup(apiGroup),
		server.WithGlobalMiddleware(addCustomHeader),
	)

	log.Info("starting server", map[string]interface{}{
		"host": cfg.Server.Host,
		"port": cfg.Server.Port,
	})

	// Start the server (this blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatal("server error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
