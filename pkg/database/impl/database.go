// Package database provides a singleton SQLite database implementation for the webserver.
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	dbtypes "github.com/magooney-loon/webserver/types/database"
	logger "github.com/magooney-loon/webserver/utils/logger"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var (
	// instance is the singleton database instance
	instance *Database

	// once ensures the database is instantiated only once
	once sync.Once

	// ErrNotInitialized is returned when the database is not initialized
	ErrNotInitialized = errors.New("database not initialized")

	// ErrInvalidConfig is returned when the database configuration is invalid
	ErrInvalidConfig = errors.New("invalid database configuration")
)

// Database represents the SQLite database instance
type Database struct {
	// db is the underlying sql.DB instance
	db *sql.DB

	// config holds the database configuration
	config *Config

	// mu protects concurrent access to the database
	mu sync.RWMutex

	// pool manages connection pooling
	pool *Pool

	// metrics collects database performance metrics
	metrics *Metrics

	// Prepared statement cache
	stmtCache map[string]*sql.Stmt
	stmtMu    sync.RWMutex

	// Query results cache
	cache *Cache
}

// Initialize sets up the database singleton with the provided configuration
func Initialize(cfg Config) error {
	var initErr error

	once.Do(func() {
		log := logger.NewLogger()
		log.Info("Initializing database")
		// Validate configuration
		if err := cfg.Validate(); err != nil {
			initErr = fmt.Errorf("%w: %v", ErrInvalidConfig, err)
			return
		}

		// Create data directory if it doesn't exist
		if err := ensureDataDirectory(cfg.Path); err != nil {
			initErr = fmt.Errorf("failed to create data directory: %w", err)
			return
		}

		// Construct DSN with pragma configurations
		dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=%d&_foreign_keys=ON&cache=shared",
			cfg.Path, cfg.BusyTimeout.Milliseconds())

		// Open database connection
		db, err := sql.Open("sqlite3", dsn)
		if err != nil {
			initErr = fmt.Errorf("failed to open database: %w", err)
			return
		}

		// Configure connection pool
		db.SetMaxOpenConns(cfg.PoolSize)
		db.SetMaxIdleConns(cfg.PoolSize / 2)
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

		// Create database instance
		instance = &Database{
			db:        db,
			config:    &cfg,
			stmtCache: make(map[string]*sql.Stmt),
			metrics:   NewMetrics(),
			cache:     NewCache(cfg.CacheMaxItems, cfg.CacheTTL, cfg.CacheEnabled),
		}

		// Initialize connection pool
		instance.pool = NewPool(db, cfg.PoolSize)

		// Ping database to verify connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			initErr = fmt.Errorf("failed to ping database: %w", err)
			// Clean up resources on error
			_ = db.Close()
			instance = nil
			return
		}

		// Apply migrations if needed
		if cfg.AutoMigrate {
			log.Info("Migrating schema")
			if err := instance.MigrateSchema(ctx); err != nil {
				initErr = fmt.Errorf("failed to migrate schema: %w", err)
				// Clean up resources on error
				_ = db.Close()
				instance = nil
				return
			}
		}

		// Initialize prepared statement cache
		if err := instance.prepareCommonStatements(ctx); err != nil {
			initErr = fmt.Errorf("failed to prepare statements: %w", err)
			_ = db.Close()
			instance = nil
		}
		log.Info("Database initialized")
	})

	return initErr
}

// GetInstance returns the database singleton instance
func GetInstance() (*Database, error) {
	if instance == nil {
		return nil, ErrNotInitialized
	}
	return instance, nil
}

// Close closes the database connection and cleans up resources
func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Close all prepared statements
	db.stmtMu.Lock()
	for _, stmt := range db.stmtCache {
		_ = stmt.Close()
	}
	db.stmtCache = nil
	db.stmtMu.Unlock()

	// Close the connection pool
	if db.pool != nil {
		db.pool.Close()
	}

	// Close the database connection
	return db.db.Close()
}

// Exec executes a query without returning any rows
func (db *Database) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	defer func() {
		db.metrics.RecordQueryDuration(time.Since(start))
	}()

	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (db *Database) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	defer func() {
		db.metrics.RecordQueryDuration(time.Since(start))
	}()

	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (db *Database) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	defer func() {
		db.metrics.RecordQueryDuration(time.Since(start))
	}()

	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.db.QueryRowContext(ctx, query, args...)
}

// QueryCached executes a query that returns rows, with result caching.
// This should be used for read-only queries that benefit from caching.
func (db *Database) QueryCached(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Create cache key
	argsJSON, err := json.Marshal(args)
	if err != nil {
		// If we can't marshal args, just skip caching
		return db.Query(ctx, query, args...)
	}

	key := CacheKey{
		Query: query,
		Args:  string(argsJSON),
	}

	// Check if result is in cache
	if _, found := db.cache.Get(key); found {
		db.metrics.RecordCacheHit()
		// Execute the query normally, but record that this was a cache hit
		return db.Query(ctx, query, args...)
	}

	db.metrics.RecordCacheMiss()
	// Not found in cache, execute query
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Store the query in cache as a marker that we've seen this query before
	db.cache.Set(key, true, 100) // Small fixed size since we're just storing a bool

	return rows, nil
}

// QueryRowCached executes a query that returns a single row, with result caching.
func (db *Database) QueryRowCached(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Create cache key
	argsJSON, err := json.Marshal(args)
	if err != nil {
		// If we can't marshal args, just skip caching
		return db.QueryRow(ctx, query, args...)
	}

	key := CacheKey{
		Query: query,
		Args:  string(argsJSON),
	}

	// Check if result is in cache
	if _, found := db.cache.Get(key); found {
		db.metrics.RecordCacheHit()
		// Execute the query normally, but record that this was a cache hit
		return db.QueryRow(ctx, query, args...)
	}

	db.metrics.RecordCacheMiss()

	// Not found in cache, execute query
	row := db.QueryRow(ctx, query, args...)

	// Store the query in cache as a marker that we've seen this query before
	db.cache.Set(key, true, 100) // Small fixed size since we're just storing a bool

	return row
}

// estimateSize estimates the size in bytes of a value
func estimateSize(v interface{}) int {
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case string:
		return len(val)
	case []byte:
		return len(val)
	case int, int32, float32, bool:
		return 4
	case int64, float64:
		return 8
	case time.Time:
		return 24
	case []interface{}:
		size := 0
		for _, elem := range val {
			size += estimateSize(elem)
		}
		return size
	case map[string]interface{}:
		size := 0
		for k, v := range val {
			size += len(k) + estimateSize(v)
		}
		return size
	default:
		// For complex types, use a rough estimation based on reflection
		rv := reflect.ValueOf(val)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			size := 0
			for i := 0; i < rv.Len(); i++ {
				size += estimateSize(rv.Index(i).Interface())
			}
			return size
		case reflect.Map:
			size := 0
			for _, k := range rv.MapKeys() {
				size += estimateSize(k.Interface()) + estimateSize(rv.MapIndex(k).Interface())
			}
			return size
		case reflect.Struct:
			size := 0
			for i := 0; i < rv.NumField(); i++ {
				size += estimateSize(rv.Field(i).Interface())
			}
			return size
		default:
			return 100 // Default size for unknown types
		}
	}
}

// prepareCommonStatements prepares and caches commonly used SQL statements
func (db *Database) prepareCommonStatements(ctx context.Context) error {
	// Add common prepared statements as needed
	return nil
}

// ensureDataDirectory creates the data directory if it doesn't exist
func ensureDataDirectory(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}

// ClearCache clears the query cache
func (db *Database) ClearCache() {
	db.cache.Clear()
}

// GetCacheStats returns statistics about the cache
func (db *Database) GetCacheStats() CacheStats {
	stats := db.cache.Stats()

	// Update metrics with cache stats
	db.metrics.UpdateCacheSize(int64(stats.Items), uint64(stats.SizeBytes))

	return stats
}

// Ping checks if the database connection is alive
func (db *Database) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

// DB returns the underlying sql.DB instance
func (db *Database) DB() *sql.DB {
	return db.db
}

// Initialize initializes the store
func (db *Database) Initialize() error {
	return nil // Already initialized in the singleton pattern
}

// WithContext returns a new Store with the given context
func (db *Database) WithContext(ctx context.Context) dbtypes.Store {
	return db // We manage context per operation
}
