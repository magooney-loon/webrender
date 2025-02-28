# Middleware Architecture

## Current Flow
```mermaid
graph LR
    Request --> Logger
    Logger --> Session
    Session --> Auth
    Auth --> CORS
    CORS --> Security
    Security --> Recovery
    Recovery --> Monitor
    Monitor --> Handler
```

## Core Components

### Chain Implementation
```go
type Chain interface {
    Add(middleware ...Middleware) Chain
    Then(http.Handler) http.Handler
}

type Middleware interface {
    Wrap(http.Handler) http.Handler
}
```

## Middleware Stack

### 1. Logger Middleware
```go
type LoggerOptions struct {
    Format     string
    TimeFormat string
    UTC        bool
}
```
- Request/response logging
- Timing information
- UTC vs local time support
- Custom format strings

### 2. Session Middleware
- Handles user sessions
- Sets/reads session cookies
- Manages session state

### 3. Auth Middleware
```go
type AuthOptions struct {
    Realm          string
    ExcludePaths   []string
    TokenValidator func(string) bool
}
```
- JWT validation
- Role-based access control
- Auth header parsing
- Path exclusions

### 4. CORS Middleware
```go
type CORSOptions struct {
    AllowedOrigins   []string
    AllowedMethods   []string
    AllowedHeaders   []string
    ExposedHeaders   []string
    AllowCredentials bool
    MaxAge           int
}
```
- Cross-origin resource sharing
- Preflight requests
- Origin validation
- Credentials handling

### 5. Security Headers
```go
type SecurityOptions struct {
    HSTS            bool
    HSTSMaxAge      int
    FrameOptions    string
    ContentSecurity string
    ReferrerPolicy  string
}
```
- HSTS configuration
- XSS Protection
- Frame Options
- Content Security Policy
- Referrer Policy

### 6. Recovery Middleware
- Panic recovery
- Error logging
- Graceful error responses

### 7. Monitoring
- Request duration
- Status codes
- Error rates
- Resource usage

## Recommended Improvements

### 1. Rate Limiting
```go
type RateLimiter struct {
    Store  RedisClient
    Limit  rate.Limit
    Burst  int
}
```
- Token bucket algorithm
- Redis-backed storage
- IP-based limiting
- User-based limiting

### 2. Request ID Tracking
```go
type RequestIDOptions struct {
    Header     string // X-Request-ID
    Generator  func() string
    Validator  func(string) bool
}
```
- UUID generation
- Header propagation
- Trace correlation

### 3. Timeout Control
```go
type TimeoutOptions struct {
    Duration time.Duration
    OnTimeout func(w http.ResponseWriter)
}
```
- Request timeouts
- Context cancellation
- Custom timeout handlers

### 4. Caching
```go
type CacheOptions struct {
    TTL           time.Duration
    KeyGenerator  func(*http.Request) string
    Store         CacheStore
}
```
- Response caching
- Cache invalidation
- Vary header support

### 5. Error Handling
```go
type ErrorHandlerOptions struct {
    Logger        Logger
    ErrorEncoder  func(error) []byte
    StatusCodes   map[error]int
}
```
- Structured error responses
- Error categorization
- Custom error pages

### 6. Compression
```go
type CompressionOptions struct {
    Level            int
    Types            []string
    MinLength        int
    ExcludedPaths    []string
}
```
- Gzip/Brotli compression
- Content type filtering
- Size thresholds

### 7. Metrics
```go
type MetricsOptions struct {
    Namespace   string
    Subsystem   string
    Labels      []string
    Registry    *prometheus.Registry
}
```
- Prometheus integration
- Custom metrics
- Label configuration

### 8. Context Propagation
```go
const (
    UserIDKey   ContextKey = "user_id"
    SessionKey  ContextKey = "session"
    TraceIDKey  ContextKey = "trace_id"
    RequestKey  ContextKey = "request"
)
```
- Request scoped values
- Type-safe context keys
- Value validation

## Best Practices

### Middleware Order
1. Recovery (first to catch panics)
2. Request ID
3. Logging
4. Compression
5. Security Headers
6. CORS
7. Rate Limiting
8. Authentication
9. Authorization
10. Business Logic

### Performance Tips
- Use sync.Pool for buffers
- Minimize allocations
- Cache expensive computations
- Use efficient data structures

### Security Guidelines
- Validate all inputs
- Sanitize outputs
- Use secure defaults
- Follow least privilege
- Implement rate limiting
- Set security headers

### Testing
```go
func TestMiddleware(t *testing.T) {
    handler := middleware.NewChain(
        middleware.NewRateLimit(),
        middleware.NewAuth(),
    ).Then(finalHandler)

    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    
    handler.ServeHTTP(rec, req)
    // Assert response
}
```

## Configuration Example
```yaml
middleware:
  ratelimit:
    enabled: true
    limit: 100
    burst: 50
  security:
    hsts: true
    hsts_max_age: 31536000
    frame_options: "DENY"
  cors:
    allowed_origins: ["*"]
    max_age: 3600
  cache:
    ttl: 300s
    size: 1000
  timeout:
    read: 5s
    write: 10s
    idle: 120s
``` 