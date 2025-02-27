package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Tx represents a database transaction
type Tx struct {
	// tx is the underlying sql.Tx instance
	tx *sql.Tx

	// db is the parent database
	db *Database

	// startTime is when the transaction was started
	startTime time.Time

	// readonly indicates if the transaction is read-only
	readonly bool
}

var (
	// ErrTxDone is returned when a transaction is used after commit or rollback
	ErrTxDone = errors.New("transaction already committed or rolled back")

	// ErrTxTimeout is returned when a transaction times out
	ErrTxTimeout = errors.New("transaction timed out")

	// ErrTxConflict is returned when a transaction fails due to a conflict
	ErrTxConflict = errors.New("transaction conflict, retry needed")
)

// TransactionOptions configures a database transaction
type TransactionOptions struct {
	// Readonly indicates if the transaction is read-only
	Readonly bool

	// Retries is the number of times to retry on conflict
	Retries int

	// RetryDelay is the delay between retries
	RetryDelay time.Duration

	// Isolation sets the transaction isolation level
	Isolation sql.IsolationLevel
}

// DefaultTransactionOptions returns default transaction options
func DefaultTransactionOptions() TransactionOptions {
	return TransactionOptions{
		Readonly:   false,
		Retries:    3,
		RetryDelay: 50 * time.Millisecond,
		Isolation:  sql.LevelSerializable,
	}
}

// ReadonlyTransactionOptions returns options for read-only transactions
func ReadonlyTransactionOptions() TransactionOptions {
	return TransactionOptions{
		Readonly:   true,
		Retries:    1,
		RetryDelay: 10 * time.Millisecond,
		Isolation:  sql.LevelReadCommitted,
	}
}

// TxFunc is a function that runs within a transaction
type TxFunc func(*Tx) error

// Transaction executes a function within a transaction
func (db *Database) Transaction(ctx context.Context, fn TxFunc) error {
	return db.TransactionWithOptions(ctx, fn, DefaultTransactionOptions())
}

// TransactionReadOnly executes a function within a read-only transaction
func (db *Database) TransactionReadOnly(ctx context.Context, fn TxFunc) error {
	return db.TransactionWithOptions(ctx, fn, ReadonlyTransactionOptions())
}

// TransactionWithOptions executes a function within a transaction with specified options
func (db *Database) TransactionWithOptions(ctx context.Context, fn TxFunc, opts TransactionOptions) error {
	var err error
	var tx *sql.Tx
	var retries int

	db.metrics.RecordTransactionStart()
	startTime := time.Now()

	// Try the transaction with retries
	for retries = 0; retries <= opts.Retries; retries++ {
		// Skip delay on first attempt
		if retries > 0 {
			select {
			case <-time.After(opts.RetryDelay):
				// Continue after delay
			case <-ctx.Done():
				// Context cancelled or timed out
				return ctx.Err()
			}
		}

		// Begin transaction
		tx, err = db.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: opts.Isolation,
			ReadOnly:  opts.Readonly,
		})
		if err != nil {
			continue // Try again if begin failed
		}

		// Create transaction wrapper
		txWrapper := &Tx{
			tx:        tx,
			db:        db,
			startTime: startTime,
			readonly:  opts.Readonly,
		}

		// Execute the transaction function
		fnErr := fn(txWrapper)

		// Handle the result
		if fnErr != nil {
			// Rollback on error from the function
			_ = tx.Rollback()

			// Check if we should retry on conflict
			if isRetriableError(fnErr) && retries < opts.Retries {
				continue // Try again
			}

			// Track metrics
			duration := time.Since(startTime)
			db.metrics.RecordTransactionEnd(duration, false, fnErr)

			return fnErr
		}

		// Attempt to commit
		if err = tx.Commit(); err != nil {
			// Check if we should retry on conflict
			if isRetriableError(err) && retries < opts.Retries {
				continue // Try again
			}

			// Track metrics
			duration := time.Since(startTime)
			db.metrics.RecordTransactionEnd(duration, false, err)

			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Track metrics for successful transaction
		duration := time.Since(startTime)
		db.metrics.RecordTransactionEnd(duration, true, nil)

		return nil // Success
	}

	// If we get here, we've exhausted all retries
	duration := time.Since(startTime)
	db.metrics.RecordTransactionEnd(duration, false, err)

	return fmt.Errorf("transaction failed after %d retries: %w", retries, err)
}

// isRetriableError determines if an error can be retried
func isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for known retriable error patterns
	if errors.Is(err, ErrTxConflict) {
		return true
	}

	// Check for SQLite busy or locked errors
	errMsg := err.Error()
	return strings.Contains(errMsg, "database is locked") ||
		strings.Contains(errMsg, "database is busy") ||
		strings.Contains(errMsg, "constraint failed") ||
		strings.Contains(errMsg, "conflict") ||
		strings.Contains(errMsg, "try again")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Exec executes a query within the transaction
func (tx *Tx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return tx.tx.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows within the transaction
func (tx *Tx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return tx.tx.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row within the transaction
func (tx *Tx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return tx.tx.QueryRowContext(ctx, query, args...)
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}
