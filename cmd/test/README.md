# Configuration System Example

This example demonstrates how to use the configuration management system. It shows:

1. Basic configuration operations (string, int)
2. JSON configuration handling
3. Real-time configuration watching
4. Server configuration management

## Running the Example

```bash
go run config_example.go
```

## Expected Output

```
=== Basic Configuration Operations ===
App Name: Test App
Max Users: 1000

=== JSON Configuration ===
Enabled Features: [auth chat notifications]

=== Configuration Watching ===
Waiting for config update...
Received config update: {Key:app.environment Value:production Type:string UpdatedAt:2024-02-28 15:04:05.123456789 +0000 UTC}

=== Server Configuration ===
Current server port: 8080
Updated server port: 9090

Configuration example completed successfully!
```

## Features Demonstrated

1. **Basic Operations**
   - Setting and getting string values
   - Setting and getting integer values
   - Error handling

2. **JSON Configuration**
   - Storing complex data structures
   - JSON marshaling/unmarshaling
   - Type safety

3. **Configuration Watching**
   - Real-time configuration updates
   - Asynchronous notification
   - Context handling

4. **Server Configuration**
   - Server port management
   - Dynamic configuration updates
   - Default values

## Database

The example uses SQLite with an in-memory database. The configuration is persisted
and can be accessed across multiple sessions if you use a file-based database
(configured via `database.Config.Path`). 