package database

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	// ID is the unique identifier for the migration (typically a number)
	ID int

	// Name is the descriptive name of the migration
	Name string

	// SQL contains the SQL statements to execute
	SQL string

	// Applied indicates if the migration has been applied
	Applied bool

	// AppliedAt is when the migration was applied
	AppliedAt time.Time
}

// MigrationOptions configures migration behavior
type MigrationOptions struct {
	// MigrationsDir is the directory containing migration files
	MigrationsDir string

	// AllowDown enables downward migrations (dangerous)
	AllowDown bool

	// AutoApply automatically applies pending migrations
	AutoApply bool

	// TableName is the name of the migrations table
	TableName string
}

// DefaultMigrationOptions returns default migration options
func DefaultMigrationOptions() MigrationOptions {
	return MigrationOptions{
		MigrationsDir: "./migrations",
		AllowDown:     false,
		AutoApply:     true,
		TableName:     "schema_migrations",
	}
}

// MigrateSchema applies pending migrations to the database
func (db *Database) MigrateSchema(ctx context.Context) error {
	return db.MigrateSchemaWithOptions(ctx, DefaultMigrationOptions())
}

// MigrateSchemaWithOptions applies migrations with specified options
func (db *Database) MigrateSchemaWithOptions(ctx context.Context, opts MigrationOptions) error {
	// Ensure migration table exists
	if err := db.ensureMigrationTable(ctx, opts.TableName); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Get applied migrations
	applied, err := db.getAppliedMigrations(ctx, opts.TableName)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get available migrations
	available, err := db.loadMigrationFiles(opts.MigrationsDir)
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	// Sort migrations by ID
	sort.Slice(available, func(i, j int) bool {
		return available[i].ID < available[j].ID
	})

	// Find pending migrations
	pending := db.findPendingMigrations(available, applied)

	// Apply pending migrations if auto-apply is enabled
	if opts.AutoApply && len(pending) > 0 {
		err = db.applyMigrations(ctx, pending, opts.TableName)
		if err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	return nil
}

// ensureMigrationTable creates the migration table if it doesn't exist
func (db *Database) ensureMigrationTable(ctx context.Context, tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL
		)
	`, tableName)

	_, err := db.Exec(ctx, query)
	return err
}

// getAppliedMigrations returns all migrations that have been applied
func (db *Database) getAppliedMigrations(ctx context.Context, tableName string) (map[int]Migration, error) {
	query := fmt.Sprintf("SELECT id, name, applied_at FROM %s", tableName)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]Migration)
	for rows.Next() {
		var m Migration
		var appliedAt string
		if err := rows.Scan(&m.ID, &m.Name, &appliedAt); err != nil {
			return nil, err
		}

		// Parse applied_at
		t, err := time.Parse("2006-01-02 15:04:05", appliedAt)
		if err != nil {
			return nil, err
		}

		m.Applied = true
		m.AppliedAt = t
		applied[m.ID] = m
	}

	return applied, nil
}

// loadMigrationFiles loads migration files from the given directory
func (db *Database) loadMigrationFiles(dir string) ([]Migration, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", dir)
	}

	var migrations []Migration
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fileName := filepath.Base(path)
		parts := strings.SplitN(fileName, "_", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid migration file name: %s", fileName)
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid migration ID: %s", parts[0])
		}

		name := strings.TrimSuffix(parts[1], ".sql")
		migrations = append(migrations, Migration{
			ID:   id,
			Name: name,
			SQL:  string(content),
		})

		return nil
	})

	return migrations, err
}

// findPendingMigrations returns migrations that haven't been applied yet
func (db *Database) findPendingMigrations(available []Migration, applied map[int]Migration) []Migration {
	var pending []Migration
	for _, m := range available {
		if _, exists := applied[m.ID]; !exists {
			pending = append(pending, m)
		}
	}
	return pending
}

// applyMigrations applies the given migrations to the database
func (db *Database) applyMigrations(ctx context.Context, migrations []Migration, tableName string) error {
	for _, m := range migrations {
		// Start a transaction for each migration
		err := db.Transaction(ctx, func(tx *Tx) error {
			// Apply the migration
			if _, err := tx.Exec(ctx, m.SQL); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", m.ID, err)
			}

			// Record the migration
			query := fmt.Sprintf("INSERT INTO %s (id, name, applied_at) VALUES (?, ?, ?)", tableName)
			if _, err := tx.Exec(ctx, query, m.ID, m.Name, time.Now().Format("2006-01-02 15:04:05")); err != nil {
				return fmt.Errorf("failed to record migration %d: %w", m.ID, err)
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// GetMigrationStatus returns the status of all migrations
func (db *Database) GetMigrationStatus(ctx context.Context, dir string, tableName string) ([]Migration, error) {
	// Get applied migrations
	applied, err := db.getAppliedMigrations(ctx, tableName)
	if err != nil {
		return nil, err
	}

	// Get available migrations
	available, err := db.loadMigrationFiles(dir)
	if err != nil {
		return nil, err
	}

	// Combine information
	var status []Migration
	for _, m := range available {
		if a, exists := applied[m.ID]; exists {
			m.Applied = true
			m.AppliedAt = a.AppliedAt
		}
		status = append(status, m)
	}

	// Sort by ID
	sort.Slice(status, func(i, j int) bool {
		return status[i].ID < status[j].ID
	})

	return status, nil
}
