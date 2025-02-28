package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds the database configuration
type Config struct {
	// Path is the file path to the SQLite database
	Path string

	// PoolSize is the maximum number of connections in the pool
	PoolSize int

	// BusyTimeout is the timeout for acquiring a lock
	BusyTimeout time.Duration

	// QueryTimeout is the timeout for query execution
	QueryTimeout time.Duration

	// ConnMaxLifetime is the maximum amount of time a connection may be reused
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle
	ConnMaxIdleTime time.Duration

	// CacheSize is the number of pages to cache in memory
	CacheSize int

	// WALMode enables Write-Ahead Logging mode
	WALMode bool

	// AutoMigrate automatically applies migrations on startup
	AutoMigrate bool

	// CacheEnabled enables in-memory caching
	CacheEnabled bool

	// CacheTTL is the time-to-live for cached entries
	CacheTTL time.Duration

	// CacheMaxItems is the maximum number of items in the cache
	CacheMaxItems int
}

// DefaultConfig returns the default database configuration
func DefaultConfig() Config {
	return Config{
		Path:            "./data/app.db",
		PoolSize:        10,
		BusyTimeout:     5 * time.Second,
		QueryTimeout:    30 * time.Second,
		ConnMaxLifetime: 1 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
		CacheSize:       2000,
		WALMode:         true,
		AutoMigrate:     false,
		CacheEnabled:    true,
		CacheTTL:        5 * time.Minute,
		CacheMaxItems:   1000,
	}
}

// LoadFromEnv loads database configuration from environment variables
func LoadFromEnv() Config {
	cfg := DefaultConfig()

	if path := os.Getenv("DB_PATH"); path != "" {
		cfg.Path = path
	}

	if poolSize := os.Getenv("DB_POOL_SIZE"); poolSize != "" {
		if size, err := strconv.Atoi(poolSize); err == nil && size > 0 {
			cfg.PoolSize = size
		}
	}

	if busyTimeout := os.Getenv("DB_BUSY_TIMEOUT"); busyTimeout != "" {
		if timeout, err := time.ParseDuration(busyTimeout); err == nil {
			cfg.BusyTimeout = timeout
		}
	}

	if queryTimeout := os.Getenv("DB_TIMEOUT"); queryTimeout != "" {
		if timeout, err := time.ParseDuration(queryTimeout); err == nil {
			cfg.QueryTimeout = timeout
		}
	}

	if cacheSize := os.Getenv("DB_CACHE_SIZE"); cacheSize != "" {
		if size, err := strconv.Atoi(cacheSize); err == nil && size > 0 {
			cfg.CacheSize = size
		}
	}

	if walMode := os.Getenv("DB_WAL_MODE"); walMode != "" {
		cfg.WALMode = walMode == "true" || walMode == "1"
	}

	if cacheEnabled := os.Getenv("CACHE_ENABLED"); cacheEnabled != "" {
		cfg.CacheEnabled = cacheEnabled == "true" || cacheEnabled == "1"
	}

	if cacheTTL := os.Getenv("CACHE_TTL"); cacheTTL != "" {
		if ttl, err := time.ParseDuration(cacheTTL); err == nil {
			cfg.CacheTTL = ttl
		}
	}

	if cacheSize := os.Getenv("CACHE_SIZE"); cacheSize != "" {
		if size, err := strconv.Atoi(cacheSize); err == nil && size > 0 {
			cfg.CacheMaxItems = size
		}
	}

	return cfg
}

// Validate validates the database configuration
func (c Config) Validate() error {
	if c.Path == "" {
		return errors.New("database path cannot be empty")
	}

	if c.PoolSize <= 0 {
		return errors.New("pool size must be positive")
	}

	if c.BusyTimeout <= 0 {
		return errors.New("busy timeout must be positive")
	}

	if c.QueryTimeout <= 0 {
		return errors.New("query timeout must be positive")
	}

	if c.CacheEnabled && c.CacheTTL <= 0 {
		return errors.New("cache TTL must be positive when cache is enabled")
	}

	if c.CacheEnabled && c.CacheMaxItems <= 0 {
		return errors.New("cache max items must be positive when cache is enabled")
	}

	return nil
}

// EnsureDirectoryExists ensures the directory for the database file exists
func (c Config) EnsureDirectoryExists() error {
	dir := filepath.Dir(c.Path)
	if dir == "." {
		return nil
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for database: %w", err)
	}

	return nil
}
