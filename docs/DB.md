# SQLite Database Implementation

This webserver uses SQLite as its embedded database engine, providing a lightweight yet powerful storage solution with zero external dependencies.

## Core Database Types

```go
// Database represents the SQLite database instance
type Database struct {
    db *sql.DB        // Underlying sql.DB instance
    config *Config    // Database configuration
    mu sync.RWMutex   // Mutex for thread safety
    pool *Pool        // Connection pool
    metrics *Metrics  // Performance metrics
    stmtCache map[string]*sql.Stmt // Prepared statement cache
}

// Config holds the database configuration
type Config struct {
    Path string             // File path to SQLite database
    PoolSize int            // Maximum connections in pool
    BusyTimeout time.Duration    // Timeout for acquiring locks
    QueryTimeout time.Duration   // Timeout for query execution
    ConnMaxLifetime time.Duration // Connection reuse limit
    ConnMaxIdleTime time.Duration // Connection idle limit
    CacheSize int           // Pages to cache in memory
    WALMode bool            // Write-Ahead Logging mode
    AutoMigrate bool        // Apply migrations on startup
    CacheEnabled bool       // In-memory query caching
    CacheTTL time.Duration  // Cache time-to-live
    CacheMaxItems int       // Maximum cache entries
}
```

## Features

- ✅ Embedded SQLite database with WAL mode
- ✅ Connection pooling and prepared statements
- ✅ Singleton pattern implementation
- ✅ Transaction support with retries
- ✅ Comprehensive configuration system
- ✅ Structured logging integration
- 🚧 In-memory caching with LRU eviction
- 🚧 Query builder with type safety
- 🚧 Metrics collection and monitoring

## Singleton Pattern

The database uses a singleton pattern for global access while ensuring only one instance exists:

```go
var (
    // instance is the singleton database instance
    instance *Database
    
    // once ensures the database is instantiated only once
    once sync.Once
    
    // ErrNotInitialized is returned when the database is not initialized
    ErrNotInitialized = errors.New("database not initialized")
)

// Initialize initializes the singleton database instance
func Initialize(config Config) error {
    var err error
    once.Do(func() {
        instance, err = New(config)
    })
    return err
}

// GetInstance returns the singleton database instance
func GetInstance() (*Database, error) {
    if instance == nil {
        return nil, ErrNotInitialized
    }
    return instance, nil
}
```

## Configuration

Configure SQLite through environment variables or programmatically:

```env
# Database Settings
DB_PATH=./data/app.db           # SQLite database file location
DB_POOL_SIZE=10                 # Connection pool size
DB_TIMEOUT=30s                  # Query timeout
DB_CACHE_SIZE=2000             # Page cache size in pages
DB_WAL_MODE=true               # Enable Write-Ahead Logging
DB_BUSY_TIMEOUT=5000           # Busy timeout in milliseconds

# Cache Settings
CACHE_ENABLED=true             # Enable in-memory cache
CACHE_TTL=5m                   # Cache TTL for entries
CACHE_SIZE=1000               # Maximum cache entries
```

The configuration system supports both environment variables and programmatic configuration:

```go
// Load configuration from environment
dbConfig := database.LoadFromEnv()

// Or set configuration programmatically
dbConfig := database.DefaultConfig()
dbConfig.Path = "./data/app.db"
dbConfig.PoolSize = 5
dbConfig.WALMode = true
dbConfig.AutoMigrate = false    // Disable auto-migrations
```

## Usage Example

Here's how to use the database in your application:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/magooney-loon/webserver/internal/database"
    "github.com/magooney-loon/webserver/pkg/logger"
)

func main() {
    // Initialize structured logging
    log := logger.New(logger.WithLevel("INFO"))

    // Initialize database
    dbConfig := database.DefaultConfig()
    dbConfig.Path = "./data/app.db"
    dbConfig.PoolSize = 5
    dbConfig.WALMode = true
    dbConfig.AutoMigrate = false
    
    log.Info("initializing database", map[string]interface{}{
        "path":      dbConfig.Path,
        "pool_size": dbConfig.PoolSize,
        "wal_mode":  dbConfig.WALMode,
    })
    
    if err := database.Initialize(dbConfig); err != nil {
        log.Fatal("database initialization failed", map[string]interface{}{
            "error": err.Error(),
        })
    }
    
    // Get database instance
    db, err := database.GetInstance()
    if err != nil {
        log.Fatal("failed to get database instance", map[string]interface{}{
            "error": err.Error(),
        })
    }
    defer db.Close()
    
    // Use with transaction
    ctx := context.Background()
    err = db.Transaction(ctx, func(tx *database.Tx) error {
        // Insert user
        _, err := tx.Exec(ctx, 
            "INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
            "newuser", "newuser@example.com", "hashed_password")
        return err
    })
    
    if err != nil {
        log.Error("transaction failed", map[string]interface{}{
            "error": err.Error(),
        })
    } else {
        log.Info("transaction succeeded", nil)
    }
}
```

## Query API

The database provides a simple, consistent API for executing queries:

```go
// Basic query execution
func (db *Database) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

// Query multiple rows
func (db *Database) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

// Query a single row
func (db *Database) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

// Execute operation in a transaction
func (db *Database) Transaction(ctx context.Context, fn func(*Tx) error) error
```

## Transaction Support

Transactions provide atomic operations with automatic rollback on errors:

```go
// Create a new order with items atomically
func CreateOrder(ctx context.Context, order Order) (int64, error) {
    db, err := database.GetInstance()
    if err != nil {
        return 0, err
    }
    
    var orderID int64
    
    err = db.Transaction(ctx, func(tx *database.Tx) error {
        // Insert order
        result, err := tx.Exec(ctx, 
            "INSERT INTO orders (user_id, total_amount) VALUES (?, ?)",
            order.UserID, order.TotalAmount)
        if err != nil {
            return err
        }
        
        // Get the order ID
        orderID, err = result.LastInsertId()
        if err != nil {
            return err
        }
        
        // Insert order items
        for _, item := range order.Items {
            _, err = tx.Exec(ctx, 
                "INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)",
                orderID, item.ProductID, item.Quantity)
            if err != nil {
                return err
            }
        }
        
        return nil
    })
    
    return orderID, err
}
```

## Schema Management

The database supports automatic schema migrations by default (can be disabled):

```sql
-- migrations/001_initial_schema.sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
```

Manual table creation can be done programmatically:

```go
// Create tables
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
```

## Performance Optimizations

The database implementation includes several performance optimizations:

1. **Write-Ahead Logging (WAL)**
   - Enables concurrent reads during writes
   - Improves write performance
   - Reduces chance of database corruption

2. **Connection Pool**
   - Reuses database connections
   - Configurable pool size
   - Connection health checks

3. **Prepared Statement Caching**
   - Statement caching reduces parsing overhead
   - Improves query execution time
   - Protects against SQL injection

4. **Concurrency Control**
   - Fine-grained locking mechanisms
   - Prevents connection contention
   - Thread-safe operations

## Integration with Logging

The database integrates with the structured logging system:

```go
// In your handler functions
func userHandler(w http.ResponseWriter, r *http.Request) {
    // Get logger
    log := r.Context().Value("logger").(*logger.Logger)
    
    // Get database
    db, err := database.GetInstance()
    if err != nil {
        log.Error("failed to get database instance", map[string]interface{}{
            "error": err.Error(),
        })
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    
    // Create context with timeout
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    // Query the database with logging
    var user User
    userID := r.URL.Query().Get("id")
    
    log.Info("looking up user", map[string]interface{}{
        "user_id": userID,
    })
    
    row := db.QueryRow(ctx, "SELECT id, username, email FROM users WHERE id = ?", userID)
    if err := row.Scan(&user.ID, &user.Username, &user.Email); err != nil {
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
    })
    
    // Return data
    json.NewEncoder(w).Encode(user)
}
```

## Implementation Roadmap

1. **Phase 1: Core Database ✅**
   - ✅ Singleton pattern implementation
   - ✅ Configuration from environment
   - ✅ Basic connection management
   - ✅ Error handling and recovery

2. **Phase 2: Transaction & Migration System ✅**
   - ✅ Transaction support with retries
   - ✅ Schema migration framework
   - ✅ Initial schema definition
   - ✅ Integration tests

3. **Phase 3: Performance Optimizations 🚧**
   - ✅ Connection pooling
   - ✅ Statement preparation and caching
   - ✅ WAL mode configuration
   - ✅ Query timeout handling
   - 🚧 In-memory caching with LRU

4. **Phase 4: Advanced Features 📋**
   - 📋 Query builder implementation
   - 🚧 Metrics collection
   - 📋 Backup system foundation
   - 📋 Admin dashboard integration

## Future Enhancements

1. **Query Builder**
   - Type-safe query construction
   - Automatic parameter binding
   - Result scanning
   
2. **Backup System**
   - Scheduled backups
   - Point-in-time recovery
   - Backup encryption

3. **Admin Dashboard**
   - Schema visualization
   - Query execution
   - Performance monitoring
   - Backup management