package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/magooney-loon/webserver/pkg/logger"
)

// Global singleton instance
var (
	instance     *Config
	instanceOnce sync.Once
	instanceLock sync.RWMutex
)

type Config struct {
	Environment string
	Server      ServerConfig
	System      SystemConfig
	API         APIConfig
	Logging     LoggingConfig
	Security    SecurityConfig
	Admin       AdminConfig
}

type ServerConfig struct {
	Port            int
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderSize   int64
	MaxBodySize     int64
}

type SystemConfig struct {
	Enabled        bool
	Prefix         string
	MetricsEnabled bool
	HealthEnabled  bool
	AdminTemplate  string
}

type APIConfig struct {
	Prefix  string
	Version string
}

type LoggingConfig struct {
	Level        string
	UseJSON      bool
	EnableFile   bool
	FilePath     string
	RotationSize int64
	MaxAge       time.Duration
	Compression  bool
}

type SecurityConfig struct {
	EnableTLS      bool
	TLSCertPath    string
	TLSKeyPath     string
	AllowedOrigins []string
	RateLimit      RateLimitConfig
	CORS           CORSConfig
	Headers        SecurityHeadersConfig
	Auth           AuthConfig
}

type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type SecurityHeadersConfig struct {
	XSSProtection           string
	ContentTypeOptions      string
	XFrameOptions           string
	ContentSecurityPolicy   string
	ReferrerPolicy          string
	StrictTransportSecurity string
	PermissionsPolicy       string
}

type RateLimitConfig struct {
	Enabled  bool
	Requests int
	Window   time.Duration
	ByIP     bool
	ByRoute  bool
	Routes   map[string]RateLimitRule
}

type RateLimitRule struct {
	Requests int
	Window   time.Duration
	Message  string
}

type AdminConfig struct {
	Enabled         bool
	Path            string
	RefreshInterval time.Duration
	Username        string
	Password        string
}

type AuthConfig struct {
	Enabled      bool
	Username     string
	Password     string
	ExcludePaths []string
}

// GetInstance returns the singleton instance of the configuration
// If the configuration hasn't been loaded yet, it loads it
func GetInstance() (*Config, error) {
	instanceOnce.Do(func() {
		cfg, err := Load()
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}
		instance = cfg
	})
	return instance, nil
}

// GetInstanceWithLogger returns the singleton instance and a configured logger
func GetInstanceWithLogger() (*Config, *logger.Logger) {
	// Initialize startup logger
	startupLog := logger.New(logger.WithJSON(true), logger.WithLevel("info"))
	startupLog.Info("starting application", nil)

	// Get or initialize the singleton instance
	cfg, err := GetInstance()
	if err != nil {
		startupLog.Fatal("failed to load configuration", map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Printf("FATAL ERROR: Configuration failed to load: %s\n", err)
		fmt.Println("Make sure ALL required environment variables are set in your .env file")
		os.Exit(1)
	}

	// Initialize application logger based on loaded config
	log := logger.New(
		logger.WithJSON(cfg.Logging.UseJSON),
		logger.WithLevel(cfg.Logging.Level),
	)

	log.Info("configuration loaded successfully", map[string]interface{}{
		"environment": cfg.Environment,
	})

	return cfg, log
}

// Reload forces a reload of the configuration from environment variables
func Reload() (*Config, error) {
	instanceLock.Lock()
	defer instanceLock.Unlock()

	// Try to load .env file
	if err := LoadEnvFile(".env"); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
		fmt.Println("Continuing with environment variables from system...")
	}

	// Load new configuration with strict requirements
	newCfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to reload configuration: %w", err)
	}

	// Update the singleton instance
	instance = newCfg
	return instance, nil
}

// LoadWithLogging loads configuration and handles logging internally
// Returns config and a configured logger or exits on failure
// DEPRECATED: Use GetInstanceWithLogger instead
func LoadWithLogging() (*Config, *logger.Logger) {
	return GetInstanceWithLogger()
}

// Load loads configuration from environment variables
// All configuration values are required with no defaults
func Load() (*Config, error) {
	// Try to load .env file (only log warning if it fails, don't exit)
	if err := LoadEnvFile(".env"); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
		fmt.Println("Continuing with environment variables from system...")
	}

	cfg := &Config{}
	var err error

	// Environment
	cfg.Environment = os.Getenv("GO_ENV")
	if cfg.Environment == "" {
		return nil, fmt.Errorf("GO_ENV is not set")
	}

	// Server config
	serverCfg := ServerConfig{}

	serverPort, err := RequireEnv("SERVER_PORT")
	if err != nil {
		return nil, err
	}
	if serverCfg.Port, err = strconv.Atoi(serverPort); err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	serverCfg.Host, err = RequireEnv("SERVER_HOST")
	if err != nil {
		return nil, err
	}

	readTimeout, err := RequireEnv("SERVER_READ_TIMEOUT")
	if err != nil {
		return nil, err
	}
	if serverCfg.ReadTimeout, err = time.ParseDuration(readTimeout); err != nil {
		return nil, fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := RequireEnv("SERVER_WRITE_TIMEOUT")
	if err != nil {
		return nil, err
	}
	if serverCfg.WriteTimeout, err = time.ParseDuration(writeTimeout); err != nil {
		return nil, fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
	}

	shutdownTimeout, err := RequireEnv("SERVER_SHUTDOWN_TIMEOUT")
	if err != nil {
		return nil, err
	}
	if serverCfg.ShutdownTimeout, err = time.ParseDuration(shutdownTimeout); err != nil {
		return nil, fmt.Errorf("invalid SERVER_SHUTDOWN_TIMEOUT: %w", err)
	}

	maxHeaderSize, err := RequireEnv("SERVER_MAX_HEADER_SIZE")
	if err != nil {
		return nil, err
	}
	if serverCfg.MaxHeaderSize, err = parseByteSize(maxHeaderSize); err != nil {
		return nil, fmt.Errorf("invalid SERVER_MAX_HEADER_SIZE: %w", err)
	}

	maxBodySize, err := RequireEnv("SERVER_MAX_BODY_SIZE")
	if err != nil {
		return nil, err
	}
	if serverCfg.MaxBodySize, err = parseByteSize(maxBodySize); err != nil {
		return nil, fmt.Errorf("invalid SERVER_MAX_BODY_SIZE: %w", err)
	}

	cfg.Server = serverCfg

	// System config
	systemCfg := SystemConfig{}

	systemEnabled, err := RequireEnv("SYSTEM_API_ENABLED")
	if err != nil {
		return nil, err
	}
	if systemCfg.Enabled, err = strconv.ParseBool(systemEnabled); err != nil {
		return nil, fmt.Errorf("invalid SYSTEM_API_ENABLED: %w", err)
	}

	systemCfg.Prefix, err = RequireEnv("SYSTEM_API_PREFIX")
	if err != nil {
		return nil, err
	}

	metricsEnabled, err := RequireEnv("METRICS_ENABLED")
	if err != nil {
		return nil, err
	}
	if systemCfg.MetricsEnabled, err = strconv.ParseBool(metricsEnabled); err != nil {
		return nil, fmt.Errorf("invalid METRICS_ENABLED: %w", err)
	}

	healthEnabled, err := RequireEnv("DETAILED_HEALTH_ENABLED")
	if err != nil {
		return nil, err
	}
	if systemCfg.HealthEnabled, err = strconv.ParseBool(healthEnabled); err != nil {
		return nil, fmt.Errorf("invalid DETAILED_HEALTH_ENABLED: %w", err)
	}

	systemCfg.AdminTemplate, err = RequireEnv("ADMIN_TEMPLATE")
	if err != nil {
		return nil, err
	}

	cfg.System = systemCfg

	// API config
	apiCfg := APIConfig{}

	apiCfg.Prefix, err = RequireEnv("API_PREFIX")
	if err != nil {
		return nil, err
	}

	apiCfg.Version, err = RequireEnv("API_VERSION")
	if err != nil {
		return nil, err
	}

	cfg.API = apiCfg

	// Logging config
	loggingCfg := LoggingConfig{}
	if loggingCfg.Level, err = RequireEnv("LOG_LEVEL"); err != nil {
		return nil, err
	}

	if useJSON, err := RequireEnv("LOG_USE_JSON"); err != nil {
		return nil, err
	} else if loggingCfg.UseJSON, err = strconv.ParseBool(useJSON); err != nil {
		return nil, fmt.Errorf("invalid LOG_USE_JSON: %w", err)
	}

	if enableFile, err := RequireEnv("LOG_ENABLE_FILE"); err != nil {
		return nil, err
	} else if loggingCfg.EnableFile, err = strconv.ParseBool(enableFile); err != nil {
		return nil, fmt.Errorf("invalid LOG_ENABLE_FILE: %w", err)
	}

	if loggingCfg.FilePath, err = RequireEnv("LOG_FILE_PATH"); err != nil {
		return nil, err
	}

	if rotationSize, err := RequireEnv("LOG_ROTATION_SIZE"); err != nil {
		return nil, err
	} else if loggingCfg.RotationSize, err = parseByteSize(rotationSize); err != nil {
		return nil, fmt.Errorf("invalid LOG_ROTATION_SIZE: %w", err)
	}

	if maxAge, err := RequireEnv("LOG_MAX_AGE"); err != nil {
		return nil, err
	} else if loggingCfg.MaxAge, err = time.ParseDuration(maxAge); err != nil {
		return nil, fmt.Errorf("invalid LOG_MAX_AGE: %w", err)
	}

	if compression, err := RequireEnv("LOG_COMPRESSION"); err != nil {
		return nil, err
	} else if loggingCfg.Compression, err = strconv.ParseBool(compression); err != nil {
		return nil, fmt.Errorf("invalid LOG_COMPRESSION: %w", err)
	}

	cfg.Logging = loggingCfg

	// Security config
	securityCfg := SecurityConfig{}

	if enableTLS, err := RequireEnv("SECURITY_ENABLE_TLS"); err != nil {
		return nil, err
	} else if securityCfg.EnableTLS, err = strconv.ParseBool(enableTLS); err != nil {
		return nil, fmt.Errorf("invalid SECURITY_ENABLE_TLS: %w", err)
	}

	if securityCfg.TLSCertPath, err = RequireEnv("TLS_CERT_PATH"); err != nil {
		return nil, err
	}

	if securityCfg.TLSKeyPath, err = RequireEnv("TLS_KEY_PATH"); err != nil {
		return nil, err
	}

	if allowedOrigins, err := RequireEnv("ALLOWED_ORIGINS"); err != nil {
		return nil, err
	} else {
		securityCfg.AllowedOrigins = strings.Split(allowedOrigins, ",")
	}

	// Auth config
	authCfg := AuthConfig{}
	if enabled, err := RequireEnv("AUTH_ENABLED"); err != nil {
		return nil, err
	} else if authCfg.Enabled, err = strconv.ParseBool(enabled); err != nil {
		return nil, fmt.Errorf("invalid AUTH_ENABLED: %w", err)
	}

	if authCfg.Username, err = RequireEnv("AUTH_USERNAME"); err != nil {
		return nil, err
	}

	if authCfg.Password, err = RequireEnv("AUTH_PASSWORD"); err != nil {
		return nil, err
	}

	if excludePaths, err := RequireEnv("AUTH_EXCLUDE_PATHS"); err != nil {
		return nil, err
	} else {
		authCfg.ExcludePaths = strings.Split(excludePaths, ",")
	}

	securityCfg.Auth = authCfg

	// Rate limit config
	rateLimitCfg := RateLimitConfig{
		Routes: make(map[string]RateLimitRule),
	}

	if enabled, err := RequireEnv("RATE_LIMIT_ENABLED"); err != nil {
		return nil, err
	} else if rateLimitCfg.Enabled, err = strconv.ParseBool(enabled); err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_ENABLED: %w", err)
	}

	if requests, err := RequireEnv("RATE_LIMIT_REQUESTS"); err != nil {
		return nil, err
	} else if rateLimitCfg.Requests, err = strconv.Atoi(requests); err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_REQUESTS: %w", err)
	}

	if window, err := RequireEnv("RATE_LIMIT_WINDOW"); err != nil {
		return nil, err
	} else if rateLimitCfg.Window, err = time.ParseDuration(window); err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW: %w", err)
	}

	if byIP, err := RequireEnv("RATE_LIMIT_BY_IP"); err != nil {
		return nil, err
	} else if rateLimitCfg.ByIP, err = strconv.ParseBool(byIP); err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_BY_IP: %w", err)
	}

	if byRoute, err := RequireEnv("RATE_LIMIT_BY_ROUTE"); err != nil {
		return nil, err
	} else if rateLimitCfg.ByRoute, err = strconv.ParseBool(byRoute); err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_BY_ROUTE: %w", err)
	}

	securityCfg.RateLimit = rateLimitCfg

	// CORS config
	corsCfg := CORSConfig{}

	if enabled, err := RequireEnv("CORS_ENABLED"); err != nil {
		return nil, err
	} else if corsCfg.Enabled, err = strconv.ParseBool(enabled); err != nil {
		return nil, fmt.Errorf("invalid CORS_ENABLED: %w", err)
	}

	if allowedOrigins, err := RequireEnv("CORS_ALLOWED_ORIGINS"); err != nil {
		return nil, err
	} else {
		corsCfg.AllowedOrigins = strings.Split(allowedOrigins, ",")
	}

	if allowedMethods, err := RequireEnv("CORS_ALLOWED_METHODS"); err != nil {
		return nil, err
	} else {
		corsCfg.AllowedMethods = strings.Split(allowedMethods, ",")
	}

	if allowedHeaders, err := RequireEnv("CORS_ALLOWED_HEADERS"); err != nil {
		return nil, err
	} else {
		corsCfg.AllowedHeaders = strings.Split(allowedHeaders, ",")
	}

	if exposedHeaders, err := RequireEnv("CORS_EXPOSED_HEADERS"); err != nil {
		return nil, err
	} else {
		corsCfg.ExposedHeaders = strings.Split(exposedHeaders, ",")
	}

	if allowCredentials, err := RequireEnv("CORS_ALLOW_CREDENTIALS"); err != nil {
		return nil, err
	} else if corsCfg.AllowCredentials, err = strconv.ParseBool(allowCredentials); err != nil {
		return nil, fmt.Errorf("invalid CORS_ALLOW_CREDENTIALS: %w", err)
	}

	if maxAge, err := RequireEnv("CORS_MAX_AGE"); err != nil {
		return nil, err
	} else if corsCfg.MaxAge, err = strconv.Atoi(maxAge); err != nil {
		return nil, fmt.Errorf("invalid CORS_MAX_AGE: %w", err)
	}

	securityCfg.CORS = corsCfg

	// Security headers config
	headersCfg := SecurityHeadersConfig{}

	if headersCfg.XSSProtection, err = RequireEnv("HEADER_XSS_PROTECTION"); err != nil {
		return nil, err
	}

	if headersCfg.ContentTypeOptions, err = RequireEnv("HEADER_CONTENT_TYPE_OPTIONS"); err != nil {
		return nil, err
	}

	if headersCfg.XFrameOptions, err = RequireEnv("HEADER_X_FRAME_OPTIONS"); err != nil {
		return nil, err
	}

	if headersCfg.ContentSecurityPolicy, err = RequireEnv("HEADER_CONTENT_SECURITY_POLICY"); err != nil {
		return nil, err
	}

	if headersCfg.ReferrerPolicy, err = RequireEnv("HEADER_REFERRER_POLICY"); err != nil {
		return nil, err
	}

	if headersCfg.StrictTransportSecurity, err = RequireEnv("HEADER_STRICT_TRANSPORT_SECURITY"); err != nil {
		return nil, err
	}

	if headersCfg.PermissionsPolicy, err = RequireEnv("HEADER_PERMISSIONS_POLICY"); err != nil {
		return nil, err
	}

	securityCfg.Headers = headersCfg
	cfg.Security = securityCfg

	// Admin config
	adminCfg := AdminConfig{}

	if enabled, err := RequireEnv("ADMIN_ENABLED"); err != nil {
		return nil, err
	} else if adminCfg.Enabled, err = strconv.ParseBool(enabled); err != nil {
		return nil, fmt.Errorf("invalid ADMIN_ENABLED: %w", err)
	}

	if adminCfg.Path, err = RequireEnv("ADMIN_PATH"); err != nil {
		return nil, err
	}

	if refreshInterval, err := RequireEnv("ADMIN_REFRESH_INTERVAL"); err != nil {
		return nil, err
	} else if adminCfg.RefreshInterval, err = time.ParseDuration(refreshInterval); err != nil {
		return nil, fmt.Errorf("invalid ADMIN_REFRESH_INTERVAL: %w", err)
	}

	if adminCfg.Username, err = RequireEnv("ADMIN_USERNAME"); err != nil {
		return nil, err
	}

	if adminCfg.Password, err = RequireEnv("ADMIN_PASSWORD"); err != nil {
		return nil, err
	}

	cfg.Admin = adminCfg

	// Validate the configuration
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// parseByteSize parses string like "10MB" into bytes
func parseByteSize(value string) (int64, error) {
	if value == "" {
		return 0, fmt.Errorf("byte size cannot be empty")
	}

	value = strings.TrimSpace(value)
	multiplier := int64(1)

	if len(value) > 2 {
		unit := strings.ToUpper(value[len(value)-2:])
		switch unit {
		case "KB":
			multiplier = 1024
			value = value[:len(value)-2]
		case "MB":
			multiplier = 1024 * 1024
			value = value[:len(value)-2]
		case "GB":
			multiplier = 1024 * 1024 * 1024
			value = value[:len(value)-2]
		case "TB":
			multiplier = 1024 * 1024 * 1024 * 1024
			value = value[:len(value)-2]
		default:
			// Check for single letter suffix
			if len(value) > 1 {
				lastChar := strings.ToUpper(value[len(value)-1:])
				switch lastChar {
				case "K":
					multiplier = 1024
					value = value[:len(value)-1]
				case "M":
					multiplier = 1024 * 1024
					value = value[:len(value)-1]
				case "G":
					multiplier = 1024 * 1024 * 1024
					value = value[:len(value)-1]
				case "T":
					multiplier = 1024 * 1024 * 1024 * 1024
					value = value[:len(value)-1]
				}
			}
		}
	}

	bytes, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size format: %w", err)
	}

	return bytes * multiplier, nil
}

// NewDefaultConfig creates a new configuration with sensible defaults
// This can be used programmatically without environment variables
func NewDefaultConfig() *Config {
	return &Config{
		Environment: EnvDevelopment,
		Server: ServerConfig{
			Port:            8080,
			Host:            "localhost",
			ReadTimeout:     time.Second * 30,
			WriteTimeout:    time.Second * 30,
			ShutdownTimeout: time.Second * 30,
			MaxHeaderSize:   1 << 20,  // 1MB
			MaxBodySize:     10 << 20, // 10MB
		},
		System: SystemConfig{
			Enabled:        true,
			Prefix:         "/system",
			MetricsEnabled: true,
			HealthEnabled:  true,
			AdminTemplate:  "admin",
		},
		API: APIConfig{
			Prefix:  "/api",
			Version: "v1",
		},
		Logging: LoggingConfig{
			Level:        "info",
			UseJSON:      false,
			EnableFile:   false,
			FilePath:     "logs/server.log",
			RotationSize: 100 << 20,       // 100MB
			MaxAge:       time.Hour * 168, // 7 days
			Compression:  true,
		},
		Security: SecurityConfig{
			EnableTLS:      false,
			TLSCertPath:    "./certs/server.crt",
			TLSKeyPath:     "./certs/server.key",
			AllowedOrigins: []string{"*"},
			RateLimit: RateLimitConfig{
				Enabled:  false,
				Requests: 100,
				Window:   time.Minute,
				ByIP:     true,
				Routes:   make(map[string]RateLimitRule),
			},
			CORS: CORSConfig{
				Enabled:          false,
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Content-Type", "Authorization"},
				ExposedHeaders:   []string{"Content-Length", "X-Request-ID"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			Headers: SecurityHeadersConfig{
				XSSProtection:           "1; mode=block",
				ContentTypeOptions:      "nosniff",
				XFrameOptions:           "DENY",
				ContentSecurityPolicy:   "default-src 'self'",
				ReferrerPolicy:          "strict-origin-when-cross-origin",
				StrictTransportSecurity: "max-age=31536000; includeSubDomains",
				PermissionsPolicy:       "camera=(), microphone=(), geolocation=()",
			},
			Auth: AuthConfig{
				Enabled:      false,
				Username:     "admin",
				Password:     "password",
				ExcludePaths: []string{"/api/v1/health", "/static"},
			},
		},
		Admin: AdminConfig{
			Enabled:         true,
			Path:            "/admin",
			RefreshInterval: time.Second * 10,
			Username:        "admin",
			Password:        "admin",
		},
	}
}

// LoadWithOptions creates a configuration using provided options or environment variables as fallback
// This allows for programmatic configuration overrides
func LoadWithOptions(options ...ConfigOption) (*Config, *logger.Logger) {
	// Start with default configuration
	cfg := NewDefaultConfig()

	// Try to load environment from .env file if present
	// This is optional and not an error if missing
	_ = LoadEnvFile(".env")

	// Apply options to override defaults
	for _, option := range options {
		option(cfg)
	}

	// Try to use environment variables to override any remaining settings
	overrideFromEnv(cfg)

	// Configure logger based on settings
	log := logger.New(
		logger.WithLevel(cfg.Logging.Level),
		logger.WithJSON(cfg.Logging.UseJSON),
	)

	// Log configuration loaded
	if cfg.Environment == EnvDevelopment {
		log.Info("configuration loaded successfully", map[string]interface{}{
			"environment": cfg.Environment,
		})
	}

	return cfg, log
}

// ConfigOption represents a function that can modify a configuration
type ConfigOption func(*Config)

// WithServerPort sets the server port
func WithServerPort(port int) ConfigOption {
	return func(cfg *Config) {
		cfg.Server.Port = port
	}
}

// WithServerHost sets the server host
func WithServerHost(host string) ConfigOption {
	return func(cfg *Config) {
		cfg.Server.Host = host
	}
}

// WithEnvironment sets the environment
func WithEnvironment(env string) ConfigOption {
	return func(cfg *Config) {
		cfg.Environment = env
	}
}

// WithAuthEnabled enables or disables authentication
func WithAuthEnabled(enabled bool) ConfigOption {
	return func(cfg *Config) {
		cfg.Security.Auth.Enabled = enabled
	}
}

// WithAuthCredentials sets the authentication credentials
func WithAuthCredentials(username, password string) ConfigOption {
	return func(cfg *Config) {
		cfg.Security.Auth.Username = username
		cfg.Security.Auth.Password = password
	}
}

// WithSystemAPI enables or disables the system API
func WithSystemAPI(enabled bool, prefix string) ConfigOption {
	return func(cfg *Config) {
		cfg.System.Enabled = enabled
		if prefix != "" {
			cfg.System.Prefix = prefix
		}
	}
}

// WithLogging configures logging
func WithLogging(level string, useJSON bool) ConfigOption {
	return func(cfg *Config) {
		cfg.Logging.Level = level
		cfg.Logging.UseJSON = useJSON
	}
}

// WithTLS enables or disables TLS
func WithTLS(enabled bool, certPath, keyPath string) ConfigOption {
	return func(cfg *Config) {
		cfg.Security.EnableTLS = enabled
		if certPath != "" {
			cfg.Security.TLSCertPath = certPath
		}
		if keyPath != "" {
			cfg.Security.TLSKeyPath = keyPath
		}
	}
}

// overrideFromEnv overrides configuration with environment variables
func overrideFromEnv(cfg *Config) {
	// Only override from environment if variables are explicitly set

	// Environment
	if env := os.Getenv("GO_ENV"); env != "" {
		cfg.Environment = env
	}

	// Server
	if port, err := strconv.Atoi(os.Getenv("SERVER_PORT")); err == nil && port > 0 {
		cfg.Server.Port = port
	}

	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	if readTimeout, err := time.ParseDuration(os.Getenv("SERVER_READ_TIMEOUT")); err == nil {
		cfg.Server.ReadTimeout = readTimeout
	}

	if writeTimeout, err := time.ParseDuration(os.Getenv("SERVER_WRITE_TIMEOUT")); err == nil {
		cfg.Server.WriteTimeout = writeTimeout
	}

	if shutdownTimeout, err := time.ParseDuration(os.Getenv("SERVER_SHUTDOWN_TIMEOUT")); err == nil {
		cfg.Server.ShutdownTimeout = shutdownTimeout
	}

	// System
	if enabled, err := strconv.ParseBool(os.Getenv("SYSTEM_API_ENABLED")); err == nil {
		cfg.System.Enabled = enabled
	}

	if prefix := os.Getenv("SYSTEM_API_PREFIX"); prefix != "" {
		cfg.System.Prefix = prefix
	}

	// Auth
	if enabled, err := strconv.ParseBool(os.Getenv("AUTH_ENABLED")); err == nil {
		cfg.Security.Auth.Enabled = enabled
	}

	if username := os.Getenv("AUTH_USERNAME"); username != "" {
		cfg.Security.Auth.Username = username
	}

	if password := os.Getenv("AUTH_PASSWORD"); password != "" {
		cfg.Security.Auth.Password = password
	}

	if paths := os.Getenv("AUTH_EXCLUDE_PATHS"); paths != "" {
		cfg.Security.Auth.ExcludePaths = strings.Split(paths, ",")
	}
}
