# SQLite Database Implementation

This webserver uses SQLite as its embedded database engine, providing a lightweight yet powerful storage solution with zero external dependencies.

## Features

- Embedded SQLite database with WAL mode
- Connection pooling and prepared statements
- In-memory caching with LRU eviction
- Automatic migrations
- Query builder with type safety
- Transaction support with retries
- Metrics collection and monitoring

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

## Usage Example

```go
package main

import (
    "context"
    "github.com/magooney-loon/webserver/internal/database"
)

func main() {
    // Initialize database
    db, err := database.New(database.Config{
        Path:       "./data/app.db",
        PoolSize:   10,
        CacheSize:  2000,
        WALEnabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Use with transaction
    err = db.Transaction(context.Background(), func(tx *database.Tx) error {
        // Perform operations
        return nil
    })
}
```

## Schema Management

Migrations are automatically applied on startup:

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

-- migrations/002_sessions.sql
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);
```

## Performance Optimizations

1. **Write-Ahead Logging (WAL)**
   - Enables concurrent reads during writes
   - Improves write performance
   - Reduces chance of database corruption

2. **Connection Pool**
   - Reuses database connections
   - Configurable pool size
   - Connection health checks

3. **Prepared Statements**
   - Statement caching
   - Query plan optimization
   - Protection against SQL injection

4. **In-Memory Cache**
   - LRU eviction policy
   - Configurable TTL
   - Automatic cache invalidation
   - Negative caching support

## Monitoring

Database metrics are exposed via the `/metrics` endpoint:

- Query latency histograms
- Connection pool stats
- Cache hit/miss ratios
- Transaction success/failure rates
- Storage utilization

## Security

1. **Query Safety**
   - All queries use prepared statements
   - Input validation and sanitization
   - Transaction isolation

2. **Access Control**
   - Row-level security
   - Query authorization
   - Audit logging

3. **Data Protection**
   - Automatic backup support
   - WAL journaling
   - Corruption detection

## Implementation Status

Current implementation status:

✅ Basic SQLite integration
✅ Connection pooling
✅ Transaction support
✅ Schema migrations
🚧 In-memory cache (WIP)
🚧 Metrics collection (WIP)
🚧 Query builder (WIP)
📋 Backup system (Planned)
📋 Admin dashboard (Planned)

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