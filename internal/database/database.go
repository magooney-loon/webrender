// Package database provides a singleton SQLite database implementation for the webserver.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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
}

// Initialize sets up the database singleton with the provided configuration
func Initialize(cfg Config) error {
	var initErr error

	once.Do(func() {
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
