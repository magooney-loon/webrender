# Server Configuration Guide

The server provides multiple ways to configure its behavior, giving you flexibility in how you deploy and run it.

## Configuration Methods

There are three main ways to configure the server, in order of precedence:

1. **Programmatic configuration** - Direct in code using configuration options
2. **Environment variables** - Set using the system environment or `.env` file
3. **Default values** - Sensible defaults that apply if not otherwise specified

## Programmatic Configuration

The most flexible approach is to use the configuration API directly in your code:

```go
// Load configuration with options
cfg, log := config.LoadWithOptions(
    // Only specify what you need to override
    config.WithEnvironment("production"),
    config.WithServerPort(8443),
    config.WithAuthEnabled(true),
    config.WithTLS(true, "/path/to/cert.pem", "/path/to/key.pem"),
)

// Use the configuration
srv := server.New(cfg, log, /* server options */)
```

### Available Configuration Options

| Option Function | Description | Example |
|-----------------|-------------|---------|
| `WithEnvironment(env)` | Set environment (development, production, test) | `WithEnvironment("production")` |
| `WithServerPort(port)` | Set server port | `WithServerPort(8080)` |
| `WithServerHost(host)` | Set server host | `WithServerHost("0.0.0.0")` |
| `WithAuthEnabled(enabled)` | Enable/disable authentication | `WithAuthEnabled(true)` |
| `WithAuthCredentials(user, pass)` | Set auth credentials | `WithAuthCredentials("admin", "secret")` |
| `WithSystemAPI(enabled, prefix)` | Configure system API endpoints | `WithSystemAPI(true, "/system")` |
| `WithLogging(level, useJSON)` | Configure logging | `WithLogging("info", true)` |
| `WithTLS(enabled, cert, key)` | Configure TLS | `WithTLS(true, "cert.pem", "key.pem")` |

## Environment Variables

You can configure the server using environment variables, either set in your environment or in a `.env` file:

```env
# Server Environment
GO_ENV=production

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_SHUTDOWN_TIMEOUT=30s
SERVER_MAX_HEADER_SIZE=1MB
SERVER_MAX_BODY_SIZE=10MB

# System API Configuration
SYSTEM_API_ENABLED=true
SYSTEM_API_PREFIX=/system

# Authentication
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=strongpassword
AUTH_EXCLUDE_PATHS=/api/v1/health,/public
```

Environment variables will override defaults but are overridden by explicit programmatic configuration.

## Default Configuration

The server ships with sensible defaults that work for many development scenarios:

```go
// These defaults apply if not overridden
Config{
    Environment: "development",
    Server: {
        Port:            8080,
        Host:            "localhost",
        ReadTimeout:     30s,
        WriteTimeout:    30s,
        ShutdownTimeout: 30s,
        MaxHeaderSize:   1MB,
        MaxBodySize:     10MB,
    },
    System: {
        Enabled:        true,
        Prefix:         "/system",
        MetricsEnabled: true,
        HealthEnabled:  true,
    },
    Logging: {
        Level:      "info",
        UseJSON:    false,
        EnableFile: false,
    },
    // ... and more defaults
}
```

## Priority Resolution

Configuration is resolved in this order:

1. Explicit values passed to `LoadWithOptions`
2. Environment variables (from system or `.env` file)
3. Default values

This means you can override just what you need while accepting defaults for everything else.

## Example: Mixed Configuration

```go
// In your application code
cfg, log := config.LoadWithOptions(
    config.WithServerPort(9000), // Override the port 
)

// In .env
GO_ENV=production
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=securepassword

// The result is:
// - Server port: 9000 (from code)
// - Environment: production (from .env)
// - Auth: enabled with admin/securepassword (from .env)
// - All other settings: default values
```

## Complete Configuration Reference

For a complete list of all configuration options and their defaults, please see the [type definitions](https://github.com/magooney-loon/webserver/blob/main/internal/config/config.go) in the codebase. 