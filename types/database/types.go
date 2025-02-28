package types

import (
	"context"
	"database/sql"
)

// Store represents a database store
type Store interface {
	// Initialize initializes the store
	Initialize() error
	// Close closes the store
	Close() error
	// DB returns the underlying sql.DB
	DB() *sql.DB
	// WithContext returns a new Store with the given context
	WithContext(ctx context.Context) Store
}

// Config represents database configuration
type Config struct {
	// Driver is the database driver name
	Driver string
	// DataSource is the database connection string
	DataSource string
	// MaxOpenConns is the maximum number of open connections
	MaxOpenConns int
	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int
	// ConnMaxLifetime is the maximum amount of time a connection may be reused
	ConnMaxLifetime int
	// Options contains additional driver-specific options
	Options map[string]interface{}
}

// QueryOptions represents options for database queries
type QueryOptions struct {
	// Timeout is the query timeout in seconds
	Timeout int
	// ReadOnly indicates if the query is read-only
	ReadOnly bool
	// UsePrepared indicates if the query should use prepared statements
	UsePrepared bool
	// UseCache indicates if the query should use caching
	UseCache bool
	// CacheTTL is the cache TTL in seconds
	CacheTTL int
}

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit() error
	// Rollback rolls back the transaction
	Rollback() error
	// WithContext returns a new Transaction with the given context
	WithContext(ctx context.Context) Transaction
}

// Result represents a database operation result
type Result interface {
	// LastInsertId returns the ID of the last inserted row
	LastInsertId() (int64, error)
	// RowsAffected returns the number of rows affected
	RowsAffected() (int64, error)
}

// Scanner represents a row scanner interface
type Scanner interface {
	// Scan copies the columns in the current row into the values pointed at by dest
	Scan(dest ...interface{}) error
}
