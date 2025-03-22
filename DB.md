# BoltDB Implementation Plan for WebRender Admin Dashboard

## Overview
This document outlines the plan for integrating BoltDB into the WebRender admin dashboard. BoltDB is a simple, fast, and reliable key-value store written in Go that's perfect for projects that need a lightweight database without the complexity of a full RDBMS.

## Why BoltDB?
- **Embedded database**: No separate server installation required
- **Pure Go implementation**: Easy integration with our existing codebase
- **ACID compliant**: Ensures data integrity with transaction support
- **Zero configuration**: Simple to set up and deploy
- **Performant**: Fast reads with B+tree indexing
- **File-based**: Easy to backup and restore

## Implementation Checklist

### Phase 1: Initial Setup
- [ ] Add BoltDB dependency to go.mod (`go get go.etcd.io/bbolt`)
- [ ] Create new package `internal/admin/database` for DB operations
- [ ] Implement database initialization function
- [ ] Define bucket structure for different data types
- [ ] Add database config options to main configuration

### Phase 2: User Management
- [ ] Create user model with proper struct tags
- [ ] Implement CRUD operations for users
  - [ ] Create/register new user
  - [ ] Read/get user by username
  - [ ] Update user details
  - [ ] Delete user
- [ ] Implement proper password hashing using bcrypt
- [ ] Update login handler to use database authentication
- [ ] Add user management UI to admin dashboard

### Phase 3: Session Management
- [ ] Store sessions in BoltDB instead of memory
- [ ] Implement session cleanup for expired sessions
- [ ] Create session bucket with TTL support
- [ ] Implement proper session serialization/deserialization

### Phase 4: System Settings & Configuration
- [ ] Create settings bucket for application configuration
- [ ] Implement settings CRUD operations
- [ ] Create settings UI in admin dashboard
- [ ] Add defaults mechanism for first-time setup

### Phase 5: Analytics & Metrics
- [ ] Design schema for storing system metrics and events
- [ ] Implement time-series data storage for dashboard metrics
- [ ] Create analytics queries for dashboard charts
- [ ] Set up automatic pruning for old metrics data
- [ ] Update dashboard UI to display real metrics from DB

### Phase 6: Integration with Component System
- [ ] Create DB-backed component state persistence
- [ ] Implement component state recovery on restart
- [ ] Add versioning system for component data

### Phase 7: Admin Features
- [ ] Implement user activity logging
- [ ] Create audit trail functionality
- [ ] Add database backup and restore features
- [ ] Implement database health monitoring (replace mock status)

### Phase 8: Database Management UI
- [ ] Create database explorer component for admin dashboard
  - [ ] Bucket browser with expandable tree view
  - [ ] Key-value viewer with pagination
  - [ ] JSON/data visualization for complex values
- [ ] Implement data editing capabilities
  - [ ] Add/edit/delete records interface
  - [ ] Validation for different data types
  - [ ] Confirmation for destructive operations
- [ ] Add search functionality
  - [ ] Full-text search across keys
  - [ ] Advanced query builder for complex searches
  - [ ] Saved searches feature
- [ ] Create data export/import tools
  - [ ] Export to JSON/CSV formats
  - [ ] Bulk import with validation
  - [ ] Transaction history for imports
- [ ] Implement access controls for database operations
  - [ ] Role-based permissions for viewing/editing
  - [ ] Audit logging for all database operations
- [ ] Add database performance monitoring
  - [ ] Query timing metrics
  - [ ] Size monitoring for buckets
  - [ ] Index optimization suggestions

## Database Schema Design

### Buckets
- `users` - User accounts and authentication
- `sessions` - User sessions data
- `settings` - Application configuration
- `metrics` - System performance metrics
- `audit` - Audit trail of admin actions
- `components` - Persisted component states

### Key Structures
- Users: `users:{username}` → User JSON
- Sessions: `sessions:{sessionID}` → Session JSON
- Settings: `settings:{key}` → Setting Value
- Metrics: `metrics:{timestamp}:{metric}` → Value
- Audit: `audit:{timestamp}:{username}` → Action JSON
- Components: `components:{componentID}` → State JSON

## Implementation Details

### Database Connection Management
```go
type DB struct {
    bolt *bbolt.DB
    path string
    mu   sync.RWMutex
}

func NewDB(path string) (*DB, error) {
    options := &bbolt.Options{
        Timeout: 1 * time.Second,
    }
    
    db, err := bbolt.Open(path, 0600, options)
    if err != nil {
        return nil, err
    }
    
    // Create required buckets
    err = db.Update(func(tx *bbolt.Tx) error {
        buckets := []string{"users", "sessions", "settings", "metrics", "audit", "components"}
        for _, bucket := range buckets {
            _, err := tx.CreateBucketIfNotExists([]byte(bucket))
            if err != nil {
                return err
            }
        }
        return nil
    })
    
    return &DB{
        bolt: db,
        path: path,
    }, err
}
```

### Database Explorer UI Component

The database explorer UI will be implemented as a WebRender component:

```go
// Example structure for the database explorer component
func NewDatabaseExplorer(id string) *component.Component {
    explorer := component.New(id)
    
    // Initial state
    explorer.State.Set("currentBucket", "")
    explorer.State.Set("currentPrefix", "")
    explorer.State.Set("page", 1)
    explorer.State.Set("pageSize", 50)
    explorer.State.Set("searchQuery", "")
    explorer.State.Set("selectedKey", "")
    explorer.State.Set("buckets", []string{})
    explorer.State.Set("keys", []string{})
    explorer.State.Set("totalKeys", 0)
    explorer.State.Set("editMode", false)
    
    // Methods for data interaction
    explorer.Methods["selectBucket"] = func(c *component.Component, args ...interface{}) error {
        bucket := args[0].(string)
        c.State.Set("currentBucket", bucket)
        c.State.Set("currentPrefix", "")
        c.State.Set("page", 1)
        
        // Load keys for this bucket
        return loadBucketKeys(c, bucket, "", 1, c.State.Get("pageSize").(int))
    }
    
    explorer.Methods["viewKey"] = func(c *component.Component, args ...interface{}) error {
        key := args[0].(string)
        c.State.Set("selectedKey", key)
        
        // Load key value
        return loadKeyValue(c, c.State.Get("currentBucket").(string), key)
    }
    
    // Additional methods for pagination, search, edit, save, etc.
    
    // Set component template
    explorer.SetTemplate(databaseExplorerTemplate)
    
    return explorer
}

// Functions to interact with the database
func loadBucketKeys(c *component.Component, bucket string, prefix string, page int, pageSize int) error {
    // Implement loading keys from bucket with pagination
    // ...
}

func loadKeyValue(c *component.Component, bucket string, key string) error {
    // Implement loading a specific key's value
    // ...
}
```

### Future Considerations
- Potential sharding for metrics data if volume grows
- Possible integration with full-text search (Bleve)
- Backup rotation system
- Data export/import functionality
- Migration system for schema changes

## Testing Strategy
- Unit tests for all database operations
- Integration tests for authentication flow
- Benchmarks for performance-critical operations
- Test fixtures for standard database states 